package v2

import (
	"fmt"
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/output"
)

// ═══════════════════════════════════════════════════════════════════════
// Stress Test Benchmarks
//
// These benchmarks reproduce the scenario from examples/stress_test.lua:
// 80×23 = 1840 individual box elements with hover/click handlers.
// Each hover triggers a full re-render of all 1840 elements.
// ═══════════════════════════════════════════════════════════════════════

const (
	stressCols = 80
	stressRows = 23
	stressCells = stressCols * stressRows // 1840
)

// stressLuaScript is the inline Lua that creates the stress test grid.
// It mirrors examples/stress_test.lua but is self-contained for benchmarks.
const stressLuaScript = `
local COLS = 80
local ROWS = 23

lumina.createComponent({
    id = "stress",
    name = "StressTest",
    x = 0, y = 0,
    w = COLS, h = ROWS + 1,
    zIndex = 0,

    render = function(state, props)
        local hoveredCell, setHoveredCell = lumina.useState("hovered", "")
        local clickedCells, setClickedCells = lumina.useState("clicked", {})
        local clickCount, setClickCount = lumina.useState("clickCount", 0)

        local function toggleCell(cellId)
            local newClicked = {}
            for k, v in pairs(clickedCells) do
                newClicked[k] = v
            end
            if newClicked[cellId] then
                newClicked[cellId] = nil
            else
                newClicked[cellId] = true
            end
            setClickedCells(newClicked)
            setClickCount(clickCount + 1)
        end

        local rowElements = {}
        for y = 0, ROWS - 1 do
            local cellsInRow = {}
            for x = 0, COLS - 1 do
                local cellId = x .. "," .. y
                local isHovered = (hoveredCell == cellId)
                local isClicked = (clickedCells[cellId] == true)

                local ch, fg, bg
                if isHovered and isClicked then
                    ch = "*"
                    fg = "#F38BA8"
                    bg = "#45475A"
                elseif isHovered then
                    ch = "█"
                    fg = "#A6E3A1"
                    bg = "#313244"
                elseif isClicked then
                    ch = "×"
                    fg = "#F38BA8"
                    bg = "#313244"
                else
                    ch = "·"
                    fg = "#585B70"
                    bg = "#1E1E2E"
                end

                local cid = cellId
                cellsInRow[#cellsInRow + 1] = {
                    type = "box",
                    id = cid,
                    style = {width = 1, height = 1, background = bg},
                    onMouseEnter = function() setHoveredCell(cid) end,
                    onMouseLeave = function() setHoveredCell("") end,
                    onClick = function() toggleCell(cid) end,
                    children = {
                        {type = "text", content = ch, style = {foreground = fg}},
                    },
                }
            end

            rowElements[#rowElements + 1] = {
                type = "hbox",
                id = "row-" .. y,
                style = {height = 1},
                children = cellsInRow,
            }
        end

        -- Status bar
        local statusText = string.format(
            " %dx%d=%d cells | Hovered:%s | [q]Quit",
            COLS, ROWS, COLS * ROWS, hoveredCell
        )
        rowElements[#rowElements + 1] = lumina.createElement("text", {
            foreground = "#89B4FA",
            bold = true,
            style = {background = "#181825", height = 1},
        }, statusText)

        return {
            type = "vbox",
            id = "stress-root",
            style = {background = "#1E1E2E"},
            focusable = true,
            children = rowElements,
        }
    end,
})
`

// newStressApp creates an App loaded with the stress test Lua script.
// Returns the app, test adapter, and Lua state.
func newStressApp(tb testing.TB) (*App, *output.TestAdapter, *lua.State) {
	tb.Helper()
	L := lua.NewState()
	tb.Cleanup(func() { L.Close() })
	ta := output.NewTestAdapter()
	app := NewApp(L, stressCols, stressRows+1, ta)
	if err := app.RunString(stressLuaScript); err != nil {
		tb.Fatalf("RunString failed: %v", err)
	}
	return app, ta, L
}

// newStressAppNop creates an App with a no-op adapter for pure render benchmarks.
func newStressAppNop(tb testing.TB) (*App, *lua.State) {
	tb.Helper()
	L := lua.NewState()
	tb.Cleanup(func() { L.Close() })
	app := NewApp(L, stressCols, stressRows+1, nopAdapter{})
	if err := app.RunString(stressLuaScript); err != nil {
		tb.Fatalf("RunString failed: %v", err)
	}
	return app, L
}

