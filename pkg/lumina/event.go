package lumina

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Key codes for keyboard navigation
const (
	KeyTab    = "\t"
	KeyEnter  = "\n"
	KeyEscape = "\x1b"
	KeyUp     = "\x1b[A"
	KeyDown   = "\x1b[B"
	KeyRight  = "\x1b[C"
	KeyLeft   = "\x1b[D"
	KeySpace  = " "
)

// Event represents a user input event.
type Event struct {
	Type      string // "click" | "keydown" | "keyup" | "focus" | "blur" | "change"
	Timestamp int64
	Target    string // component ID

	// Bubbles indicates whether this event type supports bubbling.
	// Set automatically based on event type. Non-bubbling events skip the bubble phase.
	Bubbles bool

	// CurrentTarget is the component ID currently handling the event during bubble phase.
	// Differs from Target during bubbling: Target = original source, CurrentTarget = current handler.
	CurrentTarget string

	// Keyboard
	Key       string // "Enter", "Escape", "a", etc.
	Code      string // physical key code
	Modifiers EventModifiers

	// Mouse
	X, Y   int    // global terminal coordinates
	Button string // "left" | "middle" | "right"

	// Local coordinates (relative to target element's top-left)
	LocalX int
	LocalY int

	// Target VNode reference (set by cell-based hit testing)
	TargetNode *VNode

	// Error for propagation
	stopped          bool
	defaultPrevented bool // set by PreventDefault(); checked by default action handlers
}

// EventModifiers holds keyboard modifier state.
type EventModifiers struct {
	Ctrl  bool
	Shift bool
	Alt   bool
	Meta  bool
}

// PreventDefault prevents the default browser/terminal action for this event.
func (e *Event) PreventDefault() {
	e.defaultPrevented = true
}

// DefaultPrevented returns true if PreventDefault() was called.
func (e *Event) DefaultPrevented() bool {
	return e.defaultPrevented
}

// StopPropagation stops event from bubbling.
func (e *Event) StopPropagation() {
	e.stopped = true
}

// IsStopped returns whether propagation was stopped.
func (e *Event) IsStopped() bool {
	return e.stopped
}

// eventBubbles returns true if the given event type supports bubbling by default.
func eventBubbles(eventType string) bool {
	switch eventType {
	case "click", "mousedown", "mouseup", "mousemove",
		"input", "change", "submit",
		"scroll", "wheel", "contextmenu",
		"keydown", "keyup":
		return true
	case "mouseenter", "mouseleave", "focus", "blur", "resize":
		return false
	default:
		return true // unknown events bubble by default
	}
}

// eventHandler holds a handler for a specific component.
type eventHandler struct {
	componentID string
	handler     func(*Event)
	capture     bool // true = capture phase (parent→child), false = bubble phase (child→parent)
	bridged     bool // true = registered by VNode→EventBus bridge (cleared each render)
}

// EventBus handles event dispatching.
type EventBus struct {
	handlers        map[string][]eventHandler
	shortcuts       map[string]eventHandler // "ctrl+c" → handler
	focusStack      []string                // component ID stack for focus management
	focusedID       string
	focusableIDs    []string          // ordered list of focusable component IDs
	focusableSet    map[string]bool   // O(1) dedup for RegisterFocusable
	vnodeTree       *VNodeTree        // current VNode tree for event bubbling
	bridgedHandlers []bridgedHandler  // handlers from VNode→EventBus bridge (cleared each render)
	mu              sync.RWMutex
}

// VNodeTree tracks parent-child relationships in the VNode tree for event bubbling.
type VNodeTree struct {
	Root    *VNode
	Parents map[*VNode]*VNode  // child → parent mapping
	ByID    map[string]*VNode  // id → vnode lookup
}

// BuildVNodeTree constructs a VNodeTree from a root VNode.
func BuildVNodeTree(root *VNode) *VNodeTree {
	tree := &VNodeTree{
		Root:    root,
		Parents: make(map[*VNode]*VNode),
		ByID:    make(map[string]*VNode),
	}
	buildVNodeTreeRecursive(tree, root, nil)
	fmt.Fprintf(os.Stderr, "[TREE] BuildVNodeTree: %d parents, %d byID\n", len(tree.Parents), len(tree.ByID))
	return tree
}

func buildVNodeTreeRecursive(tree *VNodeTree, node, parent *VNode) {
	if node == nil {
		return
	}
	if parent != nil {
		tree.Parents[node] = parent
	}
	if id, ok := node.Props["id"].(string); ok && id != "" {
		tree.ByID[id] = node
	}
	for _, child := range node.Children {
		buildVNodeTreeRecursive(tree, child, node)
	}
}

