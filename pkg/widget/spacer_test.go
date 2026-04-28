package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestSpacerRenderDefault(t *testing.T) {
	state := Spacer.NewState()
	props := map[string]any{}
	result := Spacer.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Spacer.Render returned nil or non-*Node")
	}
	if node.Type != "box" {
		t.Errorf("expected type 'box', got %q", node.Type)
	}
	if node.Style.Flex != 1 {
		t.Errorf("expected Flex=1, got %d", node.Style.Flex)
	}
	if node.Style.Width != 0 {
		t.Errorf("expected Width=0, got %d", node.Style.Width)
	}
	if node.Style.Height != 0 {
		t.Errorf("expected Height=0, got %d", node.Style.Height)
	}
}

func TestSpacerRenderFixedVertical(t *testing.T) {
	state := Spacer.NewState()
	props := map[string]any{"size": float64(5)}
	node := Spacer.Render(props, state).(*render.Node)
	if node.Style.Height != 5 {
		t.Errorf("expected Height=5, got %d", node.Style.Height)
	}
	if node.Style.Width != 0 {
		t.Errorf("expected Width=0, got %d", node.Style.Width)
	}
	if node.Style.Flex != 0 {
		t.Errorf("expected Flex=0 (fixed size), got %d", node.Style.Flex)
	}
}

func TestSpacerRenderFixedHorizontal(t *testing.T) {
	state := Spacer.NewState()
	props := map[string]any{"size": float64(10), "direction": "horizontal"}
	node := Spacer.Render(props, state).(*render.Node)
	if node.Style.Width != 10 {
		t.Errorf("expected Width=10, got %d", node.Style.Width)
	}
	if node.Style.Height != 0 {
		t.Errorf("expected Height=0, got %d", node.Style.Height)
	}
	if node.Style.Flex != 0 {
		t.Errorf("expected Flex=0 (fixed size), got %d", node.Style.Flex)
	}
}

func TestSpacerRenderVerticalExplicit(t *testing.T) {
	state := Spacer.NewState()
	props := map[string]any{"size": float64(3), "direction": "vertical"}
	node := Spacer.Render(props, state).(*render.Node)
	if node.Style.Height != 3 {
		t.Errorf("expected Height=3, got %d", node.Style.Height)
	}
	if node.Style.Width != 0 {
		t.Errorf("expected Width=0, got %d", node.Style.Width)
	}
}

func TestSpacerNoState(t *testing.T) {
	state := Spacer.NewState()
	if state != nil {
		t.Errorf("expected nil state, got %v", state)
	}
}

func TestSpacerNoEvents(t *testing.T) {
	state := Spacer.NewState()
	props := map[string]any{}
	events := []string{"click", "mouseenter", "mouseleave", "keydown"}
	for _, evtType := range events {
		evt := &render.WidgetEvent{Type: evtType}
		changed := Spacer.OnEvent(props, state, evt)
		if changed {
			t.Errorf("Spacer should ignore %q", evtType)
		}
	}
}

func TestSpacerThemeIndependent(t *testing.T) {
	state := Spacer.NewState()
	props := map[string]any{}
	node := Spacer.Render(props, state).(*render.Node)
	if node.Style.Background != "" {
		t.Errorf("Spacer should have no background, got %q", node.Style.Background)
	}
	if node.Style.Foreground != "" {
		t.Errorf("Spacer should have no foreground, got %q", node.Style.Foreground)
	}
}

func TestSpacerZeroSize(t *testing.T) {
	state := Spacer.NewState()
	// size=0 should be treated as flex mode (not fixed)
	props := map[string]any{"size": float64(0)}
	node := Spacer.Render(props, state).(*render.Node)
	if node.Style.Flex != 1 {
		t.Errorf("size=0 should use Flex=1, got %d", node.Style.Flex)
	}
}

func TestSpacerWidgetDefInterface(t *testing.T) {
	var w interface {
		GetName() string
		GetNewState() any
		DoRender(props map[string]any, state any) any
		DoOnEvent(props map[string]any, state any, event *render.WidgetEvent) bool
	} = Spacer

	if w.GetName() != "Spacer" {
		t.Errorf("GetName() = %q, want 'Spacer'", w.GetName())
	}

	props := map[string]any{}
	result := w.DoRender(props, nil)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}

	changed := w.DoOnEvent(props, nil, &render.WidgetEvent{Type: "click"})
	if changed {
		t.Error("DoOnEvent should return false for Spacer")
	}
}
