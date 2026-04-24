package lumina

import (
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

	// Keyboard
	Key       string // "Enter", "Escape", "a", etc.
	Code      string // physical key code
	Modifiers EventModifiers

	// Mouse
	X, Y   int    // terminal coordinates
	Button string // "left" | "middle" | "right"

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
	capture     bool // true = capture phase (parent→child), false = bubble phase (child→parent)
}

// EventBus handles event dispatching.
type EventBus struct {
	handlers     map[string][]eventHandler
	shortcuts    map[string]eventHandler // "ctrl+c" → handler
	focusStack   []string                // component ID stack for focus management
	focusedID    string
	focusableIDs []string // ordered list of focusable component IDs
	vnodeTree    *VNodeTree // current VNode tree for event bubbling
	mu           sync.RWMutex
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

	// Target phase: emit to registered handlers (direct match, non-capture)
	handlers := eb.handlers[event.Type]
	for _, h := range handlers {
		if (event.Target == "" || h.componentID == event.Target) && !h.capture {
			h.handler(event)
			if event.IsStopped() {
				return
			}
		}
	}

	// Bubble phase: walk up the VNode tree if we have one and event has a target
	if !event.IsStopped() && eb.vnodeTree != nil && event.Target != "" {
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
	var key string
	if e.Modifiers.Ctrl {
		key += "ctrl"
	}
	if e.Modifiers.Shift {
		key += "shift"
	}
	if e.Modifiers.Alt {
		key += "alt"
	}
	if e.Modifiers.Meta {
		key += "meta"
	}
	if len(key) > 0 {
		key += "+"
	}
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

// RegisterFocusable adds a component ID to the focusable list.
func (eb *EventBus) RegisterFocusable(compID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Check if already registered
	for _, id := range eb.focusableIDs {
		if id == compID {
			return
		}
	}
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

	// If this was focused, move to next
	if eb.focusedID == compID {
		eb.focusedID = ""
	}
}

// FocusNext moves focus to the next focusable component.
func (eb *EventBus) FocusNext() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if len(eb.focusableIDs) == 0 {
		return
	}

	// Find current index
	currentIdx := -1
	for i, id := range eb.focusableIDs {
		if id == eb.focusedID {
			currentIdx = i
			break
		}
	}

	// Blur current
	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID})
	}

	// Move to next
	if currentIdx == -1 || currentIdx >= len(eb.focusableIDs)-1 {
		eb.focusedID = eb.focusableIDs[0]
	} else {
		eb.focusedID = eb.focusableIDs[currentIdx+1]
	}

	// Focus new
	eb.EmitUnsafe(&Event{Type: "focus", Target: eb.focusedID})
}

// FocusPrev moves focus to the previous focusable component.
func (eb *EventBus) FocusPrev() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if len(eb.focusableIDs) == 0 {
		return
	}

	// Find current index
	currentIdx := -1
	for i, id := range eb.focusableIDs {
		if id == eb.focusedID {
			currentIdx = i
			break
		}
	}

	// Blur current
	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID})
	}

	// Move to previous
	if currentIdx <= 0 {
		eb.focusedID = eb.focusableIDs[len(eb.focusableIDs)-1]
	} else {
		eb.focusedID = eb.focusableIDs[currentIdx-1]
	}

	// Focus new
	eb.EmitUnsafe(&Event{Type: "focus", Target: eb.focusedID})
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
				Modifiers: modifiers,
			})
		}
	case KeyEscape, "Escape":
		// Close dialogs or blur
		if eb.focusedID != "" {
			eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID})
			eb.focusedID = ""
		}
	}
}

// focusNextUnsafe moves to next focusable (caller must hold lock).
func (eb *EventBus) focusNextUnsafe() {
	if len(eb.focusableIDs) == 0 {
		return
	}

	currentIdx := -1
	for i, id := range eb.focusableIDs {
		if id == eb.focusedID {
			currentIdx = i
			break
		}
	}

	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID})
	}

	if currentIdx == -1 || currentIdx >= len(eb.focusableIDs)-1 {
		eb.focusedID = eb.focusableIDs[0]
	} else {
		eb.focusedID = eb.focusableIDs[currentIdx+1]
	}

	eb.EmitUnsafe(&Event{Type: "focus", Target: eb.focusedID})
}

// focusPrevUnsafe moves to prev focusable (caller must hold lock).
func (eb *EventBus) focusPrevUnsafe() {
	if len(eb.focusableIDs) == 0 {
		return
	}

	currentIdx := -1
	for i, id := range eb.focusableIDs {
		if id == eb.focusedID {
			currentIdx = i
			break
		}
	}

	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "blur", Target: eb.focusedID})
	}

	if currentIdx <= 0 {
		eb.focusedID = eb.focusableIDs[len(eb.focusableIDs)-1]
	} else {
		eb.focusedID = eb.focusableIDs[currentIdx-1]
	}

	eb.EmitUnsafe(&Event{Type: "focus", Target: eb.focusedID})
}

// TriggerFocusedClick triggers a click event on the currently focused component.
func (eb *EventBus) TriggerFocusedClick() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.focusedID != "" {
		eb.EmitUnsafe(&Event{Type: "click", Target: eb.focusedID})
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
