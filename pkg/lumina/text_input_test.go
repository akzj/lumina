package lumina

import (
	"testing"
)

func newTestInput() *TextInputState {
	return &TextInputState{
		SelectStart: -1,
		SelectEnd:   -1,
		Width:       20,
		Height:      1,
	}
}

func newTestTextArea() *TextInputState {
	return &TextInputState{
		SelectStart: -1,
		SelectEnd:   -1,
		Width:       20,
		Height:      5,
		MultiLine:   true,
	}
}

// --- Basic Text Insertion ---

func TestInsertChar(t *testing.T) {
	s := newTestInput()
	s.InsertChar('H')
	s.InsertChar('i')
	if s.Text != "Hi" {
		t.Errorf("InsertChar: text = %q, want %q", s.Text, "Hi")
	}
	if s.CursorPos != 2 {
		t.Errorf("InsertChar: cursor = %d, want 2", s.CursorPos)
	}
}

func TestInsertCharAtMiddle(t *testing.T) {
	s := newTestInput()
	s.Text = "Hllo"
	s.CursorPos = 1
	s.InsertChar('e')
	if s.Text != "Hello" {
		t.Errorf("InsertChar middle: text = %q, want %q", s.Text, "Hello")
	}
	if s.CursorPos != 2 {
		t.Errorf("InsertChar middle: cursor = %d, want 2", s.CursorPos)
	}
}

func TestInsertString(t *testing.T) {
	s := newTestInput()
	s.InsertString("Hello World")
	if s.Text != "Hello World" {
		t.Errorf("InsertString: text = %q, want %q", s.Text, "Hello World")
	}
	if s.CursorPos != 11 {
		t.Errorf("InsertString: cursor = %d, want 11", s.CursorPos)
	}
}

// --- Backspace / Delete ---

func TestBackspace(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.CursorPos = 5

	s.Backspace()
	if s.Text != "Hell" {
		t.Errorf("Backspace: text = %q, want %q", s.Text, "Hell")
	}
	if s.CursorPos != 4 {
		t.Errorf("Backspace: cursor = %d, want 4", s.CursorPos)
	}
}

func TestBackspaceAtStart(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.CursorPos = 0

	s.Backspace()
	if s.Text != "Hello" {
		t.Errorf("Backspace at start: text = %q, want %q", s.Text, "Hello")
	}
	if s.CursorPos != 0 {
		t.Errorf("Backspace at start: cursor = %d, want 0", s.CursorPos)
	}
}

func TestBackspaceMiddle(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.CursorPos = 3

	s.Backspace()
	if s.Text != "Helo" {
		t.Errorf("Backspace middle: text = %q, want %q", s.Text, "Helo")
	}
	if s.CursorPos != 2 {
		t.Errorf("Backspace middle: cursor = %d, want 2", s.CursorPos)
	}
}

func TestDelete(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.CursorPos = 0

	s.Delete()
	if s.Text != "ello" {
		t.Errorf("Delete: text = %q, want %q", s.Text, "ello")
	}
	if s.CursorPos != 0 {
		t.Errorf("Delete: cursor = %d, want 0", s.CursorPos)
	}
}

func TestDeleteAtEnd(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.CursorPos = 5

	s.Delete()
	if s.Text != "Hello" {
		t.Errorf("Delete at end: text = %q, want %q", s.Text, "Hello")
	}
}

// --- Cursor Movement ---

func TestMoveLeft(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.CursorPos = 3

	s.MoveLeft()
	if s.CursorPos != 2 {
		t.Errorf("MoveLeft: cursor = %d, want 2", s.CursorPos)
	}
}

func TestMoveLeftAtStart(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.CursorPos = 0

	s.MoveLeft()
	if s.CursorPos != 0 {
		t.Errorf("MoveLeft at start: cursor = %d, want 0", s.CursorPos)
	}
}

func TestMoveRight(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.CursorPos = 2

	s.MoveRight()
	if s.CursorPos != 3 {
		t.Errorf("MoveRight: cursor = %d, want 3", s.CursorPos)
	}
}

func TestMoveRightAtEnd(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.CursorPos = 5

	s.MoveRight()
	if s.CursorPos != 5 {
		t.Errorf("MoveRight at end: cursor = %d, want 5", s.CursorPos)
	}
}

func TestMoveHome(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello World"
	s.CursorPos = 7

	s.MoveHome()
	if s.CursorPos != 0 {
		t.Errorf("MoveHome: cursor = %d, want 0", s.CursorPos)
	}
}

