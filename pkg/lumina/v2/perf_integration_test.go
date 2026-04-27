package v2

import (
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/perf"
)

// --- Perf Integration Tests ---

func TestPerfIntegration_RenderAll_Metrics(t *testing.T) {
	app, _ := NewTestApp(20, 10)
	app.Tracker().Enable()

	app.RegisterComponent("c1", "c1", buffer.Rect{X: 0, Y: 0, W: 10, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "A"
			return vn
		})
	app.RegisterComponent("c2", "c2", buffer.Rect{X: 10, Y: 0, W: 10, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "B"
			return vn
		})

	app.RenderAll()

	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(2),
		perf.CheckLayouts(2),
		perf.CheckPaints(2),
		perf.CheckOcclusionBuilds(1),
		perf.CheckMetric(perf.ComposeFull, 1),
		perf.CheckHitTesterRebuilds(1),
		perf.CheckHandlerFullSyncs(1),
		perf.CheckMetric(perf.WriteFullCalls, 1),
		perf.CheckMetric(perf.FlushCalls, 1),
		perf.CheckRenderComponents("c1", "c2"),
	)
}

func TestPerfIntegration_MovePositionOnly(t *testing.T) {
	app, _ := NewTestApp(40, 20)
	app.Tracker().Enable()

	app.RegisterComponent("bg", "bg", buffer.Rect{X: 0, Y: 0, W: 40, H: 20}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#000000"
			child := layout.NewVNode("text")
			child.Content = "B"
			root.AddChild(child)
			return root
		})
	app.RegisterComponent("dlg", "dlg", buffer.Rect{X: 5, Y: 5, W: 10, H: 5}, 100,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#FFFFFF"
			child := layout.NewVNode("text")
			child.Content = "D"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Now do a position-only move.
	app.MoveComponent("dlg", buffer.Rect{X: 20, Y: 5, W: 10, H: 5})
	app.RenderDirty()

	app.Tracker().AssertLastFrame(t,
		// Position-only move: no re-render of dlg (same size).
		perf.CheckRenders(0),
		perf.CheckPaints(0),
		// Occlusion rebuild needed (rect changed).
		perf.CheckOcclusionBuilds(1),
		// ComposeRects used for the move.
		perf.CheckMetric(perf.ComposeRects, 1),
		// HitTester rebuilt.
		perf.CheckHitTesterRebuilds(1),
		// Full handler sync (structural change).
		perf.CheckHandlerFullSyncs(1),
	)

	// Verify move counter.
	f := app.Tracker().LastFrame()
	if f.Get(perf.MovesPositionOnly) != 1 {
		t.Errorf("MovesPositionOnly: got %d, want 1", f.Get(perf.MovesPositionOnly))
	}
}

func TestPerfIntegration_MoveWithResize(t *testing.T) {
	app, _ := NewTestApp(40, 20)
	app.Tracker().Enable()

	app.RegisterComponent("bg", "bg", buffer.Rect{X: 0, Y: 0, W: 40, H: 20}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#000000"
			child := layout.NewVNode("text")
			child.Content = "B"
			root.AddChild(child)
			return root
		})
	app.RegisterComponent("dlg", "dlg", buffer.Rect{X: 5, Y: 5, W: 10, H: 5}, 100,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.Style.Background = "#FFFFFF"
			child := layout.NewVNode("text")
			child.Content = "D"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Move with resize — should trigger re-render.
	app.MoveComponent("dlg", buffer.Rect{X: 5, Y: 5, W: 15, H: 8})
	app.RenderDirty()

	app.Tracker().AssertLastFrame(t,
		// Resize triggers re-render of the moved component.
		perf.CheckRenders(1),
		perf.CheckRenderComponents("dlg"),
		perf.CheckOcclusionBuilds(1),
		perf.CheckHitTesterRebuilds(1),
	)

	f := app.Tracker().LastFrame()
	if f.Get(perf.MovesWithResize) != 1 {
		t.Errorf("MovesWithResize: got %d, want 1", f.Get(perf.MovesWithResize))
	}
}

