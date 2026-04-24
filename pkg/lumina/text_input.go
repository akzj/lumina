package lumina

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/akzj/go-lua/pkg/lua"
)

// TextInputState tracks cursor, selection, and text content for input/textarea.
type TextInputState struct {
	Text        string
	CursorPos   int    // rune position in text (0-based)
	SelectStart int    // rune position, -1 if no selection
	SelectEnd   int    // rune position, -1 if no selection
	ScrollX     int    // horizontal scroll offset (columns) for single-line
	ScrollY     int    // vertical scroll offset (rows) for multi-line
	Focused     bool
	Placeholder string
	MaxLength   int  // 0 = unlimited
	MultiLine   bool
	ReadOnly    bool
	Width       int  // visible width in columns
	Height      int  // visible height in rows (1 for single-line)
}

// runeLen returns the number of runes in the text.
func (s *TextInputState) runeLen() int {
	return utf8.RuneCountInString(s.Text)
}

// runeSlice returns the text as a rune slice.
func (s *TextInputState) runeSlice() []rune {
	return []rune(s.Text)
}

// InsertChar inserts a character at the current cursor position.
func (s *TextInputState) InsertChar(ch rune) {
	if s.ReadOnly {
		return
	}
	if s.MaxLength > 0 && s.runeLen() >= s.MaxLength {
		return
	}

	// If there's a selection, delete it first
	if s.HasSelection() {
		s.DeleteSelection()
	}

	runes := s.runeSlice()
	pos := s.CursorPos
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}

	// Insert rune at position
	newRunes := make([]rune, 0, len(runes)+1)
	newRunes = append(newRunes, runes[:pos]...)
	newRunes = append(newRunes, ch)
	newRunes = append(newRunes, runes[pos:]...)

	s.Text = string(newRunes)
	s.CursorPos = pos + 1
	s.ClearSelection()
}

// InsertString inserts a string at the current cursor position.
func (s *TextInputState) InsertString(str string) {
	if s.ReadOnly {
		return
	}

	// If there's a selection, delete it first
	if s.HasSelection() {
		s.DeleteSelection()
	}

	insertRunes := []rune(str)
	if s.MaxLength > 0 {
		available := s.MaxLength - s.runeLen()
		if available <= 0 {
			return
		}
		if len(insertRunes) > available {
			insertRunes = insertRunes[:available]
		}
	}

	runes := s.runeSlice()
	pos := s.CursorPos
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}

	newRunes := make([]rune, 0, len(runes)+len(insertRunes))
	newRunes = append(newRunes, runes[:pos]...)
	newRunes = append(newRunes, insertRunes...)
	newRunes = append(newRunes, runes[pos:]...)

	s.Text = string(newRunes)
	s.CursorPos = pos + len(insertRunes)
	s.ClearSelection()
}

// Backspace deletes the character before the cursor.
func (s *TextInputState) Backspace() {
	if s.ReadOnly {
		return
	}

	if s.HasSelection() {
		s.DeleteSelection()
		return
	}

	if s.CursorPos <= 0 {
		return
	}

	runes := s.runeSlice()
	pos := s.CursorPos
	if pos > len(runes) {
		pos = len(runes)
	}

	newRunes := make([]rune, 0, len(runes)-1)
	newRunes = append(newRunes, runes[:pos-1]...)
	newRunes = append(newRunes, runes[pos:]...)

	s.Text = string(newRunes)
	s.CursorPos = pos - 1
}

// Delete deletes the character at the cursor.
func (s *TextInputState) Delete() {
	if s.ReadOnly {
		return
	}

	if s.HasSelection() {
		s.DeleteSelection()
		return
	}

	runes := s.runeSlice()
	pos := s.CursorPos
	if pos >= len(runes) {
		return
	}

	newRunes := make([]rune, 0, len(runes)-1)
	newRunes = append(newRunes, runes[:pos]...)
	newRunes = append(newRunes, runes[pos+1:]...)

	s.Text = string(newRunes)
}

// MoveLeft moves the cursor one rune to the left.
func (s *TextInputState) MoveLeft() {
	if s.HasSelection() {
		// Move to start of selection
		s.CursorPos = s.selMin()
		s.ClearSelection()
		return
	}
	if s.CursorPos > 0 {
		s.CursorPos--
	}
}

