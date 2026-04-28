package render

import "github.com/akzj/go-lua/pkg/lua"

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

	// Check children in reverse order (top-most first)
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
func (e *Engine) HandleMouseMove(x, y int) {
	if e.root == nil || e.root.RootNode == nil {
		return
	}

	// Clear stale hovered pointer if node was removed from tree
	if e.hoveredNode != nil && e.hoveredNode.Removed {
		e.hoveredNode = nil
	}

	target := HitTest(e.root.RootNode, x, y)

	if target == e.hoveredNode {
		return // same node, no change
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
	}
}

// HandleClick processes a click event at screen coordinates (x, y).
// Finds the deepest node with an onClick handler (bubbling) and dispatches.
// Also handles focus: clicking a focusable node focuses it.
func (e *Engine) HandleClick(x, y int) {
	if e.root == nil || e.root.RootNode == nil {
		return
	}

	// Clear stale focused pointer if node was removed from tree
	if e.focusedNode != nil && e.focusedNode.Removed {
		e.focusedNode = nil
	}

	// Hit-test for the deepest node at this position
	hitNode := HitTest(e.root.RootNode, x, y)

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
	target := HitTestWithHandler(e.root.RootNode, x, y, "click")
	if target != nil && target.OnClick != 0 && !target.Disabled {
		e.callLuaRef(target.OnClick, x, y)
	}
}

// HandleMouseDown processes a mousedown event at screen coordinates (x, y).
// Finds the deepest node with an onMouseDown handler (bubbling) and dispatches.
func (e *Engine) HandleMouseDown(x, y int) {
	if e.root == nil || e.root.RootNode == nil {
		return
	}
	// Dispatch widget mousedown event
	hitNode := HitTest(e.root.RootNode, x, y)
	if hitNode != nil {
		e.dispatchWidgetEvent(hitNode, "mousedown", "", x, y)
	}
	target := HitTestWithHandler(e.root.RootNode, x, y, "mousedown")
	if target != nil && target.OnMouseDown != 0 && !target.Disabled {
		e.callLuaRef(target.OnMouseDown, x, y)
	}
}

// HandleMouseUp processes a mouseup event at screen coordinates (x, y).
// Finds the deepest node with an onMouseUp handler (bubbling) and dispatches.
func (e *Engine) HandleMouseUp(x, y int) {
	if e.root == nil || e.root.RootNode == nil {
		return
	}
	// Dispatch widget mouseup event
	hitNode := HitTest(e.root.RootNode, x, y)
	if hitNode != nil {
		e.dispatchWidgetEvent(hitNode, "mouseup", "", x, y)
	}
	target := HitTestWithHandler(e.root.RootNode, x, y, "mouseup")
	if target != nil && target.OnMouseUp != 0 && !target.Disabled {
		e.callLuaRef(target.OnMouseUp, x, y)
	}
}

// HandleKeyDown processes a key event.
// Priority: Tab → focus cycle, focused input editing, then onKeyDown handler.
func (e *Engine) HandleKeyDown(key string) {
	if e.root == nil || e.root.RootNode == nil {
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

	// Fall through to onKeyDown handler
	node := e.findKeyHandler(e.root.RootNode)
	if node != nil && node.OnKeyDown != 0 {
		e.callLuaRefKey(node.OnKeyDown, key)
	}
}

// HandleScroll processes a scroll event at screen coordinates (x, y).
// Priority: custom onScroll Lua handler > built-in auto-scroll for overflow=scroll nodes.
func (e *Engine) HandleScroll(x, y, delta int) {
	if e.root == nil || e.root.RootNode == nil {
		return
	}

	// First: check for a custom Lua onScroll handler (takes priority)
	target := HitTestWithHandler(e.root.RootNode, x, y, "scroll")
	if target != nil && target.OnScroll != 0 {
		e.callLuaRefScroll(target.OnScroll, delta)
		return
	}

	// No custom handler — find nearest overflow=scroll ancestor for auto-scroll
	hitNode := HitTest(e.root.RootNode, x, y)
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
func (e *Engine) dispatchWidgetEvent(node *Node, eventType, key string, x, y int) {
	comp := findOwnerComponent(node)
	if comp == nil {
		return
	}
	w, ok := e.widgets[comp.Type]
	if !ok {
		return
	}
	state, ok := e.widgetStates[comp.ID]
	if !ok {
		return
	}
	if w.DoOnEvent(comp.Props, state, eventType, key, x, y) {
		comp.Dirty = true
		e.needsRender = true
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
