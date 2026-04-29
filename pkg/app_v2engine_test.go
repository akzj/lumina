package v2

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/output"
)

// newEngineApp creates a NewApp with a fresh Lua state and TestAdapter.
func newEngineApp(t *testing.T, w, h int) (*App, *output.TestAdapter, *lua.State) {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	ta := output.NewTestAdapter()
	app := NewApp(L, w, h, ta)
	return app, ta, L
}

func TestAppV2Engine_BasicRender(t *testing.T) {
	app, ta, _ := newEngineApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "hello",
			name = "Hello",
			render = function(props)
				return lumina.createElement("text", {
					style = {foreground = "#FFFFFF"},
				}, "Hello World")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	app.RenderAll()

	// Verify text appears in output
	screen := ta.LastScreen
	if screen == nil {
		t.Fatal("no screen output")
	}

	// Check that "Hello World" is in the buffer
	found := false
	for x := 0; x < screen.Width(); x++ {
		c := screen.Get(x, 0)
		if c.Char == 'H' {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'H' in screen output")
	}
}

func TestAppV2Engine_TextContent(t *testing.T) {
	app, ta, _ := newEngineApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "content",
			name = "Content",
			render = function(props)
				return lumina.createElement("text", {
					style = {foreground = "#FFF"},
				}, "ABCDE")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	app.RenderAll()

	screen := ta.LastScreen
	if screen == nil {
		t.Fatal("no screen output")
	}

	// Verify the first 5 chars are A, B, C, D, E
	expected := "ABCDE"
	for i, ch := range expected {
		c := screen.Get(i, 0)
		if c.Char != ch {
			t.Errorf("position %d: expected %c, got %c", i, ch, c.Char)
		}
	}
}

func TestAppV2Engine_ClickStateChange(t *testing.T) {
	app, ta, _ := newEngineApp(t, 80, 24)

	err := app.RunString(`
		lumina.createComponent({
			id = "counter",
			name = "Counter",
			render = function(props)
				local count, setCount = lumina.useState("c", 0)
				return lumina.createElement("box", {
					style = {width = 80, height = 24},
					onClick = function() setCount(count + 1) end,
				}, lumina.createElement("text", {id = "val"}, tostring(count)))
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	app.RenderAll()

	// Verify initial state
	comp := app.Engine().GetComponent("counter")
	if comp == nil {
		t.Fatal("component not found")
	}
	if comp.RootNode == nil {
		t.Fatal("component has no root node")
	}

	// Check initial "0" in text child
	if len(comp.RootNode.Children) == 0 {
		t.Fatal("root node has no children")
	}
	textNode := comp.RootNode.Children[0]
	if textNode.Content != "0" {
		t.Errorf("expected initial '0', got %q", textNode.Content)
	}

	// Click → setState → RenderDirty
	app.HandleEvent(&event.Event{Type: "click", X: 10, Y: 10})
	app.RenderDirty()

	// Verify "1" after click
	textNode = comp.RootNode.Children[0]
	if textNode.Content != "1" {
		t.Errorf("expected '1' after click, got %q", textNode.Content)
	}

	// Verify output adapter received the update
	if ta.LastScreen == nil {
		t.Fatal("no screen output after RenderDirty")
	}
}

func TestAppV2Engine_HoverNoRerender(t *testing.T) {
	app, _, L := newEngineApp(t, 80, 24)

	err := app.RunString(`
		renderCount = 0
		lumina.createComponent({
			id = "static",
			name = "Static",
			render = function(props)
				renderCount = renderCount + 1
				return lumina.createElement("box", {
					style = {width = 80, height = 24},
					onMouseEnter = function() end,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	app.RenderAll()

	// Hover around — no setState → no re-render
	app.HandleEvent(&event.Event{Type: "mousemove", X: 10, Y: 5})
	app.HandleEvent(&event.Event{Type: "mousemove", X: 20, Y: 10})
	app.RenderDirty()

	L.GetGlobal("renderCount")
	count, _ := L.ToInteger(-1)
	L.Pop(1)

	if count != 1 {
		t.Errorf("expected 1 render, got %d", count)
	}
}

func TestAppV2Engine_Resize(t *testing.T) {
	app, ta, _ := newEngineApp(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "resizable",
			name = "Resizable",
			render = function(props)
				return lumina.createElement("text", {}, "test")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	app.RenderAll()

	// Resize
	app.Resize(60, 20)
	app.RenderAll()

	screen := ta.LastScreen
	if screen == nil {
		t.Fatal("no screen after resize")
	}
	if screen.Width() != 60 || screen.Height() != 20 {
		t.Errorf("expected 60x20, got %dx%d", screen.Width(), screen.Height())
	}
}

func TestAppV2Engine_MultipleClicks(t *testing.T) {
	app, _, _ := newEngineApp(t, 80, 24)

	err := app.RunString(`
		lumina.createComponent({
			id = "multi",
			name = "Multi",
			render = function(props)
				local count, setCount = lumina.useState("c", 0)
				return lumina.createElement("box", {
					style = {width = 80, height = 24},
					onClick = function() setCount(count + 1) end,
				}, lumina.createElement("text", {}, tostring(count)))
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	app.RenderAll()

	// Click 5 times
	for i := 0; i < 5; i++ {
		app.HandleEvent(&event.Event{Type: "click", X: 10, Y: 10})
		app.RenderDirty()
	}

	comp := app.Engine().GetComponent("multi")
	if comp == nil || comp.RootNode == nil {
		t.Fatal("component not found")
	}
	textNode := comp.RootNode.Children[0]
	if textNode.Content != "5" {
		t.Errorf("expected '5' after 5 clicks, got %q", textNode.Content)
	}
}

func TestAppV2Engine_F12DevToolsToggle(t *testing.T) {
	app, ta, _ := newEngineApp(t, 40, 20)

	err := app.RunString(`
		lumina.createComponent({
			id = "root",
			render = function(props)
				return { type = "text", content = "Hello" }
			end
		})
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()
	writesBefore := ta.WriteCount

	// Initially devtools is not visible.
	if app.DevTools().Visible {
		t.Fatal("devtools should be hidden initially")
	}

	// Press F12 to open devtools.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F12"})

	if !app.DevTools().Visible {
		t.Fatal("devtools should be visible after F12")
	}

	// Verify something was written to the adapter (the devtools overlay).
	if ta.WriteCount <= writesBefore {
		t.Error("expected adapter write after opening devtools")
	}

	// Verify the devtools panel is painted at the bottom of the screen.
	// The panel should occupy ~40% of height (8 rows for h=20).
	screen := ta.LastScreen
	if screen == nil {
		t.Fatal("expected screen output")
	}

	// Check that the bottom area has devtools content (tab bar with "Elements").
	panelH := 20 * 4 / 10 // 8 rows
	panelY := 20 - panelH  // row 12
	found := false
	for x := 0; x < 40; x++ {
		cell := screen.Get(x, panelY)
		if cell.Char == 'E' { // "Elements" in tab bar
			found = true
			break
		}
	}
	if !found {
		t.Error("expected devtools tab bar content at panel top")
	}

	writesBefore = ta.WriteCount

	// Press F12 again to close devtools.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F12"})

	if app.DevTools().Visible {
		t.Fatal("devtools should be hidden after second F12")
	}
}

func TestAppV2Engine_DevToolsTabSwitch(t *testing.T) {
	app, ta, _ := newEngineApp(t, 40, 20)

	err := app.RunString(`
		lumina.createComponent({
			id = "root",
			render = function(props)
				return { type = "text", content = "Hello" }
			end
		})
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	// Open devtools.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F12"})
	if !app.DevTools().Visible {
		t.Fatal("devtools should be visible")
	}

	// Default tab is Elements.
	if app.DevTools().ActiveTab != 0 { // TabElements = 0
		t.Errorf("expected Elements tab, got %d", app.DevTools().ActiveTab)
	}

	writesBefore := ta.WriteCount

	// Switch to Perf tab.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "2"})
	if app.DevTools().ActiveTab != 1 { // TabPerf = 1
		t.Errorf("expected Perf tab, got %d", app.DevTools().ActiveTab)
	}

	// Verify something was written (tab switch triggers repaint).
	if ta.WriteCount <= writesBefore {
		t.Error("expected adapter write after tab switch")
	}

	// Switch back to Elements tab.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "1"})
	if app.DevTools().ActiveTab != 0 {
		t.Errorf("expected Elements tab, got %d", app.DevTools().ActiveTab)
	}
}

func TestAppV2Engine_TableAutoFocusKeyboard(t *testing.T) {
	app, _, _ := newEngineApp(t, 80, 24)

	err := app.RunString(`
		lumina.app({
			id = "table-key-app",
			store = { tblIdx = -1 },
			render = function()
				return lumina.createElement("vbox", {
					style = { width = 80, height = 24, background = "#1E1E2E" },
				},
					lumina.createElement("Table", {
						id = "t1",
						columns = {
							{ header = "Name", key = "name", width = 12 },
							{ header = "Role", key = "role", width = 10 },
						},
						rows = {
							{ name = "Alice", role = "Admin" },
							{ name = "Bob", role = "Dev" },
						},
						selectable = true,
						striped = true,
						autoFocus = true,
						onChange = function(i)
							lumina.store.set("tblIdx", i)
						end,
					})
				)
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}
	app.RenderAll()

	if app.Engine().FocusedNode() == nil {
		t.Fatal("expected Table to receive focus via autoFocus")
	}

	app.HandleEvent(&event.Event{Type: "keydown", Key: "j"})
	app.RenderDirty()

	v, ok := app.Store().Get("tblIdx")
	if !ok {
		t.Fatal("onChange should have set tblIdx")
	}
	var idx int
	switch x := v.(type) {
	case int:
		idx = x
	case int64:
		idx = int(x)
	case float64:
		idx = int(x)
	default:
		t.Fatalf("tblIdx type %T value %v", v, v)
	}
	if idx != 1 {
		t.Fatalf("first j should select row 1 (1-based), store got %d", idx)
	}

	app.HandleEvent(&event.Event{Type: "keydown", Key: "j"})
	app.RenderDirty()
	v, _ = app.Store().Get("tblIdx")
	switch x := v.(type) {
	case int:
		idx = x
	case int64:
		idx = int(x)
	case float64:
		idx = int(x)
	default:
		t.Fatalf("tblIdx type %T", v)
	}
	if idx != 2 {
		t.Fatalf("second j should select row 2 (1-based), got %d", idx)
	}
}
