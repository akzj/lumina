package lumina

import (
	"io"
	"os"
	"sync"
)

// TermIO is the abstraction for terminal I/O.
// Both local terminal and WebSocket terminal implement this interface.
// This allows the same Lua app to run locally or over a network.
type TermIO interface {
	io.Writer            // Write terminal output (escape sequences, rendered frames)
	io.Reader            // Read terminal input (keystrokes)
	Size() (int, int)    // Terminal width, height
	SetSize(w, h int)    // Update terminal size (e.g. on resize)
	Close() error
}

// LocalTermIO wraps os.Stdin/os.Stdout to implement TermIO.
// Used for normal local terminal execution.
type LocalTermIO struct {
	mu     sync.Mutex
	width  int
	height int
}

// NewLocalTermIO creates a TermIO backed by stdin/stdout.
func NewLocalTermIO() *LocalTermIO {
	return &LocalTermIO{width: 80, height: 24}
}

// Write sends bytes to os.Stdout.
func (lt *LocalTermIO) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

// Read reads bytes from os.Stdin.
func (lt *LocalTermIO) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

// Size returns the current terminal size.
func (lt *LocalTermIO) Size() (int, int) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	return lt.width, lt.height
}

// SetSize updates the terminal dimensions.
func (lt *LocalTermIO) SetSize(w, h int) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.width = w
	lt.height = h
}

// Close is a no-op for local terminal.
func (lt *LocalTermIO) Close() error {
	return nil
}

// Ensure LocalTermIO implements TermIO at compile time.
var _ TermIO = (*LocalTermIO)(nil)
