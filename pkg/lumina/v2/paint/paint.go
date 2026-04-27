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
		p.paintText(buf, node, ox, oy)
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
	for _, child := range node.Children {
		p.paintNode(buf, child, ox, oy)
	}
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
