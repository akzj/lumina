package lumina

import (
	"sync"
	"time"
)

// Event represents a user input event.
type Event struct {
	Type      string // "click" | "keydown" | "keyup" | "focus" | "blur" | "change"
	Timestamp int64
	Target    string // component ID

	// Keyboard
	Key      string // "Enter", "Escape", "a", etc.
	Code     string // physical key code
	Modifiers EventModifiers

	// Mouse
	X, Y    int    // terminal coordinates
	Button  string // "left" | "middle" | "right"

	// Error for propagation
	stopped bool
}

// EventModifiers holds keyboard modifier state.
type EventModifiers struct {
	Ctrl  bool
	Shift bool
	Alt   bool
	Meta  bool
}

// PreventDefault prevents default behavior.
func (e *Event) PreventDefault() {
	// Mark event as handled
}

// StopPropagation stops event from bubbling.
func (e *Event) StopPropagation() {
	e.stopped = true
}

// IsStopped returns whether propagation was stopped.
func (e *Event) IsStopped() bool {
	return e.stopped
}

// eventHandler holds a handler for a specific component.
type eventHandler struct {
	componentID string
	handler     func(*Event)
}

// EventBus handles event dispatching.
type EventBus struct {
	handlers    map[string][]eventHandler
	shortcuts   map[string]eventHandler // "ctrl+c" → handler
	focusStack  []string               // component ID stack for focus management
	focusedID   string
	mu          sync.RWMutex
}

// Global event bus
var globalEventBus = &EventBus{
	handlers:  make(map[string][]eventHandler),
	shortcuts: make(map[string]eventHandler),
}

// On registers an event handler.
func (eb *EventBus) On(eventType string, compID string, handler func(*Event)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], eventHandler{
		componentID: compID,
		handler:     handler,
	})
}

// Off unregisters event handlers for a component.
func (eb *EventBus) Off(eventType string, compID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	handlers := eb.handlers[eventType]
	filtered := make([]eventHandler, 0, len(handlers))
	for _, h := range handlers {
		if h.componentID != compID {
			filtered = append(filtered, h)
		}
	}
	eb.handlers[eventType] = filtered
}

// Emit fires an event to all registered handlers.
func (eb *EventBus) Emit(event *Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// Check global shortcuts first
	if shortcut := buildShortcutKey(event); shortcut != "" {
		if handler, ok := eb.shortcuts[shortcut]; ok {
			handler.handler(event)
			if event.IsStopped() {
				return
			}
		}
	}

	// Emit to registered handlers
	handlers := eb.handlers[event.Type]
	for _, h := range handlers {
		h.handler(event)
		if event.IsStopped() {
			return
		}
	}
}

func buildShortcutKey(e *Event) string {
	var key string
	if e.Modifiers.Ctrl  { key += "ctrl" }
	if e.Modifiers.Shift { key += "shift" }
	if e.Modifiers.Alt   { key += "alt" }
	if e.Modifiers.Meta  { key += "meta" }
	if len(key) > 0 { key += "+" }
	key += e.Key
	return key
}

// SetFocus sets focus to a component.
func (eb *EventBus) SetFocus(compID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Blur previous
	if eb.focusedID != "" && eb.focusedID != compID {
		eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID})
	}

	// Focus new
	eb.focusedID = compID
	eb.EmitUnsafe(&Event{Type: "focus", Target: compID})

	// Update focus stack
	for i, id := range eb.focusStack {
		if id == compID {
			eb.focusStack = append(eb.focusStack[:i], eb.focusStack[i+1:]...)
			break
		}
	}
	eb.focusStack = append(eb.focusStack, compID)
}

// EmitUnsafe emits without locking (caller must hold lock).
func (eb *EventBus) EmitUnsafe(event *Event) {
	handlers := eb.handlers[event.Type]
	for _, h := range handlers {
		h.handler(event)
		if event.IsStopped() {
			return
		}
	}
}

// GetFocused returns the currently focused component ID.
func (eb *EventBus) GetFocused() string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.focusedID
}

// Blur removes focus from current component.
func (eb *EventBus) Blur() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID})
		eb.focusedID = ""
	}
}

// PushFocusTrap adds a component to focus trap stack.
func (eb *EventBus) PushFocusTrap(compID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.focusStack = append(eb.focusStack, compID)
}

// PopFocusTrap removes top component from focus trap stack.
func (eb *EventBus) PopFocusTrap() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if len(eb.focusStack) > 0 {
		eb.focusStack = eb.focusStack[:len(eb.focusStack)-1]
	}
}

// normalizeShortcutKey normalizes a shortcut key string.
func normalizeShortcutKey(key string) string {
	lower := ""
	for _, c := range key {
		if c >= 'A' && c <= 'Z' {
			lower += string(c + 32) // tolower
		} else if c != ' ' { // preserve + and - separators
			lower += string(c)
		}
	}
	return lower
}

// CreateEvent creates a new event with current timestamp.
func CreateEvent(eventType string) *Event {
	return &Event{
		Type:      eventType,
		Timestamp: time.Now().UnixMilli(),
	}
}
