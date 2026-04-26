package lumina

import (
	"fmt"
	"os"
	"sync"

	"github.com/akzj/go-lua/pkg/lua"
)

// normalizeKeyName maps terminal key names to user-friendly names used in onKey().
func normalizeKeyName(key string) string {
	switch key {
	case "Enter":
		return "enter"
	case "Escape":
		return "escape"
	case "Backspace":
		return "backspace"
	case "Tab":
		return "tab"
	case "ArrowLeft":
		return "left"
	case "ArrowRight":
		return "right"
	case "ArrowUp":
		return "up"
	case "ArrowDown":
		return "down"
	case "Delete":
		return "delete"
	case "Home":
		return "home"
	case "End":
		return "end"
	case "PageUp":
		return "pageup"
	case "PageDown":
		return "pagedown"
	case "Insert":
		return "insert"
	default:
		return key // regular characters like "h", "j", " ", etc. pass through
	}
}

// dispatchKeyBinding checks if a key has a lumina.onKey() binding and calls it directly.
// dispatchKeyBinding checks if a key has a lumina.onKey() binding and calls it directly.
func (app *App) dispatchKeyBinding(key string) {
	normalized := normalizeKeyName(key)
	keyBindingsMu.Lock()
	ref, ok := keyBindings[normalized]
	keyBindingsMu.Unlock()

	if ok {
		app.L.RawGetI(lua.RegistryIndex, int64(ref))
		if app.L.IsFunction(-1) {
			if status := app.L.PCall(0, 0, 0); status != lua.OK {
				msg, _ := app.L.ToString(-1)
				app.L.Pop(1)
				fmt.Fprintf(os.Stderr, "onKey(%q) error: %s\n", key, msg)
			}
		} else {
			app.L.Pop(1)
		}
	}
}

// handleScrollEvent handles scroll-related events (mouse wheel, PageUp/PageDown).
// handleTextInputEvent routes key events to the focused text input/textarea.
// Returns true if the event was consumed by a text input.
// handleScrollEvent handles scroll-related events (mouse wheel, PageUp/PageDown).
// handleTextInputEvent routes key events to the focused text input/textarea.
// Returns true if the event was consumed by a text input.
func (app *App) handleTextInputEvent(e *Event) bool {
	if e.Type != "keydown" {
		return false
	}

	focusedID := globalEventBus.GetFocused()
	if focusedID == "" {
		return false
	}

	// Check if the focused element has a text input state
	textInputMu.RLock()
	state, ok := textInputRegistry[focusedID]
	textInputMu.RUnlock()
	if !ok {
		return false
	}

	// Handle Enter for single-line input (triggers onSubmit, not consumed as text)
	if !state.MultiLine && (e.Key == "Enter" || e.Key == "\n") {
		// Trigger onSubmit callback if registered
		// The callback is stored as a Lua registry ref in the component
		app.triggerTextInputCallback(focusedID, "onSubmit", state.Text)
		return true
	}

	consumed, changed := HandleTextInputKey(state, e.Key, e.Modifiers)
	if !consumed {
		return false
	}

	if changed {
		// Trigger onChange callback
		app.triggerTextInputCallback(focusedID, "onChange", state.Text)
	}

	return consumed
}

// triggerTextInputCallback calls a Lua callback (onChange/onSubmit) for a text input.
// triggerTextInputCallback calls a Lua callback (onChange/onSubmit) for a text input.
func (app *App) triggerTextInputCallback(id, callbackName, value string) {
	// The callback refs are stored in the text input callback registry
	textCallbackMu.RLock()
	refID, ok := textCallbacks[id+":"+callbackName]
	textCallbackMu.RUnlock()
	if !ok || refID == 0 {
		return
	}

	app.L.RawGetI(lua.RegistryIndex, int64(refID))
	if app.L.Type(-1) == lua.TypeFunction {
		app.L.PushString(value)
		status := app.L.PCall(1, 0, 0)
		if status != lua.OK {
			app.L.Pop(1) // pop error
		}
	} else {
		app.L.Pop(1) // pop non-function
	}
}

