package event

import (
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/compositor"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// --- helpers ---

// makeFilledBuffer creates a buffer fully filled with a character so the
// occlusion map claims ownership of every cell.
func makeFilledBuffer(w, h int) *buffer.Buffer {
	b := buffer.New(w, h)
	b.Fill(buffer.Rect{X: 0, Y: 0, W: w, H: h}, buffer.Cell{Char: 'X'})
	return b
}

// setupSingleComponent creates a dispatcher with one component layer containing
// a VNode tree: root (box, full size) with two text children side by side.
//
//	btn1: (0,0)-(5,1)   btn2: (5,0)-(10,1)
//	rest of root: (0,0)-(10,5)
func setupSingleComponent(w, h int) (*Dispatcher, *VNodeHitTester) {
	buf := makeFilledBuffer(w, h)
	layer := &compositor.Layer{
		ID: "comp1", Buffer: buf,
		Rect: buffer.Rect{X: 0, Y: 0, W: w, H: h}, ZIndex: 0,
	}

	root := &layout.VNode{
		Type: "box", ID: "root",
		Children: []*layout.VNode{
			{Type: "text", ID: "btn1"},
			{Type: "text", ID: "btn2"},
		},
	}
	root.X, root.Y, root.W, root.H = 0, 0, w, h
	root.Children[0].X, root.Children[0].Y = 0, 0
	root.Children[0].W, root.Children[0].H = 5, 1
	root.Children[1].X, root.Children[1].Y = 5, 0
	root.Children[1].W, root.Children[1].H = 5, 1

	om := compositor.NewOcclusionMap(w, h)
	om.Build([]*compositor.Layer{layer})

	cl := &ComponentLayer{Layer: layer, VNodeTree: root}
	ht := NewVNodeHitTester([]*ComponentLayer{cl}, om)

	d := NewDispatcher()
	d.SetHitTester(ht)
	d.SetParentMap(map[string]string{
		"btn1": "root",
		"btn2": "root",
	})
	return d, ht
}

// --- Tests ---

func TestDispatcher_HitTest_Click(t *testing.T) {
	d, _ := setupSingleComponent(10, 5)

	var clicked string
	d.RegisterHandlers("btn1", HandlerMap{
		"click": func(e *Event) { clicked = e.Target },
	})

	// Click at (2, 0) — inside btn1.
	d.Dispatch(&Event{Type: "mousedown", X: 2, Y: 0})
	if clicked != "btn1" {
		t.Errorf("expected click on btn1, got %q", clicked)
	}
}

func TestDispatcher_HoverEnterLeave(t *testing.T) {
	d, _ := setupSingleComponent(10, 5)

	var entered, left string
	d.RegisterHandlers("btn1", HandlerMap{
		"mouseenter": func(e *Event) { entered = e.Target },
		"mouseleave": func(e *Event) { left = e.Target },
	})
	d.RegisterHandlers("btn2", HandlerMap{
		"mouseenter": func(e *Event) { entered = e.Target },
	})

	// Move to btn1.
	d.Dispatch(&Event{Type: "mousemove", X: 2, Y: 0})
	if entered != "btn1" {
		t.Errorf("expected mouseenter on btn1, got %q", entered)
	}

	// Move to btn2 — btn1 should get mouseleave, btn2 should get mouseenter.
	entered = ""
	d.Dispatch(&Event{Type: "mousemove", X: 7, Y: 0})
	if left != "btn1" {
		t.Errorf("expected mouseleave on btn1, got %q", left)
	}
	if entered != "btn2" {
		t.Errorf("expected mouseenter on btn2, got %q", entered)
	}
}

func TestDispatcher_HoverReenter(t *testing.T) {
	d, _ := setupSingleComponent(10, 5)

	enterCount := 0
	d.RegisterHandlers("btn1", HandlerMap{
		"mouseenter": func(e *Event) { enterCount++ },
	})

	// Enter btn1.
	d.Dispatch(&Event{Type: "mousemove", X: 2, Y: 0})
	if enterCount != 1 {
		t.Fatalf("expected 1 enter, got %d", enterCount)
	}

	// Move to btn2.
	d.Dispatch(&Event{Type: "mousemove", X: 7, Y: 0})

	// Return to btn1.
	d.Dispatch(&Event{Type: "mousemove", X: 2, Y: 0})
	if enterCount != 2 {
		t.Errorf("expected 2 enters after re-entry, got %d", enterCount)
	}
}

func TestDispatcher_ClickOccluded(t *testing.T) {
	// Two layers at same position. Layer B (z=100) occludes layer A (z=0).
	bufA := makeFilledBuffer(10, 5)
	layerA := &compositor.Layer{
		ID: "compA", Buffer: bufA,
		Rect: buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, ZIndex: 0,
	}
	bufB := makeFilledBuffer(10, 5)
	layerB := &compositor.Layer{
		ID: "compB", Buffer: bufB,
		Rect: buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, ZIndex: 100,
	}

	vnodeA := &layout.VNode{Type: "box", ID: "nodeA"}
	vnodeA.X, vnodeA.Y, vnodeA.W, vnodeA.H = 0, 0, 10, 5
	vnodeB := &layout.VNode{Type: "box", ID: "nodeB"}
	vnodeB.X, vnodeB.Y, vnodeB.W, vnodeB.H = 0, 0, 10, 5

	om := compositor.NewOcclusionMap(10, 5)
	om.Build([]*compositor.Layer{layerA, layerB})

	clA := &ComponentLayer{Layer: layerA, VNodeTree: vnodeA}
	clB := &ComponentLayer{Layer: layerB, VNodeTree: vnodeB}
	ht := NewVNodeHitTester([]*ComponentLayer{clA, clB}, om)

	d := NewDispatcher()
	d.SetHitTester(ht)

	var clickedA, clickedB bool
	d.RegisterHandlers("nodeA", HandlerMap{
		"click": func(e *Event) { clickedA = true },
	})
	d.RegisterHandlers("nodeB", HandlerMap{
		"click": func(e *Event) { clickedB = true },
	})

	d.Dispatch(&Event{Type: "mousedown", X: 5, Y: 2})

	if clickedA {
		t.Error("nodeA should NOT receive click (occluded)")
	}
	if !clickedB {
		t.Error("nodeB should receive click (top layer)")
	}
}

func TestDispatcher_SubComponentHitTest(t *testing.T) {
	// A "form" component with two buttons at different positions.
	buf := makeFilledBuffer(20, 10)
	layer := &compositor.Layer{
		ID: "form", Buffer: buf,
		Rect: buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, ZIndex: 0,
	}

	root := &layout.VNode{
		Type: "box", ID: "form-root",
		Children: []*layout.VNode{
			{Type: "text", ID: "submit-btn"},
			{Type: "text", ID: "cancel-btn"},
		},
	}
	root.X, root.Y, root.W, root.H = 0, 0, 20, 10
	root.Children[0].X, root.Children[0].Y = 2, 2
	root.Children[0].W, root.Children[0].H = 6, 1
	root.Children[1].X, root.Children[1].Y = 10, 2
	root.Children[1].W, root.Children[1].H = 6, 1

	om := compositor.NewOcclusionMap(20, 10)
	om.Build([]*compositor.Layer{layer})

	cl := &ComponentLayer{Layer: layer, VNodeTree: root}
	ht := NewVNodeHitTester([]*ComponentLayer{cl}, om)

	d := NewDispatcher()
	d.SetHitTester(ht)

	var clicked string
	d.RegisterHandlers("submit-btn", HandlerMap{
		"click": func(e *Event) { clicked = "submit" },
	})
	d.RegisterHandlers("cancel-btn", HandlerMap{
		"click": func(e *Event) { clicked = "cancel" },
	})

	// Click on submit button at (4, 2).
	d.Dispatch(&Event{Type: "mousedown", X: 4, Y: 2})
	if clicked != "submit" {
		t.Errorf("expected submit, got %q", clicked)
	}

	// Click on cancel button at (12, 2).
	clicked = ""
	d.Dispatch(&Event{Type: "mousedown", X: 12, Y: 2})
	if clicked != "cancel" {
		t.Errorf("expected cancel, got %q", clicked)
	}
}

func TestDispatcher_EventBubbling(t *testing.T) {
	d, _ := setupSingleComponent(10, 5)

	var bubbledTo string
	d.RegisterHandlers("btn1", HandlerMap{
		"click": func(e *Event) { /* child handles */ },
	})
	d.RegisterHandlers("root", HandlerMap{
		"click": func(e *Event) { bubbledTo = "root" },
	})

	// Click on btn1 with Bubbles=true — should bubble to root.
	d.Dispatch(&Event{Type: "mousedown", X: 2, Y: 0, Bubbles: true})
	if bubbledTo != "root" {
		t.Errorf("expected event to bubble to root, got %q", bubbledTo)
	}
}

func TestDispatcher_NoHandler(t *testing.T) {
	d, _ := setupSingleComponent(10, 5)

	// Click on btn1 without registering any handler — should not panic.
	d.Dispatch(&Event{Type: "mousedown", X: 2, Y: 0})
	// If we get here, no panic occurred.
}

func TestDispatcher_Focus_Tab(t *testing.T) {
	d, _ := setupSingleComponent(10, 5)

	d.RegisterFocusable("btn1", 1)
	d.RegisterFocusable("btn2", 2)

	var focused []string
	d.RegisterHandlers("btn1", HandlerMap{
		"focus": func(e *Event) { focused = append(focused, "btn1") },
	})
	d.RegisterHandlers("btn2", HandlerMap{
		"focus": func(e *Event) { focused = append(focused, "btn2") },
	})

	// Set initial focus.
	d.SetFocus("btn1")

	// Tab → should move to btn2.
	d.Dispatch(&Event{Type: "keydown", Key: "Tab"})
	if d.FocusedID() != "btn2" {
		t.Errorf("after Tab: expected btn2, got %q", d.FocusedID())
	}

	// Tab again → should wrap to btn1.
	d.Dispatch(&Event{Type: "keydown", Key: "Tab"})
	if d.FocusedID() != "btn1" {
		t.Errorf("after second Tab: expected btn1, got %q", d.FocusedID())
	}
}

func TestDispatcher_Focus_ShiftTab(t *testing.T) {
	d, _ := setupSingleComponent(10, 5)

	d.RegisterFocusable("btn1", 1)
	d.RegisterFocusable("btn2", 2)

	d.SetFocus("btn2")

	// Shift+Tab → should move to btn1.
	d.Dispatch(&Event{Type: "keydown", Key: "Shift+Tab"})
	if d.FocusedID() != "btn1" {
		t.Errorf("after Shift+Tab: expected btn1, got %q", d.FocusedID())
	}

	// Shift+Tab again → should wrap to btn2.
	d.Dispatch(&Event{Type: "keydown", Key: "Shift+Tab"})
	if d.FocusedID() != "btn2" {
		t.Errorf("after second Shift+Tab: expected btn2, got %q", d.FocusedID())
	}
}

func TestDispatcher_Focus_Click(t *testing.T) {
	d, _ := setupSingleComponent(10, 5)

	d.RegisterFocusable("btn1", 1)
	d.RegisterFocusable("btn2", 2)

	var focusedVia string
	d.RegisterHandlers("btn2", HandlerMap{
		"focus": func(e *Event) { focusedVia = "btn2" },
	})

	// Click on btn2 → should receive focus.
	d.Dispatch(&Event{Type: "mousedown", X: 7, Y: 0})
	if d.FocusedID() != "btn2" {
		t.Errorf("expected btn2 focused, got %q", d.FocusedID())
	}
	if focusedVia != "btn2" {
		t.Errorf("expected focus event on btn2, got %q", focusedVia)
	}
}

func TestDispatcher_Keyboard_ToFocused(t *testing.T) {
	d, _ := setupSingleComponent(10, 5)

	d.RegisterFocusable("btn1", 1)
	d.SetFocus("btn1")

	var receivedKey string
	d.RegisterHandlers("btn1", HandlerMap{
		"keydown": func(e *Event) { receivedKey = e.Key },
	})

	d.Dispatch(&Event{Type: "keydown", Key: "Enter"})
	if receivedKey != "Enter" {
		t.Errorf("expected Enter key, got %q", receivedKey)
	}
}