// MoveRight moves the cursor one rune to the right.
func (s *TextInputState) MoveRight() {
	if s.HasSelection() {
		// Move to end of selection
		s.CursorPos = s.selMax()
		s.ClearSelection()
		return
	}
	if s.CursorPos < s.runeLen() {
		s.CursorPos++
	}
}

// MoveHome moves the cursor to the start of the current line.
func (s *TextInputState) MoveHome() {
	s.ClearSelection()
	if !s.MultiLine {
		s.CursorPos = 0
		return
	}
	// Find start of current line
	runes := s.runeSlice()
	pos := s.CursorPos
	if pos > len(runes) {
		pos = len(runes)
	}
	for pos > 0 && runes[pos-1] != '\n' {
		pos--
	}
	s.CursorPos = pos
}

// MoveEnd moves the cursor to the end of the current line.
func (s *TextInputState) MoveEnd() {
	s.ClearSelection()
	if !s.MultiLine {
		s.CursorPos = s.runeLen()
		return
	}
	// Find end of current line
	runes := s.runeSlice()
	pos := s.CursorPos
	for pos < len(runes) && runes[pos] != '\n' {
		pos++
	}
	s.CursorPos = pos
}

// MoveUp moves the cursor up one line (multi-line only).
func (s *TextInputState) MoveUp() {
	if !s.MultiLine {
		return
	}
	s.ClearSelection()

	runes := s.runeSlice()
	line, col := s.lineCol(runes)
	if line == 0 {
		return // already on first line
	}

	// Find the start of the previous line
	prevLineStart, prevLineLen := s.lineInfo(runes, line-1)
	newCol := col
	if newCol > prevLineLen {
		newCol = prevLineLen
	}
	s.CursorPos = prevLineStart + newCol
}

// MoveDown moves the cursor down one line (multi-line only).
func (s *TextInputState) MoveDown() {
	if !s.MultiLine {
		return
	}
	s.ClearSelection()

	runes := s.runeSlice()
	line, col := s.lineCol(runes)
	lineCount := s.lineCount(runes)
	if line >= lineCount-1 {
		return // already on last line
	}

	// Find the start of the next line
	nextLineStart, nextLineLen := s.lineInfo(runes, line+1)
	newCol := col
	if newCol > nextLineLen {
		newCol = nextLineLen
	}
	s.CursorPos = nextLineStart + newCol
}

// SelectAll selects all text.
func (s *TextInputState) SelectAll() {
	s.SelectStart = 0
	s.SelectEnd = s.runeLen()
	s.CursorPos = s.runeLen()
}

// HasSelection returns true if there's an active selection.
func (s *TextInputState) HasSelection() bool {
	return s.SelectStart >= 0 && s.SelectEnd >= 0 && s.SelectStart != s.SelectEnd
}

// ClearSelection clears the selection.
func (s *TextInputState) ClearSelection() {
	s.SelectStart = -1
	s.SelectEnd = -1
}

// DeleteSelection deletes the selected text.
func (s *TextInputState) DeleteSelection() {
	if !s.HasSelection() || s.ReadOnly {
		return
	}

	lo := s.selMin()
	hi := s.selMax()
	runes := s.runeSlice()

	if lo < 0 {
		lo = 0
	}
	if hi > len(runes) {
		hi = len(runes)
	}

	newRunes := make([]rune, 0, len(runes)-(hi-lo))
	newRunes = append(newRunes, runes[:lo]...)
	newRunes = append(newRunes, runes[hi:]...)

	s.Text = string(newRunes)
	s.CursorPos = lo
	s.ClearSelection()
}

// SelectedText returns the currently selected text.
func (s *TextInputState) SelectedText() string {
	if !s.HasSelection() {
		return ""
	}
	lo := s.selMin()
	hi := s.selMax()
	runes := s.runeSlice()
	if lo < 0 {
		lo = 0
	}
	if hi > len(runes) {
		hi = len(runes)
	}
	return string(runes[lo:hi])
}