// Text input callback registry — stores Lua function refs for onChange/onSubmit.
var (
	textCallbacks  = make(map[string]int) // "id:onChange" -> Lua registry ref
	textCallbackMu sync.RWMutex
)

// RegisterTextCallback registers a Lua callback for a text input event.
// RegisterTextCallback registers a Lua callback for a text input event.
func RegisterTextCallback(id, callbackName string, refID int) {
	textCallbackMu.Lock()
	defer textCallbackMu.Unlock()
	textCallbacks[id+":"+callbackName] = refID
}

// ClearTextCallbacks removes all text input callbacks (for testing).
// ClearTextCallbacks removes all text input callbacks (for testing).
func ClearTextCallbacks() {
	textCallbackMu.Lock()
	defer textCallbackMu.Unlock()
	textCallbacks = make(map[string]int)
}

func (app *App) handleScrollEvent(e *Event) {
	markAllDirty := markAllComponentsDirty

	switch e.Type {
	case "scroll":
		// Mouse wheel scroll — e.Button is "up" or "down", NOT e.Y (which is cursor position)
		scrollAmount := 3 // lines per wheel tick
		if e.Button == "up" {
			scrollAmount = -3
		}

		// Find the scrollable container under the cursor using VNode tree walk
		targetID := ""
		if root := app.findRootVNode(); root != nil {
			targetID, _ = findScrollableVNode(root, e.X, e.Y)
		}

		// Fallback: try focused element's viewport
		if targetID == "" {
			focusedID := globalEventBus.GetFocused()
			_, hasFocusedVP := viewportRegistry[focusedID]
			if hasFocusedVP {
				targetID = focusedID
			}
		}

		// Last resort: find any viewport that needs scroll
		if targetID == "" {
			for id, vp := range viewportRegistry {
				if vp.NeedsScroll() {
					targetID = id
					break
				}
			}
		}

		if targetID != "" {
			ScrollViewport(targetID, scrollAmount)
			markAllDirty()
		}

		// Also emit wheel event for onWheel handlers
		globalEventBus.Emit(&Event{
			Type:      "wheel",
			Target:    e.Target,
			Bubbles:   true,
			X:         e.X,
			Y:         e.Y,
			Button:    e.Button, // "up" or "down"
			Timestamp: e.Timestamp,
		})

	case "keydown":
		focusedID := globalEventBus.GetFocused()
		if focusedID == "" {
			return
		}

		scrolled := false
		switch e.Key {
		case "PageUp":
			vp, ok := viewportRegistry[focusedID]
			if ok {
				vp.ScrollUp(vp.ViewH)
				scrolled = true
			}
		case "PageDown":
			vp, ok := viewportRegistry[focusedID]
			if ok {
				vp.ScrollDown(vp.ViewH)
				scrolled = true
			}
		case "Home":
			vp, ok := viewportRegistry[focusedID]
			if ok {
				vp.ScrollToTop()
				scrolled = true
			}
		case "End":
			vp, ok := viewportRegistry[focusedID]
			if ok {
				vp.ScrollToBottom()
				scrolled = true
			}
		}
		if scrolled {
			markAllDirty()
		}
	}
}

// renderAllDirty checks all components for dirty state and re-renders.
// ReloadScript performs a hot reload of the given Lua script.
// It snapshots all component states, clears the registry, re-executes
// the script, then restores states by component type name matching.
// GetWindowManager returns the app's window manager.
func (app *App) GetWindowManager() *WindowManager {
	return globalWindowManager
}

// GetDevTools returns the app's DevTools instance.
// GetDevTools returns the app's DevTools instance.
func (app *App) GetDevTools() *DevTools {
	return globalDevTools
}

// Scheduler returns the App's async coroutine scheduler.
// Scheduler returns the App's async coroutine scheduler.
func (app *App) Scheduler() *lua.Scheduler {
	return app.sched
}

// HitTestVNode finds the deepest VNode containing point (px, py).
// Returns the VNode's ID (from props["id"]) or "" if no match.
// GetGlobalEventBus returns the global event bus (for testing).
func GetGlobalEventBus() *EventBus {
	return globalEventBus
}

// ProcessPendingEvents drains and processes all pending events in the channel.
// Used in tests to process lua_callback events posted by event handlers.