// stressLuaScriptV2 uses per-cell sub-components for the V2 engine.
// Each cell is an independent component — hover only re-renders 1-2 cells.
const stressLuaScriptV2 = `
local COLS = 80
local ROWS = 23

-- Each Cell is an independent component with its own hover/click state.
local Cell = lumina.defineComponent("Cell", function(props)
    local hovered, setHovered = lumina.useState("h", false)
    local clicked, setClicked = lumina.useState("c", false)

    local ch, fg, bg
    if hovered and clicked then
        ch = "*"
        fg = "#F38BA8"
        bg = "#45475A"
    elseif hovered then
        ch = "█"
        fg = "#A6E3A1"
        bg = "#313244"
    elseif clicked then
        ch = "×"
        fg = "#F38BA8"
        bg = "#313244"
    else
        ch = "·"
        fg = "#585B70"
        bg = "#1E1E2E"
    end

    return lumina.createElement("box", {
        style = {width = 1, height = 1, background = bg},
        onMouseEnter = function() setHovered(true) end,
        onMouseLeave = function() setHovered(false) end,
        onClick = function() setClicked(not clicked) end,
    }, lumina.createElement("text", {
        style = {foreground = fg},
    }, ch))
end)

lumina.createComponent({
    id = "stress",
    name = "StressTest",

    render = function(props)
        local rowElements = {}
        for y = 0, ROWS - 1 do
            local cellsInRow = {}
            for x = 0, COLS - 1 do
                local cellId = x .. "," .. y
                cellsInRow[#cellsInRow + 1] = lumina.createElement(Cell, {
                    key = cellId,
                    id = cellId,
                })
            end
            rowElements[#rowElements + 1] = {
                type = "hbox",
                id = "row-" .. y,
                style = {height = 1},
                children = cellsInRow,
            }
        end

        rowElements[#rowElements + 1] = lumina.createElement("text", {
            foreground = "#89B4FA",
            bold = true,
            style = {background = "#181825", height = 1},
        }, string.format(
            " %dx%d=%d cells | Per-cell components | [q]Quit",
            COLS, ROWS, COLS * ROWS
        ))

        return {
            type = "vbox",
            id = "stress-root",
            style = {background = "#1E1E2E"},
            focusable = true,
            children = rowElements,
        }
    end,
})
`

// newStressAppV2 creates an App with the new V2 engine loaded with the stress test script.
func newStressAppV2(tb testing.TB) (*App, *output.TestAdapter, *lua.State) {
	tb.Helper()
	L := lua.NewState()
	tb.Cleanup(func() { L.Close() })
	ta := output.NewTestAdapter()
	app := NewApp(L, stressCols, stressRows+1, ta)
	if err := app.RunString(stressLuaScriptV2); err != nil {
		tb.Fatalf("RunString failed: %v", err)
	}
	return app, ta, L
}

// newStressAppV2Nop creates an App with the V2 engine and no-op adapter for pure render benchmarks.
func newStressAppV2Nop(tb testing.TB) (*App, *lua.State) {
	tb.Helper()
	L := lua.NewState()
	tb.Cleanup(func() { L.Close() })
	app := NewApp(L, stressCols, stressRows+1, nopAdapter{})
	if err := app.RunString(stressLuaScriptV2); err != nil {
		tb.Fatalf("RunString failed: %v", err)
	}
	return app, L
}

// ═══════════════════════════════════════════════════════════════════════
// Test: Stress test loads and renders correctly
// ═══════════════════════════════════════════════════════════════════════

func TestStressTest_InitialRender(t *testing.T) {
	app, ta, _ := newStressApp(t)

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil after RenderAll")
	}

	// Verify the grid dot character appears on screen.
	if !screenHasChar(ta, '·') {
		t.Error("expected dot character '·' on screen")
	}

	// Verify status bar text.
	if !screenHasString(ta, "1840 cells") {
		t.Error("expected '1840 cells' in status bar")
	}
}

// ═══════════════════════════════════════════════════════════════════════
// Test: Stress test performance profiling (verbose, not a benchmark)
// ═══════════════════════════════════════════════════════════════════════

