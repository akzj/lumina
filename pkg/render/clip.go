package render

// ClipRegion represents an immutable rectangular clip area.
// It can only be narrowed (intersection), never expanded.
type ClipRegion struct {
	x1, y1 int // inclusive
	x2, y2 int // exclusive
}

// FullClip returns a ClipRegion that covers the entire buffer.
func FullClip(width, height int) ClipRegion {
	return ClipRegion{x1: 0, y1: 0, x2: width, y2: height}
}

// Narrow returns a new ClipRegion that is the intersection of this region
// and the rectangle (x, y, w, h). The result is always <= the original.
func (c ClipRegion) Narrow(x, y, w, h int) ClipRegion {
	return ClipRegion{
		x1: max(c.x1, x),
		y1: max(c.y1, y),
		x2: min(c.x2, x+w),
		y2: min(c.y2, y+h),
	}
}

// Empty returns true if the clip region has zero area.
func (c ClipRegion) Empty() bool {
	return c.x1 >= c.x2 || c.y1 >= c.y2
}

// Contains returns true if (x, y) is within the clip region.
func (c ClipRegion) Contains(x, y int) bool {
	return x >= c.x1 && x < c.x2 && y >= c.y1 && y < c.y2
}

// Bounds returns the clip rectangle as (x1, y1, x2, y2).
func (c ClipRegion) Bounds() (x1, y1, x2, y2 int) {
	return c.x1, c.y1, c.x2, c.y2
}
