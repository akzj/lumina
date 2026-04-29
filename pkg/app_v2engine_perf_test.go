package v2

import (
	"testing"

	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/perf"
)

// --- V2 Engine Perf Integration Tests ---

func TestV2Perf_RenderAll_Metrics(t *testing.T) {
	app, _, _ := newEngineApp(t, 40, 10)
	tracker := app.Tracker()
	tracker.Enable()

	err := app.RunString(`
		lumina.createComponent({
			id = "root",
			name = "Root",
			render = function(props)
				return lumina.createElement("box", {
					style = {width = 40, height = 10, background = "#111111"},
				}, lumina.createElement("text", {
					style = {foreground = "#FFFFFF"},
				}, "Hello"))
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	app.RenderAll()

	f := tracker.LastFrame()
	// Should have rendered 1 component
	if f.Get(perf.ComponentsRendered) != 1 {
		t.Errorf("render count: got %d, want 1", f.Get(perf.ComponentsRendered))
	}
	// Should have painted cells (background fill + text)
	if f.Get(perf.PaintCells) == 0 {
		t.Error("PaintCells should be > 0 after RenderAll")
	}
	// Clear cells should be > 0 (PaintFull calls Clear)
	if f.Get(perf.PaintClearCells) == 0 {
		t.Error("PaintClearCells should be > 0 after RenderAll (PaintFull clears)")
	}
	// DirtyRectArea should cover the full screen (full render)
	if f.Get(perf.DirtyRectArea) != 40*10 {
		t.Errorf("DirtyRectArea: got %d, want %d", f.Get(perf.DirtyRectArea), 40*10)
	}
	// Output metrics
	if f.Get(perf.WriteFullCalls) != 1 {
		t.Errorf("WriteFullCalls: got %d, want 1", f.Get(perf.WriteFullCalls))
	}
	if f.Get(perf.FlushCalls) != 1 {
		t.Errorf("FlushCalls: got %d, want 1", f.Get(perf.FlushCalls))
	}
}

func TestV2Perf_RenderDirty_NothingDirty(t *testing.T) {
	app, _, _ := newEngineApp(t, 20, 5)
	tracker := app.Tracker()
	tracker.Enable()

	err := app.RunString(`
		lumina.createComponent({
			id = "static",
			name = "Static",
			render = function(props)
				return lumina.createElement("text", {}, "Hello")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	app.RenderAll()

	// Now nothing is dirty — RenderDirty should be near no-op.
	app.RenderDirty()

	f := tracker.LastFrame()
	if f.Get(perf.ComponentsRendered) != 0 {
		t.Errorf("render count: got %d, want 0 (nothing dirty)", f.Get(perf.ComponentsRendered))
	}
	if f.Get(perf.PaintCells) != 0 {
		t.Errorf("PaintCells: got %d, want 0 (nothing dirty)", f.Get(perf.PaintCells))
	}
	if f.Get(perf.WriteDirtyCalls) != 0 {
		t.Errorf("WriteDirtyCalls: got %d, want 0 (nothing to output)", f.Get(perf.WriteDirtyCalls))
	}
	if f.Get(perf.FlushCalls) != 0 {
		t.Errorf("FlushCalls: got %d, want 0 (nothing to output)", f.Get(perf.FlushCalls))
	}
}

func TestV2Perf_ClickStateChange_Metrics(t *testing.T) {
	app, _, _ := newEngineApp(t, 80, 24)
	tracker := app.Tracker()
	tracker.Enable()

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

	// Click triggers setState → marks dirty
	app.HandleEvent(&event.Event{Type: "click", X: 10, Y: 10})
	app.RenderDirty()

	f := tracker.LastFrame()
	// Should re-render 1 component
	if f.Get(perf.ComponentsRendered) != 1 {
		t.Errorf("render count: got %d, want 1", f.Get(perf.ComponentsRendered))
	}
	// Should paint some cells
	if f.Get(perf.PaintCells) == 0 {
		t.Error("PaintCells should be > 0 after state change")
	}
	// Should have dirty output
	if f.Get(perf.WriteDirtyCalls) != 1 {
		t.Errorf("WriteDirtyCalls: got %d, want 1", f.Get(perf.WriteDirtyCalls))
	}
	if f.Get(perf.DirtyRectsOut) != 1 {
		t.Errorf("DirtyRectsOut: got %d, want 1", f.Get(perf.DirtyRectsOut))
	}

	t.Logf("After click: PaintCells=%d, PaintClearCells=%d, DirtyRectArea=%d",
		f.Get(perf.PaintCells), f.Get(perf.PaintClearCells), f.Get(perf.DirtyRectArea))
}

func TestV2Perf_DirtyRect_Accurate(t *testing.T) {
	app, _, _ := newEngineApp(t, 80, 24)
	tracker := app.Tracker()
	tracker.Enable()

	// Create a component that renders a small text
	err := app.RunString(`
		lumina.createComponent({
			id = "small",
			name = "Small",
			render = function(props)
				local val, setVal = lumina.useState("v", "A")
				return lumina.createElement("text", {
					style = {width = 5, height = 1},
					onClick = function() setVal("B") end,
				}, val)
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	app.RenderAll()

	// Click to change state
	app.HandleEvent(&event.Event{Type: "click", X: 0, Y: 0})
	app.RenderDirty()

	f := tracker.LastFrame()
	dirtyArea := f.Get(perf.DirtyRectArea)
	totalArea := 80 * 24

	t.Logf("DirtyRectArea=%d, TotalArea=%d, ratio=%.1f%%",
		dirtyArea, totalArea, float64(dirtyArea)/float64(totalArea)*100)

	// Dirty rect should NOT be the full screen for a small text change
	if dirtyArea >= totalArea {
		t.Errorf("DirtyRectArea (%d) should be less than total screen (%d) for small text change",
			dirtyArea, totalArea)
	}
}

func TestV2Perf_Report_IncludesV2Section(t *testing.T) {
	app, _, _ := newEngineApp(t, 20, 5)
	tracker := app.Tracker()
	tracker.Enable()

	err := app.RunString(`
		lumina.createComponent({
			id = "r",
			name = "R",
			render = function(props)
				return lumina.createElement("text", {}, "X")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	app.RenderAll()

	report := tracker.Report()
	if len(report) < 50 {
		t.Errorf("Report too short: %q", report)
	}
	if !containsString(report, "Render:") {
		t.Errorf("Report should contain Render line, got:\n%s", report)
	}
}

func TestV2Perf_AssertHelpers(t *testing.T) {
	app, _, _ := newEngineApp(t, 20, 5)
	tracker := app.Tracker()
	tracker.Enable()

	err := app.RunString(`
		lumina.createComponent({
			id = "a",
			name = "A",
			render = function(props)
				return lumina.createElement("text", {}, "Hi")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	app.RenderAll()

	tracker.AssertLastFrame(t,
		perf.CheckComponentsRendered(1),
		perf.CheckPaintCellsMax(200), // 20*5=100 cells max for full screen
	)
}

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
