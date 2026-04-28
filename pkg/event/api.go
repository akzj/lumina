// Package event provides event types for Lumina v2.
package event

// Event represents an input event.
type Event struct {
	Type      string // "mousedown", "mouseup", "mousemove", "mouseenter",
	//                  "mouseleave", "click", "keydown", "keyup", "focus", "blur", "scroll"
	X, Y      int    // mouse position (screen coordinates)
	LocalX    int    // mouse position relative to target VNode
	LocalY    int
	Key       string // key name for keyboard events
	Target    string // VNode ID that should handle this event
	Bubbles   bool
	Timestamp int64
}

// EventHandler is a function that handles an event.
type EventHandler func(e *Event)

// HandlerMap maps event types to handlers.
type HandlerMap map[string]EventHandler

// EventObserver is notified when events are dispatched.
// Used by the perf tracker without creating a dependency from event → perf.
type EventObserver interface {
	OnEvent(eventType string, dispatched bool)
}
