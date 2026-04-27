// Package paint renders a VNode tree (with layout computed) into a Buffer.
package paint

import (
	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// Painter paints a VNode tree into a Buffer.
type Painter interface {
	// Paint renders the VNode tree into the given buffer.
	// The VNode must have layout computed (X, Y, W, H set as absolute screen coords).
	// offsetX, offsetY translate absolute VNode coords to buffer-local coords:
	//   bufferX = vnode.X - offsetX
	//   bufferY = vnode.Y - offsetY
	//
	// For a Component at screen position (10, 5):
	//   Paint(comp.Buffer, comp.VNodeTree, 10, 5)
	// A VNode at absolute (12, 7) paints to buffer position (2, 2).
	Paint(buf *buffer.Buffer, root *layout.VNode, offsetX, offsetY int)
}

// NewPainter creates a new Painter.
func NewPainter() Painter {
	return &painter{}
}
