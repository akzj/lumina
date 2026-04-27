// Package v2 provides the composition root for Lumina v2.
// App ties together buffer, layout, paint, compositor, event, component,
// and output into a single render-loop orchestrator.
package v2

import (
	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/compositor"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
	"github.com/akzj/lumina/pkg/lumina/v2/paint"
)

// App is the composition root — ties all v2 modules together.
type App struct {
	width      int
	height     int
	manager    *component.Manager
	compositor *compositor.Compositor
	dispatcher *event.Dispatcher
	adapter    output.Adapter

	// Internal state
	lastDirtyRects []buffer.Rect
}

// NewApp creates a new App with the given screen dimensions and output adapter.
func NewApp(w, h int, adapter output.Adapter) *App {
	painter := paint.NewPainter()
	return &App{
		width:      w,
		height:     h,
		manager:    component.NewManager(painter),
		compositor: compositor.NewCompositor(w, h),
		dispatcher: event.NewDispatcher(),
		adapter:    adapter,
	}
}

// NewTestApp creates an App with a TestAdapter for testing.
func NewTestApp(w, h int) (*App, *output.TestAdapter) {
	ta := output.NewTestAdapter()
	app := NewApp(w, h, ta)
	return app, ta
}

// RegisterComponent registers a component with the app.
// Creates buffer, sets initial rect, registers with manager.
func (a *App) RegisterComponent(id, name string, rect buffer.Rect, zIndex int, renderFn component.RenderFunc) *component.Component {
	comp := &component.Component{
		ID:         id,
		Name:       name,
		Buffer:     buffer.New(rect.W, rect.H),
		Rect:       rect,
		PrevRect:   rect,
		ZIndex:     zIndex,
		DirtyPaint: true,
		State:      make(map[string]any),
		Props:      make(map[string]any),
		RenderFn:   renderFn,
		Children:   nil,
		ChildMap:   make(map[string]*component.Component),
		Handlers:   make(map[string]event.HandlerMap),
	}
	a.manager.Register(comp)
	return comp
}

// UnregisterComponent removes a component.
func (a *App) UnregisterComponent(id string) {
	a.manager.Unregister(id)
}

// SetState updates a component's state (marks it dirty).
func (a *App) SetState(compID string, key string, value any) {
	a.manager.SetState(compID, key, value)
}

// RenderAll renders all components and composes the full screen.
func (a *App) RenderAll() {
	// Render all dirty components (they should already be marked dirty on register).
	a.manager.RenderDirty()

	// Build layers and set on compositor.
	layers := a.buildLayers()
	a.compositor.SetLayers(layers)

	// Compose full screen.
	screen := a.compositor.ComposeAll()

	// Rebuild hit tester and sync handlers/focusables for event dispatch.
	a.rebuildHitTester()
	a.syncHandlers()

	// Output.
	_ = a.adapter.WriteFull(screen)
	_ = a.adapter.Flush()

	a.lastDirtyRects = nil
	a.manager.ClearDirty()
}

// RenderDirty renders only dirty components and composes changed regions.
func (a *App) RenderDirty() {
	// 1. Capture dirty lists BEFORE rendering (RenderDirty clears DirtyPaint).
	paintDirty := a.manager.GetDirtyPaint()
	rectChanged := a.manager.GetRectChanged()

	// 2. Render dirty components (clears DirtyPaint on each).
	a.manager.RenderDirty()

	var allDirtyRects []buffer.Rect

	// 3. Always rebuild layers + occlusion map so compositor has fresh buffers.
	layers := a.buildLayers()
	a.compositor.SetLayers(layers)

	if len(rectChanged) > 0 {
		// Rect changed → recompose old + new rects.
		var rects []buffer.Rect
		for _, comp := range rectChanged {
			rects = append(rects, comp.PrevRect)
			rects = append(rects, comp.Rect)
		}
		allDirtyRects = append(allDirtyRects, a.compositor.ComposeRects(rects)...)
	}

	// 4. Compose dirty layers (paint changes).
	if len(paintDirty) > 0 {
		dirtyLayers := a.getDirtyLayers(paintDirty)
		if len(dirtyLayers) > 0 {
			rects := a.compositor.ComposeDirty(dirtyLayers)
			allDirtyRects = append(allDirtyRects, rects...)
		}
	}

	// 5. Rebuild hit tester + sync handlers.
	a.rebuildHitTester()
	a.syncHandlers()

	// 6. Output.
	if len(allDirtyRects) > 0 {
		_ = a.adapter.WriteDirty(a.compositor.Screen(), allDirtyRects)
		_ = a.adapter.Flush()
	}

	a.lastDirtyRects = allDirtyRects
	a.manager.ClearDirty()
}

