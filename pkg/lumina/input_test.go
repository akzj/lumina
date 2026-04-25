package lumina

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testInputReader creates an InputReader with a buffered channel for testing.
func testInputReader() (*InputReader, chan AppEvent) {
	events := make(chan AppEvent, 64)
	ir := &InputReader{events: events}
	return ir, events
}

// drainEvent reads one event from the channel with a timeout.
func drainEvent(t *testing.T, ch chan AppEvent) *Event {
	t.Helper()
	select {
	case ev := <-ch:
		e, ok := ev.Payload.(*Event)
		require.True(t, ok, "payload should be *Event")
		return e
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for event")
		return nil
	}
}

// assertNoEvent asserts the channel is empty.
func assertNoEvent(t *testing.T, ch chan AppEvent) {
	t.Helper()
	select {
	case ev := <-ch:
		t.Fatalf("unexpected event: %+v", ev)
	default:
		// good — no event
	}
}

func TestParseInput_RegularKey(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte("a"))

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "a", e.Key)
	assert.False(t, e.Modifiers.Ctrl)
	assert.False(t, e.Modifiers.Alt)
	assert.False(t, e.Modifiers.Shift)
}

func TestParseInput_MultipleChars(t *testing.T) {
	ir, ch := testInputReader()
	// When multiple bytes arrive that aren't an escape sequence,
	// they get sent as a single key event (paste or multi-byte char)
	ir.parseInput([]byte("hello"))

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "hello", e.Key)
}

func TestParseInput_CtrlC(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x03})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "c", e.Key)
	assert.True(t, e.Modifiers.Ctrl)
}

func TestParseInput_CtrlA(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x01})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "a", e.Key)
	assert.True(t, e.Modifiers.Ctrl)
}

func TestParseInput_CtrlD(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x04})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "d", e.Key)
	assert.True(t, e.Modifiers.Ctrl)
}

func TestParseInput_Tab(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x09})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "Tab", e.Key)
	assert.False(t, e.Modifiers.Shift)
}

func TestParseInput_ShiftTab(t *testing.T) {
	ir, ch := testInputReader()
	// Shift+Tab = ESC[Z
	ir.parseInput([]byte{0x1b, '[', 'Z'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "Tab", e.Key)
	assert.True(t, e.Modifiers.Shift)
}

func TestParseInput_Enter(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x0d}) // CR

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "Enter", e.Key)
}

func TestParseInput_EnterLF(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x0a}) // LF

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "Enter", e.Key)
}

func TestParseInput_Backspace(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x7f})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "Backspace", e.Key)
}

func TestParseInput_Escape(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x1b}) // bare ESC

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "Escape", e.Key)
}

func TestParseInput_ArrowUp(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x1b, '[', 'A'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "ArrowUp", e.Key)
}

func TestParseInput_ArrowDown(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x1b, '[', 'B'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "ArrowDown", e.Key)
}

func TestParseInput_ArrowRight(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x1b, '[', 'C'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "ArrowRight", e.Key)
}

func TestParseInput_ArrowLeft(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x1b, '[', 'D'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "ArrowLeft", e.Key)
}

func TestParseInput_Home(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x1b, '[', 'H'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "Home", e.Key)
}

func TestParseInput_End(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x1b, '[', 'F'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "End", e.Key)
}

