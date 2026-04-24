// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/akzj/go-lua/pkg/lua"
)

// Component represents a Lumina UI component instance.
// It wraps the component definition with runtime state.
type Component struct {
	ID           string // Unique identifier
	Type         string // Component type name (e.g., "Counter")
	Name         string // Component name
	Props        map[string]any
	State        map[string]any
	Dirty        atomic.Bool   // True if state changed and needs re-render
	RenderNotify chan struct{} // Notify render loop when state changes
	LastVNode    *VNode        // Previous render result for VDom diffing
	initFn       *luaFunctionRef
	renderFn     *luaFunctionRef
	cleanupFn    *luaFunctionRef
	// Hooks state
	effectHooks  []*EffectHook  // useEffect hooks in call order
	memoHooks    []*MemoHook    // useMemo hooks in call order
	hookIndex    int            // current hook index during render
	// Component tree
	Parent       *Component     // Parent component (nil for root)
	ChildComps   []*Component   // Child component instances
	mu           sync.RWMutex
}

// luaFunctionRef holds a reference to a Lua function on the stack.
type luaFunctionRef struct {
	RefID int // Reference ID in the registry
}

// ComponentRegistry stores all active components.
type ComponentRegistry struct {
	components map[string]*Component
	current    *Component // Currently rendering component
	mu         sync.RWMutex
}

var (
	globalRegistry = &ComponentRegistry{
		components: make(map[string]*Component),
	}
	currentComponentID int64
)

// getNextComponentID generates a unique component ID.
func getNextComponentID() string {
	id := atomic.AddInt64(&currentComponentID, 1)
	return fmt.Sprintf("comp_%d", id)
}

// NewComponent creates a new Component instance from a component factory table.
func NewComponent(L *lua.State, factoryIdx int, props map[string]any) (*Component, error) {
	// Get component name
	L.GetField(factoryIdx, "name")
	name, _ := L.ToString(-1)
	L.Pop(1)

	if name == "" {
		return nil, errMissingName
	}

	comp := &Component{
		ID:           getNextComponentID(),
		Type:         name,
		Name:         name,
		Props:        props,
		State:        make(map[string]any),
		RenderNotify: make(chan struct{}, 1),
	}

	// Extract init function reference
	L.GetField(factoryIdx, "init")
	if L.Type(-1) == lua.TypeFunction {
		refID := L.Ref(lua.RegistryIndex)
		comp.initFn = &luaFunctionRef{RefID: refID}
	} else {
		L.Pop(1) // pop nil
	}

	// Extract render function reference
	L.GetField(factoryIdx, "render")
	if L.Type(-1) == lua.TypeFunction {
		refID := L.Ref(lua.RegistryIndex)
		comp.renderFn = &luaFunctionRef{RefID: refID}
	} else {
		L.Pop(1) // pop nil
		return nil, errMissingRender
	}

	// Extract cleanup function (optional)
	L.GetField(factoryIdx, "cleanup")
	if L.Type(-1) == lua.TypeFunction {
		refID := L.Ref(lua.RegistryIndex)
		comp.cleanupFn = &luaFunctionRef{RefID: refID}
	} else {
		L.Pop(1) // pop nil
	}

	// Register component
	globalRegistry.mu.Lock()
	globalRegistry.components[comp.ID] = comp
	globalRegistry.mu.Unlock()

	return comp, nil
}

// SetState updates a component state value and marks it dirty.
func (c *Component) SetState(key string, value any) {
	c.mu.Lock()
	c.State[key] = value
	c.Dirty.Store(true)
	c.mu.Unlock()

	// Notify render loop that this component needs re-render
	if c.RenderNotify != nil {
		select {
		case c.RenderNotify <- struct{}{}:
		default:
		}
	}
}

// GetState returns a state value.
func (c *Component) GetState(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.State[key]
	return v, ok
}

// MarkClean marks the component as not needing re-render.
func (c *Component) MarkClean() {
	c.Dirty.Store(false)
}

