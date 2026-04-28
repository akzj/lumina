//go:build linux || darwin

package terminal

import (
	"testing"
)

// helper: parse and assert exactly one event
func mustParseOne(t *testing.T, data []byte) InputEvent {
	t.Helper()
	events := parseInput(data)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d: %+v", len(events), events)
	}
	return events[0]
}

func assertKey(t *testing.T, e InputEvent, key string) {
	t.Helper()
	if e.Type != "keydown" {
		t.Errorf("expected type keydown, got %q", e.Type)
	}
	if e.Key != key {
		t.Errorf("expected key %q, got %q", key, e.Key)
	}
}

func assertMods(t *testing.T, e InputEvent, ctrl, alt, shift bool) {
	t.Helper()
	if e.Modifiers.Ctrl != ctrl {
		t.Errorf("Ctrl: expected %v, got %v", ctrl, e.Modifiers.Ctrl)
	}
	if e.Modifiers.Alt != alt {
		t.Errorf("Alt: expected %v, got %v", alt, e.Modifiers.Alt)
	}
	if e.Modifiers.Shift != shift {
		t.Errorf("Shift: expected %v, got %v", shift, e.Modifiers.Shift)
	}
}

// --- Basic key tests ---

func TestParse_SingleChar(t *testing.T) {
	for _, ch := range []byte{'a', 'Z', '1', ' ', '/'} {
		e := mustParseOne(t, []byte{ch})
		assertKey(t, e, string(ch))
		if e.Char != string(ch) {
			t.Errorf("char %q: expected Char=%q, got %q", string(ch), string(ch), e.Char)
		}
	}
}

func TestParse_UTF8(t *testing.T) {
	// Chinese character: 中 = 0xe4 0xb8 0xad
	e := mustParseOne(t, []byte{0xe4, 0xb8, 0xad})
	assertKey(t, e, "中")
	if e.Char != "中" {
		t.Errorf("expected Char=中, got %q", e.Char)
	}

	// Emoji: 🎉 = 0xf0 0x9f 0x8e 0x89
	e = mustParseOne(t, []byte{0xf0, 0x9f, 0x8e, 0x89})
	assertKey(t, e, "🎉")
	if e.Char != "🎉" {
		t.Errorf("expected Char=🎉, got %q", e.Char)
	}
}

func TestParse_Enter(t *testing.T) {
	// CR
	e := mustParseOne(t, []byte{0x0d})
	assertKey(t, e, "Enter")

	// LF
	e = mustParseOne(t, []byte{0x0a})
	assertKey(t, e, "Enter")
}

func TestParse_Tab(t *testing.T) {
	e := mustParseOne(t, []byte{0x09})
	assertKey(t, e, "Tab")
}

func TestParse_Backspace(t *testing.T) {
	e := mustParseOne(t, []byte{0x7f})
	assertKey(t, e, "Backspace")
}

func TestParse_Escape(t *testing.T) {
	e := mustParseOne(t, []byte{0x1b})
	assertKey(t, e, "Escape")
}

// --- Ctrl keys ---

func TestParse_CtrlC(t *testing.T) {
	e := mustParseOne(t, []byte{0x03})
	assertKey(t, e, "c")
	assertMods(t, e, true, false, false)
}

func TestParse_CtrlA(t *testing.T) {
	e := mustParseOne(t, []byte{0x01})
	assertKey(t, e, "a")
	assertMods(t, e, true, false, false)
}

// --- Arrow keys ---

func TestParse_ArrowKeys(t *testing.T) {
	tests := []struct {
		data []byte
		key  string
	}{
		{[]byte{0x1b, '[', 'A'}, "ArrowUp"},
		{[]byte{0x1b, '[', 'B'}, "ArrowDown"},
		{[]byte{0x1b, '[', 'C'}, "ArrowRight"},
		{[]byte{0x1b, '[', 'D'}, "ArrowLeft"},
	}
	for _, tt := range tests {
		e := mustParseOne(t, tt.data)
		assertKey(t, e, tt.key)
	}
}

func TestParse_HomeEnd(t *testing.T) {
	// CSI H / F
	e := mustParseOne(t, []byte{0x1b, '[', 'H'})
	assertKey(t, e, "Home")

	e = mustParseOne(t, []byte{0x1b, '[', 'F'})
	assertKey(t, e, "End")

	// SS3 H / F
	e = mustParseOne(t, []byte{0x1b, 'O', 'H'})
	assertKey(t, e, "Home")

	e = mustParseOne(t, []byte{0x1b, 'O', 'F'})
	assertKey(t, e, "End")
}

