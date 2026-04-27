//go:build linux || darwin

package terminal

import (
	"io"
	"os"
)

// InputEvent represents a parsed terminal input event.
// The caller converts InputEvent → event.Event to avoid circular deps.
type InputEvent struct {
	Type      string // "keydown", "mousedown", "mouseup", "mousemove", "scroll"
	Key       string // key name: "a", "Enter", "ArrowUp", "Escape", "Tab", "F1", etc.
	Char      string // printable character (for text input); empty for non-printable
	X, Y      int    // mouse position (0-based)
	Button    string // "left", "middle", "right", "up", "down" (scroll direction)
	Modifiers Modifiers
}

// Modifiers represents keyboard modifier state.
type Modifiers struct {
	Ctrl  bool
	Alt   bool
	Shift bool
}

// InputReader reads terminal input in a goroutine and sends parsed events.
type InputReader struct {
	reader io.Reader
	events chan<- InputEvent
	done   chan struct{}
}

// NewInputReader creates an InputReader that reads from os.Stdin.
func NewInputReader(events chan<- InputEvent) *InputReader {
	return &InputReader{
		reader: os.Stdin,
		events: events,
		done:   make(chan struct{}),
	}
}

// NewInputReaderFrom creates an InputReader that reads from the given reader.
func NewInputReaderFrom(r io.Reader, events chan<- InputEvent) *InputReader {
	return &InputReader{
		reader: r,
		events: events,
		done:   make(chan struct{}),
	}
}

// Start begins reading input in a background goroutine.
func (ir *InputReader) Start() {
	go ir.readLoop()
}

// Stop signals the readLoop to exit. Note: the goroutine will actually exit
// when the next Read call returns (which may require closing the reader).
func (ir *InputReader) Stop() {
	select {
	case <-ir.done:
		// already stopped
	default:
		close(ir.done)
	}
}

func (ir *InputReader) readLoop() {
	var buf [256]byte
	for {
		select {
		case <-ir.done:
			return
		default:
		}

		n, err := ir.reader.Read(buf[:])
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}

		events := parseInput(buf[:n])
		for i := range events {
			select {
			case ir.events <- events[i]:
			case <-ir.done:
				return
			}
		}
	}
}
