package v2

import (
	"fmt"
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// =============================================================================
// CATEGORY 1: Multi-Window Compositing
// =============================================================================

func TestIntegration_Compositing_ThreeOverlappingWindows(t *testing.T) {
	app, ta := NewTestApp(40, 20)

	// bg(z=0, full screen with background)
	app.RegisterComponent("bg", "bg", buffer.Rect{X: 0, Y: 0, W: 40, H: 20}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#000000"
			child := layout.NewVNode("text")
			child.Content = "B"
			root.AddChild(child)
			return root
		})

	// panel(z=50, 10x5 at (5,5))
	app.RegisterComponent("panel", "panel", buffer.Rect{X: 5, Y: 5, W: 10, H: 5}, 50,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#111111"
			child := layout.NewVNode("text")
			child.Content = "P"
			root.AddChild(child)
			return root
		})

	// dialog(z=100, 6x3 at (8,6))
	app.RegisterComponent("dialog", "dialog", buffer.Rect{X: 8, Y: 6, W: 6, H: 3}, 100,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#222222"
			child := layout.NewVNode("text")
			child.Content = "D"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Background text 'B' at (0,0)
	if c := ta.LastScreen.Get(0, 0).Char; c != 'B' {
		t.Errorf("expected 'B' at (0,0), got %q", c)
	}

	// Panel text 'P' at its origin (5,5)
	if c := ta.LastScreen.Get(5, 5).Char; c != 'P' {
		t.Errorf("expected 'P' at (5,5), got %q", c)
	}

	// Dialog text 'D' at its origin (8,6)
	if c := ta.LastScreen.Get(8, 6).Char; c != 'D' {
		t.Errorf("expected 'D' at (8,6), got %q", c)
	}

	// Overlap point (10,7) — inside both panel and dialog area → dialog background wins (z=100)
	overlapCell := ta.LastScreen.Get(10, 7)
	if overlapCell.Background != "#222222" {
		t.Errorf("overlap (10,7): expected dialog bg '#222222', got %q", overlapCell.Background)
	}
	// Non-text area of dialog is filled with space
	if overlapCell.Char != ' ' {
		t.Errorf("overlap (10,7): expected space (bg fill), got %q", overlapCell.Char)
	}
}

func TestIntegration_Compositing_WindowMoveRevealsBackground(t *testing.T) {
	app, ta := NewTestApp(40, 20)

	app.RegisterComponent("bg", "bg", buffer.Rect{X: 0, Y: 0, W: 40, H: 20}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#000000"
			child := layout.NewVNode("text")
			child.Content = "B"
			root.AddChild(child)
			return root
		})

	app.RegisterComponent("dlg", "dialog", buffer.Rect{X: 5, Y: 5, W: 10, H: 5}, 100,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#FFFFFF"
			child := layout.NewVNode("text")
			child.Content = "D"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Verify dialog visible at (5,5) — text 'D' at dialog origin
	if c := ta.LastScreen.Get(5, 5).Char; c != 'D' {
		t.Errorf("before move: expected 'D' at (5,5), got %q", c)
	}
	// Interior of dialog at (6,6) should have dialog background
	if bg := ta.LastScreen.Get(6, 6).Background; bg != "#FFFFFF" {
		t.Errorf("before move: expected dialog bg at (6,6), got %q", bg)
	}

	// Move dialog to (20,5)
	app.MoveComponent("dlg", buffer.Rect{X: 20, Y: 5, W: 10, H: 5})
	app.RenderDirty()

	// Old position (5,5) should now show background (bg space or 'B' depending on position)
	// At (5,5), the bg component's text 'B' is at (0,0), so (5,5) shows bg space
	oldCell := ta.LastScreen.Get(5, 5)
	if oldCell.Background != "#000000" {
		t.Errorf("after move: expected bg '#000000' at (5,5), got %q", oldCell.Background)
	}

	// New position (20,5) should show dialog text 'D'
	newCell := ta.LastScreen.Get(20, 5)
	if newCell.Char != 'D' {
		t.Errorf("after move: expected 'D' at (20,5), got %q", newCell.Char)
	}
}

