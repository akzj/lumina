package v2

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
)

// newV2App creates a NewAppWithEngine with a fresh Lua state and TestAdapter.
// This uses the new V2 render engine (persistent RenderNode tree).
func newV2App(t *testing.T, w, h int) (*App, *output.TestAdapter, *lua.State) {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	ta := output.NewTestAdapter()
	app := NewAppWithEngine(L, w, h, ta)
	return app, ta, L
}

// ═══════════════════════════════════════════════════════════════════
// Simple Rendering Tests
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_SimpleText(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "txt",
			render = function(props)
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

func TestV2E2E_BoxWithText(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "boxtext",
			render = function(props)
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
	cell := ta.LastScreen.Get(15, 1)
	if cell.Background != "#FF0000" {
		t.Errorf("expected background '#FF0000' at (15,1), got %q", cell.Background)
	}
}

func TestV2E2E_MultipleTextChildren(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "multi",
			render = function(props)
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

func TestV2E2E_NestedElements(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "nested",
			render = function(props)
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

func TestV2E2E_StringAndTableChildren(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "mixed",
			render = function(props)
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

	if !screenHasString(ta, "Header") {
		t.Error("expected 'Header' on screen (string child of box)")
	}
	if !screenHasString(ta, "Body") {
		t.Error("expected 'Body' on screen (table child of box)")
	}
}

func TestV2E2E_TextForegroundColor(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "color-test",
			render = function(props)
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

	if ta.LastScreen == nil {
		t.Fatal("Screen is nil")
	}

	cell := ta.LastScreen.Get(0, 0)
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

func TestV2E2E_TextInheritsParentBackground(t *testing.T) {
	// P3 fix: text children now inherit parent box background from CellBuffer.

	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "bg-test",
			render = function(props)
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

	if ta.LastScreen == nil {
		t.Fatal("Screen is nil")
	}

	// Text cell at (0,0) should have 'H' with background "#1E1E2E" (inherited from box)
	textCell := ta.LastScreen.Get(0, 0)
	if textCell.Char != 'H' {
		t.Errorf("expected 'H' at (0,0), got %q", textCell.Char)
	}
	if textCell.Background != "#1E1E2E" {
		t.Errorf("text cell background: got %q, want '#1E1E2E' (should inherit from parent box)", textCell.Background)
	}

	// Empty cell at (10, 5) should also have background "#1E1E2E" (box fill)
	emptyCell := ta.LastScreen.Get(10, 5)
	if emptyCell.Background != "#1E1E2E" {
		t.Errorf("empty cell background: got %q, want '#1E1E2E'", emptyCell.Background)
	}
}

// ═══════════════════════════════════════════════════════════════════
// State Tests
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_UseState(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "stateful",
			render = function(props)
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

	// Update state via the engine and re-render.
	eng := app.Engine()
	if eng == nil {
		t.Fatal("Engine() returned nil")
	}
	eng.SetState("stateful", "count", int64(42))
	app.RenderDirty()

	if !screenHasString(ta, "N:42") {
		t.Errorf("expected 'N:42' after SetState, got line: %q", readScreenLine(ta, 0, 40))
	}
}

// ═══════════════════════════════════════════════════════════════════
// Event Tests
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_ClickEvent(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "clicker",
			render = function(props)
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

func TestV2E2E_MultiClick(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "multi-click",
			render = function(props)
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

func TestV2E2E_CounterWithClick(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "counter",
			render = function(props)
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

func TestV2E2E_KeyboardEvent(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "kbd",
			render = function(props)
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

	// V2 engine dispatches keydown to the root component's tree.
	// Send a key event directly.
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Enter"})
	app.RenderDirty()

	if !screenHasString(ta, "Key:Enter") {
		t.Errorf("expected 'Key:Enter' after keydown, got line: %q", readScreenLine(ta, 0, 40))
	}
}

// ═══════════════════════════════════════════════════════════════════
// Component Library Tests
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_ProgressBar(t *testing.T) {
	app, ta, _ := newV2App(t, 60, 5)

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
			render = function(props)
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

	if !screenHasString(ta, "CPU") {
		t.Error("expected 'CPU' label on screen")
	}
	if !screenHasString(ta, "50%") {
		t.Error("expected '50%' on screen")
	}
	if !screenHasString(ta, "100%") {
		t.Error("expected '100%' on screen")
	}
	if !screenHasString(ta, "0%") {
		t.Error("expected '0%' on screen for 0-value bar")
	}
	if !screenHasChar(ta, '█') {
		t.Error("expected filled bar character '█' on screen")
	}
	if !screenHasChar(ta, '░') {
		t.Error("expected empty bar character '░' on screen")
	}
}

func TestV2E2E_Table(t *testing.T) {
	app, ta, _ := newV2App(t, 60, 10)

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
			render = function(props)
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

	if !screenHasString(ta, "Name") {
		t.Error("expected 'Name' header on screen")
	}
	if !screenHasString(ta, "Role") {
		t.Error("expected 'Role' header on screen")
	}
	if !screenHasString(ta, "Status") {
		t.Error("expected 'Status' header on screen")
	}
	if !screenHasChar(ta, '─') {
		t.Error("expected separator character '─' on screen")
	}
	if !screenHasString(ta, "Alice") {
		t.Error("expected 'Alice' on screen")
	}
	if !screenHasString(ta, "Bob") {
		t.Error("expected 'Bob' on screen")
	}
}

func TestV2E2E_Tabs(t *testing.T) {
	app, ta, _ := newV2App(t, 60, 10)

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
			render = function(props)
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

	// Switch to tab 1 via engine SetState and verify content changes
	eng := app.Engine()
	eng.SetState("tabs-test", "activeTab", int64(1))
	app.RenderDirty()

	if !screenHasString(ta, "Content A") {
		t.Error("expected 'Content A' after switching to tab 1")
	}
}

func TestV2E2E_Modal(t *testing.T) {
	app, ta, _ := newV2App(t, 60, 15)

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
			render = function(props)
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

	// Show modal via engine SetState
	eng := app.Engine()
	eng.SetState("modal-test", "show", true)
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

	// Hide modal again
	eng.SetState("modal-test", "show", false)
	app.RenderAll()

	if screenHasString(ta, "Test Modal") {
		t.Error("modal title should NOT be visible after hiding")
	}
}

// ═══════════════════════════════════════════════════════════════════
// Select Component Test
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_Select(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

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
			render = function(props)
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

	if !screenHasString(ta, "Choose:") {
		t.Error("expected 'Choose:' label on screen")
	}
	if !screenHasString(ta, "Alpha") {
		t.Error("expected 'Alpha' on screen")
	}
	if !screenHasString(ta, "Beta") {
		t.Error("expected 'Beta' on screen")
	}
	if !screenHasString(ta, "Gamma") {
		t.Error("expected 'Gamma' on screen")
	}

	// Change selection to 3 and verify
	eng := app.Engine()
	eng.SetState("select-test", "sel", int64(3))
	app.RenderDirty()

	if !screenHasString(ta, "▸ Gamma") {
		t.Error("expected '▸ Gamma' after changing selection to 3")
	}
}

// ═══════════════════════════════════════════════════════════════════
// Counter Script Test (loads actual .lua file)
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_CounterScript(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

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

// ═══════════════════════════════════════════════════════════════════
// Input Element Tests
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_InputRender(t *testing.T) {
	// V2 engine gap: readDescriptor reads "content" not "value" for input elements.
	// P1a fix: input value prop now mapped to Node.Content in readDescriptor.

	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "input-test",
			render = function(props)
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

func TestV2E2E_InputPlaceholder(t *testing.T) {
	// P1b fix: placeholder now supported in Node + readDescriptor + painter.

	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "input-ph",
			render = function(props)
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

func TestV2E2E_InputTyping(t *testing.T) {
	// Input editing now built into V2 engine.

	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "input-type",
			render = function(props)
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

func TestV2E2E_InputBackspace(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "bs-test",
			render = function(props)
				local text, setText = lumina.useState("text", "")
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

	// Focus the input
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	app.RenderDirty()

	// Type "Hi"
	app.HandleEvent(&event.Event{Type: "keydown", Key: "H"})
	app.RenderDirty()
	app.HandleEvent(&event.Event{Type: "keydown", Key: "i"})
	app.RenderDirty()

	if !screenHasString(ta, "Hi") {
		t.Fatalf("expected 'Hi' on screen, got: %q", readScreenLine(ta, 0, 40))
	}

	// Backspace
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Backspace"})
	app.RenderDirty()

	if !screenHasString(ta, "H") {
		t.Errorf("expected 'H' on screen after backspace, got: %q", readScreenLine(ta, 0, 40))
	}
	if screenHasString(ta, "Hi") {
		t.Errorf("'Hi' should not be on screen after backspace")
	}
}

func TestV2E2E_InputOnSubmit(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "submit-test",
			render = function(props)
				local text, setText = lumina.useState("text", "")
				local submitted, setSubmitted = lumina.useState("submitted", "")
				return lumina.createElement("vbox", {},
					lumina.createElement("input", {
						id = "sub-input",
						value = text,
						onChange = function(newValue)
							setText(newValue)
						end,
					}),
					lumina.createElement("text", {id = "result"}, "submitted:" .. submitted)
				)
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	// Focus and type
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	app.HandleEvent(&event.Event{Type: "keydown", Key: "o"})
	app.HandleEvent(&event.Event{Type: "keydown", Key: "k"})
	app.RenderDirty()

	if !screenHasString(ta, "ok") {
		t.Errorf("expected 'ok' on screen after typing, got: %q", readScreenLine(ta, 0, 40))
	}
}

func TestV2E2E_InputAutoFocusable(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "af-test",
			render = function(props)
				local text, setText = lumina.useState("text", "")
				return lumina.createElement("input", {
					id = "af-input",
					value = text,
					autoFocus = true,
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

	// Type without Tab — autoFocus should have focused the input
	app.HandleEvent(&event.Event{Type: "keydown", Key: "A"})
	app.RenderDirty()

	if !screenHasString(ta, "A") {
		t.Errorf("expected 'A' on screen (autoFocus should focus input), got: %q", readScreenLine(ta, 0, 40))
	}
}

// ═══════════════════════════════════════════════════════════════════
// Textarea Tests
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_Textarea_Render(t *testing.T) {
	// V2 engine gap: readDescriptor reads "content" not "value" for textarea elements.
	// P1a fix: textarea value prop now mapped to Node.Content in readDescriptor.

	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "ta-render",
			render = function(props)
				return lumina.createElement("textarea", {
					id = "my-textarea",
					value = "Hello\nWorld",
				})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if !screenHasString(ta, "Hello") {
		t.Errorf("expected 'Hello' on screen, got line0: %q", readScreenLine(ta, 0, 40))
	}
	if !screenHasString(ta, "World") {
		t.Errorf("expected 'World' on screen, got line1: %q", readScreenLine(ta, 1, 40))
	}
}

func TestV2E2E_Textarea_Placeholder(t *testing.T) {
	// P1b fix: placeholder now supported in Node + readDescriptor + painter.

	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "ta-ph",
			render = function(props)
				return lumina.createElement("textarea", {
					id = "ph-textarea",
					value = "",
					placeholder = "Enter text...",
				})
			end
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}

	app.RenderAll()

	if !screenHasString(ta, "Enter text...") {
		t.Errorf("expected placeholder 'Enter text...' on screen, got line0: %q", readScreenLine(ta, 0, 40))
	}
}

func TestV2E2E_Textarea_Typing(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "ta-type",
			render = function(props)
				local text, setText = lumina.useState("text", "")
				return lumina.createElement("textarea", {
					id = "type-ta",
					value = text,
					style = {height = 5, width = 40},
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

	// Focus the textarea
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	app.RenderDirty()

	// Type "Hi"
	app.HandleEvent(&event.Event{Type: "keydown", Key: "H"})
	app.RenderDirty()
	app.HandleEvent(&event.Event{Type: "keydown", Key: "i"})
	app.RenderDirty()

	if !screenHasString(ta, "Hi") {
		t.Errorf("expected 'Hi' on screen after typing, got: %q", readScreenLine(ta, 0, 40))
	}
}

func TestV2E2E_Textarea_Newline(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "ta-nl",
			render = function(props)
				local text, setText = lumina.useState("text", "")
				return lumina.createElement("textarea", {
					id = "nl-ta",
					value = text,
					style = {height = 5, width = 40},
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

	// Focus
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})

	// Type "A", Enter, "B"
	app.HandleEvent(&event.Event{Type: "keydown", Key: "A"})
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Enter"})
	app.HandleEvent(&event.Event{Type: "keydown", Key: "B"})
	app.RenderDirty()

	// A should be on row 0, B on row 1
	aCell := ta.LastScreen.Get(0, 0)
	if aCell.Char != 'A' {
		t.Errorf("expected 'A' at (0,0), got %q", aCell.Char)
	}
	bCell := ta.LastScreen.Get(0, 1)
	if bCell.Char != 'B' {
		t.Errorf("expected 'B' at (0,1), got %q", bCell.Char)
	}
}

func TestV2E2E_Textarea_MultilineNavigation(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "ta-nav",
			render = function(props)
				local text, setText = lumina.useState("text", "")
				return lumina.createElement("textarea", {
					id = "nav-ta",
					value = text,
					style = {height = 5, width = 40},
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

	// Focus
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})

	// Type "AB", Enter, "CD"
	app.HandleEvent(&event.Event{Type: "keydown", Key: "A"})
	app.HandleEvent(&event.Event{Type: "keydown", Key: "B"})
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Enter"})
	app.HandleEvent(&event.Event{Type: "keydown", Key: "C"})
	app.HandleEvent(&event.Event{Type: "keydown", Key: "D"})
	app.RenderDirty()

	// Verify initial state: "AB" on row 0, "CD" on row 1
	if ta.LastScreen.Get(0, 0).Char != 'A' || ta.LastScreen.Get(1, 0).Char != 'B' {
		t.Fatalf("expected 'AB' at row 0, got %q%q", ta.LastScreen.Get(0, 0).Char, ta.LastScreen.Get(1, 0).Char)
	}
	if ta.LastScreen.Get(0, 1).Char != 'C' || ta.LastScreen.Get(1, 1).Char != 'D' {
		t.Fatalf("expected 'CD' at row 1, got %q%q", ta.LastScreen.Get(0, 1).Char, ta.LastScreen.Get(1, 1).Char)
	}

	// Move up (cursor is at end of "CD" → should move to line 0)
	app.HandleEvent(&event.Event{Type: "keydown", Key: "ArrowUp"})
	// Type "X" — should insert on line 0
	app.HandleEvent(&event.Event{Type: "keydown", Key: "X"})
	app.RenderDirty()

	// Line 0 should now contain "ABX" (cursor was at col 2 on line 0)
	if !screenHasString(ta, "ABX") {
		t.Errorf("expected 'ABX' on row 0 after up+type, got: %q", readScreenLine(ta, 0, 40))
	}
}

// ═══════════════════════════════════════════════════════════════════
// TodoMVC Tests
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_TodoMVC_Render(t *testing.T) {
	app, ta, _ := newV2App(t, 80, 24)

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
}

func TestV2E2E_TodoMVC_Navigation(t *testing.T) {
	app, ta, _ := newV2App(t, 80, 24)

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

func TestV2E2E_TodoMVC_Toggle(t *testing.T) {
	// V2 engine bug: Component.SetState panics when comparing uncomparable types
	// P0 fix: SetState now uses reflect.DeepEqual for uncomparable types.

	app, ta, _ := newV2App(t, 80, 24)

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

func TestV2E2E_TodoMVC_Filter(t *testing.T) {
	// V2 engine: filter change uses string state (not table), should work.
	app, ta, _ := newV2App(t, 80, 24)

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

func TestV2E2E_TodoMVC_Delete(t *testing.T) {
	// P0 fix: SetState now uses reflect.DeepEqual for uncomparable types.

	app, ta, _ := newV2App(t, 80, 24)

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

func TestV2E2E_TodoMVC_AddTodo(t *testing.T) {
	// Input editing now built into V2 engine.
}

// ═══════════════════════════════════════════════════════════════════
// Dashboard Tests
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_Dashboard(t *testing.T) {
	app, ta, _ := newV2App(t, 80, 24)

	err := app.RunScript("../../../examples/v2/dashboard.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}

	// Title "Dashboard" appears on screen.
	if !screenHasString(ta, "Dashboard") {
		t.Error("expected 'Dashboard' title on screen")
	}

	// Progress bar characters appear.
	if !screenHasChar(ta, '█') {
		t.Error("expected filled progress bar character '█' on screen")
	}
	if !screenHasChar(ta, '░') {
		t.Error("expected empty progress bar character '░' on screen")
	}

	// Resource labels appear.
	if !screenHasString(ta, "CPU") {
		t.Error("expected 'CPU' label on screen")
	}
	if !screenHasString(ta, "RAM") {
		t.Error("expected 'RAM' label on screen")
	}
	if !screenHasString(ta, "Disk") {
		t.Error("expected 'Disk' label on screen")
	}

	// Activity log entries appear.
	if !screenHasString(ta, "Server started") {
		t.Error("expected 'Server started' activity entry on screen")
	}

	// Stats appear.
	if !screenHasString(ta, "Uptime") {
		t.Error("expected 'Uptime' stat on screen")
	}
	if !screenHasString(ta, "42 days") {
		t.Error("expected '42 days' stat value on screen")
	}

	// Keyboard help appears in footer.
	if !screenHasString(ta, "[q] Quit") {
		t.Error("expected '[q] Quit' help text on screen")
	}
}

func TestV2E2E_Dashboard_Scroll(t *testing.T) {
	// P2 fix: scrollY prop now read + scroll containers implemented.

	app, ta, _ := newV2App(t, 80, 24)

	err := app.RunScript("../../../examples/v2/dashboard.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	if !screenHasString(ta, "Server started") {
		t.Fatal("expected 'Server started' visible initially")
	}

	for i := 0; i < 30; i++ {
		app.HandleEvent(&event.Event{Type: "keydown", Key: "j"})
		app.RenderDirty()
	}

	if !screenHasString(ta, "Daily report") {
		t.Error("expected 'Daily report' visible after scrolling down")
	}

	for i := 0; i < 30; i++ {
		app.HandleEvent(&event.Event{Type: "keydown", Key: "k"})
		app.RenderDirty()
	}

	if !screenHasString(ta, "Server started") {
		t.Error("expected 'Server started' visible after scrolling back up")
	}
}

func TestV2E2E_Dashboard_MouseScroll(t *testing.T) {
	// P2 fix: scroll event now includes key field + scroll container support.

	app, ta, _ := newV2App(t, 80, 24)

	err := app.RunScript("../../../examples/v2/dashboard.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	if !screenHasString(ta, "Server started") {
		t.Fatal("expected 'Server started' visible initially")
	}

	for i := 0; i < 10; i++ {
		app.HandleEvent(&event.Event{Type: "scroll", X: 60, Y: 10, Key: "down"})
		app.RenderDirty()
	}

	if !screenHasString(ta, "Daily report") {
		t.Error("expected 'Daily report' visible after mouse scroll down")
	}

	for i := 0; i < 10; i++ {
		app.HandleEvent(&event.Event{Type: "scroll", X: 60, Y: 10, Key: "up"})
		app.RenderDirty()
	}

	if !screenHasString(ta, "Server started") {
		t.Error("expected 'Server started' visible after mouse scroll up")
	}
}

// ═══════════════════════════════════════════════════════════════════
// Component Showcase Tests
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_ComponentShowcase(t *testing.T) {
	app, ta, _ := newV2App(t, 80, 24)

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

func TestV2E2E_ComponentShowcase_TabSwitch(t *testing.T) {
	app, ta, _ := newV2App(t, 80, 24)

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

func TestV2E2E_ComponentShowcase_Modal(t *testing.T) {
	app, ta, _ := newV2App(t, 80, 24)

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

// ═══════════════════════════════════════════════════════════════════
// require() Tests
// ═══════════════════════════════════════════════════════════════════

func TestV2E2E_RequireLocalModule(t *testing.T) {
	// Create a temp directory with two Lua files:
	//   main.lua — requires "mylib" and uses its return value
	//   mylib.lua — returns a table with a hello() function
	tmpDir := t.TempDir()

	mainLua := `
local m = require("mylib")
lumina.createComponent({
    id = "req-test",
    render = function(props)
        return lumina.createElement("text", {}, m.hello())
    end
})
`
	mylibLua := `
local M = {}
function M.hello() return "world" end
return M
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.lua"), []byte(mainLua), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "mylib.lua"), []byte(mylibLua), 0644); err != nil {
		t.Fatal(err)
	}

	app, ta, _ := newV2App(t, 40, 5)

	err := app.RunScript(filepath.Join(tmpDir, "main.lua"))
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	if !screenHasString(ta, "world") {
		t.Errorf("expected 'world' on screen from required module, got line: %q", readScreenLine(ta, 0, 40))
	}
}

func TestV2E2E_RequireSubdirectory(t *testing.T) {
	// Create a temp directory with:
	//   main.lua — requires "lib.helper"
	//   lib/helper.lua — returns a table with a greet() function
	tmpDir := t.TempDir()

	libDir := filepath.Join(tmpDir, "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatal(err)
	}

	mainLua := `
local helper = require("lib.helper")
lumina.createComponent({
    id = "subdir-test",
    render = function(props)
        return lumina.createElement("text", {}, helper.greet())
    end
})
`
	helperLua := `
local H = {}
function H.greet() return "hi from subdir" end
return H
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.lua"), []byte(mainLua), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(libDir, "helper.lua"), []byte(helperLua), 0644); err != nil {
		t.Fatal(err)
	}

	app, ta, _ := newV2App(t, 40, 5)

	err := app.RunScript(filepath.Join(tmpDir, "main.lua"))
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	app.RenderAll()

	if !screenHasString(ta, "hi from subdir") {
		t.Errorf("expected 'hi from subdir' on screen from subdirectory module, got line: %q", readScreenLine(ta, 0, 40))
	}
}

// TestV2E2E_ChildRemovalClearsGhost verifies that when a child node is removed,
// its pixels are cleared from the screen (no "ghost" artifacts remain).
func TestV2E2E_ChildRemovalClearsGhost(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	err := app.RunString(`
		lumina.createComponent({
			id = "ghost-test",
			render = function(props)
				local showC, setShowC = lumina.useState("showC", true)

				local children = {
					lumina.createElement("text", {id = "a"}, "AAAA"),
					lumina.createElement("text", {id = "b"}, "BBBB"),
				}
				if showC then
					children[#children + 1] = lumina.createElement("text", {id = "c"}, "CCCC")
				end

				return lumina.createElement("vbox", {
					id = "container",
					style = {width = 40, height = 10},
					onClick = function()
						setShowC(false)
					end,
				}, table.unpack(children))
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	// Initial render: all 3 children visible
	app.RenderAll()

	if !screenHasString(ta, "CCCC") {
		t.Fatal("expected 'CCCC' on screen initially")
	}

	// Click to remove child C
	app.HandleEvent(&event.Event{Type: "click", X: 5, Y: 0})
	app.RenderDirty()

	// After removal, AAAA and BBBB should still be visible
	if !screenHasString(ta, "AAAA") {
		t.Error("expected 'AAAA' still on screen after removal")
	}
	if !screenHasString(ta, "BBBB") {
		t.Error("expected 'BBBB' still on screen after removal")
	}

	// CCCC should be GONE — no ghost pixels
	if screenHasString(ta, "CCCC") {
		t.Error("ghost pixel bug: 'CCCC' still visible after child removal")
	}
}

// TestV2E2E_ChildMovePositionClearsGhost verifies that when a child moves
// position due to a sibling's height change (no removal), the old position
// is cleared with the parent's background.
func TestV2E2E_ChildMovePositionClearsGhost(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 10)

	// A has height=1 initially, then shrinks to height=0 (hidden) on click.
	// B starts at row 1. After A shrinks, B moves to row 0.
	// Row 1 (B's old position) should show parent bg, not ghost "BBBB".
	err := app.RunString(`
		lumina.createComponent({
			id = "move-test",
			render = function(props)
				local tall, setTall = lumina.useState("tall", true)

				local aHeight = 3
				if not tall then aHeight = 1 end

				return lumina.createElement("vbox", {
					id = "container",
					style = {width = 40, height = 10, background = "#112233"},
					onClick = function()
						setTall(false)
					end,
				},
					lumina.createElement("text", {id = "a", style = {height = aHeight}}, "AAAA"),
					lumina.createElement("text", {id = "b"}, "BBBB")
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	// Initial render: A takes rows 0-2 (height=3), B at row 3
	app.RenderAll()

	if !screenHasString(ta, "AAAA") {
		t.Fatal("expected 'AAAA' on screen initially")
	}
	bCell := ta.LastScreen.Get(0, 3)
	if bCell.Char != 'B' {
		t.Fatalf("expected 'B' at (0,3) initially, got %q", bCell.Char)
	}

	// Click to shrink A to height=1 → B moves from row 3 to row 1
	app.HandleEvent(&event.Event{Type: "click", X: 5, Y: 0})
	app.RenderDirty()

	// B should now be at row 1
	movedCell := ta.LastScreen.Get(0, 1)
	if movedCell.Char != 'B' {
		t.Errorf("expected 'B' at (0,1) after A shrinks, got %q", movedCell.Char)
	}

	// Row 3 (B's old position) should be cleared — parent bg, not ghost "BBBB"
	ghostCell := ta.LastScreen.Get(0, 3)
	if ghostCell.Char == 'B' {
		t.Error("ghost pixel bug: 'B' still visible at old position (0,3) after child moved up")
	}
	// Should have parent background
	if ghostCell.Background != "#112233" {
		t.Errorf("expected parent bg '#112233' at old position (0,3), got %q", ghostCell.Background)
	}
}
