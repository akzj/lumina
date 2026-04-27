// Package component manages Component lifecycle, state, rendering,
// reconciliation, and event handler extraction for Lumina v2.
package component

import (
	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/compositor"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/paint"
)

// RenderFunc is the component's render function.
// Called with (state, props), returns a VNode tree.
type RenderFunc func(state map[string]any, props map[string]any) *layout.VNode

// Component represents a stateful rendering unit.
type Component struct {
	ID          string
	Name        string
	Buffer      *buffer.Buffer
	Rect        buffer.Rect
	PrevRect    buffer.Rect
	ZIndex      int
	DirtyPaint  bool
	RectChanged bool

	State    map[string]any
	Props    map[string]any
	RenderFn RenderFunc
	VNodeTree *layout.VNode

	Parent   *Component
	Children []*Component
	ChildMap map[string]*Component

	Handlers   map[string]event.HandlerMap // vnodeID → HandlerMap
	Focusables []string
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
	m.components[comp.ID] = comp
}

// Unregister removes a component (and all its descendants) from the manager.
func (m *Manager) Unregister(id string) {
	comp := m.components[id]
	if comp == nil {
		return
	}
	// Recursively unregister children.
	for _, child := range comp.Children {
		m.Unregister(child.ID)
	}
	delete(m.components, id)
}

// Get returns the component with the given ID, or nil.
func (m *Manager) Get(id string) *Component {
	return m.components[id]
}

// SetState sets a single state key on a component and marks it dirty.
func (m *Manager) SetState(compID string, key string, value any) {
	comp := m.components[compID]
	if comp == nil {
		return
	}
	comp.State[key] = value
	comp.DirtyPaint = true
}

// GetDirtyPaint returns all components that need re-rendering.
func (m *Manager) GetDirtyPaint() []*Component {
	var dirty []*Component
	for _, comp := range m.components {
		if comp.DirtyPaint {
			dirty = append(dirty, comp)
		}
	}
	return dirty
}

// GetRectChanged returns all components whose rect changed since last clear.
func (m *Manager) GetRectChanged() []*Component {
	var changed []*Component
	for _, comp := range m.components {
		if comp.RectChanged {
			changed = append(changed, comp)
		}
	}
	return changed
}

// ClearDirty resets DirtyPaint and RectChanged on all components.
func (m *Manager) ClearDirty() {
	for _, comp := range m.components {
		comp.DirtyPaint = false
		comp.RectChanged = false
	}
}

// RenderDirty renders all dirty components: calls RenderFn, computes layout,
// paints into the component buffer, and extracts event handlers.
func (m *Manager) RenderDirty() {
	for _, comp := range m.GetDirtyPaint() {
		comp.VNodeTree = comp.RenderFn(comp.State, comp.Props)
		layout.ComputeLayout(comp.VNodeTree, comp.Rect.X, comp.Rect.Y, comp.Rect.W, comp.Rect.H)
		comp.Buffer.Clear()
		m.painter.Paint(comp.Buffer, comp.VNodeTree, comp.Rect.X, comp.Rect.Y)
		comp.ExtractHandlers()
		comp.DirtyPaint = false
	}
}

// AllLayers converts all registered components to event.ComponentLayer slices
// for use by the compositor and event system.
func (m *Manager) AllLayers() []*event.ComponentLayer {
	var layers []*event.ComponentLayer
	for _, comp := range m.components {
		layers = append(layers, &event.ComponentLayer{
			Layer: &compositor.Layer{
				ID:     comp.ID,
				Buffer: comp.Buffer,
				Rect:   comp.Rect,
				ZIndex: comp.ZIndex,
			},
			VNodeTree: comp.VNodeTree,
		})
	}
	return layers
}

// Reconcile reconciles the children of a parent component against a new set
// of child descriptors. Existing children with matching keys are updated;
// new children are created; missing children are unregistered.
func (m *Manager) Reconcile(parent *Component, newChildren []ChildDescriptor) {
	oldMap := parent.ChildMap
	newMap := make(map[string]*Component, len(newChildren))
	newList := make([]*Component, 0, len(newChildren))

	for _, desc := range newChildren {
		if existing, ok := oldMap[desc.Key]; ok {
			// Existing child — update props if changed.
			if !shallowEqual(existing.Props, desc.Props) {
				existing.Props = desc.Props
				existing.DirtyPaint = true
			}
			existing.RenderFn = desc.RenderFn
			newMap[desc.Key] = existing
			newList = append(newList, existing)
		} else {
			// New child — create.
			comp := &Component{
				ID:         parent.ID + ":" + desc.Key,
				Name:       desc.Name,
				Props:      desc.Props,
				RenderFn:   desc.RenderFn,
				State:      make(map[string]any),
				ChildMap:   make(map[string]*Component),
				Parent:     parent,
				DirtyPaint: true,
				Handlers:   make(map[string]event.HandlerMap),
			}
			m.Register(comp)
			newMap[desc.Key] = comp
			newList = append(newList, comp)
		}
	}

	// Remove children that are no longer present.
	for key, old := range oldMap {
		if _, ok := newMap[key]; !ok {
			m.Unregister(old.ID)
		}
	}

	parent.Children = newList
	parent.ChildMap = newMap
}
