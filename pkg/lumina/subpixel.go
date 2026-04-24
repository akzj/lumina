package lumina

import (
	"fmt"
)

// Color represents an RGB color for subpixel rendering.
type Color struct {
	R, G, B uint8
}

// Hex returns the color as "#rrggbb" string.
func (c Color) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// ColorFromHex parses a "#rrggbb" string into a Color.
func ColorFromHex(hex string) Color {
	if len(hex) < 7 || hex[0] != '#' {
		return Color{}
	}
	var r, g, b uint8
	fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b)
	return Color{r, g, b}
}

// NoColor is the zero/transparent color.
var NoColor = Color{0, 0, 0}

// SubPixelCanvas provides sub-cell resolution rendering using Unicode half-block characters.
// Each terminal cell is split into 2 vertical sub-pixels, giving 2x vertical resolution.
// The upper half-block ▀ (U+2580) renders fg=top, bg=bottom.
type SubPixelCanvas struct {
	CellW  int       // terminal cell width
	CellH  int       // terminal cell height
	PixW   int       // sub-pixel width (= CellW)
	PixH   int       // sub-pixel height (= CellH * 2)
	Pixels [][]Color // sub-pixel grid [PixH][PixW]
	BgColor Color    // background color (default black)
}

// NewSubPixelCanvas creates a canvas with the given terminal cell dimensions.
// The sub-pixel grid is CellW × (CellH * 2).
func NewSubPixelCanvas(cellW, cellH int) *SubPixelCanvas {
	pixH := cellH * 2
	pixels := make([][]Color, pixH)
	for y := 0; y < pixH; y++ {
		pixels[y] = make([]Color, cellW)
	}
	return &SubPixelCanvas{
		CellW:  cellW,
		CellH:  cellH,
		PixW:   cellW,
		PixH:   pixH,
		Pixels: pixels,
	}
}

// Clear resets all pixels to the background color.
func (c *SubPixelCanvas) Clear() {
	for y := 0; y < c.PixH; y++ {
		for x := 0; x < c.PixW; x++ {
			c.Pixels[y][x] = c.BgColor
		}
	}
}

// SetPixel sets a sub-pixel color at (x, y). Bounds-checked.
func (c *SubPixelCanvas) SetPixel(x, y int, color Color) {
	if x < 0 || x >= c.PixW || y < 0 || y >= c.PixH {
		return
	}
	c.Pixels[y][x] = color
}

// GetPixel returns the color at (x, y). Returns BgColor if out of bounds.
func (c *SubPixelCanvas) GetPixel(x, y int) Color {
	if x < 0 || x >= c.PixW || y < 0 || y >= c.PixH {
		return c.BgColor
	}
	return c.Pixels[y][x]
}

// RenderToFrame renders the canvas onto a Frame at the given cell offset.
// Each terminal cell maps to 2 vertical sub-pixels.
func (c *SubPixelCanvas) RenderToFrame(frame *Frame, offsetX, offsetY int) {
	for cellY := 0; cellY < c.CellH; cellY++ {
		frameY := offsetY + cellY
		if frameY < 0 || frameY >= frame.Height {
			continue
		}
		for cellX := 0; cellX < c.CellW; cellX++ {
			frameX := offsetX + cellX
			if frameX < 0 || frameX >= frame.Width {
				continue
			}

			topColor := c.Pixels[cellY*2][cellX]
			botColor := c.Pixels[cellY*2+1][cellX]

			cell := &frame.Cells[frameY][frameX]
			if topColor == botColor {
				if topColor == c.BgColor {
					cell.Char = ' '
					cell.Background = topColor.Hex()
				} else {
					cell.Char = '█' // full block
					cell.Foreground = topColor.Hex()
				}
			} else {
				cell.Char = '▀' // upper half block
				cell.Foreground = topColor.Hex()
				cell.Background = botColor.Hex()
			}
		}
	}
}

// -----------------------------------------------------------------------
// Drawing primitives
// -----------------------------------------------------------------------