func TestIntegration_Compositing_DynamicAddRemove(t *testing.T) {
	app, ta := NewTestApp(20, 10)

	app.RegisterComponent("bg", "bg", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#000000"
			child := layout.NewVNode("text")
			child.Content = "B"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Background at (0,0) shows text 'B'; at (5,5) shows bg space with bg color
	if c := ta.LastScreen.Get(0, 0).Char; c != 'B' {
		t.Errorf("initial: expected 'B' at (0,0), got %q", c)
	}
	if bg := ta.LastScreen.Get(5, 5).Background; bg != "#000000" {
		t.Errorf("initial: expected bg '#000000' at (5,5), got %q", bg)
	}

	// Add dialog
	app.RegisterComponent("dlg", "dialog", buffer.Rect{X: 5, Y: 3, W: 10, H: 4}, 100,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#FFFFFF"
			child := layout.NewVNode("text")
			child.Content = "D"
			root.AddChild(child)
			return root
		})
	app.RenderAll()

	if c := ta.LastScreen.Get(5, 3).Char; c != 'D' {
		t.Errorf("after add: expected 'D' at (5,3), got %q", c)
	}

	// Remove dialog
	app.UnregisterComponent("dlg")
	app.RenderAll()

	// After removal, (5,3) should show background space (bg fills with ' ')
	cell := ta.LastScreen.Get(5, 3)
	if cell.Background != "#000000" {
		t.Errorf("after remove: expected bg '#000000' at (5,3), got %q", cell.Background)
	}
}

func TestIntegration_Compositing_SameZIndex(t *testing.T) {
	app, ta := NewTestApp(20, 5)

	// Both at z=50, overlapping at x=5..9
	app.RegisterComponent("A", "A", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 50,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#111111"
			child := layout.NewVNode("text")
			child.Content = "A"
			root.AddChild(child)
			return root
		})

	app.RegisterComponent("B", "B", buffer.Rect{X: 5, Y: 0, W: 10, H: 5}, 50,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#222222"
			child := layout.NewVNode("text")
			child.Content = "B"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Non-overlapping: A's text at (0,0)
	if c := ta.LastScreen.Get(0, 0).Char; c != 'A' {
		t.Errorf("expected 'A' at (0,0), got %q", c)
	}

	// Non-overlapping: B's text at (5,0) — B's origin is (5,0)
	if c := ta.LastScreen.Get(5, 0).Char; c != 'B' {
		t.Errorf("expected 'B' at (5,0), got %q", c)
	}

	// B's non-text area at (14,0) should have B's background
	cell := ta.LastScreen.Get(14, 0)
	if cell.Background != "#222222" {
		t.Errorf("expected B's bg at (14,0), got %q", cell.Background)
	}

	// Overlap area (6,1): one of the two backgrounds should be visible
	overlapCell := ta.LastScreen.Get(6, 1)
	if overlapCell.Background != "#111111" && overlapCell.Background != "#222222" {
		t.Errorf("overlap (6,1): expected one of the backgrounds, got %q", overlapCell.Background)
	}
}

func TestIntegration_Compositing_TransparentCells(t *testing.T) {
	app, ta := NewTestApp(20, 10)

	// bg fills everything with background color
	app.RegisterComponent("bg", "bg", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#000000"
			child := layout.NewVNode("text")
			child.Content = "B"
			root.AddChild(child)
			return root
		})

	// Overlay only renders a single char at its first position — rest is transparent (zero cells)
	app.RegisterComponent("overlay", "overlay", buffer.Rect{X: 5, Y: 3, W: 10, H: 5}, 100,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "X"
			return vn
		})

	app.RenderAll()

	// Overlay first cell should be 'X'
	if c := ta.LastScreen.Get(5, 3).Char; c != 'X' {
		t.Errorf("overlay origin: expected 'X', got %q", c)
	}

	// Interior of overlay where no content was rendered: background shows through.
	// The overlay only renders 'X' at (5,3). Other cells in overlay area are zero cells
	// (transparent) and the background space cell (bg fill with ' ') should show through.
	cell := ta.LastScreen.Get(10, 5)
	if cell.Background != "#000000" {
		t.Errorf("transparent area: expected bg '#000000' (bg shows through), got bg=%q char=%q", cell.Background, cell.Char)
	}
}

// =============================================================================
// CATEGORY 2: Event Dispatch Complex Scenarios
// =============================================================================

