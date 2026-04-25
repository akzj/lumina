//go:build linux || darwin

package lumina

import (
	"io"
	"os"
	"time"
)

// InputReader reads and parses terminal input events.
type InputReader struct {
	term   *Terminal
	reader io.Reader // custom reader (overrides os.Stdin when set)
	events chan AppEvent
	buf    [256]byte
}

// NewInputReader creates an input reader that sends parsed events to the channel.
func NewInputReader(term *Terminal, events chan AppEvent) *InputReader {
	return &InputReader{term: term, events: events}
}

// NewInputReaderFromIO creates an input reader that reads from a TermIO.
func NewInputReaderFromIO(tio TermIO, events chan AppEvent) *InputReader {
	return &InputReader{reader: tio, events: events}
}

// Start begins reading input in a background goroutine.
func (ir *InputReader) Start() {
	go ir.readLoop()
}

func (ir *InputReader) readLoop() {
	r := ir.reader
	if r == nil {
		r = os.Stdin
	}
	for {
		n, err := r.Read(ir.buf[:])
		if err != nil {
			return
		}
		ir.parseInput(ir.buf[:n])
	}
}

// parseInput parses raw terminal bytes into structured events.
func (ir *InputReader) parseInput(data []byte) {
	if len(data) == 0 {
		return
	}

	// ESC sequences
	if data[0] == 0x1b {
		if len(data) == 1 {
			// Bare ESC key
			ir.sendKey("Escape", "Escape", EventModifiers{})
			return
		}
		if len(data) >= 2 && data[1] == '[' {
			ir.parseCSI(data[2:])
			return
		}
		if len(data) >= 2 && data[1] == 'O' {
			ir.parseSS3(data[2:])
			return
		}
		// Alt+key
		if len(data) >= 2 {
			ch := string(data[1:])
			ir.sendKey(ch, ch, EventModifiers{Alt: true})
		}
		return
	}

	// Control characters
	switch data[0] {
	case 0x01: // Ctrl+A
		ir.sendKey("a", "a", EventModifiers{Ctrl: true})
	case 0x02: // Ctrl+B
		ir.sendKey("b", "b", EventModifiers{Ctrl: true})
	case 0x03: // Ctrl+C
		ir.sendKey("c", "c", EventModifiers{Ctrl: true})
	case 0x04: // Ctrl+D
		ir.sendKey("d", "d", EventModifiers{Ctrl: true})
	case 0x05: // Ctrl+E
		ir.sendKey("e", "e", EventModifiers{Ctrl: true})
	case 0x06: // Ctrl+F
		ir.sendKey("f", "f", EventModifiers{Ctrl: true})
	case 0x09: // Tab
		ir.sendKey("Tab", "Tab", EventModifiers{})
	case 0x0a, 0x0d: // Enter (LF or CR)
		ir.sendKey("Enter", "Enter", EventModifiers{})
	case 0x11: // Ctrl+Q
		ir.sendKey("q", "q", EventModifiers{Ctrl: true})
	case 0x17: // Ctrl+W
		ir.sendKey("w", "w", EventModifiers{Ctrl: true})
	case 0x7f: // Backspace
		ir.sendKey("Backspace", "Backspace", EventModifiers{})
	default:
		if data[0] >= 0x01 && data[0] <= 0x1a {
			// Generic Ctrl+letter
			letter := string(rune('a' + data[0] - 1))
			ir.sendKey(letter, letter, EventModifiers{Ctrl: true})
		} else {
			// Regular character(s) — may be UTF-8 multi-byte
			ch := string(data)
			ir.sendKey(ch, ch, EventModifiers{})
		}
	}
}

// parseCSI parses CSI (ESC[) sequences.
func (ir *InputReader) parseCSI(data []byte) {
	if len(data) == 0 {
		return
	}

	// Simple single-char finals
	switch data[0] {
	case 'A':
		ir.sendKey("ArrowUp", "ArrowUp", EventModifiers{})
		return
	case 'B':
		ir.sendKey("ArrowDown", "ArrowDown", EventModifiers{})
		return
	case 'C':
		ir.sendKey("ArrowRight", "ArrowRight", EventModifiers{})
		return
	case 'D':
		ir.sendKey("ArrowLeft", "ArrowLeft", EventModifiers{})
		return
	case 'H':
		ir.sendKey("Home", "Home", EventModifiers{})
		return
	case 'F':
		ir.sendKey("End", "End", EventModifiers{})
		return
	case 'Z':
		// Shift+Tab (reverse tab)
		ir.sendKey("Tab", "Tab", EventModifiers{Shift: true})
		return
	case '<':
		// SGR mouse event: ESC[<button;x;yM or ESC[<button;x;ym
		ir.parseSGRMouse(data[1:])
		return
	}

	// Extended CSI: parse parameters and final byte
	ir.parseExtendedCSI(data)
}

// parseSS3 parses SS3 (ESC O) sequences — function keys on some terminals.
func (ir *InputReader) parseSS3(data []byte) {
	if len(data) == 0 {
		return
	}
	switch data[0] {
	case 'P':
		ir.sendKey("F1", "F1", EventModifiers{})
	case 'Q':
		ir.sendKey("F2", "F2", EventModifiers{})
	case 'R':
		ir.sendKey("F3", "F3", EventModifiers{})
	case 'S':
		ir.sendKey("F4", "F4", EventModifiers{})
	case 'H':
		ir.sendKey("Home", "Home", EventModifiers{})
	case 'F':
		ir.sendKey("End", "End", EventModifiers{})
	}
}

