package v2

import (
	"fmt"
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// --- Test 1: Basic render pipeline ---

func TestApp_RenderAll(t *testing.T) {
	app, ta := NewTestApp(20, 10)

	app.RegisterComponent("comp1", "box", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#FF0000"
			child := layout.NewVNode("text")
			child.Content = "Hello"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	if ta.WriteCount != 1 {
		t.Fatalf("expected WriteCount=1, got %d", ta.WriteCount)
	}
	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}
	// The text "Hello" should be painted starting at some position.
	// Check that at least one cell has 'H'.
	found := false
	for y := 0; y < ta.LastScreen.Height(); y++ {
		for x := 0; x < ta.LastScreen.Width(); x++ {
			if ta.LastScreen.Get(x, y).Char == 'H' {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("expected 'H' from 'Hello' to appear on screen")
	}
}

// --- Test 2: Cell hover scenario ---

func TestScenario_CellHover(t *testing.T) {
	app, ta := NewTestApp(10, 1)

	app.RegisterComponent("cell-0", "cell", buffer.Rect{X: 0, Y: 0, W: 10, H: 1}, 0,
		func(state, props map[string]any) *layout.VNode {
			char := "·"
			if hovered, ok := state["hovered"].(bool); ok && hovered {
				char = "█"
			}
			vn := layout.NewVNode("text")
			vn.ID = "cell-0"
			vn.Content = char
			vn.Props["onMouseEnter"] = event.EventHandler(func(e *event.Event) {
				app.SetState("cell-0", "hovered", true)
			})
			vn.Props["onMouseLeave"] = event.EventHandler(func(e *event.Event) {
				app.SetState("cell-0", "hovered", false)
			})
			return vn
		})

	app.RenderAll()

	// Before hover — should show '·'
	cell := ta.LastScreen.Get(0, 0)
	if cell.Char != '·' {
		t.Errorf("before hover: expected '·', got %q", cell.Char)
	}

	// Hover over cell-0
	app.HandleEvent(&event.Event{Type: "mousemove", X: 0, Y: 0})
	app.RenderDirty()

	cell = ta.LastScreen.Get(0, 0)
	if cell.Char != '█' {
		t.Errorf("after hover: expected '█', got %q (%d)", cell.Char, cell.Char)
	}

	// Move away
	app.HandleEvent(&event.Event{Type: "mousemove", X: 9, Y: 0})
	app.RenderDirty()

	cell = ta.LastScreen.Get(0, 0)
	if cell.Char != '·' {
		t.Errorf("after leave: expected '·', got %q", cell.Char)
	}
}

// --- Test 3: Dialog occludes grid ---

func TestScenario_DialogOccludesGrid(t *testing.T) {
	app, ta := NewTestApp(40, 20)

	// Background fills with 'B'
	app.RegisterComponent("bg", "background", buffer.Rect{X: 0, Y: 0, W: 40, H: 20}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.ID = "bg"
			vn.Content = "B"
			return vn
		})

	// Dialog at higher z-index
	app.RegisterComponent("dlg1", "dialog", buffer.Rect{X: 10, Y: 5, W: 20, H: 10}, 100,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.ID = "dlg1-root"
			vn.Content = "D"
			return vn
		})

	app.RenderAll()

	// Background area outside dialog should show 'B'
	if c := ta.LastScreen.Get(0, 0).Char; c != 'B' {
		t.Errorf("expected 'B' at (0,0), got %q", c)
	}

	// Dialog area should show 'D'
	if c := ta.LastScreen.Get(10, 5).Char; c != 'D' {
		t.Errorf("expected 'D' at (10,5), got %q", c)
	}
}

// --- Test 4: Window move reveals background ---

func TestScenario_WindowMove(t *testing.T) {
	app, ta := NewTestApp(40, 20)

	// Background fills with 'B' using a box with background color so paint fills the area.
	app.RegisterComponent("bg", "bg", buffer.Rect{X: 0, Y: 0, W: 40, H: 20}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "bg"
			root.Style.Background = "#000000"
			child := layout.NewVNode("text")
			child.Content = "B"
			root.AddChild(child)
			return root
		})

	// Dialog at z=100
	app.RegisterComponent("dlg", "dlg", buffer.Rect{X: 5, Y: 5, W: 10, H: 5}, 100,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "dlg"
			root.Style.Background = "#FFFFFF"
			child := layout.NewVNode("text")
			child.Content = "D"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	if c := ta.LastScreen.Get(5, 5).Char; c == 0 {
		t.Errorf("before move: expected non-zero at (5,5), got zero")
	}

	// Move dialog away
	app.MoveComponent("dlg", buffer.Rect{X: 25, Y: 5, W: 10, H: 5})
	app.RenderDirty()

	// Old position should now show background (space with background color, or 'B' at 0,0)
	oldCell := ta.LastScreen.Get(5, 5)
	if oldCell.Char == 0 {
		t.Errorf("after move: expected background at (5,5), got zero cell")
	}
	if oldCell.Background != "#000000" {
		t.Errorf("after move: expected bg '#000000' at (5,5), got %q", oldCell.Background)
	}

	// New position should show dialog content ('D' at first cell, spaces with bg elsewhere)
	newCell := ta.LastScreen.Get(25, 5)
	if newCell.Char != 'D' {
		t.Errorf("after move: expected 'D' at (25,5), got %q", newCell.Char)
	}
	// Interior cell should have dialog background
	interiorCell := ta.LastScreen.Get(26, 5)
	if interiorCell.Background != "#FFFFFF" {
		t.Errorf("after move: expected bg '#FFFFFF' at (26,5), got %q", interiorCell.Background)
	}
}

// --- Test 5: Tab navigation ---

func TestScenario_TabNavigation(t *testing.T) {
	app, _ := NewTestApp(40, 20)

	app.RegisterComponent("form", "form", buffer.Rect{X: 0, Y: 0, W: 40, H: 20}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("vbox")
			root.ID = "form-root"

			email := layout.NewVNode("text")
			email.ID = "email"
			email.Content = "E"
			email.Props["focusable"] = true

			password := layout.NewVNode("text")
			password.ID = "password"
			password.Content = "P"
			password.Props["focusable"] = true

			ok := layout.NewVNode("text")
			ok.ID = "ok"
			ok.Content = "O"
			ok.Props["focusable"] = true

			root.AddChild(email)
			root.AddChild(password)
			root.AddChild(ok)
			return root
		})

	app.RenderAll()

	// Tab should cycle through focusables
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	if id := app.FocusedID(); id != "email" {
		t.Errorf("after 1st tab: expected 'email', got %q", id)
	}

	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	if id := app.FocusedID(); id != "password" {
		t.Errorf("after 2nd tab: expected 'password', got %q", id)
	}

	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	if id := app.FocusedID(); id != "ok" {
		t.Errorf("after 3rd tab: expected 'ok', got %q", id)
	}

	// Wrap around
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	if id := app.FocusedID(); id != "email" {
		t.Errorf("after 4th tab: expected 'email' (wrap), got %q", id)
	}
}