func TestMoveEnd(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello World"
	s.CursorPos = 3

	s.MoveEnd()
	if s.CursorPos != 11 {
		t.Errorf("MoveEnd: cursor = %d, want 11", s.CursorPos)
	}
}

// --- Horizontal Scroll ---

func TestHorizontalScroll(t *testing.T) {
	s := newTestInput()
	s.Width = 10

	// Type 15 characters
	for i := 0; i < 15; i++ {
		s.InsertChar(rune('a' + i%26))
		s.EnsureCursorVisible()
	}

	// Cursor at 15, width 10 → scroll should be at least 6
	if s.ScrollX < 6 {
		t.Errorf("HorizontalScroll: ScrollX = %d, want >= 6", s.ScrollX)
	}

	// Cursor should be visible: CursorPos - ScrollX < Width
	if s.CursorPos-s.ScrollX >= s.Width {
		t.Error("Cursor not visible after horizontal scroll")
	}
}

func TestHorizontalScrollLeft(t *testing.T) {
	s := newTestInput()
	s.Width = 10
	s.Text = "abcdefghijklmno"
	s.CursorPos = 15
	s.ScrollX = 6

	// Move cursor to start
	for s.CursorPos > 0 {
		s.MoveLeft()
		s.EnsureCursorVisible()
	}

	if s.ScrollX != 0 {
		t.Errorf("Scroll after moving to start: ScrollX = %d, want 0", s.ScrollX)
	}
}

// --- Multi-line: Enter inserts newline ---

func TestMultiLineEnter(t *testing.T) {
	s := newTestTextArea()
	s.Text = "Hello"
	s.CursorPos = 5

	consumed, changed := HandleTextInputKey(s, "Enter", EventModifiers{})
	if !consumed {
		t.Error("Enter should be consumed in multi-line")
	}
	if !changed {
		t.Error("Enter should change text in multi-line")
	}
	if s.Text != "Hello\n" {
		t.Errorf("Multi-line Enter: text = %q, want %q", s.Text, "Hello\n")
	}
}

func TestMultiLineEnterMiddle(t *testing.T) {
	s := newTestTextArea()
	s.Text = "HelloWorld"
	s.CursorPos = 5

	s.InsertChar('\n')
	if s.Text != "Hello\nWorld" {
		t.Errorf("Multi-line Enter middle: text = %q, want %q", s.Text, "Hello\nWorld")
	}
}

// --- Multi-line: Up/Down ---

func TestMoveUpDown(t *testing.T) {
	s := newTestTextArea()
	s.Text = "Line1\nLine2\nLine3"
	s.CursorPos = 9 // in "Line2", at position 3 (after "Lin")

	s.MoveDown()
	// Should be on Line3, col 3
	if s.CursorPos != 15 { // "Line1\nLine2\n" = 12, + 3 = 15
		t.Errorf("MoveDown: cursor = %d, want 15", s.CursorPos)
	}

	s.MoveUp()
	// Should be back on Line2, col 3
	if s.CursorPos != 9 {
		t.Errorf("MoveUp: cursor = %d, want 9", s.CursorPos)
	}
}

func TestMoveUpAtFirstLine(t *testing.T) {
	s := newTestTextArea()
	s.Text = "Line1\nLine2"
	s.CursorPos = 3

	s.MoveUp()
	// Already on first line — no change
	if s.CursorPos != 3 {
		t.Errorf("MoveUp at first line: cursor = %d, want 3", s.CursorPos)
	}
}

func TestMoveDownAtLastLine(t *testing.T) {
	s := newTestTextArea()
	s.Text = "Line1\nLine2"
	s.CursorPos = 9

	s.MoveDown()
	// Already on last line — no change
	if s.CursorPos != 9 {
		t.Errorf("MoveDown at last line: cursor = %d, want 9", s.CursorPos)
	}
}

func TestMoveDownShorterLine(t *testing.T) {
	s := newTestTextArea()
	s.Text = "LongLine\nHi"
	s.CursorPos = 8 // end of "LongLine"

	s.MoveDown()
	// "Hi" is only 2 chars, so cursor should clamp to col 2
	if s.CursorPos != 11 { // "LongLine\n" = 9, + 2 = 11
		t.Errorf("MoveDown shorter line: cursor = %d, want 11", s.CursorPos)
	}
}

// --- Single-line: Enter triggers onSubmit (not text change) ---

func TestSingleLineEnter(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.CursorPos = 5

	consumed, changed := HandleTextInputKey(s, "Enter", EventModifiers{})
	if !consumed {
		t.Error("Enter should be consumed in single-line")
	}
	if changed {
		t.Error("Enter should NOT change text in single-line (it's a submit)")
	}
	if s.Text != "Hello" {
		t.Errorf("Single-line Enter: text = %q, want %q (unchanged)", s.Text, "Hello")
	}
}