// EnsureCursorVisible adjusts scroll offsets so the cursor is visible.
func (s *TextInputState) EnsureCursorVisible() {
	if s.Width <= 0 {
		return
	}

	if s.MultiLine {
		runes := s.runeSlice()
		line, col := s.lineCol(runes)

		// Horizontal scroll
		if col < s.ScrollX {
			s.ScrollX = col
		} else if col >= s.ScrollX+s.Width {
			s.ScrollX = col - s.Width + 1
		}

		// Vertical scroll
		viewH := s.Height
		if viewH <= 0 {
			viewH = 1
		}
		if line < s.ScrollY {
			s.ScrollY = line
		} else if line >= s.ScrollY+viewH {
			s.ScrollY = line - viewH + 1
		}
	} else {
		// Single-line: horizontal scroll only
		col := s.CursorPos
		if col < s.ScrollX {
			s.ScrollX = col
		} else if col >= s.ScrollX+s.Width {
			s.ScrollX = col - s.Width + 1
		}
	}

	if s.ScrollX < 0 {
		s.ScrollX = 0
	}
	if s.ScrollY < 0 {
		s.ScrollY = 0
	}
}

// --- Internal helpers ---

func (s *TextInputState) selMin() int {
	if s.SelectStart < s.SelectEnd {
		return s.SelectStart
	}
	return s.SelectEnd
}

func (s *TextInputState) selMax() int {
	if s.SelectStart > s.SelectEnd {
		return s.SelectStart
	}
	return s.SelectEnd
}

