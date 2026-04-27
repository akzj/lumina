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
		// Clear this node's region first, then repaint
		buf.ClearRect(node.X, node.Y, node.W, node.H)
		paintNode(buf, node)
		node.PaintDirty = false
		// Don't recurse into children — paintNode already painted them
		return
	}

	// Not dirty — check children
	for _, child := range node.Children {
		paintDirtyWalk(buf, child)
	}
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

	// 3. Paint children
	for _, child := range node.Children {
		paintNode(buf, child)
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
	for _, ch := range node.Content {
		if ch == '\n' {
			y++
			x = node.X
			continue
		}
		if x < node.X+node.W && y < node.Y+node.H {
			bg := node.Style.Background
			if bg == "" {
				// Inherit background from existing cell (parent painted it)
				existing := buf.Get(x, y)
				bg = existing.BG
			}
			buf.SetChar(x, y, ch, fg, bg, bold)
			x++
		}
	}
}

func paintInput(buf *CellBuffer, node *Node) {
	// Similar to paintText but with cursor support (future)
	paintText(buf, node)
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