// --- Placeholder ---

func TestPlaceholderRendering(t *testing.T) {
	ClearTextInputs()
	defer ClearTextInputs()

	root := NewVNode("input")
	root.Props["id"] = "ph-test"
	root.Props["placeholder"] = "Enter name..."

	// Register the text input
	state := GetTextInput("ph-test")
	state.Placeholder = "Enter name..."
	state.Width = 30
	state.Height = 1

	frame := VNodeToFrame(root, 40, 3)

	// Placeholder should be rendered (first char 'E')
	if frame.Cells[0][0].Char != 'E' {
		t.Errorf("Placeholder char = '%c', want 'E'", frame.Cells[0][0].Char)
	}
}

// --- MaxLength ---

func TestMaxLength(t *testing.T) {
	s := newTestInput()
	s.MaxLength = 5

	for i := 0; i < 10; i++ {
		s.InsertChar(rune('a' + i))
	}

	if len([]rune(s.Text)) != 5 {
		t.Errorf("MaxLength: text length = %d, want 5", len([]rune(s.Text)))
	}
	if s.Text != "abcde" {
		t.Errorf("MaxLength: text = %q, want %q", s.Text, "abcde")
	}
}

func TestMaxLengthInsertString(t *testing.T) {
	s := newTestInput()
	s.MaxLength = 5

	s.InsertString("Hello World")
	if len([]rune(s.Text)) != 5 {
		t.Errorf("MaxLength InsertString: text length = %d, want 5", len([]rune(s.Text)))
	}
	if s.Text != "Hello" {
		t.Errorf("MaxLength InsertString: text = %q, want %q", s.Text, "Hello")
	}
}

// --- UTF-8 Character Handling ---

func TestUTF8Insert(t *testing.T) {
	s := newTestInput()
	s.InsertChar('日')
	s.InsertChar('本')
	s.InsertChar('語')

	if s.Text != "日本語" {
		t.Errorf("UTF-8 insert: text = %q, want %q", s.Text, "日本語")
	}
	if s.CursorPos != 3 {
		t.Errorf("UTF-8 cursor: pos = %d, want 3 (rune count)", s.CursorPos)
	}
}

func TestUTF8Backspace(t *testing.T) {
	s := newTestInput()
	s.Text = "日本語"
	s.CursorPos = 3

	s.Backspace()
	if s.Text != "日本" {
		t.Errorf("UTF-8 backspace: text = %q, want %q", s.Text, "日本")
	}
	if s.CursorPos != 2 {
		t.Errorf("UTF-8 backspace: cursor = %d, want 2", s.CursorPos)
	}
}

func TestUTF8MoveLeftRight(t *testing.T) {
	s := newTestInput()
	s.Text = "日本語"
	s.CursorPos = 3

	s.MoveLeft()
	if s.CursorPos != 2 {
		t.Errorf("UTF-8 MoveLeft: cursor = %d, want 2", s.CursorPos)
	}

	s.MoveRight()
	if s.CursorPos != 3 {
		t.Errorf("UTF-8 MoveRight: cursor = %d, want 3", s.CursorPos)
	}
}

func TestUTF8Delete(t *testing.T) {
	s := newTestInput()
	s.Text = "日本語"
	s.CursorPos = 0

	s.Delete()
	if s.Text != "本語" {
		t.Errorf("UTF-8 delete: text = %q, want %q", s.Text, "本語")
	}
}

// --- Select All (Ctrl+A) ---

func TestSelectAll(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello World"

	s.SelectAll()
	if s.SelectStart != 0 {
		t.Errorf("SelectAll start = %d, want 0", s.SelectStart)
	}
	if s.SelectEnd != 11 {
		t.Errorf("SelectAll end = %d, want 11", s.SelectEnd)
	}
	if !s.HasSelection() {
		t.Error("HasSelection should be true after SelectAll")
	}
}

func TestSelectAllViaKey(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"

	consumed, _ := HandleTextInputKey(s, "a", EventModifiers{Ctrl: true})
	if !consumed {
		t.Error("Ctrl+A should be consumed")
	}
	if !s.HasSelection() {
		t.Error("Ctrl+A should create selection")
	}
}

func TestDeleteSelection(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello World"
	s.SelectStart = 5
	s.SelectEnd = 11 // select " World"

	s.DeleteSelection()
	if s.Text != "Hello" {
		t.Errorf("DeleteSelection: text = %q, want %q", s.Text, "Hello")
	}
	if s.CursorPos != 5 {
		t.Errorf("DeleteSelection: cursor = %d, want 5", s.CursorPos)
	}
	if s.HasSelection() {
		t.Error("Selection should be cleared after delete")
	}
}

