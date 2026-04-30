package v2

import (
	"fmt"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/output"
)

// ═══════════════════════════════════════════════════════════════════════
// Lux Component Benchmarks
//
// These benchmarks measure the performance of Lux library components
// (DataGrid, Form, Toast, Tree, Theme) at realistic scales.
// ═══════════════════════════════════════════════════════════════════════

// --- Helpers ---

// newLuxBenchApp creates an App with nopAdapter for Lux component benchmarks.
func newLuxBenchApp(b *testing.B, w, h int) *App {
	b.Helper()
	L := lua.NewState()
	b.Cleanup(func() { L.Close() })
	app := NewApp(L, w, h, nopAdapter{})
	return app
}

// newLuxBenchAppWithTA creates an App with TestAdapter for verification.
func newLuxBenchAppWithTA(tb testing.TB, w, h int) (*App, *output.TestAdapter) {
	tb.Helper()
	L := lua.NewState()
	tb.Cleanup(func() { L.Close() })
	ta := output.NewTestAdapter()
	app := NewApp(L, w, h, ta)
	return app, ta
}

// ═══════════════════════════════════════════════════════════════════════
// DataGrid with 1000 rows (virtual scroll)
// ═══════════════════════════════════════════════════════════════════════

const luxDataGridScript = `
local lux = require("lux")

local rows = {}
for i = 1, 1000 do
    rows[i] = {
        name = "Row " .. i,
        value = tostring(i * 100),
        status = i % 2 == 0 and "active" or "inactive",
    }
end

lumina.app {
    id = "bench-grid",
    store = { idx = 1 },
    render = function()
        local idx = lumina.useStore("idx")
        return lux.DataGrid {
            id = "grid",
            width = 80,
            height = 38,
            columns = {
                { id = "name", header = "Name", width = 30, key = "name" },
                { id = "value", header = "Value", width = 25, key = "value" },
                { id = "status", header = "Status", width = 25, key = "status" },
            },
            rows = rows,
            selectedIndex = idx,
            onChangeIndex = function(i) lumina.store.set("idx", i) end,
            virtualScroll = true,
            autoFocus = true,
        }
    end,
}
`

// TestLux_DataGrid_1000Rows_Renders verifies the DataGrid loads and renders.
func TestLux_DataGrid_1000Rows_Renders(t *testing.T) {
	app, ta := newLuxBenchAppWithTA(t, 80, 40)
	if err := app.RunString(luxDataGridScript); err != nil {
		t.Fatal(err)
	}
	app.RenderAll()
	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}
	if !screenHasString(ta, "Name") {
		t.Error("expected 'Name' header on screen")
	}
}

// BenchmarkLux_DataGrid_1000Rows_Scroll measures re-render on scroll navigation.
// With virtualScroll, only ~38 visible rows are rendered per frame.
func BenchmarkLux_DataGrid_1000Rows_Scroll(b *testing.B) {
	app := newLuxBenchApp(b, 80, 40)
	if err := app.RunString(luxDataGridScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate pressing "j" (ArrowDown) to scroll through the grid
		app.HandleEvent(&event.Event{Type: "keydown", Key: "j"})
		app.RenderDirty()
	}
}

// BenchmarkLux_DataGrid_1000Rows_Scroll_Memory measures allocations during scroll.
func BenchmarkLux_DataGrid_1000Rows_Scroll_Memory(b *testing.B) {
	app := newLuxBenchApp(b, 80, 40)
	if err := app.RunString(luxDataGridScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.HandleEvent(&event.Event{Type: "keydown", Key: "j"})
		app.RenderDirty()
	}
}

// BenchmarkLux_DataGrid_1000Rows_RenderAll measures full render cost.
func BenchmarkLux_DataGrid_1000Rows_RenderAll(b *testing.B) {
	app := newLuxBenchApp(b, 80, 40)
	if err := app.RunString(luxDataGridScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i%1000 + 1
		if err := app.RunString(fmt.Sprintf(`lumina.store.set("idx", %d)`, idx)); err != nil {
			b.Fatal(err)
		}
		app.RenderAll()
	}
}

// ═══════════════════════════════════════════════════════════════════════
// DataGrid Edit Cell
// ═══════════════════════════════════════════════════════════════════════

