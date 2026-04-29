package devtools

import (
	"testing"
	"time"

	"github.com/akzj/lumina/pkg/perf"
)

func TestPanel_Toggle(t *testing.T) {
	tracker := perf.NewTracker(10)
	p := NewPanel(tracker)

	if p.Visible {
		t.Error("panel should start hidden")
	}
	p.Toggle()
	if !p.Visible {
		t.Error("panel should be visible after toggle")
	}
	p.Toggle()
	if p.Visible {
		t.Error("panel should be hidden after second toggle")
	}
}

func TestPanel_SetTab(t *testing.T) {
	tracker := perf.NewTracker(10)
	p := NewPanel(tracker)

	if p.ActiveTab != TabElements {
		t.Errorf("default tab should be Elements, got %d", p.ActiveTab)
	}
	p.ArmElementsPick()
	p.SetTab(TabPerf)
	if p.ActiveTab != TabPerf {
		t.Errorf("expected TabPerf, got %d", p.ActiveTab)
	}
	if p.ElementsPickArmed() {
		t.Error("switching away from Elements should disarm inspect pick")
	}
	p.SetTab(TabElements)
	if p.ActiveTab != TabElements {
		t.Errorf("expected TabElements, got %d", p.ActiveTab)
	}
}

func TestPanel_ElementsScrollClamp(t *testing.T) {
	tracker := perf.NewTracker(10)
	p := NewPanel(tracker)
	p.Height = 24
	p.Width = 80
	tree := make([]NodeInfo, 50)
	for i := range tree {
		tree[i] = NodeInfo{Type: "box", Depth: 0}
	}
	p.UpdateNodeTree(tree)
	p.elementsScrollY = 1000
	p.ScrollElements(0)
	if p.elementsScrollY < 0 || p.elementsScrollY > len(tree) {
		t.Errorf("scroll should clamp into valid range, got scrollY=%d", p.elementsScrollY)
	}
}

func TestPanel_OnElementsTreeRebuiltClampsSelection(t *testing.T) {
	tracker := perf.NewTracker(10)
	p := NewPanel(tracker)
	p.UpdateNodeTree([]NodeInfo{{Type: "box"}, {Type: "box"}, {Type: "box"}})
	p.SetElementsSelection(2)
	p.UpdateNodeTree([]NodeInfo{{Type: "box"}})
	p.OnElementsTreeRebuilt(1)
	if p.ElementsSelectedIdx() != 0 {
		t.Errorf("selection should clamp to last index, got %d", p.ElementsSelectedIdx())
	}
}

func TestPanel_TickFPS(t *testing.T) {
	tracker := perf.NewTracker(10)
	p := NewPanel(tracker)

	// Initially FPS should be 0
	if p.FPS() != 0 {
		t.Errorf("initial FPS should be 0, got %d", p.FPS())
	}

	// Simulate 60 ticks over 1 second (300ms measurement window)
	p.fpsLastTime = time.Now().Add(-350 * time.Millisecond)
	p.fpsFrameCount = 20 // ~57 fps
	p.TickFPS()

	fps := p.FPS()
	if fps < 40 || fps > 80 {
		t.Errorf("FPS should be ~57, got %d", fps)
	}
}

func TestPanel_SnapshotPerf(t *testing.T) {
	tracker := perf.NewTracker(10)
	tracker.Enable()
	tracker.BeginFrame()
	tracker.Record(perf.Renders, 5)
	tracker.EndFrame()

	p := NewPanel(tracker)
	p.fps = 60

	p.SnapshotPerf()

	if p.perfSnap.Last.Get(perf.Renders) != 5 {
		t.Errorf("snapshot last renders = %d, want 5", p.perfSnap.Last.Get(perf.Renders))
	}
	if p.perfSnap.FPS != 60 {
		t.Errorf("snapshot FPS = %d, want 60", p.perfSnap.FPS)
	}
}

func TestPanel_UpdateComponents(t *testing.T) {
	tracker := perf.NewTracker(10)
	p := NewPanel(tracker)

	if len(p.Components()) != 0 {
		t.Errorf("initial components should be empty")
	}

	p.UpdateComponents([]ComponentInfo{
		{ID: "counter", Name: "Counter", X: 0, Y: 0, W: 40, H: 10, ZIndex: 0},
	})

	if len(p.Components()) != 1 {
		t.Errorf("expected 1 component, got %d", len(p.Components()))
	}
	if p.Components()[0].ID != "counter" {
		t.Errorf("expected component ID 'counter', got %q", p.Components()[0].ID)
	}
}