// SetVNodeTree sets the current VNode tree for event bubbling.
func (eb *EventBus) SetVNodeTree(tree *VNodeTree) {
	eb.mu.Lock()
	eb.vnodeTree = tree
	eb.mu.Unlock()
}

// GetVNodeTree returns the current VNode tree.
func (eb *EventBus) GetVNodeTree() *VNodeTree {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.vnodeTree
}

// Global event bus
var globalEventBus = NewEventBus()

// NewEventBus creates a new EventBus instance.
func NewEventBus() *EventBus {
	return &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
	}
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

// OnCapture registers an event handler for the capture phase (parent→child, before bubble).
func (eb *EventBus) OnCapture(eventType string, compID string, handler func(*Event)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], eventHandler{
		componentID: compID,
		handler:     handler,
		capture:     true,
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

// Emit fires an event: capture phase (top→target), then target, then bubble phase (target→top).
func (eb *EventBus) Emit(event *Event) {
	// Auto-apply bubble policy based on event type if not explicitly set.
	// This ensures events created without Bubbles field still bubble correctly.
	if !event.Bubbles && eventBubbles(event.Type) {
		event.Bubbles = true
	}

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

	// Capture phase: walk top-down from root to target's parent
	if !event.IsStopped() && eb.vnodeTree != nil && event.Target != "" {
		eb.captureDown(event)
	}

	if event.IsStopped() {
		return
	}

	// Target phase: emit to registered handlers (direct match, non-capture).
	// NOTE: Handlers with componentID="" are global handlers that fire for ALL events
	// of this type, regardless of target. This is intentional — it allows global
	// listeners (e.g., keyboard shortcuts, analytics) to observe all events.
	// Handlers with a specific componentID only fire when event.Target matches.
	event.CurrentTarget = event.Target
	handlers := eb.handlers[event.Type]
	for _, h := range handlers {
		if (h.componentID == "" || event.Target == "" || h.componentID == event.Target) && !h.capture {
			h.handler(event)
			if event.IsStopped() {
				return
			}
		}
	}

	// Bubble phase: only if event supports bubbling
	if !event.IsStopped() && event.Bubbles && eb.vnodeTree != nil && event.Target != "" {
		eb.bubbleUp(event)
	}
}

// captureDown walks from root down to the target's parent, firing capture-phase handlers.
func (eb *EventBus) captureDown(event *Event) {
	tree := eb.vnodeTree
	if tree == nil {
		return
	}

	targetNode, ok := tree.ByID[event.Target]
	if !ok {
		return
	}

	// Build path from root to target's parent
	var path []*VNode
	current := tree.Parents[targetNode]
	for current != nil {
		path = append(path, current)
		current = tree.Parents[current]
	}

	// Walk top-down (reverse the path)
	for i := len(path) - 1; i >= 0; i-- {
		if event.IsStopped() {
			return
		}
		node := path[i]
		if nodeID, ok := node.Props["id"].(string); ok && nodeID != "" {
			event.CurrentTarget = nodeID
			handlers := eb.handlers[event.Type]
			for _, h := range handlers {
				if h.componentID == nodeID && h.capture {
					h.handler(event)
					if event.IsStopped() {
						return
					}
				}
			}
		}
	}
}

// bubbleUp walks up the VNode tree from the target, invoking handlers at each level.
func (eb *EventBus) bubbleUp(event *Event) {
	tree := eb.vnodeTree
	if tree == nil {
		return
	}

	// Find the target VNode
	targetNode, ok := tree.ByID[event.Target]
	if !ok {
		return
	}

	// Walk up parent chain (bubble phase — skip capture handlers)
	current := tree.Parents[targetNode]
	for current != nil {
		if event.IsStopped() {
			return
		}

		// Check if this node has an ID with registered handlers
		if parentID, ok := current.Props["id"].(string); ok && parentID != "" {
			event.CurrentTarget = parentID
			handlers := eb.handlers[event.Type]
			for _, h := range handlers {
				if h.componentID == parentID && !h.capture {
					h.handler(event)
					if event.IsStopped() {
						return
					}
				}
			}
		}

		current = tree.Parents[current]
	}
}

func buildShortcutKey(e *Event) string {
	parts := make([]string, 0, 5)
	if e.Modifiers.Ctrl {
		parts = append(parts, "ctrl")
	}
	if e.Modifiers.Shift {
		parts = append(parts, "shift")
	}
	if e.Modifiers.Alt {
		parts = append(parts, "alt")
	}
	if e.Modifiers.Meta {
		parts = append(parts, "meta")
	}
	if len(parts) > 0 {
		return strings.Join(parts, "+") + "+" + e.Key
	}
	return e.Key
}

// SetFocus sets focus to a component.
func (eb *EventBus) SetFocus(compID string) {
	eb.mu.Lock()
	oldFocus := eb.focusedID
	eb.focusedID = compID

	// Update focus stack
	for i, id := range eb.focusStack {
		if id == compID {
			eb.focusStack = append(eb.focusStack[:i], eb.focusStack[i+1:]...)
			break
		}
	}
	eb.focusStack = append(eb.focusStack, compID)
	eb.mu.Unlock()

	// Emit blur/focus outside lock to avoid deadlock
	if oldFocus != "" && oldFocus != compID {
		eb.Emit(&Event{Type: "blur", Target: oldFocus, Bubbles: false})
	}
	if compID != "" {
		eb.Emit(&Event{Type: "focus", Target: compID, Bubbles: false})
	}
}

// EmitUnsafe emits without locking (caller must hold lock).
func (eb *EventBus) EmitUnsafe(event *Event) {
	// Capture phase: walk top-down from root to target's parent (same as Emit)
	if !event.IsStopped() && eb.vnodeTree != nil && event.Target != "" {
		eb.captureDown(event)
	}

	if event.IsStopped() {
		return
	}

	// Target phase: emit to registered handlers
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

// -----------------------------------------------------------------------
// Focus Scope — trap focus within a subset of focusable IDs
// -----------------------------------------------------------------------

// FocusScope restricts focus cycling to a subset of focusable IDs.
// When active, Tab/Shift+Tab cycle only within this scope.
type FocusScope struct {
	ID           string
	FocusableIDs []string // ordered list of IDs within this scope
}

// focusScopeStack is a stack of active focus scopes.
// The top scope (last element) is the active one.
var focusScopeStack []*FocusScope
var focusScopeMu sync.Mutex

// PushFocusScope pushes a new focus scope onto the stack.
// Focus will be trapped within this scope until PopFocusScope is called.
func PushFocusScope(scope *FocusScope) {
	focusScopeMu.Lock()
	defer focusScopeMu.Unlock()
	focusScopeStack = append(focusScopeStack, scope)
}

// PopFocusScope removes the top focus scope from the stack.
func PopFocusScope() *FocusScope {
	focusScopeMu.Lock()
	defer focusScopeMu.Unlock()
	if len(focusScopeStack) == 0 {
		return nil
	}
	top := focusScopeStack[len(focusScopeStack)-1]
	focusScopeStack = focusScopeStack[:len(focusScopeStack)-1]
	return top
}

// GetActiveFocusScope returns the top focus scope, or nil if none active.
func GetActiveFocusScope() *FocusScope {
	focusScopeMu.Lock()
	defer focusScopeMu.Unlock()
	if len(focusScopeStack) > 0 {
		return focusScopeStack[len(focusScopeStack)-1]
	}
	return nil
}

// ClearFocusScopes removes all focus scopes.
func ClearFocusScopes() {
	focusScopeMu.Lock()
	defer focusScopeMu.Unlock()
	focusScopeStack = nil
}

// RegisterFocusable adds a component ID to the focusable list.
func (eb *EventBus) RegisterFocusable(compID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// O(1) dedup via set
	if eb.focusableSet[compID] {
		return
	}
	if eb.focusableSet == nil {
		eb.focusableSet = make(map[string]bool)
	}
	eb.focusableSet[compID] = true
	eb.focusableIDs = append(eb.focusableIDs, compID)
}

// UnregisterFocusable removes a component ID from the focusable list.
func (eb *EventBus) UnregisterFocusable(compID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	filtered := make([]string, 0, len(eb.focusableIDs))
	for _, id := range eb.focusableIDs {
		if id != compID {
			filtered = append(filtered, id)
		}
	}
	eb.focusableIDs = filtered
	delete(eb.focusableSet, compID)

	// If this was focused, move to next
	if eb.focusedID == compID {
		eb.focusedID = ""
	}
}

// FocusNext moves focus to the next focusable component.
func (eb *EventBus) FocusNext() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Determine which IDs to cycle through
	ids := eb.focusableIDs
	if scope := GetActiveFocusScope(); scope != nil && len(scope.FocusableIDs) > 0 {
		ids = scope.FocusableIDs
	}

	if len(ids) == 0 {
		return
	}

	// Find current index
	currentIdx := -1
	for i, id := range ids {
		if id == eb.focusedID {
			currentIdx = i
			break
		}
	}

	// Blur current
	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID, Bubbles: false})
	}

	// Move to next (wrap around)
	if currentIdx == -1 || currentIdx >= len(ids)-1 {
		eb.focusedID = ids[0]
	} else {
		eb.focusedID = ids[currentIdx+1]
	}

	// Focus new
	eb.EmitUnsafe(&Event{Type: "focus", Target: eb.focusedID, Bubbles: false})
}

