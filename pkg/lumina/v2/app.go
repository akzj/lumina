// Package v2 provides the composition root for Lumina v2.
// App ties together buffer, layout, paint, compositor, event, component,
// and output into a single render-loop orchestrator.
package v2

import (
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/animation"
	"github.com/akzj/lumina/pkg/lumina/v2/bridge"
	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/compositor"
	"github.com/akzj/lumina/pkg/lumina/v2/devtools"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
	"github.com/akzj/lumina/pkg/lumina/v2/paint"
	"github.com/akzj/lumina/pkg/lumina/v2/perf"
	"github.com/akzj/lumina/pkg/lumina/v2/router"
)

// App is the composition root — ties all v2 modules together.
type App struct {
	width      int
	height     int
	manager    *component.Manager
	compositor *compositor.Compositor
	dispatcher *event.Dispatcher
	adapter    output.Adapter
	tracker    *perf.Tracker

	// Internal state
	lastDirtyRects []buffer.Rect
	layersDirty    bool // true when components added/removed/resized → need OcclusionMap rebuild

	// DevTools panel
	devtools *devtools.Panel

	// Runtime (populated by NewAppWithLua / Run)
	luaState  *lua.State
	bridge    *bridge.Bridge
	animMgr   *animation.Manager
	routerMgr *router.Router
	quit      chan struct{}
	running   bool
}

// trackerRenderObserver bridges component.RenderObserver to perf.Tracker.
// It skips the "__devtools" component so that DevTools' own rendering does
// not pollute the application's performance statistics.
type trackerRenderObserver struct {
	tracker *perf.Tracker
}

func (o *trackerRenderObserver) OnRender(compID string) {
	if compID == "__devtools" {
		return
	}
	o.tracker.RecordComponent(compID)
}
func (o *trackerRenderObserver) OnLayout(compID string) {
	if compID == "__devtools" {
		return
	}
	o.tracker.Record(perf.Layouts, 1)
}
func (o *trackerRenderObserver) OnPaint(compID string) {
	if compID == "__devtools" {
		return
	}
	o.tracker.Record(perf.Paints, 1)
}

// trackerEventObserver bridges event.EventObserver to perf.Tracker.
type trackerEventObserver struct {
	tracker *perf.Tracker
}

func (o *trackerEventObserver) OnEvent(eventType string, dispatched bool) {
	o.tracker.RecordEvent(eventType, dispatched)
}

// NewApp creates a new App with the given screen dimensions and output adapter.
func NewApp(w, h int, adapter output.Adapter) *App {
	painter := paint.NewPainter()
	t := perf.NewTracker(60)
	mgr := component.NewManager(painter)
	mgr.SetRenderObserver(&trackerRenderObserver{tracker: t})
	disp := event.NewDispatcher()
	disp.SetEventObserver(&trackerEventObserver{tracker: t})
	return &App{
		width:      w,
		height:     h,
		manager:    mgr,
		compositor: compositor.NewCompositor(w, h),
		dispatcher: disp,
		adapter:    adapter,
		tracker:    t,
		devtools:   devtools.NewPanel(t),
	}
}

// NewTestApp creates an App with a TestAdapter for testing.
func NewTestApp(w, h int) (*App, *output.TestAdapter) {
	ta := output.NewTestAdapter()
	app := NewApp(w, h, ta)
	return app, ta
}

// Tracker returns the performance tracker. Call Enable() to start recording.
func (a *App) Tracker() *perf.Tracker {
	return a.tracker
}

// RegisterComponent registers a component with the app.
// Creates buffer, sets initial rect, registers with manager.
func (a *App) RegisterComponent(id, name string, rect buffer.Rect, zIndex int, renderFn component.RenderFunc) *component.Component {
	comp := component.NewComponent(id, name, rect, zIndex, renderFn)
	a.manager.Register(comp)
	a.layersDirty = true
	a.tracker.Record(perf.ComponentsRegistered, 1)
	return comp
}

// UnregisterComponent removes a component.
func (a *App) UnregisterComponent(id string) {
	a.manager.Unregister(id)
	a.layersDirty = true
	a.tracker.Record(perf.ComponentsUnregistered, 1)
}

// SetState updates a component's state (marks it dirty).
func (a *App) SetState(compID string, key string, value any) {
	a.manager.SetState(compID, key, value)
	a.tracker.Record(perf.StateSets, 1)
}

