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
	if vnode == nil || vnode.Type != "input" {
		return false
	}

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
	case key == "Home":
		cursorPos = 0
	case key == "End":
		cursorPos = len(runes)
	case key == "Enter":
		a.callInputLuaCallback(vnode, "onSubmit", value)
		return true
	case key == "Escape":
		// Blur the input on Escape.
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