func TestIntegration_Events_ClickOnOccludedArea(t *testing.T) {
	app, _ := NewTestApp(20, 10)

	var log []string

	// bg at z=0 with click handler on the root box
	app.RegisterComponent("bg", "bg", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "bg-root"
			root.Style.Background = "#000000"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				log = append(log, "bg")
			})
			child := layout.NewVNode("text")
			child.Content = "B"
			root.AddChild(child)
			return root
		})

	// dialog at z=100 on top — click handler on root box
	app.RegisterComponent("dlg", "dialog", buffer.Rect{X: 5, Y: 3, W: 10, H: 4}, 100,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "dlg-root"
			root.Style.Background = "#FFFFFF"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				log = append(log, "dlg")
			})
			child := layout.NewVNode("text")
			child.Content = "D"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Click inside dialog area — only dialog handler should fire (bg is occluded)
	app.HandleEvent(&event.Event{Type: "mousedown", X: 7, Y: 4})

	if len(log) != 1 || log[0] != "dlg" {
		t.Errorf("expected only 'dlg' click, got %v", log)
	}

	// Click outside dialog (in bg area) — only bg handler should fire
	log = nil
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	if len(log) != 1 || log[0] != "bg" {
		t.Errorf("expected only 'bg' click outside dialog, got %v", log)
	}
}

func TestIntegration_Events_Bubbling(t *testing.T) {
	app, _ := NewTestApp(20, 10)

	var log []string

	app.RegisterComponent("comp", "comp", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "parent"
			root.Style.Background = "#000000"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				log = append(log, "parent")
			})

			child := layout.NewVNode("text")
			child.ID = "child"
			child.Content = "C"
			child.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				log = append(log, "child")
			})
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Click on child with bubbling enabled
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0, Bubbles: true})

	// With bubbling: child fires first, then parent
	childFired := false
	parentFired := false
	for _, entry := range log {
		if entry == "child" {
			childFired = true
		}
		if entry == "parent" {
			parentFired = true
		}
	}
	if !childFired {
		t.Error("expected child handler to fire")
	}
	if !parentFired {
		t.Error("expected parent handler to fire (bubbling)")
	}
}

func TestIntegration_Events_HoverAcrossComponents(t *testing.T) {
	app, _ := NewTestApp(20, 5)

	hoverA := false
	hoverB := false

	// Put hover handlers on the root box (which spans the whole component area)
	app.RegisterComponent("compA", "compA", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "A-root"
			root.Style.Background = "#111111"
			root.Props["onMouseEnter"] = event.EventHandler(func(e *event.Event) {
				hoverA = true
			})
			root.Props["onMouseLeave"] = event.EventHandler(func(e *event.Event) {
				hoverA = false
			})
			child := layout.NewVNode("text")
			child.Content = "A"
			root.AddChild(child)
			return root
		})

	app.RegisterComponent("compB", "compB", buffer.Rect{X: 10, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "B-root"
			root.Style.Background = "#222222"
			root.Props["onMouseEnter"] = event.EventHandler(func(e *event.Event) {
				hoverB = true
			})
			root.Props["onMouseLeave"] = event.EventHandler(func(e *event.Event) {
				hoverB = false
			})
			child := layout.NewVNode("text")
			child.Content = "B"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Hover over compA (interior area)
	app.HandleEvent(&event.Event{Type: "mousemove", X: 2, Y: 2})
	if !hoverA {
		t.Error("expected hoverA=true after mousemove to compA")
	}
	if hoverB {
		t.Error("expected hoverB=false")
	}

	// Move to compB
	app.HandleEvent(&event.Event{Type: "mousemove", X: 12, Y: 2})
	if hoverA {
		t.Error("expected hoverA=false after leaving compA")
	}
	if !hoverB {
		t.Error("expected hoverB=true after mousemove to compB")
	}

	// Move back to compA
	app.HandleEvent(&event.Event{Type: "mousemove", X: 2, Y: 2})
	if !hoverA {
		t.Error("expected hoverA=true after returning to compA")
	}
	if hoverB {
		t.Error("expected hoverB=false after leaving compB")
	}
}

func TestIntegration_Events_FocusTabAndShiftTab(t *testing.T) {
	app, _ := NewTestApp(40, 20)

	app.RegisterComponent("form1", "form1", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("vbox")
			root.ID = "form1-root"

			f1 := layout.NewVNode("text")
			f1.ID = "field-1"
			f1.Content = "F1"
			f1.Props["focusable"] = true

			f2 := layout.NewVNode("text")
			f2.ID = "field-2"
			f2.Content = "F2"
			f2.Props["focusable"] = true

			root.AddChild(f1)
			root.AddChild(f2)
			return root
		})

	app.RegisterComponent("form2", "form2", buffer.Rect{X: 20, Y: 0, W: 20, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("vbox")
			root.ID = "form2-root"

			f3 := layout.NewVNode("text")
			f3.ID = "field-3"
			f3.Content = "F3"
			f3.Props["focusable"] = true

			f4 := layout.NewVNode("text")
			f4.ID = "field-4"
			f4.Content = "F4"
			f4.Props["focusable"] = true

			root.AddChild(f3)
			root.AddChild(f4)
			return root
		})

	app.RenderAll()

	// Tab through all 4 focusables
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	first := app.FocusedID()
	if first == "" {
		t.Fatal("expected focus after first Tab, got empty")
	}

	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	second := app.FocusedID()

	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	third := app.FocusedID()

	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	fourth := app.FocusedID()

	// All should be different
	ids := map[string]bool{first: true, second: true, third: true, fourth: true}
	if len(ids) != 4 {
		t.Errorf("expected 4 unique focused IDs, got: %q, %q, %q, %q", first, second, third, fourth)
	}

	// Wrap around
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	if got := app.FocusedID(); got != first {
		t.Errorf("expected wrap to %q, got %q", first, got)
	}

	// Shift+Tab goes backwards
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Shift+Tab"})
	if got := app.FocusedID(); got != fourth {
		t.Errorf("expected Shift+Tab to go to %q, got %q", fourth, got)
	}
}