// RenderAll renders all components and composes the full screen.
func (a *App) RenderAll() {
	a.tracker.BeginFrame()

	// Render all dirty components (they should already be marked dirty on register).
	a.manager.RenderDirty()

	// Build layers and set on compositor.
	layers := a.buildLayers()
	a.compositor.SetLayers(layers)
	a.tracker.Record(perf.OcclusionBuilds, 1)

	// Compose full screen.
	screen := a.compositor.ComposeAll()
	a.tracker.Record(perf.ComposeFull, 1)

	// Rebuild hit tester and sync handlers/focusables for event dispatch.
	a.rebuildHitTester()
	a.tracker.Record(perf.HitTesterRebuilds, 1)
	a.syncHandlers()
	a.tracker.Record(perf.HandlerFullSyncs, 1)

	// Auto-focus first focusable if nothing is focused yet.
	a.autoFocusIfNeeded()

	// Output.
	_ = a.adapter.WriteFull(screen)
	a.tracker.Record(perf.WriteFullCalls, 1)
	_ = a.adapter.Flush()
	a.tracker.Record(perf.FlushCalls, 1)

	a.lastDirtyRects = nil
	a.layersDirty = false
	a.manager.ClearDirty()

	a.tracker.EndFrame()
}

// RenderDirty renders only dirty components and composes changed regions.
func (a *App) RenderDirty() {
	a.tracker.BeginFrame()

	// 1. Capture dirty lists BEFORE rendering (RenderDirty clears DirtyPaint).
	paintDirty := a.manager.GetDirtyPaint()
	rectChanged := a.manager.GetRectChanged()

	// 2. Render dirty components (clears DirtyPaint on each).
	a.manager.RenderDirty()

	var allDirtyRects []buffer.Rect

	// 3. Only rebuild layers + occlusion map when structure changed
	//    (component added/removed/moved/resized). For paint-only dirty,
	//    the existing occlusion map and hit tester are still valid.
	needsRebuild := len(rectChanged) > 0 || a.layersDirty

	if needsRebuild {
		layers := a.buildLayers()
		a.compositor.SetLayers(layers)
		a.tracker.Record(perf.OcclusionBuilds, 1)

		if len(rectChanged) > 0 {
			// Rect changed → recompose old + new rects.
			var rects []buffer.Rect
			for _, comp := range rectChanged {
				rects = append(rects, comp.PrevRect())
				rects = append(rects, comp.Rect())
			}
			allDirtyRects = append(allDirtyRects, a.compositor.ComposeRects(rects)...)
			a.tracker.Record(perf.ComposeRects, 1)
		}

		// Re-layout VNode trees for components that moved but weren't
		// re-rendered (position-only move). The VNode tree's absolute
		// coordinates (X, Y) are stale after a move; we must recompute
		// them so the hit tester finds VNodes at the new position.
		// This is cheap: just a tree walk, no renderFn/paint/buffer.
		for _, comp := range rectChanged {
			if comp.VNodeTree() != nil && !comp.IsDirtyPaint() {
				r := comp.Rect()
				layout.ComputeLayout(comp.VNodeTree(), r.X, r.Y, r.W, r.H)
				a.tracker.Record(perf.Layouts, 1)
			}
		}

		// Rebuild hit tester + sync handlers after structural change.
		a.rebuildHitTester()
		a.tracker.Record(perf.HitTesterRebuilds, 1)
		a.syncHandlers()
		a.tracker.Record(perf.HandlerFullSyncs, 1)
		a.layersDirty = false
	}

	// 4. Compose dirty layers (paint changes) — always needed.
	if len(paintDirty) > 0 {
		dirtyLayers := a.getDirtyLayers(paintDirty)
		if len(dirtyLayers) > 0 {
			// Update OcclusionMap incrementally for dirty regions only.
			// This is needed because cell content may have changed
			// (e.g., text grew from "9" to "10"), affecting which cells
			// are non-zero (opaque) vs zero (transparent).
			a.compositor.UpdateDirtyRegions(dirtyLayers)
			a.tracker.Record(perf.OcclusionUpdates, 1)
			rects := a.compositor.ComposeDirty(dirtyLayers)
			a.tracker.Record(perf.ComposeDirty, 1)
			allDirtyRects = append(allDirtyRects, rects...)
		}
	}

	// 5. Sync handlers when components were re-rendered (VNodeTree may have
	//    changed handlers/focusables), even without structural change.
	//    Only sync the dirty components — much cheaper than full syncHandlers.
	if !needsRebuild && len(paintDirty) > 0 {
		a.syncDirtyHandlers(paintDirty)
		a.tracker.Record(perf.HandlerDirtySyncs, 1)
	}

	// Auto-focus first focusable if nothing is focused yet.
	a.autoFocusIfNeeded()

	// 6. Output.
	if len(allDirtyRects) > 0 {
		a.tracker.Record(perf.DirtyRectsOut, len(allDirtyRects))
		_ = a.adapter.WriteDirty(a.compositor.Screen(), allDirtyRects)
		a.tracker.Record(perf.WriteDirtyCalls, 1)
		_ = a.adapter.Flush()
		a.tracker.Record(perf.FlushCalls, 1)
	}

	a.lastDirtyRects = allDirtyRects
	a.manager.ClearDirty()

	a.tracker.EndFrame()
}

