package render

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// HitTest finds the deepest Node at screen coordinates (x, y).
// Returns nil if no node contains the point.
// Uses cumulative scroll offset for correct nested scroll handling.
func HitTest(root *Node, x, y int) *Node {
	return hitTestWithOffset(root, x, y, 0)
}

// hitTestWithOffset performs hit testing with a cumulative scroll offset.
// scrollOffsetY accumulates as we descend through nested scroll containers.
// Each node's screen position = layout position - cumulativeScrollOffset.
func hitTestWithOffset(node *Node, x, y int, scrollOffsetY int) *Node {
	if node == nil {
		return nil
	}

	// display:none and visibility:hidden nodes are not interactive
	if node.Style.Display == "none" || node.Style.Visibility == "hidden" {
		return nil
	}

	// Component placeholders are transparent containers — their bounds from
	// parent layout don't reflect absolute-positioned children's actual positions.
	// Skip bounds check and go straight to checking children.
	if node.Type == "component" {
		if sorted := hitTestOrderChildren(node.Children); sorted != nil {
			for _, child := range sorted {
				if hit := hitTestWithOffset(child, x, y, scrollOffsetY); hit != nil {
					return hit
				}
			}
		} else {
			for i := len(node.Children) - 1; i >= 0; i-- {
				if hit := hitTestWithOffset(node.Children[i], x, y, scrollOffsetY); hit != nil {
					return hit
				}
			}
		}
		return nil // component placeholder is transparent — no match
	}

	// Node's screen position = layout position - cumulative scroll offset
	screenX := node.X
	screenY := node.Y - scrollOffsetY

	// Bounds check using screen coordinates
	if x < screenX || x >= screenX+node.W || y < screenY || y >= screenY+node.H {
		return nil
	}

	// Calculate child scroll offset
	childScrollOffset := scrollOffsetY
	if node.Style.Overflow == "scroll" {
		childScrollOffset += node.ScrollY

		// Clip: only process clicks within content area (not border/padding)
		bw := 0
		if node.Style.Border != "" && node.Style.Border != "none" {
			bw = 1
		}
		visTop := screenY + bw + node.Style.PaddingTop
		visBot := screenY + node.H - bw - node.Style.PaddingBottom
		if y < visTop || y >= visBot {
			return node // click on border/padding area, return container
		}
	}

	// Check children in reverse z-order (highest ZIndex / top-most first)
	hitChildren := hitTestOrderChildren(node.Children)
	if hitChildren != nil {
		// z-index sorting active: iterate forward through pre-sorted list
		for _, child := range hitChildren {
			// For scroll containers, skip children outside visible area
			if node.Style.Overflow == "scroll" {
				childScreenY := child.Y - childScrollOffset
				childScreenBottom := childScreenY + child.H
				bw := 0
				if node.Style.Border != "" && node.Style.Border != "none" {
					bw = 1
				}
				visTop := screenY + bw + node.Style.PaddingTop
				visBot := screenY + node.H - bw - node.Style.PaddingBottom
				if childScreenBottom <= visTop || childScreenY >= visBot {
					continue // scrolled out of view
				}
			}

			if hit := hitTestWithOffset(child, x, y, childScrollOffset); hit != nil {
				return hit
			}
		}
	} else {
		// Fast path: no z-index, reverse order (same as before)
		for i := len(node.Children) - 1; i >= 0; i-- {
			child := node.Children[i]

			// For scroll containers, skip children outside visible area
			if node.Style.Overflow == "scroll" {
				childScreenY := child.Y - childScrollOffset
				childScreenBottom := childScreenY + child.H
				bw := 0
				if node.Style.Border != "" && node.Style.Border != "none" {
					bw = 1
				}
				visTop := screenY + bw + node.Style.PaddingTop
				visBot := screenY + node.H - bw - node.Style.PaddingBottom
				if childScreenBottom <= visTop || childScreenY >= visBot {
					continue // scrolled out of view
				}
			}

			if hit := hitTestWithOffset(child, x, y, childScrollOffset); hit != nil {
				return hit
			}
		}
	}

	return node
}

