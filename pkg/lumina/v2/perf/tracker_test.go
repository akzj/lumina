package perf

import (
	"testing"
)

func TestTracker_Disabled_NoOp(t *testing.T) {
	tr := NewTracker(10)
	// Not enabled — all calls should be no-ops.
	tr.BeginFrame()
	tr.Record(Renders, 5)
	tr.RecordComponent("comp1")
	tr.RecordEvent("click", true)
	tr.EndFrame()

	f := tr.LastFrame()
	if f.Get(Renders) != 0 {
		t.Errorf("disabled tracker should not record: Renders=%d", f.Get(Renders))
	}
	if f.Get(EventsDispatched) != 0 {
		t.Errorf("disabled tracker should not record events")
	}
}

func TestTracker_BeginEndFrame(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.Record(Renders, 1)
	tr.EndFrame()

	f := tr.LastFrame()
	if f.Get(Renders) != 1 {
		t.Errorf("Renders: got %d, want 1", f.Get(Renders))
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
	tr.Record(Renders, 2)
	tr.Record(Layouts, 3)
	tr.Record(Paints, 1)
	tr.Record(OcclusionBuilds, 1)
	tr.Record(ComposeFull, 1)
	tr.Record(HitTesterRebuilds, 1)
	tr.Record(HandlerFullSyncs, 1)
	tr.Record(WriteDirtyCalls, 4)
	tr.Record(FlushCalls, 2)
	tr.EndFrame()

	f := tr.LastFrame()
	checks := []struct {
		m    Metric
		want int
	}{
		{Renders, 2},
		{Layouts, 3},
		{Paints, 1},
		{OcclusionBuilds, 1},
		{ComposeFull, 1},
		{HitTesterRebuilds, 1},
		{HandlerFullSyncs, 1},
		{WriteDirtyCalls, 4},
		{FlushCalls, 2},
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
	if f.Get(Renders) != 2 {
		t.Errorf("Renders: got %d, want 2", f.Get(Renders))
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
	if f.Get(EventsDispatched) != 2 {
		t.Errorf("EventsDispatched: got %d, want 2", f.Get(EventsDispatched))
	}
	if f.Get(EventsMissed) != 1 {
		t.Errorf("EventsMissed: got %d, want 1", f.Get(EventsMissed))
	}
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

	// Record 5 frames in a ring buffer of size 3.
	for i := 1; i <= 5; i++ {
		tr.BeginFrame()
		tr.Record(Renders, i)
		tr.EndFrame()
	}

	hist := tr.History()
	if len(hist) != 3 {
		t.Fatalf("History len: got %d, want 3", len(hist))
	}
	// Should contain frames 3, 4, 5 (oldest first).
	expected := []int{3, 4, 5}
	for i, h := range hist {
		if got := h.Get(Renders); got != expected[i] {
			t.Errorf("History[%d].Renders: got %d, want %d", i, got, expected[i])
		}
	}
}

func TestTracker_TotalStats_Accumulates(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	for i := 0; i < 3; i++ {
		tr.BeginFrame()
		tr.Record(Renders, 2)
		tr.Record(Layouts, 1)
		tr.EndFrame()
	}

	total := tr.TotalStats()
	if total.Frames != 3 {
		t.Errorf("Frames: got %d, want 3", total.Frames)
	}
	if total.Get(Renders) != 6 {
		t.Errorf("Total Renders: got %d, want 6", total.Get(Renders))
	}
	if total.Get(Layouts) != 3 {
		t.Errorf("Total Layouts: got %d, want 3", total.Get(Layouts))
	}
	if total.TotalDuration <= 0 {
		t.Error("TotalDuration should be positive")
	}
}

func TestTracker_Reset(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.Record(Renders, 5)
	tr.EndFrame()

	tr.Reset()

	total := tr.TotalStats()
	if total.Frames != 0 {
		t.Errorf("after reset, Frames: got %d, want 0", total.Frames)
	}
	if total.Get(Renders) != 0 {
		t.Errorf("after reset, Renders: got %d, want 0", total.Get(Renders))
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
	tr.Record(Renders, 7)
	tr.EndFrame()

	if alertCalled != 1 {
		t.Errorf("alert called %d times, want 1", alertCalled)
	}
	if alertFrame.Get(Renders) != 7 {
		t.Errorf("alert frame Renders: got %d, want 7", alertFrame.Get(Renders))
	}
}

func TestTracker_Report_NotEmpty(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	tr.BeginFrame()
	tr.Record(Renders, 1)
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
	tr.Record(Renders, 1)
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
	tr.Record(Layouts, 3)
	tr.RecordComponent("c1")
	tr.RecordComponent("c2")
	tr.EndFrame()

	// RecordComponent increments Renders, so 2 calls = Renders=2.
	tr.AssertLastFrame(t,
		CheckRenders(2),
		CheckLayouts(3),
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

	// Order should not matter.
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
	// The second frame should also have a duration.
	tr.EndFrame()

	total := tr.TotalStats()
	if total.MaxFrameDuration <= 0 {
		t.Error("MaxFrameDuration should be positive")
	}
}

func TestTracker_FrameResets_BetweenFrames(t *testing.T) {
	tr := NewTracker(10)
	tr.Enable()

	// Frame 1: record some counters.
	tr.BeginFrame()
	tr.Record(Renders, 5)
	tr.RecordEvent("click", true)
	tr.EndFrame()

	// Frame 2: record different counters.
	tr.BeginFrame()
	tr.Record(Layouts, 1)
	tr.EndFrame()

	f := tr.LastFrame()
	if f.Get(Renders) != 0 {
		t.Errorf("Frame 2 Renders should be 0, got %d", f.Get(Renders))
	}
	if f.Get(Layouts) != 1 {
		t.Errorf("Frame 2 Layouts should be 1, got %d", f.Get(Layouts))
	}
	if f.Get(EventsDispatched) != 0 {
		t.Errorf("Frame 2 EventsDispatched should be 0, got %d", f.Get(EventsDispatched))
	}
}
