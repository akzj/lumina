//go:build linux || darwin

package terminal

import "unicode/utf8"

// parseInput parses raw terminal bytes into InputEvents.
// Returns a slice because one read may contain multiple events.
func parseInput(data []byte) []InputEvent {
	if len(data) == 0 {
		return nil
	}

	// ESC sequences
	if data[0] == 0x1b {
		return parseEscape(data)
	}

	// Control characters
	if data[0] < 0x20 || data[0] == 0x7f {
		return []InputEvent{parseControl(data[0])}
	}

	// Regular characters (UTF-8)
	return parseText(data)
}

// parseEscape handles ESC-prefixed sequences.
func parseEscape(data []byte) []InputEvent {
	if len(data) == 1 {
		return []InputEvent{{Type: "keydown", Key: "Escape"}}
	}

	switch data[1] {
	case '[':
		return parseCSI(data[2:])
	case 'O':
		return parseSS3(data[2:])
	default:
		// Alt+key: ESC followed by a character
		ch := string(data[1:])
		return []InputEvent{{
			Type:      "keydown",
			Key:       ch,
			Char:      ch,
			Modifiers: Modifiers{Alt: true},
		}}
	}
}

// parseCSI handles CSI (ESC[) sequences.
func parseCSI(data []byte) []InputEvent {
	if len(data) == 0 {
		return nil
	}

	// Simple single-char finals
	switch data[0] {
	case 'A':
		return []InputEvent{{Type: "keydown", Key: "ArrowUp"}}
	case 'B':
		return []InputEvent{{Type: "keydown", Key: "ArrowDown"}}
	case 'C':
		return []InputEvent{{Type: "keydown", Key: "ArrowRight"}}
	case 'D':
		return []InputEvent{{Type: "keydown", Key: "ArrowLeft"}}
	case 'H':
		return []InputEvent{{Type: "keydown", Key: "Home"}}
	case 'F':
		return []InputEvent{{Type: "keydown", Key: "End"}}
	case 'Z':
		return []InputEvent{{Type: "keydown", Key: "Tab", Modifiers: Modifiers{Shift: true}}}
	case '<':
		return parseSGRMouse(data[1:])
	}

	// Extended CSI: parameterized sequences
	return parseExtendedCSI(data)
}

// parseSS3 handles SS3 (ESC O) sequences — function keys on some terminals.
func parseSS3(data []byte) []InputEvent {
	if len(data) == 0 {
		return nil
	}
	switch data[0] {
	case 'P':
		return []InputEvent{{Type: "keydown", Key: "F1"}}
	case 'Q':
		return []InputEvent{{Type: "keydown", Key: "F2"}}
	case 'R':
		return []InputEvent{{Type: "keydown", Key: "F3"}}
	case 'S':
		return []InputEvent{{Type: "keydown", Key: "F4"}}
	case 'H':
		return []InputEvent{{Type: "keydown", Key: "Home"}}
	case 'F':
		return []InputEvent{{Type: "keydown", Key: "End"}}
	}
	return nil
}

// parseExtendedCSI handles parameterized CSI sequences like ESC[3~, ESC[1;5A.
func parseExtendedCSI(data []byte) []InputEvent {
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
			if hasParam {
				params = append(params, current)
			}
			final = b
			break
		}
	}

	// Tilde sequences: ESC[N~
	if final == '~' && len(params) >= 1 {
		key := tildeKeyName(params[0])
		if key != "" {
			return []InputEvent{{Type: "keydown", Key: key}}
		}
		return nil
	}

	// Modified arrow keys: ESC[1;modA
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
		return []InputEvent{{Type: "keydown", Key: key, Modifiers: mods}}
	}

	// Modified Home/End: ESC[1;modH / ESC[1;modF
	if len(params) >= 2 && (final == 'H' || final == 'F') {
		mods := decodeXtermModifier(params[1])
		key := "Home"
		if final == 'F' {
			key = "End"
		}
		return []InputEvent{{Type: "keydown", Key: key, Modifiers: mods}}
	}

	return nil
}