// HitTestWithHandler finds the deepest Node at (x,y) that has a handler for the given event.
// First does a normal hit-test, then walks up the tree (bubbling) to find the nearest handler.
func HitTestWithHandler(root *Node, x, y int, eventType string) *Node {
	node := HitTest(root, x, y)
	if node == nil {
		return nil
	}

	// Walk up looking for a handler (event bubbling)
	for n := node; n != nil; n = n.Parent {
		if hasHandler(n, eventType) {
			return n
		}
	}
	return nil
}

func hasHandler(n *Node, eventType string) bool {
	switch eventType {
	case "click":
		return n.OnClick != 0
	case "mouseenter":
		return n.OnMouseEnter != 0
	case "mouseleave":
		return n.OnMouseLeave != 0
	case "keydown":
		return n.OnKeyDown != 0
	case "scroll":
		return n.OnScroll != 0
	case "mousedown":
		return n.OnMouseDown != 0
	case "mouseup":
		return n.OnMouseUp != 0
	case "submit":
		return n.OnSubmit != 0
	case "outsideclick":
		return n.OnOutsideClick != 0
	default:
		return false
	}
}

// HandleMouseMove processes a mouse move event.
// Performs hit-test, fires onMouseEnter/onMouseLeave as needed.
// hitTestLayers performs hit-test across all layers, from top to bottom.
// Returns the hit node and which layer it belongs to.
// For modal layers, if the click misses the layer's content, returns (nil, layer).
func (e *Engine) hitTestLayers(x, y int) (*Node, *Layer) {
	for i := len(e.layers) - 1; i >= 0; i-- {
		layer := e.layers[i]
		if layer.Root == nil {
			continue
		}
		node := HitTest(layer.Root, x, y)
		if node != nil {
			return node, layer
		}
		if layer.Modal {
			return nil, layer // Modal miss → block, return layer for onOutsideClick
		}
	}
	return nil, nil
}

// HitTestScreen returns the deepest node at (x, y) across all layers, top layer first.
// Used by DevTools inspect picker; same semantics as pointer hit-testing.
func (e *Engine) HitTestScreen(x, y int) *Node {
	n, _ := e.hitTestLayers(x, y)
	return n
}

// hitTestLayersWithHandler performs hit-test with handler across all layers.
func (e *Engine) hitTestLayersWithHandler(x, y int, eventType string) (*Node, *Layer) {
	for i := len(e.layers) - 1; i >= 0; i-- {
		layer := e.layers[i]
		if layer.Root == nil {
			continue
		}
		target := HitTestWithHandler(layer.Root, x, y, eventType)
		if target != nil {
			return target, layer
		}
		// Even without handler, if we hit this layer's area, don't pass through
		if HitTest(layer.Root, x, y) != nil {
			return nil, layer
		}
		if layer.Modal {
			return nil, layer
		}
	}
	return nil, nil
}