const luxDataGridEditScript = `
local lux = require("lux")

local rows = {}
for i = 1, 100 do
    rows[i] = { name = "Item " .. i, price = tostring(i * 10) }
end

lumina.app {
    id = "bench-edit",
    store = {
        idx = 1,
        editCell = nil,
        editValue = nil,
    },
    render = function()
        local idx = lumina.useStore("idx")
        local editCell = lumina.useStore("editCell")
        local editValue = lumina.useStore("editValue")
        return lux.DataGrid {
            id = "editgrid",
            width = 60,
            height = 20,
            columns = {
                { id = "name", header = "Name", width = 30, key = "name" },
                { id = "price", header = "Price", width = 20, key = "price" },
            },
            rows = rows,
            selectedIndex = idx,
            onChangeIndex = function(i) lumina.store.set("idx", i) end,
            editable = true,
            editingCell = editCell,
            editValue = editValue,
            onEditStart = function(row, col)
                lumina.store.set("editCell", { rowIndex = row, columnId = col })
                lumina.store.set("editValue", nil)
            end,
            onEditValueChange = function(text)
                lumina.store.set("editValue", text)
            end,
            onCellChange = function(row, col, val)
                lumina.store.set("editCell", nil)
                lumina.store.set("editValue", nil)
            end,
            onEditCancel = function(row, col)
                lumina.store.set("editCell", nil)
                lumina.store.set("editValue", nil)
            end,
            autoFocus = true,
        }
    end,
}
`

// BenchmarkLux_DataGrid_EditCell measures entering/exiting edit mode.
func BenchmarkLux_DataGrid_EditCell(b *testing.B) {
	app := newLuxBenchApp(b, 60, 22)
	if err := app.RunString(luxDataGridEditScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Enter edit mode
		app.HandleEvent(&event.Event{Type: "keydown", Key: "Enter"})
		app.RenderDirty()
		// Exit edit mode (Escape)
		app.HandleEvent(&event.Event{Type: "keydown", Key: "Escape"})
		app.RenderDirty()
	}
}

// ═══════════════════════════════════════════════════════════════════════
// Form with 10 fields
// ═══════════════════════════════════════════════════════════════════════

const luxFormScript = `
local lux = require("lux")

local fields = {}
for i = 1, 10 do
    fields[i] = {
        id = "field" .. i,
        type = "text",
        label = "Field " .. i,
        placeholder = "Enter value " .. i,
        required = (i <= 3),
    }
end

lumina.app {
    id = "bench-form",
    store = {
        values = {},
        errors = {},
    },
    render = function()
        local values = lumina.useStore("values")
        local errors = lumina.useStore("errors")
        return lux.Form {
            id = "form",
            width = 50,
            fields = fields,
            values = values or {},
            errors = errors or {},
            onFieldChange = function(fieldId, val)
                local v = lumina.store.get("values") or {}
                v[fieldId] = val
                lumina.store.set("values", v)
            end,
            onSubmit = function(vals) end,
            onReset = function()
                lumina.store.set("values", {})
                lumina.store.set("errors", {})
            end,
        }
    end,
}
`

// TestLux_Form_10Fields_Renders verifies the form renders correctly.
func TestLux_Form_10Fields_Renders(t *testing.T) {
	app, ta := newLuxBenchAppWithTA(t, 60, 40)
	if err := app.RunString(luxFormScript); err != nil {
		t.Fatal(err)
	}
	app.RenderAll()
	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}
	if !screenHasString(ta, "Field 1") {
		t.Error("expected 'Field 1' label on screen")
	}
}

// BenchmarkLux_Form_10Fields_RenderAll measures full render of a 10-field form.
func BenchmarkLux_Form_10Fields_RenderAll(b *testing.B) {
	app := newLuxBenchApp(b, 60, 40)
	if err := app.RunString(luxFormScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate field change via Lua store API → triggers re-render
		code := fmt.Sprintf(`lumina.store.set("values", { field1 = "value%d" })`, i)
		if err := app.RunString(code); err != nil {
			b.Fatal(err)
		}
		app.RenderAll()
	}
}

// BenchmarkLux_Form_10Fields_RenderDirty measures incremental render on field change.
func BenchmarkLux_Form_10Fields_RenderDirty(b *testing.B) {
	app := newLuxBenchApp(b, 60, 40)
	if err := app.RunString(luxFormScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code := fmt.Sprintf(`lumina.store.set("values", { field1 = "val%d" })`, i)
		if err := app.RunString(code); err != nil {
			b.Fatal(err)
		}
		app.RenderDirty()
	}
}

// ═══════════════════════════════════════════════════════════════════════
// Theme Switching
// ═══════════════════════════════════════════════════════════════════════

const luxThemeSwitchScript = `
local lux = require("lux")

lumina.app {
    id = "bench-theme",
    store = { theme = "mocha" },
    render = function()
        local children = {}
        -- Mix of themed components
        children[#children + 1] = lux.Card {
            key = "card1",
            title = "Dashboard",
            width = 40,
            height = 6,
            children = {
                lumina.createElement("text", { key = "t1" }, "Content here"),
            },
        }
        children[#children + 1] = lux.Alert {
            key = "alert1",
            variant = "info",
            title = "Notice",
            message = "This is a themed alert",
            width = 40,
        }
        children[#children + 1] = lux.Badge {
            key = "badge1",
            text = "Active",
            variant = "success",
        }
        children[#children + 1] = lux.Progress {
            key = "prog1",
            value = 65,
            width = 30,
        }
        return lumina.createElement("vbox", {
            style = { width = 50, height = 20 },
        }, table.unpack(children))
    end,
}
`

