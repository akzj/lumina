package lumina

import (
	"bytes"
	"io"
	"sync"
)

// BufferTermIO implements TermIO backed by in-memory buffers.
// Useful for headless testing and CI — no real terminal needed.
type BufferTermIO struct {
	mu     sync.Mutex
	width  int
	height int
	output *bytes.Buffer
	input  *bytes.Buffer
}

// NewBufferTermIO creates a TermIO backed by in-memory buffers.
// If output is nil, a new buffer is allocated.
func NewBufferTermIO(w, h int, output *bytes.Buffer) *BufferTermIO {
	if output == nil {
		output = &bytes.Buffer{}
	}
	return &BufferTermIO{
		width:  w,
		height: h,
		output: output,
		input:  &bytes.Buffer{},
	}
}

// Write writes bytes to the output buffer.
func (b *BufferTermIO) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.output.Write(p)
}

// Read reads from the input buffer. Returns io.EOF when empty.
func (b *BufferTermIO) Read(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n, err := b.input.Read(p)
	if err == io.EOF {
		return 0, io.EOF
	}
	return n, err
}

// Size returns the configured terminal dimensions.
func (b *BufferTermIO) Size() (int, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.width, b.height
}

// SetSize updates the terminal dimensions.
func (b *BufferTermIO) SetSize(w, h int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.width = w
	b.height = h
}

// Close is a no-op for buffer-backed TermIO.
func (b *BufferTermIO) Close() error {
	return nil
}

// Output returns the output buffer contents as a string.
func (b *BufferTermIO) Output() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.output.String()
}

// Reset clears both input and output buffers.
func (b *BufferTermIO) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.output.Reset()
	b.input.Reset()
}

// WriteInput writes data into the input buffer (simulates keystrokes).
func (b *BufferTermIO) WriteInput(data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.input.Write(data)
}

// Ensure BufferTermIO implements TermIO at compile time.
var _ TermIO = (*BufferTermIO)(nil)
