// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

// HitTestFrame uses Cell.OwnerNode for O(1) hit testing.
// Returns the VNode that owns the cell at screen position (x, y), or nil.
func HitTestFrame(frame *Frame, x, y int) *VNode {
	if frame == nil || x < 0 || y < 0 || x >= frame.Width || y >= frame.Height {
		return nil
	}
	return frame.Cells[y][x].OwnerNode
}

// GlobalToLocal converts global screen coordinates to local coordinates
// relative to a VNode's content area (inside border + padding).
func GlobalToLocal(vnode *VNode, globalX, globalY int) (localX, localY int) {
	borderW := 0
	if vnode.Style.Border != "" {
		borderW = 1
	}
	localX = globalX - vnode.X - borderW - vnode.Style.PaddingLeft
	localY = globalY - vnode.Y - borderW - vnode.Style.PaddingTop
	return
}

// LocalToGlobal converts local coordinates (relative to VNode content area)
// to global screen coordinates.
func LocalToGlobal(vnode *VNode, localX, localY int) (globalX, globalY int) {
	borderW := 0
	if vnode.Style.Border != "" {
		borderW = 1
	}
	globalX = localX + vnode.X + borderW + vnode.Style.PaddingLeft
	globalY = localY + vnode.Y + borderW + vnode.Style.PaddingTop
	return
}

// ContainsPoint checks if a global point is within a VNode's bounding box.
func ContainsPoint(vnode *VNode, x, y int) bool {
	return x >= vnode.X && x < vnode.X+vnode.W &&
		y >= vnode.Y && y < vnode.Y+vnode.H
}

// ContainsPointContent checks if a global point is within a VNode's content area
// (inside border + padding).
func ContainsPointContent(vnode *VNode, x, y int) bool {
	borderW := 0
	if vnode.Style.Border != "" {
		borderW = 1
	}
	cx := vnode.X + borderW + vnode.Style.PaddingLeft
	cy := vnode.Y + borderW + vnode.Style.PaddingTop
	cw := vnode.W - 2*borderW - vnode.Style.PaddingLeft - vnode.Style.PaddingRight
	ch := vnode.H - 2*borderW - vnode.Style.PaddingTop - vnode.Style.PaddingBottom
	if cw < 0 || ch < 0 {
		return false
	}
	return x >= cx && x < cx+cw && y >= cy && y < cy+ch
}
