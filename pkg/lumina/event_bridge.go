package lumina

import (
	"fmt"
	"os"

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

	// Release Lua refs from previous render cycle NOW — after clearing old handlers
	// but before registering new ones. This ensures refs are only freed when we're
	// about to replace them with new refs from the current render's VNode tree.
	SwapRenderRefs(app.L)

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

		// Register onMouseDown handler
		if ref, ok := getLuaRef(vnode.Props["onMouseDown"]); ok {
			app.registerBridgedLuaHandler("mousedown", id, ref)
		}

		// Register onMouseUp handler
		if ref, ok := getLuaRef(vnode.Props["onMouseUp"]); ok {
			app.registerBridgedLuaHandler("mouseup", id, ref)
		}

		// Register onMouseMove handler
		if ref, ok := getLuaRef(vnode.Props["onMouseMove"]); ok {
			app.registerBridgedLuaHandler("mousemove", id, ref)
		}

		// Register onMouseEnter handler
		if ref, ok := getLuaRef(vnode.Props["onMouseEnter"]); ok {
			app.registerBridgedLuaHandler("mouseenter", id, ref)
		}

		// Register onMouseLeave handler
		if ref, ok := getLuaRef(vnode.Props["onMouseLeave"]); ok {
			app.registerBridgedLuaHandler("mouseleave", id, ref)
		}

		// Register onDragOver handler
		if ref, ok := getLuaRef(vnode.Props["onDragOver"]); ok {
			app.registerBridgedLuaHandler("dragover", id, ref)
		}

		// Register onDrop handler
		if ref, ok := getLuaRef(vnode.Props["onDrop"]); ok {
			app.registerBridgedLuaHandler("drop", id, ref)
		}

		// Register onWheel handler
		if ref, ok := getLuaRef(vnode.Props["onWheel"]); ok {
			app.registerBridgedLuaHandler("wheel", id, ref)
		}

		// Register onInput handler
		if ref, ok := getLuaRef(vnode.Props["onInput"]); ok {
			app.registerBridgedLuaHandler("input", id, ref)
		}

		// Register onResize handler
		if ref, ok := getLuaRef(vnode.Props["onResize"]); ok {
			app.registerBridgedLuaHandler("resize", id, ref)
		}

		// Register onContextMenu handler
		if ref, ok := getLuaRef(vnode.Props["onContextMenu"]); ok {
			app.registerBridgedLuaHandler("contextmenu", id, ref)
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
// function by registry reference. When called from the main thread (inside a
// batch — e.g. during handleEvent → Emit), the Lua function is invoked
// synchronously so that state changes are visible before EndBatch renders.
// When called from another goroutine, it falls back to PostEvent (async).
func (app *App) registerBridgedLuaHandler(eventType, compID string, luaRef int) {
	globalEventBus.RegisterBridgedHandler(eventType, compID, func(e *Event) {
		if app.IsBatching() {
			// Already on main thread — call Lua directly (synchronous).
			// This ensures e.g. mouseleave → setHovered(false) takes effect
			// before the current batch's EndBatch triggers a render.
			app.invokeLuaCallback(luaRef, e)
		} else {
			// From another goroutine — post async for main-thread execution.
			app.PostEvent(AppEvent{
				Type:    "lua_callback",
				Payload: LuaCallbackEvent{RefID: luaRef, Event: e},
			})
		}
	})
}

// invokeLuaCallback calls a Lua function by registry reference on the main thread.
// Must only be called from the main goroutine (inside handleEvent / eventLoop).
func (app *App) invokeLuaCallback(luaRef int, e *Event) {
	if e.Type == "mouseenter" || e.Type == "mouseleave" {
		fmt.Fprintf(os.Stderr, "[EVENT] %s target=%s\n", e.Type, e.Target)
	}
	app.L.RawGetI(lua.RegistryIndex, int64(luaRef))
	if app.L.IsFunction(-1) {
		pushEventToLua(app.L, e)
		if status := app.L.PCall(1, 0, 0); status != lua.OK {
			app.L.Pop(1) // pop error
		}
	} else {
		app.L.Pop(1) // pop non-function
	}
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
		key == "onSubmit" || key == "onScroll" ||
		key == "onMouseDown" || key == "onMouseUp" || key == "onMouseMove" ||
		key == "onMouseEnter" || key == "onMouseLeave" ||
		key == "onWheel" || key == "onInput" || key == "onResize" || key == "onContextMenu" ||
		key == "onDragOver" || key == "onDrop"
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
	// DON'T call trackRenderRef(ref) — prop refs are managed per-component
	// to avoid SwapRenderRefs freeing them and causing registry slot reuse
	// that corrupts component renderFn refs.
	return LuaFuncRef{Ref: ref}
}

// Render ref tracking — release Lua registry refs from the previous render cycle.
var (
	previousRenderRefs []int
	currentRenderRefs  []int
)

func trackRenderRef(ref int) {
	currentRenderRefs = append(currentRenderRefs, ref)
}

// ResetRenderRefs clears all tracked render refs (for test isolation).
func ResetRenderRefs() {
	previousRenderRefs = nil
	currentRenderRefs = nil
}

// SwapRenderRefs releases all Lua registry refs from the previous render cycle
// and promotes the current cycle's refs to "previous". Call at the start of each render.
func SwapRenderRefs(L *lua.State) {
	for _, ref := range previousRenderRefs {
		L.Unref(lua.RegistryIndex, ref)
	}
	previousRenderRefs = currentRenderRefs
	currentRenderRefs = nil
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

	// Clear all focusable IDs (re-registered each render cycle)
	eb.focusableIDs = nil
	eb.focusableSet = nil
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
