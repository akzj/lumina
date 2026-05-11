package render

import (
	"fmt"
	"log"
	"os"

	"github.com/akzj/go-lua/pkg/lua"
)

// HitTest finds the deepest Node at screen coordinates (x, y).
// Returns nil if no node contains the point.
// Uses cumulative scroll offset for correct nested scroll handling.
func HitTest(root *Node, x, y int) *Node {
	return hitTestWithOffset(root, x, y, 0, 0)
}

// hitTestWithOffset performs hit testing with a cumulative scroll offset.
// scrollOffsetY accumulates as we descend through nested scroll containers.
// Each node's screen position = layout position - cumulativeScrollOffset.
func hitTestWithOffset(node *Node, x, y int, scrollOffsetX, scrollOffsetY int) *Node {
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
				if hit := hitTestWithOffset(child, x, y, scrollOffsetX, scrollOffsetY); hit != nil {
					return hit
				}
			}
		} else {
			for i := len(node.Children) - 1; i >= 0; i-- {
				if hit := hitTestWithOffset(node.Children[i], x, y, scrollOffsetX, scrollOffsetY); hit != nil {
					return hit
				}
			}
		}
		return nil // component placeholder is transparent — no match
	}

	// Node's screen position = layout position - cumulative scroll offset
	screenX := node.X - scrollOffsetX
	screenY := node.Y - scrollOffsetY

	// Bounds check using screen coordinates
	if x < screenX || x >= screenX+node.W || y < screenY || y >= screenY+node.H {
		return nil
	}

	// Calculate child scroll offsets
	childScrollOffsetX := scrollOffsetX
	childScrollOffsetY := scrollOffsetY
	if node.Style.Overflow == "scroll" {
		childScrollOffsetX += node.ScrollX
		childScrollOffsetY += node.ScrollY

		// Clip: only process clicks within content area (not border/padding)
		bw := 0
		if hasBorder(node.Style) {
			bw = 1
		}
		visTop := screenY + bw + node.Style.PaddingTop
		visBot := screenY + node.H - bw - node.Style.PaddingBottom
		visLeft := screenX + bw + node.Style.PaddingLeft
		visRight := screenX + node.W - bw - node.Style.PaddingRight
		if y < visTop || y >= visBot || x < visLeft || x >= visRight {
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
				childScreenY := child.Y - childScrollOffsetY
				childScreenBottom := childScreenY + child.H
				bw := 0
				if hasBorder(node.Style) {
					bw = 1
				}
				visTop := screenY + bw + node.Style.PaddingTop
				visBot := screenY + node.H - bw - node.Style.PaddingBottom
				if childScreenBottom <= visTop || childScreenY >= visBot {
					continue // scrolled out of view
				}
			}

			if hit := hitTestWithOffset(child, x, y, childScrollOffsetX, childScrollOffsetY); hit != nil {
				return hit
			}
		}
	} else {
		// Fast path: no z-index, reverse order (same as before)
		for i := len(node.Children) - 1; i >= 0; i-- {
			child := node.Children[i]

			// For scroll containers, skip children outside visible area
			if node.Style.Overflow == "scroll" {
				childScreenY := child.Y - childScrollOffsetY
				childScreenBottom := childScreenY + child.H
				bw := 0
				if hasBorder(node.Style) {
					bw = 1
				}
				visTop := screenY + bw + node.Style.PaddingTop
				visBot := screenY + node.H - bw - node.Style.PaddingBottom
				if childScreenBottom <= visTop || childScreenY >= visBot {
					continue // scrolled out of view
				}
			}

			if hit := hitTestWithOffset(child, x, y, childScrollOffsetX, childScrollOffsetY); hit != nil {
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
	case "rightclick":
		return n.OnRightClick != 0
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

	// Scrollbar drag in progress?
	if e.scrollbarDragNode != nil {
		e.handleScrollbarDrag(x, y)
		return
	}

	// Mouse capture: if a node has captured mouse, send events directly to it
	// Note: the captured node can be removed by a drag-triggered re-render. Use the
	// cached handler ref so resize/drag state keeps receiving events until mouseup.
	if e.capturedNode != nil {
		if e.captureMoveRef != 0 {
			e.callLuaRef(e.captureMoveRef, x, y)
		}
		return
	}

	// Clear stale hovered pointer if node was removed from tree.
	// Fire the cached onMouseLeave ref (parent chain is broken by markRemovedRecursive).
	if e.hoveredNode != nil && e.hoveredNode.Removed {
		e.hoveredNode = nil
		if e.hoverLeaveRef != 0 {
			e.callLuaRef(e.hoverLeaveRef, 0, 0)
			e.L.Unref(lua.RegistryIndex, int(e.hoverLeaveRef))
			e.hoverLeaveRef = 0
		}
	}

	target, _ := e.hitTestLayers(x, y)

	if target == e.hoveredNode {
		// Fire onMouseMove on hovered node even when it hasn't changed
		if target != nil {
			for n := target; n != nil; n = n.Parent {
				if n.OnMouseMove != 0 {
					e.callLuaRef(n.OnMouseMove, x, y)
					break
				}
			}
		}
		return
	}

	old := e.hoveredNode
	e.hoveredNode = target

	// Cache the OnMouseLeave ref from the new target's ancestor chain
	e.hoverLeaveRef = 0
	if target != nil {
		for n := target; n != nil; n = n.Parent {
			if n.OnMouseLeave != 0 {
				e.hoverLeaveRef = n.OnMouseLeave
				break
			}
		}
	}

	// Update hoverStyle state: clear Hovered on old path, set on new path
	if old != nil && !old.Removed {
		for n := old; n != nil; n = n.Parent {
			if n.HoverStyle != nil && n.Hovered {
				n.Hovered = false
				n.PaintDirty = true
				if n.Parent != nil {
					n.Parent.PaintDirty = true
				}
				e.needsRender = true
			}
		}
	}
	if target != nil {
		for n := target; n != nil; n = n.Parent {
			if n.HoverStyle != nil && !n.Hovered {
				n.Hovered = true
				n.PaintDirty = true
				if n.Parent != nil {
					n.Parent.PaintDirty = true
				}
				e.needsRender = true
			}
		}
	}

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
		// Fire onMouseMove on new hovered node
		for n := target; n != nil; n = n.Parent {
			if n.OnMouseMove != 0 {
				e.callLuaRef(n.OnMouseMove, x, y)
				break
			}
		}
	}
}

// HandleClick processes a click event at screen coordinates (x, y).
// Finds the deepest node with an onClick handler (bubbling) and dispatches.
// Also handles focus: clicking a focusable node focuses it.
// button is "left", "right", "middle", or "".
// Right-clicks dispatch to OnRightClick instead of OnClick.
func (e *Engine) HandleClick(x, y int, button string) {
	e.currentButton = button

	if len(e.layers) == 0 {
		return
	}

	// Update hover state to match click position.
	// Terminal clients often send click without a preceding mouse-move,
	// so hover may be stale. This prevents ghost hover backgrounds.
	e.HandleMouseMove(x, y)

	// Clear stale focused pointer if node was removed from tree
	if e.focusedNode != nil && e.focusedNode.Removed {
		e.focusedNode = nil
	}

	// Phase A: Scrollbar click — check if click is on a scrollbar column
	for i := len(e.layers) - 1; i >= 0; i-- {
		if e.layers[i].Root != nil {
			if e.handleScrollbarClick(e.layers[i].Root, x, y) {
				e.needsRender = true
				return
			}
		}
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

	// Check focusTarget is still valid (onOutsideClick handler may have removed it)
	if focusTarget != nil && !focusTarget.Removed {
		e.setFocus(focusTarget)
	} else if e.focusedNode != nil {
		// Clicked on non-focusable area or target was removed → blur current
		e.setFocus(nil)
	}

	// Dispatch click: right-click → OnRightClick, left/middle → OnClick
	if button == "right" {
		for n := hitNode; n != nil; n = n.Parent {
			if n.OnRightClick != 0 && !n.Disabled {
				e.callLuaRef(n.OnRightClick, x, y)
				break
			}
		}
	} else {
		for n := hitNode; n != nil; n = n.Parent {
			if n.OnClick != 0 && !n.Disabled {
				e.callLuaRef(n.OnClick, x, y)
				break
			}
		}
	}
}

// HandleMouseDown processes a mousedown event at screen coordinates (x, y).
// Finds the deepest node with an onMouseDown handler (bubbling) and dispatches.
func (e *Engine) HandleMouseDown(x, y int, button string) {
	e.clickPrevented = false // reset at start of each mousedown
	e.currentButton = button

	if len(e.layers) == 0 {
		return
	}

	// Check if mousedown is on a scrollbar thumb → start drag
	if e.handleScrollbarDragStart(x, y) {
		return
	}

	target, _ := e.hitTestLayersWithHandler(x, y, "mousedown")
	if target != nil && target.OnMouseDown != 0 && !target.Disabled {
		result := e.callLuaRef(target.OnMouseDown, x, y)
		if result.DefaultPrevented {
			e.clickPrevented = true
		}
	}

	// Mouse capture is only needed for drag/resize interactions that keep handling
	// motion after the pointer leaves the original node. Plain buttons also have
	// onMouseDown/onMouseUp for pressed state, but should not capture the mouse.
	if target != nil && findMouseMoveRef(target) != 0 {
		e.capturedNode = target
		e.captureMoveRef = findMouseMoveRef(target)
		e.captureMouseUpRef = findMouseUpRef(target)
	}
}

// ClickPrevented returns true if the last mousedown handler called preventDefault.
func (e *Engine) ClickPrevented() bool {
	return e.clickPrevented
}

// HandleMouseUp processes a mouseup event at screen coordinates (x, y).
// Sends mouseup to captured node first (critical for drag release), then falls
// back to hitTest if no capture was active.
func (e *Engine) HandleMouseUp(x, y int) {
	// End scrollbar drag if active
	if e.scrollbarDragNode != nil {
		e.scrollbarDragNode = nil
		return
	}

	// Send mouseup to captured node first (critical for drag release)
	captured := e.capturedNode
	e.capturedNode = nil // release capture
	mouseUpRef := e.captureMouseUpRef
	e.captureMoveRef = 0
	e.captureMouseUpRef = 0

	// Deliver mouseup even if captured node was Removed — this resets drag/resize state.
	if captured != nil {
		if mouseUpRef != 0 {
			e.callLuaRef(mouseUpRef, x, y)
		}
		e.drainPendingUnrefs()
		return // captured node handled it, don't also dispatch to hitTest target
	}

	// No capture — fall back to hitTest
	if len(e.layers) == 0 {
		return
	}
	target, _ := e.hitTestLayersWithHandler(x, y, "mouseup")
	if target != nil && target.OnMouseUp != 0 && !target.Disabled {
		e.callLuaRef(target.OnMouseUp, x, y)
	}
}

func findMouseMoveRef(node *Node) LuaRef {
	for n := node; n != nil; n = n.Parent {
		if n.OnMouseMove != 0 {
			return n.OnMouseMove
		}
	}
	return 0
}

func findMouseUpRef(node *Node) LuaRef {
	for n := node; n != nil; n = n.Parent {
		if n.OnMouseUp != 0 {
			return n.OnMouseUp
		}
	}
	return 0
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
	if key == "Shift+Tab" {
		e.FocusPrev()
		return
	}

	// Native input/textarea removed — all text input is handled by Lua components via onKeyDown.

	// Page up/down: scroll the nearest vertical scroll container (under focus,
	// or at viewport center if nothing is focused).
	if key == "PageUp" || key == "PageDown" {
		if e.handlePageScrollKey(key) {
			return
		}
	}

	// Fall through to onKeyDown handler.
	// Priority: walk up from focused node first (bubble), then DFS fallback.
	var keyHandlerNode *Node
	if e.focusedNode != nil && !e.focusedNode.Removed {
		// First: check the focused node itself
		if e.focusedNode.OnKeyDown != 0 {
			keyHandlerNode = e.focusedNode
		}
		// Then: check children (for component placeholders whose rendered
		// content has the onKeyDown handler)
		if keyHandlerNode == nil && len(e.focusedNode.Children) > 0 {
			keyHandlerNode = findKeyHandlerInSubtree(e.focusedNode)
		}
		// Then: walk up from focused node (existing bubble logic)
		if keyHandlerNode == nil {
			for n := e.focusedNode; n != nil; n = n.Parent {
				if n.OnKeyDown != 0 {
					keyHandlerNode = n
					break
				}
			}
		}
	}

	// Fallback: DFS search from top layer down (for apps with no focused node)
	if keyHandlerNode == nil {
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
	}
	if keyHandlerNode != nil && keyHandlerNode.OnKeyDown != 0 {
		e.callLuaRefKey(keyHandlerNode.OnKeyDown, key)
	}
}

// handlePageScrollKey applies one viewport of ScrollY for PageUp/PageDown on the
// scrollable ancestor of the focused node (or the top layer hit at screen center).
// Returns true if a scroll container was updated.
func (e *Engine) handlePageScrollKey(key string) bool {
	var anchor *Node
	if e.focusedNode != nil && !e.focusedNode.Removed {
		anchor = e.focusedNode
	} else {
		for i := len(e.layers) - 1; i >= 0; i-- {
			if e.layers[i].Root == nil {
				continue
			}
			cx := e.width / 2
			cy := e.height / 2
			if cx < 0 {
				cx = 0
			}
			if cy < 0 {
				cy = 0
			}
			anchor = HitTest(e.layers[i].Root, cx, cy)
			break
		}
	}
	scroll := findScrollableAncestor(anchor)
	if scroll == nil {
		return false
	}
	page := scrollViewportLines(scroll)
	if page < 1 {
		page = 1
	}
	d := page
	if key == "PageUp" {
		d = -page
	}
	maxScroll := computeMaxScrollY(scroll)
	if maxScroll <= 0 {
		return false
	}
	newY := scroll.ScrollY + d
	if newY < 0 {
		newY = 0
	}
	if newY > maxScroll {
		newY = maxScroll
	}
	if newY == scroll.ScrollY {
		return true // at boundary; still "handled" so we don't fall through to Lua
	}
	scroll.ScrollY = newY
	scroll.TargetScrollY = scroll.ScrollY
	scroll.PaintDirty = true
	e.needsRender = true
	return true
}

// scrollViewportLines is the inner visible height in rows of a scroll container.
func scrollViewportLines(n *Node) int {
	if n == nil {
		return 0
	}
	if n.Style.Overflow != "scroll" {
		return 0
	}
	bw := 0
	if hasBorder(n.Style) {
		bw = 1
	}
	h := n.H - 2*bw - n.Style.PaddingTop - n.Style.PaddingBottom
	if h < 1 {
		return 1
	}
	return h
}

// HandleScroll processes a scroll event at screen coordinates (x, y).
// Always performs auto-scroll first (if a scroll container exists), then calls any
// custom Lua onScroll handler so it receives the updated scrollY.
func (e *Engine) HandleScroll(x, y, delta int) {
	if len(e.layers) == 0 {
		return
	}

	// Step 1: Always do auto-scroll first (if there's a scroll container)
	hitNode, _ := e.hitTestLayers(x, y)
	scrollNode := findScrollableAncestor(hitNode)
	if scrollNode != nil {
		e.autoScroll(scrollNode, delta)
	}

	// Step 2: Then call onScroll handler (with updated scrollY)
	target, _ := e.hitTestLayersWithHandler(x, y, "scroll")
	if target != nil && target.OnScroll != 0 {
		e.callLuaRefScroll(target.OnScroll, delta, scrollNode)
	}

	// Step 3: Re-evaluate hover after scroll offset changed.
	// The node under (x, y) may have changed due to scroll, so fire
	// onMouseLeave/onMouseEnter as needed (e.g., dismiss tooltip).
	e.HandleMouseMove(x, y)
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

// autoScroll adjusts a scroll container's TargetScrollY by delta, clamped to [0, maxScroll].
// The actual ScrollY is animated toward TargetScrollY by TickSmoothScroll (called at 60Hz).
func (e *Engine) autoScroll(node *Node, delta int) {
	maxScroll := computeMaxScrollY(node)
	if maxScroll <= 0 {
		// Content fits vertically — redirect to horizontal scroll if content overflows horizontally
		maxScrollX := computeMaxScrollX(node)
		if maxScrollX > 0 {
			e.autoScrollX(node, delta)
		}
		return
	}

	const step = 3 // scroll 3 lines per wheel tick
	newTarget := node.TargetScrollY + delta*step

	// Clamp
	if newTarget < 0 {
		newTarget = 0
	}
	if newTarget > maxScroll {
		newTarget = maxScroll
	}

	if newTarget == node.TargetScrollY {
		return // no change
	}

	node.TargetScrollY = newTarget
	// Don't set ScrollY directly — TickSmoothScroll will animate it
	e.needsRender = true
}




// handleScrollbarClick walks the tree (DFS) looking for scroll containers with visible
// scrollbars. If the click (x,y) lands on a scrollbar, it handles page up/down or
// thumb-jump and returns true. The scrollbar geometry matches paintScrollbar.
func (e *Engine) handleScrollbarClick(node *Node, x, y int) bool {
	if node == nil {
		return false
	}
	if node.Style.Display == "none" || node.Style.Visibility == "hidden" {
		return false
	}

	// Check children first (DFS) — deepest scroll container wins
	for _, child := range node.Children {
		if e.handleScrollbarClick(child, x, y) {
			return true
		}
	}

	// Check this node: is it a scroll container with a visible scrollbar?
	if node.Style.Overflow != "scroll" {
		return false
	}
	if node.Style.Scrollbar == "none" {
		return false
	}
	maxScroll := computeMaxScrollY(node)
	if maxScroll <= 0 {
		return false
	}

	bw := 0
	if hasBorder(node.Style) {
		bw = 1
	}
	scrollbarX := node.X + node.W - bw - node.Style.PaddingRight - 1
	innerY1 := node.Y + bw + node.Style.PaddingTop
	innerY2 := node.Y + node.H - bw - node.Style.PaddingBottom

	// Is click on the scrollbar column?
	if x != scrollbarX {
		return false
	}
	if y < innerY1 || y >= innerY2 {
		return false
	}

	visibleH := innerY2 - innerY1
	totalH := node.ScrollHeight
	if totalH <= 0 || visibleH <= 0 {
		return false
	}

	// Thumb geometry (same as paintScrollbar)
	thumbSize := visibleH * visibleH / totalH
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > visibleH {
		thumbSize = visibleH
	}
	trackSpace := visibleH - thumbSize
	thumbPos := 0
	if maxScroll > 0 && trackSpace > 0 {
		thumbPos = node.ScrollY * trackSpace / maxScroll
	}

	clickRel := y - innerY1 // click position relative to top of track

	if clickRel >= thumbPos && clickRel < thumbPos+thumbSize {
		// Click on thumb itself → do nothing (drag handles this)
		return true
	}

	// Click on track (above or below thumb) → jump to proportional position
	// Center the thumb on the click position
	targetThumbPos := clickRel - thumbSize/2
	if targetThumbPos < 0 {
		targetThumbPos = 0
	}
	if targetThumbPos > trackSpace {
		targetThumbPos = trackSpace
	}
	newSY := 0
	if trackSpace > 0 {
		newSY = targetThumbPos * maxScroll / trackSpace
	}
	if newSY < 0 {
		newSY = 0
	}
	if newSY > maxScroll {
		newSY = maxScroll
	}
	if e.AnimManager != nil && e.NowMs != nil {
		fromSY := node.ScrollY
		animID := fmt.Sprintf("__scrollbar_scroll_%p", node)
		nowMs := e.NowMs()
		e.AnimManager.StartAnim(animID, float64(fromSY), float64(newSY), 300, "easeOut", false, func(value float64) {
			node.ScrollY = int(value)
			node.TargetScrollY = node.ScrollY
			node.PaintDirty = true
			e.needsRender = true
		}, func() {
			node.ScrollY = newSY
			node.TargetScrollY = node.ScrollY
			node.PaintDirty = true
			e.needsRender = true
		}, nowMs)
	} else {
		node.ScrollY = newSY
		node.TargetScrollY = node.ScrollY
		node.PaintDirty = true
	}
	return true
}

// handleScrollbarDragStart checks if (x,y) is on a scrollbar thumb and starts drag if so.
func (e *Engine) handleScrollbarDragStart(x, y int) bool {
	for i := len(e.layers) - 1; i >= 0; i-- {
		if e.layers[i].Root != nil {
			if node := e.findScrollbarThumbAt(e.layers[i].Root, x, y); node != nil {
				e.scrollbarDragNode = node
				e.scrollbarDragStartY = y
				e.scrollbarDragStartScrollY = node.ScrollY
				return true
			}
		}
	}
	return false
}

// findScrollbarThumbAt walks the tree and returns the scroll node whose thumb is at (x,y).
func (e *Engine) findScrollbarThumbAt(node *Node, x, y int) *Node {
	// DFS - check children first (deeper nodes have priority)
	for _, child := range node.Children {
		if found := e.findScrollbarThumbAt(child, x, y); found != nil {
			return found
		}
	}

	if node.Style.Overflow != "scroll" {
		return nil
	}
	if node.Style.Scrollbar == "none" {
		return nil
	}
	maxScroll := computeMaxScrollY(node)
	if maxScroll <= 0 {
		return nil
	}

	bw := 0
	if hasBorder(node.Style) {
		bw = 1
	}
	scrollbarX := node.X + node.W - bw - node.Style.PaddingRight - 1
	innerY1 := node.Y + bw + node.Style.PaddingTop
	innerY2 := node.Y + node.H - bw - node.Style.PaddingBottom

	if x != scrollbarX {
		return nil
	}
	if y < innerY1 || y >= innerY2 {
		return nil
	}

	// Check if click is on thumb
	visibleH := innerY2 - innerY1
	totalH := node.ScrollHeight
	if totalH <= 0 || visibleH <= 0 {
		return nil
	}
	thumbSize := visibleH * visibleH / totalH
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > visibleH {
		thumbSize = visibleH
	}
	trackSpace := visibleH - thumbSize
	thumbPos := 0
	if maxScroll > 0 && trackSpace > 0 {
		thumbPos = node.ScrollY * trackSpace / maxScroll
	}

	clickRel := y - innerY1
	if clickRel >= thumbPos && clickRel < thumbPos+thumbSize {
		return node // Click is on thumb
	}
	return nil
}

// handleScrollbarDrag processes mouse move during scrollbar drag.
func (e *Engine) handleScrollbarDrag(x, y int) {
	node := e.scrollbarDragNode
	bw := 0
	if hasBorder(node.Style) {
		bw = 1
	}
	innerY1 := node.Y + bw + node.Style.PaddingTop
	innerY2 := node.Y + node.H - bw - node.Style.PaddingBottom
	visibleH := innerY2 - innerY1
	if visibleH <= 0 {
		return
	}

	totalH := node.ScrollHeight
	maxScroll := computeMaxScrollY(node)
	if maxScroll <= 0 {
		return
	}

	thumbSize := visibleH * visibleH / totalH
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > visibleH {
		thumbSize = visibleH
	}
	trackSpace := visibleH - thumbSize
	if trackSpace <= 0 {
		return
	}

	// Convert mouse Y delta to scroll delta
	dy := y - e.scrollbarDragStartY
	scrollDelta := dy * maxScroll / trackSpace
	newScrollY := e.scrollbarDragStartScrollY + scrollDelta

	if newScrollY < 0 {
		newScrollY = 0
	}
	if newScrollY > maxScroll {
		newScrollY = maxScroll
	}

	if newScrollY != node.ScrollY {
		node.ScrollY = newScrollY
		node.TargetScrollY = node.ScrollY
		node.PaintDirty = true
		e.needsRender = true
	}
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
	scrollNode.TargetScrollY = scrollNode.ScrollY
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

// ScrollNodeByID finds a node by its ID across all layers and adjusts its ScrollY by delta.
// Returns the new ScrollY value. The target must be a node with overflow:"scroll" style.
// If the ID matches a non-scroll node (e.g., a component placeholder), it searches children
// for the actual scroll container with the same ID.
func (e *Engine) ScrollNodeByID(id string, delta int) int {
	for _, layer := range e.layers {
		if layer.Root != nil {
			if found := e.findNodeByID(layer.Root, id); found != nil {
				// If the found node is not a scroll container, search its children
				// (e.g., component placeholder wrapping a scroll vbox with the same ID)
				scrollNode := found
				if scrollNode.Style.Overflow != "scroll" {
					scrollNode = e.findScrollNodeByID(scrollNode, id)
					if scrollNode == nil {
						return 0
					}
				}
				maxScroll := computeMaxScrollY(scrollNode)
				newSY := scrollNode.ScrollY + delta
				if newSY < 0 {
					newSY = 0
				}
				if newSY > maxScroll {
					newSY = maxScroll
				}
				if newSY != scrollNode.ScrollY {
					scrollNode.ScrollY = newSY
					scrollNode.TargetScrollY = scrollNode.ScrollY
					scrollNode.PaintDirty = true
					e.needsRender = true
				}
				return newSY
			}
		}
	}
	return 0
}

// ScrollNodeByIDH finds a node by its ID and adjusts its ScrollX by delta.
// Returns the new ScrollX value.
func (e *Engine) ScrollNodeByIDH(id string, delta int) int {
	for _, layer := range e.layers {
		if layer.Root != nil {
			if found := e.findNodeByID(layer.Root, id); found != nil {
				scrollNode := found
				if scrollNode.Style.Overflow != "scroll" {
					scrollNode = e.findScrollNodeByID(scrollNode, id)
					if scrollNode == nil {
						return 0
					}
				}
				maxScroll := computeMaxScrollX(scrollNode)
				newSX := scrollNode.ScrollX + delta
				if newSX < 0 {
					newSX = 0
				}
				if newSX > maxScroll {
					newSX = maxScroll
				}
				if newSX != scrollNode.ScrollX {
					scrollNode.ScrollX = newSX
					scrollNode.PaintDirty = true
					e.needsRender = true
				}
				return newSX
			}
		}
	}
	return 0
}

// findScrollNodeByID searches children of node for a node with the given ID
// that has overflow:"scroll" style.
func (e *Engine) findScrollNodeByID(node *Node, id string) *Node {
	for _, child := range node.Children {
		if child.ID == id && child.Style.Overflow == "scroll" {
			return child
		}
		if found := e.findScrollNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// findNodeByID searches the tree rooted at node for a node with the given ID.
// Returns nil if not found.
func (e *Engine) findNodeByID(node *Node, id string) *Node {
	if node.ID == id {
		return node
	}
	for _, child := range node.Children {
		if found := e.findNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// computeMaxScrollY calculates the maximum scroll offset for a scroll container.
// Uses the stored ScrollHeight (set by layout) minus the visible content height.
func computeMaxScrollY(node *Node) int {
	bw := 0
	if hasBorder(node.Style) {
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

func computeMaxScrollX(node *Node) int {
	bw := 0
	if hasBorder(node.Style) {
		bw = 1
	}
	contentW := node.W - 2*bw - node.Style.PaddingLeft - node.Style.PaddingRight
	if contentW <= 0 {
		return 0
	}
	maxScroll := node.ScrollWidth - contentW
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// autoScrollX adjusts a scroll container's ScrollX by delta, clamped to [0, maxScroll].
func (e *Engine) autoScrollX(node *Node, delta int) {
	maxScroll := computeMaxScrollX(node)
	if maxScroll <= 0 {
		return // content fits, no scrolling needed
	}

	const step = 3 // scroll 3 columns per wheel tick
	newScrollX := node.ScrollX + delta*step

	// Clamp
	if newScrollX < 0 {
		newScrollX = 0
	}
	if newScrollX > maxScroll {
		newScrollX = maxScroll
	}

	if newScrollX == node.ScrollX {
		return // no change
	}

	node.ScrollX = newScrollX
	node.PaintDirty = true
	e.needsRender = true
}

// HandleScrollH handles horizontal scroll events (e.g. Shift+wheel).
func (e *Engine) HandleScrollH(x, y, delta int) {
	if len(e.layers) == 0 {
		return
	}
	hitNode, _ := e.hitTestLayers(x, y)
	scrollNode := findScrollableAncestor(hitNode)
	if scrollNode != nil {
		e.autoScrollX(scrollNode, delta)
	}

	// Re-evaluate hover after horizontal scroll offset changed.
	e.HandleMouseMove(x, y)
}

// EventResult holds the result of calling a Lua event handler.
// Handlers can call event.stopPropagation() and event.preventDefault().
type EventResult struct {
	Stopped          bool // stopPropagation was called
	DefaultPrevented bool // preventDefault was called
}

// reportEventError logs and notifies Lua about an event handler error.
func (e *Engine) reportEventError(errMsg string, handlerType string) {
	fmt.Fprintf(os.Stderr, "[lumina] event handler error (%s): %s\n", handlerType, errMsg)
	log.Printf("[lumina] event handler error (%s): %s", handlerType, errMsg)
	if e.onErrorRef != 0 {
		L := e.L
		L.RawGetI(lua.RegistryIndex, e.onErrorRef)
		L.PushString(errMsg)
		L.PushString("event:" + handlerType)
		L.PCall(2, 0, 0) // ignore errors in error handler itself
	}
}

// callLuaRef calls a Lua function by registry ref with an event table {x=x, y=y, button=button}.
// The event table includes stopPropagation() and preventDefault() methods.
func (e *Engine) callLuaRef(ref LuaRef, x, y int) EventResult {
	result := EventResult{}
	L := e.L
	L.RawGetI(lua.RegistryIndex, ref)
	if !L.IsFunction(-1) {
		L.Pop(1)
		return result
	}
	// Push event table: {x=x, y=y, button=button, stopPropagation=func, preventDefault=func}
	L.NewTable()
	tblIdx := L.AbsIndex(-1)
	L.PushInteger(int64(x))
	L.SetField(tblIdx, "x")
	L.PushInteger(int64(y))
	L.SetField(tblIdx, "y")
	L.PushString(e.currentButton)
	L.SetField(tblIdx, "button")
	// event.stopPropagation() — Go closure that sets result.Stopped
	L.PushFunction(func(L *lua.State) int {
		result.Stopped = true
		return 0
	})
	L.SetField(tblIdx, "stopPropagation")
	// event.preventDefault() — Go closure that sets result.DefaultPrevented
	L.PushFunction(func(L *lua.State) int {
		result.DefaultPrevented = true
		return 0
	})
	L.SetField(tblIdx, "preventDefault")
	if status := L.PCall(1, 0, 0); status != lua.OK {
		errMsg, _ := L.ToString(-1)
		L.Pop(1)
		e.reportEventError(errMsg, "onClick")
	}
	return result
}

// callLuaRefKey calls a Lua function with an event table {key=key}.
// The event table includes stopPropagation() and preventDefault() methods.
func (e *Engine) callLuaRefKey(ref LuaRef, key string) EventResult {
	result := EventResult{}
	L := e.L
	L.RawGetI(lua.RegistryIndex, ref)
	if !L.IsFunction(-1) {
		L.Pop(1)
		return result
	}
	L.NewTable()
	tblIdx := L.AbsIndex(-1)
	L.PushString(key)
	L.SetField(tblIdx, "key")
	L.PushFunction(func(L *lua.State) int {
		result.Stopped = true
		return 0
	})
	L.SetField(tblIdx, "stopPropagation")
	L.PushFunction(func(L *lua.State) int {
		result.DefaultPrevented = true
		return 0
	})
	L.SetField(tblIdx, "preventDefault")
	if status := L.PCall(1, 0, 0); status != lua.OK {
		errMsg, _ := L.ToString(-1)
		L.Pop(1)
		e.reportEventError(errMsg, "onKeyDown")
	}
	return result
}

// callLuaRefScroll calls a Lua function with an event table {delta=delta, key="up"/"down",
// scrollY=scrollY, scrollHeight=scrollHeight}.
func (e *Engine) callLuaRefScroll(ref LuaRef, delta int, scrollNode *Node) {
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
	// Include absolute scrollY and scrollHeight so handlers can compute visible range.
	// Use TargetScrollY so Lua handlers see the intended scroll position (smooth scroll
	// animates ScrollY toward TargetScrollY, but Lua logic should use the target).
	if scrollNode != nil {
		L.PushInteger(int64(scrollNode.TargetScrollY))
		L.SetField(tblIdx, "scrollY")
		L.PushInteger(int64(scrollNode.ScrollHeight))
		L.SetField(tblIdx, "scrollHeight")
		// Add visibleH and maxScroll for infinite scroll support
		maxScroll := computeMaxScrollY(scrollNode)
		L.PushInteger(int64(maxScroll))
		L.SetField(tblIdx, "maxScroll")
		bw := 0
		if hasBorder(scrollNode.Style) {
			bw = 1
		}
		visibleH := scrollNode.H - 2*bw - scrollNode.Style.PaddingTop - scrollNode.Style.PaddingBottom
		if visibleH < 0 {
			visibleH = 0
		}
		L.PushInteger(int64(visibleH))
		L.SetField(tblIdx, "visibleH")
	}
	if status := L.PCall(1, 0, 0); status != lua.OK {
		errMsg, _ := L.ToString(-1)
		L.Pop(1)
		e.reportEventError(errMsg, "onScroll")
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

// findKeyHandlerInSubtree finds the first node with OnKeyDown in the subtree (DFS).
// Used when a focused component placeholder has no handler itself but its rendered
// children do (e.g. Lua Textarea renders a vbox with onKeyDown inside a component).
func findKeyHandlerInSubtree(node *Node) *Node {
	if node == nil {
		return nil
	}
	if node.OnKeyDown != 0 {
		return node
	}
	for _, child := range node.Children {
		if found := findKeyHandlerInSubtree(child); found != nil {
			return found
		}
	}
	return nil
}

// callLuaRefSimple calls a Lua function by registry ref with an event table
// containing only stopPropagation() and preventDefault() methods.
func (e *Engine) callLuaRefSimple(ref LuaRef) EventResult {
	result := EventResult{}
	L := e.L
	L.RawGetI(lua.RegistryIndex, ref)
	if !L.IsFunction(-1) {
		L.Pop(1)
		return result
	}
	// Push event table with just the control methods
	L.NewTable()
	tblIdx := L.AbsIndex(-1)
	L.PushFunction(func(L *lua.State) int {
		result.Stopped = true
		return 0
	})
	L.SetField(tblIdx, "stopPropagation")
	L.PushFunction(func(L *lua.State) int {
		result.DefaultPrevented = true
		return 0
	})
	L.SetField(tblIdx, "preventDefault")
	if status := L.PCall(1, 0, 0); status != lua.OK {
		errMsg, _ := L.ToString(-1)
		L.Pop(1)
		e.reportEventError(errMsg, "event")
	}
	return result
}