// FocusPrev moves focus to the previous focusable component.
func (eb *EventBus) FocusPrev() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Determine which IDs to cycle through
	ids := eb.focusableIDs
	if scope := GetActiveFocusScope(); scope != nil && len(scope.FocusableIDs) > 0 {
		ids = scope.FocusableIDs
	}

	if len(ids) == 0 {
		return
	}

	// Find current index
	currentIdx := -1
	for i, id := range ids {
		if id == eb.focusedID {
			currentIdx = i
			break
		}
	}

	// Blur current
	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID, Bubbles: false})
	}

	// Move to previous (wrap around)
	if currentIdx <= 0 {
		eb.focusedID = ids[len(ids)-1]
	} else {
		eb.focusedID = ids[currentIdx-1]
	}

	// Focus new
	eb.EmitUnsafe(&Event{Type: "focus", Target: eb.focusedID, Bubbles: false})
}

// HandleKeyEvent handles keyboard navigation.
func (eb *EventBus) HandleKeyEvent(key string, modifiers EventModifiers) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	switch key {
	case KeyTab, "Tab":
		if modifiers.Shift {
			eb.focusPrevUnsafe()
		} else {
			eb.focusNextUnsafe()
		}
	case KeyEnter, KeySpace, "Enter":
		// Trigger click on focused component
		if eb.focusedID != "" {
			eb.EmitUnsafe(&Event{
				Type:      "click",
				Target:    eb.focusedID,
				Bubbles:   true,
				Modifiers: modifiers,
			})
		}
	case KeyEscape, "Escape":
		// Close dialogs or blur
		if eb.focusedID != "" {
			eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID, Bubbles: false})
			eb.focusedID = ""
		}
	}
}

