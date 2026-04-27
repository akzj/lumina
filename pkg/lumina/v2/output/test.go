package output

import (
	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
)

// TestAdapter captures rendered output for test assertions.
// It is exported so tests outside this package can use it.
type TestAdapter struct {
	LastScreen *buffer.Buffer
	DirtyRects []buffer.Rect
	WriteCount int
}

// NewTestAdapter creates a new test adapter.
func NewTestAdapter() *TestAdapter {
	return &TestAdapter{}
}

// WriteFull clones the screen buffer and stores it.
func (t *TestAdapter) WriteFull(screen *buffer.Buffer) error {
	t.LastScreen = cloneBuffer(screen)
	t.DirtyRects = nil
	t.WriteCount++
	return nil
}

// WriteDirty clones the screen buffer and records the dirty rects.
func (t *TestAdapter) WriteDirty(screen *buffer.Buffer, dirtyRects []buffer.Rect) error {
	t.LastScreen = cloneBuffer(screen)
	t.DirtyRects = make([]buffer.Rect, len(dirtyRects))
	copy(t.DirtyRects, dirtyRects)
	t.WriteCount++
	return nil
}

// Flush is a no-op for test adapter.
func (t *TestAdapter) Flush() error { return nil }

// Close is a no-op for test adapter.
func (t *TestAdapter) Close() error { return nil }

// cloneBuffer creates a deep copy of a buffer.
func cloneBuffer(src *buffer.Buffer) *buffer.Buffer {
	w, h := src.Width(), src.Height()
	dst := buffer.New(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(x, y, src.Get(x, y))
		}
	}
	return dst
}