// lineCol returns the (line, col) for the current cursor position.
// Both are 0-based.
func (s *TextInputState) lineCol(runes []rune) (int, int) {
	line := 0
	col := 0
	pos := s.CursorPos
	if pos > len(runes) {
		pos = len(runes)
	}
	for i := 0; i < pos; i++ {
		if runes[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return line, col
}

// lineCount returns the number of lines in the text.
func (s *TextInputState) lineCount(runes []rune) int {
	count := 1
	for _, r := range runes {
		if r == '\n' {
			count++
		}
	}
	return count
}

// lineInfo returns (startPos, length) for the given 0-based line number.
// length does not include the trailing newline.
func (s *TextInputState) lineInfo(runes []rune, targetLine int) (int, int) {
	line := 0
	lineStart := 0
	for i, r := range runes {
		if line == targetLine {
			if r == '\n' {
				return lineStart, i - lineStart
			}
		} else if r == '\n' {
			line++
			lineStart = i + 1
		}
	}
	// Last line (no trailing newline)
	if line == targetLine {
		return lineStart, len(runes) - lineStart
	}
	return len(runes), 0
}

// Lines returns the text split into lines.
func (s *TextInputState) Lines() []string {
	return strings.Split(s.Text, "\n")
}

// --- TextInput Registry ---

var (
	textInputRegistry = make(map[string]*TextInputState)
	textInputMu       sync.RWMutex
)

// GetTextInput returns the TextInputState for the given ID, creating one if needed.
func GetTextInput(id string) *TextInputState {
	textInputMu.RLock()
	ti, ok := textInputRegistry[id]
	textInputMu.RUnlock()
	if ok {
		return ti
	}

	textInputMu.Lock()
	defer textInputMu.Unlock()
	if ti, ok := textInputRegistry[id]; ok {
		return ti
	}
	ti = &TextInputState{
		SelectStart: -1,
		SelectEnd:   -1,
	}
	textInputRegistry[id] = ti
	return ti
}

// SetTextInput stores a TextInputState for the given ID.
func SetTextInput(id string, state *TextInputState) {
	textInputMu.Lock()
	defer textInputMu.Unlock()
	textInputRegistry[id] = state
}

// RemoveTextInput removes a TextInputState from the registry.
func RemoveTextInput(id string) {
	textInputMu.Lock()
	defer textInputMu.Unlock()
	delete(textInputRegistry, id)
}

// ClearTextInputs removes all text input states (for testing).
func ClearTextInputs() {
	textInputMu.Lock()
	defer textInputMu.Unlock()
	textInputRegistry = make(map[string]*TextInputState)
}

// HandleTextInputKey processes a key event for a text input.
// Returns true if the key was consumed, and whether the text changed.
func HandleTextInputKey(state *TextInputState, key string, mods EventModifiers) (consumed bool, changed bool) {
	if state == nil {
		return false, false
	}

	oldText := state.Text

	// Ctrl combinations
	if mods.Ctrl {
		switch key {
		case "a":
			state.SelectAll()
			return true, false
		case "c":
			// Copy — just return selected text (caller handles clipboard)
			return true, false
		case "v":
			// Paste — caller should call InsertString with clipboard content
			return true, false
		default:
			return false, false
		}
	}

	switch key {
	case "Backspace":
		state.Backspace()
	case "Delete":
		state.Delete()
	case "Left", "\x1b[D":
		state.MoveLeft()
	case "Right", "\x1b[C":
		state.MoveRight()
	case "Home":
		state.MoveHome()
	case "End":
		state.MoveEnd()
	case "Up", "\x1b[A":
		state.MoveUp()
	case "Down", "\x1b[B":
		state.MoveDown()
	case "Enter", "\n":
		if state.MultiLine {
			state.InsertChar('\n')
		} else {
			// Single-line: Enter = submit (caller handles)
			return true, false
		}
	case "Tab", "\t":
		// Don't consume Tab — let it move focus
		return false, false
	default:
		// Regular character input
		if len(key) == 1 {
			ch := rune(key[0])
			if ch >= 32 { // printable ASCII
				state.InsertChar(ch)
			} else {
				return false, false
			}
		} else {
			// Multi-byte character or unhandled key
			runes := []rune(key)
			if len(runes) == 1 && runes[0] >= 32 {
				state.InsertChar(runes[0])
			} else {
				return false, false
			}
		}
	}

	state.EnsureCursorVisible()
	return true, state.Text != oldText
}

// --- Lua API ---

// luaSetInputValue implements lumina.setInputValue(id, value)
func luaSetInputValue(L *lua.State) int {
	id := L.CheckString(1)
	value := L.CheckString(2)

	state := GetTextInput(id)
	state.Text = value
	// Move cursor to end of new text
	state.CursorPos = utf8.RuneCountInString(value)
	state.ClearSelection()
	state.EnsureCursorVisible()
	return 0
}

// luaGetInputValue implements lumina.getInputValue(id) -> string
func luaGetInputValue(L *lua.State) int {
	id := L.CheckString(1)

	textInputMu.RLock()
	state, ok := textInputRegistry[id]
	textInputMu.RUnlock()

	if !ok {
		L.PushString("")
		return 1
	}
	L.PushString(state.Text)
	return 1
}

// luaRegisterInput implements lumina.registerInput(id, opts)
// opts = { multiLine=bool, placeholder=string, maxLength=int, onChange=fn, onSubmit=fn }
func luaRegisterInput(L *lua.State) int {
	id := L.CheckString(1)

	state := GetTextInput(id)
	state.SelectStart = -1
	state.SelectEnd = -1

	if L.Type(2) == lua.TypeTable {
		// multiLine
		L.GetField(2, "multiLine")
		if L.Type(-1) == lua.TypeBoolean {
			state.MultiLine = L.ToBoolean(-1)
		}
		L.Pop(1)

		// placeholder
		L.GetField(2, "placeholder")
		if s, ok := L.ToString(-1); ok {
			state.Placeholder = s
		}
		L.Pop(1)

		// maxLength
		L.GetField(2, "maxLength")
		if v, ok := L.ToInteger(-1); ok {
			state.MaxLength = int(v)
		}
		L.Pop(1)

		// readOnly
		L.GetField(2, "readOnly")
		if L.Type(-1) == lua.TypeBoolean {
			state.ReadOnly = L.ToBoolean(-1)
		}
		L.Pop(1)

		// onChange callback
		L.GetField(2, "onChange")
		if L.Type(-1) == lua.TypeFunction {
			refID := L.Ref(lua.RegistryIndex)
			RegisterTextCallback(id, "onChange", refID)
		} else {
			L.Pop(1)
		}

		// onSubmit callback
		L.GetField(2, "onSubmit")
		if L.Type(-1) == lua.TypeFunction {
			refID := L.Ref(lua.RegistryIndex)
			RegisterTextCallback(id, "onSubmit", refID)
		} else {
			L.Pop(1)
		}
	}

	return 0
}

// luaFocusInput implements lumina.focusInput(id)
// Sets focus to the given input and marks it as focused.
func luaFocusInput(L *lua.State) int {
	id := L.CheckString(1)

	state := GetTextInput(id)
	state.Focused = true

	// Also set focus in the event bus
	globalEventBus.SetFocus(id)
	return 0
}
