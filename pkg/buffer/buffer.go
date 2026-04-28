package buffer

// Buffer is a 2D grid of Cells, backed by a flat []Cell slice (single allocation).
type Buffer struct {
	cells  []Cell // flat array, row-major: cells[y*width + x]
	width  int
	height int
}

// New creates a buffer of the given size. All cells are zero-valued.
func New(w, h int) *Buffer {
	if w <= 0 || h <= 0 {
		return &Buffer{}
	}
	return &Buffer{
		cells:  make([]Cell, w*h),
		width:  w,
		height: h,
	}
}

// Width returns the buffer width.
func (b *Buffer) Width() int { return b.width }

// Height returns the buffer height.
func (b *Buffer) Height() int { return b.height }

// Get returns the cell at (x, y). Returns zero Cell if out of bounds.
func (b *Buffer) Get(x, y int) Cell {
	if x < 0 || y < 0 || x >= b.width || y >= b.height {
		return Cell{}
	}
	return b.cells[y*b.width+x]
}

// Set sets the cell at (x, y). No-op if out of bounds.
func (b *Buffer) Set(x, y int, c Cell) {
	if x < 0 || y < 0 || x >= b.width || y >= b.height {
		return
	}
	b.cells[y*b.width+x] = c
}

// Fill fills a rectangular region with the given cell.
// Clips to buffer bounds.
func (b *Buffer) Fill(r Rect, c Cell) {
	// Clip the rect to buffer bounds.
	clipped := r.Intersect(Rect{X: 0, Y: 0, W: b.width, H: b.height})
	if clipped.W <= 0 || clipped.H <= 0 {
		return
	}
	for y := clipped.Y; y < clipped.Y+clipped.H; y++ {
		rowStart := y*b.width + clipped.X
		for x := 0; x < clipped.W; x++ {
			b.cells[rowStart+x] = c
		}
	}
}

// Resize reallocates the buffer to a new size.
// Existing content is preserved where it fits; new cells are zero-valued.
func (b *Buffer) Resize(w, h int) {
	if w <= 0 || h <= 0 {
		b.cells = nil
		b.width = 0
		b.height = 0
		return
	}
	newCells := make([]Cell, w*h)
	copyW := min(b.width, w)
	copyH := min(b.height, h)
	for y := 0; y < copyH; y++ {
		copy(newCells[y*w:y*w+copyW], b.cells[y*b.width:y*b.width+copyW])
	}
	b.cells = newCells
	b.width = w
	b.height = h
}

// Clear resets all cells to zero value.
func (b *Buffer) Clear() {
	clear(b.cells)
}

// Blit copies src buffer into dst buffer at offset (dx, dy).
// Only copies cells where src cell is non-zero (transparent skip).
// clip limits the destination area.
// Returns the dirty rect (the area actually written to dst).
func Blit(dst, src *Buffer, dx, dy int, clip Rect) Rect {
	// The source rect in dst coordinates.
	srcInDst := Rect{X: dx, Y: dy, W: src.width, H: src.height}
	// Intersect with clip and dst bounds.
	dstBounds := Rect{X: 0, Y: 0, W: dst.width, H: dst.height}
	region := srcInDst.Intersect(clip).Intersect(dstBounds)
	if region.W <= 0 || region.H <= 0 {
		return Rect{}
	}

	// Track actual dirty bounds.
	dirtyMinX, dirtyMinY := region.X+region.W, region.Y+region.H
	dirtyMaxX, dirtyMaxY := region.X, region.Y

	for y := region.Y; y < region.Y+region.H; y++ {
		srcY := y - dy
		for x := region.X; x < region.X+region.W; x++ {
			srcX := x - dx
			sc := src.cells[srcY*src.width+srcX]
			if sc.Zero() {
				continue // transparent — skip
			}
			dst.cells[y*dst.width+x] = sc
			if x < dirtyMinX {
				dirtyMinX = x
			}
			if x >= dirtyMaxX {
				dirtyMaxX = x + 1
			}
			if y < dirtyMinY {
				dirtyMinY = y
			}
			if y >= dirtyMaxY {
				dirtyMaxY = y + 1
			}
		}
	}

	if dirtyMinX >= dirtyMaxX || dirtyMinY >= dirtyMaxY {
		return Rect{} // nothing was actually written
	}
	return Rect{X: dirtyMinX, Y: dirtyMinY, W: dirtyMaxX - dirtyMinX, H: dirtyMaxY - dirtyMinY}
}

// Equal returns true if two buffers have identical content.
func Equal(a, b *Buffer) bool {
	if a.width != b.width || a.height != b.height {
		return false
	}
	for i := range a.cells {
		if a.cells[i] != b.cells[i] {
			return false
		}
	}
	return true
}