func TestIntegration_Events_DispatchAfterUnregister(t *testing.T) {
	app, _ := NewTestApp(20, 10)

	clicked := false

	app.RegisterComponent("btn", "btn", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.ID = "btn-1"
			vn.Content = "Click"
			vn.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clicked = true
			})
			return vn
		})

	app.RenderAll()

	// Verify handler works
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	if !clicked {
		t.Fatal("handler should fire before unregister")
	}

	// Unregister and re-render to clear handlers
	clicked = false
	app.UnregisterComponent("btn")
	app.RenderAll()

	// Click at old position — should NOT fire
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	if clicked {
		t.Error("stale handler fired after unregister + RenderAll")
	}
}

// =============================================================================
// CATEGORY 3: Reconciliation Integration
// =============================================================================

func TestIntegration_Reconcile_AddChildren(t *testing.T) {
	app, ta := NewTestApp(30, 10)

	app.RegisterComponent("parent", "parent", buffer.Rect{X: 0, Y: 0, W: 30, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("hbox")
			root.ID = "parent-root"
			root.Style.Background = "#000000"

			items, _ := state["items"].([]string)
			if items == nil {
				items = []string{"a", "b"}
			}
			for _, item := range items {
				child := layout.NewVNode("text")
				child.ID = "item-" + item
				child.Content = item
				root.AddChild(child)
			}
			return root
		})

	app.RenderAll()

	// Initial: 'a' and 'b' visible
	if c := ta.LastScreen.Get(0, 0).Char; c != 'a' {
		t.Errorf("initial: expected 'a' at (0,0), got %q", c)
	}

	// Add item 'c'
	app.SetState("parent", "items", []string{"a", "b", "c"})
	app.RenderDirty()

	// Verify 'c' appears somewhere on screen
	found := false
	for y := 0; y < ta.LastScreen.Height(); y++ {
		for x := 0; x < ta.LastScreen.Width(); x++ {
			if ta.LastScreen.Get(x, y).Char == 'c' {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("after adding 'c': expected 'c' to appear on screen")
	}
}

func TestIntegration_Reconcile_RemoveChildren(t *testing.T) {
	app, ta := NewTestApp(30, 10)

	app.RegisterComponent("parent", "parent", buffer.Rect{X: 0, Y: 0, W: 30, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("hbox")
			root.ID = "parent-root"
			root.Style.Background = "#000000"

			items, _ := state["items"].([]string)
			if items == nil {
				items = []string{"a", "b", "c"}
			}
			for _, item := range items {
				child := layout.NewVNode("text")
				child.ID = "item-" + item
				child.Content = item
				root.AddChild(child)
			}
			return root
		})

	app.RenderAll()

	// Verify 'b' is present
	foundB := false
	for y := 0; y < ta.LastScreen.Height(); y++ {
		for x := 0; x < ta.LastScreen.Width(); x++ {
			if ta.LastScreen.Get(x, y).Char == 'b' {
				foundB = true
				break
			}
		}
		if foundB {
			break
		}
	}
	if !foundB {
		t.Error("initial: expected 'b' on screen")
	}

	// Remove 'b'
	app.SetState("parent", "items", []string{"a", "c"})
	app.RenderDirty()

	// 'b' should no longer be on screen
	for y := 0; y < ta.LastScreen.Height(); y++ {
		for x := 0; x < ta.LastScreen.Width(); x++ {
			if ta.LastScreen.Get(x, y).Char == 'b' {
				t.Errorf("after remove: 'b' still visible at (%d,%d)", x, y)
				return
			}
		}
	}
}

func TestIntegration_Reconcile_RenderCountTracking(t *testing.T) {
	app, _ := NewTestApp(30, 10)

	renderCounts := map[string]int{}

	app.RegisterComponent("parent", "parent", buffer.Rect{X: 0, Y: 0, W: 30, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			renderCounts["parent"]++
			root := layout.NewVNode("hbox")
			root.ID = "parent-root"
			root.Style.Background = "#000000"

			child1 := layout.NewVNode("text")
			child1.ID = "c1"
			child1.Content = "A"
			root.AddChild(child1)

			child2 := layout.NewVNode("text")
			child2.ID = "c2"
			child2.Content = "B"
			root.AddChild(child2)

			return root
		})

	app.RenderAll()
	if renderCounts["parent"] != 1 {
		t.Errorf("expected 1 render, got %d", renderCounts["parent"])
	}

	// SetState makes it dirty → re-rendered
	app.SetState("parent", "x", 1)
	app.RenderDirty()
	if renderCounts["parent"] != 2 {
		t.Errorf("expected 2 renders after SetState, got %d", renderCounts["parent"])
	}

	// No state change → RenderDirty should NOT re-render
	app.RenderDirty()
	if renderCounts["parent"] != 2 {
		t.Errorf("expected still 2 renders (no dirty), got %d", renderCounts["parent"])
	}
}

func TestIntegration_Reconcile_ChildStatePreserved(t *testing.T) {
	// Tests that reconciliation via Manager preserves child state when key matches
	app, _ := NewTestApp(30, 10)

	parent := app.RegisterComponent("parent", "parent", buffer.Rect{X: 0, Y: 0, W: 30, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("hbox")
			root.ID = "parent-root"
			root.Style.Background = "#000000"
			child := layout.NewVNode("text")
			child.Content = "P"
			root.AddChild(child)
			return root
		})

	// Create children via Reconcile
	childRenderCount := 0
	children := []component.ChildDescriptor{
		{
			Key:  "child-a",
			Name: "childA",
			Props: map[string]any{"label": "A"},
			RenderFn: func(state, props map[string]any) *layout.VNode {
				childRenderCount++
				vn := layout.NewVNode("text")
				vn.Content = "A"
				return vn
			},
		},
		{
			Key:  "child-b",
			Name: "childB",
			Props: map[string]any{"label": "B"},
			RenderFn: func(state, props map[string]any) *layout.VNode {
				vn := layout.NewVNode("text")
				vn.Content = "B"
				return vn
			},
		},
	}

	app.manager.Reconcile(parent, children)

	// Set state on child-a
	childAID := parent.ID + ":child-a"
	app.SetState(childAID, "counter", 42)

	// Reconcile again with same keys — child state should be preserved
	app.manager.Reconcile(parent, children)

	childA := app.manager.Get(childAID)
	if childA == nil {
		t.Fatal("child-a should still exist after reconcile")
	}
	if childA.State["counter"] != 42 {
		t.Errorf("expected child-a state[counter]=42, got %v", childA.State["counter"])
	}

	_ = childRenderCount
}

// =============================================================================
// CATEGORY 4: Render Cycle Correctness
// =============================================================================

func TestIntegration_RenderCycle_BatchedSetState(t *testing.T) {
	app, ta := NewTestApp(20, 5)

	app.RegisterComponent("comp", "comp", buffer.Rect{X: 0, Y: 0, W: 20, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			x, _ := state["x"].(int)
			y, _ := state["y"].(int)
			vn := layout.NewVNode("text")
			vn.Content = fmt.Sprintf("%d,%d", x, y)
			return vn
		})

	app.RenderAll()

	// Multiple SetState calls before RenderDirty
	app.SetState("comp", "x", 1)
	app.SetState("comp", "y", 2)
	app.RenderDirty()

	// Should show "1,2"
	if c := ta.LastScreen.Get(0, 0).Char; c != '1' {
		t.Errorf("expected '1' at (0,0), got %q", c)
	}
	if c := ta.LastScreen.Get(2, 0).Char; c != '2' {
		t.Errorf("expected '2' at (2,0), got %q", c)
	}
}

func TestIntegration_RenderCycle_NoDirtyNoOp(t *testing.T) {
	app, ta := NewTestApp(10, 5)

	app.RegisterComponent("c", "c", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "X"
			return vn
		})

	app.RenderAll()
	writesBefore := ta.WriteCount

	// Nothing dirty
	app.RenderDirty()

	if ta.WriteCount != writesBefore {
		t.Errorf("expected no new writes, WriteCount went from %d to %d", writesBefore, ta.WriteCount)
	}
	if rects := app.DirtyRects(); len(rects) != 0 {
		t.Errorf("expected no dirty rects, got %d", len(rects))
	}
}

func TestIntegration_RenderCycle_RapidStateChanges(t *testing.T) {
	app, ta := NewTestApp(10, 1)

	app.RegisterComponent("counter", "counter", buffer.Rect{X: 0, Y: 0, W: 10, H: 1}, 0,
		func(state, props map[string]any) *layout.VNode {
			count := 0
			if c, ok := state["count"].(int); ok {
				count = c
			}
			vn := layout.NewVNode("text")
			vn.Content = fmt.Sprintf("%d", count)
			return vn
		})

	app.RenderAll()

	// 100 rapid updates
	for i := 1; i <= 100; i++ {
		app.SetState("counter", "count", i)
		app.RenderDirty()
	}

	// Final screen should show "100"
	s := ""
	for x := 0; x < ta.LastScreen.Width(); x++ {
		c := ta.LastScreen.Get(x, 0).Char
		if c == 0 {
			break
		}
		s += string(c)
	}
	if s != "100" {
		t.Errorf("after 100 updates: expected '100' on screen, got %q", s)
	}
}

func TestIntegration_RenderCycle_NilVNodeTree(t *testing.T) {
	app, ta := NewTestApp(10, 5)

	app.RegisterComponent("nil-comp", "nil", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			return nil
		})

	// Should not panic
	app.RenderAll()
	app.RenderDirty()

	// Screen should be empty (all zero cells)
	for y := 0; y < ta.LastScreen.Height(); y++ {
		for x := 0; x < ta.LastScreen.Width(); x++ {
			if c := ta.LastScreen.Get(x, y); !c.Zero() {
				t.Errorf("expected zero cell at (%d,%d), got char=%q bg=%q", x, y, c.Char, c.Background)
				return
			}
		}
	}
}

