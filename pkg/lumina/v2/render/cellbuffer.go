package render

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
}

// NewCellBuffer creates a buffer of the given size.
func NewCellBuffer(width, height int) *CellBuffer {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return &CellBuffer{
		cells:  make([]Cell, width*height),
		width:  width,
		height: height,
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
		cb.cells[y*cb.width+x] = c
	}
}

// SetChar writes a character with colors at (x, y).
func (cb *CellBuffer) SetChar(x, y int, ch rune, fg, bg string, bold bool) {
	if x >= 0 && x < cb.width && y >= 0 && y < cb.height {
		cb.cells[y*cb.width+x] = Cell{Ch: ch, FG: fg, BG: bg, Bold: bold}
	}
}

// Clear resets all cells to zero.
func (cb *CellBuffer) Clear() {
	for i := range cb.cells {
		cb.cells[i] = Cell{}
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