// DrawLine draws a line from (x1,y1) to (x2,y2) using Bresenham's algorithm.
func (c *SubPixelCanvas) DrawLine(x1, y1, x2, y2 int, color Color) {
	dx := abs(x2 - x1)
	dy := -abs(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx + dy

	for {
		c.SetPixel(x1, y1, color)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x1 += sx
		}
		if e2 <= dx {
			err += dx
			y1 += sy
		}
	}
}

// DrawRect draws an outlined rectangle.
func (c *SubPixelCanvas) DrawRect(x, y, w, h int, color Color) {
	// Top and bottom edges
	for px := x; px < x+w; px++ {
		c.SetPixel(px, y, color)
		c.SetPixel(px, y+h-1, color)
	}
	// Left and right edges
	for py := y; py < y+h; py++ {
		c.SetPixel(x, py, color)
		c.SetPixel(x+w-1, py, color)
	}
}

// FillRect fills a solid rectangle.
func (c *SubPixelCanvas) FillRect(x, y, w, h int, color Color) {
	for py := y; py < y+h; py++ {
		for px := x; px < x+w; px++ {
			c.SetPixel(px, py, color)
		}
	}
}

// DrawCircle draws a circle outline using the midpoint circle algorithm.
func (c *SubPixelCanvas) DrawCircle(cx, cy, r int, color Color) {
	x := r
	y := 0
	d := 1 - r

	for x >= y {
		// Draw all 8 octants
		c.SetPixel(cx+x, cy+y, color)
		c.SetPixel(cx-x, cy+y, color)
		c.SetPixel(cx+x, cy-y, color)
		c.SetPixel(cx-x, cy-y, color)
		c.SetPixel(cx+y, cy+x, color)
		c.SetPixel(cx-y, cy+x, color)
		c.SetPixel(cx+y, cy-x, color)
		c.SetPixel(cx-y, cy-x, color)

		y++
		if d <= 0 {
			d += 2*y + 1
		} else {
			x--
			d += 2*(y-x) + 1
		}
	}
}

// DrawRoundedRect draws a rectangle with rounded corners.
func (c *SubPixelCanvas) DrawRoundedRect(x, y, w, h, radius int, color Color) {
	if radius <= 0 {
		c.DrawRect(x, y, w, h, color)
		return
	}

	// Clamp radius
	maxR := w / 2
	if h/2 < maxR {
		maxR = h / 2
	}
	if radius > maxR {
		radius = maxR
	}

	// Straight edges (excluding corners)
	for px := x + radius; px < x+w-radius; px++ {
		c.SetPixel(px, y, color)     // top
		c.SetPixel(px, y+h-1, color) // bottom
	}
	for py := y + radius; py < y+h-radius; py++ {
		c.SetPixel(x, py, color)     // left
		c.SetPixel(x+w-1, py, color) // right
	}

	// Corner arcs using midpoint circle algorithm
	drawCornerArc(c, x+radius, y+radius, radius, color, -1, -1)           // top-left
	drawCornerArc(c, x+w-1-radius, y+radius, radius, color, 1, -1)        // top-right
	drawCornerArc(c, x+radius, y+h-1-radius, radius, color, -1, 1)        // bottom-left
	drawCornerArc(c, x+w-1-radius, y+h-1-radius, radius, color, 1, 1)     // bottom-right
}

// drawCornerArc draws a quarter circle arc for rounded corners.
// sx, sy indicate the quadrant (-1/-1 = top-left, 1/-1 = top-right, etc.)
func drawCornerArc(c *SubPixelCanvas, cx, cy, r int, color Color, sx, sy int) {
	x := r
	y := 0
	d := 1 - r

	for x >= y {
		c.SetPixel(cx+sx*x, cy+sy*y, color)
		c.SetPixel(cx+sx*y, cy+sy*x, color)

		y++
		if d <= 0 {
			d += 2*y + 1
		} else {
			x--
			d += 2*(y-x) + 1
		}
	}
}

// abs returns the absolute value of an int.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}


