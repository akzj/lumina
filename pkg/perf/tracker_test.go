package perf

import (
	"testing"
)

func TestTracker_Disabled_NoOp(t *testing.T) {
	tr := NewTracker(10)
	tr.BeginFrame()
	tr.Record(PaintCells, 5)
	tr.RecordComponent("comp1")
	tr.RecordEvent("click", true)
	tr.EndFrame()

	f := tr.LastFrame()
	if f.Get(PaintCells) != 0 {
		t.Errorf("disabled tracker should not record: PaintCells=%d", f.Get(PaintCells))
	}
	if f.EventsByType["click"] != 0 {
		t.Errorf("disabled tracker should not record events")
	}
}

func TestTracker_BeginEndFrame(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.Record(FlushCalls, 1)
	tr.EndFrame()

	f := tr.LastFrame()
	if f.Get(FlushCalls) != 1 {
		t.Errorf("FlushCalls: got %d, want 1", f.Get(FlushCalls))
	}
	if f.Duration <= 0 {
		t.Errorf("Duration should be positive, got %v", f.Duration)
	}
	if f.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}
}

func TestTracker_Record_Counters(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.Record(DirtyRectsOut, 2)
	tr.Record(WriteDirtyCalls, 3)
	tr.Record(WriteFullCalls, 1)
	tr.Record(FlushCalls, 2)
	tr.Record(ComponentsRendered, 4)
	tr.Record(PaintCells, 10)
	tr.Record(PaintClearCells, 5)
	tr.Record(DirtyRectArea, 100)
	tr.EndFrame()

	f := tr.LastFrame()
	checks := []struct {
		m    Metric
		want int
	}{
		{DirtyRectsOut, 2},
		{WriteDirtyCalls, 3},
		{WriteFullCalls, 1},
		{FlushCalls, 2},
		{ComponentsRendered, 4},
		{PaintCells, 10},
		{PaintClearCells, 5},
		{DirtyRectArea, 100},
	}
	for _, c := range checks {
		if got := f.Get(c.m); got != c.want {
			t.Errorf("metric %d: got %d, want %d", c.m, got, c.want)
		}
	}
}

func TestTracker_RecordComponent(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.RecordComponent("comp-a")
	tr.RecordComponent("comp-b")
	tr.EndFrame()

	f := tr.LastFrame()
	if f.Get(ComponentsRendered) != 2 {
		t.Errorf("render count: got %d, want 2", f.Get(ComponentsRendered))
	}
	if len(f.RenderComponents) != 2 {
		t.Fatalf("RenderComponents len: got %d, want 2", len(f.RenderComponents))
	}
	if f.RenderComponents[0] != "comp-a" || f.RenderComponents[1] != "comp-b" {
		t.Errorf("RenderComponents: got %v, want [comp-a, comp-b]", f.RenderComponents)
	}
}

func TestTracker_RecordEvent(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.RecordEvent("click", true)
	tr.RecordEvent("mousemove", true)
	tr.RecordEvent("keydown", false)
	tr.EndFrame()

	f := tr.LastFrame()
	if f.EventsByType["click"] != 1 {
		t.Errorf("EventsByType[click]: got %d, want 1", f.EventsByType["click"])
	}
	if f.EventsByType["mousemove"] != 1 {
		t.Errorf("EventsByType[mousemove]: got %d, want 1", f.EventsByType["mousemove"])
	}
	if f.EventsByType["keydown"] != 1 {
		t.Errorf("EventsByType[keydown]: got %d, want 1", f.EventsByType["keydown"])
	}
}

func TestTracker_History_RingBuffer(t *testing.T) {
	tr := NewTracker(3)
	tr.Enable()

	for i := 1; i <= 5; i++ {
		tr.BeginFrame()
		tr.Record(FlushCalls, i)
		tr.EndFrame()
	}

	hist := tr.History()
	if len(hist) != 3 {
		t.Fatalf("History len: got %d, want 3", len(hist))
	}
	expected := []int{3, 4, 5}
	for i, h := range hist {
		if got := h.Get(FlushCalls); got != expected[i] {
			t.Errorf("History[%d].FlushCalls: got %d, want %d", i, got, expected[i])
		}
	}
}