// --- Test 6: SetState triggers re-render ---

func TestApp_SetState_ReRender(t *testing.T) {
	app, ta := NewTestApp(10, 1)

	app.RegisterComponent("counter", "counter", buffer.Rect{X: 0, Y: 0, W: 10, H: 1}, 0,
		func(state, props map[string]any) *layout.VNode {
			count := 0
			if c, ok := state["count"].(int); ok {
				count = c
			}
			vn := layout.NewVNode("text")
			vn.Content = fmt.Sprintf("N:%d", count)
			return vn
		})

	app.RenderAll()

	// Initial: "N:0"
	if c := ta.LastScreen.Get(0, 0).Char; c != 'N' {
		t.Errorf("initial: expected 'N' at (0,0), got %q", c)
	}
	if c := ta.LastScreen.Get(2, 0).Char; c != '0' {
		t.Errorf("initial: expected '0' at (2,0), got %q", c)
	}

	app.SetState("counter", "count", 7)
	app.RenderDirty()

	// After state change: "N:7"
	if c := ta.LastScreen.Get(2, 0).Char; c != '7' {
		t.Errorf("after setState: expected '7' at (2,0), got %q", c)
	}
}

// --- Test 7: Multiple components z-order ---

func TestApp_MultipleComponents_ZOrder(t *testing.T) {
	app, ta := NewTestApp(10, 5)

	// Background z=0
	app.RegisterComponent("bg", "bg", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "B"
			return vn
		})

	// Foreground z=100, partial overlap
	app.RegisterComponent("fg", "fg", buffer.Rect{X: 3, Y: 1, W: 4, H: 3}, 100,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "F"
			return vn
		})

	app.RenderAll()

	// Background visible outside foreground
	if c := ta.LastScreen.Get(0, 0).Char; c != 'B' {
		t.Errorf("expected 'B' at (0,0), got %q", c)
	}

	// Foreground visible on top
	if c := ta.LastScreen.Get(3, 1).Char; c != 'F' {
		t.Errorf("expected 'F' at (3,1), got %q", c)
	}
}

