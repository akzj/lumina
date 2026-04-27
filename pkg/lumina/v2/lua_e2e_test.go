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

// --- Test: TodoMVC Chinese/Unicode input ---

func TestLuaE2E_TodoMVC_ChineseInput(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/todo_mvc.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// Enter input mode by pressing "a".
	app.HandleEvent(&event.Event{Type: "keydown", Key: "a"})
	app.RenderDirty()

	// Type Chinese characters: "你好"
	app.HandleEvent(&event.Event{Type: "keydown", Key: "你"})
	app.RenderDirty()
	app.HandleEvent(&event.Event{Type: "keydown", Key: "好"})
	app.RenderDirty()

	// Verify both Chinese characters appear on screen (in the input bar).
	// Note: CJK wide characters occupy 2 cells each with a zero-padding cell,
	// so screenHasString can't match them consecutively. Use screenHasChar.
	if !screenHasChar(ta, '你') {
		t.Error("Chinese character '你' not found on screen after typing")
	}
	if !screenHasChar(ta, '好') {
		t.Error("Chinese character '好' not found on screen after typing")
	}

	// Test backspace removes one Chinese character (not just last byte).
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Backspace"})
	app.RenderDirty()

	// "你" should remain, "好" should be gone.
	if !screenHasChar(ta, '你') {
		t.Error("'你' should still be on screen after one backspace")
	}
	if screenHasChar(ta, '好') {
		t.Error("'好' should be removed after one backspace")
	}

	// Submit with Enter — "你" becomes a new todo item.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Enter"})
	app.RenderDirty()

	// "你" should appear in the todo list with a "[ ]" checkbox.
	if !screenHasChar(ta, '你') {
		t.Error("Chinese todo '你' not found in todo list after submit")
	}
	// Header should show 6 items.
	if !screenHasString(ta, "6 items") {
		t.Error("expected '6 items' in header after adding Chinese todo")
	}
}

// --- Test: Input component Chinese/Unicode typing ---

