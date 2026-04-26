package lumina

import (
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

func (app *App) renderAllDirty() {
	// Frame rate limiting: skip if less than 16ms since last render
	now := time.Now()
	if !app.lastRenderTime.IsZero() && now.Sub(app.lastRenderTime) < 16*time.Millisecond {
		return // will catch on next tick
	}

	// Check for any root that needs work (self-dirty OR has dirty children)
	hasDirty := false
	for _, comp := range globalRegistry.components {
		if comp.IsRoot && (comp.Dirty.Load() || comp.HasDirtyChild.Load()) {
			hasDirty = true
			break
		}
	}
	if !hasDirty {
		return
	}

	// FPS tracking
	app.frameCount++
	if now.Sub(app.fpsLastTime) >= time.Second {
		app.fps = app.frameCount
		app.frameCount = 0
		app.fpsLastTime = now
	}

	// Collect root components that need work
	var roots []*Component
	for _, comp := range globalRegistry.components {
		if comp.IsRoot {
			roots = append(roots, comp)
		}
	}

	adapter := GetOutputAdapter()
	if adapter == nil {
		return
	}

	// Pause Lua GC during entire render pass
	app.L.SetGCStopped(true)

	for _, comp := range roots {
		if comp.Dirty.Load() {
			// Root itself is dirty — full re-render (runs root's Lua PCall)
			app.renderComponent(comp, adapter)
		} else if comp.HasDirtyChild.Load() {
			// Root is clean but has dirty children — partial re-render:
			// reuse root's VNode tree, only re-render dirty child components.
			app.renderDirtyChildren(comp, adapter)
		}
	}

	// Resume GC — let the debt-based pacer trigger collection naturally
	app.L.SetGCStopped(false)
}

// renderDirtyChildren handles the case where a root component is NOT self-dirty
// but has dirty descendants. It reuses the root's cached VNode tree, walks it
// to re-render only dirty child components in-place, then runs the normal
// diff/layout/write pipeline. This avoids the root's Lua PCall which would
// create 1800+ createElement tables for unchanged children.
func (app *App) renderDirtyChildren(comp *Component, adapter OutputAdapter) {
	comp.HasDirtyChild.Store(false)

	if comp.LastVNode == nil {
		// No cached tree — fall back to full render
		comp.Dirty.Store(true)
		app.renderComponent(comp, adapter)
		return
	}

	// Walk the VNode tree and re-render dirty child components in-place.
	// This modifies comp.LastVNode's subtrees where children are dirty.
	changed := reRenderDirtySubtree(app.L, comp.LastVNode)
	if !changed {
		return // nothing actually changed
	}

	// Now run the normal diff/layout/write pipeline with the updated VNode tree.
	// The tree is the same object (comp.LastVNode) with dirty subtrees replaced.
	newVNode := comp.LastVNode

	w, h := app.getWidth(), app.getHeight()
	sizeChanged := (w != app.lastRenderWidth) || (h != app.lastRenderHeight)

	var frame *Frame
	if sizeChanged || app.lastFrame == nil {
		// Size changed or no previous frame — must do full layout + paint
		frame = VNodeToFrame(newVNode, w, h)
		app.lastFrame = frame
		app.lastRenderWidth = w
		app.lastRenderHeight = h
		// Size change needs full event re-bridging
		app.bridgeVNodeEvents(newVNode)
	} else {
		// INCREMENTAL: size unchanged, reuse existing frame.
		// Skip computeFlexLayout — layout positions are cached from last full render.
		// VNode positions (X, Y, W, H) were set by the last computeFlexLayout call
		// and haven't changed (hover only changes colors, not sizes).
		// Just repaint all VNodes into the existing frame.
		frame = app.lastFrame
		fullClip := Rect{X: 0, Y: 0, W: w, H: h}
		renderVNode(frame, newVNode, fullClip)
		frame.MarkDirty()
		// SKIP bridgeVNodeEvents — handlers still valid from last full render.
	}

	// Clear scroll dirty flags
	ClearAllScrollDirty()

	// Composite overlays
	overlays := globalOverlayManager.GetVisible()
	if len(overlays) > 0 {
		compositor := NewCompositor(w, h)
		frame = compositor.Compose(frame, overlays)
	}

	// Composite managed windows
	windows := globalWindowManager.GetVisible()
	if len(windows) > 0 {
		compositor := NewCompositor(w, h)
		var winOverlays []*Overlay
		for _, win := range windows {
			winVNode := BuildWindowVNode(win)
			winOverlays = append(winOverlays, &Overlay{
				ID: "window-" + win.ID, VNode: winVNode,
				X: win.X, Y: win.Y, W: win.W, H: win.H,
				ZIndex: win.ZIndex, Visible: true,
			})
		}
		frame = compositor.Compose(frame, winOverlays)
	}

	frame.FocusedID = globalEventBus.GetFocused()

	// DevTools inspector overlay
	if IsInspectorVisible() && app.lastFrame != nil {
		var highlightNode *VNode
		if globalInspector.selectedID != "" {
			highlightNode = findVNodeByID(newVNode, globalInspector.selectedID)
		} else if globalInspector.highlightID != "" {
			highlightNode = findVNodeByID(newVNode, globalInspector.highlightID)
		}
		if highlightNode != nil {
			RenderHighlight(frame, highlightNode)
		}
		CallDevToolsRender(app.L)
		dtOverlay := globalOverlayManager.Get("devtools-panel")
		if dtOverlay != nil {
			dtCompositor := NewCompositor(w, h)
			frame = dtCompositor.Compose(frame, []*Overlay{dtOverlay})
		}
	}

	adapter.Write(frame)
	app.lastRenderTime = time.Now()
}

// reRenderDirtySubtree walks a VNode tree and re-renders any child component
// whose Dirty flag is set. Returns true if any subtree was changed.
func reRenderDirtySubtree(L *lua.State, vnode *VNode) bool {
	if vnode == nil {
		return false
	}

	changed := false

	// If this VNode is backed by a dirty component, re-render it
	if vnode.ComponentRef != nil && vnode.ComponentRef.Dirty.Load() {
		comp := vnode.ComponentRef
		comp.MarkClean()
		comp.HasDirtyChild.Store(false)

		prevComp := GetCurrentComponent()
		SetCurrentComponent(comp)
		comp.ResetHookIndex()

		if comp.PushRenderFn(L) {
			// Build instance table: state + props (same as luaComponentToVNode)
			fields := map[string]any{
				"_instance": comp.ID,
			}
			for k, v := range comp.State {
				fields[k] = v
			}
			fields["props"] = comp.Props

			L.NewTableFrom(fields)
			// Expand function refs from props onto instance table
			for k, v := range comp.Props {
				if k != "children" {
					if ref, ok := v.(LuaFuncRef); ok {
						L.RawGetI(lua.RegistryIndex, int64(ref.Ref))
						L.SetField(-2, k)
					}
				}
			}

			status := L.PCall(1, 1, 0)
			if status == lua.OK {
				newChild := LuaVNodeToVNode(L, -1)
				L.Pop(1)

				// Replace VNode content in-place (keep the same pointer so
				// parent references remain valid)
				newChild.ComponentRef = comp
				newChild.ComponentKey = vnode.ComponentKey
				comp.LastVNode = newChild

				// Copy new VNode data into existing vnode (in-place update)
				vnode.Type = newChild.Type
				vnode.Props = newChild.Props
				vnode.Children = newChild.Children
				vnode.Content = newChild.Content
				vnode.Style = newChild.Style
				vnode.ComponentRef = newChild.ComponentRef
				vnode.ComponentKey = newChild.ComponentKey
				changed = true
			} else {
				L.Pop(1) // pop error
			}
		}

		SetCurrentComponent(prevComp)
		// Don't recurse into freshly rendered children — they're already up to date
		return changed
	}

	// If this component has dirty children, recurse
	if vnode.ComponentRef != nil && vnode.ComponentRef.HasDirtyChild.Load() {
		vnode.ComponentRef.HasDirtyChild.Store(false)
	}

	// Recurse into children
	for _, child := range vnode.Children {
		if reRenderDirtySubtree(L, child) {
			changed = true
		}
	}

	return changed
}

// renderComponent re-renders a single component on the main thread.
func (app *App) renderComponent(comp *Component, adapter OutputAdapter) {
	SetCurrentComponent(comp)
	comp.ResetHookIndex()
	defer SetCurrentComponent(nil)

	// Cache dimensions locally (atomic reads)
	w, h := app.getWidth(), app.getHeight()

	if !comp.PushRenderFn(app.L) {
		return
	}

	status := app.L.PCall(0, 1, 0)
	if status != lua.OK {
		app.L.Pop(1)
		comp.MarkClean() // prevent infinite retry on persistent errors
		return
	}

	newVNode := LuaVNodeToVNode(app.L, -1)
	app.L.Pop(1)

	// Diff against previous render to detect no-change case.
	var frame *Frame
	sizeChanged := (w != app.lastRenderWidth) || (h != app.lastRenderHeight)
	inspectorDirty := needsInspectorRerender.Load()
	needsInspectorRerender.Store(false)
	if comp.LastVNode != nil {
		patches := DiffVNode(comp.LastVNode, newVNode)
		scrollDirty := AnyViewportScrollDirty()
		if len(patches) == 0 && !scrollDirty && !sizeChanged && !inspectorDirty {
			// Nothing changed — skip rendering.
			comp.MarkClean()
			ClearAllScrollDirty()
			return
		}

		// Incremental rendering: if few patches and we have a previous frame,
		// re-layout the new tree and apply only changed subtrees via parent
		// container re-rendering (handles sibling position shifts correctly).
		// Scroll-dirty or size-change forces full re-render since layout positions change.
		// Also need full re-render if inspector visibility changed.
		if len(patches) <= 10 && app.lastFrame != nil && !ShouldFullRerender(patches, newVNode) && !scrollDirty && !sizeChanged && !inspectorDirty {
			frame = app.lastFrame
			// Re-layout the entire new tree (cheap) so positions are correct
			computeFlexLayout(newVNode, 0, 0, w, h)
			ApplyPatches(frame, newVNode, patches, w, h)
			// Reconcile components: cleanup any removed in incremental update
			ReconcileComponents(app.L, comp.LastVNode, newVNode)
			comp.LastVNode = newVNode
			app.lastFrame = frame
			app.lastRenderWidth = w
			app.lastRenderHeight = h
			goto compositeAndWrite
		}
	}
	// Full re-render (first render, large change, or no previous frame)
	frame = VNodeToFrame(newVNode, w, h)

	// Reconcile components: cleanup any that were in old tree but not in new
	if comp.LastVNode != nil {
		ReconcileComponents(app.L, comp.LastVNode, newVNode)
	}

	comp.LastVNode = newVNode
	app.lastFrame = frame
	app.lastRenderWidth = w
	app.lastRenderHeight = h

compositeAndWrite:
	// Clear scroll dirty flags after re-render (layout applied new scroll positions)
	ClearAllScrollDirty()
	// Bridge VNode event handlers to EventBus
	app.bridgeVNodeEvents(newVNode)

	// Composite overlays on top of the base frame using the layer compositor
	overlays := globalOverlayManager.GetVisible()
	if len(overlays) > 0 {
		compositor := NewCompositor(w, h)
		frame = compositor.Compose(frame, overlays)
	}

	// Composite managed windows on top of overlays
	windows := globalWindowManager.GetVisible()
	if len(windows) > 0 {
		compositor := NewCompositor(w, h)
		var winOverlays []*Overlay
		for _, win := range windows {
			winVNode := BuildWindowVNode(win)
			winOverlays = append(winOverlays, &Overlay{
				ID:      "window-" + win.ID,
				VNode:   winVNode,
				X:       win.X,
				Y:       win.Y,
				W:       win.W,
				H:       win.H,
				ZIndex:  win.ZIndex,
				Visible: true,
			})
		}
		frame = compositor.Compose(frame, winOverlays)
	}

	frame.FocusedID = globalEventBus.GetFocused()

	// DevTools inspector overlay
	if IsInspectorVisible() && app.lastFrame != nil {
		// Highlight the hovered/selected element
		var highlightNode *VNode
		if globalInspector.selectedID != "" {
			highlightNode = findVNodeByID(newVNode, globalInspector.selectedID)
		} else if globalInspector.highlightID != "" {
			highlightNode = findVNodeByID(newVNode, globalInspector.highlightID)
		}
		if highlightNode != nil {
			RenderHighlight(frame, highlightNode)
		}

		// Lua DevTools panel renders via overlay system (showOverlay).
		// Call Lua to rebuild the overlay VNode on each render cycle.
		CallDevToolsRender(app.L)

		// The overlay is now managed by globalOverlayManager via showOverlay().
		// Compose it along with any other overlays.
		dtOverlay := globalOverlayManager.Get("devtools-panel")
		if dtOverlay != nil {
			dtCompositor := NewCompositor(w, h)
			frame = dtCompositor.Compose(frame, []*Overlay{dtOverlay})
		}
	} else if app.lastFrame != nil && needsInspectorRerender.Load() {
		// Inspector was just hidden — Lua hides the overlay
		CallDevToolsRender(app.L)
		// Also clear the inspector area in case overlay cleanup is needed
		panelW := globalInspector.panelWidth
		if panelW > w/2 {
			panelW = w / 2
		}
		startX := w - panelW
		if startX < 0 {
			startX = 0
		}
		// Clear cells in the inspector area to force re-render
		for y := 0; y < h && y < frame.Height; y++ {
			for x := startX; x < w && x < frame.Width; x++ {
				frame.Cells[y][x] = Cell{}
			}
		}
	}

	adapter.Write(frame)
	app.lastRenderTime = time.Now()
	comp.MarkClean()
}

// InitialRender renders all components once (for testing/compatibility).
// InitialRender renders all components once (for testing/compatibility).
func (app *App) InitialRender() {
	components := make([]*Component, 0, len(globalRegistry.components))
	for _, comp := range globalRegistry.components {
		components = append(components, comp)
	}

	adapter := GetOutputAdapter()
	if adapter == nil {
		return
	}

	for _, comp := range components {
		SetCurrentComponent(comp)

		if !comp.PushRenderFn(app.L) {
			continue
		}

		status := app.L.PCall(0, 1, 0)
		if status != lua.OK {
			app.L.Pop(1)
			continue
		}

		frame := RenderLuaVNode(app.L, -1, app.getWidth(), app.getHeight())
		app.L.Pop(1)
		frame.FocusedID = globalEventBus.GetFocused()
		adapter.Write(frame)
	}
	SetCurrentComponent(nil)
}

// Stop stops the application by posting a quit event.
// RenderOnce renders all dirty components once and returns.
// Useful for headless testing — render a single frame without an event loop.
func (app *App) RenderOnce() {
	app.renderAllDirty()
}

// Close closes the application and cleans up resources.
// HitTestVNode finds the deepest VNode containing point (px, py).
// Returns the VNode's ID (from props["id"]) or "" if no match.
func HitTestVNode(vnode *VNode, px, py int) string {
	if vnode == nil {
		return ""
	}
	// Check if point is within this node's bounds
	if px < vnode.X || px >= vnode.X+vnode.W || py < vnode.Y || py >= vnode.Y+vnode.H {
		return ""
	}
	// Check children (deepest match wins — reverse order for z-order)
	for i := len(vnode.Children) - 1; i >= 0; i-- {
		if id := HitTestVNode(vnode.Children[i], px, py); id != "" {
			return id
		}
	}
	// Return this node's ID if it has one
	if id, ok := vnode.Props["id"].(string); ok && id != "" {
		return id
	}
	return ""
}

// findRootVNode returns the last rendered VNode tree from the root component.
// findRootVNode returns the last rendered VNode tree from the root component.
func (app *App) findRootVNode() *VNode {
	for _, comp := range globalRegistry.components {
		if comp.LastVNode != nil {
			return comp.LastVNode
		}
	}
	return nil
}

// GetGlobalEventBus returns the global event bus (for testing).
