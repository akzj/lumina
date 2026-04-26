package lumina

import (
	"time"
)

// Event pipeline: ordered stages for processing input events.
// Extracted from handleEvent's input_event case for clarity and
// to prevent bugs where multiple handlers fire for the same event.

// inputPipelineCtx carries state between pipeline stages.
type inputPipelineCtx struct {
	e           *Event
	textHandled bool // set by stageTextInput, read by stageKeyboardNav
}

// runInputPipeline processes an input event through ordered stages.
// Each stage can mark the event consumed by returning true.
// When a stage returns consumed=true, no further stages run.
func (app *App) runInputPipeline(e *Event) {
	ctx := &inputPipelineCtx{e: e}

	for _, stage := range []func(*App, *inputPipelineCtx) bool{
		stageMouseThrottle,   // Drop high-frequency mousemove events (16ms throttle)
		stageQuitShortcut,    // Ctrl+C/Q → quit
		stageDevTools,        // F12 toggle + inspector mouse hover/selection
		stageModalIntercept,  // Modal: Escape closes, others retarget
		stageHitTest,         // Mouse: Cell.OwnerNode → e.Target, e.LocalX/Y
		stageFocus,           // mousedown: SetFocus(e.Target)
		stageWindowManager,   // Window chrome: close/max/min/drag/resize (CONSUMES)
		stageDragAndDrop,     // DnD: start/move/end
		stageHover,           // mousemove: synthesize mouseenter/mouseleave
		stageEmit,            // EventBus.Emit(e) — main dispatch to bridged handlers
		stageClickSynthesize, // mousedown: synthesize click + contextmenu
		stageTextInput,       // keydown: handleTextInputEvent → sets ctx.textHandled
		stageKeyboardNav,     // keydown: HandleKeyEvent (if !textHandled && !hasBinding)
		stageOnKeyBinding,    // keydown: dispatchKeyBinding (lumina.onKey)
		stageScroll,          // scroll/keydown: handleScrollEvent
	} {
		if stage(app, ctx) {
			return // consumed
		}
	}
}

// markAllComponentsDirty marks all registered components as needing re-render.
func markAllComponentsDirty() {
	for _, comp := range globalRegistry.components {
		comp.Dirty.Store(true)
	}
}

// stageMouseThrottle drops duplicate mousemove events at the same coordinates.
// Terminal ANY_EVENT mouse tracking (mode 1003) sends continuous mousemove
// reports even when the mouse is stationary. We drop these duplicates
// and also throttle same-position events to at most once per 16ms.
// Events with changed coordinates always pass through immediately.
func stageMouseThrottle(app *App, ctx *inputPipelineCtx) bool {
	if ctx.e.Type != "mousemove" {
		return false
	}
	e := ctx.e
	now := time.Now()

	// If coordinates changed, always process
	if e.X != app.lastMouseX || e.Y != app.lastMouseY {
		app.lastMouseX = e.X
		app.lastMouseY = e.Y
		app.lastMouseMoveTime = now
		return false
	}

	// Same coordinates — throttle to 16ms
	if !app.lastMouseMoveTime.IsZero() && now.Sub(app.lastMouseMoveTime) < 16*time.Millisecond {
		return true // consumed (dropped)
	}
	app.lastMouseMoveTime = now
	return false
}

// stageQuitShortcut handles Ctrl+C / Ctrl+Q to quit (always works, even with modal).
func stageQuitShortcut(app *App, ctx *inputPipelineCtx) bool {
	e := ctx.e
	if e.Type == "keydown" && e.Modifiers.Ctrl && (e.Key == "c" || e.Key == "q") {
		app.running = false
		return true
	}
	return false
}

