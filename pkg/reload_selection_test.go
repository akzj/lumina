package v2

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/output"
)

// TestHotReload_F5_SelectionHighlight_MovesOnClick reproduces a bug where,
// after F5 full reload, clicking different rows updates state correctly
// (content changes) but the selection highlight (background color) stays
// stuck on the first clicked row instead of moving to the newly clicked one.
//
// Pre-reload: clicking rows correctly moves the highlight.
// Post-reload: clicking rows updates text but highlight is stuck.
//
// The test verifies both the content (text) and the visual style (background)
// to distinguish between "state updated but paint didn't reflect it" vs
// "state itself didn't update".
func TestHotReload_F5_SelectionHighlight_MovesOnClick(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.lua")

	// Script: two rows with selection state. Clicking a row selects it.
	// Selected row gets background "#00FF00", unselected gets "#000000".
	script := `
		local RowComp = lumina.defineComponent("TestRow", function(props)
			local bg = "#000000"
			if props.isSelected then bg = "#00FF00" end
			return lumina.createElement("hbox", {
				style = { height = 1, width = 20 },
				background = bg,
				onClick = function()
					if props.onSelect then props.onSelect() end
				end,
			},
				lumina.createElement("text", {}, props.label or "")
			)
		end)

		lumina.app {
			id = "reload-test",
			render = function()
				local sel, setSel = lumina.useState("sel", "a")
				return lumina.createElement("vbox", {
					style = { width = 40, height = 10 },
				},
					lumina.createElement(RowComp, {
						key = "row-a",
						label = "Row A",
						isSelected = (sel == "a"),
						onSelect = function() setSel("a") end,
					}),
					lumina.createElement(RowComp, {
						key = "row-b",
						label = "Row B",
						isSelected = (sel == "b"),
						onSelect = function() setSel("b") end,
					})
				)
			end,
		}
	`

	if err := os.WriteFile(mainPath, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}

	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.scriptPath = mainPath
	app.RenderAll()

	// Helper: get background color at position (x, y) from the engine's cell buffer.
	getBG := func(x, y int) string {
		return app.engine.Buffer().Get(x, y).BG
	}

	// --- Phase 1: Pre-reload behavior (should work correctly) ---

	// Initial state: sel="a", so Row A (y=0) is selected (#00FF00), Row B (y=1) is not (#000000)
	bgA := getBG(0, 0)
	bgB := getBG(0, 1)
	if bgA != "#00FF00" {
		t.Fatalf("pre-reload initial: Row A bg = %q, want #00FF00", bgA)
	}
	if bgB != "#000000" {
		t.Fatalf("pre-reload initial: Row B bg = %q, want #000000", bgB)
	}

	// Click Row B (y=1) → should select B
	app.HandleEvent(&event.Event{Type: "click", X: 5, Y: 1})
	app.RenderDirty()

	bgA = getBG(0, 0)
	bgB = getBG(0, 1)
	if bgB != "#00FF00" {
		t.Fatalf("pre-reload after click B: Row B bg = %q, want #00FF00", bgB)
	}
	if bgA != "#000000" {
		t.Fatalf("pre-reload after click B: Row A bg = %q, want #000000 (should lose highlight)", bgA)
	}

	// Click Row A (y=0) → should select A again
	app.HandleEvent(&event.Event{Type: "click", X: 5, Y: 0})
	app.RenderDirty()

	bgA = getBG(0, 0)
	bgB = getBG(0, 1)
	if bgA != "#00FF00" {
		t.Fatalf("pre-reload after click A: Row A bg = %q, want #00FF00", bgA)
	}
	if bgB != "#000000" {
		t.Fatalf("pre-reload after click A: Row B bg = %q, want #000000", bgB)
	}

	// --- Phase 2: F5 full reload ---
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F5"})

	// After reload, initial state is sel="a" again
	bgA = getBG(0, 0)
	bgB = getBG(0, 1)
	if bgA != "#00FF00" {
		t.Fatalf("post-reload initial: Row A bg = %q, want #00FF00", bgA)
	}
	if bgB != "#000000" {
		t.Fatalf("post-reload initial: Row B bg = %q, want #000000", bgB)
	}

	// --- Phase 3: Post-reload click behavior (this is where the bug manifests) ---

	// Click Row B (y=1) → should select B, deselect A
	app.HandleEvent(&event.Event{Type: "click", X: 5, Y: 1})
	app.RenderDirty()

	bgA = getBG(0, 0)
	bgB = getBG(0, 1)
	if bgB != "#00FF00" {
		t.Errorf("post-reload after click B: Row B bg = %q, want #00FF00 (highlight should move to B)", bgB)
	}
	if bgA != "#000000" {
		t.Errorf("post-reload after click B: Row A bg = %q, want #000000 (highlight should leave A)", bgA)
	}

	// Click Row A (y=0) → should select A, deselect B
	app.HandleEvent(&event.Event{Type: "click", X: 5, Y: 0})
	app.RenderDirty()

	bgA = getBG(0, 0)
	bgB = getBG(0, 1)
	if bgA != "#00FF00" {
		t.Errorf("post-reload after click A: Row A bg = %q, want #00FF00 (highlight should move to A)", bgA)
	}
	if bgB != "#000000" {
		t.Errorf("post-reload after click A: Row B bg = %q, want #000000 (highlight should leave B)", bgB)
	}
}

