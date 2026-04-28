package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestToastRenderHidden(t *testing.T) {
	state := Toast.NewState()
	props := map[string]any{"visible": false, "message": "Hello"}
	result := Toast.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Toast.Render returned nil or non-*Node")
	}
	if node.Style.Width != 0 || node.Style.Height != 0 {
		t.Errorf("hidden toast should be zero-size, got %dx%d", node.Style.Width, node.Style.Height)
	}
}

func TestToastRenderDefault(t *testing.T) {
	state := Toast.NewState()
	props := map[string]any{"visible": true, "message": "Hello world"}
	result := Toast.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Toast.Render returned nil or non-*Node")
	}
	if node.Type != "hbox" {
		t.Errorf("expected root type 'hbox', got %q", node.Type)
	}
	if node.Style.Border != "rounded" {
		t.Errorf("expected border 'rounded', got %q", node.Style.Border)
	}
	// Should have icon + message + close = 3 children
	if len(node.Children) != 3 {
		t.Fatalf("expected 3 children (icon+message+close), got %d", len(node.Children))
	}
	// Default variant → info icon
	iconNode := node.Children[0]
	if iconNode.Content != "ℹ " {
		t.Errorf("default icon: got %q, want 'ℹ '", iconNode.Content)
	}
	msgNode := node.Children[1]
	if msgNode.Content != "Hello world" {
		t.Errorf("message: got %q, want 'Hello world'", msgNode.Content)
	}
	closeNode := node.Children[2]
	if closeNode.Content != " ✕" {
		t.Errorf("close: got %q, want ' ✕'", closeNode.Content)
	}
}

func TestToastRenderDefaultMessage(t *testing.T) {
	state := Toast.NewState()
	props := map[string]any{"visible": true}
	node := Toast.Render(props, state).(*render.Node)
	msgNode := node.Children[1]
	if msgNode.Content != "Notification" {
		t.Errorf("default message: got %q, want 'Notification'", msgNode.Content)
	}
}

func TestToastRenderSuccess(t *testing.T) {
	th := CurrentTheme
	state := Toast.NewState()
	props := map[string]any{"visible": true, "message": "Done", "variant": "success"}
	node := Toast.Render(props, state).(*render.Node)
	iconNode := node.Children[0]
	if iconNode.Content != "✓ " {
		t.Errorf("success icon: got %q, want '✓ '", iconNode.Content)
	}
	if iconNode.Style.Foreground != th.Success {
		t.Errorf("success fg: got %q, want %q", iconNode.Style.Foreground, th.Success)
	}
}

func TestToastRenderWarning(t *testing.T) {
	th := CurrentTheme
	state := Toast.NewState()
	props := map[string]any{"visible": true, "message": "Caution", "variant": "warning"}
	node := Toast.Render(props, state).(*render.Node)
	iconNode := node.Children[0]
	if iconNode.Content != "⚠ " {
		t.Errorf("warning icon: got %q, want '⚠ '", iconNode.Content)
	}
	if iconNode.Style.Foreground != th.Warning {
		t.Errorf("warning fg: got %q, want %q", iconNode.Style.Foreground, th.Warning)
	}
}

func TestToastRenderError(t *testing.T) {
	th := CurrentTheme
	state := Toast.NewState()
	props := map[string]any{"visible": true, "message": "Failed", "variant": "error"}
	node := Toast.Render(props, state).(*render.Node)
	iconNode := node.Children[0]
	if iconNode.Content != "✗ " {
		t.Errorf("error icon: got %q, want '✗ '", iconNode.Content)
	}
	if iconNode.Style.Foreground != th.Error {
		t.Errorf("error fg: got %q, want %q", iconNode.Style.Foreground, th.Error)
	}
}

func TestToastClickDismiss(t *testing.T) {
	state := Toast.NewState()
	props := map[string]any{"visible": true, "message": "Hi"}
	evt := &render.WidgetEvent{Type: "click"}
	Toast.OnEvent(props, state, evt)
	if evt.FireOnChange != "dismiss" {
		t.Errorf("click should fire onChange 'dismiss', got %v", evt.FireOnChange)
	}
}

func TestToastEscapeDismiss(t *testing.T) {
	state := Toast.NewState()
	props := map[string]any{"visible": true, "message": "Hi"}
	evt := &render.WidgetEvent{Type: "keydown", Key: "Escape"}
	Toast.OnEvent(props, state, evt)
	if evt.FireOnChange != "dismiss" {
		t.Errorf("Escape should fire onChange 'dismiss', got %v", evt.FireOnChange)
	}
}

func TestToastIgnoresEventsWhenHidden(t *testing.T) {
	state := Toast.NewState()
	props := map[string]any{"visible": false, "message": "Hi"}

	evt := &render.WidgetEvent{Type: "click"}
	changed := Toast.OnEvent(props, state, evt)
	if changed {
		t.Error("hidden toast should not handle click")
	}
	if evt.FireOnChange != nil {
		t.Error("hidden toast should not fire onChange")
	}

	evt2 := &render.WidgetEvent{Type: "keydown", Key: "Escape"}
	changed = Toast.OnEvent(props, state, evt2)
	if changed {
		t.Error("hidden toast should not handle Escape")
	}
	if evt2.FireOnChange != nil {
		t.Error("hidden toast should not fire onChange")
	}
}

func TestToastThemeColors(t *testing.T) {
	th := CurrentTheme
	state := Toast.NewState()
	props := map[string]any{"visible": true, "message": "Test"}
	node := Toast.Render(props, state).(*render.Node)

	if node.Style.Background != th.Surface0 {
		t.Errorf("background: got %q, want %q", node.Style.Background, th.Surface0)
	}
	// Default variant → primary icon color
	iconNode := node.Children[0]
	if iconNode.Style.Foreground != th.Primary {
		t.Errorf("default icon fg: got %q, want %q", iconNode.Style.Foreground, th.Primary)
	}
	msgNode := node.Children[1]
	if msgNode.Style.Foreground != th.Text {
		t.Errorf("message fg: got %q, want %q", msgNode.Style.Foreground, th.Text)
	}
	closeNode := node.Children[2]
	if closeNode.Style.Foreground != th.Muted {
		t.Errorf("close fg: got %q, want %q", closeNode.Style.Foreground, th.Muted)
	}
}

func TestToastParentPointers(t *testing.T) {
	state := Toast.NewState()
	props := map[string]any{"visible": true, "message": "Test"}
	node := Toast.Render(props, state).(*render.Node)
	for i, ch := range node.Children {
		if ch.Parent != node {
			t.Errorf("child %d Parent not set correctly", i)
		}
	}
}

func TestToastWidgetDefInterface(t *testing.T) {
	var w render.WidgetDef = Toast
	if w.GetName() != "Toast" {
		t.Errorf("GetName: got %q, want 'Toast'", w.GetName())
	}
	s := w.GetNewState()
	if s == nil {
		t.Fatal("GetNewState returned nil")
	}
	props := map[string]any{"visible": true}
	result := w.DoRender(props, s)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}
}
