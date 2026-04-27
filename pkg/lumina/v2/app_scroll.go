package v2

import (
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// scrollStep is the number of rows to scroll per mouse wheel tick.
const scrollStep = 3

// handleScrollEvent handles built-in mouse wheel scrolling for overflow="scroll"
// containers. It hit-tests to find the VNode under the mouse, walks up the tree
// to find the nearest scroll container, adjusts its ScrollY, and marks the
// component dirty. Returns true if the event was handled.
func (a *App) handleScrollEvent(e *event.Event) bool {
	// Find which component and VNode is under the mouse.
	for _, comp := range a.manager.GetAll() {
		if comp.VNodeTree() == nil {
			continue
		}
		r := comp.Rect()
		if e.X < r.X || e.X >= r.X+r.W || e.Y < r.Y || e.Y >= r.Y+r.H {
			continue
		}
		// Walk the VNode tree to find the deepest scroll container containing (e.X, e.Y).
		scrollNode := findScrollContainer(comp.VNodeTree(), e.X, e.Y)
		if scrollNode == nil {
			continue
		}

		// Calculate content area dimensions.
		style := scrollNode.Style
		borderW := 0
		if style.Border == "single" || style.Border == "double" || style.Border == "rounded" {
			borderW = 1
		}
		padTop := style.PaddingTop
		if padTop == 0 {
			padTop = style.Padding
		}
		padBottom := style.PaddingBottom
		if padBottom == 0 {
			padBottom = style.Padding
		}
		padLeft := style.PaddingLeft
		if padLeft == 0 {
			padLeft = style.Padding
		}
		padRight := style.PaddingRight
		if padRight == 0 {
			padRight = style.Padding
		}

		contentH := scrollNode.H - 2*borderW - padTop - padBottom
		if contentH <= 0 {
			continue
		}

		// Calculate total content height from children.
		absContentY := scrollNode.Y + borderW + padTop
		_ = padLeft
		_ = padRight
		totalContentH := 0
		for _, child := range scrollNode.Children {
			childBottom := (child.Y - absContentY) + child.H
			if childBottom > totalContentH {
				totalContentH = childBottom
			}
		}

		maxScroll := totalContentH - contentH
		if maxScroll < 0 {
			maxScroll = 0
		}

		// Adjust scroll offset.
		oldScrollY := scrollNode.ScrollY
		switch e.Key {
		case "up":
			scrollNode.ScrollY -= scrollStep
		case "down":
			scrollNode.ScrollY += scrollStep
		default:
			continue
		}

		// Clamp.
		if scrollNode.ScrollY < 0 {
			scrollNode.ScrollY = 0
		}
		if scrollNode.ScrollY > maxScroll {
			scrollNode.ScrollY = maxScroll
		}

		if scrollNode.ScrollY != oldScrollY {
			// Mark component dirty so it repaints with new scroll offset.
			comp.MarkDirty()
			// Also dispatch the scroll event to Lua handlers.
			a.dispatcher.Dispatch(e)
			return true
		}
		return false
	}
	return false
}

// findScrollContainer walks the VNode tree and returns the deepest VNode
// with overflow="scroll" that contains the point (x, y).
func findScrollContainer(vnode *layout.VNode, x, y int) *layout.VNode {
	if vnode == nil {
		return nil
	}
	// Check if point is inside this VNode.
	if x < vnode.X || x >= vnode.X+vnode.W || y < vnode.Y || y >= vnode.Y+vnode.H {
		return nil
	}
	// Check children in reverse order (deepest first).
	for i := len(vnode.Children) - 1; i >= 0; i-- {
		if found := findScrollContainer(vnode.Children[i], x, y); found != nil {
			return found
		}
	}
	// Check this node.
	if vnode.Style.Overflow == "scroll" {
		return vnode
	}
	return nil
}
