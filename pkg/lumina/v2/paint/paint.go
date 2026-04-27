package paint

import (
	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

type painter struct{}

func (p *painter) Paint(buf *buffer.Buffer, root *layout.VNode, offsetX, offsetY int) {
	if root == nil || buf == nil {
		return
	}
	p.paintNode(buf, root, offsetX, offsetY)
}

func (p *painter) paintNode(buf *buffer.Buffer, node *layout.VNode, ox, oy int) {
	switch node.Type {
	case "fragment":
		for _, child := range node.Children {
			p.paintNode(buf, child, ox, oy)
		}
		return

	case "text":
		p.paintText(buf, node, ox, oy)
		return

	case "input":
		p.paintInput(buf, node, ox, oy)
		return

	case "textarea":
		p.paintText(buf, node, ox, oy)
		return

	default:
		// Container types: box, vbox, hbox, etc.
		p.paintContainer(buf, node, ox, oy)
	}
}

// paintContainer renders a container node: background, border, then children.
func (p *painter) paintContainer(buf *buffer.Buffer, node *layout.VNode, ox, oy int) {
	bx := node.X - ox
	by := node.Y - oy
	w := node.W
	h := node.H
	style := node.Style

	// 1. Fill background
	if style.Background != "" {
		bgCell := buffer.Cell{
			Char:       ' ',
			Background: style.Background,
		}
		buf.Fill(buffer.Rect{X: bx, Y: by, W: w, H: h}, bgCell)
	}

	// 2. Draw border
	if style.Border != "" && style.Border != "none" {
		p.paintBorder(buf, bx, by, w, h, style)
	}

	// 3. Recurse children
	if style.Overflow == "scroll" {
		p.paintScrollChildren(buf, node, ox, oy)
	} else {
		for _, child := range node.Children {
			p.paintNode(buf, child, ox, oy)
		}
	}
}

// paintScrollChildren renders children of a scroll container with clipping
// and a scrollbar. Children are painted into a temporary buffer at their
// natural (layout) positions, then only the visible portion (determined by
// ScrollY) is copied into the main buffer.
func (p *painter) paintScrollChildren(buf *buffer.Buffer, node *layout.VNode, ox, oy int) {
	style := node.Style
	bx := node.X - ox
	by := node.Y - oy
	w := node.W
	h := node.H

	borderW := 0
	if hasBorder(style) {
		borderW = 1
	}

	contentX := bx + borderW + resolvedPaddingLeft(style)
	contentY := by + borderW + resolvedPaddingTop(style)
	scrollbarW := 1 // always reserve 1 column for scrollbar track
	contentW := w - 2*borderW - resolvedPaddingLeft(style) - resolvedPaddingRight(style) - scrollbarW
	contentH := h - 2*borderW - resolvedPaddingTop(style) - resolvedPaddingBottom(style)

	if contentW <= 0 || contentH <= 0 {
		return
	}

	// The layout engine already positioned children at absolute screen coords.
	// The content origin in absolute coords is:
	absContentX := node.X + borderW + resolvedPaddingLeft(style)
	absContentY := node.Y + borderW + resolvedPaddingTop(style)

	// Calculate total content height from children's layout positions.
	totalContentH := 0
	for _, child := range node.Children {
		childBottom := (child.Y - absContentY) + child.H
		if childBottom > totalContentH {
			totalContentH = childBottom
		}
	}

	// Clamp scrollY
	scrollY := node.ScrollY
	maxScroll := totalContentH - contentH
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scrollY < 0 {
		scrollY = 0
	}
	if scrollY > maxScroll {
		scrollY = maxScroll
	}

	// Paint children into a temporary buffer sized to the full content area.
	// Temp buffer coords: (0,0) = (absContentX, absContentY) in screen space.
	if totalContentH < 1 {
		totalContentH = 1
	}
	tempBuf := buffer.New(contentW, totalContentH)
	for _, child := range node.Children {
		p.paintNode(tempBuf, child, absContentX, absContentY)
	}

	// Copy visible portion from temp buffer to main buffer.
	// When a temp cell has content but no background, inherit the background
	// from the main buffer (which already has the container's background fill).
	for dy := 0; dy < contentH; dy++ {
		srcY := scrollY + dy
		if srcY < 0 || srcY >= totalContentH {
			continue
		}
		dstY := contentY + dy
		for dx := 0; dx < contentW; dx++ {
			cell := tempBuf.Get(dx, srcY)
			if !cell.Zero() {
				if cell.Background == "" {
					existing := buf.Get(contentX+dx, dstY)
					cell.Background = existing.Background
				}
				buf.Set(contentX+dx, dstY, cell)
			}
		}
	}

	// Draw scrollbar
	p.paintScrollbar(buf, contentX+contentW, contentY, contentH, scrollY, totalContentH, style.Background)
}

// paintScrollbar renders a vertical scrollbar track and thumb.
func (p *painter) paintScrollbar(buf *buffer.Buffer, x, y, viewH, scrollY, totalH int, containerBg string) {
	if viewH <= 0 {
		return
	}

	trackBg := containerBg
	if trackBg == "" {
		trackBg = "#313244"
	}

	if totalH <= viewH {
		// Content fits — draw empty track
		for dy := 0; dy < viewH; dy++ {
			buf.Set(x, y+dy, buffer.Cell{
				Char:       ' ',
				Background: trackBg,
			})
		}
		return
	}

	// Calculate thumb position and size
	thumbH := viewH * viewH / totalH
	if thumbH < 1 {
		thumbH = 1
	}
	scrollRange := totalH - viewH
	trackRange := viewH - thumbH
	thumbY := 0
	if scrollRange > 0 && trackRange > 0 {
		thumbY = scrollY * trackRange / scrollRange
	}
	if thumbY < 0 {
		thumbY = 0
	}
	if thumbY+thumbH > viewH {
		thumbY = viewH - thumbH
	}

	for dy := 0; dy < viewH; dy++ {
		var cell buffer.Cell
		if dy >= thumbY && dy < thumbY+thumbH {
			cell = buffer.Cell{
				Char:       '█',
				Foreground: "#585B70",
				Background: trackBg,
			}
		} else {
			cell = buffer.Cell{
				Char:       ' ',
				Background: trackBg,
			}
		}
		buf.Set(x, y+dy, cell)
	}
}

// resolvedPaddingBottom returns the effective bottom padding.
func resolvedPaddingBottom(s layout.Style) int {
	if s.PaddingBottom > 0 {
		return s.PaddingBottom
	}
	return s.Padding
}

// paintBorder draws a border around the given rectangle.
func (p *painter) paintBorder(buf *buffer.Buffer, bx, by, w, h int, style layout.Style) {
	if w < 2 || h < 2 {
		return
	}

	var tl, tr, bl, br, hLine, vLine rune
	switch style.Border {
	case "single":
		tl, tr, bl, br, hLine, vLine = '┌', '┐', '└', '┘', '─', '│'
	case "double":
		tl, tr, bl, br, hLine, vLine = '╔', '╗', '╚', '╝', '═', '║'
	case "rounded":
		tl, tr, bl, br, hLine, vLine = '╭', '╮', '╰', '╯', '─', '│'
	default:
		return
	}

	fg := style.Foreground

	bc := func(ch rune) buffer.Cell {
		return buffer.Cell{Char: ch, Foreground: fg}
	}

	// Top border
	buf.Set(bx, by, bc(tl))
	for i := 1; i < w-1; i++ {
		buf.Set(bx+i, by, bc(hLine))
	}
	buf.Set(bx+w-1, by, bc(tr))

	// Bottom border
	buf.Set(bx, by+h-1, bc(bl))
	for i := 1; i < w-1; i++ {
		buf.Set(bx+i, by+h-1, bc(hLine))
	}
	buf.Set(bx+w-1, by+h-1, bc(br))

	// Vertical borders
	for i := 1; i < h-1; i++ {
		buf.Set(bx, by+i, bc(vLine))
		buf.Set(bx+w-1, by+i, bc(vLine))
	}
}

// paintText renders text content into the buffer.
func (p *painter) paintText(buf *buffer.Buffer, node *layout.VNode, ox, oy int) {
	if node.Content == "" {
		return
	}

	style := node.Style
	borderW := 0
	if style.Border == "single" || style.Border == "double" || style.Border == "rounded" {
		borderW = 1
	}

	// If this is a container-like text (has background/border), draw those first.
	bx := node.X - ox
	by := node.Y - oy

	if style.Background != "" {
		bgCell := buffer.Cell{
			Char:       ' ',
			Background: style.Background,
		}
		buf.Fill(buffer.Rect{X: bx, Y: by, W: node.W, H: node.H}, bgCell)
	}

	if style.Border != "" && style.Border != "none" {
		p.paintBorder(buf, bx, by, node.W, node.H, style)
	}

	startX := bx + borderW + resolvedPaddingLeft(style)
	startY := by + borderW + resolvedPaddingTop(style)
	availW := node.W - 2*borderW - resolvedPaddingLeft(style) - resolvedPaddingRight(style)
	if availW <= 0 {
		availW = node.W
	}

	cell := buffer.Cell{
		Foreground: style.Foreground,
		Background: style.Background,
		Bold:       style.Bold,
		Dim:        style.Dim,
		Underline:  style.Underline,
	}

	col := 0
	row := 0
	for _, ch := range node.Content {
		if ch == '\n' {
			col = 0
			row++
			continue
		}

		w := runeWidth(ch)
		if w == 0 {
			continue
		}

		// Wrap if this character would exceed available width
		if availW > 0 && col+w > availW {
			col = 0
			row++
		}

		px := startX + col
		py := startY + row
		cell.Char = ch
		cell.Wide = (w == 2)

		// Inherit parent's background if text has no explicit background
		if style.Background == "" {
			existing := buf.Get(px, py)
			cell.Background = existing.Background
		}

		buf.Set(px, py, cell)

		// For wide chars, place a zero-char padding cell in the next column
		if w == 2 {
			padCell := buffer.Cell{
				Char:       0,
				Foreground: cell.Foreground,
				Background: cell.Background,
			}
			buf.Set(px+1, py, padCell)
		}
		col += w
	}
}


// paintInput renders an input node: background, border, value/placeholder, cursor.
func (p *painter) paintInput(buf *buffer.Buffer, node *layout.VNode, ox, oy int) {
	style := node.Style
	bx := node.X - ox
	by := node.Y - oy
	w := node.W
	h := node.H

	// 1. Fill background
	if style.Background != "" {
		bgCell := buffer.Cell{
			Char:       ' ',
			Background: style.Background,
		}
		buf.Fill(buffer.Rect{X: bx, Y: by, W: w, H: h}, bgCell)
	}

	// 2. Draw border if any
	if style.Border != "" && style.Border != "none" {
		p.paintBorder(buf, bx, by, w, h, style)
	}

	borderW := 0
	if hasBorder(style) {
		borderW = 1
	}

	startX := bx + borderW + resolvedPaddingLeft(style)
	startY := by + borderW + resolvedPaddingTop(style)
	availW := w - 2*borderW - resolvedPaddingLeft(style) - resolvedPaddingRight(style)
	if availW <= 0 {
		return
	}

	// 3. Determine display text
	value := node.Content
	placeholder := ""
	if pv, ok := node.Props["placeholder"]; ok {
		if s, ok := pv.(string); ok {
			placeholder = s
		}
	}

	displayText := value
	isPlaceholder := false
	if displayText == "" && placeholder != "" {
		displayText = placeholder
		isPlaceholder = true
	}

	// 4. Get cursor position (only when focused)
	cursorPos := -1 // -1 = no cursor
	if f, ok := node.Props["focused"]; ok {
		if fb, ok := f.(bool); ok && fb {
			cursorPos = len([]rune(value)) // default: end of text
			if cp, ok := node.Props["cursorPos"]; ok {
				switch v := cp.(type) {
				case int:
					cursorPos = v
				case int64:
					cursorPos = int(v)
				case float64:
					cursorPos = int(v)
				}
			}
		}
	}

	// 5. Determine scroll offset for long text
	scrollOffset := 0
	runes := []rune(displayText)
	if cursorPos >= 0 && !isPlaceholder {
		// Ensure cursor is visible within availW
		if cursorPos > availW-1 {
			scrollOffset = cursorPos - availW + 1
		}
	}

	// 6. Render text
	fg := style.Foreground
	if isPlaceholder {
		fg = "#6C7086" // muted color for placeholder
	}

	cell := buffer.Cell{
		Foreground: fg,
		Background: style.Background,
		Bold:       style.Bold,
		Dim:        isPlaceholder,
	}

	col := 0
	for i := scrollOffset; i < len(runes) && col < availW; i++ {
		ch := runes[i]
		rw := runeWidth(ch)
		if rw == 0 {
			continue
		}
		if col+rw > availW {
			break
		}

		px := startX + col
		py := startY

		// Cursor highlight
		if i == cursorPos && !isPlaceholder {
			cell.Background = "#585B70"
		} else if style.Background != "" {
			cell.Background = style.Background
		} else {
			existing := buf.Get(px, py)
			cell.Background = existing.Background
		}

		cell.Char = ch
		cell.Wide = (rw == 2)
		buf.Set(px, py, cell)

		if rw == 2 {
			padCell := buffer.Cell{
				Char:       0,
				Foreground: cell.Foreground,
				Background: cell.Background,
			}
			buf.Set(px+1, py, padCell)
		}
		col += rw
	}

	// 7. Draw cursor at end of text (block cursor)
	if cursorPos >= 0 && !isPlaceholder {
		cursorCol := 0
		for i := scrollOffset; i < cursorPos && i < len(runes); i++ {
			cursorCol += runeWidth(runes[i])
		}
		if cursorCol >= 0 && cursorCol < availW && cursorPos >= len(runes) {
			px := startX + cursorCol
			py := startY
			cursorCell := buffer.Cell{
				Char:       ' ',
				Background: "#585B70",
				Foreground: fg,
			}
			buf.Set(px, py, cursorCell)
		}
	}
}

// hasBorder returns true if the style has a visible border.
func hasBorder(s layout.Style) bool {
	return s.Border == "single" || s.Border == "double" || s.Border == "rounded"
}

// resolvedPaddingLeft returns the effective left padding.
func resolvedPaddingLeft(s layout.Style) int {
	if s.PaddingLeft > 0 {
		return s.PaddingLeft
	}
	return s.Padding
}

// resolvedPaddingRight returns the effective right padding.
func resolvedPaddingRight(s layout.Style) int {
	if s.PaddingRight > 0 {
		return s.PaddingRight
	}
	return s.Padding
}

// resolvedPaddingTop returns the effective top padding.
func resolvedPaddingTop(s layout.Style) int {
	if s.PaddingTop > 0 {
		return s.PaddingTop
	}
	return s.Padding
}

// runeWidth returns the display width of a rune in terminal columns.
func runeWidth(r rune) int {
	if r == 0 {
		return 0
	}
	// Control characters
	if r < 0x20 || (r >= 0x7F && r < 0xA0) {
		return 0
	}
	// Combining characters (zero width)
	if r >= 0x0300 && r <= 0x036F {
		return 0
	}
	// CJK ranges (double width)
	if r >= 0x1100 && r <= 0x115F {
		return 2
	}
	if r >= 0x2E80 && r <= 0x303E {
		return 2
	}
	if r >= 0x3041 && r <= 0x33BF {
		return 2
	}
	if r >= 0x3400 && r <= 0x4DBF {
		return 2
	}
	if r >= 0x4E00 && r <= 0xA4CF {
		return 2
	}
	if r >= 0xAC00 && r <= 0xD7AF {
		return 2
	}
	if r >= 0xF900 && r <= 0xFAFF {
		return 2
	}
	if r >= 0xFE30 && r <= 0xFE6F {
		return 2
	}
	if r >= 0xFF01 && r <= 0xFF60 {
		return 2
	}
	if r >= 0xFFE0 && r <= 0xFFE6 {
		return 2
	}
	if r >= 0x20000 && r <= 0x2FA1F {
		return 2
	}
	return 1
}