// --- Test 8: Click dispatches to correct target ---

func TestScenario_ClickDispatch(t *testing.T) {
	app, _ := NewTestApp(20, 10)

	clicked := ""

	app.RegisterComponent("btn", "btn", buffer.Rect{X: 5, Y: 3, W: 5, H: 1}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.ID = "btn-1"
			vn.Content = "Click"
			vn.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clicked = "btn-1"
			})
			return vn
		})

	app.RenderAll()

	// Click in the button area
	app.HandleEvent(&event.Event{Type: "mousedown", X: 5, Y: 3})
	if clicked != "btn-1" {
		t.Errorf("expected click on 'btn-1', got %q", clicked)
	}
}

// --- Test 9: Resize resets compositor ---

func TestApp_Resize(t *testing.T) {
	app, ta := NewTestApp(10, 5)

	app.RegisterComponent("c", "c", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "X"
			return vn
		})

	app.RenderAll()
	if ta.LastScreen.Width() != 10 || ta.LastScreen.Height() != 5 {
		t.Errorf("expected 10x5, got %dx%d", ta.LastScreen.Width(), ta.LastScreen.Height())
	}

	// Resize
	app.Resize(20, 10)
	app.RenderAll()
	if ta.LastScreen.Width() != 20 || ta.LastScreen.Height() != 10 {
		t.Errorf("after resize: expected 20x10, got %dx%d", ta.LastScreen.Width(), ta.LastScreen.Height())
	}
}

// --- Test 10: Unregister component removes from screen ---

func TestApp_UnregisterComponent(t *testing.T) {
	app, ta := NewTestApp(10, 5)

	app.RegisterComponent("c1", "c1", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "A"
			return vn
		})

	app.RenderAll()
	if c := ta.LastScreen.Get(0, 0).Char; c != 'A' {
		t.Errorf("expected 'A', got %q", c)
	}

	app.UnregisterComponent("c1")
	app.RenderAll()

	// Screen should be empty (no components)
	if c := ta.LastScreen.Get(0, 0).Char; c != 0 {
		t.Errorf("after unregister: expected zero, got %q", c)
	}
}

// --- Test 11: syncHandlers clears stale handlers (BUG-1) ---

func TestApp_SyncHandlers_ClearsStale(t *testing.T) {
	app, _ := NewTestApp(10, 5)

	clicked := ""
	app.RegisterComponent("c1", "c1", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.ID = "btn-stale"
			vn.Content = "X"
			vn.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clicked = "stale-handler-fired"
			})
			return vn
		})

	app.RenderAll()

	// Verify handler works initially
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	if clicked != "stale-handler-fired" {
		t.Fatalf("expected handler to fire initially, got %q", clicked)
	}

	// Unregister the component
	clicked = ""
	app.UnregisterComponent("c1")
	app.RenderAll()

	// Now click in the same area — stale handler should NOT fire
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	if clicked != "" {
		t.Errorf("stale handler fired after unregister: got %q", clicked)
	}
}