// stageDevTools handles F12 toggle and inspector mouse hover/selection.
func stageDevTools(app *App, ctx *inputPipelineCtx) bool {
	e := ctx.e

	// F12 toggles DevTools inspector
	if e.Type == "keydown" && e.Key == "F12" {
		ToggleInspector()
		// Use Lua-based DevTools panel
		CallDevToolsRender(app.L)
		// Force re-render
		markAllComponentsDirty()
	}

	// DevTools inspector: update hover/selection on mouse events
	if IsInspectorVisible() {
		switch e.Type {
		case "mousemove":
			if app.lastFrame != nil && e.X >= 0 && e.Y >= 0 &&
				e.X < app.lastFrame.Width && e.Y < app.lastFrame.Height {
				node := app.lastFrame.Cells[e.Y][e.X].OwnerNode
				if node != nil {
					if id, ok := node.Props["id"].(string); ok && id != "" {
						SetInspectorHighlight(id)
						// Trigger Lua DevTools re-render after highlight change
						CallDevToolsRender(app.L)
					}
				}
			}
		case "mousedown":
			// Click selects element for inspection
			if app.lastFrame != nil && e.X >= 0 && e.Y >= 0 &&
				e.X < app.lastFrame.Width && e.Y < app.lastFrame.Height {
				node := app.lastFrame.Cells[e.Y][e.X].OwnerNode
				if node != nil {
					if id, ok := node.Props["id"].(string); ok && id != "" {
						SetInspectorSelected(id)
						// Trigger Lua DevTools re-render after selection change
						CallDevToolsRender(app.L)
					}
				}
			}
		}
	}

	return false
}

// stageModalIntercept handles modal overlay input routing.
// Escape closes modal, other events are retargeted to modal.
func stageModalIntercept(app *App, ctx *inputPipelineCtx) bool {
	e := ctx.e
	if topModal := globalOverlayManager.GetTopModal(); topModal != nil {
		if e.Type == "keydown" && e.Key == KeyEscape {
			globalOverlayManager.Hide(topModal.ID)
			return true
		}
		// Route event to modal's VNode tree if it has one
		// (target events to the modal overlay, not the base layer)
		if topModal.VNode != nil {
			e.Target = topModal.ID
		}
	}
	return false
}

// stageHitTest performs mouse event hit-testing: maps (x,y) to target component.
func stageHitTest(app *App, ctx *inputPipelineCtx) bool {
	e := ctx.e
	if e.Type != "mousedown" && e.Type != "mouseup" && e.Type != "mousemove" {
		return false
	}

	// O(1) hit-test using Cell.OwnerNode from last rendered frame
	var targetNode *VNode
	if app.lastFrame != nil && e.X >= 0 && e.Y >= 0 &&
		e.X < app.lastFrame.Width && e.Y < app.lastFrame.Height {
		targetNode = app.lastFrame.Cells[e.Y][e.X].OwnerNode
	}

	if targetNode != nil {
		if id, ok := targetNode.Props["id"].(string); ok && id != "" {
			e.Target = id
		} else {
			// Walk up the VNode tree to find nearest ancestor with an id
			tree := globalEventBus.GetVNodeTree()
			if tree != nil {
				for node := tree.Parents[targetNode]; node != nil; node = tree.Parents[node] {
					if id, ok := node.Props["id"].(string); ok && id != "" {
						e.Target = id
						targetNode = node
						break
					}
				}
			}
		}
		e.TargetNode = targetNode
		// Compute local coordinates (relative to target VNode top-left)
		e.LocalX = e.X - targetNode.X
		e.LocalY = e.Y - targetNode.Y
	}

	// Fall back to VNode tree walk if Cell has no owner
	if e.Target == "" {
		if root := app.findRootVNode(); root != nil {
			targetID := HitTestVNode(root, e.X, e.Y)
			if targetID != "" {
				e.Target = targetID
			}
		}
	}

	return false
}

// stageFocus sets focus on mousedown target.
func stageFocus(app *App, ctx *inputPipelineCtx) bool {
	e := ctx.e
	if e.Type == "mousedown" && e.Target != "" {
		globalEventBus.SetFocus(e.Target)
	}
	return false
}