func TestIntegration_RenderCycle_ResizeThenRenderDirty(t *testing.T) {
	app, ta := NewTestApp(10, 5)

	app.RegisterComponent("c", "c", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "X"
			return vn
		})

	app.RenderAll()
	if c := ta.LastScreen.Get(0, 0).Char; c != 'X' {
		t.Fatalf("initial: expected 'X', got %q", c)
	}

	// Resize and use RenderDirty (not RenderAll)
	app.Resize(20, 10)
	app.RenderDirty()

	if ta.LastScreen.Width() != 20 || ta.LastScreen.Height() != 10 {
		t.Errorf("after resize: expected 20x10, got %dx%d", ta.LastScreen.Width(), ta.LastScreen.Height())
	}
	// Component should still render 'X'
	if c := ta.LastScreen.Get(0, 0).Char; c != 'X' {
		t.Errorf("after resize: expected 'X' at (0,0), got %q", c)
	}
}

// =============================================================================
// CATEGORY 5: Stress / Scale Tests
// =============================================================================

func TestIntegration_Stress_ManyCellComponents(t *testing.T) {
	app, ta := NewTestApp(50, 50)

	// Register 100 1x1 components at various positions
	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("cell-%d", i)
		x := i % 50
		y := i / 50
		char := rune('A' + (i % 26))
		app.RegisterComponent(id, "cell", buffer.Rect{X: x, Y: y, W: 1, H: 1}, i,
			func(state, props map[string]any) *layout.VNode {
				vn := layout.NewVNode("text")
				vn.Content = string(char)
				return vn
			})
	}

	app.RenderAll()

	// Verify some cells are rendered
	if c := ta.LastScreen.Get(0, 0).Char; c == 0 {
		t.Error("expected non-zero cell at (0,0)")
	}

	// Change state of one component → RenderDirty should produce minimal dirty rects
	app.SetState("cell-50", "highlight", true)
	app.RenderDirty()

	// Should have produced dirty rects (not zero)
	// (We can't assert exact count without knowing the compositor internals,
	// but we can verify it didn't crash and produced output)
	if ta.WriteCount < 2 {
		t.Errorf("expected at least 2 writes (initial + dirty), got %d", ta.WriteCount)
	}
}