func TestTracker_TotalStats_Accumulates(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	for i := 0; i < 3; i++ {
		tr.BeginFrame()
		tr.Record(PaintCells, 2)
		tr.Record(WriteDirtyCalls, 1)
		tr.EndFrame()
	}

	total := tr.TotalStats()
	if total.Frames != 3 {
		t.Errorf("Frames: got %d, want 3", total.Frames)
	}
	if total.Get(PaintCells) != 6 {
		t.Errorf("Total PaintCells: got %d, want 6", total.Get(PaintCells))
	}
	if total.Get(WriteDirtyCalls) != 3 {
		t.Errorf("Total WriteDirtyCalls: got %d, want 3", total.Get(WriteDirtyCalls))
	}
	if total.TotalDuration <= 0 {
		t.Error("TotalDuration should be positive")
	}
}

func TestTracker_Reset(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.Record(FlushCalls, 5)
	tr.EndFrame()

	tr.Reset()

	total := tr.TotalStats()
	if total.Frames != 0 {
		t.Errorf("after reset, Frames: got %d, want 0", total.Frames)
	}
	if total.Get(FlushCalls) != 0 {
		t.Errorf("after reset, FlushCalls: got %d, want 0", total.Get(FlushCalls))
	}
}

func TestTracker_Alert(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	var alertCalled int
	var alertFrame FrameStats
	tr.SetAlert(func(f FrameStats) {
		alertCalled++
		alertFrame = f
	})

	tr.BeginFrame()
	tr.Record(PaintCells, 7)
	tr.EndFrame()

	if alertCalled != 1 {
		t.Errorf("alert called %d times, want 1", alertCalled)
	}
	if alertFrame.Get(PaintCells) != 7 {
		t.Errorf("alert frame PaintCells: got %d, want 7", alertFrame.Get(PaintCells))
	}
}

func TestTracker_Report_NotEmpty(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.Record(FlushCalls, 1)
	tr.RecordComponent("test-comp")
	tr.RecordEvent("click", true)
	tr.EndFrame()

	report := tr.Report()
	if report == "" {
		t.Error("Report should not be empty")
	}
	if len(report) < 50 {
		t.Errorf("Report too short: %q", report)
	}
}

func TestTracker_TotalReport_NotEmpty(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.Record(FlushCalls, 1)
	tr.EndFrame()

	report := tr.TotalReport()
	if report == "" {
		t.Error("TotalReport should not be empty")
	}
}

func TestTracker_AssertLastFrame(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.Record(WriteDirtyCalls, 3)
	tr.RecordComponent("c1")
	tr.RecordComponent("c2")
	tr.EndFrame()

	tr.AssertLastFrame(t,
		CheckComponentsRendered(2),
		CheckMetric(WriteDirtyCalls, 3),
		CheckRenderComponents("c1", "c2"),
	)
}

func TestTracker_AssertLastFrame_RenderComponents_OrderIndependent(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.RecordComponent("b")
	tr.RecordComponent("a")
	tr.EndFrame()

	tr.AssertLastFrame(t,
		CheckRenderComponents("a", "b"),
	)
}

func TestTracker_MaxFrameDuration(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.EndFrame()

	tr.BeginFrame()
	tr.EndFrame()

	total := tr.TotalStats()
	if total.MaxFrameDuration <= 0 {
		t.Error("MaxFrameDuration should be positive")
	}
}

func TestTracker_FrameResets_BetweenFrames(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.Record(PaintCells, 5)
	tr.RecordEvent("click", true)
	tr.EndFrame()

	tr.BeginFrame()
	tr.Record(WriteDirtyCalls, 1)
	tr.EndFrame()

	f := tr.LastFrame()
	if f.Get(PaintCells) != 0 {
		t.Errorf("Frame 2 PaintCells should be 0, got %d", f.Get(PaintCells))
	}
	if f.Get(WriteDirtyCalls) != 1 {
		t.Errorf("Frame 2 WriteDirtyCalls should be 1, got %d", f.Get(WriteDirtyCalls))
	}
	if f.EventsByType["click"] != 0 {
		t.Errorf("Frame 2 EventsByType should be fresh, got click=%d", f.EventsByType["click"])
	}
}
