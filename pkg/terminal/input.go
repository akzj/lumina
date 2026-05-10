//go:build linux || darwin

package terminal

import (
	"io"
	"log"
	"os"
	"sync"
	"unicode/utf8"
)

var inputDebugLog *log.Logger
var inputDebugOnce sync.Once

func getInputDebugLog() *log.Logger {
	inputDebugOnce.Do(func() {
		f, err := os.OpenFile("input_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return
		}
		inputDebugLog = log.New(f, "", log.LstdFlags|log.Lmicroseconds)
	})
	return inputDebugLog
}

// InputEvent represents a parsed terminal input event.
// The caller converts InputEvent → event.Event to avoid circular deps.
type InputEvent struct {
	Type      string // "keydown", "mousedown", "mouseup", "mousemove", "scroll"
	Key       string // key name: "a", "Enter", "ArrowUp", "Escape", "Tab", "F1", etc.
	Char      string // printable character (for text input); empty for non-printable
	X, Y      int    // mouse position (0-based)
	Button    string // mouse: "left","middle","right"; wheel: "up","down","left","right"
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
	var parser inputParser
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
		if n > 0 {
			if dl := getInputDebugLog(); dl != nil {
				dl.Printf("[RAW] %d bytes: %x | ascii: %q", n, buf[:n], buf[:n])
			}
		}
		if n == 0 {
			events := parser.Flush()
			for i := range events {
				select {
				case ir.events <- events[i]:
				case <-ir.done:
					return
				}
			}
			continue
		}

		events := parser.Parse(buf[:n])
		if dl := getInputDebugLog(); dl != nil {
			for _, ev := range events {
				dl.Printf("[EVENT] type=%s key=%q char=%q mods={ctrl=%v alt=%v shift=%v}",
					ev.Type, ev.Key, ev.Char, ev.Modifiers.Ctrl, ev.Modifiers.Alt, ev.Modifiers.Shift)
			}
		}
		for i := range events {
			select {
			case ir.events <- events[i]:
			case <-ir.done:
				return
			}
		}
	}
}

type inputParser struct {
	pending []byte
}

func (p *inputParser) Parse(data []byte) []InputEvent {
	if len(data) > 0 {
		p.pending = append(p.pending, data...)
	}

	var events []InputEvent
	for len(p.pending) > 0 {
		consumed, evs, needMore := parseOneStreamEvent(p.pending)
		if needMore {
			break
		}
		if consumed <= 0 {
			p.pending = p.pending[1:]
			continue
		}
		events = append(events, evs...)
		p.pending = p.pending[consumed:]
	}
	return events
}

func (p *inputParser) Flush() []InputEvent {
	if len(p.pending) == 0 {
		return nil
	}
	if len(p.pending) == 1 && p.pending[0] == 0x1b {
		p.pending = p.pending[:0]
		return []InputEvent{{Type: "keydown", Key: "Escape"}}
	}
	// Drop timed-out partial escape/UTF-8 sequences. They are terminal protocol
	// fragments, not user text, and must not leak into focused inputs.
	p.pending = p.pending[:0]
	return nil
}

func parseOneStreamEvent(data []byte) (consumed int, events []InputEvent, needMore bool) {
	if len(data) == 0 {
		return 0, nil, true
	}

	if data[0] == 0x1b {
		return parseOneEscapeEvent(data)
	}

	if data[0] < 0x20 || data[0] == 0x7f {
		return 1, []InputEvent{parseControl(data[0])}, false
	}

	if !utf8.FullRune(data) {
		return 0, nil, true
	}
	_, size := utf8.DecodeRune(data)
	return size, parseText(data[:size]), false
}

func parseOneEscapeEvent(data []byte) (consumed int, events []InputEvent, needMore bool) {
	if len(data) == 1 {
		return 0, nil, true
	}

	switch data[1] {
	case '[':
		if len(data) == 2 {
			return 0, nil, true
		}
		if data[2] == '<' {
			end := findSGRMouseEnd(data[3:])
			if end < 0 {
				return 0, nil, true
			}
			n := 3 + end + 1
			return n, parseInput(data[:n]), false
		}
		if isSimpleCSIFinal(data[2]) {
			return 3, parseInput(data[:3]), false
		}
		end := findCSIFinal(data[2:])
		if end < 0 {
			return 0, nil, true
		}
		n := 2 + end + 1
		return n, parseInput(data[:n]), false
	case 'O':
		if len(data) < 3 {
			return 0, nil, true
		}
		return 3, parseInput(data[:3]), false
	case 0x0d, 0x0a:
		return 2, parseInput(data[:2]), false
	default:
		if !utf8.FullRune(data[1:]) {
			return 0, nil, true
		}
		_, size := utf8.DecodeRune(data[1:])
		n := 1 + size
		return n, parseInput(data[:n]), false
	}
}

func findSGRMouseEnd(data []byte) int {
	for i, b := range data {
		if b == 'M' || b == 'm' {
			return i
		}
	}
	return -1
}

func isSimpleCSIFinal(b byte) bool {
	switch b {
	case 'A', 'B', 'C', 'D', 'H', 'F', 'Z':
		return true
	default:
		return false
	}
}

func findCSIFinal(data []byte) int {
	for i, b := range data {
		if (b >= 'A' && b <= 'Z') || b == '~' {
			return i
		}
	}
	return -1
}
