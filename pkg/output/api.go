// Package output provides adapters to render a screen Buffer to different outputs.
package output

import (
	"github.com/akzj/lumina/pkg/buffer"
)

// Adapter writes a screen buffer to an output target.
type Adapter interface {
	// WriteFull writes the entire screen buffer.
	WriteFull(screen *buffer.Buffer) error

	// WriteDirty writes only the changed regions.
	WriteDirty(screen *buffer.Buffer, dirtyRects []buffer.Rect) error

	// SetCursor positions the hardware cursor at (x, y) and makes it visible.
	// If visible is false, the cursor is hidden. Coordinates are 0-based.
	SetCursor(x, y int, visible bool)

	// Flush flushes buffered output.
	Flush() error

	// Close closes the adapter.
	Close() error
}