func TestParse_ShiftTab(t *testing.T) {
	e := mustParseOne(t, []byte{0x1b, '[', 'Z'})
	assertKey(t, e, "Tab")
	assertMods(t, e, false, false, true)
}

// --- Function keys ---

func TestParse_FunctionKeys(t *testing.T) {
	tests := []struct {
		data []byte
		key  string
	}{
		{[]byte{0x1b, '[', '1', '5', '~'}, "F5"},
		{[]byte{0x1b, '[', '1', '7', '~'}, "F6"},
		{[]byte{0x1b, '[', '1', '8', '~'}, "F7"},
		{[]byte{0x1b, '[', '1', '9', '~'}, "F8"},
		{[]byte{0x1b, '[', '2', '0', '~'}, "F9"},
		{[]byte{0x1b, '[', '2', '1', '~'}, "F10"},
		{[]byte{0x1b, '[', '2', '3', '~'}, "F11"},
		{[]byte{0x1b, '[', '2', '4', '~'}, "F12"},
	}
	for _, tt := range tests {
		e := mustParseOne(t, tt.data)
		assertKey(t, e, tt.key)
	}
}

func TestParse_SS3_FunctionKeys(t *testing.T) {
	tests := []struct {
		data []byte
		key  string
	}{
		{[]byte{0x1b, 'O', 'P'}, "F1"},
		{[]byte{0x1b, 'O', 'Q'}, "F2"},
		{[]byte{0x1b, 'O', 'R'}, "F3"},
		{[]byte{0x1b, 'O', 'S'}, "F4"},
	}
	for _, tt := range tests {
		e := mustParseOne(t, tt.data)
		assertKey(t, e, tt.key)
	}
}

func TestParse_Delete(t *testing.T) {
	e := mustParseOne(t, []byte{0x1b, '[', '3', '~'})
	assertKey(t, e, "Delete")
}

func TestParse_PageUpDown(t *testing.T) {
	e := mustParseOne(t, []byte{0x1b, '[', '5', '~'})
	assertKey(t, e, "PageUp")

	e = mustParseOne(t, []byte{0x1b, '[', '6', '~'})
	assertKey(t, e, "PageDown")
}

// --- Modified keys ---

func TestParse_ModifiedArrow(t *testing.T) {
	// ESC[1;5A = Ctrl+ArrowUp
	e := mustParseOne(t, []byte{0x1b, '[', '1', ';', '5', 'A'})
	assertKey(t, e, "ArrowUp")
	assertMods(t, e, true, false, false)

	// ESC[1;2B = Shift+ArrowDown
	e = mustParseOne(t, []byte{0x1b, '[', '1', ';', '2', 'B'})
	assertKey(t, e, "ArrowDown")
	assertMods(t, e, false, false, true)

	// ESC[1;3C = Alt+ArrowRight
	e = mustParseOne(t, []byte{0x1b, '[', '1', ';', '3', 'C'})
	assertKey(t, e, "ArrowRight")
	assertMods(t, e, false, true, false)

	// ESC[1;8D = Shift+Alt+Ctrl+ArrowLeft
	e = mustParseOne(t, []byte{0x1b, '[', '1', ';', '8', 'D'})
	assertKey(t, e, "ArrowLeft")
	assertMods(t, e, true, true, true)
}

func TestParse_AltKey(t *testing.T) {
	// ESC + 'a' → Alt+a
	e := mustParseOne(t, []byte{0x1b, 'a'})
	assertKey(t, e, "a")
	assertMods(t, e, false, true, false)
}

// --- Mouse events ---

func TestParse_MouseClick(t *testing.T) {
	// ESC[<0;10;5M → left click at (9, 4)
	data := []byte{0x1b, '[', '<', '0', ';', '1', '0', ';', '5', 'M'}
	e := mustParseOne(t, data)
	if e.Type != "mousedown" {
		t.Errorf("expected mousedown, got %q", e.Type)
	}
	if e.X != 9 || e.Y != 4 {
		t.Errorf("expected (9,4), got (%d,%d)", e.X, e.Y)
	}
	if e.Button != "left" {
		t.Errorf("expected left, got %q", e.Button)
	}
}

