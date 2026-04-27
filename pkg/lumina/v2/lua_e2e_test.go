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

// --- Test: TodoMVC script loads and renders ---

func TestLuaE2E_TodoMVC_Render(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/todo_mvc.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}

	// Verify title is on screen.
	if !screenHasString(ta, "Todo MVC") {
		t.Error("expected 'Todo MVC' title on screen")
	}

	// Verify first todo is visible.
	if !screenHasString(ta, "Learn Lumina v2") {
		t.Error("expected 'Learn Lumina v2' on screen")
	}

	// Verify footer help text.
	if !screenHasString(ta, "[j/k] Navigate") {
		t.Error("expected '[j/k] Navigate' help text on screen")
	}

	// Verify the first todo is selected (has "> " prefix).
	if !screenHasString(ta, "> [x] Learn Lumina v2") {
		t.Error("expected '> [x] Learn Lumina v2' (selected first item)")
	}
}

// --- Test: TodoMVC keyboard navigation (j/k) ---

func TestLuaE2E_TodoMVC_Navigation(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/todo_mvc.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// Initially, first item is selected (has "> " prefix).
	if !screenHasString(ta, "> [x] Learn Lumina v2") {
		t.Fatal("expected first item selected initially")
	}

	// Press "j" to move down.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "j"})
	app.RenderDirty()

	// Now second item should be selected.
	if !screenHasString(ta, "> [ ] Build a TUI app") {
		t.Error("expected second item selected after 'j'")
	}

	// Press "k" to move back up.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "k"})
	app.RenderDirty()

	// Back to first item.
	if !screenHasString(ta, "> [x] Learn Lumina v2") {
		t.Error("expected first item selected after 'k'")
	}
}

// --- Test: TodoMVC toggle done ---

func TestLuaE2E_TodoMVC_Toggle(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/todo_mvc.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// First item starts as done: "[x]"
	if !screenHasString(ta, "> [x] Learn Lumina v2") {
		t.Fatal("expected first item done initially")
	}

	// Press Space to toggle.
	app.HandleEvent(&event.Event{Type: "keydown", Key: " "})
	app.RenderDirty()

	// Should now be undone: "[ ]"
	if !screenHasString(ta, "> [ ] Learn Lumina v2") {
		t.Error("expected first item toggled to undone after Space")
	}

	// Press Space again to toggle back.
	app.HandleEvent(&event.Event{Type: "keydown", Key: " "})
	app.RenderDirty()

	if !screenHasString(ta, "> [x] Learn Lumina v2") {
		t.Error("expected first item toggled back to done after 2nd Space")
	}
}

// --- Test: TodoMVC filter cycling ---

func TestLuaE2E_TodoMVC_Filter(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/todo_mvc.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// Initially "all" filter is active — all 5 todos visible.
	if !screenHasString(ta, "Deploy to production") {
		t.Fatal("expected all todos visible initially")
	}

	// Press "f" to switch to "active" filter.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "f"})
	app.RenderDirty()

	// "Learn Lumina v2" is done, so it should NOT be visible in active filter.
	if screenHasString(ta, "Learn Lumina v2") {
		t.Error("done todo should be hidden in 'active' filter")
	}
	// Active todos should still be visible.
	if !screenHasString(ta, "Build a TUI app") {
		t.Error("expected active todo visible in 'active' filter")
	}

	// Press "f" again → "completed" filter.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "f"})
	app.RenderDirty()

	// Only completed todo should be visible.
	if !screenHasString(ta, "Learn Lumina v2") {
		t.Error("expected completed todo visible in 'completed' filter")
	}
	if screenHasString(ta, "Build a TUI app") {
		t.Error("active todo should be hidden in 'completed' filter")
	}
}

// --- Test: TodoMVC add todo ---

func TestLuaE2E_TodoMVC_AddTodo(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/todo_mvc.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// Press "a" to enter input mode.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "a"})
	app.RenderDirty()

	// Type "New task".
	for _, ch := range "New task" {
		app.HandleEvent(&event.Event{Type: "keydown", Key: string(ch)})
		app.RenderDirty()
	}

	// Should see input text on screen.
	if !screenHasString(ta, "New task") {
		t.Error("expected 'New task' visible in input bar")
	}

	// Press Enter to submit.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Enter"})
	app.RenderDirty()

	// New todo should appear in the list.
	if !screenHasString(ta, "New task") {
		t.Error("expected 'New task' in todo list after Enter")
	}

	// Should show 6 items now in header.
	if !screenHasString(ta, "6 items") {
		t.Error("expected '6 items' in header after adding todo")
	}
}

// --- Test: TodoMVC delete todo ---