func TestSelectedText(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello World"
	s.SelectStart = 6
	s.SelectEnd = 11

	if got := s.SelectedText(); got != "World" {
		t.Errorf("SelectedText = %q, want %q", got, "World")
	}
}

func TestBackspaceWithSelection(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello World"
	s.CursorPos = 11
	s.SelectStart = 5
	s.SelectEnd = 11

	s.Backspace()
	if s.Text != "Hello" {
		t.Errorf("Backspace with selection: text = %q, want %q", s.Text, "Hello")
	}
}

func TestInsertReplacesSelection(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello World"
	s.CursorPos = 11
	s.SelectStart = 0
	s.SelectEnd = 11

	s.InsertChar('X')
	if s.Text != "X" {
		t.Errorf("Insert replaces selection: text = %q, want %q", s.Text, "X")
	}
}

// --- ReadOnly Mode ---

func TestReadOnly(t *testing.T) {
	s := newTestInput()
	s.Text = "ReadOnly"
	s.ReadOnly = true

	s.InsertChar('X')
	if s.Text != "ReadOnly" {
		t.Errorf("ReadOnly insert: text = %q, want %q", s.Text, "ReadOnly")
	}

	s.Backspace()
	if s.Text != "ReadOnly" {
		t.Errorf("ReadOnly backspace: text = %q, want %q", s.Text, "ReadOnly")
	}

	s.Delete()
	if s.Text != "ReadOnly" {
		t.Errorf("ReadOnly delete: text = %q, want %q", s.Text, "ReadOnly")
	}
}

func TestReadOnlyCursorMovement(t *testing.T) {
	s := newTestInput()
	s.Text = "ReadOnly"
	s.ReadOnly = true
	s.CursorPos = 4

	// Cursor movement should still work
	s.MoveLeft()
	if s.CursorPos != 3 {
		t.Errorf("ReadOnly MoveLeft: cursor = %d, want 3", s.CursorPos)
	}
}

// --- HandleTextInputKey ---

func TestHandleKeyRegularChar(t *testing.T) {
	s := newTestInput()

	consumed, changed := HandleTextInputKey(s, "a", EventModifiers{})
	if !consumed {
		t.Error("Regular char should be consumed")
	}
	if !changed {
		t.Error("Regular char should change text")
	}
	if s.Text != "a" {
		t.Errorf("HandleKey 'a': text = %q, want %q", s.Text, "a")
	}
}

func TestHandleKeyBackspace(t *testing.T) {
	s := newTestInput()
	s.Text = "ab"
	s.CursorPos = 2

	consumed, changed := HandleTextInputKey(s, "Backspace", EventModifiers{})
	if !consumed || !changed {
		t.Error("Backspace should be consumed and change text")
	}
	if s.Text != "a" {
		t.Errorf("HandleKey Backspace: text = %q, want %q", s.Text, "a")
	}
}

func TestHandleKeyTab(t *testing.T) {
	s := newTestInput()

	consumed, _ := HandleTextInputKey(s, "Tab", EventModifiers{})
	if consumed {
		t.Error("Tab should NOT be consumed (let it move focus)")
	}
}

func TestHandleKeyUnknown(t *testing.T) {
	s := newTestInput()

	consumed, _ := HandleTextInputKey(s, "F1", EventModifiers{})
	if consumed {
		t.Error("Unknown key should NOT be consumed")
	}
}

// --- TextInput Registry ---

func TestTextInputRegistry(t *testing.T) {
	ClearTextInputs()
	defer ClearTextInputs()

	state := GetTextInput("test-input")
	state.Text = "Hello"
	state.CursorPos = 5

	state2 := GetTextInput("test-input")
	if state2.Text != "Hello" {
		t.Errorf("Registry: text = %q, want %q", state2.Text, "Hello")
	}
}

func TestTextInputRegistryPersistence(t *testing.T) {
	ClearTextInputs()
	defer ClearTextInputs()

	state := GetTextInput("persist-test")
	state.Text = "Persisted"
	state.CursorPos = 9
	state.ScrollX = 5

	// Simulate re-render: get same state
	state2 := GetTextInput("persist-test")
	if state2.Text != "Persisted" || state2.ScrollX != 5 {
		t.Errorf("Persistence: text=%q scrollX=%d, want text=%q scrollX=5",
			state2.Text, state2.ScrollX, "Persisted")
	}
}

