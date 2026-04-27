package compositor

import (
	"sort"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
)

// NewOcclusionMap creates an occlusion map of the given screen size.
func NewOcclusionMap(w, h int) *OcclusionMap {
	if w <= 0 || h <= 0 {
		return &OcclusionMap{}
	}
	return &OcclusionMap{
		owners: make([]*Layer, w*h),
		width:  w,
		height: h,
	}
}

// Build rebuilds the occlusion map from layers.
// Layers are processed from LOWEST z-index to HIGHEST so the highest z wins.
// Only non-zero cells in a layer's buffer claim ownership — transparent cells
// let lower layers show through.
func (om *OcclusionMap) Build(layers []*Layer) {
	// Clear all owners.
	clear(om.owners)

	if len(layers) == 0 {
		return
	}

	// Sort layers by ZIndex ascending (lowest first).
	sorted := make([]*Layer, len(layers))
	copy(sorted, layers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ZIndex < sorted[j].ZIndex
	})

	screenBounds := buffer.Rect{X: 0, Y: 0, W: om.width, H: om.height}

	for _, layer := range sorted {
		if layer.Buffer == nil {
			continue
		}
		// Clip layer rect to screen bounds.
		visible := layer.Rect.Intersect(screenBounds)
		if visible.W <= 0 || visible.H <= 0 {
			continue
		}

		for y := visible.Y; y < visible.Y+visible.H; y++ {
			for x := visible.X; x < visible.X+visible.W; x++ {
				// Local coordinates in the layer's buffer.
				localX := x - layer.Rect.X
				localY := y - layer.Rect.Y
				cell := layer.Buffer.Get(localX, localY)
				if !cell.Zero() {
					om.owners[y*om.width+x] = layer
				}
			}
		}
	}
}

// Owner returns the layer ID that owns cell (x, y). Empty string if no owner.
func (om *OcclusionMap) Owner(x, y int) string {
	if x < 0 || y < 0 || x >= om.width || y >= om.height {
		return ""
	}
	l := om.owners[y*om.width+x]
	if l == nil {
		return ""
	}
	return l.ID
}

// OwnerLayer returns the Layer that owns cell (x, y). Nil if no owner.
func (om *OcclusionMap) OwnerLayer(x, y int) *Layer {
	if x < 0 || y < 0 || x >= om.width || y >= om.height {
		return nil
	}
	return om.owners[y*om.width+x]
}

// NewCompositor creates a compositor with the given screen dimensions.
func NewCompositor(w, h int) *Compositor {
	return &Compositor{
		screen: buffer.New(w, h),
		om:     NewOcclusionMap(w, h),
	}
}

// SetLayers sets the full layer stack and rebuilds the occlusion map.
func (c *Compositor) SetLayers(layers []*Layer) {
	c.layers = layers
	c.om.Build(layers)
}

// ComposeAll composes all layers into the screen buffer. Returns the screen.
func (c *Compositor) ComposeAll() *buffer.Buffer {
	c.screen.Clear()

	screenH := c.screen.Height()
	screenW := c.screen.Width()

	for y := 0; y < screenH; y++ {
		for x := 0; x < screenW; x++ {
			layer := c.om.OwnerLayer(x, y)
			if layer == nil {
				continue
			}
			localX := x - layer.Rect.X
			localY := y - layer.Rect.Y
			c.screen.Set(x, y, layer.Buffer.Get(localX, localY))
		}
	}
	return c.screen
}

// ComposeDirty composes only changed regions.
// For each dirty layer, blits the dirty sub-region (or entire layer rect)
// but only writes cells where the occlusion map says this layer owns the cell.
// Returns the dirty rects on screen.
func (c *Compositor) ComposeDirty(dirtyLayers []*Layer) []buffer.Rect {
	if len(dirtyLayers) == 0 {
		return nil
	}

	screenBounds := buffer.Rect{X: 0, Y: 0, W: c.screen.Width(), H: c.screen.Height()}
	var dirtyRects []buffer.Rect

	for _, layer := range dirtyLayers {
		if layer.Buffer == nil {
			continue
		}

		// Determine the dirty region in screen coordinates.
		var rect buffer.Rect
		if layer.DirtyRect != nil {
			// DirtyRect is in local coordinates — convert to screen coords.
			rect = buffer.Rect{
				X: layer.Rect.X + layer.DirtyRect.X,
				Y: layer.Rect.Y + layer.DirtyRect.Y,
				W: layer.DirtyRect.W,
				H: layer.DirtyRect.H,
			}
		} else {
			rect = layer.Rect
		}

		// Clip to screen bounds.
		clipped := rect.Intersect(screenBounds)
		if clipped.W <= 0 || clipped.H <= 0 {
			continue
		}

		for y := clipped.Y; y < clipped.Y+clipped.H; y++ {
			for x := clipped.X; x < clipped.X+clipped.W; x++ {
				if c.om.Owner(x, y) == layer.ID {
					localX := x - layer.Rect.X
					localY := y - layer.Rect.Y
					c.screen.Set(x, y, layer.Buffer.Get(localX, localY))
				}
			}
		}

		dirtyRects = append(dirtyRects, clipped)
	}

	return dirtyRects
}

// ComposeRects recomposes specific screen regions by looking up the occlusion map.
// Unowned cells are cleared. Returns the recomposed rects.
func (c *Compositor) ComposeRects(rects []buffer.Rect) []buffer.Rect {
	if len(rects) == 0 {
		return nil
	}

	screenBounds := buffer.Rect{X: 0, Y: 0, W: c.screen.Width(), H: c.screen.Height()}
	var result []buffer.Rect

	for _, rect := range rects {
		clipped := rect.Intersect(screenBounds)
		if clipped.W <= 0 || clipped.H <= 0 {
			continue
		}

		for y := clipped.Y; y < clipped.Y+clipped.H; y++ {
			for x := clipped.X; x < clipped.X+clipped.W; x++ {
				layer := c.om.OwnerLayer(x, y)
				if layer != nil {
					localX := x - layer.Rect.X
					localY := y - layer.Rect.Y
					c.screen.Set(x, y, layer.Buffer.Get(localX, localY))
				} else {
					c.screen.Set(x, y, buffer.Cell{})
				}
			}
		}

		result = append(result, clipped)
	}

	return result
}

// Screen returns the current screen buffer.
func (c *Compositor) Screen() *buffer.Buffer {
	return c.screen
}

// OcclusionMap returns the current occlusion map.
func (c *Compositor) OcclusionMap() *OcclusionMap {
	return c.om
}