func TestLuaE2E_TodoMVC_Delete(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/todo_mvc.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// Verify first item exists.
	if !screenHasString(ta, "Learn Lumina v2") {
		t.Fatal("expected 'Learn Lumina v2' initially")
	}

	// Press "d" to delete the selected (first) item.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "d"})
	app.RenderDirty()

	// First item should be gone.
	if screenHasString(ta, "Learn Lumina v2") {
		t.Error("expected 'Learn Lumina v2' removed after delete")
	}

	// Should show 4 items now.
	if !screenHasString(ta, "4 items") {
		t.Error("expected '4 items' in header after delete")
	}
}

// --- Test: TodoMVC auto-focus (keyboard works without Tab) ---

func TestLuaE2E_TodoMVC_AutoFocus(t *testing.T) {
	app, _, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/todo_mvc.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// After RenderAll, the focusable VNode "todo-root" should be auto-focused.
	if id := app.FocusedID(); id != "todo-root" {
		t.Errorf("expected auto-focus on 'todo-root', got %q", id)
	}
}

// --- Test: Input renders value on screen ---

func TestLuaE2E_InputRender(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "input-test",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("input", {
					id = "my-input",
					value = "Hello World",
					foreground = "#CDD6F4",
				})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if !screenHasString(ta, "Hello World") {
		t.Error("expected 'Hello World' on screen from input value")
	}
}

// --- Test: Input shows placeholder when empty ---

func TestLuaE2E_InputPlaceholder(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "input-ph",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("input", {
					id = "ph-input",
					value = "",
					placeholder = "Type here",
				})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if !screenHasString(ta, "Type here") {
		t.Error("expected 'Type here' placeholder on screen")
	}
}

// --- Test: Input typing updates screen via onChange ---

func TestLuaE2E_InputTyping(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "input-type",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				local text, setText = lumina.useState("text", "")
				return lumina.createElement("input", {
					id = "type-input",
					value = text,
					onChange = function(newValue)
						setText(newValue)
					end,
				})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	// Focus the input (Tab cycles to the first focusable).
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	app.RenderDirty()

	// Type "Hi"
	app.HandleEvent(&event.Event{Type: "keydown", Key: "H"})
	app.RenderDirty()
	app.HandleEvent(&event.Event{Type: "keydown", Key: "i"})
	app.RenderDirty()

	if !screenHasString(ta, "Hi") {
		t.Errorf("expected 'Hi' on screen after typing, got line: %q", readScreenLine(ta, 0, 40))
	}
}

// --- Test: Input backspace removes character ---

func TestLuaE2E_InputBackspace(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "input-bs",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				local text, setText = lumina.useState("text", "ABC")
				return lumina.createElement("input", {
					id = "bs-input",
					value = text,
					onChange = function(newValue)
						setText(newValue)
					end,
				})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	// Verify initial value.
	if !screenHasString(ta, "ABC") {
		t.Fatal("expected 'ABC' on initial render")
	}

	// Focus the input.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	app.RenderDirty()

	// Press Backspace (cursor is at end, removes 'C').
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Backspace"})
	app.RenderDirty()

	if !screenHasString(ta, "AB") {
		t.Errorf("expected 'AB' after backspace, got line: %q", readScreenLine(ta, 0, 40))
	}

	// If "ABC" is still there, backspace didn't work.
	if screenHasString(ta, "ABC") {
		t.Error("'ABC' should be gone after backspace")
	}
}

// --- Test: Input onSubmit fires on Enter ---

func TestLuaE2E_InputOnSubmit(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "input-submit",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				local text, setText = lumina.useState("text", "hello")
				local submitted, setSubmitted = lumina.useState("submitted", "none")
				return lumina.createElement("box", {},
					lumina.createElement("input", {
						id = "submit-input",
						value = text,
						onChange = function(newValue)
							setText(newValue)
						end,
						onSubmit = function(value)
							setSubmitted(value)
						end,
					}),
					lumina.createElement("text", {}, "Submitted:" .. submitted)
				)
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if !screenHasString(ta, "Submitted:none") {
		t.Fatal("expected 'Submitted:none' on initial render")
	}

	// Focus the input.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	app.RenderDirty()

	// Press Enter.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Enter"})
	app.RenderDirty()

	if !screenHasString(ta, "Submitted:hello") {
		t.Errorf("expected 'Submitted:hello' after Enter, got line: %q", readScreenLine(ta, 0, 40))
	}
}

// --- Test: Input auto-focusable (no need for focusable=true in props) ---

func TestLuaE2E_InputAutoFocusable(t *testing.T) {
	app, _, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "input-af",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				return lumina.createElement("input", {
					id = "af-input",
					value = "test",
				})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	// Tab should focus the input (it's auto-focusable).
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})

	if id := app.FocusedID(); id != "af-input" {
		t.Errorf("expected focus on 'af-input', got %q", id)
	}
}
