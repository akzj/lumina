package event

import (
	"sort"

	"github.com/akzj/lumina/pkg/lumina/v2/compositor"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// --- VNodeHitTester ---

// NewVNodeHitTester creates a hit tester from component layers and an occlusion map.
func NewVNodeHitTester(layers []*ComponentLayer, om *compositor.OcclusionMap) *VNodeHitTester {
	m := make(map[string]*ComponentLayer, len(layers))
	for _, cl := range layers {
		m[cl.Layer.ID] = cl
	}
	return &VNodeHitTester{layers: m, om: om}
}

// HitTest performs a two-phase hit test:
//
//	Phase 1: Use the occlusion map to find which layer owns the cell at (x, y).
//	Phase 2: Walk that layer's VNode tree to find the deepest VNode with an ID.
func (ht *VNodeHitTester) HitTest(x, y int) string {
	// Phase 1: layer-level via occlusion map.
	layer := ht.om.OwnerLayer(x, y)
	if layer == nil {
		return ""
	}

	// Phase 2: VNode tree walk.
	cl := ht.layers[layer.ID]
	if cl == nil || cl.VNodeTree == nil {
		return ""
	}
	return findDeepest(cl.VNodeTree, x, y)
}

// findDeepest walks the VNode tree and returns the deepest VNode ID containing (x, y).
// Children are checked in reverse order (last child = visually on top).
func findDeepest(vnode *layout.VNode, x, y int) string {
	if vnode == nil {
		return ""
	}
	// Check if point is inside this VNode's layout rect (absolute screen coords).
	if x < vnode.X || x >= vnode.X+vnode.W || y < vnode.Y || y >= vnode.Y+vnode.H {
		return ""
	}
	// Check children in reverse order.
	for i := len(vnode.Children) - 1; i >= 0; i-- {
		if id := findDeepest(vnode.Children[i], x, y); id != "" {
			return id
		}
	}
	// Return this node's ID if it has one.
	return vnode.ID
}

// --- Dispatcher ---

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers:  make(map[string]HandlerMap),
		parentMap: make(map[string]string),
	}
}

// SetHitTester sets the hit tester used for mouse event dispatch.
func (d *Dispatcher) SetHitTester(ht HitTester) {
	d.hitTester = ht
}

// SetEventObserver sets an event observer for performance tracking.
func (d *Dispatcher) SetEventObserver(obs EventObserver) {
	d.eventObserver = obs
}

// RegisterHandlers registers event handlers for a VNode.
func (d *Dispatcher) RegisterHandlers(vnodeID string, handlers HandlerMap) {
	d.handlers[vnodeID] = handlers
}

// UnregisterHandlers removes all handlers for a VNode.
func (d *Dispatcher) UnregisterHandlers(vnodeID string) {
	delete(d.handlers, vnodeID)
}

// ClearAllHandlers removes all registered handlers.
func (d *Dispatcher) ClearAllHandlers() {
	d.handlers = make(map[string]HandlerMap)
}

// ClearAllFocusables removes all focusable entries.
func (d *Dispatcher) ClearAllFocusables() {
	d.focusables = nil
}

// SetParentMap sets the VNode parent relationships for event bubbling.
func (d *Dispatcher) SetParentMap(parentMap map[string]string) {
	d.parentMap = parentMap
}

// MergeParentMap merges entries into the existing parent map without replacing it.
// Used for incremental updates when only some components were re-rendered.
func (d *Dispatcher) MergeParentMap(entries map[string]string) {
	for k, v := range entries {
		d.parentMap[k] = v
	}
}

// HoveredID returns the currently hovered VNode ID.
func (d *Dispatcher) HoveredID() string {
	return d.hoveredID
}

// FocusedID returns the currently focused VNode ID.
func (d *Dispatcher) FocusedID() string {
	return d.focusedID
}

// Dispatch dispatches an event through the system.
func (d *Dispatcher) Dispatch(e *Event) {
	switch {
	case e.Type == "mousemove" || e.Type == "mousedown" || e.Type == "mouseup":
		d.dispatchMouse(e)
	case e.Type == "scroll":
		d.dispatchScroll(e)
	case e.Type == "keydown" || e.Type == "keyup":
		d.dispatchKey(e)
	default:
		// Direct dispatch (e.g. synthetic events).
		if e.Target != "" {
			d.emit(e.Type, e.Target, e)
		}
	}
}

// dispatchScroll dispatches a scroll event. It hit-tests to find the target,
// then emits the "scroll" event on that target. The event bubbles up via
// the parent map so that a scroll container ancestor can handle it.
func (d *Dispatcher) dispatchScroll(e *Event) {
	targetID := ""
	if d.hitTester != nil {
		targetID = d.hitTester.HitTest(e.X, e.Y)
	}
	e.Target = targetID
	e.Bubbles = true // scroll events bubble to find scroll containers

	// Notify observer.
	if d.eventObserver != nil {
		d.eventObserver.OnEvent(e.Type, targetID != "")
	}

	if targetID != "" {
		d.emit(e.Type, targetID, e)
	}
}