func (e *Engine) HandleMouseMove(x, y int) {
	if len(e.layers) == 0 {
		return
	}

	// If a widget has captured the mouse (dragging), dispatch directly to it
	if e.capturedComp != nil {
		w, ok := e.widgets[e.capturedComp.Type]
		if !ok {
			e.capturedComp = nil
		} else {
			state, stateOk := e.widgetStates[e.capturedComp.ID]
			if !stateOk {
				e.capturedComp = nil
			} else {
				appW, appH := e.appBounds()
				evt := &WidgetEvent{Type: "mousemove", X: x, Y: y, ScreenW: appW, ScreenH: appH}
				if e.capturedComp.RootNode != nil {
					evt.WidgetX = e.capturedComp.RootNode.X
					evt.WidgetY = e.capturedComp.RootNode.Y
					evt.WidgetW = e.capturedComp.RootNode.W
					evt.WidgetH = e.capturedComp.RootNode.H
					if scrollNode := findScrollNode(e.capturedComp.RootNode); scrollNode != nil {
						evt.ScrollY = scrollNode.ScrollY
						evt.ContentHeight = scrollNode.ScrollHeight
					}
				}
				consumed := w.DoOnEvent(e.capturedComp.Props, state, evt)
				if consumed {
					if evt.ScrollBy == 0 {
						e.capturedComp.Dirty = true
					}
					e.needsRender = true
				}
				// Fire onChange if widget requested it
				if evt.FireOnChange != nil && e.capturedComp.RootNode != nil && e.capturedComp.RootNode.OnChange != 0 {
					e.callLuaRefWithValue(e.capturedComp.RootNode.OnChange, evt.FireOnChange)
				}
				// Process scroll request from captured widget
				if evt.ScrollBy != 0 && e.capturedComp.RootNode != nil {
					e.scrollNodeBy(e.capturedComp.RootNode, evt.ScrollBy)
				}
				return // captured widget gets ALL mouse events, skip normal dispatch
			}
		}
	}

	// Clear stale hovered pointer if node was removed from tree
	if e.hoveredNode != nil && e.hoveredNode.Removed {
		e.hoveredNode = nil
	}

	target, _ := e.hitTestLayers(x, y)

	if target == e.hoveredNode {
		// Same node — still dispatch mousemove to widget
		if target != nil {
			e.dispatchWidgetEvent(target, "mousemove", "", x, y)
		}
		return
	}

	old := e.hoveredNode
	e.hoveredNode = target

	// Fire onMouseLeave on old node (bubble up to find handler)
	if old != nil && !old.Removed {
		e.dispatchWidgetEvent(old, "mouseleave", "", x, y)
		for n := old; n != nil; n = n.Parent {
			if n.OnMouseLeave != 0 {
				e.callLuaRef(n.OnMouseLeave, x, y)
				break
			}
		}
	}

	// Fire onMouseEnter on new node (bubble up to find handler)
	if target != nil {
		e.dispatchWidgetEvent(target, "mouseenter", "", x, y)
		for n := target; n != nil; n = n.Parent {
			if n.OnMouseEnter != 0 {
				e.callLuaRef(n.OnMouseEnter, x, y)
				break
			}
		}
		// Also dispatch mousemove to widget at current position
		e.dispatchWidgetEvent(target, "mousemove", "", x, y)
	}
}

// HandleClick processes a click event at screen coordinates (x, y).
// Finds the deepest node with an onClick handler (bubbling) and dispatches.
// Also handles focus: clicking a focusable node focuses it.
func (e *Engine) HandleClick(x, y int) {
	if len(e.layers) == 0 {
		return
	}
	// During widget capture (drag), all mouse events go to the captured widget
	if e.capturedComp != nil {
		return // captured widget only receives mousemove/mouseup
	}

	// Clear stale focused pointer if node was removed from tree
	if e.focusedNode != nil && e.focusedNode.Removed {
		e.focusedNode = nil
	}

	// Hit-test across layers for the deepest node at this position
	hitNode, hitLayer := e.hitTestLayers(x, y)

	// Modal miss: click outside modal layer content → block event
	if hitNode == nil && hitLayer != nil && hitLayer.Modal {
		// Fire onOutsideClick on focused node if present
		if e.focusedNode != nil && !e.focusedNode.Removed && e.focusedNode.OnOutsideClick != 0 {
			e.callLuaRef(e.focusedNode.OnOutsideClick, x, y)
		}
		return
	}

	// Focus management: clicking on a focusable node focuses it
	// Walk up from hitNode to find the nearest focusable ancestor
	var focusTarget *Node
	for n := hitNode; n != nil; n = n.Parent {
		if n.Focusable && !n.Disabled {
			focusTarget = n
			break
		}
	}

	// Fire onOutsideClick on the previously focused node if focus is moving away
	if e.focusedNode != nil && focusTarget != e.focusedNode && !e.focusedNode.Removed {
		if e.focusedNode.OnOutsideClick != 0 {
			e.callLuaRef(e.focusedNode.OnOutsideClick, x, y)
		}
	}

	if focusTarget != nil {
		e.setFocus(focusTarget)
	} else if e.focusedNode != nil {
		// Clicked on non-focusable area → blur current
		e.setFocus(nil)
	}

	// Dispatch widget click event
	if hitNode != nil {
		e.dispatchWidgetEvent(hitNode, "click", "", x, y)
	}

	// Dispatch onClick (bubble up from hit node, skip disabled)
	target, _ := e.hitTestLayersWithHandler(x, y, "click")
	if target != nil && target.OnClick != 0 && !target.Disabled {
		e.callLuaRef(target.OnClick, x, y)
	}
}