func TestIntegration_Stress_RapidHoverAcrossComponents(t *testing.T) {
	app, _ := NewTestApp(20, 1)

	hoverState := make(map[string]bool)

	// 20 components in a row (1x1 each)
	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("cell-%d", i)
		app.RegisterComponent(id, "cell", buffer.Rect{X: i, Y: 0, W: 1, H: 1}, 0,
			func(state, props map[string]any) *layout.VNode {
				vn := layout.NewVNode("text")
				vn.ID = id
				vn.Content = "·"
				vn.Props["onMouseEnter"] = event.EventHandler(func(e *event.Event) {
					hoverState[e.Target] = true
				})
				vn.Props["onMouseLeave"] = event.EventHandler(func(e *event.Event) {
					hoverState[e.Target] = false
				})
				return vn
			})
	}

	app.RenderAll()

	// Sweep mouse across all 20 cells
	for x := 0; x < 20; x++ {
		app.HandleEvent(&event.Event{Type: "mousemove", X: x, Y: 0})
		app.RenderDirty()
	}

	// After sweeping to position 19, only cell-19 should be hovered
	// (all previous should have received mouseleave)
	hoveredCount := 0
	for _, v := range hoverState {
		if v {
			hoveredCount++
		}
	}
	// At most 1 should be hovered (the last one)
	if hoveredCount > 1 {
		t.Errorf("expected at most 1 hovered cell, got %d", hoveredCount)
	}
}

