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

	// Flush flushes buffered output.
	Flush() error

	// Close closes the adapter.
	Close() error
}
