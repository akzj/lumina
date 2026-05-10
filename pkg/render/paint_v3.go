package render

// paint_v3.go — V3 paint engine. Single paint path, clip-safe by construction.
// All cell writes go through CellWriter which enforces clip bounds.
// No separate clipped/non-clipped variants.

// PaintFullV3 paints the entire node tree into the buffer using the V3 engine.
// The buffer is cleared first.
func PaintFullV3(buf *CellBuffer, root *Node) {
	if buf == nil || root == nil {
		return
	}
	buf.Clear()
	w := NewCellWriter(buf)
	paintNodeV3(w, root, 0)
}

// v3MaxDepth prevents stack overflow from cyclic trees.
const v3MaxDepth = 500

// paintNodeV3 is the single recursive paint function for all node types.
// It handles background, border, text, children, overflow, scroll — everything.
func paintNodeV3(w CellWriter, node *Node, depth int) {
	if node == nil || node.W <= 0 || node.H <= 0 {
		return
	}
	if node.Style.Display == "none" || node.Style.Visibility == "hidden" {
		return
	}
	if depth > v3MaxDepth {
		return
	}

	// Early out: if node is entirely outside clip, skip
	ox, oy := w.Offset()
	screenX := node.X + ox
	screenY := node.Y + oy
	clip := w.Clip()
	cx1, cy1, cx2, cy2 := clip.Bounds()
	if screenX >= cx2 || screenX+node.W <= cx1 || screenY >= cy2 || screenY+node.H <= cy1 {
		return
	}

	// Apply hover style
	var savedStyle Style
	if node.Hovered && node.HoverStyle != nil {
		savedStyle = node.Style
		node.Style = mergeHoverStyle(node.Style, node.HoverStyle)
		defer func() { node.Style = savedStyle }()
	}

	switch node.Type {
	case "text":
		paintTextV3(w, node)
	case "box", "vbox", "hbox":
		paintBoxV3(w, node, depth)
	case "component":
		// Component placeholder: transparent container, just paint children
		for _, child := range paintOrderChildren(node.Children) {
			paintNodeV3(w, child, depth+1)
		}
	}
}

// paintBoxV3 paints a box node: background, border, then children.
func paintBoxV3(w CellWriter, node *Node, depth int) {
	ox, oy := w.Offset()
	screenX := node.X + ox
	screenY := node.Y + oy

	// 1. Fill background
	if node.Style.Background != "" {
		for y := screenY; y < screenY+node.H; y++ {
			for x := screenX; x < screenX+node.W; x++ {
				w.SetChar(x-ox, y-oy, ' ', "", node.Style.Background, false)
			}
		}
	}

	// 2. Draw border
	if hasBorder(node.Style) {
		paintBorderV3(w, node)
	}

	// 3. Paint children
	if node.Style.Overflow == "scroll" {
		paintScrollChildrenV3(w, node, depth)
	} else if node.Style.Overflow == "hidden" {
		paintHiddenChildrenV3(w, node, depth)
	} else {
		for _, child := range paintOrderChildren(node.Children) {
			paintNodeV3(w, child, depth+1)
		}
	}
}

// paintHiddenChildrenV3 paints children clipped to the node's content area.
func paintHiddenChildrenV3(w CellWriter, node *Node, depth int) {
	ox, oy := w.Offset()
	screenX := node.X + ox
	screenY := node.Y + oy

	bw := 0
	if hasBorder(node.Style) {
		bw = 1
	}
	contentX := screenX + bw + node.Style.PaddingLeft
	contentY := screenY + bw + node.Style.PaddingTop
	contentW := node.W - 2*bw - node.Style.PaddingLeft - node.Style.PaddingRight
	contentH := node.H - 2*bw - node.Style.PaddingTop - node.Style.PaddingBottom

	if contentW <= 0 || contentH <= 0 {
		return
	}

	childWriter := w.WithClip(contentX, contentY, contentW, contentH)

	for _, child := range paintOrderChildren(node.Children) {
		paintNodeV3(childWriter, child, depth+1)
	}
}