// TestHotReload_F5_SelectionHighlight_ScrollView reproduces the bug with a
// ScrollView wrapping the rows — matching the real app's component hierarchy
// (Root → ScrollView → AgentRow). The ScrollView adds an intermediate
// component in the reconciliation chain that may cause stale node references.
//
// Also uses mousedown+mouseup (real terminal event path) and multiple useState
// calls per click handler (matching the real app's onSelectAgent behavior).
func TestHotReload_F5_SelectionHighlight_ScrollView(t *testing.T) {
	app, ta, _ := newV2App(t, 40, 20)

	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.lua")

	// Script with ScrollView wrapping rows + multiple setState per click.
	// This mimics the real app: Root → SessionPage → AgentTree → ScrollView → AgentRow.
	script := `
		local ScrollView = require("lux.scrollview")

		local RowComp = lumina.defineComponent("AgentRow", function(props)
			local bg = "#111111"
			if props.isSelected then bg = "#00FF00" end
			return lumina.createElement("hbox", {
				style = { height = 1, width = 40 },
				background = bg,
				onClick = function()
					if props.onSelect then props.onSelect() end
				end,
			},
				lumina.createElement("text", {}, props.label or "")
			)
		end)

		-- Intermediate container component (like SessionPage → AgentTree)
		local AgentTree = lumina.defineComponent("AgentTree", function(props)
			return lumina.createElement(ScrollView, {
				key = "agent-scroll",
				height = 10,
				style = { width = 40 },
			},
				lumina.createElement(RowComp, {
					key = "agent-a",
					label = "Agent Alpha",
					isSelected = props.selected == "a",
					onSelect = function() props.onSelect("a") end,
				}),
				lumina.createElement(RowComp, {
					key = "agent-b",
					label = "Agent Beta",
					isSelected = props.selected == "b",
					onSelect = function() props.onSelect("b") end,
				}),
				lumina.createElement(RowComp, {
					key = "agent-c",
					label = "Agent Gamma",
					isSelected = props.selected == "c",
					onSelect = function() props.onSelect("c") end,
				})
			)
		end)

		lumina.app {
			id = "scroll-reload-test",
			render = function()
				local sel, setSel = lumina.useState("sel", "a")
				local detail, setDetail = lumina.useState("detail", "none")
				local counter, setCounter = lumina.useState("counter", 0)

				local function onSelectAgent(agentId)
					-- Multiple setState calls per click (like real app)
					setSel(agentId)
					setDetail("detail-" .. agentId)
					setCounter(counter + 1)
				end

				return lumina.createElement("vbox", {
					style = { width = 40, height = 20 },
				},
					lumina.createElement(AgentTree, {
						key = "tree",
						selected = sel,
						onSelect = onSelectAgent,
					}),
					lumina.createElement("text", {
						id = "detail-text",
					}, "Detail: " .. detail .. " sel=" .. sel)
				)
			end,
		}
	`

	if err := os.WriteFile(mainPath, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}
	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.scriptPath = mainPath
	app.RenderAll()

	getBG := func(x, y int) string {
		return app.engine.Buffer().Get(x, y).BG
	}

	// Flush screen for text checks
	flushScreen := func() {
		screen := app.engine.ToBuffer()
		_ = ta.WriteFull(screen)
	}

	// Helper: simulate mousedown + mouseup (real terminal path)
	click := func(x, y int) {
		app.HandleEvent(&event.Event{Type: "mousedown", X: x, Y: y, Button: "left"})
		app.HandleEvent(&event.Event{Type: "mouseup", X: x, Y: y, Button: "left"})
		app.RenderDirty()
	}

	// --- Phase 1: Pre-reload verification ---
	flushScreen()

	// Row A at y=0, Row B at y=1, Row C at y=2 (inside ScrollView)
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Fatalf("pre-reload initial: Row A bg = %q, want #00FF00", bg)
	}
	if bg := getBG(0, 1); bg != "#111111" {
		t.Fatalf("pre-reload initial: Row B bg = %q, want #111111", bg)
	}

	// Click Row B
	click(5, 1)
	flushScreen()

	if bg := getBG(0, 1); bg != "#00FF00" {
		t.Fatalf("pre-reload click B: Row B bg = %q, want #00FF00", bg)
	}
	if bg := getBG(0, 0); bg != "#111111" {
		t.Fatalf("pre-reload click B: Row A bg = %q, want #111111 (should deselect)", bg)
	}
	if !screenHasString(ta, "sel=b") {
		t.Fatalf("pre-reload click B: expected 'sel=b' on screen")
	}

	// Click Row C
	click(5, 2)
	flushScreen()

	if bg := getBG(0, 2); bg != "#00FF00" {
		t.Fatalf("pre-reload click C: Row C bg = %q, want #00FF00", bg)
	}
	if bg := getBG(0, 1); bg != "#111111" {
		t.Fatalf("pre-reload click C: Row B bg = %q, want #111111", bg)
	}

	// --- Phase 2: F5 full reload ---
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F5"})

	// After reload, sel resets to "a"
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Fatalf("post-reload initial: Row A bg = %q, want #00FF00", bg)
	}
	for _, row := range []int{1, 2} {
		if bg := getBG(0, row); bg != "#111111" {
			t.Fatalf("post-reload initial: Row %d bg = %q, want #111111", row, bg)
		}
	}

	// --- Phase 3: Post-reload clicks (bug reproduction) ---

	// Click Row B (y=1)
	click(5, 1)
	flushScreen()

	if bg := getBG(0, 1); bg != "#00FF00" {
		t.Errorf("post-reload click B: Row B bg = %q, want #00FF00 (highlight should move to B)", bg)
	}
	if bg := getBG(0, 0); bg != "#111111" {
		t.Errorf("post-reload click B: Row A bg = %q, want #111111 (should deselect)", bg)
	}
	if !screenHasString(ta, "sel=b") {
		t.Errorf("post-reload click B: expected 'sel=b' on screen (state should update)")
	}

	// Click Row C (y=2) — highlight should MOVE from B to C
	click(5, 2)
	flushScreen()

	if bg := getBG(0, 2); bg != "#00FF00" {
		t.Errorf("post-reload click C: Row C bg = %q, want #00FF00 (highlight should move to C)", bg)
	}
	if bg := getBG(0, 1); bg != "#111111" {
		t.Errorf("post-reload click C: Row B bg = %q, want #111111 (should lose highlight)", bg)
	}
	if !screenHasString(ta, "sel=c") {
		t.Errorf("post-reload click C: expected 'sel=c' on screen")
	}

	// Click Row A (y=0) — highlight should MOVE from C to A
	click(5, 0)
	flushScreen()

	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Errorf("post-reload click A: Row A bg = %q, want #00FF00 (highlight should move to A)", bg)
	}
	for _, row := range []int{1, 2} {
		if bg := getBG(0, row); bg != "#111111" {
			t.Errorf("post-reload click A: Row %d bg = %q, want #111111", row, bg)
		}
	}
	if !screenHasString(ta, "sel=a") {
		t.Errorf("post-reload click A: expected 'sel=a' on screen")
	}
}