// stageWindowManager handles window chrome clicks (title bar / controls)
// and mouse drag/resize operations. These CONSUME the event.
func stageWindowManager(app *App, ctx *inputPipelineCtx) bool {
	e := ctx.e
	if e.Type != "mousedown" && e.Type != "mousemove" && e.Type != "mouseup" {
		return false
	}

	if win := globalWindowManager.WindowAtPoint(e.X, e.Y); win != nil {
		localY := e.Y - win.Y
		localX := e.X - win.X

		switch e.Type {
		case "mousedown":
			// Title bar click (row 0)
			if localY == 0 {
				// Check which control button was clicked
				// Button region is right-aligned: [ _ ][ □ ][ X ] = 9 chars
				if localX >= win.W-9 {
					// Close button — last 3 chars " X "
					globalWindowManager.CloseWindow(win.ID)
					return true // consume event, don't emit to bus
				} else if localX >= win.W-13 {
					// Maximize button — next 3 chars " □ "
					if win.Maximized {
						globalWindowManager.RestoreWindow(win.ID)
					} else {
						globalWindowManager.MaximizeWindow(win.ID)
					}
					return true
				} else if localX >= win.W-16 {
					// Minimize button — next 3 chars " _ "
					globalWindowManager.MinimizeWindow(win.ID)
					return true
				}
				// Otherwise: start drag
				globalWindowManager.StartDrag(win.ID, localX, localY)
				globalWindowManager.FocusWindow(win.ID)
				return true
			}
			// Resize handle (bottom-right corner)
			if localY == win.H-1 && localX == win.W-1 {
				globalWindowManager.StartResize(win.ID, e.X, e.Y)
				return true
			}

		case "mousemove":
			if globalWindowManager.IsDragging() {
				globalWindowManager.UpdateDrag(e.X, e.Y)
				return true
			}
			if globalWindowManager.IsResizing() {
				globalWindowManager.UpdateResize(e.X, e.Y)
				return true
			}

		case "mouseup":
			if globalWindowManager.IsDragging() {
				globalWindowManager.StopDrag()
				return true
			}
			if globalWindowManager.IsResizing() {
				globalWindowManager.StopResize()
				return true
			}
		}
	}

	return false
}

// stageDragAndDrop handles drag-and-drop: start on mousedown, move on mousemove, drop on mouseup.
// Does NOT consume — falls through to other stages.
func stageDragAndDrop(app *App, ctx *inputPipelineCtx) bool {
	e := ctx.e
	switch e.Type {
	case "mousedown":
		if e.TargetNode != nil {
			if draggable, ok := e.TargetNode.Props["draggable"].(bool); ok && draggable {
				dragType, _ := e.TargetNode.Props["dragType"].(string)
				globalDragState.StartDrag(e.Target, dragType, e.TargetNode.Props["dragData"])
				globalDragState.UpdatePosition(e.X, e.Y)
			}
		}

	case "mousemove":
		if globalDragState.Dragging() {
			globalDragState.UpdatePosition(e.X, e.Y)
			// Check if hovering over a drop target
			if e.TargetNode != nil {
				if _, hasOnDrop := e.TargetNode.Props["onDrop"]; hasOnDrop {
					globalDragState.SetDropTarget(e.Target)
					globalEventBus.Emit(&Event{
						Type: "dragover", Target: e.Target, Bubbles: true,
						X: e.X, Y: e.Y, LocalX: e.LocalX, LocalY: e.LocalY,
						Timestamp: e.Timestamp,
					})
				}
			}
		}

	case "mouseup":
		if globalDragState.Dragging() {
			sourceID, dropTargetID, _ := globalDragState.EndDrag()
			if dropTargetID != "" {
				globalEventBus.Emit(&Event{
					Type: "drop", Target: dropTargetID, Bubbles: true,
					X: e.X, Y: e.Y,
					Timestamp: e.Timestamp,
				})
			}
			_ = sourceID // available for future use
		}
	}
	return false
}