// HandleEvent dispatches an input event through the event system.
// F12 and DevTools tab-switching keys are intercepted before normal dispatch.
// Input VNodes get built-in keyboard handling (text editing, cursor movement).
func (a *App) HandleEvent(e *event.Event) {
	if e.Type == "keydown" {
		if e.Key == "F12" {
			a.toggleDevTools()
			return
		}
		// Tab switching when devtools is visible.
		if a.devtools.Visible {
			switch e.Key {
			case "1":
				a.devtools.SetTab(devtools.TabElements)
				a.refreshDevTools()
				return
			case "2":
				a.devtools.SetTab(devtools.TabPerf)
				a.refreshDevTools()
				return
			}
		}
		// Built-in input handling: if focused VNode is type="input",
		// handle text editing before normal dispatch.
		if a.handleInputKeyDown(e) {
			return
		}
	}
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
	oldRect := comp.Rect()
	comp.Move(newRect)
	if newRect.W != oldRect.W || newRect.H != oldRect.H {
		a.tracker.Record(perf.MovesWithResize, 1)
	} else {
		a.tracker.Record(perf.MovesPositionOnly, 1)
	}
}

// Resize resizes the screen.
func (a *App) Resize(w, h int) {
	a.width = w
	a.height = h
	a.compositor = compositor.NewCompositor(w, h)
	a.layersDirty = true
	// Mark all components dirty so next render repaints everything.
	for _, comp := range a.manager.GetAll() {
		comp.MarkDirty()
	}
}

// DevTools returns the DevTools panel (for testing/inspection).
func (a *App) DevTools() *devtools.Panel {
	return a.devtools
}

// toggleDevTools shows or hides the DevTools panel.
func (a *App) toggleDevTools() {
	a.devtools.Toggle()
	if a.devtools.Visible {
		// Enable perf tracking when devtools opens.
		a.tracker.Enable()

		panelH := a.height * 4 / 10
		if panelH < 8 {
			panelH = 8
		}
		a.devtools.Width = a.width
		a.devtools.Height = panelH

		a.updateDevToolsComponentInfo()

		rect := buffer.Rect{X: 0, Y: a.height - panelH, W: a.width, H: panelH}
		a.RegisterComponent("__devtools", "DevTools", rect, 9999, a.devtools.Render)
	} else {
		a.UnregisterComponent("__devtools")
	}
	a.RenderAll()
}

// tickDevTools is called every frame tick from the event loop.
// It updates FPS, snapshots perf data, and marks the devtools component dirty
// so it auto-refreshes each frame with up-to-date stats.
func (a *App) tickDevTools() {
	a.devtools.TickFPS()
	if !a.devtools.Visible {
		return
	}
	// Snapshot perf data BEFORE marking devtools dirty, so the snapshot
	// captures application stats without the devtools' own render cycle.
	a.devtools.SnapshotPerf()
	a.updateDevToolsComponentInfo()
	// Mark dirty so RenderDirty (called right after) will re-render it.
	comp := a.manager.Get("__devtools")
	if comp != nil {
		comp.MarkDirty()
	}
}

// refreshDevTools triggers an immediate re-render of the DevTools component.
// Used for tab switching where we want instant feedback without waiting for tick.
func (a *App) refreshDevTools() {
	a.devtools.SnapshotPerf()
	a.updateDevToolsComponentInfo()
	a.SetState("__devtools", "__refresh", time.Now().UnixNano())
	a.RenderDirty()
}

// updateDevToolsComponentInfo snapshots all registered components for the
// Elements tab, excluding the devtools component itself.
func (a *App) updateDevToolsComponentInfo() {
	all := a.manager.GetAll()
	infos := make([]devtools.ComponentInfo, 0, len(all))
	for _, comp := range all {
		if comp.ID() == "__devtools" {
			continue
		}
		r := comp.Rect()
		infos = append(infos, devtools.ComponentInfo{
			ID:        comp.ID(),
			Name:      comp.Name(),
			X:         r.X,
			Y:         r.Y,
			W:         r.W,
			H:         r.H,
			ZIndex:    comp.ZIndex(),
			VNodeTree: comp.VNodeTree(),
		})
	}
	a.devtools.UpdateComponents(infos)
}

// autoFocusIfNeeded focuses the first focusable VNode if nothing is currently
// focused. This ensures keyboard-driven components work immediately after
// their first render without requiring a manual Tab press.
func (a *App) autoFocusIfNeeded() {
	if a.dispatcher.FocusedID() == "" && a.dispatcher.HasFocusables() {
		a.dispatcher.FocusNext()
	}
}

// --- internal helpers ---

