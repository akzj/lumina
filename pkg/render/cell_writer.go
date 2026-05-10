package render

// CellWriter is the ONLY interface for writing cells to the buffer in V3.
// It encapsulates clip region + scroll offset. All writes are automatically clipped.
// CellWriter is a value type — methods return new writers (immutable pattern).
type CellWriter struct {
	buf     *CellBuffer
	clip    ClipRegion
	offsetX int // cumulative scroll offset X
	offsetY int // cumulative scroll offset Y
}

// NewCellWriter creates a root CellWriter covering the full buffer.
func NewCellWriter(buf *CellBuffer) CellWriter {
	return CellWriter{
		buf:  buf,
		clip: FullClip(buf.Width(), buf.Height()),
	}
}

// Set writes a cell at logical position (x, y), applying offset and clip.
// Out-of-clip writes are silently discarded (the core safety guarantee).
func (cw CellWriter) Set(x, y int, c Cell) {
	sx, sy := x+cw.offsetX, y+cw.offsetY
	if !cw.clip.Contains(sx, sy) {
		return
	}
	cw.buf.Set(sx, sy, c)
}

// SetChar writes a character with style at logical position (x, y).
func (cw CellWriter) SetChar(x, y int, ch rune, fg, bg string, bold bool) {
	sx, sy := x+cw.offsetX, y+cw.offsetY
	if !cw.clip.Contains(sx, sy) {
		return
	}
	cw.buf.SetChar(sx, sy, ch, fg, bg, bold)
}

// SetCell writes a full Cell struct at logical position (x, y).
// This is the most flexible write method, supporting all style attributes.
func (cw CellWriter) SetCell(x, y int, ch rune, fg, bg string, bold, dim, underline, italic, strikethrough, inverse bool) {
	sx, sy := x+cw.offsetX, y+cw.offsetY
	if !cw.clip.Contains(sx, sy) {
		return
	}
	cw.buf.Set(sx, sy, Cell{
		Ch: ch, FG: fg, BG: bg,
		Bold: bold, Dim: dim, Underline: underline,
		Italic: italic, Strikethrough: strikethrough, Inverse: inverse,
	})
}

// SetWideChar writes a wide (CJK) character that occupies 2 cells.
// The main cell gets the character, the next cell gets Wide=true padding.
// Both cells are individually clipped.
func (cw CellWriter) SetWideChar(x, y int, ch rune, fg, bg string, bold, dim, underline, italic, strikethrough, inverse bool) {
	// Main cell
	cw.SetCell(x, y, ch, fg, bg, bold, dim, underline, italic, strikethrough, inverse)
	// Padding cell (right half)
	sx2, sy2 := x+1+cw.offsetX, y+cw.offsetY
	if cw.clip.Contains(sx2, sy2) {
		cw.buf.Set(sx2, sy2, Cell{Wide: true, BG: bg})
	}
}

// Clip returns the current clip region.
func (cw CellWriter) Clip() ClipRegion {
	return cw.clip
}

// Offset returns the current scroll offset.
func (cw CellWriter) Offset() (x, y int) {
	return cw.offsetX, cw.offsetY
}

// WithClip returns a new CellWriter with a narrowed clip region.
// The clip can only shrink, never grow (intersection with current clip).
// The (x, y, width, height) are in SCREEN coordinates (already offset-applied).
func (cw CellWriter) WithClip(x, y, width, height int) CellWriter {
	return CellWriter{
		buf:     cw.buf,
		clip:    cw.clip.Narrow(x, y, width, height),
		offsetX: cw.offsetX,
		offsetY: cw.offsetY,
	}
}

// WithOffset returns a new CellWriter with additional scroll offset.
func (cw CellWriter) WithOffset(dx, dy int) CellWriter {
	return CellWriter{
		buf:     cw.buf,
		clip:    cw.clip,
		offsetX: cw.offsetX + dx,
		offsetY: cw.offsetY + dy,
	}
}

// ClearRect fills a rectangle with spaces using the given background color.
// Respects clip — only cells within clip are written.
// (x, y) are logical coordinates (offset will be applied).
func (cw CellWriter) ClearRect(x, y, width, height int, bg string) {
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			cw.SetChar(col, row, ' ', "", bg, false)
		}
	}
}
