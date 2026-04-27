// Package compositor composes multiple component buffers into a single
// screen buffer using z-index ordering and an occlusion map.
package compositor

import "github.com/akzj/lumina/pkg/lumina/v2/buffer"

// Layer represents a Component's buffer positioned on screen.
type Layer struct {
	ID        string         // unique component/window ID
	Buffer    *buffer.Buffer // the component's rendered content
	Rect      buffer.Rect   // screen position and size
	ZIndex    int            // compositing order (higher = on top)
	DirtyRect *buffer.Rect   // sub-region that changed (nil = entire buffer dirty)
}

// OcclusionMap maps each screen cell to the Layer that owns it (highest z-index
// with a non-zero cell at that position).
type OcclusionMap struct {
	owners []*Layer // flat array [y*width + x]
	width  int
	height int
}

// Compositor composes layers into a screen buffer.
type Compositor struct {
	screen *buffer.Buffer
	om     *OcclusionMap
	layers []*Layer
}