// HandleMouseDown processes a mousedown event at screen coordinates (x, y).
// Finds the deepest node with an onMouseDown handler (bubbling) and dispatches.
func (e *Engine) HandleMouseDown(x, y int) {
	if len(e.layers) == 0 {
		return
	}
	// During widget capture (drag), all mouse events go to the captured widget
	if e.capturedComp != nil {
		return // captured widget only receives mousemove/mouseup
	}
	// Dispatch widget mousedown event
	hitNode, _ := e.hitTestLayers(x, y)
	if hitNode != nil {
		e.dispatchWidgetEvent(hitNode, "mousedown", "", x, y)
	}
	target, _ := e.hitTestLayersWithHandler(x, y, "mousedown")
	if target != nil && target.OnMouseDown != 0 && !target.Disabled {
		e.callLuaRef(target.OnMouseDown, x, y)
	}
}

// HandleMouseUp processes a mouseup event at screen coordinates (x, y).
// Finds the deepest node with an onMouseUp handler (bubbling) and dispatches.
func (e *Engine) HandleMouseUp(x, y int) {
	if len(e.layers) == 0 {
		return
	}

	// If a widget has captured the mouse, dispatch to it and release capture
	if e.capturedComp != nil {
		w, ok := e.widgets[e.capturedComp.Type]
		if ok {
			state, stateOk := e.widgetStates[e.capturedComp.ID]
			if stateOk {
				appW, appH := e.appBounds()
				evt := &WidgetEvent{Type: "mouseup", X: x, Y: y, ScreenW: appW, ScreenH: appH}
				if e.capturedComp.RootNode != nil {
					evt.WidgetX = e.capturedComp.RootNode.X
					evt.WidgetY = e.capturedComp.RootNode.Y
					evt.WidgetW = e.capturedComp.RootNode.W
					evt.WidgetH = e.capturedComp.RootNode.H
					if scrollNode := findScrollNode(e.capturedComp.RootNode); scrollNode != nil {
						evt.ScrollY = scrollNode.ScrollY
						evt.ContentHeight = scrollNode.ScrollHeight
					}
				}
				if w.DoOnEvent(e.capturedComp.Props, state, evt) {
					if evt.ScrollBy == 0 {
						e.capturedComp.Dirty = true
					}
					e.needsRender = true
				}
				// Fire onChange if widget requested it
				if evt.FireOnChange != nil && e.capturedComp.RootNode != nil && e.capturedComp.RootNode.OnChange != 0 {
					e.callLuaRefWithValue(e.capturedComp.RootNode.OnChange, evt.FireOnChange)
				}
				// Process scroll request from captured widget
				if evt.ScrollBy != 0 && e.capturedComp.RootNode != nil {
					e.scrollNodeBy(e.capturedComp.RootNode, evt.ScrollBy)
				}
			}
			// stateOk == false: widget state missing, skip dispatch
		}
		e.capturedComp = nil // release capture
		return
	}

	// Dispatch widget mouseup event
	hitNode, _ := e.hitTestLayers(x, y)
	if hitNode != nil {
		e.dispatchWidgetEvent(hitNode, "mouseup", "", x, y)
	}
	target, _ := e.hitTestLayersWithHandler(x, y, "mouseup")
	if target != nil && target.OnMouseUp != 0 && !target.Disabled {
		e.callLuaRef(target.OnMouseUp, x, y)
	}
}

