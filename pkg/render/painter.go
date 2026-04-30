package render

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
		buf.ClearRect(root.X, root.Y, root.W, root.H)
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

	if node.PaintDirty {
		// If any ancestor has children with fixed/absolute positioning,
		// escalate to that ancestor for full repaint to maintain correct z-order.
		if overlayParent := findOverlayAncestor(node); overlayParent != nil && !overlayParent.PaintDirty {
			overlayParent.PaintDirty = true
			bg := findAncestorBackground(overlayParent)
			clearRectWithBG(buf, overlayParent.X, overlayParent.Y, overlayParent.W, overlayParent.H, bg)
			paintNode(buf, overlayParent)
			overlayParent.PaintDirty = false
			clearPaintDirtyBelow(overlayParent)
			return
		}

		// If this node is inside a scroll container (at any ancestor level),
		// escalate to that scroll ancestor so paintScrollChildren handles
		// coordinate transformation and clipping correctly.
		// This is critical for nested scroll: inner scroll containers inside
		// outer scroll containers need the outer to repaint with correct offsets.
		scrollAncestor := findScrollableAncestor(node.Parent)
		if scrollAncestor != nil && !scrollAncestor.PaintDirty {
			scrollAncestor.PaintDirty = true
			bg := findAncestorBackground(scrollAncestor)
			clearRectWithBG(buf, scrollAncestor.X, scrollAncestor.Y, scrollAncestor.W, scrollAncestor.H, bg)
			paintNode(buf, scrollAncestor)
			scrollAncestor.PaintDirty = false
			clearPaintDirtyBelow(scrollAncestor)
			// Repaint overlapping windows above this scroll container
			repaintOverlappingSiblings(buf, scrollAncestor)
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

	switch node.Type {
	case "text":
		paintText(buf, node)
	case "box", "vbox", "hbox":
		paintBox(buf, node)
	case "input", "textarea":
		paintInput(buf, node)
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
	if node.Style.Border != "" && node.Style.Border != "none" {
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
	paintHiddenChildrenClipped(buf, node, 0, 0, 1<<30, 1<<30)
}

// paintScrollChildren paints children with a scroll offset, clipping to the content area.
// Delegates to paintScrollChildrenClipped with unbounded outer clip.
func paintScrollChildren(buf *CellBuffer, node *Node) {
	paintScrollChildrenClipped(buf, node, 0, 0, 1<<30, 1<<30)
}

// paintScrollChildrenClipped paints scroll container children with both the
// inner scroll offset AND an outer clip rect (from a parent scroll container).
// The effective clip is the intersection of the outer clip and the inner content area.
func paintScrollChildrenClipped(buf *CellBuffer, node *Node, outerClipX1, outerClipY1, outerClipX2, outerClipY2 int) {
	// Clamp scrollY to valid range
	maxScroll := computeMaxScrollY(node)
	if node.ScrollY > maxScroll {
		node.ScrollY = maxScroll
	}
	if node.ScrollY < 0 {
		node.ScrollY = 0
	}

	scrollY := node.ScrollY

	// Shift children by scroll offset
	shiftNodeTreeY(node, -scrollY)
	defer shiftNodeTreeY(node, scrollY)

	// Inner clip: content area inside border + padding
	bw := 0
	if node.Style.Border != "" && node.Style.Border != "none" {
		bw = 1
	}
	innerX1 := node.X + bw + node.Style.PaddingLeft
	innerY1 := node.Y + bw + node.Style.PaddingTop
	innerX2 := node.X + node.W - bw - node.Style.PaddingRight
	innerY2 := node.Y + node.H - bw - node.Style.PaddingBottom

	// Effective clip = intersection of outer and inner
	clipX1 := max(outerClipX1, innerX1)
	clipY1 := max(outerClipY1, innerY1)
	clipX2 := min(outerClipX2, innerX2)
	clipY2 := min(outerClipY2, innerY2)

	if clipX1 >= clipX2 || clipY1 >= clipY2 {
		return // No visible area
	}

	for _, child := range paintOrderChildren(node.Children) {
		paintNodeClipped(buf, child, clipX1, clipY1, clipX2, clipY2)
	}

	// Paint scrollbar in the reserved right column (use inner clip for position)
	paintScrollbar(buf, node, innerX2, innerY1, innerY2, maxScroll)
}

// paintHiddenChildrenClipped paints overflow:hidden children with both the
// inner content clip AND an outer clip rect (from a parent clipped container).
// The effective clip is the intersection of the outer clip and the inner content area.
func paintHiddenChildrenClipped(buf *CellBuffer, node *Node, outerClipX1, outerClipY1, outerClipX2, outerClipY2 int) {
	bw := 0
	if node.Style.Border != "" && node.Style.Border != "none" {
		bw = 1
	}
	innerX1 := node.X + bw + node.Style.PaddingLeft
	innerY1 := node.Y + bw + node.Style.PaddingTop
	innerX2 := node.X + node.W - bw - node.Style.PaddingRight
	innerY2 := node.Y + node.H - bw - node.Style.PaddingBottom

	// Effective clip = intersection of outer and inner
	clipX1 := max(outerClipX1, innerX1)
	clipY1 := max(outerClipY1, innerY1)
	clipX2 := min(outerClipX2, innerX2)
	clipY2 := min(outerClipY2, innerY2)

	if clipX1 >= clipX2 || clipY1 >= clipY2 {
		return // No visible area
	}

	for _, child := range paintOrderChildren(node.Children) {
		paintNodeClipped(buf, child, clipX1, clipY1, clipX2, clipY2)
	}
}

// shiftNodeTreeY shifts all children (recursively) of a node by dy.
// Does NOT shift the node itself (only its children).
//
// paintScrollbar draws a vertical scrollbar in the reserved right column of a scroll container.
// scrollbarX is the x-coordinate of the scrollbar column.
// clipY1/clipY2 define the vertical extent of the content area.
// maxScroll is the maximum scroll offset (0 means no overflow, no scrollbar drawn).
func paintScrollbar(buf *CellBuffer, node *Node, scrollbarX, clipY1, clipY2, maxScroll int) {
	if maxScroll <= 0 {
		return // content fits, no scrollbar needed
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

	// Determine colors: use dim track and bright thumb
	trackBG := node.Style.Background
	thumbFG := "#6c7086" // dim gray for track
	thumbBright := "#cdd6f4" // bright for thumb

	for row := 0; row < visibleH; row++ {
		y := clipY1 + row
		if row >= thumbPos && row < thumbPos+thumbSize {
			// Thumb
			buf.Set(scrollbarX, y, Cell{Ch: '█', FG: thumbBright, BG: trackBG})
		} else {
			// Track
			buf.Set(scrollbarX, y, Cell{Ch: '░', FG: thumbFG, BG: trackBG, Dim: true})
		}
	}
}

// SAFETY: This temporarily mutates node Y positions for scroll painting.
// This is only safe because the engine is single-threaded. If concurrent
// access is ever added, this must be replaced with a paint-time offset
// parameter instead of mutating the tree.
func shiftNodeTreeY(node *Node, dy int) {
	for _, child := range node.Children {
		shiftNodeY(child, dy)
	}
}

// shiftNodeY shifts a node and all its descendants by dy.
func shiftNodeY(node *Node, dy int) {
	node.Y += dy
	for _, child := range node.Children {
		shiftNodeY(child, dy)
	}
}

// paintNodeClipped paints a node, but only writes cells within the clip rect [clipX1, clipY1) to [clipX2, clipY2).
func paintNodeClipped(buf *CellBuffer, node *Node, clipX1, clipY1, clipX2, clipY2 int) {
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
	// Skip entirely if the node is outside the clip rect
	if node.Y >= clipY2 || node.Y+node.H <= clipY1 || node.X >= clipX2 || node.X+node.W <= clipX1 {
		return
	}

	switch node.Type {
	case "text":
		paintTextClipped(buf, node, clipX1, clipY1, clipX2, clipY2)
	case "box", "vbox", "hbox":
		paintBoxClipped(buf, node, clipX1, clipY1, clipX2, clipY2)
	case "input", "textarea":
		paintInputClipped(buf, node, clipX1, clipY1, clipX2, clipY2)
	case "component":
		for _, child := range paintOrderChildren(node.Children) {
			paintNodeClipped(buf, child, clipX1, clipY1, clipX2, clipY2)
		}
	}
}

func paintBoxClipped(buf *CellBuffer, node *Node, clipX1, clipY1, clipX2, clipY2 int) {
	// Fill background (clipped)
	if node.Style.Background != "" {
		for y := node.Y; y < node.Y+node.H; y++ {
			if y < clipY1 || y >= clipY2 {
				continue
			}
			for x := node.X; x < node.X+node.W; x++ {
				if x >= clipX1 && x < clipX2 {
					buf.SetChar(x, y, ' ', "", node.Style.Background, false)
				}
			}
		}
	}

	// Border (must match paintBox): scroll/hidden parents paint children via
	// paintNodeClipped → paintBoxClipped; omitting borders dropped LuxButton
	// outlines inside overflow:scroll main regions.
	if node.Style.Border != "" && node.Style.Border != "none" {
		paintBorderClipped(buf, node, clipX1, clipY1, clipX2, clipY2)
	}

	// Paint children — handle nested scroll/hidden containers
	if node.Style.Overflow == "scroll" {
		// Nested scroll container inside an outer scroll: apply inner scroll
		// offset and use the intersection of outer clip and inner content area.
		paintScrollChildrenClipped(buf, node, clipX1, clipY1, clipX2, clipY2)
	} else if node.Style.Overflow == "hidden" {
		// overflow:hidden inside a clipped parent: use intersection of
		// outer clip and inner content area.
		paintHiddenChildrenClipped(buf, node, clipX1, clipY1, clipX2, clipY2)
	} else {
		for _, child := range paintOrderChildren(node.Children) {
			paintNodeClipped(buf, child, clipX1, clipY1, clipX2, clipY2)
		}
	}
}

func paintTextClipped(buf *CellBuffer, node *Node, clipX1, clipY1, clipX2, clipY2 int) {
	// Fill background (clipped)
	if node.Style.Background != "" {
		for y := node.Y; y < node.Y+node.H; y++ {
			if y < clipY1 || y >= clipY2 {
				continue
			}
			for x := node.X; x < node.X+node.W; x++ {
				if x >= clipX1 && x < clipX2 {
					buf.SetChar(x, y, ' ', "", node.Style.Background, false)
				}
			}
		}
	}

	fg := node.Style.Foreground
	bold := node.Style.Bold
	italic := node.Style.Italic
	strikethrough := node.Style.Strikethrough
	inverse := node.Style.Inverse
	dim := node.Style.Dim
	underline := node.Style.Underline

	x := node.X
	y := node.Y
	rightEdge := node.X + node.W
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
		if x+w-1 < rightEdge {
			if y >= clipY1 && y < clipY2 && x >= clipX1 && x < clipX2 {
				bg := node.Style.Background
				if bg == "" {
					existing := buf.Get(x, y)
					bg = existing.BG
				}
				buf.Set(x, y, Cell{Ch: ch, FG: fg, BG: bg, Bold: bold, Dim: dim, Underline: underline, Italic: italic, Strikethrough: strikethrough, Inverse: inverse})
				if w == 2 && x+1 >= clipX1 && x+1 < clipX2 {
					buf.Set(x+1, y, Cell{Wide: true, BG: bg})
				}
			}
			x += w
		}
	}
}

// paintRuneCell writes a single rune to the buffer at (x, y) with the node's
// style attributes. It resolves background from the node or inherits from the
// existing cell. Returns the rune's display width. If rightEdge is exceeded,
// the cell is not written and 0 is returned.
func paintRuneCell(buf *CellBuffer, x, y int, ch rune, node *Node, rightEdge int) int {
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
	// DON'T fill background if not set (preserve parent's background)
	if node.Style.Background != "" {
		for y := node.Y; y < node.Y+node.H; y++ {
			for x := node.X; x < node.X+node.W; x++ {
				buf.SetChar(x, y, ' ', "", node.Style.Background, false)
			}
		}
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

func paintInput(buf *CellBuffer, node *Node) {
	if node.Content == "" && node.Placeholder != "" {
		// Render placeholder with dim style
		fg := node.Style.Foreground
		if fg == "" {
			fg = "#585B70" // dim gray default
		}
		x := node.X
		y := node.Y
		for _, ch := range node.Placeholder {
			w := runeWidth(ch)
			if x+w-1 < node.X+node.W && y < node.Y+node.H {
				bg := node.Style.Background
				if bg == "" {
					existing := buf.Get(x, y)
					bg = existing.BG
				}
				buf.Set(x, y, Cell{Ch: ch, FG: fg, BG: bg, Dim: true})
				if w == 2 {
					if x+1 < node.X+node.W {
						buf.Set(x+1, y, Cell{Wide: true, BG: bg})
					}
				}
				x += w
			}
		}
		// Show cursor at start if focused
		if node.Focused {
			paintInputCursor(buf, node, node.X, node.Y)
		}
		return
	}
	paintText(buf, node)

	// Show cursor if focused
	if node.Focused {
		cursorX := node.X + inputCursorScreenOffset(node)
		paintInputCursor(buf, node, cursorX, node.Y)
	}
}

// inputCursorScreenOffset calculates the screen X offset for the cursor position.
func inputCursorScreenOffset(node *Node) int {
	runes := []rune(node.Content)
	offset := 0
	for i := 0; i < node.CursorPos && i < len(runes); i++ {
		offset += runeWidth(runes[i])
	}
	return offset
}

// paintInputCursor renders a cursor at the given screen position using inverted colors.
func paintInputCursor(buf *CellBuffer, node *Node, x, y int) {
	if x >= node.X+node.W || y >= node.Y+node.H {
		return
	}
	existing := buf.Get(x, y)
	// Invert colors for cursor visibility
	fg := existing.BG
	bg := existing.FG
	if fg == "" {
		fg = "#1E1E2E" // default dark background
	}
	if bg == "" {
		bg = "#CDD6F4" // default light foreground
	}
	ch := existing.Ch
	if ch == 0 {
		ch = ' ' // cursor on empty space shows as block
	}
	buf.Set(x, y, Cell{Ch: ch, FG: fg, BG: bg})
}

// paintInputClipped renders an input/textarea node inside a clip rect,
// handling placeholder text, content text, and cursor correctly.
func paintInputClipped(buf *CellBuffer, node *Node, clipX1, clipY1, clipX2, clipY2 int) {
	if node.Content == "" && node.Placeholder != "" {
		// Render placeholder with dim style (clipped)
		fg := node.Style.Foreground
		if fg == "" {
			fg = "#585B70" // dim gray default
		}
		x := node.X
		y := node.Y
		for _, ch := range node.Placeholder {
			w := runeWidth(ch)
			if x+w-1 < node.X+node.W && y < node.Y+node.H {
				if y >= clipY1 && y < clipY2 && x >= clipX1 && x < clipX2 {
					bg := node.Style.Background
					if bg == "" {
						existing := buf.Get(x, y)
						bg = existing.BG
					}
					buf.Set(x, y, Cell{Ch: ch, FG: fg, BG: bg, Dim: true})
					if w == 2 && x+1 >= clipX1 && x+1 < clipX2 {
						buf.Set(x+1, y, Cell{Wide: true, BG: bg})
					}
				}
				x += w
			}
		}
		// Show cursor at start if focused
		if node.Focused {
			cx, cy := node.X, node.Y
			if cy >= clipY1 && cy < clipY2 && cx >= clipX1 && cx < clipX2 {
				paintInputCursor(buf, node, cx, cy)
			}
		}
		return
	}
	// Render text content (clipped)
	paintTextClipped(buf, node, clipX1, clipY1, clipX2, clipY2)

	// Show cursor if focused
	if node.Focused {
		cursorX := node.X + inputCursorScreenOffset(node)
		cursorY := node.Y
		if cursorY >= clipY1 && cursorY < clipY2 && cursorX >= clipX1 && cursorX < clipX2 {
			paintInputCursor(buf, node, cursorX, cursorY)
		}
	}
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
func paintBorderClipped(buf *CellBuffer, node *Node, clipX1, clipY1, clipX2, clipY2 int) {
	x, y, w, h := node.X, node.Y, node.W, node.H
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