// paintScrollChildrenV3 paints scroll container children with clip + offset.
func paintScrollChildrenV3(w CellWriter, node *Node, depth int) {
	ox, oy := w.Offset()
	screenX := node.X + ox
	screenY := node.Y + oy

	// Clamp scroll values
	maxScrollY := computeMaxScrollY(node)
	if node.ScrollY > maxScrollY {
		node.ScrollY = maxScrollY
	}
	if node.ScrollY < 0 {
		node.ScrollY = 0
	}
	maxScrollX := computeMaxScrollX(node)
	if node.ScrollX > maxScrollX {
		node.ScrollX = maxScrollX
	}
	if node.ScrollX < 0 {
		node.ScrollX = 0
	}

	bw := 0
	if hasBorder(node.Style) {
		bw = 1
	}
	contentX := screenX + bw + node.Style.PaddingLeft
	contentY := screenY + bw + node.Style.PaddingTop
	contentW := node.W - 2*bw - node.Style.PaddingLeft - node.Style.PaddingRight
	contentH := node.H - 2*bw - node.Style.PaddingTop - node.Style.PaddingBottom

	if contentW <= 0 || contentH <= 0 {
		return
	}

	// Narrow clip to content area, then apply scroll offset
	childWriter := w.WithClip(contentX, contentY, contentW, contentH)
	childWriter = childWriter.WithOffset(-node.ScrollX, -node.ScrollY)

	for _, child := range paintOrderChildren(node.Children) {
		paintNodeV3(childWriter, child, depth+1)
	}

	// Paint scrollbar (in the parent's clip space, not scrolled)
	if node.Style.Scrollbar != "none" && maxScrollY > 0 {
		sbX := contentX + contentW - 1
		paintScrollbarV3(w, node, sbX, contentY, contentY+contentH, contentX, contentX+contentW, maxScrollY)
	}
}