// stageHover handles mousemove hover tracking: synthesizes mouseenter/mouseleave events.
func stageHover(app *App, ctx *inputPipelineCtx) bool {
	e := ctx.e
	if e.Type != "mousemove" {
		return false
	}

	newHoverID := e.Target
	if newHoverID != app.hoveredID {
		oldHoverID := app.hoveredID
		app.hoveredID = newHoverID

		// Synthesize mouseleave for old element
		if oldHoverID != "" {
			globalEventBus.Emit(&Event{
				Type:      "mouseleave",
				Target:    oldHoverID,
				Bubbles:   false,
				X:         e.X,
				Y:         e.Y,
				Timestamp: e.Timestamp,
			})
		}

		// Synthesize mouseenter for new element
		if newHoverID != "" {
			globalEventBus.Emit(&Event{
				Type:      "mouseenter",
				Target:    newHoverID,
				Bubbles:   false,
				X:         e.X,
				Y:         e.Y,
				Timestamp: e.Timestamp,
			})
		}

		// NOTE: No markAllComponentsDirty() here.
		// The bridged mouseenter/mouseleave handlers already mark
		// affected components dirty via setState/store.dispatch.
		// Marking ALL components dirty caused O(N) re-renders on
		// every hover change, leading to 80% CPU on fast mouse movement.
	}

	return false
}

// stageEmit dispatches the event to the EventBus (handles focus, shortcuts, registered handlers).
// Never consumes.
func stageEmit(app *App, ctx *inputPipelineCtx) bool {
	globalEventBus.Emit(ctx.e)
	return false
}

// stageClickSynthesize emits synthetic click and contextmenu events on mousedown.
// Never consumes.
func stageClickSynthesize(app *App, ctx *inputPipelineCtx) bool {
	e := ctx.e
	if e.Type != "mousedown" || e.Target == "" {
		return false
	}

	clickEvent := &Event{
		Type:       "click",
		Target:     e.Target,
		Bubbles:    true,
		X:          e.X,
		Y:          e.Y,
		LocalX:     e.LocalX,
		LocalY:     e.LocalY,
		Button:     e.Button,
		Modifiers:  e.Modifiers,
		Timestamp:  e.Timestamp,
		TargetNode: e.TargetNode,
	}
	globalEventBus.Emit(clickEvent)

	// Right-click → emit contextmenu
	if e.Button == "right" {
		globalEventBus.Emit(&Event{
			Type:      "contextmenu",
			Target:    e.Target,
			Bubbles:   true,
			X:         e.X,
			Y:         e.Y,
			Timestamp: e.Timestamp,
		})
	}

	return false
}

// stageTextInput handles text input events for focused input/textarea elements.
// Sets ctx.textHandled for downstream stages. Never consumes.
func stageTextInput(app *App, ctx *inputPipelineCtx) bool {
	if ctx.e.Type == "keydown" {
		ctx.textHandled = app.handleTextInputEvent(ctx.e)
	}
	return false
}

// stageKeyboardNav handles built-in keyboard navigation (Tab, Enter, Escape, etc.).
// Skipped if text input consumed the event or if there's a user-defined onKey binding.
// Never consumes.
func stageKeyboardNav(app *App, ctx *inputPipelineCtx) bool {
	e := ctx.e
	if e.Type != "keydown" || ctx.textHandled {
		return false
	}

	normalized := normalizeKeyName(e.Key)
	keyBindingsMu.Lock()
	_, hasBinding := keyBindings[normalized]
	keyBindingsMu.Unlock()
	if !hasBinding {
		globalEventBus.HandleKeyEvent(e.Key, e.Modifiers)
	}

	return false
}

// stageOnKeyBinding dispatches lumina.onKey() bindings. Never consumes.
func stageOnKeyBinding(app *App, ctx *inputPipelineCtx) bool {
	if ctx.e.Type == "keydown" {
		app.dispatchKeyBinding(ctx.e.Key)
	}
	return false
}

// stageScroll handles scroll events (mouse wheel and PageUp/PageDown). Never consumes.
func stageScroll(app *App, ctx *inputPipelineCtx) bool {
	app.handleScrollEvent(ctx.e)
	return false
}