// tildeKeyName maps the numeric param in ESC[N~ to a key name.
func tildeKeyName(n int) string {
	switch n {
	case 2:
		return "Insert"
	case 3:
		return "Delete"
	case 5:
		return "PageUp"
	case 6:
		return "PageDown"
	case 15:
		return "F5"
	case 17:
		return "F6"
	case 18:
		return "F7"
	case 19:
		return "F8"
	case 20:
		return "F9"
	case 21:
		return "F10"
	case 23:
		return "F11"
	case 24:
		return "F12"
	}
	return ""
}

// parseSGRMouse parses SGR extended mouse format: button;x;yM or button;x;ym
// (data starts after the '<' character)
func parseSGRMouse(data []byte) []InputEvent {
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
				return nil // malformed
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
		return nil
	}

	button := nums[0]
	x := nums[1] - 1 // SGR coordinates are 1-based; convert to 0-based
	y := nums[2] - 1

	mods := Modifiers{
		Shift: button&4 != 0,
		Alt:   button&8 != 0,
		Ctrl:  button&16 != 0,
	}

	// Scroll events (bit 6 = 64)
	if button&64 != 0 {
		direction := "up"
		if button&1 != 0 {
			direction = "down"
		}
		return []InputEvent{{
			Type:      "scroll",
			X:         x,
			Y:         y,
			Button:    direction,
			Modifiers: mods,
		}}
	}

	// Button name (bits 0-1)
	buttonName := "left"
	switch button & 0x03 {
	case 1:
		buttonName = "middle"
	case 2:
		buttonName = "right"
	}

	// Motion bit (bit 5 = 32)
	isMotion := button&32 != 0

	var eventType string
	if isMotion {
		eventType = "mousemove"
	} else if press {
		eventType = "mousedown"
	} else {
		eventType = "mouseup"
	}

	return []InputEvent{{
		Type:      eventType,
		X:         x,
		Y:         y,
		Button:    buttonName,
		Modifiers: mods,
	}}
}

// parseControl handles single control characters.
func parseControl(b byte) InputEvent {
	switch b {
	case 0x09:
		return InputEvent{Type: "keydown", Key: "Tab"}
	case 0x0a, 0x0d:
		return InputEvent{Type: "keydown", Key: "Enter"}
	case 0x7f:
		return InputEvent{Type: "keydown", Key: "Backspace"}
	default:
		// Ctrl+letter: 0x01=Ctrl+a, 0x02=Ctrl+b, etc.
		if b >= 0x01 && b <= 0x1a {
			letter := string(rune('a' + rune(b) - 1))
			return InputEvent{
				Type:      "keydown",
				Key:       letter,
				Modifiers: Modifiers{Ctrl: true},
			}
		}
		return InputEvent{Type: "keydown", Key: ""}
	}
}

// parseText handles regular printable UTF-8 text.
func parseText(data []byte) []InputEvent {
	var events []InputEvent
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size <= 1 {
			// Invalid UTF-8, skip byte
			data = data[1:]
			continue
		}
		ch := string(r)
		events = append(events, InputEvent{
			Type: "keydown",
			Key:  ch,
			Char: ch,
		})
		data = data[size:]
	}
	return events
}

// decodeXtermModifier decodes the xterm modifier parameter.
// Xterm modifiers are 1-based: 1=none, 2=Shift, 3=Alt, 5=Ctrl, etc.
// Formula: mod-1, then bit 0=Shift, bit 1=Alt, bit 2=Ctrl.
func decodeXtermModifier(mod int) Modifiers {
	mod-- // xterm modifiers are 1-based
	return Modifiers{
		Shift: mod&1 != 0,
		Alt:   mod&2 != 0,
		Ctrl:  mod&4 != 0,
	}
}