// HandleKeyDown processes a key event.
// Priority: Tab → focus cycle, focused input editing, then onKeyDown handler.
func (e *Engine) HandleKeyDown(key string) {
	if len(e.layers) == 0 {
		return
	}

	// Clear stale focused pointer if node was removed from tree
	if e.focusedNode != nil && e.focusedNode.Removed {
		e.focusedNode = nil
	}

	// Tab cycles focus
	if key == "Tab" {
		e.FocusNext()
		return
	}

	// If an input/textarea is focused, try input handling first
	if e.focusedNode != nil {
		if e.HandleInputKeyDown(key) {
			return // consumed by input system
		}
	}

	// Dispatch keydown to Go widget that owns the focused node.
	// dispatchWidgetEvent walks up from the node to find the owning widget
	// component and calls DoOnEvent("keydown", key). If the widget consumed
	// the event, don't dispatch to Lua onKeyDown handler.
	if e.focusedNode != nil {
		if e.dispatchWidgetEvent(e.focusedNode, "keydown", key, 0, 0) {
			return // consumed by Go widget, skip Lua onKeyDown
		}
	}

	// Fall through to onKeyDown handler — search from top layer down
	var keyHandlerNode *Node
	for i := len(e.layers) - 1; i >= 0; i-- {
		if e.layers[i].Root != nil {
			keyHandlerNode = e.findKeyHandler(e.layers[i].Root)
			if keyHandlerNode != nil {
				break
			}
			if e.layers[i].Modal {
				break // Don't pass through modal
			}
		}
	}
	if keyHandlerNode != nil && keyHandlerNode.OnKeyDown != 0 {
		e.callLuaRefKey(keyHandlerNode.OnKeyDown, key)
	}
}

// HandleScroll processes a scroll event at screen coordinates (x, y).
// Priority: custom onScroll Lua handler > built-in auto-scroll for overflow=scroll nodes.
func (e *Engine) HandleScroll(x, y, delta int) {
	if len(e.layers) == 0 {
		return
	}
	// During widget capture (drag), all mouse events go to the captured widget
	if e.capturedComp != nil {
		return // captured widget only receives mousemove/mouseup
	}

	// First: check for a custom Lua onScroll handler (takes priority)
	target, _ := e.hitTestLayersWithHandler(x, y, "scroll")
	if target != nil && target.OnScroll != 0 {
		e.callLuaRefScroll(target.OnScroll, delta)
		return
	}

	// No custom handler — find nearest overflow=scroll ancestor for auto-scroll
	hitNode, _ := e.hitTestLayers(x, y)
	scrollNode := findScrollableAncestor(hitNode)
	if scrollNode != nil {
		e.autoScroll(scrollNode, delta)
	}
}

// findScrollableAncestor walks up from node to find the nearest ancestor with overflow=scroll.
func findScrollableAncestor(node *Node) *Node {
	for n := node; n != nil; n = n.Parent {
		if n.Style.Overflow == "scroll" {
			return n
		}
	}
	return nil
}

// autoScroll adjusts a scroll container's ScrollY by delta, clamped to [0, maxScroll].
func (e *Engine) autoScroll(node *Node, delta int) {
	maxScroll := computeMaxScrollY(node)
	if maxScroll <= 0 {
		return // content fits, no scrolling needed
	}

	const step = 3 // scroll 3 lines per wheel tick
	newScrollY := node.ScrollY + delta*step

	// Clamp
	if newScrollY < 0 {
		newScrollY = 0
	}
	if newScrollY > maxScroll {
		newScrollY = maxScroll
	}

	if newScrollY == node.ScrollY {
		return // no change
	}

	node.ScrollY = newScrollY
	node.PaintDirty = true
	e.needsRender = true
}

// scrollNodeBy adjusts a node's ScrollY by the given number of lines (positive=down, negative=up).
// appBounds returns the app's logical dimensions (root component's rendered size).
// Falls back to engine dimensions (terminal size) if root is not available.
func (e *Engine) appBounds() (int, int) {
	if e.root != nil && e.root.RootNode != nil && e.root.RootNode.W > 0 && e.root.RootNode.H > 0 {
		return e.root.RootNode.W, e.root.RootNode.H
	}
	return e.width, e.height
}

// Unlike autoScroll, this does NOT multiply by a step factor — the caller provides the exact delta.
// The node must have overflow:"scroll" style and valid ScrollHeight from layout.
func (e *Engine) scrollNodeBy(node *Node, lines int) {
	// Find the scroll container: if the root node itself has overflow scroll, use it.
	// Otherwise look for the first child with overflow scroll.
	scrollNode := findScrollNode(node)
	if scrollNode == nil {
		return
	}
	maxScroll := computeMaxScrollY(scrollNode)
	if maxScroll <= 0 {
		return
	}

	newScrollY := scrollNode.ScrollY + lines
	if newScrollY < 0 {
		newScrollY = 0
	}
	if newScrollY > maxScroll {
		newScrollY = maxScroll
	}
	if newScrollY == scrollNode.ScrollY {
		return
	}

	scrollNode.ScrollY = newScrollY
	scrollNode.PaintDirty = true
	e.needsRender = true
}

