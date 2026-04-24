package lumina

import (
	"strings"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// pushEventToLua pushes an Event as a Lua table onto the stack.
func pushEventToLua(L *lua.State, e *Event) {
	L.NewTable()
	L.PushString("type"); L.PushString(e.Type); L.SetTable(-3)
	L.PushString("key"); L.PushString(e.Key); L.SetTable(-3)
	L.PushString("code"); L.PushString(e.Code); L.SetTable(-3)
	L.PushString("x"); L.PushInteger(int64(e.X)); L.SetTable(-3)
	L.PushString("y"); L.PushInteger(int64(e.Y)); L.SetTable(-3)
	L.PushString("button"); L.PushString(e.Button); L.SetTable(-3)
	L.PushString("target"); L.PushString(e.Target); L.SetTable(-3)
	L.PushString("timestamp"); L.PushInteger(e.Timestamp); L.SetTable(-3)

	// Modifiers sub-table
	L.PushString("modifiers")
	L.NewTable()
	L.PushString("ctrl"); L.PushBoolean(e.Modifiers.Ctrl); L.SetTable(-3)
	L.PushString("shift"); L.PushBoolean(e.Modifiers.Shift); L.SetTable(-3)
	L.PushString("alt"); L.PushBoolean(e.Modifiers.Alt); L.SetTable(-3)
	L.PushString("meta"); L.PushBoolean(e.Modifiers.Meta); L.SetTable(-3)
	L.SetTable(-3)
}

// registerEvent(eventType, componentID?, handler) — register event handler
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

	globalEventBus.On(eventType, compID, func(e *Event) {
		L.RawGetI(lua.RegistryIndex, int64(refID))
		if L.Type(-1) == lua.TypeFunction {
			pushEventToLua(L, e)
			status := L.PCall(1, 0, 0)
			if status != lua.OK {
				// Error in handler, just log and continue
				L.Pop(1)
			}
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
		L.GetField(3, "key")
		if s, _ := L.ToString(-1); s != "" {
			e.Key = s
		}
		L.Pop(1)

		L.GetField(3, "x")
		if n, ok := L.ToInteger(-1); ok {
			e.X = int(n)
		}
		L.Pop(1)

		L.GetField(3, "y")
		if n, ok := L.ToInteger(-1); ok {
			e.Y = int(n)
		}
		L.Pop(1)
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

	L.GetField(1, "key")
	key, _ := L.ToString(-1)
	L.Pop(1)

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

	globalEventBus.shortcuts[normKey] = eventHandler{
		componentID: "global",
		handler: func(e *Event) {
			L.RawGetI(lua.RegistryIndex, int64(refID))
			if L.Type(-1) == lua.TypeFunction {
				pushEventToLua(L, e)
				status := L.PCall(1, 0, 0)
				if status != lua.OK {
					L.Pop(1)
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
		L.GetField(2, "ctrl")
		e.Modifiers.Ctrl = L.ToBoolean(-1)
		L.Pop(1)

		L.GetField(2, "shift")
		e.Modifiers.Shift = L.ToBoolean(-1)
		L.Pop(1)

		L.GetField(2, "alt")
		e.Modifiers.Alt = L.ToBoolean(-1)
		L.Pop(1)
	}

	globalEventBus.Emit(e)
	return 0
}