// HandleEvent dispatches an input event through the event system.
func (a *App) HandleEvent(e *event.Event) {
	a.dispatcher.Dispatch(e)
}

// Screen returns the current screen buffer.
func (a *App) Screen() *buffer.Buffer {
	return a.compositor.Screen()
}

// DirtyRects returns the dirty rects from the last render.
func (a *App) DirtyRects() []buffer.Rect {
	return a.lastDirtyRects
}

// FocusedID returns the currently focused VNode ID.
func (a *App) FocusedID() string {
	return a.dispatcher.FocusedID()
}

// MoveComponent moves a component to a new rect.
func (a *App) MoveComponent(id string, newRect buffer.Rect) {
	comp := a.manager.Get(id)
	if comp == nil {
		return
	}
	comp.PrevRect = comp.Rect
	comp.Rect = newRect
	comp.RectChanged = true
	comp.DirtyPaint = true

	// Resize buffer if dimensions changed.
	if newRect.W != comp.Buffer.Width() || newRect.H != comp.Buffer.Height() {
		comp.Buffer = buffer.New(newRect.W, newRect.H)
	}
}

// Resize resizes the screen.
func (a *App) Resize(w, h int) {
	a.width = w
	a.height = h
	a.compositor = compositor.NewCompositor(w, h)
	// Mark all components dirty so next render repaints everything.
	for _, comp := range a.manager.GetAll() {
		comp.DirtyPaint = true
	}
}

// --- internal helpers ---

// buildLayers extracts compositor layers from all registered components.
func (a *App) buildLayers() []*compositor.Layer {
	compLayers := a.manager.AllLayers()
	layers := make([]*compositor.Layer, len(compLayers))
	for i, cl := range compLayers {
		layers[i] = cl.Layer
	}
	return layers
}

// rebuildHitTester rebuilds the VNode hit tester from current component layers.
func (a *App) rebuildHitTester() {
	compLayers := a.manager.AllLayers()
	ht := event.NewVNodeHitTester(compLayers, a.compositor.OcclusionMap())
	a.dispatcher.SetHitTester(ht)
}

// syncHandlers syncs event handlers and focusables from all components
// into the dispatcher.
func (a *App) syncHandlers() {
	// Clear stale handlers and focusables from previous render cycle.
	a.dispatcher.ClearAllHandlers()
	a.dispatcher.ClearAllFocusables()

	compLayers := a.manager.AllLayers()

	// Build parent map and register handlers.
	parentMap := make(map[string]string)
	for _, cl := range compLayers {
		if cl.VNodeTree != nil {
			buildParentMap(cl.VNodeTree, "", parentMap)
		}
	}
	a.dispatcher.SetParentMap(parentMap)

	// Register all handlers from all components.
	// First, collect all component objects via manager.
	for _, cl := range compLayers {
		comp := a.manager.Get(cl.Layer.ID)
		if comp == nil {
			continue
		}
		for vnodeID, hm := range comp.Handlers {
			a.dispatcher.RegisterHandlers(vnodeID, hm)
		}
		for i, fID := range comp.Focusables {
			a.dispatcher.RegisterFocusable(fID, i)
		}
	}
}

// buildParentMap recursively builds a vnodeID → parentVNodeID map.
func buildParentMap(vn *layout.VNode, parentID string, m map[string]string) {
	if vn == nil {
		return
	}
	if vn.ID != "" && parentID != "" {
		m[vn.ID] = parentID
	}
	currentID := vn.ID
	if currentID == "" {
		currentID = parentID
	}
	for _, child := range vn.Children {
		buildParentMap(child, currentID, m)
	}
}

// getDirtyLayers converts dirty components to compositor layers.
func (a *App) getDirtyLayers(dirtyComps []*component.Component) []*compositor.Layer {
	var layers []*compositor.Layer
	for _, comp := range dirtyComps {
		layers = append(layers, &compositor.Layer{
			ID:     comp.ID,
			Buffer: comp.Buffer,
			Rect:   comp.Rect,
			ZIndex: comp.ZIndex,
		})
	}
	return layers
}
