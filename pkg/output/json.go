package output

import (
	"encoding/json"
	"io"

	"github.com/akzj/lumina/pkg/buffer"
)

// RenderResult is the JSON representation of a rendered screen.
type RenderResult struct {
	Width      int          `json:"width"`
	Height     int          `json:"height"`
	Cells      [][]CellJSON `json:"cells"`
	DirtyRects []RectJSON   `json:"dirty_rects,omitempty"`
}

// CellJSON is the JSON representation of a single cell.
type CellJSON struct {
	Char          string `json:"char"`
	Fg            string `json:"fg,omitempty"`
	Bg            string `json:"bg,omitempty"`
	Bold          bool   `json:"bold,omitempty"`
	Dim           bool   `json:"dim,omitempty"`
	Underline     bool   `json:"underline,omitempty"`
	Italic        bool   `json:"italic,omitempty"`
	Strikethrough bool   `json:"strikethrough,omitempty"`
	Inverse       bool   `json:"inverse,omitempty"`
}

// RectJSON is the JSON representation of a rectangle.
type RectJSON struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// jsonAdapter renders a buffer as JSON.
type jsonAdapter struct {
	w io.Writer
}

// NewJSONAdapter creates a JSON adapter that writes to w.
func NewJSONAdapter(w io.Writer) Adapter {
	return &jsonAdapter{w: w}
}

// WriteFull converts the entire buffer to JSON and writes it.
func (j *jsonAdapter) WriteFull(screen *buffer.Buffer) error {
	result := bufferToRenderResult(screen)
	return json.NewEncoder(j.w).Encode(result)
}

// WriteDirty converts only dirty cells to JSON with dirty rect info.
func (j *jsonAdapter) WriteDirty(screen *buffer.Buffer, dirtyRects []buffer.Rect) error {
	result := bufferToDirtyRenderResult(screen, dirtyRects)
	result.DirtyRects = make([]RectJSON, len(dirtyRects))
	for i, r := range dirtyRects {
		result.DirtyRects[i] = RectJSON{X: r.X, Y: r.Y, W: r.W, H: r.H}
	}
	return json.NewEncoder(j.w).Encode(result)
}

// Flush is a no-op for JSON adapter.
func (j *jsonAdapter) Flush() error { return nil }

// Close is a no-op for JSON adapter.
func (j *jsonAdapter) Close() error { return nil }

// bufferToRenderResult converts a buffer to a RenderResult.
func bufferToRenderResult(screen *buffer.Buffer) RenderResult {
	w, h := screen.Width(), screen.Height()
	cells := make([][]CellJSON, h)
	for y := 0; y < h; y++ {
		row := make([]CellJSON, w)
		for x := 0; x < w; x++ {
			c := screen.Get(x, y)
			ch := c.Char
			if ch == 0 {
				ch = ' '
			}
			row[x] = CellJSON{
				Char:          string(ch),
				Fg:            c.Foreground,
				Bg:            c.Background,
				Bold:          c.Bold,
				Dim:           c.Dim,
				Underline:     c.Underline,
				Italic:        c.Italic,
				Strikethrough: c.Strikethrough,
				Inverse:       c.Inverse,
			}
		}
		cells[y] = row
	}
	return RenderResult{
		Width:  w,
		Height: h,
		Cells:  cells,
	}
}

// bufferToDirtyRenderResult converts only the cells within dirty rects to a RenderResult.
// Rows outside dirty rects are nil (omitted in JSON). This significantly reduces payload
// for small updates (e.g., 1-cell change in a 200×50 terminal).
func bufferToDirtyRenderResult(screen *buffer.Buffer, dirtyRects []buffer.Rect) RenderResult {
	w, h := screen.Width(), screen.Height()
	cells := make([][]CellJSON, h)

	for _, r := range dirtyRects {
		for y := r.Y; y < r.Y+r.H && y < h; y++ {
			if cells[y] == nil {
				cells[y] = make([]CellJSON, w)
			}
			for x := r.X; x < r.X+r.W && x < w; x++ {
				c := screen.Get(x, y)
				ch := c.Char
				if ch == 0 {
					ch = ' '
				}
				cells[y][x] = CellJSON{
					Char:          string(ch),
					Fg:            c.Foreground,
					Bg:            c.Background,
					Bold:          c.Bold,
					Dim:           c.Dim,
					Underline:     c.Underline,
					Italic:        c.Italic,
					Strikethrough: c.Strikethrough,
					Inverse:       c.Inverse,
				}
			}
		}
	}

	return RenderResult{
		Width:  w,
		Height: h,
		Cells:  cells,
	}
}