// --- Test 12: RenderFn returning nil doesn't panic (BUG-3) ---

func TestApp_NilRenderFn(t *testing.T) {
	app, _ := NewTestApp(10, 5)

	app.RegisterComponent("nil-comp", "nil", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			return nil // intentionally return nil
		})

	// Should not panic
	app.RenderAll()
	app.RenderDirty()
}

// --- Test 13: Resize marks components dirty (EDGE-1) ---

func TestApp_Resize_MarksDirty(t *testing.T) {
	app, ta := NewTestApp(10, 5)

	app.RegisterComponent("c", "c", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "Z"
			return vn
		})

	app.RenderAll()
	if c := ta.LastScreen.Get(0, 0).Char; c != 'Z' {
		t.Fatalf("initial: expected 'Z', got %q", c)
	}

	// Resize and use RenderDirty (not RenderAll) — should still repaint
	app.Resize(20, 10)
	app.RenderDirty()

	if ta.LastScreen == nil {
		t.Fatal("after resize+RenderDirty: LastScreen is nil")
	}
	if ta.LastScreen.Width() != 20 || ta.LastScreen.Height() != 10 {
		t.Errorf("after resize: expected 20x10, got %dx%d", ta.LastScreen.Width(), ta.LastScreen.Height())
	}
}

// --- Test 14: HandleEvent on empty app doesn't panic ---

func TestApp_HandleEvent_EmptyApp(t *testing.T) {
	app, _ := NewTestApp(10, 5)
	// No components registered — should not panic
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	app.HandleEvent(&event.Event{Type: "mousemove", X: 5, Y: 3})
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
}

// --- Test 15: RenderDirty with nothing dirty is a no-op ---

func TestApp_RenderDirty_NoDirty(t *testing.T) {
	app, ta := NewTestApp(10, 5)

	app.RegisterComponent("c", "c", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "A"
			return vn
		})

	app.RenderAll()
	writesBefore := ta.WriteCount

	// Nothing dirty — RenderDirty should be a no-op
	app.RenderDirty()

	if ta.WriteCount != writesBefore {
		t.Errorf("expected no new writes, but WriteCount went from %d to %d", writesBefore, ta.WriteCount)
	}
}

// --- Test 16: DevTools F12 toggle ---

func TestApp_DevTools_F12Toggle(t *testing.T) {
	app, _ := NewTestApp(80, 24)

	app.RegisterComponent("main", "Main", buffer.Rect{X: 0, Y: 0, W: 80, H: 24}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "Hello App"
			return vn
		})

	app.RenderAll()

	// DevTools should not be visible initially
	if app.DevTools().Visible {
		t.Fatal("devtools should be hidden initially")
	}
	if app.manager.Get("__devtools") != nil {
		t.Fatal("__devtools component should not exist initially")
	}

	// Press F12 to open
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F12"})

	if !app.DevTools().Visible {
		t.Fatal("devtools should be visible after F12")
	}
	dtComp := app.manager.Get("__devtools")
	if dtComp == nil {
		t.Fatal("__devtools component should be registered after F12")
	}
	if dtComp.ZIndex() != 9999 {
		t.Errorf("devtools zIndex = %d, want 9999", dtComp.ZIndex())
	}

	// Check that devtools panel takes bottom 40% of screen
	r := dtComp.Rect()
	expectedH := 24 * 4 / 10
	if r.H != expectedH {
		t.Errorf("devtools height = %d, want %d", r.H, expectedH)
	}
	if r.Y != 24-expectedH {
		t.Errorf("devtools Y = %d, want %d", r.Y, 24-expectedH)
	}
	if r.W != 80 {
		t.Errorf("devtools width = %d, want 80", r.W)
	}

	// Check the screen has devtools content
	screen := app.Screen()
	foundElements := false
	for y := 0; y < 24; y++ {
		line := ""
		for x := 0; x < 80; x++ {
			c := screen.Get(x, y)
			if c.Char != 0 {
				line += string(c.Char)
			}
		}
		if len(line) > 0 && (contains(line, "Elements") || contains(line, "Perf")) {
			foundElements = true
			break
		}
	}
	if !foundElements {
		t.Error("devtools panel content not found on screen after F12")
	}

	// Press F12 again to close
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F12"})

	if app.DevTools().Visible {
		t.Fatal("devtools should be hidden after second F12")
	}
	if app.manager.Get("__devtools") != nil {
		t.Fatal("__devtools component should be unregistered after second F12")
	}
}

