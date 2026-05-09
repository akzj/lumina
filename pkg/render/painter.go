package render

import (
	"fmt"
	"os"
)

// painter.go — paint engine primitives (vbox, hbox, text, component) to CellBuffer.

// PaintFull paints the entire node tree into the buffer.
// The buffer is cleared first. Used for initial render.
// Clears all PaintDirty flags after painting.
func PaintFull(buf *CellBuffer, root *Node) {
	if buf == nil || root == nil {
		return
	}
	buf.Clear()
	paintNode(buf, root)
	clearPaintDirty(root)
}

// clearPaintDirty recursively clears PaintDirty flags on all nodes.
func clearPaintDirty(node *Node) {
	if node == nil {
		return
	}
	node.PaintDirty = false
	node.PositionChanged = false
	for _, child := range node.Children {
		clearPaintDirty(child)
	}
}

// clearPaintDirtyBelow clears PaintDirty on all descendants (not the node itself).
func clearPaintDirtyBelow(node *Node) {
	for _, child := range node.Children {
		child.PaintDirty = false
		child.PositionChanged = false
		clearPaintDirtyBelow(child)
	}
}

// PaintDirty paints only PaintDirty nodes into the buffer.
// Does NOT clear the buffer — only overwrites dirty regions.
// Clears PaintDirty flags after painting.
// If the root node itself is dirty, does a full repaint — this handles
// the case where the root component re-rendered (e.g., bringToFront
// reordered children). Full repaint ensures all absolute-positioned
// children are correctly painted in z-order.
func PaintDirty(buf *CellBuffer, root *Node) {
	if buf == nil || root == nil {
		return
	}
	if root.PaintDirty {
		// Use background-aware clear so any cells not covered by paintNode
		// still show the root's background color (not terminal default).
		bg := root.Style.Background
		if bg == "" {
			bg = findAncestorBackground(root)
		}
		clearRectWithBG(buf, root.X, root.Y, root.W, root.H, bg)
		paintNode(buf, root)
		clearPaintDirty(root)
		return
	}
	paintDirtyWalk(buf, root)
}

// PaintDirtyOverlay paints dirty nodes for overlay layers.
// Does NOT clear regions — overlays paint on top of the main layer.
// This prevents overlay layer ClearRect from wiping main layer content.
func PaintDirtyOverlay(buf *CellBuffer, root *Node) {
	if buf == nil || root == nil {
		return
	}
	if root.PaintDirty {
		paintNode(buf, root)
		clearPaintDirty(root)
		return
	}
	paintDirtyWalk(buf, root)
}

func paintDirtyWalk(buf *CellBuffer, node *Node) {
	if node == nil {
		return
	}
	// display:none nodes are invisible — skip painting and clear dirty flags
	// to prevent hidden components from triggering parent container repaints.
	if node.Style.Display == "none" {
		if node.PaintDirty {
			clearPaintDirty(node)
		}
		return
	}

	if node.PaintDirty {
		// If any ancestor has children with fixed/absolute positioning,
		// escalate to that ancestor for full repaint to maintain correct z-order.
		// EXCEPTION: scroll containers clip their own content and don't affect
		// z-order of siblings — no need to escalate to a larger ancestor which
		// would unnecessarily clear/repaint unrelated sibling panels.
		if node.Style.Overflow != "scroll" && node.Style.Overflow != "hidden" {
			if overlayParent := findOverlayAncestor(node); overlayParent != nil && !overlayParent.PaintDirty {
				overlayParent.PaintDirty = true
				bg := findAncestorBackground(overlayParent)
				clearRectWithBG(buf, overlayParent.X, overlayParent.Y, overlayParent.W, overlayParent.H, bg)
				paintNode(buf, overlayParent)
				overlayParent.PaintDirty = false
				clearPaintDirtyBelow(overlayParent)
				return
			}
		}

		// If this node is inside a scroll container (at any ancestor level),
		// escalate to that scroll ancestor so paintScrollChildren handles
		// coordinate transformation and clipping correctly.
		// This is critical for nested scroll: inner scroll containers inside
		// outer scroll containers need the outer to repaint with correct offsets.
		scrollAncestor := findScrollableAncestor(node.Parent)
		if scrollAncestor != nil {
			if !scrollAncestor.PaintDirty {
				scrollAncestor.PaintDirty = true
				bg := findAncestorBackground(scrollAncestor)
				clearRectWithBG(buf, scrollAncestor.X, scrollAncestor.Y, scrollAncestor.W, scrollAncestor.H, bg)
				paintNode(buf, scrollAncestor)
				scrollAncestor.PaintDirty = false
				clearPaintDirtyBelow(scrollAncestor)
				// Repaint overlapping windows above this scroll container
				repaintOverlappingSiblings(buf, scrollAncestor)
			} else {
				// Scroll ancestor is already dirty — it will be repainted later
				// in the walk with proper scroll offset/clip. Just clear our
				// dirty flag and skip; the ancestor's repaint will cover us.
				node.PaintDirty = false
				clearPaintDirtyBelow(node)
			}
			return
		}
		// If this node is inside an overflow:hidden container, escalate to that
		// container so its full area gets cleared before repainting. This prevents
		// residual content when children change (e.g. tab switches).
		hiddenAncestor := findHiddenAncestor(node.Parent)
		if hiddenAncestor != nil {
			if !hiddenAncestor.PaintDirty {
				// Normal escalation: repaint the entire hidden ancestor
				hiddenAncestor.PaintDirty = true
				bg := findAncestorBackground(hiddenAncestor)
				clearRectWithBG(buf, hiddenAncestor.X, hiddenAncestor.Y, hiddenAncestor.W, hiddenAncestor.H, bg)
				paintNode(buf, hiddenAncestor)
				hiddenAncestor.PaintDirty = false
				clearPaintDirtyBelow(hiddenAncestor)
			} else {
				// Hidden ancestor is already dirty — it will be repainted later
				// in the walk. Just clear our dirty flag and skip; the ancestor's
				// repaint will cover us with proper clipping.
				node.PaintDirty = false
				clearPaintDirtyBelow(node)
			}
			return
		}

		// Component placeholders have stacked bounds that may overlap siblings'
		// absolute-positioned children. Escalate to parent so all siblings
		// get repainted after ClearRect.
		if node.Type == "component" {
			parent := findRepaintParent(node)
			if parent != nil {
				if node.PositionChanged {
					bg := findAncestorBackground(node)
					clearRectWithBG(buf, node.OldX, node.OldY, node.OldW, node.OldH, bg)
					node.PositionChanged = false
				}
				node.PaintDirty = false
				if !parent.PaintDirty {
					parent.PaintDirty = true
					bg := findAncestorBackground(parent)
					clearRectWithBG(buf, parent.X, parent.Y, parent.W, parent.H, bg)
					paintNode(buf, parent)
					parent.PaintDirty = false
					clearPaintDirtyBelow(parent)
				}
				return
			}
		}
		// For absolute-positioned nodes that moved, escalate to parent container
		// so all overlapping siblings (other windows) get repainted too.
		if node.PositionChanged && node.Style.Position == "absolute" {
			parent := findRepaintParent(node)
			if parent != nil && !parent.PaintDirty {
				bg := findAncestorBackground(node)
				clearRectWithBG(buf, node.OldX, node.OldY, node.OldW, node.OldH, bg)
				node.PositionChanged = false
				node.PaintDirty = false
				parent.PaintDirty = true
				bg = findAncestorBackground(parent)
				clearRectWithBG(buf, parent.X, parent.Y, parent.W, parent.H, bg)
				paintNode(buf, parent)
				parent.PaintDirty = false
				clearPaintDirtyBelow(parent)
				return
			}
		}
		// If node moved/resized, clear the old region to avoid ghost artifacts
		if node.PositionChanged {
			bg := findAncestorBackground(node)
			clearRectWithBG(buf, node.OldX, node.OldY, node.OldW, node.OldH, bg)
			node.PositionChanged = false
		}
		// Clear this node's region first, then repaint
		bg := findAncestorBackground(node)
		clearRectWithBG(buf, node.X, node.Y, node.W, node.H, bg)
		paintNode(buf, node)
		node.PaintDirty = false
		// Clear all descendants' PaintDirty flags — paintNode already painted them
		clearPaintDirtyBelow(node)
		// If this node is inside an absolute-positioned window, repaint any
		// overlapping siblings that are later in z-order (above this window).
		repaintOverlappingSiblings(buf, node)
		return
	}

	// Not dirty — check children
	for _, child := range node.Children {
		paintDirtyWalk(buf, child)
	}
}

