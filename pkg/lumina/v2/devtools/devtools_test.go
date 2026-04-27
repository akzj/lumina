package devtools

import (
	"strings"
	"testing"
	"time"

	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/perf"
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
	p.SetTab(TabPerf)
	if p.ActiveTab != TabPerf {
		t.Errorf("expected TabPerf, got %d", p.ActiveTab)
	}
	p.SetTab(TabElements)
	if p.ActiveTab != TabElements {
		t.Errorf("expected TabElements, got %d", p.ActiveTab)
	}
}

func TestPanel_RenderElements_Empty(t *testing.T) {
	tracker := perf.NewTracker(10)
	p := NewPanel(tracker)
	p.Visible = true
	p.ActiveTab = TabElements
	p.Width = 80
	p.Height = 20

	vn := p.Render(nil, nil)
	if vn == nil {
		t.Fatal("Render returned nil")
	}
	if vn.Type != "box" {
		t.Errorf("root type = %q, want box", vn.Type)
	}
	if vn.ID != "__devtools_root" {
		t.Errorf("root ID = %q, want __devtools_root", vn.ID)
	}
	// Should have tab bar + elements content
	if len(vn.Children) < 2 {
		t.Errorf("expected at least 2 children (tabbar + content), got %d", len(vn.Children))
	}

	// Elements content should say "No components registered"
	elemBox := vn.Children[1]
	if elemBox.ID != "__devtools_elements" {
		t.Errorf("elements box ID = %q, want __devtools_elements", elemBox.ID)
	}
	found := false
	for _, child := range elemBox.Children {
		if strings.Contains(child.Content, "No components") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'No components registered' text in empty elements tab")
	}
}

func TestPanel_RenderElements_WithComponents(t *testing.T) {
	tracker := perf.NewTracker(10)
	p := NewPanel(tracker)
	p.Visible = true
	p.ActiveTab = TabElements
	p.Width = 80
	p.Height = 20

	// Build a VNode tree
	textNode := layout.NewVNode("text")
	textNode.Content = "Hello"
	boxNode := layout.NewVNode("box")
	boxNode.ID = "main"
	boxNode.AddChild(textNode)

	p.UpdateComponents([]ComponentInfo{
		{ID: "counter", Name: "Counter", X: 0, Y: 0, W: 40, H: 10, ZIndex: 0, VNodeTree: boxNode},
	})

	vn := p.Render(nil, nil)
	if vn == nil {
		t.Fatal("Render returned nil")
	}

	// Elements content should have component info
	elemBox := vn.Children[1]
	foundHeader := false
	foundVNode := false
	for _, child := range elemBox.Children {
		if strings.Contains(child.Content, "Counter") && strings.Contains(child.Content, "counter") {
			foundHeader = true
		}
		if strings.Contains(child.Content, "<box") {
			foundVNode = true
		}
	}
	if !foundHeader {
		t.Error("expected component header with 'Counter [counter]'")
	}
	if !foundVNode {
		t.Error("expected VNode tree with '<box'")
	}
}

func TestPanel_RenderPerf(t *testing.T) {
	tracker := perf.NewTracker(10)
	tracker.Enable()
	tracker.BeginFrame()
	tracker.Record(perf.Renders, 3)
	tracker.Record(perf.Layouts, 3)
	tracker.Record(perf.Paints, 3)
	tracker.EndFrame()

	p := NewPanel(tracker)
	p.Visible = true
	p.ActiveTab = TabPerf
	p.Width = 80
	p.Height = 20

	// Snapshot perf data before render (as the app would do)
	p.SnapshotPerf()

	vn := p.Render(nil, nil)
	if vn == nil {
		t.Fatal("Render returned nil")
	}

	// Should have tab bar + perf content
	if len(vn.Children) < 2 {
		t.Fatalf("expected at least 2 children, got %d", len(vn.Children))
	}

	perfBox := vn.Children[1]
	if perfBox.ID != "__devtools_perf" {
		t.Errorf("perf box ID = %q, want __devtools_perf", perfBox.ID)
	}

	// Check that perf data is rendered
	allText := collectText(perfBox)
	if !strings.Contains(allText, "Renders") {
		t.Error("perf tab should contain 'Renders'")
	}
	if !strings.Contains(allText, "Frame Stats") {
		t.Error("perf tab should contain 'Frame Stats'")
	}
	if !strings.Contains(allText, "Runtime") {
		t.Error("perf tab should contain 'Runtime'")
	}
	if !strings.Contains(allText, "FPS") {
		t.Error("perf tab should contain 'FPS'")
	}
}

func TestPanel_TabBar(t *testing.T) {
	tracker := perf.NewTracker(10)
	p := NewPanel(tracker)
	p.Visible = true
	p.Width = 80
	p.Height = 20

	// Elements tab active
	p.ActiveTab = TabElements
	vn := p.Render(nil, nil)
	tabBar := vn.Children[0]
	if tabBar.ID != "__devtools_tabbar" {
		t.Errorf("tabbar ID = %q, want __devtools_tabbar", tabBar.ID)
	}

	// Check tab styling: Elements should be bold
	elemTab := findByID(tabBar, "__devtools_tab_elements")
	if elemTab == nil {
		t.Fatal("elements tab not found")
	}
	if !elemTab.Style.Bold {
		t.Error("active Elements tab should be bold")
	}

	perfTab := findByID(tabBar, "__devtools_tab_perf")
	if perfTab == nil {
		t.Fatal("perf tab not found")
	}
	if perfTab.Style.Bold {
		t.Error("inactive Perf tab should not be bold")
	}

	// Switch to Perf tab
	p.ActiveTab = TabPerf
	vn = p.Render(nil, nil)
	tabBar = vn.Children[0]
	elemTab = findByID(tabBar, "__devtools_tab_elements")
	perfTab = findByID(tabBar, "__devtools_tab_perf")
	if elemTab.Style.Bold {
		t.Error("inactive Elements tab should not be bold")
	}
	if !perfTab.Style.Bold {
		t.Error("active Perf tab should be bold")
	}
}

func TestPanel_TabBarShowsFPS(t *testing.T) {
	tracker := perf.NewTracker(10)
	p := NewPanel(tracker)
	p.Visible = true
	p.Width = 80
	p.Height = 20

	// Simulate FPS measurement
	p.fpsLastTime = time.Now().Add(-400 * time.Millisecond) // 400ms ago
	p.fpsFrameCount = 24                                     // 24 frames in 400ms = 60fps
	p.TickFPS()
	p.SnapshotPerf()

	vn := p.Render(nil, nil)
	tabBar := vn.Children[0]
	allText := collectText(tabBar)
	if !strings.Contains(allText, "FPS") {
		t.Errorf("tab bar should show FPS, got: %s", allText)
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

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		contains string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
	}
	for _, tc := range tests {
		got := formatBytes(tc.input)
		if got != tc.contains {
			t.Errorf("formatBytes(%d) = %q, want %q", tc.input, got, tc.contains)
		}
	}
}

// --- helpers ---

func collectText(node *layout.VNode) string {
	var sb strings.Builder
	collectTextRecursive(node, &sb)
	return sb.String()
}

func collectTextRecursive(node *layout.VNode, sb *strings.Builder) {
	if node.Content != "" {
		sb.WriteString(node.Content)
		sb.WriteString(" ")
	}
	for _, child := range node.Children {
		collectTextRecursive(child, sb)
	}
}

func findByID(node *layout.VNode, id string) *layout.VNode {
	if node.ID == id {
		return node
	}
	for _, child := range node.Children {
		if found := findByID(child, id); found != nil {
			return found
		}
	}
	return nil
}
