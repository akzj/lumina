package render

import "github.com/mattn/go-runewidth"

// Cell represents a single terminal cell.
type Cell struct {
	Ch        rune
	FG        string // color string e.g. "#FF0000" or "" for default
	BG        string // color string e.g. "#1E1E2E" or "" for default
	Bold      bool
	Dim       bool
	Underline bool
	// Wide: true if this cell is the right half of a CJK character
	Wide bool
}

// CellBuffer is a pre-allocated 2D grid of cells.
type CellBuffer struct {
	cells  []Cell // flat array: cells[y*width + x]
	width  int
	height int

	// Per-frame stats (reset via ResetStats).
	writeCount int // cells written since last ResetStats
	clearCount int // cells cleared since last ResetStats
	dirtyMinX  int // bounding box of dirty region
	dirtyMinY  int
	dirtyMaxX  int // exclusive
	dirtyMaxY  int // exclusive
}

// NewCellBuffer creates a buffer of the given size.
func NewCellBuffer(width, height int) *CellBuffer {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	cb := &CellBuffer{
		cells:  make([]Cell, width*height),
		width:  width,
		height: height,
	}
	cb.ResetStats()
	return cb
}

// ResetStats resets per-frame write/clear counters and dirty bounding box.
func (cb *CellBuffer) ResetStats() {
	cb.writeCount = 0
	cb.clearCount = 0
	cb.dirtyMinX = cb.width
	cb.dirtyMinY = cb.height
	cb.dirtyMaxX = 0
	cb.dirtyMaxY = 0
}

// CellBufferStats holds per-frame cell buffer statistics.
type CellBufferStats struct {
	WriteCount int
	ClearCount int
	DirtyX     int // bounding box X (0 if no dirty)
	DirtyY     int // bounding box Y
	DirtyW     int // bounding box width (0 if no dirty)
	DirtyH     int // bounding box height
}

// Stats returns per-frame write/clear counts and the dirty bounding box.
func (cb *CellBuffer) Stats() CellBufferStats {
	s := CellBufferStats{
		WriteCount: cb.writeCount,
		ClearCount: cb.clearCount,
	}
	if cb.dirtyMaxX > cb.dirtyMinX && cb.dirtyMaxY > cb.dirtyMinY {
		s.DirtyX = cb.dirtyMinX
		s.DirtyY = cb.dirtyMinY
		s.DirtyW = cb.dirtyMaxX - cb.dirtyMinX
		s.DirtyH = cb.dirtyMaxY - cb.dirtyMinY
	}
	return s
}

// trackDirty updates the dirty bounding box for a cell write at (x, y).
func (cb *CellBuffer) trackDirty(x, y int) {
	if x < cb.dirtyMinX {
		cb.dirtyMinX = x
	}
	if y < cb.dirtyMinY {
		cb.dirtyMinY = y
	}
	if x+1 > cb.dirtyMaxX {
		cb.dirtyMaxX = x + 1
	}
	if y+1 > cb.dirtyMaxY {
		cb.dirtyMaxY = y + 1
	}
}

// Width returns the buffer width.
func (cb *CellBuffer) Width() int { return cb.width }

// Height returns the buffer height.
func (cb *CellBuffer) Height() int { return cb.height }

// Get returns the cell at (x, y). Out-of-bounds returns zero Cell.
func (cb *CellBuffer) Get(x, y int) Cell {
	if x < 0 || x >= cb.width || y < 0 || y >= cb.height {
		return Cell{}
	}
	return cb.cells[y*cb.width+x]
}

// Set writes a cell at (x, y). Out-of-bounds is silently ignored.
func (cb *CellBuffer) Set(x, y int, c Cell) {
	if x >= 0 && x < cb.width && y >= 0 && y < cb.height {
		idx := y*cb.width + x
		old := cb.cells[idx]
		// Clean up wide character orphans:
		// If overwriting the first cell of a wide char, clear its padding cell
		if !old.Wide && old.Ch != 0 && runewidth.RuneWidth(old.Ch) == 2 && x+1 < cb.width {
			cb.cells[idx+1] = Cell{BG: c.BG}
		}
		// If overwriting the padding cell of a wide char, clear the first half
		if old.Wide && x-1 >= 0 {
			cb.cells[idx-1] = Cell{BG: cb.cells[idx-1].BG}
		}
		cb.cells[idx] = c
		cb.writeCount++
		cb.trackDirty(x, y)
	}
}

// SetChar writes a character with colors at (x, y).
func (cb *CellBuffer) SetChar(x, y int, ch rune, fg, bg string, bold bool) {
	if x >= 0 && x < cb.width && y >= 0 && y < cb.height {
		idx := y*cb.width + x
		old := cb.cells[idx]
		// Clean up wide character orphans:
		// If overwriting the first cell of a wide char, clear its padding cell
		if !old.Wide && old.Ch != 0 && runewidth.RuneWidth(old.Ch) == 2 && x+1 < cb.width {
			cb.cells[idx+1] = Cell{BG: bg}
		}
		// If overwriting the padding cell of a wide char, clear the first half
		if old.Wide && x-1 >= 0 {
			cb.cells[idx-1] = Cell{BG: cb.cells[idx-1].BG}
		}
		cb.cells[idx] = Cell{Ch: ch, FG: fg, BG: bg, Bold: bold}
		cb.writeCount++
		cb.trackDirty(x, y)
	}
}

// Clear resets all cells to zero.
func (cb *CellBuffer) Clear() {
	for i := range cb.cells {
		cb.cells[i] = Cell{}
	}
	cb.clearCount += len(cb.cells)
	if cb.width > 0 && cb.height > 0 {
		cb.dirtyMinX = 0
		cb.dirtyMinY = 0
		cb.dirtyMaxX = cb.width
		cb.dirtyMaxY = cb.height
	}
}

// ClearRect clears a rectangular region.
func (cb *CellBuffer) ClearRect(x, y, w, h int) {
	for row := y; row < y+h && row < cb.height; row++ {
		if row < 0 {
			continue
		}
		for col := x; col < x+w && col < cb.width; col++ {
			if col < 0 {
				continue
			}
			cb.cells[row*cb.width+col] = Cell{}
			cb.clearCount++
			cb.trackDirty(col, row)
		}
	}
}

// Resize creates a new buffer with the given size, copying existing content.
func (cb *CellBuffer) Resize(newWidth, newHeight int) {
	if newWidth < 0 {
		newWidth = 0
	}
	if newHeight < 0 {
		newHeight = 0
	}
	newCells := make([]Cell, newWidth*newHeight)
	minW := cb.width
	if newWidth < minW {
		minW = newWidth
	}
	minH := cb.height
	if newHeight < minH {
		minH = newHeight
	}
	for y := 0; y < minH; y++ {
		for x := 0; x < minW; x++ {
			newCells[y*newWidth+x] = cb.cells[y*cb.width+x]
		}
	}
	cb.cells = newCells
	cb.width = newWidth
	cb.height = newHeight
}