func TestPerfIntegration_SetState_OnlyDirty(t *testing.T) {
	app, _ := NewTestApp(50, 10)
	app.Tracker().Enable()

	// Register 5 components.
	for i := 0; i < 5; i++ {
		id := string(rune('A'+i)) + "-comp"
		x := i * 10
		app.RegisterComponent(id, id, buffer.Rect{X: x, Y: 0, W: 10, H: 10}, 0,
			func(state, props map[string]any) *layout.VNode {
				vn := layout.NewVNode("text")
				vn.Content = "X"
				return vn
			})
	}
	app.RenderAll()

	// SetState on only one component.
	app.SetState("C-comp", "count", 1)
	app.RenderDirty()

	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(1),
		perf.CheckRenderComponents("C-comp"),
		// No structural change → no occlusion rebuild.
		perf.CheckOcclusionBuilds(0),
		perf.CheckHitTesterRebuilds(0),
		// Dirty handler sync (not full).
		perf.CheckHandlerDirtySyncs(1),
		perf.CheckHandlerFullSyncs(0),
	)
}

func TestPerfIntegration_EventDispatch(t *testing.T) {
	app, _ := NewTestApp(20, 10)
	app.Tracker().Enable()

	app.RegisterComponent("btn", "btn", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.ID = "btn-1"
			vn.Content = "Click"
			vn.Props["onClick"] = event.EventHandler(func(e *event.Event) {})
			return vn
		})

	app.RenderAll()

	// Click on the button — should dispatch.
	app.Tracker().BeginFrame()
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	app.Tracker().EndFrame()

	app.Tracker().AssertLastFrame(t,
		perf.CheckEventsDispatched(1),
		perf.CheckEventsMissed(0),
	)

	// Click on empty area (no component) — should miss.
	app.Tracker().BeginFrame()
	app.HandleEvent(&event.Event{Type: "mousedown", X: 15, Y: 8})
	app.Tracker().EndFrame()

	app.Tracker().AssertLastFrame(t,
		perf.CheckEventsDispatched(0),
		perf.CheckEventsMissed(1),
	)
}

func TestPerfIntegration_NoDirty_NoOp(t *testing.T) {
	app, _ := NewTestApp(10, 5)
	app.Tracker().Enable()

	app.RegisterComponent("c", "c", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "A"
			return vn
		})

	app.RenderAll()

	// Nothing dirty — RenderDirty should be a near no-op.
	app.RenderDirty()

	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(0),
		perf.CheckLayouts(0),
		perf.CheckPaints(0),
		perf.CheckOcclusionBuilds(0),
		perf.CheckMetric(perf.ComposeFull, 0),
		perf.CheckMetric(perf.ComposeDirty, 0),
		perf.CheckHitTesterRebuilds(0),
		perf.CheckHandlerFullSyncs(0),
		perf.CheckHandlerDirtySyncs(0),
		perf.CheckMetric(perf.WriteDirtyCalls, 0),
		perf.CheckMetric(perf.FlushCalls, 0),
	)
}