// repaintOverlappingSiblings finds the absolute/fixed-positioned ancestor of node
// (the window), then repaints any later siblings in z-order that overlap the
// repainted area. This prevents scroll repaints from leaking over windows above.
// If the node is a normal flow child (not inside absolute/fixed), it falls back
// to checking later siblings for any fixed/absolute overlays that need repainting.
func repaintOverlappingSiblings(buf *CellBuffer, node *Node) {
	// Walk up to find the absolute/fixed-positioned ancestor (the window node)
	var absNode *Node
	for n := node; n != nil; n = n.Parent {
		if n.Style.Position == "absolute" || n.Style.Position == "fixed" {
			absNode = n
			break
		}
	}
	if absNode == nil {
		// Not inside an absolute/fixed node — this is a normal flow child.
		// Check later siblings for fixed/absolute overlays that overlap.
		repaintFixedOverlappingSiblings(buf, node)
		return
	}
	if absNode.Parent == nil {
		return
	}

	// Find the container that holds all windows as children.
	// Walk through component wrappers to find the actual parent container.
	parent := absNode.Parent
	containerChild := absNode // the child of parent that contains our window
	for parent != nil && parent.Type == "component" && parent.Parent != nil {
		containerChild = parent
		parent = parent.Parent
	}

	// Find our index in parent's children
	idx := -1
	for i, ch := range parent.Children {
		if ch == containerChild {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}

	// Use the ACTUAL absolute/fixed-positioned window bounds for overlap check
	rx, ry, rw, rh := absNode.X, absNode.Y, absNode.W, absNode.H
	if rw <= 0 || rh <= 0 {
		rx, ry, rw, rh = node.X, node.Y, node.W, node.H
	}

	// Repaint siblings that paint ABOVE us (higher z-order) and overlap.
	// With z-index: a sibling is "above" if it has higher ZIndex, or same ZIndex
	// but comes later in the array. Use paint-order to determine this.
	ordered := paintOrderChildren(parent.Children)
	// Find our position in paint order
	paintIdx := -1
	for i, ch := range ordered {
		if ch == containerChild {
			paintIdx = i
			break
		}
	}
	if paintIdx < 0 {
		return
	}
	// Repaint everything that paints after us (above us visually)
	for i := paintIdx + 1; i < len(ordered); i++ {
		sibling := ordered[i]
		// Find the absolute/fixed-positioned child INSIDE the sibling (component wrapper)
		sibAbs := findAbsoluteOrFixedChild(sibling)
		if sibAbs != nil {
			if rectsOverlap(rx, ry, rw, rh, sibAbs.X, sibAbs.Y, sibAbs.W, sibAbs.H) {
				bg := findAncestorBackground(sibAbs)
				clearRectWithBG(buf, sibAbs.X, sibAbs.Y, sibAbs.W, sibAbs.H, bg)
				paintNode(buf, sibAbs)
			}
		} else {
			// Fallback: use sibling's own bounds
			if rectsOverlap(rx, ry, rw, rh, sibling.X, sibling.Y, sibling.W, sibling.H) {
				bg := findAncestorBackground(sibling)
				clearRectWithBG(buf, sibling.X, sibling.Y, sibling.W, sibling.H, bg)
				paintNode(buf, sibling)
			}
		}
	}
}

// repaintFixedOverlappingSiblings handles the case where a normal flow child
// is repainted and needs to check if any later siblings with position:fixed
// or position:absolute overlap and need repainting.
func repaintFixedOverlappingSiblings(buf *CellBuffer, node *Node) {
	parent := node.Parent
	if parent == nil {
		return
	}
	// Walk through component wrappers to find the actual parent container
	target := node
	for parent != nil && parent.Type == "component" && parent.Parent != nil {
		target = parent
		parent = parent.Parent
	}

	// Find our index in parent's children
	idx := -1
	for i, ch := range parent.Children {
		if ch == target {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}

	rx, ry, rw, rh := node.X, node.Y, node.W, node.H

	// Repaint siblings that paint above us and are fixed/absolute and overlap.
	// Use paint order (z-index sorted) to find siblings that paint after target.
	ordered := paintOrderChildren(parent.Children)
	paintIdx := -1
	for i, ch := range ordered {
		if ch == target {
			paintIdx = i
			break
		}
	}
	if paintIdx < 0 {
		return
	}
	for i := paintIdx + 1; i < len(ordered); i++ {
		sibling := ordered[i]
		sibFixed := findAbsoluteOrFixedChild(sibling)
		if sibFixed != nil {
			if rectsOverlap(rx, ry, rw, rh, sibFixed.X, sibFixed.Y, sibFixed.W, sibFixed.H) {
				bg := findAncestorBackground(sibFixed)
				clearRectWithBG(buf, sibFixed.X, sibFixed.Y, sibFixed.W, sibFixed.H, bg)
				paintNode(buf, sibFixed)
			}
		}
	}
}

// findAbsoluteOrFixedChild finds the first absolute/fixed-positioned descendant within a node.
// This traverses through component wrappers to find the actual window/overlay vbox.
func findAbsoluteOrFixedChild(node *Node) *Node {
	if node.Style.Position == "absolute" || node.Style.Position == "fixed" {
		return node
	}
	for _, ch := range node.Children {
		if found := findAbsoluteOrFixedChild(ch); found != nil {
			return found
		}
	}
	return nil
}

// rectsOverlap returns true if two rectangles overlap.
func rectsOverlap(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// parentHasOverlayChildren returns true if the node has any child (direct or through
// component wrappers) with position:fixed or position:absolute.
// findOverlayAncestor walks up from node to find the nearest ancestor that has
// children with position:fixed or position:absolute. Returns that ancestor.
// This is used to escalate dirty painting to ensure correct z-order when
// overlays are present.
func findOverlayAncestor(node *Node) *Node {
	for n := node.Parent; n != nil; n = n.Parent {
		if n.Type == "component" {
			continue // skip component wrappers
		}
		if parentHasOverlayChildren(n) {
			return n
		}
	}
	return nil
}
// findHiddenAncestor walks up from node.Parent to find the nearest ancestor
// with overflow:"hidden". This is used to escalate dirty painting to the
// hidden container so its full area gets cleared before repainting.
func findHiddenAncestor(node *Node) *Node {
	for n := node; n != nil; n = n.Parent {
		if n.Type == "component" {
			continue // skip component wrappers
		}
		if n.Style.Overflow == "hidden" {
			return n
		}
	}
	return nil
}

// findAncestorBackground walks up the tree to find the nearest ancestor
// with a non-empty background color. This simulates CSS background inheritance.
func findAncestorBackground(node *Node) string {
	for n := node.Parent; n != nil; n = n.Parent {
		if n.Style.Background != "" {
			return n.Style.Background
		}
	}
	return ""
}

// clearRectWithBG clears a rectangular area and fills it with the given
// background color. If bg is empty, falls back to plain ClearRect.
func clearRectWithBG(buf *CellBuffer, x, y, w, h int, bg string) {
	if bg == "" {
		buf.ClearRect(x, y, w, h)
		return
	}
	for row := y; row < y+h && row < buf.Height(); row++ {
		if row < 0 {
			continue
		}
		for col := x; col < x+w && col < buf.Width(); col++ {
			if col < 0 {
				continue
			}
			buf.SetChar(col, row, ' ', "", bg, false)
		}
	}
}


func parentHasOverlayChildren(node *Node) bool {
	for _, child := range node.Children {
		if child.Style.Position == "fixed" || child.Style.Position == "absolute" {
			return true
		}
		// Check through component wrappers
		if child.Type == "component" {
			if findAbsoluteOrFixedChild(child) != nil {
				return true
			}
		}
	}
	return false
}

// findRepaintParent walks up from node to find the first non-component ancestor.
// This is the container that holds all overlapping windows.
func findRepaintParent(node *Node) *Node {
	for n := node.Parent; n != nil; n = n.Parent {
		if n.Type != "component" {
			return n
		}
	}
	return nil
}

// paintDepth tracks recursion depth in paintNode to prevent stack overflow
// from cycles in the node tree (e.g., after a hot reload).
var paintDepth int

const maxPaintDepth = 500

// unboundedClip is used as a "no clip" sentinel in clipped paint functions.
const unboundedClip = 1 << 30

// mergeHoverStyle merges hover overrides onto the base style.
// Only non-zero/non-empty hover fields override the base.
func mergeHoverStyle(base Style, hover *Style) Style {
	if hover == nil {
		return base
	}
	result := base
	if hover.Foreground != "" {
		result.Foreground = hover.Foreground
	}
	if hover.Background != "" {
		result.Background = hover.Background
	}
	if hover.Bold {
		result.Bold = true
	}
	if hover.Dim {
		result.Dim = true
	}
	if hover.Underline {
		result.Underline = true
	}
	if hover.Italic {
		result.Italic = true
	}
	if hover.Strikethrough {
		result.Strikethrough = true
	}
	if hover.Inverse {
		result.Inverse = true
	}
	if hover.Border != "" {
		result.Border = hover.Border
	}
	if hover.BorderColor != "" {
		result.BorderColor = hover.BorderColor
	}
	return result
}

func paintNode(buf *CellBuffer, node *Node) {
	if node == nil || node.W <= 0 || node.H <= 0 {
		return
	}
	if node.Style.Display == "none" || node.Style.Visibility == "hidden" {
		return
	}
	paintDepth++
	if paintDepth > maxPaintDepth {
		paintDepth--
		return
	}
	defer func() { paintDepth-- }()


	// DEBUG: detect painting of nodes whose parent is nil (detached/root nodes)
	// that are NOT the layer root (which legitimately has no parent)
	if node.Parent == nil && node.X > 0 && node.W < 200 {
		fmt.Fprintf(os.Stderr, "DEBUG_DETACHED_PAINT: type=%q id=%q X=%d W=%d H=%d overflow=%q numChildren=%d\n",
			node.Type, node.ID, node.X, node.W, node.H, node.Style.Overflow, len(node.Children))
	}

	// Apply hover style: temporarily swap node.Style with merged version
	var savedStyle Style
	if node.Hovered && node.HoverStyle != nil {
		savedStyle = node.Style
		node.Style = mergeHoverStyle(node.Style, node.HoverStyle)
		defer func() { node.Style = savedStyle }()
	}

	switch node.Type {
	case "text":
		paintText(buf, node)
	case "box", "vbox", "hbox":
		paintBox(buf, node)
	case "component":
		// Component placeholder: transparent container, just paint children
		for _, child := range paintOrderChildren(node.Children) {
			paintNode(buf, child)
		}
	}
}

func paintBox(buf *CellBuffer, node *Node) {
	// 1. Fill background
	if node.Style.Background != "" {
		for y := node.Y; y < node.Y+node.H; y++ {
			for x := node.X; x < node.X+node.W; x++ {
				buf.SetChar(x, y, ' ', "", node.Style.Background, false)
			}
		}
	}

	// 2. Draw border
	if hasBorder(node.Style) {
		paintBorder(buf, node)
	}

	// 3. Paint children (with scroll offset / clipping if applicable)
	if node.Style.Overflow == "scroll" {
		paintScrollChildren(buf, node)
	} else if node.Style.Overflow == "hidden" {
		paintHiddenChildren(buf, node)
	} else {
		for _, child := range paintOrderChildren(node.Children) {
			paintNode(buf, child)
		}
	}
}

// paintHiddenChildren paints children clipped to the node's content area.
// Unlike scroll, there is no scroll offset — children are simply clipped.
// paintHiddenChildren paints overflow:hidden children clipped to the content area.
// Delegates to paintHiddenChildrenClipped with unbounded outer clip.
func paintHiddenChildren(buf *CellBuffer, node *Node) {
	paintHiddenChildrenClipped(buf, node, 0, 0, unboundedClip, unboundedClip, 0, 0)
}

// paintScrollChildren paints children with a scroll offset, clipping to the content area.
// Delegates to paintScrollChildrenClipped with unbounded outer clip.
func paintScrollChildren(buf *CellBuffer, node *Node) {
	paintScrollChildrenClipped(buf, node, 0, 0, unboundedClip, unboundedClip, 0, 0)
}

// paintScrollChildrenClipped paints scroll container children with both the
// inner scroll offset AND an outer clip rect (from a parent scroll container).
// The effective clip is the intersection of the outer clip and the inner content area.
// parentOffsetY is the cumulative paint offset from ancestor scroll containers.
func paintScrollChildrenClipped(buf *CellBuffer, node *Node, outerClipX1, outerClipY1, outerClipX2, outerClipY2, parentOffsetX, parentOffsetY int) {
	// Clamp scrollY to valid range
	maxScrollY := computeMaxScrollY(node)
	if node.ScrollY > maxScrollY {
		node.ScrollY = maxScrollY
	}
	if node.ScrollY < 0 {
		node.ScrollY = 0
	}

	// Clamp scrollX to valid range
	maxScrollX := computeMaxScrollX(node)
	if node.ScrollX > maxScrollX {
		node.ScrollX = maxScrollX
	}
	if node.ScrollX < 0 {
		node.ScrollX = 0
	}

	scrollY := node.ScrollY
	scrollX := node.ScrollX

	// Combine parent offset with this container's scroll offset for children
	childOffsetX := parentOffsetX - scrollX
	childOffsetY := parentOffsetY - scrollY

	// Inner clip: content area inside border + padding (using screen coordinates)
	screenX := node.X + parentOffsetX
	screenY := node.Y + parentOffsetY
	bw := 0
	if hasBorder(node.Style) {
		bw = 1
	}
	innerX1 := screenX + bw + node.Style.PaddingLeft
	innerY1 := screenY + bw + node.Style.PaddingTop
	innerX2 := screenX + node.W - bw - node.Style.PaddingRight
	innerY2 := screenY + node.H - bw - node.Style.PaddingBottom

	// Effective clip = intersection of outer and inner
	clipX1 := max(outerClipX1, innerX1)
	clipY1 := max(outerClipY1, innerY1)
	clipX2 := min(outerClipX2, innerX2)
	clipY2 := min(outerClipY2, innerY2)

	if clipX1 >= clipX2 || clipY1 >= clipY2 {
		return // No visible area
	}

	for _, child := range paintOrderChildren(node.Children) {
		paintNodeClipped(buf, child, clipX1, clipY1, clipX2, clipY2, childOffsetX, childOffsetY)
	}

	// Paint scrollbar in the reserved right column (unless hidden)
	if node.Style.Scrollbar != "none" {
		sbX := innerX2 - 1
		// DEBUG: detect scrollbar position anomaly — scrollbar should be inside
		// the effective clip area. If it's outside, log for diagnosis.
		if sbX < clipX1 || sbX >= clipX2 {
			// Scrollbar X is outside effective clip — this is the leak!
			fmt.Fprintf(os.Stderr, "DEBUG_SB_LEAK: scrollbarX=%d clipX1=%d clipX2=%d node.X=%d node.W=%d screenX=%d innerX2=%d outerClipX1=%d outerClipX2=%d parentOffsetX=%d node.ID=%q\n",
				sbX, clipX1, clipX2, node.X, node.W, screenX, innerX2, outerClipX1, outerClipX2, parentOffsetX, node.ID)
			// Print parent chain
			for i, p := 0, node.Parent; p != nil && i < 5; i, p = i+1, p.Parent {
				fmt.Fprintf(os.Stderr, "  DEBUG_SB_LEAK: parent[%d] type=%q id=%q X=%d W=%d overflow=%q\n",
					i, p.Type, p.ID, p.X, p.W, p.Style.Overflow)
			}
		}
		paintScrollbar(buf, node, sbX, innerY1, innerY2, clipX1, clipX2, maxScrollY)
	}
}

// paintHiddenChildrenClipped paints overflow:hidden children with both the
// inner content clip AND an outer clip rect (from a parent clipped container).
// The effective clip is the intersection of the outer clip and the inner content area.
func paintHiddenChildrenClipped(buf *CellBuffer, node *Node, outerClipX1, outerClipY1, outerClipX2, outerClipY2, offsetX, offsetY int) {
	screenX := node.X + offsetX
	screenY := node.Y + offsetY
	bw := 0
	if hasBorder(node.Style) {
		bw = 1
	}
	innerX1 := screenX + bw + node.Style.PaddingLeft
	innerY1 := screenY + bw + node.Style.PaddingTop
	innerX2 := screenX + node.W - bw - node.Style.PaddingRight
	innerY2 := screenY + node.H - bw - node.Style.PaddingBottom

	// Effective clip = intersection of outer and inner
	clipX1 := max(outerClipX1, innerX1)
	clipY1 := max(outerClipY1, innerY1)
	clipX2 := min(outerClipX2, innerX2)
	clipY2 := min(outerClipY2, innerY2)

	if clipX1 >= clipX2 || clipY1 >= clipY2 {
		return // No visible area
	}

	for _, child := range paintOrderChildren(node.Children) {
		paintNodeClipped(buf, child, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY)
	}
}

// paintScrollbar draws a vertical scrollbar in the reserved right column of a scroll container.
// scrollbarX is the x-coordinate of the scrollbar column.
// clipY1/clipY2 define the vertical extent of the content area.
// clipX1/clipX2 define the horizontal clip bounds (scrollbar is only drawn if within these bounds).
// maxScroll is the maximum scroll offset (0 means no overflow, no scrollbar drawn).
func paintScrollbar(buf *CellBuffer, node *Node, scrollbarX, clipY1, clipY2, clipX1, clipX2, maxScroll int) {
	if maxScroll <= 0 {
		return // content fits, no scrollbar needed
	}
	// Defensive: scrollbar must be within horizontal clip bounds and within the node's own area
	if scrollbarX < clipX1 || scrollbarX >= clipX2 {
		return
	}
	if scrollbarX < node.X || scrollbarX >= node.X+node.W {
		return
	}
	// Extra safety: verify against nearest overflow:hidden ancestor bounds.
	// If the scroll container's layout coordinates are stale/wrong, this prevents
	// the scrollbar from leaking into other panels.
	for p := node.Parent; p != nil; p = p.Parent {
		if p.Type == "component" {
			continue
		}
		if p.Style.Overflow == "hidden" || p.Style.Overflow == "scroll" {
			if scrollbarX < p.X || scrollbarX >= p.X+p.W {
				return // scrollbar would be outside parent clip — skip
			}
			break
		}
	}

	visibleH := clipY2 - clipY1
	if visibleH <= 0 {
		return
	}

	totalH := node.ScrollHeight
	if totalH <= 0 {
		return
	}

	// Calculate thumb size and position
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
	if thumbPos < 0 {
		thumbPos = 0
	}
	if thumbPos > trackSpace {
		thumbPos = trackSpace
	}

	// Determine colors: use style overrides or defaults
	trackBG := node.Style.Background
	thumbFG := "#6c7086" // dim gray for track (default)
	thumbBright := "#cdd6f4" // bright for thumb (default)
	if node.Style.ScrollbarTrackColor != "" {
		trackBG = node.Style.ScrollbarTrackColor
	}
	if node.Style.ScrollbarThumbColor != "" {
		thumbBright = node.Style.ScrollbarThumbColor
	}

	for row := 0; row < visibleH; row++ {
		y := clipY1 + row
		// DEBUG: detect scrollbar writing to left panel area
		if scrollbarX < 33 {
			fmt.Fprintf(os.Stderr, "DEBUG_SB_WRITE: scrollbarX=%d y=%d node.X=%d node.W=%d clipX1=%d clipX2=%d\n",
				scrollbarX, y, node.X, node.W, clipX1, clipX2)
			// Don't return — scrollbarX < 33 is valid for the left panel's own scrollbar
		}
		if row >= thumbPos && row < thumbPos+thumbSize {
			// Thumb
			buf.Set(scrollbarX, y, Cell{Ch: '█', FG: thumbBright, BG: trackBG})
		} else {
			// Track
			buf.Set(scrollbarX, y, Cell{Ch: '░', FG: thumbFG, BG: trackBG, Dim: true})
		}
	}
}



// paintNodeClipped paints a node, but only writes cells within the clip rect [clipX1, clipY1) to [clipX2, clipY2).
// offsetY is a paint-time vertical offset (used for scroll containers: -scrollY).
// The node's tree positions are NOT mutated; the offset is applied during painting only.
func paintNodeClipped(buf *CellBuffer, node *Node, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY int) {
	if node == nil || node.W <= 0 || node.H <= 0 {
		return
	}
	if node.Style.Display == "none" || node.Style.Visibility == "hidden" {
		return
	}
	// Depth guard: prevent stack overflow from cyclic trees (same as paintNode)
	paintDepth++
	if paintDepth > maxPaintDepth {
		paintDepth--
		return
	}
	defer func() { paintDepth-- }()
	// Skip entirely if the node is outside the clip rect (with offset applied)
	screenX := node.X + offsetX
	screenY := node.Y + offsetY
	if screenY >= clipY2 || screenY+node.H <= clipY1 || screenX >= clipX2 || screenX+node.W <= clipX1 {
		return
	}

	// Apply hover style: temporarily swap node.Style with merged version
	if node.Hovered && node.HoverStyle != nil {
		savedStyle := node.Style
		node.Style = mergeHoverStyle(node.Style, node.HoverStyle)
		defer func() { node.Style = savedStyle }()
	}

	switch node.Type {
	case "text":
		paintTextClipped(buf, node, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY)
	case "box", "vbox", "hbox":
		paintBoxClipped(buf, node, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY)
	case "component":
		for _, child := range paintOrderChildren(node.Children) {
			paintNodeClipped(buf, child, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY)
		}
	}
}

func paintBoxClipped(buf *CellBuffer, node *Node, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY int) {
	screenX := node.X + offsetX
	screenY := node.Y + offsetY

	// DEBUG: detect cross-panel painting leak.
	// If a node with W > 40 is being painted into a narrow clip (< 35 wide),
	// it likely means a right-panel node is being drawn through left-panel clip.
	if node.W > 40 && clipX2-clipX1 < 35 && node.Style.Background != "" {
		fmt.Fprintf(os.Stderr, "DEBUG_BOX_LEAK: node.X=%d node.W=%d screenX=%d clipX1=%d clipX2=%d bg=%q type=%q id=%q offsetX=%d\n",
			node.X, node.W, screenX, clipX1, clipX2, node.Style.Background, node.Type, node.ID, offsetX)
		for i, p := 0, node.Parent; p != nil && i < 6; i, p = i+1, p.Parent {
			fmt.Fprintf(os.Stderr, "  parent[%d] type=%q id=%q X=%d W=%d overflow=%q bg=%q\n",
				i, p.Type, p.ID, p.X, p.W, p.Style.Overflow, p.Style.Background)
		}
	}

	// Fill background (clipped)
	if node.Style.Background != "" {
		for y := screenY; y < screenY+node.H; y++ {
			if y < clipY1 || y >= clipY2 {
				continue
			}
			for x := screenX; x < screenX+node.W; x++ {
				if x >= clipX1 && x < clipX2 {
					buf.SetChar(x, y, ' ', "", node.Style.Background, false)
				}
			}
		}
	}

	// Border (must match paintBox): scroll/hidden parents paint children via
	// paintNodeClipped → paintBoxClipped; omitting borders dropped LuxButton
	// outlines inside overflow:scroll main regions.
	if hasBorder(node.Style) {
		paintBorderClipped(buf, node, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY)
	}

	// Paint children — handle nested scroll/hidden containers
	if node.Style.Overflow == "scroll" {
		// Nested scroll container inside an outer scroll: apply inner scroll
		// offset and use the intersection of outer clip and inner content area.
		paintScrollChildrenClipped(buf, node, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY)
	} else if node.Style.Overflow == "hidden" {
		// overflow:hidden inside a clipped parent: use intersection of
		// outer clip and inner content area.
		paintHiddenChildrenClipped(buf, node, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY)
	} else {
		for _, child := range paintOrderChildren(node.Children) {
			paintNodeClipped(buf, child, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY)
		}
	}
}

func paintTextClipped(buf *CellBuffer, node *Node, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY int) {
	screenX := node.X + offsetX
	screenY := node.Y + offsetY
	// Fill background (clipped)
	if node.Style.Background != "" {
		for y := screenY; y < screenY+node.H; y++ {
			if y < clipY1 || y >= clipY2 {
				continue
			}
			for x := screenX; x < screenX+node.W; x++ {
				if x >= clipX1 && x < clipX2 {
					buf.SetChar(x, y, ' ', "", node.Style.Background, false)
				}
			}
		}
	}

	// If spans are present, use span-based painting
	if len(node.Spans) > 0 {
		paintTextSpansClipped(buf, node, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY)
		return
	}

	fg := node.Style.Foreground
	bold := node.Style.Bold
	italic := node.Style.Italic
	strikethrough := node.Style.Strikethrough
	inverse := node.Style.Inverse
	dim := node.Style.Dim
	underline := node.Style.Underline
	noWrap := node.Style.WhiteSpace == "nowrap"
	ellipsis := node.Style.TextOverflow == "ellipsis"
	textAlign := node.Style.TextAlign
	rightEdge := screenX + node.W
	availW := node.W

	if noWrap {
		// Nowrap clipped path: render each line independently
		lines := splitLines(node.Content)
		for lineIdx, line := range lines {
			y := screenY + lineIdx
			if y >= screenY+node.H {
				break
			}
			if y < clipY1 || y >= clipY2 {
				continue
			}
			lineW := stringWidth(line)
			x := alignedX(screenX, availW, lineW, textAlign)

			truncated := ellipsis && lineW > availW
			maxW := availW
			if truncated {
				maxW = availW - 1 // leave room for ellipsis
			}

			col := x
			colW := 0
			for _, ch := range line {
				if ch == '\t' {
					for i := 0; i < 4; i++ {
						if colW >= availW {
							break
						}
						if col >= clipX1 && col < clipX2 {
							bg := node.Style.Background
							if bg == "" {
								existing := buf.Get(col, y)
								bg = existing.BG
							}
							buf.Set(col, y, Cell{Ch: ' ', FG: fg, BG: bg, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
						}
						col++
						colW++
					}
					continue
				}
				w := runeWidth(ch)
				if truncated && colW+w > maxW {
					// Paint ellipsis
					if col >= clipX1 && col < clipX2 {
						bg := node.Style.Background
						if bg == "" {
							existing := buf.Get(col, y)
							bg = existing.BG
						}
						buf.Set(col, y, Cell{Ch: '…', FG: fg, BG: bg, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
					}
					break
				}
				if col >= clipX1 && col < clipX2 {
					bg := node.Style.Background
					if bg == "" {
						existing := buf.Get(col, y)
						bg = existing.BG
					}
					if w == 2 && col+1 >= clipX2 {
						buf.Set(col, y, Cell{Ch: ' ', FG: fg, BG: bg, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
					} else {
						buf.Set(col, y, Cell{Ch: ch, FG: fg, BG: bg, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
						if w == 2 && col+1 >= clipX1 && col+1 < clipX2 {
							buf.Set(col+1, y, Cell{Wide: true, BG: bg})
						}
					}
				} else if w == 2 && col == clipX1-1 && clipX1 < clipX2 {
					bg := node.Style.Background
					if bg == "" {
						existing := buf.Get(clipX1, y)
						bg = existing.BG
					}
					buf.Set(clipX1, y, Cell{Ch: ' ', FG: fg, BG: bg, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
				}
				col += w
				colW += w
			}
		}
	} else {
		// Wrapping mode (original logic)
		x := screenX
		y := screenY
		for _, ch := range node.Content {
			if ch == '\n' {
				y++
				x = screenX
				continue
			}
			if ch == '\t' {
				// Render tab as 4 spaces
				for i := 0; i < 4; i++ {
					if x >= rightEdge {
						break
					}
					if y >= clipY1 && y < clipY2 && x >= clipX1 && x < clipX2 {
						bg := node.Style.Background
						if bg == "" {
							existing := buf.Get(x, y)
							bg = existing.BG
						}
						buf.Set(x, y, Cell{Ch: ' ', FG: fg, BG: bg, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
					}
					x++
				}
				continue
			}
			w := runeWidth(ch)
			// Wrap to next line if character doesn't fit
			if x+w > rightEdge {
				y++
				x = screenX
			}
			if y >= screenY+node.H {
				break
			}
			if x+w-1 < rightEdge {
				if y >= clipY1 && y < clipY2 {
					if x >= clipX1 && x < clipX2 {
						bg := node.Style.Background
						if bg == "" {
							existing := buf.Get(x, y)
							bg = existing.BG
						}
						if w == 2 && x+1 >= clipX2 {
							// Wide char doesn't fully fit at right clip boundary — write space
							buf.Set(x, y, Cell{Ch: ' ', FG: fg, BG: bg, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
						} else {
							buf.Set(x, y, Cell{Ch: ch, FG: fg, BG: bg, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
							if w == 2 && x+1 >= clipX1 && x+1 < clipX2 {
								buf.Set(x+1, y, Cell{Wide: true, BG: bg})
							}
						}
					} else if w == 2 && x == clipX1-1 && clipX1 < clipX2 {
						// Wide char straddles left clip boundary — write space at clipX1
						bg := node.Style.Background
						if bg == "" {
							existing := buf.Get(clipX1, y)
							bg = existing.BG
						}
						buf.Set(clipX1, y, Cell{Ch: ' ', FG: fg, BG: bg, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
					}
				}
				x += w
			}
		}
	}
}

// paintRuneCell writes a single rune to the buffer at (x, y) with the node's
// style attributes. It resolves background from the node or inherits from the
// existing cell. Returns the rune's display width. If rightEdge is exceeded,
// the cell is not written and 0 is returned.
func paintRuneCell(buf *CellBuffer, x, y int, ch rune, node *Node, rightEdge int) int {
	if ch == '\t' {
		// Render tab as 4 spaces
		adv := 0
		for i := 0; i < 4; i++ {
			if x+i >= rightEdge {
				break
			}
			bg := node.Style.Background
			if bg == "" {
				existing := buf.Get(x+i, y)
				bg = existing.BG
			}
			buf.Set(x+i, y, Cell{
				Ch:            ' ',
				FG:            node.Style.Foreground,
				BG:            bg,
				Bold:          node.Style.Bold,
				Dim:           node.Style.Dim,
				Underline:     node.Style.Underline,
				Italic:        node.Style.Italic,
				Strikethrough: node.Style.Strikethrough,
				Inverse:       node.Style.Inverse,
			})
			adv++
		}
		return adv
	}
	w := runeWidth(ch)
	if x+w > rightEdge {
		return 0
	}
	bg := node.Style.Background
	if bg == "" {
		existing := buf.Get(x, y)
		bg = existing.BG
	}
	buf.Set(x, y, Cell{
		Ch:            ch,
		FG:            node.Style.Foreground,
		BG:            bg,
		Bold:          node.Style.Bold,
		Dim:           node.Style.Dim,
		Underline:     node.Style.Underline,
		Italic:        node.Style.Italic,
		Strikethrough: node.Style.Strikethrough,
		Inverse:       node.Style.Inverse,
	})
	if w == 2 && x+1 < rightEdge {
		buf.Set(x+1, y, Cell{Wide: true, BG: bg})
	}
	return w
}

func paintText(buf *CellBuffer, node *Node) {
	// DEBUG: detect text nodes being painted non-clipped that cross panel boundary
	if node.X+node.W > 32 && node.X < 32 {
		fmt.Fprintf(os.Stderr, "DEBUG_TEXT_OVERFLOW: paintText(NON-CLIPPED) node.X=%d node.W=%d rightEdge=%d content=%q id=%q\n",
			node.X, node.W, node.X+node.W, truncStr(node.Content, 40), node.ID)
		for i, p := 0, node.Parent; p != nil && i < 5; i, p = i+1, p.Parent {
			fmt.Fprintf(os.Stderr, "  parent[%d] type=%q id=%q X=%d W=%d overflow=%q\n",
				i, p.Type, p.ID, p.X, p.W, p.Style.Overflow)
		}
		if node.Parent != nil && node.Parent.Parent == nil {
			fmt.Fprintf(os.Stderr, "  *** DETACHED: parent vbox has NO grandparent (Parent.Parent=nil). This node is in a detached subtree!\n")
		}
	}
	// DON'T fill background if not set (preserve parent's background)
	if node.Style.Background != "" {
		for y := node.Y; y < node.Y+node.H; y++ {
			for x := node.X; x < node.X+node.W; x++ {
				buf.SetChar(x, y, ' ', "", node.Style.Background, false)
			}
		}
	}

	// If spans are present, use span-based painting
	if len(node.Spans) > 0 {
		paintTextSpans(buf, node)
		return
	}

	// Text alignment and overflow
	textAlign := node.Style.TextAlign
	noWrap := node.Style.WhiteSpace == "nowrap"
	ellipsis := node.Style.TextOverflow == "ellipsis"

	rightEdge := node.X + node.W
	availW := node.W

	if noWrap {
		// No-wrap mode: render each line without wrapping, clip or ellipsis
		lines := splitLines(node.Content)
		for lineIdx, line := range lines {
			y := node.Y + lineIdx
			if y >= node.Y+node.H {
				break
			}
			runes := []rune(line)
			lineW := stringWidth(line)

			// Calculate starting X based on alignment
			x := alignedX(node.X, availW, lineW, textAlign)

			// Determine truncation point if ellipsis
			var truncIdx int
			var truncated bool
			if ellipsis && lineW > availW {
				truncated = true
				truncIdx = truncateRunesForWidth(runes, availW-1) // leave room for "…"
			}

			col := x
			for i, ch := range runes {
				if truncated && i >= truncIdx {
					// Paint ellipsis
					col += paintRuneCell(buf, col, y, '…', node, rightEdge)
					break
				}
				adv := paintRuneCell(buf, col, y, ch, node, rightEdge)
				if adv == 0 {
					break // clipped
				}
				col += adv
			}
		}
	} else {
		// Normal wrapping mode (existing behavior + alignment + decorations)
		x := node.X
		y := node.Y

		// For alignment in wrapping mode, we need to process line by line
		if textAlign == "center" || textAlign == "right" {
			lines := splitLines(node.Content)
			for lineIdx, line := range lines {
				y = node.Y + lineIdx
				if y >= node.Y+node.H {
					break
				}
				// TODO: wrapping + alignment is complex; for now, align first line of each \n-segment
				lineW := stringWidth(line)
				x = alignedX(node.X, availW, lineW, textAlign)

				for _, ch := range line {
					w := runeWidth(ch)
					if x+w > rightEdge {
						y++
						x = node.X
					}
					if y >= node.Y+node.H {
						break
					}
					adv := paintRuneCell(buf, x, y, ch, node, rightEdge)
					if adv > 0 {
						x += adv
					}
				}
			}
		} else {
			// Default left-aligned wrapping (original behavior + decorations)
			for _, ch := range node.Content {
				if ch == '\n' {
					y++
					x = node.X
					continue
				}
				w := runeWidth(ch)
				// Wrap to next line if character doesn't fit
				if x+w > rightEdge {
					y++
					x = node.X
				}
				if y >= node.Y+node.H {
					break
				}
				adv := paintRuneCell(buf, x, y, ch, node, rightEdge)
				if adv > 0 {
					x += adv
				}
			}
		}
	}
}

// resolveSpanStyle resolves a span's effective style, inheriting from the node's Style.
func resolveSpanStyle(span *Span, nodeStyle *Style) (fg, bg string, bold, dim, underline, italic, strikethrough, inverse bool) {
	fg = span.Foreground
	if fg == "" {
		fg = nodeStyle.Foreground
	}
	bg = span.Background
	if bg == "" {
		bg = nodeStyle.Background
	}
	if span.Bold != nil {
		bold = *span.Bold
	} else {
		bold = nodeStyle.Bold
	}
	if span.Dim != nil {
		dim = *span.Dim
	} else {
		dim = nodeStyle.Dim
	}
	if span.Underline != nil {
		underline = *span.Underline
	} else {
		underline = nodeStyle.Underline
	}
	if span.Italic != nil {
		italic = *span.Italic
	} else {
		italic = nodeStyle.Italic
	}
	if span.Strikethrough != nil {
		strikethrough = *span.Strikethrough
	} else {
		strikethrough = nodeStyle.Strikethrough
	}
	if span.Inverse != nil {
		inverse = *span.Inverse
	} else {
		inverse = nodeStyle.Inverse
	}
	return
}

// paintRuneCellStyled writes a single rune to the buffer with explicit style.
// Returns the rune's display width. If rightEdge is exceeded, 0 is returned.
func paintRuneCellStyled(buf *CellBuffer, x, y int, ch rune, fg, bg string, bold, dim, underline, italic, strikethrough, inverse bool, rightEdge int) int {
	if ch == '\t' {
		// Render tab as 4 spaces
		adv := 0
		for i := 0; i < 4; i++ {
			if x+i >= rightEdge {
				break
			}
			cellBG := bg
			if cellBG == "" {
				existing := buf.Get(x+i, y)
				cellBG = existing.BG
			}
			buf.Set(x+i, y, Cell{
				Ch:            ' ',
				FG:            fg,
				BG:            cellBG,
				Bold:          bold,
				Dim:           dim,
				Underline:     underline,
				Italic:        italic,
				Strikethrough: strikethrough,
				Inverse:       inverse,
			})
			adv++
		}
		return adv
	}
	w := runeWidth(ch)
	if x+w > rightEdge {
		return 0
	}
	cellBG := bg
	if cellBG == "" {
		existing := buf.Get(x, y)
		cellBG = existing.BG
	}
	buf.Set(x, y, Cell{
		Ch:            ch,
		FG:            fg,
		BG:            cellBG,
		Bold:          bold,
		Dim:           dim,
		Underline:     underline,
		Italic:        italic,
		Strikethrough: strikethrough,
		Inverse:       inverse,
	})
	if w == 2 && x+1 < rightEdge {
		buf.Set(x+1, y, Cell{Wide: true, BG: cellBG})
	}
	return w
}

// paintTextSpans renders spans in a text node (non-clipped path).
func paintTextSpans(buf *CellBuffer, node *Node) {
	// DEBUG: detect span text nodes being painted non-clipped that cross panel boundary
	if node.X+node.W > 32 && node.X < 32 {
		var spanPreview string
		for i, sp := range node.Spans {
			if i > 2 {
				break
			}
			spanPreview += sp.Text
		}
		fmt.Fprintf(os.Stderr, "DEBUG_TEXT_OVERFLOW: paintTextSpans(NON-CLIPPED) node.X=%d node.W=%d rightEdge=%d spans=%q id=%q\n",
			node.X, node.W, node.X+node.W, truncStr(spanPreview, 40), node.ID)
		for i, p := 0, node.Parent; p != nil && i < 5; i, p = i+1, p.Parent {
			fmt.Fprintf(os.Stderr, "  parent[%d] type=%q id=%q X=%d W=%d overflow=%q\n",
				i, p.Type, p.ID, p.X, p.W, p.Style.Overflow)
		}
		if node.Parent != nil && node.Parent.Parent == nil {
			fmt.Fprintf(os.Stderr, "  *** DETACHED: parent vbox has NO grandparent (Parent.Parent=nil). This node is in a detached subtree!\n")
		}
	}
	textAlign := node.Style.TextAlign
	noWrap := node.Style.WhiteSpace == "nowrap"
	ellipsis := node.Style.TextOverflow == "ellipsis"
	rightEdge := node.X + node.W
	availW := node.W

	if noWrap {
		// No-wrap mode: each line rendered independently, clipped or ellipsized.
		// Build a flat list of (rune, spanIndex) to correctly track which span
		// each character belongs to across multi-line content.
		type runeSpan struct {
			ch    rune
			spanI int
		}
		var allRunes []runeSpan
		for si := range node.Spans {
			for _, ch := range node.Spans[si].Text {
				allRunes = append(allRunes, runeSpan{ch, si})
			}
		}

		// Split into lines by '\n'
		var lines [][]runeSpan
		current := []runeSpan{}
		for _, rs := range allRunes {
			if rs.ch == '\n' {
				lines = append(lines, current)
				current = []runeSpan{}
			} else {
				current = append(current, rs)
			}
		}
		lines = append(lines, current)

		// Paint each line
		for lineIdx, lineRunes := range lines {
			y := node.Y + lineIdx
			if y >= node.Y+node.H {
				break
			}
			// Calculate line width for alignment
			lineW := 0
			for _, rs := range lineRunes {
				lineW += runeWidth(rs.ch)
			}
			x := alignedX(node.X, availW, lineW, textAlign)

			truncated := ellipsis && lineW > availW
			maxW := availW
			if truncated {
				maxW = availW - 1 // leave room for ellipsis
			}

			col := x
			colW := 0
			for _, rs := range lineRunes {
				span := &node.Spans[rs.spanI]
				fg, bg, bold, dim, underline, italic, strikethrough, inverse := resolveSpanStyle(span, &node.Style)
				w := runeWidth(rs.ch)
				if truncated && colW+w > maxW {
					// Paint ellipsis with last span's style
					paintRuneCellStyled(buf, col, y, '…', fg, bg, bold, dim, underline, italic, strikethrough, inverse, rightEdge)
					break // break this line, continue to next line
				}
				adv := paintRuneCellStyled(buf, col, y, rs.ch, fg, bg, bold, dim, underline, italic, strikethrough, inverse, rightEdge)
				if adv == 0 {
					break // clipped on this line, continue to next
				}
				col += adv
				colW += w
			}
		}
	} else {
		// Wrapping mode
		x := node.X
		y := node.Y

		if textAlign == "center" || textAlign == "right" {
			// For alignment, compute full text and use line-by-line approach
			fullText := nodeTextContent(node)
			lines := splitLines(fullText)
			// Build a flat list of (rune, spanIndex) for painting
			type styledRune struct {
				ch        rune
				spanIndex int
			}
			var allRunes []styledRune
			for si := range node.Spans {
				for _, ch := range node.Spans[si].Text {
					allRunes = append(allRunes, styledRune{ch, si})
				}
			}

			runeIdx := 0
			for lineIdx, line := range lines {
				y = node.Y + lineIdx
				if y >= node.Y+node.H {
					break
				}
				lineW := stringWidth(line)
				x = alignedX(node.X, availW, lineW, textAlign)

				for _, ch := range line {
					if runeIdx >= len(allRunes) {
						break
					}
					sr := allRunes[runeIdx]
					_ = sr // we use the span index from allRunes
					span := &node.Spans[sr.spanIndex]
					fg, bg, bold, dim, underline, italic, strikethrough, inverse := resolveSpanStyle(span, &node.Style)
					w := runeWidth(ch)
					if x+w > rightEdge {
						y++
						x = node.X
					}
					if y >= node.Y+node.H {
						break
					}
					adv := paintRuneCellStyled(buf, x, y, ch, fg, bg, bold, dim, underline, italic, strikethrough, inverse, rightEdge)
					if adv > 0 {
						x += adv
					}
					runeIdx++
				}
				// Skip the newline character in allRunes
				if runeIdx < len(allRunes) && allRunes[runeIdx].ch == '\n' {
					runeIdx++
				}
			}
		} else {
			// Default left-aligned wrapping
			for si := range node.Spans {
				span := &node.Spans[si]
				fg, bg, bold, dim, underline, italic, strikethrough, inverse := resolveSpanStyle(span, &node.Style)
				for _, ch := range span.Text {
					if ch == '\n' {
						y++
						x = node.X
						continue
					}
					w := runeWidth(ch)
					if x+w > rightEdge {
						y++
						x = node.X
					}
					if y >= node.Y+node.H {
						return
					}
					adv := paintRuneCellStyled(buf, x, y, ch, fg, bg, bold, dim, underline, italic, strikethrough, inverse, rightEdge)
					if adv > 0 {
						x += adv
					}
				}
			}
		}
	}
}

// paintTextSpansClipped renders spans in a text node with clipping (for scroll containers).
func paintTextSpansClipped(buf *CellBuffer, node *Node, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY int) {
	screenX := node.X + offsetX
	screenY := node.Y + offsetY
	rightEdge := screenX + node.W
	availW := node.W
	noWrap := node.Style.WhiteSpace == "nowrap"
	ellipsis := node.Style.TextOverflow == "ellipsis"
	textAlign := node.Style.TextAlign

	if noWrap {
		// No-wrap mode: build flat rune list with span association, split by newlines
		type runeSpan struct {
			ch    rune
			spanI int
		}
		var allRunes []runeSpan
		for si := range node.Spans {
			for _, ch := range node.Spans[si].Text {
				allRunes = append(allRunes, runeSpan{ch, si})
			}
		}

		// Split into lines by '\n'
		var lines [][]runeSpan
		current := []runeSpan{}
		for _, rs := range allRunes {
			if rs.ch == '\n' {
				lines = append(lines, current)
				current = []runeSpan{}
			} else {
				current = append(current, rs)
			}
		}
		lines = append(lines, current)

		// Paint each line
		for lineIdx, lineRunes := range lines {
			y := screenY + lineIdx
			if y >= screenY+node.H {
				break
			}
			if y < clipY1 || y >= clipY2 {
				continue
			}
			// Calculate line width for alignment
			lineW := 0
			for _, rs := range lineRunes {
				lineW += runeWidth(rs.ch)
			}
			x := alignedX(screenX, availW, lineW, textAlign)

			truncated := ellipsis && lineW > availW
			maxW := availW
			if truncated {
				maxW = availW - 1 // leave room for ellipsis
			}

			col := x
			colW := 0
			for _, rs := range lineRunes {
				span := &node.Spans[rs.spanI]
				fg, bg, bold, dim, underline, italic, strikethrough, inverse := resolveSpanStyle(span, &node.Style)

				w := runeWidth(rs.ch)
				if truncated && colW+w > maxW {
					// Paint ellipsis
					if col >= clipX1 && col < clipX2 {
						cellBG := bg
						if cellBG == "" {
							existing := buf.Get(col, y)
							cellBG = existing.BG
						}
						buf.Set(col, y, Cell{Ch: '…', FG: fg, BG: cellBG, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
					}
					break
				}
				if col >= clipX1 && col < clipX2 {
					cellBG := bg
					if cellBG == "" {
						existing := buf.Get(col, y)
						cellBG = existing.BG
					}
					if w == 2 && col+1 >= clipX2 {
						buf.Set(col, y, Cell{Ch: ' ', FG: fg, BG: cellBG, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
					} else {
						buf.Set(col, y, Cell{Ch: rs.ch, FG: fg, BG: cellBG, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
						if w == 2 && col+1 >= clipX1 && col+1 < clipX2 {
							buf.Set(col+1, y, Cell{Wide: true, BG: cellBG})
						}
					}
				} else if w == 2 && col == clipX1-1 && clipX1 < clipX2 {
					cellBG := bg
					if cellBG == "" {
						existing := buf.Get(clipX1, y)
						cellBG = existing.BG
					}
					buf.Set(clipX1, y, Cell{Ch: ' ', FG: fg, BG: cellBG, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
				}
				col += w
				colW += w
			}
		}
	} else {
		// Wrapping mode (original logic)
		x := screenX
		y := screenY
		for si := range node.Spans {
			span := &node.Spans[si]
			fg, bg, bold, dim, underline, italic, strikethrough, inverse := resolveSpanStyle(span, &node.Style)
			for _, ch := range span.Text {
				if ch == '\n' {
					y++
					x = screenX
					continue
				}
				if ch == '\t' {
					for i := 0; i < 4; i++ {
						if x >= rightEdge {
							break
						}
						if y >= clipY1 && y < clipY2 && x >= clipX1 && x < clipX2 {
							cellBG := bg
							if cellBG == "" {
								existing := buf.Get(x, y)
								cellBG = existing.BG
							}
							buf.Set(x, y, Cell{Ch: ' ', FG: fg, BG: cellBG, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
						}
						x++
					}
					continue
				}
				w := runeWidth(ch)
				if x+w > rightEdge {
					y++
					x = screenX
				}
				if y >= screenY+node.H {
					return
				}
				if x+w-1 < rightEdge {
					if y >= clipY1 && y < clipY2 {
						if x >= clipX1 && x < clipX2 {
							cellBG := bg
							if cellBG == "" {
								existing := buf.Get(x, y)
								cellBG = existing.BG
							}
							if w == 2 && x+1 >= clipX2 {
								buf.Set(x, y, Cell{Ch: ' ', FG: fg, BG: cellBG, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
							} else {
								buf.Set(x, y, Cell{Ch: ch, FG: fg, BG: cellBG, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
								if w == 2 && x+1 >= clipX1 && x+1 < clipX2 {
									buf.Set(x+1, y, Cell{Wide: true, BG: cellBG})
								}
							}
						} else if w == 2 && x == clipX1-1 && clipX1 < clipX2 {
							cellBG := bg
							if cellBG == "" {
								existing := buf.Get(clipX1, y)
								cellBG = existing.BG
							}
							buf.Set(clipX1, y, Cell{Ch: ' ', FG: fg, BG: cellBG, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
						}
					}
					x += w
				}
			}
		}
	}
}

// alignedX returns the starting X position for a line of text given alignment.
func alignedX(nodeX, availW, lineW int, textAlign string) int {
	x := nodeX
	switch textAlign {
	case "center":
		if offset := (availW - lineW) / 2; offset > 0 {
			x += offset
		}
	case "right":
		if offset := availW - lineW; offset > 0 {
			x += offset
		}
	}
	return x
}

func paintBorder(buf *CellBuffer, node *Node) {
	x, y, w, h := node.X, node.Y, node.W, node.H
	if w < 2 || h < 2 {
		return
	}

	fg := node.Style.BorderColor
	if fg == "" {
		fg = node.Style.Foreground
	}
	bg := node.Style.Background

	// Border characters based on style
	var tl, tr, bl, br, hz, vt rune
	switch node.Style.Border {
	case "single":
		tl, tr, bl, br, hz, vt = '┌', '┐', '└', '┘', '─', '│'
	case "double":
		tl, tr, bl, br, hz, vt = '╔', '╗', '╚', '╝', '═', '║'
	case "rounded":
		tl, tr, bl, br, hz, vt = '╭', '╮', '╰', '╯', '─', '│'
	default:
		return
	}

	// Corners
	buf.SetChar(x, y, tl, fg, bg, false)
	buf.SetChar(x+w-1, y, tr, fg, bg, false)
	buf.SetChar(x, y+h-1, bl, fg, bg, false)
	buf.SetChar(x+w-1, y+h-1, br, fg, bg, false)

	// Top and bottom edges
	for col := x + 1; col < x+w-1; col++ {
		buf.SetChar(col, y, hz, fg, bg, false)
		buf.SetChar(col, y+h-1, hz, fg, bg, false)
	}

	// Left and right edges
	for row := y + 1; row < y+h-1; row++ {
		buf.SetChar(x, row, vt, fg, bg, false)
		buf.SetChar(x+w-1, row, vt, fg, bg, false)
	}
}

// paintBorderClipped draws a box border like paintBorder but only writes cells
// inside [clipX1,clipY1)–[clipX2,clipY2). Used from paintBoxClipped.
func paintBorderClipped(buf *CellBuffer, node *Node, clipX1, clipY1, clipX2, clipY2, offsetX, offsetY int) {
	x, y, w, h := node.X+offsetX, node.Y+offsetY, node.W, node.H
	if w < 2 || h < 2 {
		return
	}

	fg := node.Style.BorderColor
	if fg == "" {
		fg = node.Style.Foreground
	}
	bg := node.Style.Background

	var tl, tr, bl, br, hz, vt rune
	switch node.Style.Border {
	case "single":
		tl, tr, bl, br, hz, vt = '┌', '┐', '└', '┘', '─', '│'
	case "double":
		tl, tr, bl, br, hz, vt = '╔', '╗', '╚', '╝', '═', '║'
	case "rounded":
		tl, tr, bl, br, hz, vt = '╭', '╮', '╰', '╯', '─', '│'
	default:
		return
	}

	set := func(px, py int, ch rune) {
		if px >= clipX1 && px < clipX2 && py >= clipY1 && py < clipY2 {
			buf.SetChar(px, py, ch, fg, bg, false)
		}
	}

	set(x, y, tl)
	set(x+w-1, y, tr)
	set(x, y+h-1, bl)
	set(x+w-1, y+h-1, br)
	for col := x + 1; col < x+w-1; col++ {
		set(col, y, hz)
		set(col, y+h-1, hz)
	}
	for row := y + 1; row < y+h-1; row++ {
		set(x, row, vt)
		set(x+w-1, row, vt)
	}
}

// splitLines splits text by newline characters.
func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	lines := make([]string, 0, 4)
	start := 0
	for i, ch := range s {
		if ch == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

// truncateRunesForWidth returns the number of runes from the slice that fit in maxW display columns.
func truncateRunesForWidth(runes []rune, maxW int) int {
	w := 0
	for i, r := range runes {
		rw := runeWidth(r)
		if w+rw > maxW {
			return i
		}
		w += rw
	}
	return len(runes)
}

// truncStr truncates a string to max characters (DEBUG helper).
func truncStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
