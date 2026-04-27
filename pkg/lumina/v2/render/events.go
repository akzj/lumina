package render

import "github.com/akzj/go-lua/pkg/lua"

// HitTest finds the deepest Node at screen coordinates (x, y).
// Returns nil if no node contains the point.
func HitTest(root *Node, x, y int) *Node {
	if root == nil {
		return nil
	}
	// Check if point is within this node's bounds
	if x < root.X || x >= root.X+root.W || y < root.Y || y >= root.Y+root.H {
		return nil
	}
	// Check children in reverse order (last child is on top / higher z-index)
	for i := len(root.Children) - 1; i >= 0; i-- {
		if hit := HitTest(root.Children[i], x, y); hit != nil {
			return hit
		}
	}
	// No child hit — this node is the target
	return root
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
		for n := old; n != nil; n = n.Parent {
			if n.OnMouseLeave != 0 {
				e.callLuaRef(n.OnMouseLeave, x, y)
				break
			}
		}
	}

	// Fire onMouseEnter on new node (bubble up to find handler)
	if target != nil {
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
// Also handles focus: clicking an input/textarea focuses it.
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

	// Focus management: clicking on input/textarea focuses it
	if hitNode != nil && (hitNode.Type == "input" || hitNode.Type == "textarea") {
		old := e.focusedNode
		e.focusedNode = hitNode
		if old != nil && old != hitNode && !old.Removed {
			old.PaintDirty = true
		}
		hitNode.PaintDirty = true
	}

	// Dispatch onClick (bubble up from hit node)
	target := HitTestWithHandler(e.root.RootNode, x, y, "click")
	if target != nil && target.OnClick != 0 {
		e.callLuaRef(target.OnClick, x, y)
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
// Finds the deepest node with an onScroll handler (bubbling) and dispatches.
func (e *Engine) HandleScroll(x, y, delta int) {
	if e.root == nil || e.root.RootNode == nil {
		return
	}
	target := HitTestWithHandler(e.root.RootNode, x, y, "scroll")
	if target != nil && target.OnScroll != 0 {
		e.callLuaRefScroll(target.OnScroll, delta)
	}
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