func TestPerfIntegration_ComplexScenario(t *testing.T) {
	app, _ := NewTestApp(40, 20)
	app.Tracker().Enable()

	// Background.
	app.RegisterComponent("bg", "bg", buffer.Rect{X: 0, Y: 0, W: 40, H: 20}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "bg"
			root.Style.Background = "#000000"
			child := layout.NewVNode("text")
			child.Content = "B"
			root.AddChild(child)
			return root
		})

	// Counter component.
	app.RegisterComponent("counter", "counter", buffer.Rect{X: 0, Y: 0, W: 10, H: 1}, 10,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.ID = "counter-text"
			vn.Content = "0"
			vn.Props["onClick"] = event.EventHandler(func(e *event.Event) {})
			return vn
		})

	// Dialog.
	app.RegisterComponent("dlg", "dlg", buffer.Rect{X: 10, Y: 5, W: 15, H: 8}, 100,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "dlg"
			root.Style.Background = "#FFFFFF"
			child := layout.NewVNode("text")
			child.Content = "D"
			root.AddChild(child)
			return root
		})

	app.RenderAll()

	// Frame 1: SetState on counter + position-only move on dialog.
	app.SetState("counter", "n", 1)
	app.MoveComponent("dlg", buffer.Rect{X: 15, Y: 5, W: 15, H: 8})
	app.RenderDirty()

	f := app.Tracker().LastFrame()
	// counter re-rendered, dialog NOT re-rendered (same size).
	if f.Get(perf.Renders) != 1 {
		t.Errorf("Frame 1 Renders: got %d, want 1", f.Get(perf.Renders))
	}
	// Structural change (rect changed) → occlusion rebuild.
	if f.Get(perf.OcclusionBuilds) != 1 {
		t.Errorf("Frame 1 OcclusionBuilds: got %d, want 1", f.Get(perf.OcclusionBuilds))
	}
	// HitTester rebuilt.
	if f.Get(perf.HitTesterRebuilds) != 1 {
		t.Errorf("Frame 1 HitTesterRebuilds: got %d, want 1", f.Get(perf.HitTesterRebuilds))
	}
	// Position-only move counter.
	if f.Get(perf.MovesPositionOnly) != 1 {
		t.Errorf("Frame 1 MovesPositionOnly: got %d, want 1", f.Get(perf.MovesPositionOnly))
	}
	// SetState counter.
	if f.Get(perf.StateSets) != 1 {
		t.Errorf("Frame 1 StateSets: got %d, want 1", f.Get(perf.StateSets))
	}

	// Frame 2: Just an event, no render.
	app.Tracker().BeginFrame()
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	app.Tracker().EndFrame()

	f2 := app.Tracker().LastFrame()
	if f2.Get(perf.EventsDispatched) != 1 {
		t.Errorf("Frame 2 EventsDispatched: got %d, want 1", f2.Get(perf.EventsDispatched))
	}
	if f2.Get(perf.Renders) != 0 {
		t.Errorf("Frame 2 Renders: got %d, want 0", f2.Get(perf.Renders))
	}

	// Check total stats.
	total := app.Tracker().TotalStats()
	if total.Frames < 3 {
		t.Errorf("Total frames should be >= 3 (RenderAll + RenderDirty + event frame), got %d", total.Frames)
	}
}

func TestPerfIntegration_Disabled_ZeroOverhead(t *testing.T) {
	app, _ := NewTestApp(10, 5)
	// Tracker is NOT enabled (default state).

	app.RegisterComponent("c", "c", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "A"
			return vn
		})

	app.RenderAll()
	app.SetState("c", "x", 1)
	app.RenderDirty()

	// Tracker should have no data.
	f := app.Tracker().LastFrame()
	if f.Get(perf.Renders) != 0 {
		t.Errorf("disabled tracker should not record: Renders=%d", f.Get(perf.Renders))
	}
	total := app.Tracker().TotalStats()
	if total.Frames != 0 {
		t.Errorf("disabled tracker should have 0 frames, got %d", total.Frames)
	}
}

func TestPerfIntegration_Report_NotEmpty(t *testing.T) {
	app, _ := NewTestApp(10, 5)
	app.Tracker().Enable()

	app.RegisterComponent("c", "c", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			vn := layout.NewVNode("text")
			vn.Content = "A"
			return vn
		})

	app.RenderAll()

	report := app.Tracker().Report()
	if len(report) < 100 {
		t.Errorf("Report too short: %q", report)
	}

	totalReport := app.Tracker().TotalReport()
	if len(totalReport) < 50 {
		t.Errorf("TotalReport too short: %q", totalReport)
	}
}

func TestPerfIntegration_RegisterUnregister_Counters(t *testing.T) {
	app, _ := NewTestApp(10, 5)
	app.Tracker().Enable()

	app.Tracker().BeginFrame()
	app.RegisterComponent("c1", "c1", buffer.Rect{X: 0, Y: 0, W: 5, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			return layout.NewVNode("text")
		})
	app.RegisterComponent("c2", "c2", buffer.Rect{X: 5, Y: 0, W: 5, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			return layout.NewVNode("text")
		})
	app.UnregisterComponent("c1")
	app.Tracker().EndFrame()

	app.Tracker().AssertLastFrame(t,
		perf.CheckMetric(perf.ComponentsRegistered, 2),
		perf.CheckMetric(perf.ComponentsUnregistered, 1),
	)
}
