package lumina

import (
	"fmt"
	"strings"

	"github.com/akzj/go-lua/pkg/lua"
)

// bridgeVNodeEvents walks the VNode tree and registers event handlers + focusable
// elements with the global EventBus. This bridges the render output to the input
// system, connecting onClick/onChange/onFocus handlers to the event dispatcher.
//
// Called after each render cycle to keep handlers in sync with the current VNode tree.
func (app *App) bridgeVNodeEvents(root *VNode) {
	if root == nil {
		return
	}

	// Clear previous bridged handlers (stale from last render)
	globalEventBus.ClearBridgedHandlers()

	// Build VNode tree for event bubbling
	tree := BuildVNodeTree(root)
	globalEventBus.SetVNodeTree(tree)

	// Walk the tree and register handlers
	app.walkVNodeForEvents(root)
}

// walkVNodeForEvents recursively walks the VNode tree, registering event handlers
// and focusable elements found in VNode props.
func (app *App) walkVNodeForEvents(vnode *VNode) {
	if vnode == nil {
		return
	}

	id := ""
	if idVal, ok := vnode.Props["id"].(string); ok {
		id = idVal
	}

	// Auto-generate ID if element has event handlers but no explicit ID
	if id == "" {
		hasHandler := false
		for key := range vnode.Props {
			if isEventProp(key) {
				hasHandler = true
				break
			}
		}
		if hasHandler {
			id = fmt.Sprintf("auto_%p", vnode)
			vnode.Props["id"] = id
		}
	}

	if id != "" {
		// Register onClick handler
		if ref, ok := getLuaRef(vnode.Props["onClick"]); ok {
			app.registerBridgedLuaHandler("click", id, ref)
			globalEventBus.RegisterFocusable(id)
		}

		// Register onChange handler
		if ref, ok := getLuaRef(vnode.Props["onChange"]); ok {
			app.registerBridgedLuaHandler("change", id, ref)
		}

		// Register onFocus handler
		if ref, ok := getLuaRef(vnode.Props["onFocus"]); ok {
			app.registerBridgedLuaHandler("focus", id, ref)
		}

		// Register onBlur handler
		if ref, ok := getLuaRef(vnode.Props["onBlur"]); ok {
			app.registerBridgedLuaHandler("blur", id, ref)
		}

		// Register onKeyDown handler
		if ref, ok := getLuaRef(vnode.Props["onKeyDown"]); ok {
			app.registerBridgedLuaHandler("keydown", id, ref)
		}

		// Register onKeyUp handler
		if ref, ok := getLuaRef(vnode.Props["onKeyUp"]); ok {
			app.registerBridgedLuaHandler("keyup", id, ref)
		}

		// Register onSubmit handler
		if ref, ok := getLuaRef(vnode.Props["onSubmit"]); ok {
			app.registerBridgedLuaHandler("submit", id, ref)
		}

		// Register onScroll handler
		if ref, ok := getLuaRef(vnode.Props["onScroll"]); ok {
			app.registerBridgedLuaHandler("scroll", id, ref)
		}

		// If element type is inherently focusable, register it
		if isFocusableType(vnode.Type) {
			globalEventBus.RegisterFocusable(id)
		}
	}

	// Recurse into children
	for _, child := range vnode.Children {
		app.walkVNodeForEvents(child)
	}
}

// registerBridgedLuaHandler registers a bridged event handler that calls a Lua
// function by registry reference. The handler posts a LuaCallbackEvent to the
// App's event channel for safe main-thread execution.
func (app *App) registerBridgedLuaHandler(eventType, compID string, luaRef int) {
	globalEventBus.RegisterBridgedHandler(eventType, compID, func(e *Event) {
		app.PostEvent(AppEvent{
			Type:    "lua_callback",
			Payload: LuaCallbackEvent{RefID: luaRef, Event: e},
		})
	})
}

// getLuaRef extracts a Lua registry reference from a VNode prop value.
// Returns (ref, true) if the value is a valid Lua function reference.
func getLuaRef(val any) (int, bool) {
	switch v := val.(type) {
	case int:
		if v > 0 {
			return v, true
		}
	case int64:
		if v > 0 {
			return int(v), true
		}
	case LuaFuncRef:
		return v.Ref, true
	}
	return 0, false
}

// isEventProp returns true if the key is an event handler prop name.
func isEventProp(key string) bool {
	return key == "onClick" || key == "onChange" || key == "onFocus" ||
		key == "onBlur" || key == "onKeyDown" || key == "onKeyUp" ||
		key == "onSubmit" || key == "onScroll"
}

// isFocusableType returns true if the VNode type is inherently interactive/focusable.
func isFocusableType(t string) bool {
	return t == "button" || t == "input" || t == "textarea" ||
		t == "select" || t == "checkbox" || t == "radio"
}

// LuaFuncRef wraps a Lua registry reference to a function.
// Used to distinguish function references from plain integers in VNode props.
type LuaFuncRef struct {
	Ref int
}

// storeLuaFuncRef stores a Lua function at the given stack index as a registry
// reference and returns a LuaFuncRef. The function is popped from the stack.
func storeLuaFuncRef(L *lua.State, idx int) LuaFuncRef {
	L.PushValue(idx)
	ref := L.Ref(lua.RegistryIndex)
	return LuaFuncRef{Ref: ref}
}

// ClearBridgedHandlers removes all handlers registered by the VNode→EventBus bridge.
// Called at the start of each render cycle before re-registering.
func (eb *EventBus) ClearBridgedHandlers() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Remove bridged handlers from the main handlers map
	for _, bh := range eb.bridgedHandlers {
		handlers := eb.handlers[bh.eventType]
		filtered := make([]eventHandler, 0, len(handlers))
		for _, h := range handlers {
			if !h.bridged {
				filtered = append(filtered, h)
			}
		}
		eb.handlers[bh.eventType] = filtered
	}
	eb.bridgedHandlers = nil

	// Clear auto-registered focusable IDs (those with "auto_" prefix)
	filtered := make([]string, 0, len(eb.focusableIDs))
	for _, id := range eb.focusableIDs {
		if !strings.HasPrefix(id, "auto_") {
			filtered = append(filtered, id)
		}
	}
	eb.focusableIDs = filtered
}

// RegisterBridgedHandler registers a handler that will be cleared on next render cycle.
func (eb *EventBus) RegisterBridgedHandler(eventType, componentID string, handler func(*Event)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	h := eventHandler{
		componentID: componentID,
		handler:     handler,
		bridged:     true,
	}
	eb.handlers[eventType] = append(eb.handlers[eventType], h)
	eb.bridgedHandlers = append(eb.bridgedHandlers, bridgedHandler{eventType, componentID, handler})
}

// bridgedHandler tracks handlers registered by the VNode→EventBus bridge.
type bridgedHandler struct {
	eventType   string
	componentID string
	handler     func(*Event)
}