// --- Multi-line Home/End ---

func TestMultiLineHome(t *testing.T) {
	s := newTestTextArea()
	s.Text = "Line1\nLine2\nLine3"
	s.CursorPos = 9 // middle of Line2

	s.MoveHome()
	// Should go to start of Line2 (pos 6)
	if s.CursorPos != 6 {
		t.Errorf("MultiLine Home: cursor = %d, want 6", s.CursorPos)
	}
}

func TestMultiLineEnd(t *testing.T) {
	s := newTestTextArea()
	s.Text = "Line1\nLine2\nLine3"
	s.CursorPos = 6 // start of Line2

	s.MoveEnd()
	// Should go to end of Line2 (pos 11, before \n)
	if s.CursorPos != 11 {
		t.Errorf("MultiLine End: cursor = %d, want 11", s.CursorPos)
	}
}

// --- Lines() helper ---

func TestLines(t *testing.T) {
	s := newTestTextArea()
	s.Text = "Line1\nLine2\nLine3"

	lines := s.Lines()
	if len(lines) != 3 {
		t.Errorf("Lines count = %d, want 3", len(lines))
	}
	if lines[0] != "Line1" || lines[1] != "Line2" || lines[2] != "Line3" {
		t.Errorf("Lines = %v", lines)
	}
}

func TestLinesEmpty(t *testing.T) {
	s := newTestTextArea()
	s.Text = ""

	lines := s.Lines()
	if len(lines) != 1 || lines[0] != "" {
		t.Errorf("Lines empty = %v, want [\"\"]", lines)
	}
}

// --- Vertical Scroll (multi-line) ---

func TestVerticalScroll(t *testing.T) {
	s := newTestTextArea()
	s.Width = 20
	s.Height = 3

	// Create text with 10 lines
	s.Text = "L1\nL2\nL3\nL4\nL5\nL6\nL7\nL8\nL9\nL10"
	s.CursorPos = len([]rune(s.Text)) // end of text

	s.EnsureCursorVisible()

	// With 10 lines and height 3, scrollY should be at least 7
	if s.ScrollY < 7 {
		t.Errorf("VerticalScroll: ScrollY = %d, want >= 7", s.ScrollY)
	}
}

// --- Rendering Integration ---

func TestInputRendering(t *testing.T) {
	ClearTextInputs()
	defer ClearTextInputs()

	state := GetTextInput("render-test")
	state.Text = "Hello"
	state.Width = 20
	state.Height = 1

	root := NewVNode("input")
	root.Props["id"] = "render-test"
	root.Props["value"] = "Hello"

	frame := VNodeToFrame(root, 40, 3)

	// "Hello" should be rendered starting at (0,0)
	if frame.Cells[0][0].Char != 'H' {
		t.Errorf("Input render [0][0] = '%c', want 'H'", frame.Cells[0][0].Char)
	}
	if frame.Cells[0][4].Char != 'o' {
		t.Errorf("Input render [0][4] = '%c', want 'o'", frame.Cells[0][4].Char)
	}
}

func TestTextAreaRendering(t *testing.T) {
	ClearTextInputs()
	defer ClearTextInputs()

	state := GetTextInput("ta-render")
	state.Text = "Line1\nLine2"
	state.Width = 20
	state.Height = 5
	state.MultiLine = true

	root := NewVNode("textarea")
	root.Props["id"] = "ta-render"
	root.Props["value"] = "Line1\nLine2"

	frame := VNodeToFrame(root, 40, 10)

	if frame.Cells[0][0].Char != 'L' {
		t.Errorf("TextArea render [0][0] = '%c', want 'L'", frame.Cells[0][0].Char)
	}
}

// --- MoveLeft/Right with selection clears selection ---

func TestMoveLeftClearsSelection(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.SelectStart = 1
	s.SelectEnd = 4
	s.CursorPos = 4

	s.MoveLeft()
	if s.HasSelection() {
		t.Error("MoveLeft should clear selection")
	}
	// Should move to start of selection
	if s.CursorPos != 1 {
		t.Errorf("MoveLeft from selection: cursor = %d, want 1", s.CursorPos)
	}
}

func TestMoveRightClearsSelection(t *testing.T) {
	s := newTestInput()
	s.Text = "Hello"
	s.SelectStart = 1
	s.SelectEnd = 4
	s.CursorPos = 1

	s.MoveRight()
	if s.HasSelection() {
		t.Error("MoveRight should clear selection")
	}
	// Should move to end of selection
	if s.CursorPos != 4 {
		t.Errorf("MoveRight from selection: cursor = %d, want 4", s.CursorPos)
	}
}