// findScrollNode finds the first node with overflow:"scroll" in the tree rooted at node.
// Checks the node itself first, then searches children depth-first.
func findScrollNode(node *Node) *Node {
	if node.Style.Overflow == "scroll" {
		return node
	}
	for _, ch := range node.Children {
		if found := findScrollNode(ch); found != nil {
			return found
		}
	}
	return nil
}

// computeMaxScrollY calculates the maximum scroll offset for a scroll container.
// Uses the stored ScrollHeight (set by layout) minus the visible content height.
func computeMaxScrollY(node *Node) int {
	bw := 0
	if node.Style.Border != "" && node.Style.Border != "none" {
		bw = 1
	}
	contentH := node.H - 2*bw - node.Style.PaddingTop - node.Style.PaddingBottom
	if contentH <= 0 {
		return 0
	}
	maxScroll := node.ScrollHeight - contentH
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// callLuaRef calls a Lua function by registry ref with an event table {x=x, y=y}.
func (e *Engine) callLuaRef(ref LuaRef, x, y int) {
	L := e.L
	L.RawGetI(lua.RegistryIndex, ref)
	if !L.IsFunction(-1) {
		L.Pop(1)
		return
	}
	// Push event table: {x=x, y=y}
	L.NewTable()
	tblIdx := L.AbsIndex(-1)
	L.PushInteger(int64(x))
	L.SetField(tblIdx, "x")
	L.PushInteger(int64(y))
	L.SetField(tblIdx, "y")
	if status := L.PCall(1, 0, 0); status != lua.OK {
		L.Pop(1) // pop error message to prevent stack pollution
	}
}

// callLuaRefKey calls a Lua function with an event table {key=key}.
func (e *Engine) callLuaRefKey(ref LuaRef, key string) {
	L := e.L
	L.RawGetI(lua.RegistryIndex, ref)
	if !L.IsFunction(-1) {
		L.Pop(1)
		return
	}
	L.NewTable()
	tblIdx := L.AbsIndex(-1)
	L.PushString(key)
	L.SetField(tblIdx, "key")
	if status := L.PCall(1, 0, 0); status != lua.OK {
		L.Pop(1) // pop error message to prevent stack pollution
	}
}

// callLuaRefScroll calls a Lua function with an event table {delta=delta, key="up"/"down"}.
func (e *Engine) callLuaRefScroll(ref LuaRef, delta int) {
	L := e.L
	L.RawGetI(lua.RegistryIndex, ref)
	if !L.IsFunction(-1) {
		L.Pop(1)
		return
	}
	L.NewTable()
	tblIdx := L.AbsIndex(-1)
	L.PushInteger(int64(delta))
	L.SetField(tblIdx, "delta")
	// Also add key field for compatibility with Lua scripts expecting "up"/"down"
	if delta < 0 {
		L.PushString("up")
	} else {
		L.PushString("down")
	}
	L.SetField(tblIdx, "key")
	if status := L.PCall(1, 0, 0); status != lua.OK {
		L.Pop(1) // pop error message to prevent stack pollution
	}
}

// findKeyHandler walks the tree (DFS) to find a node with an onKeyDown handler.
func (e *Engine) findKeyHandler(node *Node) *Node {
	if node == nil {
		return nil
	}
	if node.OnKeyDown != 0 {
		return node
	}
	for _, child := range node.Children {
		if found := e.findKeyHandler(child); found != nil {
			return found
		}
	}
	return nil
}

// callLuaRefSimple calls a Lua function by registry ref with no arguments.
func (e *Engine) callLuaRefSimple(ref LuaRef) {
	L := e.L
	L.RawGetI(lua.RegistryIndex, ref)
	if !L.IsFunction(-1) {
		L.Pop(1)
		return
	}
	if status := L.PCall(0, 0, 0); status != lua.OK {
		L.Pop(1) // pop error message to prevent stack pollution
	}
}

// dispatchWidgetEvent walks up from node to find an owning Go widget component,
// then calls WidgetDef.DoOnEvent. If it returns true (state changed), the
// component is marked dirty for re-render.
// If the widget sets FireOnChange on the event, the engine fires the onChange
// Lua callback on the widget's root node.
// Returns true if a widget was found AND DoOnEvent returned true (event consumed).
func (e *Engine) dispatchWidgetEvent(node *Node, eventType, key string, x, y int) bool {
	comp := findOwnerComponent(node)
	if comp == nil {
		return false
	}
	w, ok := e.widgets[comp.Type]
	if !ok {
		return false
	}
	state, ok := e.widgetStates[comp.ID]
	if !ok {
		return false
	}
	appW, appH := e.appBounds()
	evt := &WidgetEvent{Type: eventType, Key: key, X: x, Y: y, ScreenW: appW, ScreenH: appH}
	// Populate widget screen bounds so widgets know their position
	if comp.RootNode != nil {
		evt.WidgetX = comp.RootNode.X
		evt.WidgetY = comp.RootNode.Y
		evt.WidgetW = comp.RootNode.W
		evt.WidgetH = comp.RootNode.H
		// Populate scroll state from the scroll container
		if scrollNode := findScrollNode(comp.RootNode); scrollNode != nil {
			evt.ScrollY = scrollNode.ScrollY
			evt.ContentHeight = scrollNode.ScrollHeight
		}
	}
	consumed := w.DoOnEvent(comp.Props, state, evt)
	if consumed {
		// Don't mark dirty if only ScrollBy was set — scrolling is handled
		// by the engine directly on the existing node tree (no re-render needed).
		if evt.ScrollBy == 0 {
			comp.Dirty = true
		}
		e.needsRender = true
	}

	// Mouse capture: widget requests to capture all subsequent mouse events
	if evt.CaptureMouse {
		e.capturedComp = comp
	}

	// Fire onChange if widget requested it
	if evt.FireOnChange != nil && comp.RootNode != nil && comp.RootNode.OnChange != 0 {
		e.callLuaRefWithValue(comp.RootNode.OnChange, evt.FireOnChange)
	}

	// Process layer requests from widget
	if evt.CreateLayer != nil {
		req := evt.CreateLayer
		e.CreateLayer(req.ID, req.Root, req.Modal)
	}
	if evt.RemoveLayer != "" {
		e.RemoveLayer(evt.RemoveLayer)
	}

	// Scroll request: widget wants to scroll its root node by N lines
	if evt.ScrollBy != 0 && comp.RootNode != nil {
		e.scrollNodeBy(comp.RootNode, evt.ScrollBy)
	}

	return consumed
}

// callLuaRefWithValue calls a Lua function by registry ref with a single typed value argument.
func (e *Engine) callLuaRefWithValue(ref LuaRef, value any) {
	L := e.L
	L.RawGetI(lua.RegistryIndex, ref)
	if !L.IsFunction(-1) {
		L.Pop(1)
		return
	}
	switch v := value.(type) {
	case bool:
		L.PushBoolean(v)
	case string:
		L.PushString(v)
	case float64:
		L.PushNumber(v)
	case int:
		L.PushInteger(int64(v))
	case int64:
		L.PushInteger(v)
	case map[string]any:
		L.CreateTable(0, len(v))
		for key, val := range v {
			L.PushString(key)
			switch tv := val.(type) {
			case bool:
				L.PushBoolean(tv)
			case string:
				L.PushString(tv)
			case float64:
				L.PushNumber(tv)
			case int:
				L.PushInteger(int64(tv))
			case int64:
				L.PushInteger(tv)
			default:
				L.PushNil()
			}
			L.SetTable(-3)
		}
	default:
		L.PushNil()
	}
	if status := L.PCall(1, 0, 0); status != lua.OK {
		L.Pop(1) // pop error message to prevent stack pollution
	}
}

// findOwnerComponent walks up the node tree to find the nearest Component.
func findOwnerComponent(node *Node) *Component {
	for n := node; n != nil; n = n.Parent {
		if n.Component != nil {
			return n.Component
		}
	}
	return nil
}