func TestStressPerf_HoverCycle(t *testing.T) {
	app, _, _ := newStressApp(t)

	tracker := app.Tracker()
	tracker.Enable()

	// Initial full render.
	app.RenderAll()
	initialFrame := tracker.LastFrame()
	t.Logf("=== Initial RenderAll ===")
	t.Logf("  Duration: %v", initialFrame.Duration)
	t.Logf("  Report:\n%s", tracker.Report())

	// Simulate 20 hover events (state change → RenderDirty).
	const hoverCount = 20
	start := time.Now()
	for i := 0; i < hoverCount; i++ {
		x := i % stressCols
		y := i % stressRows
		cellID := fmt.Sprintf("%d,%d", x, y)
		app.SetState("stress", "hovered", cellID)
		app.RenderDirty()
	}
	totalHover := time.Since(start)

	t.Logf("=== Hover Performance (%d frames) ===", hoverCount)
	t.Logf("  Total: %v", totalHover)
	t.Logf("  Avg per frame: %v", totalHover/time.Duration(hoverCount))
	t.Logf("  Effective FPS: %.1f", float64(hoverCount)/totalHover.Seconds())
	t.Logf("  Last frame report:\n%s", tracker.Report())
	t.Logf("  Total report:\n%s", tracker.TotalReport())

	// Sanity check: each hover should trigger exactly 1 render.
	lastFrame := tracker.LastFrame()
	if lastFrame.Get(1) < 1 { // perf.Renders
		t.Logf("  WARNING: last frame had 0 renders")
	}
}

func TestStressPerf_ClickCycle(t *testing.T) {
	app, _, _ := newStressApp(t)

	tracker := app.Tracker()
	tracker.Enable()

	app.RenderAll()

	// Simulate 10 click events via HandleEvent (full event dispatch path).
	const clickCount = 10
	start := time.Now()
	for i := 0; i < clickCount; i++ {
		x := i % stressCols
		y := i % stressRows
		app.HandleEvent(&event.Event{Type: "mousedown", X: x, Y: y})
		app.RenderDirty()
	}
	totalClick := time.Since(start)

	t.Logf("=== Click Performance (%d frames) ===", clickCount)
	t.Logf("  Total: %v", totalClick)
	t.Logf("  Avg per frame: %v", totalClick/time.Duration(clickCount))
	t.Logf("  Effective FPS: %.1f", float64(clickCount)/totalClick.Seconds())
	t.Logf("  Total report:\n%s", tracker.TotalReport())
}

// ═══════════════════════════════════════════════════════════════════════
// Test: Stress test with actual Lua script file
// ═══════════════════════════════════════════════════════════════════════

func TestStressPerf_ActualScript(t *testing.T) {
	app, ta, _ := newLuaApp(t, stressCols, stressRows+1)

	err := app.RunScript("../examples/stress_test.lua")
	if err != nil {
		// stress_test.lua now uses V2 APIs (defineComponent) — skip for V1 pipeline
		t.Skipf("Skipping V1 actual script test (script uses V2 APIs): %v", err)
	}

	tracker := app.Tracker()
	tracker.Enable()

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}

	t.Logf("=== Actual stress_test.lua Initial Render ===")
	t.Logf("  Duration: %v", tracker.LastFrame().Duration)
	t.Logf("  Report:\n%s", tracker.Report())

	// Simulate 10 hover cycles via SetState.
	const hoverCount = 10
	start := time.Now()
	for i := 0; i < hoverCount; i++ {
		cellID := fmt.Sprintf("%d,%d", i%stressCols, i%stressRows)
		app.SetState("stress", "hovered", cellID)
		app.RenderDirty()
	}
	totalHover := time.Since(start)

	t.Logf("=== Actual Script Hover (%d frames) ===", hoverCount)
	t.Logf("  Total: %v", totalHover)
	t.Logf("  Avg per frame: %v", totalHover/time.Duration(hoverCount))
	t.Logf("  Effective FPS: %.1f", float64(hoverCount)/totalHover.Seconds())
}

// ═══════════════════════════════════════════════════════════════════════
// Benchmarks: Stress test render cycle
// ═══════════════════════════════════════════════════════════════════════

// BenchmarkStress_RenderAll measures the full initial render of 1840 elements.
func BenchmarkStress_RenderAll(b *testing.B) {
	app, _ := newStressAppNop(b)
	app.RenderAll() // warm up

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Force full re-render by marking dirty.
		app.SetState("stress", "hovered", fmt.Sprintf("0,%d", i%stressRows))
		app.RenderAll()
	}
}

