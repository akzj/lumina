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
	effectHooks        []*EffectHook        // useEffect hooks in call order
	layoutEffectHooks  []*LayoutEffectHook  // useLayoutEffect hooks in call order
	memoHooks          []*MemoHook          // useMemo hooks in call order
	externalStoreHooks []*ExternalStoreHook // useSyncExternalStore hooks
	animationHooks     []string             // animation IDs owned by this component
	effectHookIndex       int // current index for useEffect hooks
	memoHookIndex         int // current index for useMemo hooks
	layoutEffectHookIndex int // current index for useLayoutEffect hooks
	generalHookIndex      int // current index for useReducer, useId, useTransition, useDeferredValue, useSyncExternalStore, useAnimation
	debugValues        []string             // useDebugValue labels (reset each render)
	// Component tree
	Parent       *Component     // Parent component (nil for root)
	ChildComps   []*Component   // Child component instances
	// Error boundary
	IsErrorBoundary bool            // true if this component catches child render errors
	IsRoot          bool            // true if this component was created by lumina.mount()
	FallbackFn      *luaFunctionRef // fallback(errorMsg) → VNode
	CaughtError     string          // last caught error message (empty = no error)
	// Memoization
	Memoized  bool           // true if wrapped with lumina.memo()
	LastProps map[string]any // props from last successful render (for memo comparison)
	// Context
	ContextValues map[int64]any // context ID → value (provided by this component)
	// Prop function refs — Lua registry refs for function-valued props.
	// Managed per-component to avoid SwapRenderRefs freeing them and
	// causing registry slot reuse that corrupts renderFn refs.
	propRefs []int
	mu       sync.RWMutex
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

	// Register component and mark dirty for initial render
	globalRegistry.mu.Lock()
	globalRegistry.components[comp.ID] = comp
	globalRegistry.mu.Unlock()
	comp.Dirty.Store(true)

	return comp, nil
}

// SetState updates a component state value and marks it dirty.
func (c *Component) SetState(key string, value any) {
	c.mu.Lock()
	c.State[key] = value
	c.Dirty.Store(true)
	c.mu.Unlock()

	// Walk up to root component and mark it dirty too.
	// Child components only re-render when the root re-renders and
	// hits luaComponentToVNode inline, so the root must be dirty.
	root := c
	for root.Parent != nil {
		root = root.Parent
	}
	if root != c {
		root.Dirty.Store(true)
	}

	// Notify render loop that this component needs re-render
	if root.RenderNotify != nil {
		select {
		case root.RenderNotify <- struct{}{}:
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

// RenderRefID returns the Lua registry reference ID for the render function.
func (c *Component) RenderRefID() int {
	if c.renderFn == nil {
		return -1
	}
	return c.renderFn.RefID
}

// PushInitFn pushes the init function onto the stack.
func (c *Component) PushInitFn(L *lua.State) bool {
	if c.initFn == nil {
		return false
	}
	L.RawGetI(lua.RegistryIndex, int64(c.initFn.RefID))
	return L.Type(-1) == lua.TypeFunction
}

// ReleasePropRefs releases all Lua registry refs for function-valued props.
// Called before updating props (to release old refs) and during Cleanup.
func (c *Component) ReleasePropRefs(L *lua.State) {
	for _, ref := range c.propRefs {
		L.Unref(lua.RegistryIndex, ref)
	}
	c.propRefs = nil
}

// Cleanup removes the component from registry and cleans up references.
// Runs effect and layout-effect cleanups before releasing Lua refs.
func (c *Component) Cleanup(L *lua.State) {
	// Run effect cleanups (useEffect return functions)
	RunEffectCleanups(L, c)
	// Run layout effect cleanups (useLayoutEffect return functions)
	RunLayoutEffectCleanups(L, c)

	// Release prop function refs
	c.ReleasePropRefs(L)

	if c.initFn != nil {
		L.Unref(lua.RegistryIndex, c.initFn.RefID)
	}
	if c.renderFn != nil {
		L.Unref(lua.RegistryIndex, c.renderFn.RefID)
	}
	if c.cleanupFn != nil {
		L.Unref(lua.RegistryIndex, c.cleanupFn.RefID)
	}

	// Remove from parent's child list so findChildComponent won't
	// find this cleaned-up component after resize (stale renderFn refs).
	if c.Parent != nil {
		c.Parent.RemoveChild(c)
	}

	globalRegistry.mu.Lock()
	delete(globalRegistry.components, c.ID)
	globalRegistry.mu.Unlock()
}

// SetCurrentComponent sets the currently rendering component.
// IMPORTANT: This uses a global variable protected by a mutex, which means
// only one component can be rendering at a time. This is safe because the
// Lua VM is single-threaded — all component renders happen sequentially on
// the main goroutine. If concurrent rendering is ever needed, this must be
// replaced with a per-goroutine or per-VM context (e.g., stored on the Lua State).
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

// ClearComponents removes all components from the global registry.
// Used for test isolation.
func ClearComponents() {
	globalRegistry.mu.Lock()
	globalRegistry.components = make(map[string]*Component)
	globalRegistry.current = nil
	globalRegistry.mu.Unlock()
}


// ReconcileComponents compares old and new VNode trees, calling Cleanup on removed components.
func ReconcileComponents(L *lua.State, oldTree, newTree *VNode) {
	oldIDs := collectComponentIDs(oldTree)
	newIDs := collectComponentIDs(newTree)
	for id := range oldIDs {
		if _, exists := newIDs[id]; !exists {
			globalRegistry.mu.RLock()
			comp, ok := globalRegistry.components[id]
			globalRegistry.mu.RUnlock()
			if ok {
				comp.Cleanup(L)
			}
		}
	}
}

// collectComponentIDs walks a VNode tree and returns all component IDs found.
func collectComponentIDs(node *VNode) map[string]bool {
	ids := make(map[string]bool)
	if node == nil {
		return ids
	}
	if node.ComponentRef != nil {
		ids[node.ComponentRef.ID] = true
	}
	for _, child := range node.Children {
		for id := range collectComponentIDs(child) {
			ids[id] = true
		}
	}
	return ids
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

// ResetHookIndex resets all per-type hook call counters for a new render pass.
func (c *Component) ResetHookIndex() {
	c.effectHookIndex = 0
	c.memoHookIndex = 0
	c.layoutEffectHookIndex = 0
	c.generalHookIndex = 0
	c.debugValues = c.debugValues[:0] // reset debug values each render
}

// UpdateProps updates the component's props and marks it dirty if they changed.
// Returns true if props actually changed.
func (c *Component) UpdateProps(newProps map[string]any) bool {
	changed := !propsEqualSkipFuncs(c.Props, newProps)
	c.mu.Lock()
	// Always update props so new function refs (closures) are used,
	// even if only function props changed.
	c.Props = newProps
	if changed {
		c.Dirty.Store(true)
	}
	c.mu.Unlock()
	return changed
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