// paintBorderV3 draws a box border using the CellWriter.
func paintBorderV3(w CellWriter, node *Node) {
	ox, oy := w.Offset()
	x, y := node.X+ox, node.Y+oy
	bw, bh := node.W, node.H
	if bw < 2 || bh < 2 {
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

	// Use logical coordinates (subtract offset since SetChar adds it back)
	lx, ly := x-ox, y-oy

	// Corners
	w.SetChar(lx, ly, tl, fg, bg, false)
	w.SetChar(lx+bw-1, ly, tr, fg, bg, false)
	w.SetChar(lx, ly+bh-1, bl, fg, bg, false)
	w.SetChar(lx+bw-1, ly+bh-1, br, fg, bg, false)

	// Top and bottom edges
	for col := lx + 1; col < lx+bw-1; col++ {
		w.SetChar(col, ly, hz, fg, bg, false)
		w.SetChar(col, ly+bh-1, hz, fg, bg, false)
	}

	// Left and right edges
	for row := ly + 1; row < ly+bh-1; row++ {
		w.SetChar(lx, row, vt, fg, bg, false)
		w.SetChar(lx+bw-1, row, vt, fg, bg, false)
	}
}

// paintScrollbarV3 draws a vertical scrollbar.
func paintScrollbarV3(w CellWriter, node *Node, scrollbarX, clipY1, clipY2, clipX1, clipX2, maxScroll int) {
	if maxScroll <= 0 {
		return
	}
	if scrollbarX < clipX1 || scrollbarX >= clipX2 {
		return
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

	// Determine colors
	trackBG := node.Style.Background
	thumbFG := "#6c7086"     // dim gray for track (default)
	thumbBright := "#cdd6f4" // bright for thumb (default)
	if node.Style.ScrollbarTrackColor != "" {
		trackBG = node.Style.ScrollbarTrackColor
	}
	if node.Style.ScrollbarThumbColor != "" {
		thumbBright = node.Style.ScrollbarThumbColor
	}

	// Write scrollbar cells using the writer (which will clip automatically)
	// We need to write in screen coordinates, but SetChar expects logical coords.
	// The writer's offset is already applied, so we need to subtract it.
	ox, oy := w.Offset()
	for row := 0; row < visibleH; row++ {
		sy := clipY1 + row
		// Convert screen coord to logical coord for the writer
		lx := scrollbarX - ox
		ly := sy - oy
		if row >= thumbPos && row < thumbPos+thumbSize {
			w.Set(lx, ly, Cell{Ch: '█', FG: thumbBright, BG: trackBG})
		} else {
			w.Set(lx, ly, Cell{Ch: '░', FG: thumbFG, BG: trackBG, Dim: true})
		}
	}
}

// paintTextV3 paints a text node using the CellWriter.
func paintTextV3(w CellWriter, node *Node) {
	ox, oy := w.Offset()
	screenX := node.X + ox
	screenY := node.Y + oy

	// Fill background
	if node.Style.Background != "" {
		for y := screenY; y < screenY+node.H; y++ {
			for x := screenX; x < screenX+node.W; x++ {
				w.SetChar(x-ox, y-oy, ' ', "", node.Style.Background, false)
			}
		}
	}

	// If spans are present, use span-based painting
	if len(node.Spans) > 0 {
		paintTextSpansV3(w, node)
		return
	}

	// Text alignment and overflow
	textAlign := node.Style.TextAlign
	noWrap := node.Style.WhiteSpace == "nowrap"
	ellipsis := node.Style.TextOverflow == "ellipsis"
	rightEdge := node.X + node.W
	availW := node.W

	if noWrap {
		lines := splitLines(node.Content)
		for lineIdx, line := range lines {
			y := node.Y + lineIdx
			if y >= node.Y+node.H {
				break
			}
			runes := []rune(line)
			lineW := stringWidth(line)

			x := alignedX(node.X, availW, lineW, textAlign)

			var truncIdx int
			var truncated bool
			if ellipsis && lineW > availW {
				truncated = true
				truncIdx = truncateRunesForWidth(runes, availW-1)
			}

			col := x
			for i, ch := range runes {
				if truncated && i >= truncIdx {
					paintRuneCellV3(w, col, y, '…', node, rightEdge)
					break
				}
				adv := paintRuneCellV3(w, col, y, ch, node, rightEdge)
				if adv == 0 {
					break
				}
				col += adv
			}
		}
	} else {
		// Normal wrapping mode
		x := node.X
		y := node.Y

		if textAlign == "center" || textAlign == "right" {
			lines := splitLines(node.Content)
			for lineIdx, line := range lines {
				y = node.Y + lineIdx
				if y >= node.Y+node.H {
					break
				}
				lineW := stringWidth(line)
				x = alignedX(node.X, availW, lineW, textAlign)

				for _, ch := range line {
					cw := runeWidth(ch)
					if x+cw > rightEdge {
						y++
						x = node.X
					}
					if y >= node.Y+node.H {
						break
					}
					adv := paintRuneCellV3(w, x, y, ch, node, rightEdge)
					if adv > 0 {
						x += adv
					}
				}
			}
		} else {
			// Default left-aligned wrapping
			for _, ch := range node.Content {
				if ch == '\n' {
					y++
					x = node.X
					continue
				}
				cw := runeWidth(ch)
				if x+cw > rightEdge {
					y++
					x = node.X
				}
				if y >= node.Y+node.H {
					break
				}
				adv := paintRuneCellV3(w, x, y, ch, node, rightEdge)
				if adv > 0 {
					x += adv
				}
			}
		}
	}
}

// paintRuneCellV3 writes a single rune to the buffer via CellWriter.
// Returns the rune's display width. If rightEdge is exceeded, returns 0.
func paintRuneCellV3(w CellWriter, x, y int, ch rune, node *Node, rightEdge int) int {
	if ch == '\t' {
		adv := 0
		for i := 0; i < 4; i++ {
			if x+i >= rightEdge {
				break
			}
			bg := node.Style.Background
			if bg == "" {
				// For V3, we can't easily read existing cell BG through the writer,
				// so use empty string (CellBuffer.Set handles orphan cleanup)
				ox, oy := w.Offset()
				sx, sy := x+i+ox, y+oy
				if w.Clip().Contains(sx, sy) {
					existing := w.buf.Get(sx, sy)
					bg = existing.BG
				}
			}
			w.Set(x+i, y, Cell{
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
	rw := runeWidth(ch)
	if x+rw > rightEdge {
		return 0
	}
	bg := node.Style.Background
	if bg == "" {
		ox, oy := w.Offset()
		sx, sy := x+ox, y+oy
		if w.Clip().Contains(sx, sy) {
			existing := w.buf.Get(sx, sy)
			bg = existing.BG
		}
	}
	w.Set(x, y, Cell{
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
	if rw == 2 {
		// Wide character padding cell
		ox, oy := w.Offset()
		sx2, sy2 := x+1+ox, y+oy
		if w.Clip().Contains(sx2, sy2) {
			w.buf.Set(sx2, sy2, Cell{Wide: true, BG: bg})
		}
	}
	return rw
}

// paintRuneCellStyledV3 writes a single rune with explicit style via CellWriter.
func paintRuneCellStyledV3(w CellWriter, x, y int, ch rune, fg, bg string, bold, dim, underline, italic, strikethrough, inverse bool, rightEdge int) int {
	if ch == '\t' {
		adv := 0
		for i := 0; i < 4; i++ {
			if x+i >= rightEdge {
				break
			}
			cellBG := bg
			if cellBG == "" {
				ox, oy := w.Offset()
				sx, sy := x+i+ox, y+oy
				if w.Clip().Contains(sx, sy) {
					existing := w.buf.Get(sx, sy)
					cellBG = existing.BG
				}
			}
			w.Set(x+i, y, Cell{
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
	rw := runeWidth(ch)
	if x+rw > rightEdge {
		return 0
	}
	cellBG := bg
	if cellBG == "" {
		ox, oy := w.Offset()
		sx, sy := x+ox, y+oy
		if w.Clip().Contains(sx, sy) {
			existing := w.buf.Get(sx, sy)
			cellBG = existing.BG
		}
	}
	w.Set(x, y, Cell{
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
	if rw == 2 {
		ox, oy := w.Offset()
		sx2, sy2 := x+1+ox, y+oy
		if w.Clip().Contains(sx2, sy2) {
			w.buf.Set(sx2, sy2, Cell{Wide: true, BG: cellBG})
		}
	}
	return rw
}

// paintTextSpansV3 renders spans in a text node via CellWriter.
func paintTextSpansV3(w CellWriter, node *Node) {
	textAlign := node.Style.TextAlign
	noWrap := node.Style.WhiteSpace == "nowrap"
	ellipsis := node.Style.TextOverflow == "ellipsis"
	rightEdge := node.X + node.W
	availW := node.W

	if noWrap {
		// Build flat list of (rune, spanIndex)
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
			lineW := 0
			for _, rs := range lineRunes {
				lineW += runeWidth(rs.ch)
			}
			x := alignedX(node.X, availW, lineW, textAlign)

			truncated := ellipsis && lineW > availW
			maxW := availW
			if truncated {
				maxW = availW - 1
			}

			col := x
			colW := 0
			for _, rs := range lineRunes {
				span := &node.Spans[rs.spanI]
				fg, bg, bold, dim, underline, italic, strikethrough, inverse := resolveSpanStyle(span, &node.Style)
				rw := runeWidth(rs.ch)
				if truncated && colW+rw > maxW {
					paintRuneCellStyledV3(w, col, y, '…', fg, bg, bold, dim, underline, italic, strikethrough, inverse, rightEdge)
					break
				}
				adv := paintRuneCellStyledV3(w, col, y, rs.ch, fg, bg, bold, dim, underline, italic, strikethrough, inverse, rightEdge)
				if adv == 0 {
					break
				}
				col += adv
				colW += rw
			}
		}
	} else {
		// Wrapping mode
		x := node.X
		y := node.Y

		if textAlign == "center" || textAlign == "right" {
			fullText := nodeTextContent(node)
			lines := splitLines(fullText)
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
					span := &node.Spans[sr.spanIndex]
					fg, bg, bold, dim, underline, italic, strikethrough, inverse := resolveSpanStyle(span, &node.Style)
					rw := runeWidth(ch)
					if x+rw > rightEdge {
						y++
						x = node.X
					}
					if y >= node.Y+node.H {
						break
					}
					adv := paintRuneCellStyledV3(w, x, y, ch, fg, bg, bold, dim, underline, italic, strikethrough, inverse, rightEdge)
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
					rw := runeWidth(ch)
					if x+rw > rightEdge {
						y++
						x = node.X
					}
					if y >= node.Y+node.H {
						return
					}
					adv := paintRuneCellStyledV3(w, x, y, ch, fg, bg, bold, dim, underline, italic, strikethrough, inverse, rightEdge)
					if adv > 0 {
						x += adv
					}
				}
			}
		}
	}
}

// PaintDirtyV3 incrementally repaints only dirty nodes.
// It walks the tree top-down, accumulating clip through overflow containers.
// When a dirty node is found, its area is cleared and repainted.
// Key advantage over V2: no escalation logic needed — the CellWriter carries
// the correct clip as we walk down the tree.
func PaintDirtyV3(buf *CellBuffer, root *Node) {
	if buf == nil || root == nil {
		return
	}
	w := NewCellWriter(buf)
	paintDirtyWalkV3(w, root, 0)
}

func paintDirtyWalkV3(w CellWriter, node *Node, depth int) {
	if node == nil || node.W <= 0 || node.H <= 0 {
		return
	}
	if node.Style.Display == "none" {
		if node.PaintDirty {
			clearPaintDirty(node)
		}
		return
	}
	if depth > v3MaxDepth {
		return
	}

	// Early out: if node is entirely outside current clip, skip
	ox, oy := w.Offset()
	screenX := node.X + ox
	screenY := node.Y + oy
	clip := w.Clip()
	cx1, cy1, cx2, cy2 := clip.Bounds()
	if screenX >= cx2 || screenX+node.W <= cx1 || screenY >= cy2 || screenY+node.H <= cy1 {
		if node.PaintDirty {
			clearPaintDirty(node)
		}
		return
	}

	if node.PaintDirty {
		// Clear the node's area (within clip) then repaint it fully.
		// If node moved, also clear old position.
		if node.PositionChanged {
			bg := findAncestorBackground(node)
			w.ClearRect(node.OldX, node.OldY, node.OldW, node.OldH, bg)
			node.PositionChanged = false
		}

		bg := findAncestorBackground(node)
		w.ClearRect(node.X, node.Y, node.W, node.H, bg)
		paintNodeV3(w, node, depth)
		node.PaintDirty = false
		clearPaintDirtyBelow(node)
		return
	}

	// Not dirty — propagate clip to children and recurse
	switch node.Type {
	case "text":
		// Text nodes have no children — nothing to do if not dirty
		return
	case "component":
		for _, child := range node.Children {
			paintDirtyWalkV3(w, child, depth+1)
		}
	case "box", "vbox", "hbox":
		if node.Style.Overflow == "scroll" {
			paintDirtyWalkScrollV3(w, node, depth)
		} else if node.Style.Overflow == "hidden" {
			paintDirtyWalkHiddenV3(w, node, depth)
		} else {
			for _, child := range node.Children {
				paintDirtyWalkV3(w, child, depth+1)
			}
		}
	}
}

func paintDirtyWalkHiddenV3(w CellWriter, node *Node, depth int) {
	ox, oy := w.Offset()
	screenX := node.X + ox
	screenY := node.Y + oy

	bw := 0
	if hasBorder(node.Style) {
		bw = 1
	}
	contentX := screenX + bw + node.Style.PaddingLeft
	contentY := screenY + bw + node.Style.PaddingTop
	contentW := node.W - 2*bw - node.Style.PaddingLeft - node.Style.PaddingRight
	contentH := node.H - 2*bw - node.Style.PaddingTop - node.Style.PaddingBottom
	if contentW <= 0 || contentH <= 0 {
		return
	}

	childWriter := w.WithClip(contentX, contentY, contentW, contentH)
	for _, child := range node.Children {
		paintDirtyWalkV3(childWriter, child, depth+1)
	}
}

func paintDirtyWalkScrollV3(w CellWriter, node *Node, depth int) {
	ox, oy := w.Offset()
	screenX := node.X + ox
	screenY := node.Y + oy

	// Clamp scroll
	maxScrollY := computeMaxScrollY(node)
	if node.ScrollY > maxScrollY {
		node.ScrollY = maxScrollY
	}
	if node.ScrollY < 0 {
		node.ScrollY = 0
	}
	maxScrollX := computeMaxScrollX(node)
	if node.ScrollX > maxScrollX {
		node.ScrollX = maxScrollX
	}
	if node.ScrollX < 0 {
		node.ScrollX = 0
	}

	bw := 0
	if hasBorder(node.Style) {
		bw = 1
	}
	contentX := screenX + bw + node.Style.PaddingLeft
	contentY := screenY + bw + node.Style.PaddingTop
	contentW := node.W - 2*bw - node.Style.PaddingLeft - node.Style.PaddingRight
	contentH := node.H - 2*bw - node.Style.PaddingTop - node.Style.PaddingBottom
	if contentW <= 0 || contentH <= 0 {
		return
	}

	childWriter := w.WithClip(contentX, contentY, contentW, contentH)
	childWriter = childWriter.WithOffset(-node.ScrollX, -node.ScrollY)

	for _, child := range node.Children {
		paintDirtyWalkV3(childWriter, child, depth+1)
	}
}
