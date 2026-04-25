// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"testing"
)

func TestFocusNavigation(t *testing.T) {
	globalEventBus = NewEventBus() // isolate from other tests
	// Test FocusNext
	globalEventBus.RegisterFocusable("comp_1")
	globalEventBus.RegisterFocusable("comp_2")
	globalEventBus.RegisterFocusable("comp_3")

	// Initial focus
	globalEventBus.SetFocus("comp_1")

	// Test FocusNext
	globalEventBus.FocusNext()
	if globalEventBus.GetFocused() != "comp_2" {
		t.Errorf("Expected comp_2, got %s", globalEventBus.GetFocused())
	}

	globalEventBus.FocusNext()
	if globalEventBus.GetFocused() != "comp_3" {
		t.Errorf("Expected comp_3, got %s", globalEventBus.GetFocused())
	}

	// Wrap around
	globalEventBus.FocusNext()
	if globalEventBus.GetFocused() != "comp_1" {
		t.Errorf("Expected comp_1 (wrap), got %s", globalEventBus.GetFocused())
	}

	// Test FocusPrev
	globalEventBus.FocusPrev()
	if globalEventBus.GetFocused() != "comp_3" {
		t.Errorf("Expected comp_3, got %s", globalEventBus.GetFocused())
	}

	globalEventBus.FocusPrev()
	if globalEventBus.GetFocused() != "comp_2" {
		t.Errorf("Expected comp_2, got %s", globalEventBus.GetFocused())
	}

	// Wrap around
	globalEventBus.FocusPrev()
	if globalEventBus.GetFocused() != "comp_1" {
		t.Errorf("Expected comp_1 (wrap), got %s", globalEventBus.GetFocused())
	}

	// Clean up
	globalEventBus.UnregisterFocusable("comp_1")
	globalEventBus.UnregisterFocusable("comp_2")
	globalEventBus.UnregisterFocusable("comp_3")
}

func TestHandleKeyEventTab(t *testing.T) {
	globalEventBus = NewEventBus() // isolate from other tests
	// Register focusable components
	globalEventBus.RegisterFocusable("btn_1")
	globalEventBus.RegisterFocusable("btn_2")

	// Set initial focus
	globalEventBus.SetFocus("btn_1")

	// Test Tab key advances focus btn_1 → btn_2
	globalEventBus.HandleKeyEvent(KeyTab, EventModifiers{})
	if globalEventBus.GetFocused() != "btn_2" {
		t.Errorf("Tab 1: Expected btn_2, got %s", globalEventBus.GetFocused())
	}

	// Tab again wraps btn_2 → btn_1
	globalEventBus.HandleKeyEvent(KeyTab, EventModifiers{})
	if globalEventBus.GetFocused() != "btn_1" {
		t.Errorf("Tab wrap: Expected btn_1, got %s", globalEventBus.GetFocused())
	}

	// Clean up
	globalEventBus.UnregisterFocusable("btn_1")
	globalEventBus.UnregisterFocusable("btn_2")
}

func TestHandleKeyEventShiftTab(t *testing.T) {
	globalEventBus = NewEventBus() // isolate from other tests
	globalEventBus.RegisterFocusable("btn_1")
	globalEventBus.RegisterFocusable("btn_2")

	// Set focus on btn_2
	globalEventBus.SetFocus("btn_2")

	// Shift+Tab from btn_2 → btn_1
	globalEventBus.HandleKeyEvent(KeyTab, EventModifiers{Shift: true})
	if globalEventBus.GetFocused() != "btn_1" {
		t.Errorf("Shift+Tab from btn_2: Expected btn_1, got %s", globalEventBus.GetFocused())
	}

	// Shift+Tab from btn_1 → btn_2 (wrap to end)
	globalEventBus.HandleKeyEvent(KeyTab, EventModifiers{Shift: true})
	if globalEventBus.GetFocused() != "btn_2" {
		t.Errorf("Shift+Tab wrap: Expected btn_2, got %s", globalEventBus.GetFocused())
	}

	// Clean up
	globalEventBus.UnregisterFocusable("btn_1")
	globalEventBus.UnregisterFocusable("btn_2")
}

func TestHandleKeyEventEscape(t *testing.T) {
	globalEventBus.RegisterFocusable("btn_1")

	globalEventBus.SetFocus("btn_1")

	// Escape clears focus
	globalEventBus.HandleKeyEvent(KeyEscape, EventModifiers{})
	if globalEventBus.GetFocused() != "" {
		t.Errorf("Escape: Expected no focus, got %s", globalEventBus.GetFocused())
	}

	// Clean up
	globalEventBus.UnregisterFocusable("btn_1")
}

func TestRegisterUnregisterFocusable(t *testing.T) {
	// Initially empty
	if globalEventBus.IsFocusable("comp_x") {
		t.Error("comp_x should not be focusable initially")
	}

	// Register
	globalEventBus.RegisterFocusable("comp_x")
	if !globalEventBus.IsFocusable("comp_x") {
		t.Error("comp_x should be focusable after register")
	}

	// Duplicate register should not duplicate
	globalEventBus.RegisterFocusable("comp_x")
	ids := globalEventBus.GetFocusableIDs()
	count := 0
	for _, id := range ids {
		if id == "comp_x" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Expected 1 comp_x, got %d", count)
	}

	// Unregister
	globalEventBus.UnregisterFocusable("comp_x")
	if globalEventBus.IsFocusable("comp_x") {
		t.Error("comp_x should not be focusable after unregister")
	}
}

func TestFocusIndicatorRendering(t *testing.T) {
	// Test that renderFocusIndicator doesn't panic with valid VNode
	frame := NewFrame(20, 10)

	vnode := &VNode{
		X:       2,
		Y:       2,
		W:       5,
		H:       3,
		Focused: true,
	}

	// Should not panic
	renderFocusIndicator(frame, vnode, Rect{X: 0, Y: 0, W: 20, H: 10})

	// Verify border cells were set (corners should have focus chars)
	if frame.Cells[2][2].Char != '[' {
		t.Errorf("Expected top-left '[' at (2,2), got %c", frame.Cells[2][2].Char)
	}
	if frame.Cells[2][6].Char != ']' {
		t.Errorf("Expected top-right ']' at (2,6), got %c", frame.Cells[2][6].Char)
	}

	// Verify foreground color is set for focus indicator
	if frame.Cells[2][2].Foreground != "#FFFF00" {
		t.Errorf("Expected yellow foreground for focus, got %s", frame.Cells[2][2].Foreground)
	}
}

func TestVNodeToFrameWithFocus(t *testing.T) {
	vnode := NewVNode("box")
	vnode.Props["id"] = "focused_comp"
	vnode.Props["background"] = "#333333"
	vnode.W = 10
	vnode.H = 5

	// Without focus
	frame := VNodeToFrame(vnode, 20, 10)
	if frame.FocusedID != "" {
		t.Errorf("Expected empty FocusedID, got %s", frame.FocusedID)
	}

	// With focus
	frame2 := VNodeToFrameWithFocus(vnode, 20, 10, "focused_comp")
	if frame2.FocusedID != "focused_comp" {
		t.Errorf("Expected FocusedID 'focused_comp', got %s", frame2.FocusedID)
	}
	// VNode should be marked as focused
	if !vnode.Focused {
		t.Error("VNode should be marked as Focused")
	}
}