// BenchmarkStress_RenderDirty_Hover measures the incremental render on hover.
// This is the hot path: user moves mouse → state change → RenderDirty.
func BenchmarkStress_RenderDirty_Hover(b *testing.B) {
	app, _ := newStressAppNop(b)
	app.RenderAll() // initial render

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cellID := fmt.Sprintf("%d,%d", i%stressCols, i%stressRows)
		app.SetState("stress", "hovered", cellID)
		app.RenderDirty()
	}
}

// BenchmarkStress_RenderDirty_Hover_WithTestAdapter measures with the
// TestAdapter (includes buffer cloning overhead).
func BenchmarkStress_RenderDirty_Hover_WithTestAdapter(b *testing.B) {
	app, _, _ := newStressApp(b)
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cellID := fmt.Sprintf("%d,%d", i%stressCols, i%stressRows)
		app.SetState("stress", "hovered", cellID)
		app.RenderDirty()
	}
}

// BenchmarkStress_HandleEvent_MouseMove measures the event dispatch path
// for mouse moves (hit testing through 1840 elements).
func BenchmarkStress_HandleEvent_MouseMove(b *testing.B) {
	app, _ := newStressAppNop(b)
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.HandleEvent(&event.Event{
			Type: "mousemove",
			X:    i % stressCols,
			Y:    i % stressRows,
		})
	}
}

// BenchmarkStress_FullHoverCycle measures the complete hover cycle:
// mouseenter event dispatch + state change + RenderDirty.
func BenchmarkStress_FullHoverCycle(b *testing.B) {
	app, _ := newStressAppNop(b)
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % stressCols
		y := i % stressRows
		// Simulate mouseenter → state change → render
		app.HandleEvent(&event.Event{Type: "mousemove", X: x, Y: y})
		app.RenderDirty()
	}
}

// BenchmarkStress_RenderAll_Memory measures allocations during full render.
func BenchmarkStress_RenderAll_Memory(b *testing.B) {
	app, _ := newStressAppNop(b)
	app.RenderAll() // warm up

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.SetState("stress", "hovered", fmt.Sprintf("0,%d", i%stressRows))
		app.RenderAll()
	}
}

// BenchmarkStress_RenderDirty_Memory measures allocations during hover render.
func BenchmarkStress_RenderDirty_Memory(b *testing.B) {
	app, _ := newStressAppNop(b)
	app.RenderAll()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cellID := fmt.Sprintf("%d,%d", i%stressCols, i%stressRows)
		app.SetState("stress", "hovered", cellID)
		app.RenderDirty()
	}
}

// ═══════════════════════════════════════════════════════════════════════
// V2 Engine: Test — Stress test loads and renders correctly
// ═══════════════════════════════════════════════════════════════════════

func TestStressV2_InitialRender(t *testing.T) {
	app, ta, _ := newStressAppV2(t)

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil after RenderAll")
	}

	// Verify the grid dot character appears on screen.
	if !screenHasChar(ta, '·') {
		t.Error("expected dot character '·' on screen")
	}

	// Verify status bar text.
	if !screenHasString(ta, "1840 cells") {
		t.Error("expected '1840 cells' in status bar")
	}
}

// ═══════════════════════════════════════════════════════════════════════
// V2 Engine: Perf test — Hover cycle
// ═══════════════════════════════════════════════════════════════════════

func TestStressPerfV2_HoverCycle(t *testing.T) {
	app, _, _ := newStressAppV2(t)
	app.RenderAll()

	// With per-cell components, hover uses HandleEvent (mousemove)
	// which triggers Cell's onMouseEnter → setHovered(true) → only that cell re-renders.
	const hoverCount = 20
	start := time.Now()
	for i := 0; i < hoverCount; i++ {
		x := i % stressCols
		y := i % stressRows
		app.HandleEvent(&event.Event{Type: "mousemove", X: x, Y: y})
		app.RenderDirty()
	}
	totalHover := time.Since(start)

	t.Logf("=== V2 Engine Hover Performance (%d frames) ===", hoverCount)
	t.Logf("  Total: %v", totalHover)
	t.Logf("  Avg per frame: %v", totalHover/time.Duration(hoverCount))
	t.Logf("  Effective FPS: %.1f", float64(hoverCount)/totalHover.Seconds())
}

// ═══════════════════════════════════════════════════════════════════════
// V2 Engine: Perf test — Actual script file
// ═══════════════════════════════════════════════════════════════════════

