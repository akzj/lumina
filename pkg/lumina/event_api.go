package lumina

import (
	"strings"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// pushEventToLua pushes an Event as a Lua table onto the stack.
func pushEventToLua(L *lua.State, e *Event) {
	L.PushAny(map[string]any{
		"type": e.Type, "key": e.Key, "code": e.Code,
		"x": int64(e.X), "y": int64(e.Y), "button": e.Button,
		"target": e.Target, "timestamp": e.Timestamp,
		"modifiers": map[string]any{
			"ctrl": e.Modifiers.Ctrl, "shift": e.Modifiers.Shift,
			"alt": e.Modifiers.Alt, "meta": e.Modifiers.Meta,
		},
	})
}

// registerEvent(eventType, componentID?, handler) — register event handler
//
// In the new architecture, event handlers do NOT call L.PCall directly.
// Instead they post a LuaCallbackEvent to the App's event channel,
// which is dispatched on the main thread.
func registerEvent(L *lua.State) int {
	eventType := L.CheckString(1)
	compID := ""
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		if s, _ := L.ToString(2); s != "" {
			compID = s
		}
	}
	if L.Type(3) != lua.TypeFunction {
		L.PushString("registerEvent: third argument must be handler function")
		L.Error()
		return 0
	}

	// Store handler reference in registry
	refID := L.Ref(lua.RegistryIndex)

	// Get the App from UserValue for safe event posting
	app := GetApp(L)

	globalEventBus.On(eventType, compID, func(e *Event) {
		if app != nil {
			// Post to main thread — safe
			app.PostEvent(AppEvent{
				Type: "lua_callback",
				Payload: LuaCallbackEvent{RefID: refID, Event: e},
			})
		} else {
			// Fallback: direct call (for tests without App)
			pushEventToLua(L, e)
			_ = L.CallRef(refID, 1, 0)
		}
	})

	return 0
}

// unregisterEvent(eventType, componentID?) — unregister event handler
func unregisterEvent(L *lua.State) int {
	eventType := L.CheckString(1)
	compID := ""
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		if s, _ := L.ToString(2); s != "" {
			compID = s
		}
	}
	globalEventBus.Off(eventType, compID)
	return 0
}

// emitEvent(targetID, eventType, eventData?) — emit an event
func emitEvent(L *lua.State) int {
	targetID := L.CheckString(1)
	eventType := L.CheckString(2)

	e := &Event{
		Type:   eventType,
		Target: targetID,
	}

	// Optional event data table
	if L.GetTop() >= 3 && L.Type(3) == lua.TypeTable {
		if s := L.GetFieldString(3, "key"); s != "" {
			e.Key = s
		}
		e.X = int(L.GetFieldInt(3, "x"))
		e.Y = int(L.GetFieldInt(3, "y"))
	}

	globalEventBus.Emit(e)
	return 0
}

// registerShortcut(config) — register a keyboard shortcut
func registerShortcut(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("registerShortcut: expected table")
		L.Error()
		return 0
	}

	key := L.GetFieldString(1, "key")

	if key == "" {
		L.PushString("registerShortcut: 'key' is required")
		L.Error()
		return 0
	}

	if L.Type(2) != lua.TypeFunction {
		L.PushString("registerShortcut: second argument must be handler function")
		L.Error()
		return 0
	}

	// Store handler reference
	refID := L.Ref(lua.RegistryIndex)

	// Normalize key (lowercase, remove spaces)
	normKey := normalizeShortcutKey(key)

	// Get the App for safe event posting
	app := GetApp(L)

	globalEventBus.shortcuts[normKey] = eventHandler{
		componentID: "global",
		handler: func(e *Event) {
			if app != nil {
				app.PostEvent(AppEvent{
					Type:    "lua_callback",
					Payload: LuaCallbackEvent{RefID: refID, Event: e},
				})
			} else {
				L.RawGetI(lua.RegistryIndex, int64(refID))
				if L.Type(-1) == lua.TypeFunction {
					pushEventToLua(L, e)
					status := L.PCall(1, 0, 0)
					if status != lua.OK {
						L.Pop(1)
					}
				}
			}
		},
	}

	return 0
}

// setFocus(componentID) — set focus to a component
func setFocus(L *lua.State) int {
	compID := L.CheckString(1)
	globalEventBus.SetFocus(compID)
	return 0
}

// getFocused() → componentID — get currently focused component
func getFocused(L *lua.State) int {
	L.PushString(globalEventBus.GetFocused())
	return 1
}

// blur() — remove focus from current component
func blur(L *lua.State) int {
	globalEventBus.Blur()
	return 0
}

// focusNext() — move focus to next focusable component
func focusNext(L *lua.State) int {
	globalEventBus.FocusNext()
	return 0
}

// focusPrev() — move focus to previous focusable component
func focusPrev(L *lua.State) int {
	globalEventBus.FocusPrev()
	return 0
}

// registerFocusable(componentID) — register a component as focusable
func registerFocusable(L *lua.State) int {
	compID := L.CheckString(1)
	globalEventBus.RegisterFocusable(compID)
	return 0
}

// unregisterFocusable(componentID) — unregister a component from focusable list
func unregisterFocusable(L *lua.State) int {
	compID := L.CheckString(1)
	globalEventBus.UnregisterFocusable(compID)
	return 0
}

// isFocusable(componentID) → boolean — check if component is focusable
func isFocusable(L *lua.State) int {
	compID := L.CheckString(1)
	L.PushBoolean(globalEventBus.IsFocusable(compID))
	return 1
}

// getFocusableIDs() → table — get list of focusable component IDs
func getFocusableIDs(L *lua.State) int {
	ids := globalEventBus.GetFocusableIDs()
	L.PushAny(ids)
	return 1
}

// emitKeyEvent(key, modifiers?) — emit a key event for testing
func emitKeyEvent(L *lua.State) int {
	key := L.CheckString(1)

	e := &Event{
		Type:      "keydown",
		Key:       strings.ToLower(key),
		Timestamp: time.Now().UnixMilli(),
	}

	// Optional modifiers table
	if L.GetTop() >= 2 && L.Type(2) == lua.TypeTable {
		e.Modifiers.Ctrl = L.GetFieldBool(2, "ctrl")
		e.Modifiers.Shift = L.GetFieldBool(2, "shift")
		e.Modifiers.Alt = L.GetFieldBool(2, "alt")
	}

	globalEventBus.Emit(e)
	return 0
}

// registerCaptureEvent registers an event handler for the capture phase.
// lumina.onCapture(eventType, componentID, handler)
func registerCaptureEvent(L *lua.State) int {
	eventType := L.CheckString(1)
	compID := ""
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		if s, _ := L.ToString(2); s != "" {
			compID = s
		}
	}
	if L.Type(3) != lua.TypeFunction {
		L.PushString("registerCaptureEvent: third argument must be handler function")
		L.Error()
		return 0
	}

	refID := L.Ref(lua.RegistryIndex)
	app := GetApp(L)

	globalEventBus.OnCapture(eventType, compID, func(e *Event) {
		if app != nil {
			app.PostEvent(AppEvent{
				Type:    "lua_callback",
				Payload: LuaCallbackEvent{RefID: refID, Event: e},
			})
		} else {
			pushEventToLua(L, e)
			_ = L.CallRef(refID, 1, 0)
		}
	})

	return 0
}