// --- Test 17: DevTools tab switching ---

func TestApp_DevTools_TabSwitch(t *testing.T) {
	app, _ := NewTestApp(80, 24)

	app.RegisterComponent("main", "Main", buffer.Rect{X: 0, Y: 0, W: 80, H: 24}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "Hello"
			return vn
		})

	app.RenderAll()

	// Open devtools
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F12"})

	if app.DevTools().ActiveTab != 0 { // TabElements
		t.Errorf("default tab should be Elements (0), got %d", app.DevTools().ActiveTab)
	}

	// Switch to Perf tab
	app.HandleEvent(&event.Event{Type: "keydown", Key: "2"})
	if app.DevTools().ActiveTab != 1 { // TabPerf
		t.Errorf("tab should be Perf (1) after pressing 2, got %d", app.DevTools().ActiveTab)
	}

	// Switch back to Elements
	app.HandleEvent(&event.Event{Type: "keydown", Key: "1"})
	if app.DevTools().ActiveTab != 0 {
		t.Errorf("tab should be Elements (0) after pressing 1, got %d", app.DevTools().ActiveTab)
	}

	// Tab keys should NOT be intercepted when devtools is closed
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F12"}) // close
	app.DevTools().SetTab(0)                                     // reset

	// Now "2" should be dispatched normally (not intercepted)
	app.HandleEvent(&event.Event{Type: "keydown", Key: "2"})
	if app.DevTools().ActiveTab != 0 {
		t.Errorf("tab should still be Elements when devtools is closed, got %d", app.DevTools().ActiveTab)
	}
}

// --- Test 18: DevTools Elements tab shows component info ---

func TestApp_DevTools_ElementsContent(t *testing.T) {
	app, _ := NewTestApp(80, 24)

	app.RegisterComponent("counter", "Counter", buffer.Rect{X: 5, Y: 3, W: 30, H: 8}, 2,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "counter-root"
			child := layout.NewVNode("text")
			child.Content = "Count: 42"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Open devtools
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F12"})

	// The devtools component should have rendered with Elements tab
	dtComp := app.manager.Get("__devtools")
	if dtComp == nil {
		t.Fatal("__devtools not registered")
	}

	vn := dtComp.VNodeTree()
	if vn == nil {
		t.Fatal("devtools VNodeTree is nil")
	}

	// Walk the VNode tree to find component info
	allText := collectAllText(vn)
	if !contains(allText, "Counter") {
		t.Errorf("Elements tab should show component name 'Counter', got: %s", truncate(allText, 200))
	}
	if !contains(allText, "counter") {
		t.Errorf("Elements tab should show component ID 'counter', got: %s", truncate(allText, 200))
	}
}

// helper: contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// helper: collect all text content from a VNode tree
func collectAllText(vn *layout.VNode) string {
	if vn == nil {
		return ""
	}
	result := vn.Content
	for _, child := range vn.Children {
		result += " " + collectAllText(child)
	}
	return result
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// Suppress unused import warning
var _ = (*component.Component)(nil)
