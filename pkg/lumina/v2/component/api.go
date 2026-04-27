// Package component manages Component lifecycle, state, rendering,
// reconciliation, and event handler extraction for Lumina v2.
package component

import (
	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/paint"
)

// HandlerFunc is an opaque event handler stored on a component.
// The event package will type-assert this to event.EventHandler when dispatching.
type HandlerFunc = any

// HandlerMap maps event type names (e.g. "click") to handler functions.
type HandlerMap map[string]HandlerFunc

// RenderFunc is the component's render function.
// Called with (state, props), returns a VNode tree.
type RenderFunc func(state map[string]any, props map[string]any) *layout.VNode

// Component represents a stateful rendering unit.
type Component struct {
	id          string
	name        string
	buf         *buffer.Buffer
	rect        buffer.Rect
	prevRect    buffer.Rect
	zIndex      int
	dirtyPaint  bool
	rectChanged bool

	state    map[string]any
	props    map[string]any
	renderFn RenderFunc
	vnodeTree *layout.VNode

	parent   *Component
	children []*Component
	childMap map[string]*Component

	handlers   map[string]HandlerMap // vnodeID → HandlerMap
	focusables []string

	hookStore map[string]any // dedicated storage for bridge hooks
}

// NewComponent creates a new Component with the given parameters.
func NewComponent(id, name string, rect buffer.Rect, zIndex int, renderFn RenderFunc) *Component {
	return &Component{
		id:         id,
		name:       name,
		buf:        buffer.New(rect.W, rect.H),
		rect:       rect,
		prevRect:   rect,
		zIndex:     zIndex,
		dirtyPaint: true,
		state:      make(map[string]any),
		props:      make(map[string]any),
		renderFn:   renderFn,
		childMap:   make(map[string]*Component),
		handlers:   make(map[string]HandlerMap),
		hookStore:  make(map[string]any),
	}
}

// --- Accessors ---

// ID returns the component's unique identifier.
func (c *Component) ID() string { return c.id }

// Name returns the component's name.
func (c *Component) Name() string { return c.name }

// Buffer returns the component's render buffer.
func (c *Component) Buffer() *buffer.Buffer { return c.buf }

// Rect returns the component's current screen rectangle.
func (c *Component) Rect() buffer.Rect { return c.rect }

// PrevRect returns the component's previous screen rectangle.
func (c *Component) PrevRect() buffer.Rect { return c.prevRect }

// ZIndex returns the component's z-order index.
func (c *Component) ZIndex() int { return c.zIndex }

// IsDirtyPaint returns true if the component needs re-rendering.
func (c *Component) IsDirtyPaint() bool { return c.dirtyPaint }

// IsRectChanged returns true if the component's rect changed since last clear.
func (c *Component) IsRectChanged() bool { return c.rectChanged }

// State returns the component's state map.
func (c *Component) State() map[string]any { return c.state }

// Props returns the component's props map.
func (c *Component) Props() map[string]any { return c.props }

// RenderFn returns the component's render function.
func (c *Component) GetRenderFn() RenderFunc { return c.renderFn }

// VNodeTree returns the component's current VNode tree (from last render).
func (c *Component) VNodeTree() *layout.VNode { return c.vnodeTree }

// Handlers returns the component's event handlers (vnodeID → HandlerMap).
func (c *Component) Handlers() map[string]HandlerMap { return c.handlers }

// Focusables returns the IDs of focusable VNodes in this component.
func (c *Component) Focusables() []string { return c.focusables }

// Children returns the component's child components.
func (c *Component) Children() []*Component { return c.children }

// Parent returns the component's parent, or nil.
func (c *Component) Parent() *Component { return c.parent }

// ChildMap returns the component's child map (key → *Component).
func (c *Component) ChildMap() map[string]*Component { return c.childMap }

// HookStore returns dedicated storage for bridge hooks (useState, useEffect, etc.).
// This replaces the previous pattern of abusing Props for hook storage.
func (c *Component) HookStore() map[string]any { return c.hookStore }

// --- Mutation methods ---

// MarkDirty marks the component as needing re-rendering.
func (c *Component) MarkDirty() { c.dirtyPaint = true }

// SetState sets a single state key and marks the component dirty.
func (c *Component) SetState(key string, value any) {
	c.state[key] = value
	c.dirtyPaint = true
}

// Move moves the component to a new rect, updating prevRect and resizing
// the buffer if dimensions changed.
func (c *Component) Move(newRect buffer.Rect) {
	c.prevRect = c.rect
	c.rect = newRect
	c.rectChanged = true
	c.dirtyPaint = true
	if newRect.W != c.buf.Width() || newRect.H != c.buf.Height() {
		c.buf = buffer.New(newRect.W, newRect.H)
	}
}