func TestLuaE2E_InputChineseTyping(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 5)

	err := app.RunString(`
		lumina.createComponent({
			id = "cjk-test",
			x = 0, y = 0, w = 40, h = 5,
			render = function(state, props)
				local text, setText = lumina.useState("text", "")
				return lumina.createElement("input", {
					id = "cjk-input",
					value = text,
					onChange = function(newValue) setText(newValue) end,
				})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	// Focus the input.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	app.RenderDirty()

	// Type Chinese characters: "中文"
	app.HandleEvent(&event.Event{Type: "keydown", Key: "中"})
	app.RenderDirty()
	app.HandleEvent(&event.Event{Type: "keydown", Key: "文"})
	app.RenderDirty()

	// Verify "中文" appears on screen.
	if !screenHasChar(ta, '中') {
		t.Error("Chinese character '中' not found on screen")
	}
	if !screenHasChar(ta, '文') {
		t.Error("Chinese character '文' not found on screen")
	}

	// Backspace should remove '文' (last character).
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Backspace"})
	app.RenderDirty()

	if !screenHasChar(ta, '中') {
		t.Error("'中' should remain after one backspace")
	}
	if screenHasChar(ta, '文') {
		t.Error("'文' should be removed after backspace")
	}
}

// --- Test: Dashboard example loads and renders all panels ---

func TestLuaE2E_Dashboard(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/dashboard.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}

	// 1. Title "Dashboard" appears on screen.
	if !screenHasString(ta, "Dashboard") {
		t.Error("expected 'Dashboard' title on screen")
	}

	// 2. Progress bar characters appear (█ for filled, ░ for empty).
	if !screenHasChar(ta, '█') {
		t.Error("expected filled progress bar character '█' on screen")
	}
	if !screenHasChar(ta, '░') {
		t.Error("expected empty progress bar character '░' on screen")
	}

	// 3. Resource labels appear.
	if !screenHasString(ta, "CPU") {
		t.Error("expected 'CPU' label on screen")
	}
	if !screenHasString(ta, "RAM") {
		t.Error("expected 'RAM' label on screen")
	}
	if !screenHasString(ta, "Disk") {
		t.Error("expected 'Disk' label on screen")
	}

	// 4. Activity log entries appear.
	if !screenHasString(ta, "Server started") {
		t.Error("expected 'Server started' activity entry on screen")
	}

	// 5. Stats appear.
	if !screenHasString(ta, "Uptime") {
		t.Error("expected 'Uptime' stat on screen")
	}
	if !screenHasString(ta, "42 days") {
		t.Error("expected '42 days' stat value on screen")
	}

	// 6. Keyboard help appears in footer.
	if !screenHasString(ta, "[q] Quit") {
		t.Error("expected '[q] Quit' help text on screen")
	}
}

// --- Test: Dashboard keyboard scroll changes activity log view ---

func TestLuaE2E_Dashboard_Scroll(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/dashboard.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// The first activity entry should be visible initially.
	if !screenHasString(ta, "Server started") {
		t.Fatal("expected 'Server started' visible initially")
	}

	// A late entry (entry ~40+) should NOT be visible before scrolling.
	if screenHasString(ta, "Daily report") {
		t.Fatal("'Daily report' should not be visible before scrolling")
	}

	// Press 'j' many times to scroll down in the activity log.
	// With 45 entries and ~20 visible rows, scrolling 30 times should
	// push early entries off screen and reveal late entries.
	for i := 0; i < 30; i++ {
		app.HandleEvent(&event.Event{Type: "keydown", Key: "j"})
		app.RenderDirty()
	}

	// After scrolling down significantly, late entries should be visible.
	if !screenHasString(ta, "Daily report") {
		t.Error("expected 'Daily report' visible after scrolling down")
	}

	// Early entries should have scrolled off screen.
	if screenHasString(ta, "Server started") {
		t.Error("'Server started' should have scrolled off screen")
	}

	// Press 'k' many times to scroll back to top.
	for i := 0; i < 30; i++ {
		app.HandleEvent(&event.Event{Type: "keydown", Key: "k"})
		app.RenderDirty()
	}

	// First entry should be visible again.
	if !screenHasString(ta, "Server started") {
		t.Error("expected 'Server started' visible after scrolling back up")
	}
}

// --- Test: Dashboard auto-focus (keyboard works without Tab) ---

func TestLuaE2E_Dashboard_AutoFocus(t *testing.T) {
	app, _, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/dashboard.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// The focusable root "dashboard-root" should be auto-focused.
	if id := app.FocusedID(); id != "dashboard-root" {
		t.Errorf("expected auto-focus on 'dashboard-root', got %q", id)
	}
}

// --- Test: Dashboard mouse scroll (onScroll event) ---

func TestLuaE2E_Dashboard_MouseScroll(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/dashboard.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// First entry visible initially.
	if !screenHasString(ta, "Server started") {
		t.Fatal("expected 'Server started' visible initially")
	}

	// A late entry should NOT be visible before scrolling.
	if screenHasString(ta, "Daily report") {
		t.Fatal("'Daily report' should not be visible before scrolling")
	}

	// Send mouse wheel scroll events on the activity log area (right half).
	// The activity log is on the right side (x=40+), rows 3-20.
	for i := 0; i < 10; i++ {
		app.HandleEvent(&event.Event{Type: "scroll", X: 60, Y: 10, Key: "down"})
		app.RenderDirty()
	}

	// After scrolling down, late entries should appear.
	if !screenHasString(ta, "Daily report") {
		t.Error("expected 'Daily report' visible after mouse scroll down")
	}

	// Scroll back up.
	for i := 0; i < 10; i++ {
		app.HandleEvent(&event.Event{Type: "scroll", X: 60, Y: 10, Key: "up"})
		app.RenderDirty()
	}

	// First entry should be visible again.
	if !screenHasString(ta, "Server started") {
		t.Error("expected 'Server started' visible after mouse scroll up")
	}
}

// ═══════════════════════════════════════════════════════════════════
// Component Library E2E Tests
// ═══════════════════════════════════════════════════════════════════

// --- Test: ProgressBar renders bar characters and percentage ---

func TestLuaE2E_ProgressBar(t *testing.T) {
	app, ta, _ := newLuaApp(t, 60, 5)

	err := app.RunString(`
		local function ProgressBar(props)
			local value = math.max(0, math.min(100, props.value or 0))
			local width = props.width or 20
			local color = props.color or "#A6E3A1"
			local label = props.label or ""

			local filled = math.floor(value / 100 * width)
			local empty = width - filled
			local bar = string.rep("█", filled) .. string.rep("░", empty)
			local pct = string.format("%3d%%", value)

			local pctColor = "#A6E3A1"
			if value > 80 then pctColor = "#F38BA8"
			elseif value > 60 then pctColor = "#F9E2AF"
			end

			local children = {}
			if label ~= "" then
				children[#children + 1] = lumina.createElement("text", {
					foreground = "#CDD6F4",
				}, label)
			end
			children[#children + 1] = lumina.createElement("text", {
				foreground = color,
			}, bar)
			children[#children + 1] = lumina.createElement("text", {
				foreground = pctColor,
			}, pct)

			return lumina.createElement("hbox", {
				style = {gap = 1},
			}, table.unpack(children))
		end

		lumina.createComponent({
			id = "pb-test",
			x = 0, y = 0, w = 60, h = 5,
			render = function(state, props)
				return lumina.createElement("vbox", {},
					ProgressBar({label = "CPU ", value = 50, width = 10}),
					ProgressBar({value = 100, width = 10}),
					ProgressBar({value = 0, width = 10})
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}

	// Row 0: "CPU " label + bar + " 50%"
	if !screenHasString(ta, "CPU") {
		t.Error("expected 'CPU' label on screen")
	}
	if !screenHasString(ta, "50%") {
		t.Error("expected '50%' on screen")
	}

	// Row 1: 100% bar should have filled characters
	if !screenHasString(ta, "100%") {
		t.Error("expected '100%' on screen")
	}

	// Row 2: 0% bar should have empty characters
	if !screenHasString(ta, "0%") {
		t.Error("expected '0%' on screen for 0-value bar")
	}

	// Verify bar characters exist (█ and ░)
	if !screenHasChar(ta, '█') {
		t.Error("expected filled bar character '█' on screen")
	}
	if !screenHasChar(ta, '░') {
		t.Error("expected empty bar character '░' on screen")
	}
}

// --- Test: Table renders headers and data rows ---

func TestLuaE2E_Table(t *testing.T) {
	app, ta, _ := newLuaApp(t, 60, 10)

	err := app.RunString(`
		local function DataTable(props)
			local headers = props.headers or {}
			local rows = props.rows or {}
			local selectedRow = props.selectedRow or -1
			local colWidths = props.colWidths

			if not colWidths then
				colWidths = {}
				for i, h in ipairs(headers) do
					colWidths[i] = #tostring(h) + 2
				end
				for _, row in ipairs(rows) do
					for i, cell in ipairs(row) do
						local w = #tostring(cell) + 2
						if w > (colWidths[i] or 0) then colWidths[i] = w end
					end
				end
			end

			local totalWidth = 0
			for _, w in ipairs(colWidths) do totalWidth = totalWidth + w end

			local children = {}

			local headerCells = {}
			for i, h in ipairs(headers) do
				local text = tostring(h)
				local cw = colWidths[i] or #text
				local padded = text .. string.rep(" ", math.max(0, cw - #text))
				headerCells[#headerCells + 1] = lumina.createElement("text", {
					foreground = "#89B4FA", bold = true,
				}, padded)
			end
			children[#children + 1] = lumina.createElement("hbox", {},
				table.unpack(headerCells))

			children[#children + 1] = lumina.createElement("text", {
				foreground = "#585B70",
			}, string.rep("─", totalWidth))

			for ri, row in ipairs(rows) do
				local rowCells = {}
				local isSelected = (ri == selectedRow)
				for i, cell in ipairs(row) do
					local text = tostring(cell)
					local cw = colWidths[i] or #text
					local padded = text .. string.rep(" ", math.max(0, cw - #text))
					local fg = isSelected and "#1E1E2E" or "#CDD6F4"
					local cellProps = {foreground = fg}
					if isSelected then cellProps.background = "#89B4FA" end
					rowCells[#rowCells + 1] = lumina.createElement("text",
						cellProps, padded)
				end
				children[#children + 1] = lumina.createElement("hbox", {},
					table.unpack(rowCells))
			end

			local outerProps = {}
			if props.style then outerProps.style = props.style end
			return lumina.createElement("vbox", outerProps, table.unpack(children))
		end

		lumina.createComponent({
			id = "table-test",
			x = 0, y = 0, w = 60, h = 10,
			render = function(state, props)
				return DataTable({
					headers = {"Name", "Role", "Status"},
					rows = {
						{"Alice", "Admin", "Active"},
						{"Bob", "Dev", "Away"},
					},
					selectedRow = 1,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}

	// Verify headers
	if !screenHasString(ta, "Name") {
		t.Error("expected 'Name' header on screen")
	}
	if !screenHasString(ta, "Role") {
		t.Error("expected 'Role' header on screen")
	}
	if !screenHasString(ta, "Status") {
		t.Error("expected 'Status' header on screen")
	}

	// Verify separator
	if !screenHasChar(ta, '─') {
		t.Error("expected separator character '─' on screen")
	}

	// Verify data rows
	if !screenHasString(ta, "Alice") {
		t.Error("expected 'Alice' on screen")
	}
	if !screenHasString(ta, "Bob") {
		t.Error("expected 'Bob' on screen")
	}
	if !screenHasString(ta, "Admin") {
		t.Error("expected 'Admin' on screen")
	}
}

// --- Test: Tabs renders tab buttons and active content ---

func TestLuaE2E_Tabs(t *testing.T) {
	app, ta, _ := newLuaApp(t, 60, 10)

	err := app.RunString(`
		local function Tabs(props)
			local tabs = props.tabs or {}
			local activeTab = props.activeTab or 1
			local separatorLen = props.separatorLen or 40

			local tabButtons = {}
			for i, tab in ipairs(tabs) do
				local isActive = (i == activeTab)
				tabButtons[#tabButtons + 1] = lumina.createElement("text", {
					foreground = isActive and "#1E1E2E" or "#CDD6F4",
					background = isActive and "#89B4FA" or "#313244",
					bold = isActive,
				}, " " .. tab.label .. " ")
			end

			local children = {}
			children[#children + 1] = lumina.createElement("hbox", {
				style = {gap = 1},
			}, table.unpack(tabButtons))
			children[#children + 1] = lumina.createElement("text", {
				foreground = "#585B70",
			}, string.rep("─", separatorLen))

			if tabs[activeTab] and tabs[activeTab].content then
				children[#children + 1] = tabs[activeTab].content
			end

			local outerProps = {}
			if props.style then outerProps.style = props.style end
			return lumina.createElement("vbox", outerProps, table.unpack(children))
		end

		lumina.createComponent({
			id = "tabs-test",
			x = 0, y = 0, w = 60, h = 10,
			render = function(state, props)
				local activeTab, setActiveTab = lumina.useState("activeTab", 2)

				return Tabs({
					activeTab = activeTab,
					separatorLen = 50,
					tabs = {
						{label = "First", content = lumina.createElement("text", {}, "Content A")},
						{label = "Second", content = lumina.createElement("text", {}, "Content B")},
						{label = "Third", content = lumina.createElement("text", {}, "Content C")},
					},
				})
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}

	// Verify tab labels appear
	if !screenHasString(ta, "First") {
		t.Error("expected 'First' tab label on screen")
	}
	if !screenHasString(ta, "Second") {
		t.Error("expected 'Second' tab label on screen")
	}
	if !screenHasString(ta, "Third") {
		t.Error("expected 'Third' tab label on screen")
	}

	// Active tab is 2 → "Content B" should be visible
	if !screenHasString(ta, "Content B") {
		t.Error("expected 'Content B' (active tab content) on screen")
	}

	// Switch to tab 1 and verify content changes
	app.SetState("tabs-test", "activeTab", int64(1))
	app.RenderDirty()

	if !screenHasString(ta, "Content A") {
		t.Error("expected 'Content A' after switching to tab 1")
	}
}

// --- Test: Select renders options with highlighted selection ---

func TestLuaE2E_Select(t *testing.T) {
	app, ta, _ := newLuaApp(t, 40, 10)

	err := app.RunString(`
		local function Select(props)
			local options = props.options or {}
			local selected = props.selected or 1
			local label = props.label or ""

			local children = {}
			if label ~= "" then
				children[#children + 1] = lumina.createElement("text", {
					foreground = "#89B4FA", bold = true,
				}, label)
			end

			for i, opt in ipairs(options) do
				local isSelected = (i == selected)
				local prefix = isSelected and "▸ " or "  "
				local fg = isSelected and "#A6E3A1" or "#CDD6F4"
				local cellProps = {foreground = fg, bold = isSelected}
				if isSelected then cellProps.background = "#313244" end
				children[#children + 1] = lumina.createElement("text",
					cellProps, prefix .. opt)
			end

			local outerProps = {}
			if props.style then outerProps.style = props.style end
			return lumina.createElement("vbox", outerProps, table.unpack(children))
		end

		lumina.createComponent({
			id = "select-test",
			x = 0, y = 0, w = 40, h = 10,
			render = function(state, props)
				local sel, setSel = lumina.useState("sel", 2)

				return Select({
					label = "Choose:",
					options = {"Alpha", "Beta", "Gamma"},
					selected = sel,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}

	// Verify label
	if !screenHasString(ta, "Choose:") {
		t.Error("expected 'Choose:' label on screen")
	}

	// Verify all options appear
	if !screenHasString(ta, "Alpha") {
		t.Error("expected 'Alpha' on screen")
	}
	if !screenHasString(ta, "Beta") {
		t.Error("expected 'Beta' on screen")
	}
	if !screenHasString(ta, "Gamma") {
		t.Error("expected 'Gamma' on screen")
	}

	// Verify selection indicator on "Beta" (selected = 2)
	if !screenHasString(ta, "▸ Beta") {
		t.Error("expected '▸ Beta' (selected indicator) on screen")
	}

	// Non-selected items should have space prefix
	if !screenHasString(ta, "  Alpha") {
		t.Error("expected '  Alpha' (non-selected prefix) on screen")
	}

	// Change selection to 3 and verify
	app.SetState("select-test", "sel", int64(3))
	app.RenderDirty()

	if !screenHasString(ta, "▸ Gamma") {
		t.Error("expected '▸ Gamma' after changing selection to 3")
	}
}

// --- Test: Modal visible/hidden toggle ---

func TestLuaE2E_Modal(t *testing.T) {
	app, ta, _ := newLuaApp(t, 60, 15)

	err := app.RunString(`
		local function Modal(props)
			if not props.visible then
				return lumina.createElement("text", {}, "")
			end

			local w = props.width or 40
			local title = props.title or "Dialog"

			return lumina.createElement("box", {
				style = {
					border = "rounded",
					background = "#1E1E2E",
					padding = 1,
				},
			},
				lumina.createElement("text", {
					foreground = "#89B4FA", bold = true,
				}, title),
				lumina.createElement("text", {
					foreground = "#585B70",
				}, string.rep("─", math.max(0, w - 4))),
				props.children or lumina.createElement("text", {}, ""),
				lumina.createElement("text", {}, ""),
				lumina.createElement("text", {
					foreground = "#6C7086",
				}, "[Esc] Close")
			)
		end

		lumina.createComponent({
			id = "modal-test",
			x = 0, y = 0, w = 60, h = 15,
			render = function(state, props)
				local show, setShow = lumina.useState("show", false)

				return lumina.createElement("vbox", {},
					lumina.createElement("text", {}, "Background Content"),
					Modal({
						visible = show,
						title = "Test Modal",
						width = 40,
						children = lumina.createElement("text", {
							foreground = "#CDD6F4",
						}, "Modal body text"),
					})
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	// Initial render: modal hidden
	app.RenderAll()

	if !screenHasString(ta, "Background Content") {
		t.Error("expected 'Background Content' on screen")
	}
	if screenHasString(ta, "Test Modal") {
		t.Error("modal title should NOT be visible when hidden")
	}

	// Show modal
	app.SetState("modal-test", "show", true)
	app.RenderDirty()

	if !screenHasString(ta, "Test Modal") {
		t.Error("expected 'Test Modal' title when modal is visible")
	}
	if !screenHasString(ta, "Modal body text") {
		t.Error("expected 'Modal body text' when modal is visible")
	}
	if !screenHasString(ta, "[Esc] Close") {
		t.Error("expected '[Esc] Close' hint when modal is visible")
	}

	// Hide modal again — use RenderAll because RenderDirty does not clear
	// regions that are no longer occupied by VNodes (expected behavior).
	app.SetState("modal-test", "show", false)
	app.RenderAll()

	if screenHasString(ta, "Test Modal") {
		t.Error("modal title should NOT be visible after hiding")
	}
}

// --- Test: Component Showcase loads and renders without error ---

func TestLuaE2E_ComponentShowcase(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/components_showcase.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}

	// Verify the header renders
	if !screenHasString(ta, "Component Library Showcase") {
		t.Error("expected showcase title on screen")
	}

	// Verify tab labels are visible
	if !screenHasString(ta, "Progress") {
		t.Error("expected 'Progress' tab label")
	}
	if !screenHasString(ta, "Table") {
		t.Error("expected 'Table' tab label")
	}
	if !screenHasString(ta, "Select") {
		t.Error("expected 'Select' tab label")
	}

	// Default tab is 1 (Progress) — verify progress bar content
	if !screenHasChar(ta, '█') {
		t.Error("expected filled bar character on Progress tab")
	}

	// Verify footer
	if !screenHasString(ta, "Quit") {
		t.Error("expected 'Quit' in footer")
	}
}

// --- Test: Component Showcase tab switching via keyboard ---

func TestLuaE2E_ComponentShowcase_TabSwitch(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/components_showcase.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// Switch to tab 2 (Table)
	app.HandleEvent(&event.Event{Type: "keydown", Key: "2"})
	app.RenderDirty()

	// Table tab should show headers
	if !screenHasString(ta, "Name") {
		t.Error("expected 'Name' header on Table tab")
	}
	if !screenHasString(ta, "Alice") {
		t.Error("expected 'Alice' data on Table tab")
	}

	// Switch to tab 3 (Select)
	app.HandleEvent(&event.Event{Type: "keydown", Key: "3"})
	app.RenderDirty()

	// Select tab should show theme options
	if !screenHasString(ta, "Theme:") {
		t.Error("expected 'Theme:' label on Select tab")
	}
	if !screenHasString(ta, "Dark Mode") {
		t.Error("expected 'Dark Mode' option on Select tab")
	}
}

// --- Test: Component Showcase modal toggle ---

func TestLuaE2E_ComponentShowcase_Modal(t *testing.T) {
	app, ta, _ := newLuaApp(t, 80, 24)

	err := app.RunScript("../../../examples/v2/components_showcase.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	// Modal should not be visible initially
	if screenHasString(ta, "Example Modal") {
		t.Error("modal should NOT be visible initially")
	}

	// Press 'm' to show modal
	app.HandleEvent(&event.Event{Type: "keydown", Key: "m"})
	app.RenderDirty()

	if !screenHasString(ta, "Example Modal") {
		t.Error("expected 'Example Modal' after pressing 'm'")
	}

	// Press Escape to close modal
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Escape"})
	app.RenderDirty()

	if screenHasString(ta, "Example Modal") {
		t.Error("modal should be hidden after pressing Escape")
	}
}
