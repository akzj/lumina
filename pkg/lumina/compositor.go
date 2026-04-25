// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

// Compositor merges multiple overlay layers onto a base frame.
// Each overlay renders into its own private Frame, then the compositor
// copies non-transparent cells onto the output bottom-up by ZIndex.
type Compositor struct {
	width  int
	height int
}

// NewCompositor creates a new Compositor for the given screen dimensions.
func NewCompositor(width, height int) *Compositor {
	return &Compositor{width: width, height: height}
}

// Compose merges all visible overlays onto the base frame.
// Overlays must be pre-sorted by ZIndex (ascending).
// The base frame is modified in place and returned.
func (c *Compositor) Compose(baseFrame *Frame, overlays []*Overlay) *Frame {
	for _, ov := range overlays {
		if !ov.Visible || ov.VNode == nil {
			continue
		}

		// If modal, dim everything below
		if ov.Modal {
			renderBackdrop(baseFrame)
		}

		// Render overlay into its own temporary frame
		ovFrame := NewFrame(ov.W, ov.H)
		computeFlexLayout(ov.VNode, 0, 0, ov.W, ov.H)
		ovClip := Rect{X: 0, Y: 0, W: ov.W, H: ov.H}
		renderVNode(ovFrame, ov.VNode, ovClip)

		// Composite overlay frame onto output at (ov.X, ov.Y)
		compositeFrame(baseFrame, ovFrame, ov.X, ov.Y)

		// Mark overlay region as dirty so the ANSI adapter re-renders it
		baseFrame.DirtyRects = append(baseFrame.DirtyRects, Rect{
			X: ov.X, Y: ov.Y, W: ov.W, H: ov.H,
		})
	}

	return baseFrame
}

// compositeFrame copies non-transparent cells from src onto dst at offset (dx, dy).
// Transparent cells in src are skipped, preserving whatever is in dst below.
func compositeFrame(dst *Frame, src *Frame, dx, dy int) {
	for y := 0; y < src.Height; y++ {
		dstY := y + dy
		if dstY < 0 || dstY >= dst.Height {
			continue
		}
		for x := 0; x < src.Width; x++ {
			dstX := x + dx
			if dstX < 0 || dstX >= dst.Width {
				continue
			}
			cell := src.Cells[y][x]
			if cell.Transparent {
				continue // preserve lower layer
			}
			dst.Cells[dstY][dstX] = cell
		}
	}
}