// BenchmarkLux_ThemeSwitch measures theme switching + full re-render.
func BenchmarkLux_ThemeSwitch(b *testing.B) {
	app := newLuxBenchApp(b, 50, 20)
	if err := app.RunString(luxThemeSwitchScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	themes := []string{"mocha", "latte", "nord", "dracula"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		theme := themes[i%len(themes)]
		// Switch theme via Lua API
		if err := app.RunString(fmt.Sprintf(`lumina.setTheme("%s")`, theme)); err != nil {
			b.Fatal(err)
		}
		app.RenderAll()
	}
}

// BenchmarkLux_ThemeSwitch_Memory measures allocations during theme switch.
func BenchmarkLux_ThemeSwitch_Memory(b *testing.B) {
	app := newLuxBenchApp(b, 50, 20)
	if err := app.RunString(luxThemeSwitchScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	themes := []string{"mocha", "latte", "nord", "dracula"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		theme := themes[i%len(themes)]
		if err := app.RunString(fmt.Sprintf(`lumina.setTheme("%s")`, theme)); err != nil {
			b.Fatal(err)
		}
		app.RenderAll()
	}
}

// ═══════════════════════════════════════════════════════════════════════
// Toast Stack (5 notifications)
// ═══════════════════════════════════════════════════════════════════════

const luxToastScript = `
local lux = require("lux")

lumina.app {
    id = "bench-toast",
    store = {
        toasts = {
            { id = 1, message = "File saved successfully", variant = "success" },
            { id = 2, message = "Network connection lost", variant = "error" },
            { id = 3, message = "Update available", variant = "info" },
            { id = 4, message = "Battery low", variant = "warning" },
            { id = 5, message = "Download complete", variant = "success" },
        },
    },
    render = function()
        local toasts = lumina.useStore("toasts")
        return lux.Toast {
            id = "toasts",
            items = toasts or {},
            maxVisible = 5,
            width = 45,
            onDismiss = function(id)
                local items = lumina.store.get("toasts") or {}
                local newItems = {}
                for _, item in ipairs(items) do
                    if item.id ~= id then
                        newItems[#newItems + 1] = item
                    end
                end
                lumina.store.set("toasts", newItems)
            end,
        }
    end,
}
`

// TestLux_Toast_Renders verifies toast stack renders.
func TestLux_Toast_Renders(t *testing.T) {
	app, ta := newLuxBenchAppWithTA(t, 50, 15)
	if err := app.RunString(luxToastScript); err != nil {
		t.Fatal(err)
	}
	app.RenderAll()
	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}
	if !screenHasString(ta, "File saved") {
		t.Error("expected toast message on screen")
	}
}

// BenchmarkLux_Toast_Stack_RenderDirty measures re-render when toasts change.
func BenchmarkLux_Toast_Stack_RenderDirty(b *testing.B) {
	app := newLuxBenchApp(b, 50, 15)
	if err := app.RunString(luxToastScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Add a new toast and re-render
		code := fmt.Sprintf(`
			local items = lumina.store.get("toasts") or {}
			items[#items + 1] = { id = %d, message = "Toast %d", variant = "info" }
			lumina.store.set("toasts", items)
		`, i+100, i+100)
		if err := app.RunString(code); err != nil {
			b.Fatal(err)
		}
		app.RenderDirty()
	}
}

// BenchmarkLux_Toast_Stack_RenderAll measures full render of toast stack.
func BenchmarkLux_Toast_Stack_RenderAll(b *testing.B) {
	app := newLuxBenchApp(b, 50, 15)
	if err := app.RunString(luxToastScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.RenderAll()
	}
}

// ═══════════════════════════════════════════════════════════════════════
// Tree with deep nesting (100 nodes, 5 levels)
// ═══════════════════════════════════════════════════════════════════════

const luxTreeScript = `
local lux = require("lux")

-- Generate a tree: 5 levels, 4 children per node = ~340 total nodes
local function buildTree(prefix, depth, maxDepth)
    local items = {}
    local count = depth == 0 and 5 or 4
    for i = 1, count do
        local id = prefix .. "-" .. i
        local node = { id = id, label = "Node " .. id }
        if depth < maxDepth then
            node.children = buildTree(id, depth + 1, maxDepth)
        end
        items[#items + 1] = node
    end
    return items
end

local treeItems = buildTree("n", 0, 4)

-- Expand first 2 levels
local expanded = {"n-1", "n-1-1", "n-1-2", "n-2", "n-2-1", "n-3"}

lumina.app {
    id = "bench-tree",
    store = {
        selected = "n-1",
        expanded = expanded,
    },
    render = function()
        local selected = lumina.useStore("selected")
        local exp = lumina.useStore("expanded")
        return lux.Tree {
            id = "tree",
            items = treeItems,
            expandedIds = exp or {},
            selectedId = selected,
            onSelect = function(id)
                lumina.store.set("selected", id)
            end,
            onToggle = function(id, isExpanded, newExpanded)
                lumina.store.set("expanded", newExpanded)
            end,
            width = 40,
            height = 30,
        }
    end,
}
`

// TestLux_Tree_Renders verifies tree renders correctly.
func TestLux_Tree_Renders(t *testing.T) {
	app, ta := newLuxBenchAppWithTA(t, 50, 32)
	if err := app.RunString(luxTreeScript); err != nil {
		t.Fatal(err)
	}
	app.RenderAll()
	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}
	if !screenHasString(ta, "Node n-1") {
		t.Error("expected 'Node n-1' on screen")
	}
}