// focusNextUnsafe moves to next focusable (caller must hold lock).
func (eb *EventBus) focusNextUnsafe() {
	// Respect FocusScope if active
	ids := eb.focusableIDs
	if scope := GetActiveFocusScope(); scope != nil && len(scope.FocusableIDs) > 0 {
		ids = scope.FocusableIDs
	}

	if len(ids) == 0 {
		return
	}

	currentIdx := -1
	for i, id := range ids {
		if id == eb.focusedID {
			currentIdx = i
			break
		}
	}

	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID, Bubbles: false})
	}

	if currentIdx == -1 || currentIdx >= len(ids)-1 {
		eb.focusedID = ids[0]
	} else {
		eb.focusedID = ids[currentIdx+1]
	}

	eb.EmitUnsafe(&Event{Type: "focus", Target: eb.focusedID, Bubbles: false})
}

// focusPrevUnsafe moves to prev focusable (caller must hold lock).
func (eb *EventBus) focusPrevUnsafe() {
	// Respect FocusScope if active
	ids := eb.focusableIDs
	if scope := GetActiveFocusScope(); scope != nil && len(scope.FocusableIDs) > 0 {
		ids = scope.FocusableIDs
	}

	if len(ids) == 0 {
		return
	}

	currentIdx := -1
	for i, id := range ids {
		if id == eb.focusedID {
			currentIdx = i
			break
		}
	}

	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID, Bubbles: false})
	}

	if currentIdx <= 0 {
		eb.focusedID = ids[len(ids)-1]
	} else {
		eb.focusedID = ids[currentIdx-1]
	}

	eb.EmitUnsafe(&Event{Type: "focus", Target: eb.focusedID, Bubbles: false})
}

// TriggerFocusedClick triggers a click event on the currently focused component.
func (eb *EventBus) TriggerFocusedClick() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "click", Target: eb.focusedID, Bubbles: true})
	}
}

// GetFocusableIDs returns the list of focusable component IDs.
func (eb *EventBus) GetFocusableIDs() []string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	ids := make([]string, len(eb.focusableIDs))
	copy(ids, eb.focusableIDs)
	return ids
}

// IsFocusable checks if a component ID is registered as focusable.
func (eb *EventBus) IsFocusable(compID string) bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, id := range eb.focusableIDs {
		if id == compID {
			return true
		}
	}
	return false
}

// Blur removes focus from current component.
func (eb *EventBus) Blur() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID, Bubbles: false})
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
		Bubbles:   eventBubbles(eventType),
		Timestamp: time.Now().UnixMilli(),
	}
}