func TestParseInput_Delete(t *testing.T) {
	ir, ch := testInputReader()
	// ESC[3~
	ir.parseInput([]byte{0x1b, '[', '3', '~'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "Delete", e.Key)
}

func TestParseInput_Insert(t *testing.T) {
	ir, ch := testInputReader()
	// ESC[2~
	ir.parseInput([]byte{0x1b, '[', '2', '~'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "Insert", e.Key)
}

func TestParseInput_PageUp(t *testing.T) {
	ir, ch := testInputReader()
	// ESC[5~
	ir.parseInput([]byte{0x1b, '[', '5', '~'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "PageUp", e.Key)
}

func TestParseInput_PageDown(t *testing.T) {
	ir, ch := testInputReader()
	// ESC[6~
	ir.parseInput([]byte{0x1b, '[', '6', '~'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "PageDown", e.Key)
}

func TestParseInput_F1(t *testing.T) {
	ir, ch := testInputReader()
	// ESC OP
	ir.parseInput([]byte{0x1b, 'O', 'P'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "F1", e.Key)
}

func TestParseInput_F5(t *testing.T) {
	ir, ch := testInputReader()
	// ESC[15~
	ir.parseInput([]byte{0x1b, '[', '1', '5', '~'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "F5", e.Key)
}

func TestParseInput_F12(t *testing.T) {
	ir, ch := testInputReader()
	// ESC[24~
	ir.parseInput([]byte{0x1b, '[', '2', '4', '~'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "F12", e.Key)
}

func TestParseInput_AltKey(t *testing.T) {
	ir, ch := testInputReader()
	// Alt+x = ESC x
	ir.parseInput([]byte{0x1b, 'x'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "x", e.Key)
	assert.True(t, e.Modifiers.Alt)
}

func TestParseInput_CtrlArrowUp(t *testing.T) {
	ir, ch := testInputReader()
	// ESC[1;5A = Ctrl+Up
	ir.parseInput([]byte{0x1b, '[', '1', ';', '5', 'A'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "ArrowUp", e.Key)
	assert.True(t, e.Modifiers.Ctrl)
	assert.False(t, e.Modifiers.Shift)
}

func TestParseInput_ShiftArrowRight(t *testing.T) {
	ir, ch := testInputReader()
	// ESC[1;2C = Shift+Right
	ir.parseInput([]byte{0x1b, '[', '1', ';', '2', 'C'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "ArrowRight", e.Key)
	assert.True(t, e.Modifiers.Shift)
	assert.False(t, e.Modifiers.Ctrl)
}

func TestParseInput_CtrlShiftArrowDown(t *testing.T) {
	ir, ch := testInputReader()
	// ESC[1;6B = Ctrl+Shift+Down (modifier 6 = 1+Shift+Ctrl = 1+1+4 = 6)
	ir.parseInput([]byte{0x1b, '[', '1', ';', '6', 'B'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "ArrowDown", e.Key)
	assert.True(t, e.Modifiers.Ctrl)
	assert.True(t, e.Modifiers.Shift)
}

func TestParseInput_MousePress(t *testing.T) {
	ir, ch := testInputReader()
	// SGR mouse press: ESC[<0;10;20M (left button press at SGR 10,20 → 0-based 9,19)
	ir.parseInput([]byte{0x1b, '[', '<', '0', ';', '1', '0', ';', '2', '0', 'M'})

	e := drainEvent(t, ch)
	assert.Equal(t, "mousedown", e.Type)
	assert.Equal(t, "left", e.Button)
	assert.Equal(t, 9, e.X)  // SGR is 1-based, internal is 0-based
	assert.Equal(t, 19, e.Y) // SGR is 1-based, internal is 0-based
}

func TestParseInput_MouseRelease(t *testing.T) {
	ir, ch := testInputReader()
	// SGR mouse release: ESC[<0;10;20m (left button release at SGR 10,20 → 0-based 9,19)
	ir.parseInput([]byte{0x1b, '[', '<', '0', ';', '1', '0', ';', '2', '0', 'm'})

	e := drainEvent(t, ch)
	assert.Equal(t, "mouseup", e.Type)
	assert.Equal(t, "left", e.Button)
	assert.Equal(t, 9, e.X)
	assert.Equal(t, 19, e.Y)
}

func TestParseInput_MouseRightClick(t *testing.T) {
	ir, ch := testInputReader()
	// SGR right button press: ESC[<2;5;15M (SGR 5,15 → 0-based 4,14)
	ir.parseInput([]byte{0x1b, '[', '<', '2', ';', '5', ';', '1', '5', 'M'})

	e := drainEvent(t, ch)
	assert.Equal(t, "mousedown", e.Type)
	assert.Equal(t, "right", e.Button)
	assert.Equal(t, 4, e.X)
	assert.Equal(t, 14, e.Y)
}

func TestParseInput_ScrollUp(t *testing.T) {
	ir, ch := testInputReader()
	// SGR scroll up: ESC[<64;10;20M (button 64 = scroll up, SGR 10,20 → 0-based 9,19)
	ir.parseInput([]byte{0x1b, '[', '<', '6', '4', ';', '1', '0', ';', '2', '0', 'M'})

	e := drainEvent(t, ch)
	assert.Equal(t, "scroll", e.Type)
	assert.Equal(t, "up", e.Button)
	assert.Equal(t, 9, e.X)
	assert.Equal(t, 19, e.Y)
}

func TestParseInput_ScrollDown(t *testing.T) {
	ir, ch := testInputReader()
	// SGR scroll down: ESC[<65;10;20M (button 65 = scroll down, SGR 10,20 → 0-based 9,19)
	ir.parseInput([]byte{0x1b, '[', '<', '6', '5', ';', '1', '0', ';', '2', '0', 'M'})

	e := drainEvent(t, ch)
	assert.Equal(t, "scroll", e.Type)
	assert.Equal(t, "down", e.Button)
	assert.Equal(t, 9, e.X)
	assert.Equal(t, 19, e.Y)
}

func TestParseInput_CtrlMousePress(t *testing.T) {
	ir, ch := testInputReader()
	// SGR Ctrl+left click: ESC[<16;10;20M (button 0 + Ctrl flag 16 = 16)
	ir.parseInput([]byte{0x1b, '[', '<', '1', '6', ';', '1', '0', ';', '2', '0', 'M'})

	e := drainEvent(t, ch)
	assert.Equal(t, "mousedown", e.Type)
	assert.Equal(t, "left", e.Button)
	assert.True(t, e.Modifiers.Ctrl)
}

func TestParseInput_UTF8(t *testing.T) {
	ir, ch := testInputReader()
	// UTF-8 for '日' = 0xE6, 0x97, 0xA5
	ir.parseInput([]byte{0xE6, 0x97, 0xA5})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "日", e.Key)
}

func TestParseInput_EmptyInput(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{})
	assertNoEvent(t, ch)
}

func TestParseInput_CtrlQ(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x11}) // Ctrl+Q

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "q", e.Key)
	assert.True(t, e.Modifiers.Ctrl)
}

func TestParseInput_CtrlW(t *testing.T) {
	ir, ch := testInputReader()
	ir.parseInput([]byte{0x17}) // Ctrl+W

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "w", e.Key)
	assert.True(t, e.Modifiers.Ctrl)
}

func TestDecodeXtermModifier(t *testing.T) {
	tests := []struct {
		mod   int
		ctrl  bool
		shift bool
		alt   bool
	}{
		{1, false, false, false}, // no modifier
		{2, false, true, false},  // Shift
		{3, false, false, true},  // Alt
		{4, false, true, true},   // Shift+Alt
		{5, true, false, false},  // Ctrl
		{6, true, true, false},   // Ctrl+Shift
		{7, true, false, true},   // Ctrl+Alt
		{8, true, true, true},    // Ctrl+Shift+Alt
	}

	for _, tt := range tests {
		m := decodeXtermModifier(tt.mod)
		assert.Equal(t, tt.ctrl, m.Ctrl, "mod=%d ctrl", tt.mod)
		assert.Equal(t, tt.shift, m.Shift, "mod=%d shift", tt.mod)
		assert.Equal(t, tt.alt, m.Alt, "mod=%d alt", tt.mod)
	}
}

func TestParseInput_ModifiedHome(t *testing.T) {
	ir, ch := testInputReader()
	// ESC[1;5H = Ctrl+Home
	ir.parseInput([]byte{0x1b, '[', '1', ';', '5', 'H'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "Home", e.Key)
	assert.True(t, e.Modifiers.Ctrl)
}

func TestParseInput_ModifiedEnd(t *testing.T) {
	ir, ch := testInputReader()
	// ESC[1;2F = Shift+End
	ir.parseInput([]byte{0x1b, '[', '1', ';', '2', 'F'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "End", e.Key)
	assert.True(t, e.Modifiers.Shift)
}

func TestParseInput_SS3Home(t *testing.T) {
	ir, ch := testInputReader()
	// ESC OH = Home (SS3 variant)
	ir.parseInput([]byte{0x1b, 'O', 'H'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "Home", e.Key)
}

func TestParseInput_SS3End(t *testing.T) {
	ir, ch := testInputReader()
	// ESC OF = End (SS3 variant)
	ir.parseInput([]byte{0x1b, 'O', 'F'})

	e := drainEvent(t, ch)
	assert.Equal(t, "keydown", e.Type)
	assert.Equal(t, "End", e.Key)
}

func TestParseInput_MiddleMouseButton(t *testing.T) {
	ir, ch := testInputReader()
	// SGR middle button press: ESC[<1;5;15M
	ir.parseInput([]byte{0x1b, '[', '<', '1', ';', '5', ';', '1', '5', 'M'})

	e := drainEvent(t, ch)
	assert.Equal(t, "mousedown", e.Type)
	assert.Equal(t, "middle", e.Button)
}

func TestParseInput_EventTimestamp(t *testing.T) {
	ir, ch := testInputReader()
	before := time.Now().UnixMilli()
	ir.parseInput([]byte("x"))
	after := time.Now().UnixMilli()

	e := drainEvent(t, ch)
	assert.GreaterOrEqual(t, e.Timestamp, before)
	assert.LessOrEqual(t, e.Timestamp, after)
}