func (d *Dispatcher) dispatchMouse(e *Event) {
	targetID := ""
	if d.hitTester != nil {
		targetID = d.hitTester.HitTest(e.X, e.Y)
	}
	e.Target = targetID

	// Notify observer.
	if d.eventObserver != nil {
		d.eventObserver.OnEvent(e.Type, targetID != "")
	}

	// Hover tracking on mousemove.
	if e.Type == "mousemove" {
		if targetID != d.hoveredID {
			if d.hoveredID != "" {
				d.emit("mouseleave", d.hoveredID, &Event{
					Type:   "mouseleave",
					X:      e.X,
					Y:      e.Y,
					Target: d.hoveredID,
				})
			}
			if targetID != "" {
				d.emit("mouseenter", targetID, &Event{
					Type:   "mouseenter",
					X:      e.X,
					Y:      e.Y,
					Target: targetID,
				})
			}
			d.hoveredID = targetID
		}
	}

	// Click + focus on mousedown.
	if e.Type == "mousedown" {
		if targetID != "" {
			d.emit("click", targetID, &Event{
				Type:    "click",
				X:       e.X,
				Y:       e.Y,
				Target:  targetID,
				Bubbles: e.Bubbles,
			})
		}
		if d.isFocusable(targetID) {
			d.SetFocus(targetID)
		}
	}

	// Also dispatch the original event type (mousedown, mouseup, mousemove).
	if targetID != "" {
		d.emit(e.Type, targetID, e)
	}
}

func (d *Dispatcher) dispatchKey(e *Event) {
	// Notify observer.
	if d.eventObserver != nil {
		d.eventObserver.OnEvent(e.Type, d.focusedID != "")
	}

	if e.Type == "keydown" {
		if e.Key == "Tab" {
			d.FocusNext()
			return
		}
		if e.Key == "Shift+Tab" {
			d.FocusPrev()
			return
		}
	}
	// Dispatch to focused VNode.
	if d.focusedID != "" {
		e.Target = d.focusedID
		d.emit(e.Type, d.focusedID, e)
	}
}

// emit calls the handler for the given event type on the target VNode,
// then bubbles to parents if the event has Bubbles set.
func (d *Dispatcher) emit(eventType, targetID string, e *Event) {
	if targetID == "" {
		return
	}
	if hm, ok := d.handlers[targetID]; ok {
		if handler, ok := hm[eventType]; ok {
			handler(e)
		}
	}
	// Bubble to parent if event supports bubbling.
	if e.Bubbles {
		if parentID, ok := d.parentMap[targetID]; ok && parentID != "" {
			d.emit(eventType, parentID, e)
		}
	}
}

// --- Focus management ---

// SetFocus sets focus to the given VNode, emitting "blur" on old and "focus" on new.
func (d *Dispatcher) SetFocus(vnodeID string) {
	if vnodeID == d.focusedID {
		return
	}
	oldID := d.focusedID
	d.focusedID = vnodeID

	if oldID != "" {
		d.emit("blur", oldID, &Event{Type: "blur", Target: oldID})
	}
	if vnodeID != "" {
		d.emit("focus", vnodeID, &Event{Type: "focus", Target: vnodeID})
	}
}

// RegisterFocusable adds a VNode to the focusable list sorted by tabIndex.
func (d *Dispatcher) RegisterFocusable(vnodeID string, tabIndex int) {
	// Remove existing entry if present (avoid duplicates).
	d.removeFocusable(vnodeID)

	entry := focusEntry{vnodeID: vnodeID, tabIndex: tabIndex}
	d.focusables = append(d.focusables, entry)
	sort.SliceStable(d.focusables, func(i, j int) bool {
		return d.focusables[i].tabIndex < d.focusables[j].tabIndex
	})
}

// UnregisterFocusable removes a VNode from the focusable list.
// If the removed VNode was focused, focus moves to the next focusable.
func (d *Dispatcher) UnregisterFocusable(vnodeID string) {
	wasFocused := d.focusedID == vnodeID
	d.removeFocusable(vnodeID)
	if wasFocused {
		if len(d.focusables) > 0 {
			d.SetFocus(d.focusables[0].vnodeID)
		} else {
			d.focusedID = ""
		}
	}
}

// FocusNext moves focus to the next focusable VNode (wraps around).
func (d *Dispatcher) FocusNext() {
	if len(d.focusables) == 0 {
		return
	}
	idx := d.focusIndex()
	next := (idx + 1) % len(d.focusables)
	d.SetFocus(d.focusables[next].vnodeID)
}

// FocusPrev moves focus to the previous focusable VNode (wraps around).
func (d *Dispatcher) FocusPrev() {
	if len(d.focusables) == 0 {
		return
	}
	idx := d.focusIndex()
	prev := (idx - 1 + len(d.focusables)) % len(d.focusables)
	d.SetFocus(d.focusables[prev].vnodeID)
}

// HasFocusables returns true if there are any focusable VNodes registered.
func (d *Dispatcher) HasFocusables() bool {
	return len(d.focusables) > 0
}

// GetFocusableIDs returns the IDs of all registered focusable VNodes, sorted by tabIndex.
func (d *Dispatcher) GetFocusableIDs() []string {
	ids := make([]string, len(d.focusables))
	for i, f := range d.focusables {
		ids[i] = f.vnodeID
	}
	return ids
}

// isFocusable checks if a VNode is in the focusable list.
func (d *Dispatcher) isFocusable(vnodeID string) bool {
	for _, f := range d.focusables {
		if f.vnodeID == vnodeID {
			return true
		}
	}
	return false
}

// focusIndex returns the index of the currently focused VNode in focusables.
// Returns -1 if not found.
func (d *Dispatcher) focusIndex() int {
	for i, f := range d.focusables {
		if f.vnodeID == d.focusedID {
			return i
		}
	}
	return -1
}

// removeFocusable removes a VNode from the focusable list.
func (d *Dispatcher) removeFocusable(vnodeID string) {
	for i, f := range d.focusables {
		if f.vnodeID == vnodeID {
			d.focusables = append(d.focusables[:i], d.focusables[i+1:]...)
			return
		}
	}
}
