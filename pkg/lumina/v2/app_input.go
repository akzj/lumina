package v2

import (
	"unicode/utf8"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// handleInputKeyDown intercepts keydown events for focused input VNodes.
// If the focused VNode is type="input", it handles text editing (insert,
// backspace, delete, cursor movement) and calls onChange/onSubmit Lua
// callbacks. Returns true if the event was handled.
func (a *App) handleInputKeyDown(e *event.Event) bool {
	focusedID := a.dispatcher.FocusedID()
	if focusedID == "" {
		return false
	}

	// Find the focused VNode across all components.
	vnode, comp := a.findVNode(focusedID)
	if vnode == nil || (vnode.Type != "input" && vnode.Type != "textarea") {
		return false
	}
	isTextarea := vnode.Type == "textarea"

	// Tab and Shift+Tab should still cycle focus, not be captured by input.
	if e.Key == "Tab" || e.Key == "Shift+Tab" {
		return false
	}

	value := vnode.Content
	runes := []rune(value)
	cursorPos := len(runes) // default: end of text

	if cp, ok := vnode.Props["cursorPos"]; ok {
		cursorPos = toIntProp(cp)
	}

	// Clamp cursor to valid range.
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	key := e.Key
	changed := false

	switch {
	case key == "Backspace":
		if cursorPos > 0 {
			runes = append(runes[:cursorPos-1], runes[cursorPos:]...)
			cursorPos--
			changed = true
		}
	case key == "Delete":
		if cursorPos < len(runes) {
			runes = append(runes[:cursorPos], runes[cursorPos+1:]...)
			changed = true
		}
	case key == "ArrowLeft":
		if cursorPos > 0 {
			cursorPos--
		}
	case key == "ArrowRight":
		if cursorPos < len(runes) {
			cursorPos++
		}
	case key == "ArrowUp" && isTextarea:
		cursorPos = textareaCursorUp(runes, cursorPos)
	case key == "ArrowDown" && isTextarea:
		cursorPos = textareaCursorDown(runes, cursorPos)
	case key == "Home":
		if isTextarea {
			cursorPos = textareaLineStart(runes, cursorPos)
		} else {
			cursorPos = 0
		}
	case key == "End":
		if isTextarea {
			cursorPos = textareaLineEnd(runes, cursorPos)
		} else {
			cursorPos = len(runes)
		}
	case key == "Enter":
		if isTextarea {
			// Insert newline character.
			newRunes := make([]rune, len(runes)+1)
			copy(newRunes, runes[:cursorPos])
			newRunes[cursorPos] = '\n'
			copy(newRunes[cursorPos+1:], runes[cursorPos:])
			runes = newRunes
			cursorPos++
			changed = true
		} else {
			a.callInputLuaCallback(vnode, "onSubmit", value)
			return true
		}
	case key == "Escape":
		if isTextarea {
			a.callInputLuaCallback(vnode, "onSubmit", string(runes))
		}
		// Blur on Escape.
		a.dispatcher.SetFocus("")
		return true
	case isPrintableKey(key):
		// Insert the character at cursor position.
		ch := []rune(key)[0]
		if key == "Space" {
			ch = ' '
		}
		newRunes := make([]rune, len(runes)+1)
		copy(newRunes, runes[:cursorPos])
		newRunes[cursorPos] = ch
		copy(newRunes[cursorPos+1:], runes[cursorPos:])
		runes = newRunes
		cursorPos++
		changed = true
	default:
		// Unknown key — let the dispatcher handle it normally.
		return false
	}

	// Update the component state to trigger re-render.
	newValue := string(runes)

	// Store cursor position in component state so it persists across renders.
	comp.SetState("__inputCursor_"+focusedID, int64(cursorPos))

	if changed {
		comp.SetState("__inputValue_"+focusedID, newValue)
		a.callInputLuaCallback(vnode, "onChange", newValue)
	} else {
		// Cursor moved without text change — still need re-render for cursor.
		comp.MarkDirty()
	}

	return true
}

// callInputLuaCallback calls a Lua callback (onChange or onSubmit) stored as a
// registry ref in the VNode's Props.
func (a *App) callInputLuaCallback(vnode *layout.VNode, propName string, value string) {
	if a.bridge == nil {
		return
	}
	handler, ok := vnode.Props[propName]
	if !ok {
		return
	}
	ref, ok := inputToLuaRef(handler)
	if !ok {
		return
	}
	L := a.bridge.L
	L.RawGetI(lua.RegistryIndex, int64(ref))
	if !L.IsFunction(-1) {
		L.Pop(1)
		return
	}
	L.PushString(value)
	if status := L.PCall(1, 0, 0); status != lua.OK {
		L.Pop(1) // pop error
	}
}

// findVNode searches all component VNode trees for a VNode with the given ID.
// Returns the VNode and its owning component, or (nil, nil) if not found.
func (a *App) findVNode(vnodeID string) (*layout.VNode, *component.Component) {
	for _, comp := range a.manager.GetAll() {
		if comp.VNodeTree() != nil {
			if vn := findVNodeByID(comp.VNodeTree(), vnodeID); vn != nil {
				return vn, comp
			}
		}
	}
	return nil, nil
}

// findVNodeByID recursively searches a VNode tree for a node with the given ID.
func findVNodeByID(vn *layout.VNode, id string) *layout.VNode {
	if vn == nil {
		return nil
	}
	if vn.ID == id {
		return vn
	}
	for _, child := range vn.Children {
		if found := findVNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// isPrintableKey returns true if the key represents a single printable character.
func isPrintableKey(key string) bool {
	if key == "" {
		return false
	}
	// Space key (often sent as "Space" string).
	if key == "Space" || key == " " {
		return true
	}
	// Single character keys (a-z, 0-9, symbols, unicode).
	r, size := utf8.DecodeRuneInString(key)
	if size == len(key) && r != utf8.RuneError {
		// Exclude control characters.
		if r >= 0x20 && r != 0x7F {
			return true
		}
	}
	return false
}

// toIntProp converts a prop value to int.
func toIntProp(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	}
	return 0
}

// injectInputProps walks the VNode tree of a component and sets "focused"
// and "cursorPos" props on input VNodes based on the dispatcher's focus state
// and the component's state. Called as a PostRenderHook after renderFn+layout
// but before paint.
func (a *App) injectInputProps(comp *component.Component) {
	focusedID := a.dispatcher.FocusedID()
	if comp.VNodeTree() == nil {
		return
	}
	injectInputPropsWalk(comp.VNodeTree(), focusedID, comp)
}

// injectInputPropsWalk recursively walks the VNode tree and injects
// focused/cursorPos props on input VNodes.
func injectInputPropsWalk(vn *layout.VNode, focusedID string, comp *component.Component) {
	if vn == nil {
		return
	}
	if (vn.Type == "input" || vn.Type == "textarea") && vn.ID != "" {
		isFocused := vn.ID == focusedID
		vn.Props["focused"] = isFocused
		if isFocused {
			// Read cursor position from component state.
			if cp, ok := comp.State()["__inputCursor_"+vn.ID]; ok {
				vn.Props["cursorPos"] = cp
			}
			// Read value override from component state (set by handleInputKeyDown).
			if val, ok := comp.State()["__inputValue_"+vn.ID]; ok {
				if s, ok := val.(string); ok {
					vn.Content = s
				}
			}
		}
	}
	for _, child := range vn.Children {
		injectInputPropsWalk(child, focusedID, comp)
	}
}

// inputToLuaRef extracts a Lua registry reference from a handler value.
func inputToLuaRef(val any) (int, bool) {
	switch v := val.(type) {
	case int:
		if v > 0 {
			return v, true
		}
	case int64:
		if v > 0 {
			return int(v), true
		}
	}
	return 0, false
}

// --- Textarea cursor navigation helpers ---

// textareaLineStart returns the rune offset of the start of the current line.
func textareaLineStart(runes []rune, cursorPos int) int {
	for i := cursorPos - 1; i >= 0; i-- {
		if runes[i] == '\n' {
			return i + 1
		}
	}
	return 0
}

// textareaLineEnd returns the rune offset of the end of the current line
// (position just before the next '\n' or at end of text).
func textareaLineEnd(runes []rune, cursorPos int) int {
	for i := cursorPos; i < len(runes); i++ {
		if runes[i] == '\n' {
			return i
		}
	}
	return len(runes)
}

// textareaCursorUp moves the cursor up one line, preserving column position.
func textareaCursorUp(runes []rune, cursorPos int) int {
	// Find current line start and column
	lineStart := textareaLineStart(runes, cursorPos)
	col := cursorPos - lineStart

	if lineStart == 0 {
		// Already on first line — go to start
		return 0
	}

	// Previous line ends at lineStart-1 (the '\n')
	prevLineEnd := lineStart - 1
	prevLineStart := 0
	for i := prevLineEnd - 1; i >= 0; i-- {
		if runes[i] == '\n' {
			prevLineStart = i + 1
			break
		}
	}

	prevLineLen := prevLineEnd - prevLineStart
	if col > prevLineLen {
		col = prevLineLen
	}
	return prevLineStart + col
}

// textareaCursorDown moves the cursor down one line, preserving column position.
func textareaCursorDown(runes []rune, cursorPos int) int {
	// Find current line start and column
	lineStart := textareaLineStart(runes, cursorPos)
	col := cursorPos - lineStart

	// Find end of current line
	lineEnd := textareaLineEnd(runes, cursorPos)

	if lineEnd >= len(runes) {
		// Already on last line — go to end
		return len(runes)
	}

	// Next line starts at lineEnd+1 (after the '\n')
	nextLineStart := lineEnd + 1
	nextLineEnd := len(runes)
	for i := nextLineStart; i < len(runes); i++ {
		if runes[i] == '\n' {
			nextLineEnd = i
			break
		}
	}

	nextLineLen := nextLineEnd - nextLineStart
	if col > nextLineLen {
		col = nextLineLen
	}
	return nextLineStart + col
}