func TestIntegration_Stress_DeepVNodeTree(t *testing.T) {
	app, ta := NewTestApp(20, 10)

	var deepestClicked bool

	app.RegisterComponent("deep", "deep", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			// 5-level nested tree: box > vbox > hbox > box > text
			level1 := layout.NewVNode("box")
			level1.ID = "L1"
			level1.Style.Background = "#000000"

			level2 := layout.NewVNode("vbox")
			level2.ID = "L2"
			level1.AddChild(level2)

			level3 := layout.NewVNode("hbox")
			level3.ID = "L3"
			level2.AddChild(level3)

			level4 := layout.NewVNode("box")
			level4.ID = "L4"
			level3.AddChild(level4)

			level5 := layout.NewVNode("text")
			level5.ID = "L5-text"
			level5.Content = "Z"
			level5.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				deepestClicked = true
			})
			level4.AddChild(level5)

			return level1
		})

	app.RenderAll()

	// Verify 'Z' is rendered
	found := false
	for y := 0; y < ta.LastScreen.Height(); y++ {
		for x := 0; x < ta.LastScreen.Width(); x++ {
			if ta.LastScreen.Get(x, y).Char == 'Z' {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("expected 'Z' from deep nested tree to appear on screen")
	}

	// Click at origin — should hit the deepest node with an ID
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	if !deepestClicked {
		t.Error("expected deepest VNode (L5-text) click handler to fire")
	}
}

// =============================================================================
// CATEGORY 6: Cross-Module Edge Cases
// =============================================================================

func TestIntegration_CrossModule_BufferResizeOnMove(t *testing.T) {
	app, ta := NewTestApp(30, 15)

	app.RegisterComponent("comp", "comp", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#111111"
			child := layout.NewVNode("text")
			child.Content = "A"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Verify initial render
	if c := ta.LastScreen.Get(0, 0).Char; c != 'A' {
		t.Errorf("initial: expected 'A' at (0,0), got %q", c)
	}

	// Move to larger rect
	app.MoveComponent("comp", buffer.Rect{X: 0, Y: 0, W: 20, H: 10})
	app.RenderDirty()

	// Content should still render correctly at the new size
	if c := ta.LastScreen.Get(0, 0).Char; c != 'A' {
		t.Errorf("after resize move: expected 'A' at (0,0), got %q", c)
	}

	// Verify the component's buffer was resized
	comp := app.manager.Get("comp")
	if comp == nil {
		t.Fatal("component not found")
	}
	if comp.Buffer.Width() != 20 || comp.Buffer.Height() != 10 {
		t.Errorf("expected buffer 20x10, got %dx%d", comp.Buffer.Width(), comp.Buffer.Height())
	}
}

func TestIntegration_CrossModule_OverlappingDirtyRects(t *testing.T) {
	app, ta := NewTestApp(30, 15)

	// 3 overlapping components
	app.RegisterComponent("c1", "c1", buffer.Rect{X: 0, Y: 0, W: 15, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#111111"
			child := layout.NewVNode("text")
			child.Content = "1"
			root.AddChild(child)
			return root
		})

	app.RegisterComponent("c2", "c2", buffer.Rect{X: 5, Y: 3, W: 15, H: 10}, 50,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#222222"
			child := layout.NewVNode("text")
			child.Content = "2"
			root.AddChild(child)
			return root
		})

	app.RegisterComponent("c3", "c3", buffer.Rect{X: 10, Y: 5, W: 15, H: 10}, 100,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#333333"
			child := layout.NewVNode("text")
			child.Content = "3"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Verify correct z-order at overlap point
	if c := ta.LastScreen.Get(10, 5).Char; c != '3' {
		t.Errorf("initial overlap: expected '3' at (10,5), got %q", c)
	}

	// Make all 3 dirty simultaneously
	app.SetState("c1", "x", 1)
	app.SetState("c2", "x", 1)
	app.SetState("c3", "x", 1)
	app.RenderDirty()

	// Screen should still show correct compositing
	if c := ta.LastScreen.Get(0, 0).Char; c != '1' {
		t.Errorf("after dirty: expected '1' at (0,0), got %q", c)
	}
	if c := ta.LastScreen.Get(5, 3).Char; c != '2' {
		t.Errorf("after dirty: expected '2' at (5,3), got %q", c)
	}
	if c := ta.LastScreen.Get(10, 5).Char; c != '3' {
		t.Errorf("after dirty: expected '3' at (10,5), got %q", c)
	}
}

func TestIntegration_CrossModule_FullLifecycle(t *testing.T) {
	app, ta := NewTestApp(20, 10)

	clicked := false
	count := 0

	// 1. Register
	app.RegisterComponent("comp", "comp", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			c, _ := state["count"].(int)
			root := layout.NewVNode("box")
			root.ID = "root"
			root.Style.Background = "#000000"
			child := layout.NewVNode("text")
			child.ID = "btn"
			child.Content = fmt.Sprintf("N:%d", c)
			child.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clicked = true
			})
			child.Props["focusable"] = true
			root.AddChild(child)
			return root
		})

	// 2. RenderAll
	app.RenderAll()
	if c := ta.LastScreen.Get(0, 0).Char; c != 'N' {
		t.Fatalf("step 2: expected 'N' at (0,0), got %q", c)
	}

	// 3. HandleEvent (click)
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	if !clicked {
		t.Error("step 3: click handler should have fired")
	}

	// 4. SetState + RenderDirty
	count = 5
	app.SetState("comp", "count", count)
	app.RenderDirty()
	if c := ta.LastScreen.Get(2, 0).Char; c != '5' {
		t.Errorf("step 4: expected '5' at (2,0), got %q", c)
	}

	// 5. MoveComponent + RenderDirty
	app.MoveComponent("comp", buffer.Rect{X: 5, Y: 2, W: 15, H: 8})
	app.RenderDirty()
	if c := ta.LastScreen.Get(5, 2).Char; c != 'N' {
		t.Errorf("step 5: expected 'N' at (5,2) after move, got %q", c)
	}

	// 6. Resize + RenderAll
	app.Resize(30, 15)
	app.RenderAll()
	if ta.LastScreen.Width() != 30 || ta.LastScreen.Height() != 15 {
		t.Errorf("step 6: expected 30x15, got %dx%d", ta.LastScreen.Width(), ta.LastScreen.Height())
	}

	// 7. Unregister + RenderAll
	app.UnregisterComponent("comp")
	app.RenderAll()

	// Screen should be empty
	allEmpty := true
	for y := 0; y < ta.LastScreen.Height(); y++ {
		for x := 0; x < ta.LastScreen.Width(); x++ {
			if !ta.LastScreen.Get(x, y).Zero() {
				allEmpty = false
				break
			}
		}
		if !allEmpty {
			break
		}
	}
	if !allEmpty {
		t.Error("step 7: screen should be empty after unregister")
	}
}

// Suppress unused import warnings
var _ = (*component.Component)(nil)