// ChildDescriptor describes a child component to create/update during reconciliation.
type ChildDescriptor struct {
	Key      string
	Name     string
	Props    map[string]any
	RenderFn RenderFunc
}

// Manager manages the component tree.
type Manager struct {
	components map[string]*Component
	painter    paint.Painter
}

// NewManager creates a new component Manager with the given painter.
func NewManager(painter paint.Painter) *Manager {
	return &Manager{
		components: make(map[string]*Component),
		painter:    painter,
	}
}

// Register adds a component to the manager.
func (m *Manager) Register(comp *Component) {
	m.components[comp.id] = comp
}

// Unregister removes a component (and all its descendants) from the manager.
func (m *Manager) Unregister(id string) {
	comp := m.components[id]
	if comp == nil {
		return
	}
	// Recursively unregister children.
	for _, child := range comp.children {
		m.Unregister(child.id)
	}
	delete(m.components, id)
}

// Get returns the component with the given ID, or nil.
func (m *Manager) Get(id string) *Component {
	return m.components[id]
}

// GetAll returns all registered components.
func (m *Manager) GetAll() []*Component {
	all := make([]*Component, 0, len(m.components))
	for _, c := range m.components {
		all = append(all, c)
	}
	return all
}

// SetState sets a single state key on a component and marks it dirty.
func (m *Manager) SetState(compID string, key string, value any) {
	comp := m.components[compID]
	if comp == nil {
		return
	}
	comp.state[key] = value
	comp.dirtyPaint = true
}

// GetDirtyPaint returns all components that need re-rendering.
func (m *Manager) GetDirtyPaint() []*Component {
	var dirty []*Component
	for _, comp := range m.components {
		if comp.dirtyPaint {
			dirty = append(dirty, comp)
		}
	}
	return dirty
}

// GetRectChanged returns all components whose rect changed since last clear.
func (m *Manager) GetRectChanged() []*Component {
	var changed []*Component
	for _, comp := range m.components {
		if comp.rectChanged {
			changed = append(changed, comp)
		}
	}
	return changed
}

// ClearDirty resets DirtyPaint and RectChanged on all components.
func (m *Manager) ClearDirty() {
	for _, comp := range m.components {
		comp.dirtyPaint = false
		comp.rectChanged = false
	}
}

// RenderDirty renders all dirty components: calls RenderFn, computes layout,
// paints into the component buffer, and extracts event handlers.
func (m *Manager) RenderDirty() {
	for _, comp := range m.GetDirtyPaint() {
		comp.vnodeTree = comp.renderFn(comp.state, comp.props)
		if comp.vnodeTree == nil {
			comp.dirtyPaint = false
			continue // skip layout/paint for nil render result
		}
		layout.ComputeLayout(comp.vnodeTree, comp.rect.X, comp.rect.Y, comp.rect.W, comp.rect.H)
		comp.buf.Clear()
		m.painter.Paint(comp.buf, comp.vnodeTree, comp.rect.X, comp.rect.Y)
		comp.ExtractHandlers()
		comp.dirtyPaint = false
	}
}

// Reconcile reconciles the children of a parent component against a new set
// of child descriptors. Existing children with matching keys are updated;
// new children are created; missing children are unregistered.
func (m *Manager) Reconcile(parent *Component, newChildren []ChildDescriptor) {
	oldMap := parent.childMap
	newMap := make(map[string]*Component, len(newChildren))
	newList := make([]*Component, 0, len(newChildren))

	for _, desc := range newChildren {
		if existing, ok := oldMap[desc.Key]; ok {
			// Existing child — update props if changed.
			if !shallowEqual(existing.props, desc.Props) {
				existing.props = desc.Props
				existing.dirtyPaint = true
			}
			existing.renderFn = desc.RenderFn
			newMap[desc.Key] = existing
			newList = append(newList, existing)
		} else {
			// New child — create.
			comp := &Component{
				id:         parent.id + ":" + desc.Key,
				name:       desc.Name,
				buf:        buffer.New(1, 1),
				rect:       buffer.Rect{W: 1, H: 1},
				props:      desc.Props,
				renderFn:   desc.RenderFn,
				state:      make(map[string]any),
				childMap:   make(map[string]*Component),
				parent:     parent,
				dirtyPaint: true,
				handlers:   make(map[string]HandlerMap),
				hookStore:  make(map[string]any),
			}
			m.Register(comp)
			newMap[desc.Key] = comp
			newList = append(newList, comp)
		}
	}

	// Remove children that are no longer present.
	for key, old := range oldMap {
		if _, ok := newMap[key]; !ok {
			m.Unregister(old.id)
		}
	}

	parent.children = newList
	parent.childMap = newMap
}