// parseExtendedCSI handles parameterized CSI sequences like ESC[3~, ESC[1;5A, etc.
func (ir *InputReader) parseExtendedCSI(data []byte) {
	// Collect parameter bytes (digits and semicolons) and find the final byte
	params := make([]int, 0, 4)
	current := 0
	hasParam := false
	var final byte

	for _, b := range data {
		if b >= '0' && b <= '9' {
			current = current*10 + int(b-'0')
			hasParam = true
		} else if b == ';' {
			params = append(params, current)
			current = 0
			hasParam = false
		} else {
			// Final byte
			if hasParam {
				params = append(params, current)
			}
			final = b
			break
		}
	}

	// Handle tilde sequences: ESC[N~
	if final == '~' && len(params) >= 1 {
		switch params[0] {
		case 2:
			ir.sendKey("Insert", "Insert", EventModifiers{})
		case 3:
			ir.sendKey("Delete", "Delete", EventModifiers{})
		case 5:
			ir.sendKey("PageUp", "PageUp", EventModifiers{})
		case 6:
			ir.sendKey("PageDown", "PageDown", EventModifiers{})
		case 15:
			ir.sendKey("F5", "F5", EventModifiers{})
		case 17:
			ir.sendKey("F6", "F6", EventModifiers{})
		case 18:
			ir.sendKey("F7", "F7", EventModifiers{})
		case 19:
			ir.sendKey("F8", "F8", EventModifiers{})
		case 20:
			ir.sendKey("F9", "F9", EventModifiers{})
		case 21:
			ir.sendKey("F10", "F10", EventModifiers{})
		case 23:
			ir.sendKey("F11", "F11", EventModifiers{})
		case 24:
			ir.sendKey("F12", "F12", EventModifiers{})
		}
		return
	}

	// Handle modified arrow keys: ESC[1;modA  (mod: 2=Shift, 3=Alt, 5=Ctrl, etc.)
	if len(params) >= 2 && (final == 'A' || final == 'B' || final == 'C' || final == 'D') {
		mods := decodeXtermModifier(params[1])
		var key string
		switch final {
		case 'A':
			key = "ArrowUp"
		case 'B':
			key = "ArrowDown"
		case 'C':
			key = "ArrowRight"
		case 'D':
			key = "ArrowLeft"
		}
		ir.sendKey(key, key, mods)
		return
	}

	// Handle modified Home/End: ESC[1;modH / ESC[1;modF
	if len(params) >= 2 && (final == 'H' || final == 'F') {
		mods := decodeXtermModifier(params[1])
		key := "Home"
		if final == 'F' {
			key = "End"
		}
		ir.sendKey(key, key, mods)
		return
	}
}

// parseSGRMouse parses SGR extended mouse format: button;x;yM or button;x;ym
func (ir *InputReader) parseSGRMouse(data []byte) {
	var nums [3]int
	numIdx := 0
	var press bool
	foundFinal := false

	for _, b := range data {
		if b >= '0' && b <= '9' {
			nums[numIdx] = nums[numIdx]*10 + int(b-'0')
		} else if b == ';' {
			numIdx++
			if numIdx > 2 {
				return // malformed
			}
		} else if b == 'M' {
			press = true
			foundFinal = true
			break
		} else if b == 'm' {
			press = false
			foundFinal = true
			break
		}
	}

	if !foundFinal {
		return
	}

	button := nums[0]
	x := nums[1] - 1 // SGR mouse coordinates are 1-based; convert to 0-based
	y := nums[2] - 1

	mods := EventModifiers{
		Shift: button&4 != 0,
		Alt:   button&8 != 0,
		Ctrl:  button&16 != 0,
	}

	// Check motion bit (bit 5 = 32)
	isMotion := button&32 != 0

	// Scroll events (button bit 6 set)
	if button&64 != 0 {
		direction := "up"
		if button&1 != 0 {
			direction = "down"
		}
		ir.events <- AppEvent{
			Type: "input_event",
			Payload: &Event{
				Type:      "scroll",
				Bubbles:   true,
				X:         x,
				Y:         y,
				Button:    direction,
				Modifiers: mods,
				Timestamp: time.Now().UnixMilli(),
			},
		}
		return
	}

	// Button name (mask out motion + modifier bits)
	buttonName := "left"
	switch button & 0x03 {
	case 1:
		buttonName = "middle"
	case 2:
		buttonName = "right"
	}

	// Determine event type
	var eventType string
	if isMotion {
		eventType = "mousemove"
	} else if press {
		eventType = "mousedown"
	} else {
		eventType = "mouseup"
	}

	ir.events <- AppEvent{
		Type: "input_event",
		Payload: &Event{
			Type:      eventType,
			X:         x,
			Y:         y,
			Button:    buttonName,
			Bubbles:   eventBubbles(eventType),
			Modifiers: mods,
			Timestamp: time.Now().UnixMilli(),
		},
	}
}

// sendKey sends a keyboard event to the event channel.
func (ir *InputReader) sendKey(key, code string, mods EventModifiers) {
	ir.events <- AppEvent{
		Type: "input_event",
		Payload: &Event{
			Type:      "keydown",
			Key:       key,
			Code:      code,
			Bubbles:   true,
			Modifiers: mods,
			Timestamp: time.Now().UnixMilli(),
		},
	}
}

// decodeXtermModifier decodes the xterm modifier parameter.
// 1=none, 2=Shift, 3=Alt, 4=Shift+Alt, 5=Ctrl, 6=Shift+Ctrl, 7=Alt+Ctrl, 8=Shift+Alt+Ctrl
func decodeXtermModifier(mod int) EventModifiers {
	mod-- // xterm modifiers are 1-based
	return EventModifiers{
		Shift: mod&1 != 0,
		Alt:   mod&2 != 0,
		Ctrl:  mod&4 != 0,
	}
}