// TestHotReload_F5_SelectionHighlight_NestedComponents tests 4-level nesting:
// Root → PageWrapper → ListContainer → Row. This matches the depth of the
// real app (Root → SessionPage → AgentTree → AgentRow) without requiring
// ScrollView, isolating the nesting depth as a variable.
func TestHotReload_F5_SelectionHighlight_NestedComponents(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	ta := output.NewTestAdapter()
	app := NewApp(L, 40, 10, ta)

	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.lua")

	script := `
		local RowComp = lumina.defineComponent("NestRow", function(props)
			local bg = "#222222"
			if props.isSelected then bg = "#00FF00" end
			return lumina.createElement("hbox", {
				style = { height = 1, width = 40 },
				background = bg,
				onClick = function()
					if props.onSelect then props.onSelect() end
				end,
			},
				lumina.createElement("text", {}, props.label or "")
			)
		end)

		-- Level 2: ListContainer wraps rows
		local ListContainer = lumina.defineComponent("ListContainer", function(props)
			local children = {}
			for i, item in ipairs(props.items) do
				children[i] = lumina.createElement(RowComp, {
					key = "row-" .. item.id,
					label = item.label,
					isSelected = (props.selectedId == item.id),
					onSelect = function() props.onSelect(item.id) end,
				})
			end
			return lumina.createElement("vbox", {
				style = { width = 40, height = 8 },
			}, table.unpack(children))
		end)

		-- Level 1: PageWrapper
		local PageWrapper = lumina.defineComponent("PageWrapper", function(props)
			return lumina.createElement("vbox", {
				style = { width = 40, height = 10 },
			},
				lumina.createElement(ListContainer, {
					key = "list",
					items = props.items,
					selectedId = props.selectedId,
					onSelect = props.onSelect,
				}),
				lumina.createElement("text", {}, "selected=" .. (props.selectedId or "nil"))
			)
		end)

		-- Root level
		lumina.app {
			id = "nested-reload-test",
			render = function()
				local sel, setSel = lumina.useState("sel", "x")
				local info, setInfo = lumina.useState("info", "")

				local items = {
					{ id = "x", label = "Item X" },
					{ id = "y", label = "Item Y" },
					{ id = "z", label = "Item Z" },
				}

				return lumina.createElement(PageWrapper, {
					key = "page",
					items = items,
					selectedId = sel,
					onSelect = function(id)
						setSel(id)
						setInfo("clicked-" .. id)
					end,
				})
			end,
		}
	`

	if err := os.WriteFile(mainPath, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}
	if err := app.RunScript(mainPath); err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	app.scriptPath = mainPath
	app.RenderAll()

	getBG := func(x, y int) string {
		return app.engine.Buffer().Get(x, y).BG
	}

	flushScreen := func() {
		screen := app.engine.ToBuffer()
		_ = ta.WriteFull(screen)
	}

	// mousedown+mouseup path (real terminal behavior)
	click := func(x, y int) {
		app.HandleEvent(&event.Event{Type: "mousedown", X: x, Y: y, Button: "left"})
		app.HandleEvent(&event.Event{Type: "mouseup", X: x, Y: y, Button: "left"})
		app.RenderDirty()
	}

	// --- Phase 1: Pre-reload ---
	flushScreen()

	// Row X at y=0 (initially selected), Y at y=1, Z at y=2
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Fatalf("pre-reload: Row X bg = %q, want #00FF00 (initially selected)", bg)
	}

	// Click Row Y
	click(5, 1)
	flushScreen()
	if bg := getBG(0, 1); bg != "#00FF00" {
		t.Fatalf("pre-reload click Y: bg = %q, want #00FF00", bg)
	}
	if bg := getBG(0, 0); bg != "#222222" {
		t.Fatalf("pre-reload click Y: Row X bg = %q, want #222222 (deselected)", bg)
	}
	if !screenHasString(ta, "selected=y") {
		t.Fatalf("pre-reload: expected 'selected=y' on screen")
	}

	// --- Phase 2: F5 reload ---
	app.HandleEvent(&event.Event{Type: "keydown", Key: "F5"})
	flushScreen()

	// After reload, sel resets to "x"
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Fatalf("post-reload initial: Row X bg = %q, want #00FF00", bg)
	}
	if !screenHasString(ta, "selected=x") {
		t.Fatalf("post-reload: expected 'selected=x' on screen")
	}

	// --- Phase 3: Post-reload clicks ---

	// Click Row Y (y=1)
	click(5, 1)
	flushScreen()
	if bg := getBG(0, 1); bg != "#00FF00" {
		t.Errorf("post-reload click Y: bg = %q, want #00FF00", bg)
	}
	if bg := getBG(0, 0); bg != "#222222" {
		t.Errorf("post-reload click Y: Row X bg = %q, want #222222 (should deselect)", bg)
	}
	if !screenHasString(ta, "selected=y") {
		t.Errorf("post-reload: expected 'selected=y' on screen")
	}

	// Click Row Z (y=2) — highlight should MOVE from Y to Z
	click(5, 2)
	flushScreen()
	if bg := getBG(0, 2); bg != "#00FF00" {
		t.Errorf("post-reload click Z: bg = %q, want #00FF00 (highlight should move)", bg)
	}
	if bg := getBG(0, 1); bg != "#222222" {
		t.Errorf("post-reload click Z: Row Y bg = %q, want #222222 (should lose highlight)", bg)
	}
	if !screenHasString(ta, "selected=z") {
		t.Errorf("post-reload: expected 'selected=z' on screen")
	}

	// Click Row X (y=0) — highlight should MOVE from Z to X
	click(5, 0)
	flushScreen()
	if bg := getBG(0, 0); bg != "#00FF00" {
		t.Errorf("post-reload click X: bg = %q, want #00FF00 (highlight should move)", bg)
	}
	for _, row := range []int{1, 2} {
		if bg := getBG(0, row); bg != "#222222" {
			t.Errorf("post-reload click X: Row %d bg = %q, want #222222", row, bg)
		}
	}
}