// GetID returns the component's unique ID.
func (c *Component) GetID() string {
	return c.ID
}

// GetType returns the component's type name.
func (c *Component) GetType() string {
	return c.Type
}

// PushRenderFn pushes the render function onto the stack.
func (c *Component) PushRenderFn(L *lua.State) bool {
	if c.renderFn == nil {
		return false
	}
	L.RawGetI(lua.RegistryIndex, int64(c.renderFn.RefID))
	return L.Type(-1) == lua.TypeFunction
}

// PushInitFn pushes the init function onto the stack.
func (c *Component) PushInitFn(L *lua.State) bool {
	if c.initFn == nil {
		return false
	}
	L.RawGetI(lua.RegistryIndex, int64(c.initFn.RefID))
	return L.Type(-1) == lua.TypeFunction
}

// Cleanup removes the component from registry and cleans up references.
func (c *Component) Cleanup(L *lua.State) {
	if c.initFn != nil {
		L.Unref(lua.RegistryIndex, c.initFn.RefID)
	}
	if c.renderFn != nil {
		L.Unref(lua.RegistryIndex, c.renderFn.RefID)
	}
	if c.cleanupFn != nil {
		L.Unref(lua.RegistryIndex, c.cleanupFn.RefID)
	}

	globalRegistry.mu.Lock()
	delete(globalRegistry.components, c.ID)
	globalRegistry.mu.Unlock()
}

// SetCurrentComponent sets the currently rendering component.
func SetCurrentComponent(comp *Component) {
	globalRegistry.mu.Lock()
	globalRegistry.current = comp
	globalRegistry.mu.Unlock()
}

// GetCurrentComponent returns the currently rendering component.
func GetCurrentComponent() *Component {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return globalRegistry.current
}

// GetComponentByID returns a component by its ID.
func GetComponentByID(id string) (*Component, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	comp, ok := globalRegistry.components[id]
	return comp, ok
}

// ComponentToUserdata creates a userdata wrapping a Component and pushes it to the stack.
func ComponentToUserdata(L *lua.State, comp *Component) {
	L.NewUserdata(0, 0)
	L.SetUserdataValue(-1, comp)
}

// UserdataToComponent extracts a Component from a userdata on the stack.
func UserdataToComponent(L *lua.State, idx int) *Component {
	val := L.UserdataValue(idx)
	if comp, ok := val.(*Component); ok {
		return comp
	}
	return nil
}

// ResetHookIndex resets the hook call counter for a new render pass.
func (c *Component) ResetHookIndex() {
	c.hookIndex = 0
}

// UpdateProps updates the component's props and marks it dirty if they changed.
// Returns true if props actually changed.
func (c *Component) UpdateProps(newProps map[string]any) bool {
	if propsEqual(c.Props, newProps) {
		return false
	}
	c.mu.Lock()
	c.Props = newProps
	c.Dirty.Store(true)
	c.mu.Unlock()
	return true
}

// AddChild adds a child component to this component's tree.
func (c *Component) AddChild(child *Component) {
	c.mu.Lock()
	c.ChildComps = append(c.ChildComps, child)
	c.mu.Unlock()
	child.mu.Lock()
	child.Parent = c
	child.mu.Unlock()
}

// RemoveChild removes a child component from this component's tree.
func (c *Component) RemoveChild(child *Component) {
	c.mu.Lock()
	for i, ch := range c.ChildComps {
		if ch.ID == child.ID {
			c.ChildComps = append(c.ChildComps[:i], c.ChildComps[i+1:]...)
			break
		}
	}
	c.mu.Unlock()
	child.mu.Lock()
	child.Parent = nil
	child.mu.Unlock()
}

// GetChildren returns a copy of the child components slice.
func (c *Component) GetChildren() []*Component {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*Component, len(c.ChildComps))
	copy(result, c.ChildComps)
	return result
}

// Error variables for component creation
var (
	errMissingName   = errors.New("component factory missing 'name' field")
	errMissingRender = errors.New("component missing 'render' function")
)