func TestParse_MouseRelease(t *testing.T) {
	// ESC[<0;10;5m → left release at (9, 4)
	data := []byte{0x1b, '[', '<', '0', ';', '1', '0', ';', '5', 'm'}
	e := mustParseOne(t, data)
	if e.Type != "mouseup" {
		t.Errorf("expected mouseup, got %q", e.Type)
	}
	if e.X != 9 || e.Y != 4 {
		t.Errorf("expected (9,4), got (%d,%d)", e.X, e.Y)
	}
}

func TestParse_MouseMove(t *testing.T) {
	// ESC[<32;10;5M → motion at (9, 4)
	data := []byte{0x1b, '[', '<', '3', '2', ';', '1', '0', ';', '5', 'M'}
	e := mustParseOne(t, data)
	if e.Type != "mousemove" {
		t.Errorf("expected mousemove, got %q", e.Type)
	}
	if e.X != 9 || e.Y != 4 {
		t.Errorf("expected (9,4), got (%d,%d)", e.X, e.Y)
	}
}

func TestParse_MouseRight(t *testing.T) {
	// ESC[<2;10;5M → right click
	data := []byte{0x1b, '[', '<', '2', ';', '1', '0', ';', '5', 'M'}
	e := mustParseOne(t, data)
	if e.Type != "mousedown" {
		t.Errorf("expected mousedown, got %q", e.Type)
	}
	if e.Button != "right" {
		t.Errorf("expected right, got %q", e.Button)
	}
}

func TestParse_MouseScroll(t *testing.T) {
	// ESC[<64;10;5M → scroll up
	data := []byte{0x1b, '[', '<', '6', '4', ';', '1', '0', ';', '5', 'M'}
	e := mustParseOne(t, data)
	if e.Type != "scroll" {
		t.Errorf("expected scroll, got %q", e.Type)
	}
	if e.Button != "up" {
		t.Errorf("expected up, got %q", e.Button)
	}
	if e.X != 9 || e.Y != 4 {
		t.Errorf("expected (9,4), got (%d,%d)", e.X, e.Y)
	}

	// ESC[<65;10;5M → scroll down
	data = []byte{0x1b, '[', '<', '6', '5', ';', '1', '0', ';', '5', 'M'}
	e = mustParseOne(t, data)
	if e.Type != "scroll" {
		t.Errorf("expected scroll, got %q", e.Type)
	}
	if e.Button != "down" {
		t.Errorf("expected down, got %q", e.Button)
	}
}

func TestParse_MouseWithModifiers(t *testing.T) {
	// ESC[<20;10;5M → Ctrl+left (button 0 + Ctrl bit 16 = 16, but 16&0x03=0 → left)
	// Actually: button=20 → bits: 16(Ctrl) + 4(Shift) + 0(left) = 20
	// 20 & 0x03 = 0 → left, 20 & 4 = 4 → Shift, 20 & 16 = 16 → Ctrl
	data := []byte{0x1b, '[', '<', '2', '0', ';', '1', '0', ';', '5', 'M'}
	e := mustParseOne(t, data)
	if e.Type != "mousedown" {
		t.Errorf("expected mousedown, got %q", e.Type)
	}
	if e.Button != "left" {
		t.Errorf("expected left, got %q", e.Button)
	}
	if !e.Modifiers.Ctrl {
		t.Error("expected Ctrl modifier")
	}
	if !e.Modifiers.Shift {
		t.Error("expected Shift modifier")
	}
}

func TestParse_Insert(t *testing.T) {
	e := mustParseOne(t, []byte{0x1b, '[', '2', '~'})
	assertKey(t, e, "Insert")
}

func TestParse_ModifiedHomeEnd(t *testing.T) {
	// ESC[1;5H = Ctrl+Home
	e := mustParseOne(t, []byte{0x1b, '[', '1', ';', '5', 'H'})
	assertKey(t, e, "Home")
	assertMods(t, e, true, false, false)

	// ESC[1;2F = Shift+End
	e = mustParseOne(t, []byte{0x1b, '[', '1', ';', '2', 'F'})
	assertKey(t, e, "End")
	assertMods(t, e, false, false, true)
}

func TestParse_MultipleUTF8Chars(t *testing.T) {
	// "ab" in one read → should produce 2 events
	events := parseInput([]byte{'a', 'b'})
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	assertKey(t, events[0], "a")
	assertKey(t, events[1], "b")
}

func TestParse_EmptyInput(t *testing.T) {
	events := parseInput([]byte{})
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}

	events = parseInput(nil)
	if len(events) != 0 {
		t.Fatalf("expected 0 events for nil, got %d", len(events))
	}
}
