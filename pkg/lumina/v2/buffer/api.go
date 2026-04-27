// Package buffer provides a 2D grid of terminal character cells.
// It is the foundational visual layer for Lumina v2.
package buffer

// Cell represents a single terminal character cell. Pure visual data.
type Cell struct {
	Char       rune
	Foreground string // "#rrggbb" or "" (inherit)
	Background string // "#rrggbb" or "" (transparent)
	Bold       bool
	Dim        bool
	Underline  bool
}

// Zero returns true if this cell has never been written to.
func (c Cell) Zero() bool {
	return c.Char == 0 && c.Foreground == "" && c.Background == ""
}

// Rect represents a rectangular region.
type Rect struct {
	X, Y, W, H int
}

// Contains checks if point (px, py) is inside the rect.
func (r Rect) Contains(px, py int) bool {
	return px >= r.X && px < r.X+r.W && py >= r.Y && py < r.Y+r.H
}

// Intersect returns the intersection of two rects. Zero rect if no overlap.
func (r Rect) Intersect(other Rect) Rect {
	x1 := max(r.X, other.X)
	y1 := max(r.Y, other.Y)
	x2 := min(r.X+r.W, other.X+other.W)
	y2 := min(r.Y+r.H, other.Y+other.H)
	if x2 <= x1 || y2 <= y1 {
		return Rect{}
	}
	return Rect{X: x1, Y: y1, W: x2 - x1, H: y2 - y1}
}

// Union returns the smallest rect containing both rects.
func (r Rect) Union(other Rect) Rect {
	// If either rect is empty, return the other.
	if r.W <= 0 || r.H <= 0 {
		return other
	}
	if other.W <= 0 || other.H <= 0 {
		return r
	}
	x1 := min(r.X, other.X)
	y1 := min(r.Y, other.Y)
	x2 := max(r.X+r.W, other.X+other.W)
	y2 := max(r.Y+r.H, other.Y+other.H)
	return Rect{X: x1, Y: y1, W: x2 - x1, H: y2 - y1}
}