func TestStressPerfV2_ActualScript(t *testing.T) {
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	ta := output.NewTestAdapter()
	app := NewApp(L, stressCols, stressRows+1, ta)

	err := app.RunScript("../examples/stress_test.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}

	tracker := app.Tracker()
	tracker.Enable()

	app.RenderAll()

	if ta.LastScreen == nil {
		t.Fatal("LastScreen is nil")
	}

	t.Logf("=== V2 Actual stress_test.lua Initial Render ===")
	t.Logf("  Duration: %v", tracker.LastFrame().Duration)

	// With per-cell components, hover uses HandleEvent (mousemove).
	const hoverCount = 10
	start := time.Now()
	for i := 0; i < hoverCount; i++ {
		x := i % stressCols
		y := i % stressRows
		app.HandleEvent(&event.Event{Type: "mousemove", X: x, Y: y})
		app.RenderDirty()
	}
	totalHover := time.Since(start)

	t.Logf("=== V2 Actual Script Hover (%d frames) ===", hoverCount)
	t.Logf("  Total: %v", totalHover)
	t.Logf("  Avg per frame: %v", totalHover/time.Duration(hoverCount))
	t.Logf("  Effective FPS: %.1f", float64(hoverCount)/totalHover.Seconds())
}

// ═══════════════════════════════════════════════════════════════════════
// V2 Engine Benchmarks (new render pipeline)
// ═══════════════════════════════════════════════════════════════════════

// BenchmarkStressV2_RenderAll measures the full initial render with V2 engine.
func BenchmarkStressV2_RenderAll(b *testing.B) {
	app, _ := newStressAppV2Nop(b)
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Engine().SetState("stress", "hovered", fmt.Sprintf("0,%d", i%stressRows))
		app.RenderAll()
	}
}

// BenchmarkStressV2_RenderDirty_Hover measures the incremental render on hover with V2 engine.
// With per-cell components, hover triggers only 1-2 cell re-renders via HandleEvent.
func BenchmarkStressV2_RenderDirty_Hover(b *testing.B) {
	app, _ := newStressAppV2Nop(b)
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % stressCols
		y := i % stressRows
		app.HandleEvent(&event.Event{Type: "mousemove", X: x, Y: y})
		app.RenderDirty()
	}
}

// BenchmarkStressV2_RenderDirty_Hover_WithTestAdapter measures V2 with TestAdapter.
func BenchmarkStressV2_RenderDirty_Hover_WithTestAdapter(b *testing.B) {
	app, _, _ := newStressAppV2(b)
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % stressCols
		y := i % stressRows
		app.HandleEvent(&event.Event{Type: "mousemove", X: x, Y: y})
		app.RenderDirty()
	}
}

// BenchmarkStressV2_HandleEvent_MouseMove measures V2 event dispatch for mouse moves.
func BenchmarkStressV2_HandleEvent_MouseMove(b *testing.B) {
	app, _ := newStressAppV2Nop(b)
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.HandleEvent(&event.Event{
			Type: "mousemove",
			X:    i % stressCols,
			Y:    i % stressRows,
		})
	}
}

// BenchmarkStressV2_FullHoverCycle measures the complete V2 hover cycle:
// mouseenter event dispatch + state change + RenderDirty.
func BenchmarkStressV2_FullHoverCycle(b *testing.B) {
	app, _ := newStressAppV2Nop(b)
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % stressCols
		y := i % stressRows
		app.HandleEvent(&event.Event{Type: "mousemove", X: x, Y: y})
		app.RenderDirty()
	}
}

// BenchmarkStressV2_RenderAll_Memory measures V2 allocations during full render.
func BenchmarkStressV2_RenderAll_Memory(b *testing.B) {
	app, _ := newStressAppV2Nop(b)
	app.RenderAll()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.Engine().SetState("stress", "hovered", fmt.Sprintf("0,%d", i%stressRows))
		app.RenderAll()
	}
}

// BenchmarkStressV2_RenderDirty_Memory measures V2 allocations during hover render.
func BenchmarkStressV2_RenderDirty_Memory(b *testing.B) {
	app, _ := newStressAppV2Nop(b)
	app.RenderAll()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % stressCols
		y := i % stressRows
		app.HandleEvent(&event.Event{Type: "mousemove", X: x, Y: y})
		app.RenderDirty()
	}
}
