package v2

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
	"github.com/akzj/lumina/pkg/lumina/v2/perf"
)

// --- Helpers ---

// readScreenLine reads up to maxLen rune characters from screen row y,
// stopping at the first zero-char cell.
func readScreenLine(ta *output.TestAdapter, y, maxLen int) string {
	var line []rune
	for x := 0; x < maxLen; x++ {
		c := ta.LastScreen.Get(x, y)
		if c.Char == 0 {
			break
		}
		line = append(line, c.Char)
	}
	return string(line)
}

// screenHasChar returns true if the screen contains the given rune anywhere.
func screenHasChar(ta *output.TestAdapter, r rune) bool {
	for y := 0; y < ta.LastScreen.Height(); y++ {
		for x := 0; x < ta.LastScreen.Width(); x++ {
			if ta.LastScreen.Get(x, y).Char == r {
				return true
			}
		}
	}
	return false
}

// screenHasString returns true if the string appears starting at some (x, y).
func screenHasString(ta *output.TestAdapter, s string) bool {
	runes := []rune(s)
	if len(runes) == 0 {
		return true
	}
	h := ta.LastScreen.Height()
	w := ta.LastScreen.Width()
	for y := 0; y < h; y++ {
		for x := 0; x <= w-len(runes); x++ {
			match := true
			for i, r := range runes {
				if ta.LastScreen.Get(x+i, y).Char != r {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}

// newLuaApp creates a NewAppWithLua with a fresh Lua state and TestAdapter.
func newLuaApp(t *testing.T, w, h int) (*App, *output.TestAdapter, *lua.State) {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	ta := output.NewTestAdapter()
	app := NewAppWithLua(L, w, h, ta)
	return app, ta, L
}

// --- Test 1: Simple text via createElement string child ---

func TestLuaE2E_SimpleText(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "txt",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("text", {}, "Hello World")
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}

	line := readScreenLine(ta, 0, 40)
	if line != "Hello World" {
		t.Errorf("expected 'Hello World' on screen, got %q", line)
	}
}

// --- Test 2: Box with background + text child ---

func TestLuaE2E_BoxWithText(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "boxtext",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("box", {style = {background = "#FF0000"}},
					lumina.createElement("text", {}, "Inside Box"))
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if !screenHasString(ta, "Inside Box") {
		t.Error("expected 'Inside Box' on screen")
	}

	// Verify background color is set on cells.
	// Check a cell beyond the text content where only background fill exists.
	cell := ta.LastScreen.Get(15, 1)
	if cell.Background != "#FF0000" {
		t.Errorf("expected background '#FF0000' at (15,1), got %q", cell.Background)
	}
}

// --- Test 3: Multiple text children (one per line) ---

func TestLuaE2E_MultipleTextChildren(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "multi",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("box", {},
					lumina.createElement("text", {}, "Line 1"),
					lumina.createElement("text", {}, "Line 2"),
					lumina.createElement("text", {}, "Line 3"))
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	for _, text := range []string{"Line 1", "Line 2", "Line 3"} {
		if !screenHasString(ta, text) {
			t.Errorf("expected %q on screen", text)
		}
	}
}

// --- Test 4: useState with SetState + re-render ---

func TestLuaE2E_UseState(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "stateful",
			x = 0, y = 0, w = 40, h = 3,
			render = function(state, props)
				local count, setCount = lumina.useState("count", 0)
				return lumina.createElement("text", {}, "N:" .. tostring(count))
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	// Initial render: "N:0"
	app.RenderAll()
	if !screenHasString(ta, "N:0") {
		t.Error("expected 'N:0' on initial render")
	}

	// Update state and re-render.
	app.SetState("stateful", "count", 42)
	app.RenderDirty()

	if !screenHasString(ta, "N:42") {
		t.Errorf("expected 'N:42' after SetState, got line: %q", readScreenLine(ta, 0, 40))
	}
}

// --- Test 5: Load actual counter.lua example ---

func TestLuaE2E_CounterScript(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunScript("../../../examples/v2/counter.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil after rendering counter.lua")
	}

	// counter.lua should produce visible text (at least "Count: 0").
	if !screenHasString(ta, "Count: 0") {
		// Fall back: at least some non-zero chars should be on screen.
		hasContent := false
		for y := 0; y < ta.LastScreen.Height(); y++ {
			for x := 0; x < ta.LastScreen.Width(); x++ {
				if ta.LastScreen.Get(x, y).Char != 0 {
					hasContent = true
					break
				}
			}
			if hasContent {
				break
			}
		}
		if !hasContent {
			t.Error("counter.lua rendered a blank screen — no visible content")
		}
	}
}

// --- Test 6: Click event triggers state change ---

func TestLuaE2E_ClickEvent(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "clicker",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				local count, setCount = lumina.useState("clicks", 0)
				return lumina.createElement("text", {
					id = "click-target",
					onClick = function()
						setCount(count + 1)
					end,
				}, "Clicks:" .. tostring(count))
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()
	if !screenHasString(ta, "Clicks:0") {
		t.Fatal("expected 'Clicks:0' on initial render")
	}

	// Click on the component area.
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	app.RenderDirty()

	if !screenHasString(ta, "Clicks:1") {
		t.Errorf("expected 'Clicks:1' after click, got line: %q", readScreenLine(ta, 0, 40))
	}
}

// --- Test 7: Multiple components at different positions ---

func TestLuaE2E_MultipleComponents(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "comp-a",
			x = 0, y = 0, w = 20, h = 5,
			render = function(state, props)
				return lumina.createElement("text", {}, "AAA")
			end
		})
		lumina.createComponent({
			id = "comp-b",
			x = 20, y = 0, w = 20, h = 5,
			render = function(state, props)
				return lumina.createElement("text", {}, "BBB")
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	// comp-a at x=0 should show "AAA"
	lineA := readScreenLine(ta, 0, 20)
	if lineA != "AAA" {
		t.Errorf("expected 'AAA' at x=0, got %q", lineA)
	}

	// comp-b at x=20 should show "BBB"
	var lineB []rune
	for x := 20; x < 40; x++ {
		c := ta.LastScreen.Get(x, 0)
		if c.Char == 0 {
			break
		}
		lineB = append(lineB, c.Char)
	}
	if string(lineB) != "BBB" {
		t.Errorf("expected 'BBB' at x=20, got %q", string(lineB))
	}
}

// --- Test 8: Deeply nested elements ---

func TestLuaE2E_NestedElements(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "nested",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("box", {},
					lumina.createElement("box", {},
						lumina.createElement("box", {},
							lumina.createElement("text", {}, "Deep"))))
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if !screenHasString(ta, "Deep") {
		t.Error("expected 'Deep' on screen from nested elements")
	}
}

// --- Test 9: Mixed string and table children ---

func TestLuaE2E_StringAndTableChildren(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "mixed",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("box", {},
					"Header",
					lumina.createElement("text", {}, "Body"))
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	// The "Header" string child should become the box's content.
	// The "Body" table child should appear as a child text node.
	if !screenHasString(ta, "Header") {
		t.Error("expected 'Header' on screen (string child of box)")
	}
	if !screenHasString(ta, "Body") {
		t.Error("expected 'Body' on screen (table child of box)")
	}
}

// --- Test 10: Perf tracker during Lua render ---

func TestLuaE2E_PerfTrackerDuringLuaRender(t *testing.T) {
	app, _, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "perf-comp",
			x = 0, y = 0, w = 20, h = 5,
			render = function(state, props)
				return lumina.createElement("text", {}, "Perf")
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	// Enable perf tracker before rendering.
	tracker := app.Tracker()
	tracker.Enable()

	app.RenderAll()

	frame := tracker.LastFrame()

	// Exactly 1 component was rendered.
	if got := frame.Get(perf.Renders); got != 1 {
		t.Errorf("Renders = %d, want 1", got)
	}
	// Layout should have been called.
	if got := frame.Get(perf.Layouts); got < 1 {
		t.Errorf("Layouts = %d, want >= 1", got)
	}
	// Paint should have been called.
	if got := frame.Get(perf.Paints); got < 1 {
		t.Errorf("Paints = %d, want >= 1", got)
	}
	// Full compose should have happened.
	if got := frame.Get(perf.ComposeFull); got != 1 {
		t.Errorf("ComposeFull = %d, want 1", got)
	}
}

// --- Test 11: Multi-click increments counter 3 times ---

func TestLuaE2E_MultiClick(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "multi-click",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				local count, setCount = lumina.useState("clicks", 0)
				return lumina.createElement("text", {
					id = "mc-target",
					onClick = function()
						setCount(count + 1)
					end,
				}, "N:" .. tostring(count))
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()
	if !screenHasString(ta, "N:0") {
		t.Fatal("expected 'N:0' on initial render")
	}

	// Click 3 times.
	for i := 1; i <= 3; i++ {
		app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
		app.RenderDirty()
	}

	if !screenHasString(ta, "N:3") {
		t.Errorf("expected 'N:3' after 3 clicks, got line: %q", readScreenLine(ta, 0, 40))
	}
}

// --- Test 12: Keyboard event dispatched to focused Lua component ---

func TestLuaE2E_KeyboardEvent(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "kbd",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				local key, setKey = lumina.useState("lastKey", "none")
				return lumina.createElement("text", {
					id = "kbd-target",
					focusable = true,
					onKeyDown = function(e)
						setKey(e.key)
					end,
				}, "Key:" .. tostring(key))
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()
	if !screenHasString(ta, "Key:none") {
		t.Fatal("expected 'Key:none' on initial render")
	}

	// Focus the target via Tab.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})

	// Send a key event.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Enter"})
	app.RenderDirty()

	if !screenHasString(ta, "Key:Enter") {
		t.Errorf("expected 'Key:Enter' after keydown, got line: %q", readScreenLine(ta, 0, 40))
	}
}

// --- Test 13: Counter.lua with onClick (full Lua→Event→State→Render→Screen) ---

func TestLuaE2E_CounterWithClick(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	// Inline counter with onClick on the box — the full closed-loop test.
	err := app.RunString(`
		lumina.createComponent({
			id = "counter",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				local count, setCount = lumina.useState("count", 0)
				return lumina.createElement("box", {
					id = "counter-box",
					style = {background = "#1E1E2E"},
					onClick = function()
						setCount(count + 1)
					end,
				},
					lumina.createElement("text", {}, "Count: " .. tostring(count)),
					lumina.createElement("text", {}, "Click to increment"))
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	// Step 1: Initial render → Count: 0
	app.RenderAll()
	if !screenHasString(ta, "Count: 0") {
		t.Fatal("expected 'Count: 0' on initial render")
	}

	// Step 2: Click → Count: 1
	app.HandleEvent(&event.Event{Type: "mousedown", X: 5, Y: 2})
	app.RenderDirty()
	if !screenHasString(ta, "Count: 1") {
		t.Errorf("expected 'Count: 1' after 1st click, got line: %q", readScreenLine(ta, 0, 40))
	}

	// Step 3: Click again → Count: 2
	app.HandleEvent(&event.Event{Type: "mousedown", X: 5, Y: 2})
	app.RenderDirty()
	if !screenHasString(ta, "Count: 2") {
		t.Errorf("expected 'Count: 2' after 2nd click, got line: %q", readScreenLine(ta, 0, 40))
	}

	// Step 4: Click a third time → Count: 3
	app.HandleEvent(&event.Event{Type: "mousedown", X: 5, Y: 2})
	app.RenderDirty()
	if !screenHasString(ta, "Count: 3") {
		t.Errorf("expected 'Count: 3' after 3rd click, got line: %q", readScreenLine(ta, 0, 40))
	}
}

// --- Test 14: Text inherits parent box background ---

func TestLuaE2E_TextInheritsParentBackground(t *testing.T) {
	app, _, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "bg-test",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("box", {
					style = {background = "#1E1E2E"},
				},
					lumina.createElement("text", {}, "Hello")
				)
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()
	screen := app.Screen()

	if screen == nil {
		t.Fatal("Screen is nil")
	}

	// Text cell at (0,0) should have 'H' with background "#1E1E2E" (inherited from box)
	textCell := screen.Get(0, 0)
	if textCell.Char != 'H' {
		t.Errorf("expected 'H' at (0,0), got %q", textCell.Char)
	}
	if textCell.Background != "#1E1E2E" {
		t.Errorf("text cell background: got %q, want '#1E1E2E' (should inherit from parent box)", textCell.Background)
	}

	// Empty cell at (10, 5) should also have background "#1E1E2E" (box fill)
	emptyCell := screen.Get(10, 5)
	if emptyCell.Background != "#1E1E2E" {
		t.Errorf("empty cell background: got %q, want '#1E1E2E'", emptyCell.Background)
	}

	// Both should be the same background
	if textCell.Background != emptyCell.Background {
		t.Errorf("text bg %q != empty bg %q — text should inherit parent background",
			textCell.Background, emptyCell.Background)
	}
}

// --- Test: Top-level style props in createElement ---

func TestLuaE2E_TextForegroundColor(t *testing.T) {
	app, _, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "color-test",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("text", {
					foreground = "#89B4FA",
					bold = true,
				}, "Colored")
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()
	screen := app.Screen()

	if screen == nil {
		t.Fatal("Screen is nil")
	}

	// First character should be 'C' from "Colored"
	cell := screen.Get(0, 0)
	if cell.Char != 'C' {
		t.Errorf("expected 'C' at (0,0), got %c (%d)", cell.Char, cell.Char)
	}
	if cell.Foreground != "#89B4FA" {
		t.Errorf("foreground = %q, want %q", cell.Foreground, "#89B4FA")
	}
	if !cell.Bold {
		t.Error("expected bold=true at (0,0)")
	}
}
