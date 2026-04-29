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
	for _, child := range node.Children {
		clearPaintDirty(child)
	}
}

// clearPaintDirtyBelow clears PaintDirty on all descendants (not the node itself).
func clearPaintDirtyBelow(node *Node) {
	for _, child := range node.Children {
		child.PaintDirty = false
		clearPaintDirtyBelow(child)
	}
}


// PaintDirty paints only PaintDirty nodes into the buffer.
// Does NOT clear the buffer — only overwrites dirty regions.
// Clears PaintDirty flags after painting.
func PaintDirty(buf *CellBuffer, root *Node) {
	if buf == nil || root == nil {
		return
	}
	paintDirtyWalk(buf, root)
}

func paintDirtyWalk(buf *CellBuffer, node *Node) {
	if node == nil {
		return
	}

	if node.PaintDirty {
		// If this node is inside a scroll container (at any ancestor level),
		// escalate to that scroll ancestor so paintScrollChildren handles
		// coordinate transformation and clipping correctly.
		// This is critical for nested scroll: inner scroll containers inside
		// outer scroll containers need the outer to repaint with correct offsets.
		scrollAncestor := findScrollableAncestor(node.Parent)
		if scrollAncestor != nil && !scrollAncestor.PaintDirty {
			scrollAncestor.PaintDirty = true
			buf.ClearRect(scrollAncestor.X, scrollAncestor.Y, scrollAncestor.W, scrollAncestor.H)
			paintNode(buf, scrollAncestor)
			scrollAncestor.PaintDirty = false
			clearPaintDirtyBelow(scrollAncestor)
			return
		}
		// For absolute-positioned nodes that moved, escalate to parent container
		// so all overlapping siblings (other windows) get repainted too.
		if node.PositionChanged && node.Style.Position == "absolute" {
			parent := findRepaintParent(node)
			if parent != nil && !parent.PaintDirty {
				buf.ClearRect(node.OldX, node.OldY, node.OldW, node.OldH)
				node.PositionChanged = false
				node.PaintDirty = false
				parent.PaintDirty = true
				buf.ClearRect(parent.X, parent.Y, parent.W, parent.H)
				paintNode(buf, parent)
				parent.PaintDirty = false
				clearPaintDirtyBelow(parent)
				return
			}
		}
		// If node moved/resized, clear the old region to avoid ghost artifacts
		if node.PositionChanged {
			buf.ClearRect(node.OldX, node.OldY, node.OldW, node.OldH)
			node.PositionChanged = false
		}
		// Clear this node's region first, then repaint
		buf.ClearRect(node.X, node.Y, node.W, node.H)
		paintNode(buf, node)
		node.PaintDirty = false
		// Clear all descendants' PaintDirty flags — paintNode already painted them
		clearPaintDirtyBelow(node)
		return
	}

	// Not dirty — check children
	for _, child := range node.Children {
		paintDirtyWalk(buf, child)
	}
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

func paintNode(buf *CellBuffer, node *Node) {
	if node == nil || node.W <= 0 || node.H <= 0 {
		return
	}

	switch node.Type {
	case "text":
		paintText(buf, node)
	case "box", "vbox", "hbox":
		paintBox(buf, node)
	case "input", "textarea":
		paintInput(buf, node)
	case "component":
		// Component placeholder: transparent container, just paint children
		for _, child := range node.Children {
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

	// 3. Paint children (with scroll offset if applicable)
	if node.Style.Overflow == "scroll" {
		paintScrollChildren(buf, node)
	} else {
		for _, child := range node.Children {
			paintNode(buf, child)
		}
	}
}

// paintScrollChildren paints children with a scroll offset, clipping to the content area
// (inside border + padding). Temporarily shifts child Y positions by -scrollY, then restores.
func paintScrollChildren(buf *CellBuffer, node *Node) {
	// Clamp scrollY to valid range (content may have shrunk since last scroll)
	maxScroll := computeMaxScrollY(node)
	if node.ScrollY > maxScroll {
		node.ScrollY = maxScroll
	}
	if node.ScrollY < 0 {
		node.ScrollY = 0
	}

	scrollY := node.ScrollY

	// Shift all child subtree Y positions
	shiftNodeTreeY(node, -scrollY)
	defer shiftNodeTreeY(node, scrollY) // restore

	// Clip to content area (inside border + padding)
	bw := 0
	if node.Style.Border != "" && node.Style.Border != "none" {
		bw = 1
	}
	clipX1 := node.X + bw + node.Style.PaddingLeft
	clipY1 := node.Y + bw + node.Style.PaddingTop
	clipX2 := node.X + node.W - bw - node.Style.PaddingRight
	clipY2 := node.Y + node.H - bw - node.Style.PaddingBottom

	for _, child := range node.Children {
		paintNodeClipped(buf, child, clipX1, clipY1, clipX2, clipY2)
	}
}
// paintScrollChildrenClipped paints scroll container children with both the
// inner scroll offset AND an outer clip rect (from a parent scroll container).
// The effective clip is the intersection of the outer clip and the inner content area.
func paintScrollChildrenClipped(buf *CellBuffer, node *Node, outerClipX1, outerClipY1, outerClipX2, outerClipY2 int) {
	// Clamp scrollY
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

	for _, child := range node.Children {
		paintNodeClipped(buf, child, clipX1, clipY1, clipX2, clipY2)
	}
}


// shiftNodeTreeY shifts all children (recursively) of a node by dy.
// Does NOT shift the node itself (only its children).
//
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
		paintTextClipped(buf, node, clipX1, clipY1, clipX2, clipY2)
	case "component":
		for _, child := range node.Children {
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

	// Paint children — handle nested scroll containers
	if node.Style.Overflow == "scroll" {
		// Nested scroll container inside an outer scroll: apply inner scroll
		// offset and use the intersection of outer clip and inner content area.
		paintScrollChildrenClipped(buf, node, clipX1, clipY1, clipX2, clipY2)
	} else {
		for _, child := range node.Children {
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
				buf.SetChar(x, y, ch, fg, bg, bold)
				if w == 2 && x+1 >= clipX1 && x+1 < clipX2 {
					buf.Set(x+1, y, Cell{Wide: true, BG: bg})
				}
			}
			x += w
		}
	}
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

	// Write text content
	fg := node.Style.Foreground
	bold := node.Style.Bold

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
			bg := node.Style.Background
			if bg == "" {
				// Inherit background from existing cell (parent painted it)
				existing := buf.Get(x, y)
				bg = existing.BG
			}
			buf.SetChar(x, y, ch, fg, bg, bold)
			if w == 2 {
				// Set padding cell for wide character
				if x+1 < rightEdge {
					buf.Set(x+1, y, Cell{Wide: true, BG: bg})
				}
			}
			x += w
		}
	}
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

func paintBorder(buf *CellBuffer, node *Node) {
	x, y, w, h := node.X, node.Y, node.W, node.H
	if w < 2 || h < 2 {
		return
	}

	fg := node.Style.Foreground
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