// BenchmarkLux_Tree_DeepNesting_Navigate measures navigation through a deep tree.
// Alternates up/down to ensure each iteration triggers a state change + re-render.
func BenchmarkLux_Tree_DeepNesting_Navigate(b *testing.B) {
	app := newLuxBenchApp(b, 50, 32)
	if err := app.RunString(luxTreeScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate down/up to keep triggering state changes
		if i%2 == 0 {
			app.HandleEvent(&event.Event{Type: "keydown", Key: "j"})
		} else {
			app.HandleEvent(&event.Event{Type: "keydown", Key: "k"})
		}
		app.RenderDirty()
	}
}

// BenchmarkLux_Tree_DeepNesting_ExpandCollapse measures expand/collapse cycles.
func BenchmarkLux_Tree_DeepNesting_ExpandCollapse(b *testing.B) {
	app := newLuxBenchApp(b, 50, 32)
	if err := app.RunString(luxTreeScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Toggle expand (Enter on a node with children)
		app.HandleEvent(&event.Event{Type: "keydown", Key: "Enter"})
		app.RenderDirty()
	}
}

// BenchmarkLux_Tree_DeepNesting_Memory measures allocations during tree navigation.
func BenchmarkLux_Tree_DeepNesting_Memory(b *testing.B) {
	app := newLuxBenchApp(b, 50, 32)
	if err := app.RunString(luxTreeScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			app.HandleEvent(&event.Event{Type: "keydown", Key: "j"})
		} else {
			app.HandleEvent(&event.Event{Type: "keydown", Key: "k"})
		}
		app.RenderDirty()
	}
}

// ═══════════════════════════════════════════════════════════════════════
// Component Mount/Unmount Cycle
// ═══════════════════════════════════════════════════════════════════════

const luxMountScript = `
local lux = require("lux")

lumina.app {
    id = "bench-mount",
    store = { show = true },
    render = function()
        local show = lumina.useStore("show")
        if show then
            return lux.Card {
                id = "card",
                title = "Mounted Component",
                width = 40,
                height = 8,
                children = {
                    lumina.createElement(lux.Badge, { key = "b1", text = "Active", variant = "success" }),
                    lumina.createElement(lux.Progress, { key = "p1", value = 50, width = 30 }),
                    lumina.createElement(lux.Spinner, { key = "s1", label = "Loading..." }),
                },
            }
        else
            return lumina.createElement("box", {
                id = "empty",
                style = { width = 40, height = 8 },
            })
        end
    end,
}
`

// BenchmarkLux_ComponentMount measures mount/unmount cycle of Lux components.
func BenchmarkLux_ComponentMount(b *testing.B) {
	app := newLuxBenchApp(b, 50, 10)
	if err := app.RunString(luxMountScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Toggle visibility via Lua store → triggers re-render
		show := "true"
		if i%2 != 0 {
			show = "false"
		}
		if err := app.RunString(fmt.Sprintf(`lumina.store.set("show", %s)`, show)); err != nil {
			b.Fatal(err)
		}
		app.RenderDirty()
	}
}

// BenchmarkLux_ComponentMount_Memory measures allocations during mount/unmount.
func BenchmarkLux_ComponentMount_Memory(b *testing.B) {
	app := newLuxBenchApp(b, 50, 10)
	if err := app.RunString(luxMountScript); err != nil {
		b.Fatal(err)
	}
	app.RenderAll()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		show := "true"
		if i%2 != 0 {
			show = "false"
		}
		if err := app.RunString(fmt.Sprintf(`lumina.store.set("show", %s)`, show)); err != nil {
			b.Fatal(err)
		}
		app.RenderDirty()
	}
}