// compToLayer converts a component to a compositor layer.
func compToLayer(comp *component.Component) *compositor.Layer {
	return &compositor.Layer{
		ID:     comp.ID(),
		Buffer: comp.Buffer(),
		Rect:   comp.Rect(),
		ZIndex: comp.ZIndex(),
	}
}

// compToComponentLayer converts a component to an event.ComponentLayer.
func compToComponentLayer(comp *component.Component) *event.ComponentLayer {
	return &event.ComponentLayer{
		Layer:     compToLayer(comp),
		VNodeTree: comp.VNodeTree(),
	}
}

// buildLayers extracts compositor layers from all registered components.
func (a *App) buildLayers() []*compositor.Layer {
	all := a.manager.GetAll()
	layers := make([]*compositor.Layer, len(all))
	for i, comp := range all {
		layers[i] = compToLayer(comp)
	}
	return layers
}

// rebuildHitTester rebuilds the VNode hit tester from current component layers.
func (a *App) rebuildHitTester() {
	all := a.manager.GetAll()
	compLayers := make([]*event.ComponentLayer, len(all))
	for i, comp := range all {
		compLayers[i] = compToComponentLayer(comp)
	}
	ht := event.NewVNodeHitTester(compLayers, a.compositor.OcclusionMap())
	a.dispatcher.SetHitTester(ht)
}

// convertHandlerMap converts component.HandlerMap (any values) to event.HandlerMap.
// Handles both Go event.EventHandler values and Lua registry refs (int).
// When a bridge is available, Lua refs are wrapped via bridge.WrapLuaHandler.
func (a *App) convertHandlerMap(chm component.HandlerMap) event.HandlerMap {
	ehm := make(event.HandlerMap, len(chm))
	for evtType, handler := range chm {
		if h, ok := handler.(event.EventHandler); ok {
			ehm[evtType] = h
		} else if a.bridge != nil {
			// Lua handlers are stored as int registry refs by LuaTableToVNode.
			if ref, ok := toLuaRef(handler); ok {
				ehm[evtType] = a.bridge.WrapLuaHandler(ref)
			}
		}
	}
	return ehm
}

// toLuaRef extracts a Lua registry reference from a handler value.
// Lua handlers are stored as int (from L.Ref) by LuaTableToVNode.
func toLuaRef(val any) (int, bool) {
	switch v := val.(type) {
	case int:
		if v > 0 {
			return v, true
		}
	case int64:
		if v > 0 {
			return int(v), true
		}
	}
	return 0, false
}

// syncHandlers syncs event handlers and focusables from all components
// into the dispatcher.
func (a *App) syncHandlers() {
	// Clear stale handlers and focusables from previous render cycle.
	a.dispatcher.ClearAllHandlers()
	a.dispatcher.ClearAllFocusables()

	all := a.manager.GetAll()

	// Build parent map and register handlers.
	parentMap := make(map[string]string)
	for _, comp := range all {
		if comp.VNodeTree() != nil {
			buildParentMap(comp.VNodeTree(), "", parentMap)
		}
	}
	a.dispatcher.SetParentMap(parentMap)

	// Register all handlers from all components.
	for _, comp := range all {
		for vnodeID, chm := range comp.Handlers() {
			ehm := a.convertHandlerMap(chm)
			if len(ehm) > 0 {
				a.dispatcher.RegisterHandlers(vnodeID, ehm)
			}
		}
		for i, fID := range comp.Focusables() {
			a.dispatcher.RegisterFocusable(fID, i)
		}
	}
}

// syncDirtyHandlers updates handlers and focusables only for the given dirty
// components, without clearing and re-registering everything. This is O(dirty)
// instead of O(all_components).
func (a *App) syncDirtyHandlers(dirtyComps []*component.Component) {
	for _, comp := range dirtyComps {
		// Remove old handlers for this component's VNode IDs.
		// The dispatcher doesn't track per-component handler sets, so we
		// re-register from the component's extracted handler map.
		// Since ExtractHandlers was called during RenderDirty, comp.Handlers
		// is already up-to-date.
		for vnodeID, chm := range comp.Handlers() {
			ehm := a.convertHandlerMap(chm)
			if len(ehm) > 0 {
				a.dispatcher.RegisterHandlers(vnodeID, ehm)
			}
		}
		// Re-register focusables.
		for i, fID := range comp.Focusables() {
			a.dispatcher.RegisterFocusable(fID, i)
		}
		// Rebuild parent map for this component's VNode tree.
		if comp.VNodeTree() != nil {
			parentMap := make(map[string]string)
			buildParentMap(comp.VNodeTree(), "", parentMap)
			// Merge into dispatcher's parent map.
			a.dispatcher.MergeParentMap(parentMap)
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
		layers = append(layers, compToLayer(comp))
	}
	return layers
}
