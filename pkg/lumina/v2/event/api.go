// Package event provides event dispatch (mouse/keyboard), hit-testing via
// occlusion map + VNode tree walk, and focus management for Lumina v2.
package event

import (
	"github.com/akzj/lumina/pkg/lumina/v2/compositor"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// Event represents an input event.
type Event struct {
	Type      string // "mousedown", "mouseup", "mousemove", "mouseenter",
	//                  "mouseleave", "click", "keydown", "keyup", "focus", "blur"
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

// HitTester resolves screen coordinates to a target ID.
type HitTester interface {
	HitTest(x, y int) string
}

// ComponentLayer extends compositor.Layer with a VNode tree for sub-component events.
type ComponentLayer struct {
	*compositor.Layer
	VNodeTree *layout.VNode
}

// VNodeHitTester implements HitTester using an occlusion map + VNode tree walk.
type VNodeHitTester struct {
	layers map[string]*ComponentLayer // layerID → ComponentLayer
	om     *compositor.OcclusionMap
}

// focusEntry is a focusable VNode with its tab index.
type focusEntry struct {
	vnodeID  string
	tabIndex int
}

// EventObserver is notified when events are dispatched.
// Used by the perf tracker without creating a dependency from event → perf.
type EventObserver interface {
	OnEvent(eventType string, dispatched bool)
}

// Dispatcher dispatches events to handlers.
type Dispatcher struct {
	hitTester     HitTester
	handlers      map[string]HandlerMap // vnodeID → HandlerMap
	hoveredID     string
	focusedID     string
	focusables    []focusEntry          // sorted by tabIndex
	parentMap     map[string]string     // vnodeID → parentVNodeID (for bubbling)
	eventObserver EventObserver
}
