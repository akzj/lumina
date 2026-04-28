package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestLabelRenderDefault(t *testing.T) {
	state := Label.NewState()
	props := map[string]any{}
	result := Label.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Label.Render returned nil or non-*Node")
	}
	if node.Type != "text" {
		t.Errorf("expected type 'text', got %q", node.Type)
	}
	if node.Content != "Label" {
		t.Errorf("expected default content 'Label', got %q", node.Content)
	}
	if node.Style.Foreground != "#CDD6F4" {
		t.Errorf("expected fg '#CDD6F4', got %q", node.Style.Foreground)
	}
}

func TestLabelRenderCustomText(t *testing.T) {
	state := Label.NewState()
	props := map[string]any{"text": "Username"}
	node := Label.Render(props, state).(*render.Node)
	if node.Content != "Username" {
		t.Errorf("expected content 'Username', got %q", node.Content)
	}
}

func TestLabelIgnoresAllEvents(t *testing.T) {
	state := Label.NewState()
	props := map[string]any{"text": "Test"}

	events := []string{"click", "mouseenter", "mouseleave", "mousedown", "mouseup", "keydown", "focus", "blur"}
	for _, evtType := range events {
		changed := Label.OnEvent(props, state, &render.WidgetEvent{Type: evtType})
		if changed {
			t.Errorf("Label should ignore %q event", evtType)
		}
	}
}

func TestLabelNilState(t *testing.T) {
	state := Label.NewState()
	if state != nil {
		t.Errorf("Label state should be nil, got %v", state)
	}
}

func TestLabelWidgetDefInterface(t *testing.T) {
	var w interface {
		GetName() string
		GetNewState() any
		DoRender(props map[string]any, state any) any
		DoOnEvent(props map[string]any, state any, event *render.WidgetEvent) bool
	} = Label

	if w.GetName() != "Label" {
		t.Errorf("GetName() = %q, want 'Label'", w.GetName())
	}

	state := w.GetNewState()
	result := w.DoRender(map[string]any{"text": "Hello"}, state)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}
	node := result.(*render.Node)
	if node.Content != "Hello" {
		t.Errorf("DoRender content: got %q, want 'Hello'", node.Content)
	}

	changed := w.DoOnEvent(map[string]any{}, state, &render.WidgetEvent{Type: "click"})
	if changed {
		t.Error("DoOnEvent should return false for Label")
	}
}
