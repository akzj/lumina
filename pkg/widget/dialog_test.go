package widget

import (
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestDialogRenderClosed(t *testing.T) {
	state := Dialog.NewState()
	props := map[string]any{"open": false}
	result := Dialog.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Dialog.Render returned nil or non-*Node")
	}
	if node.Style.Width != 0 || node.Style.Height != 0 {
		t.Errorf("closed dialog should be zero-size, got %dx%d", node.Style.Width, node.Style.Height)
	}
}

func TestDialogRenderOpen(t *testing.T) {
	state := Dialog.NewState()
	props := map[string]any{
		"open":    true,
		"title":   "Confirm",
		"message": "Are you sure?",
	}
	result := Dialog.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Dialog.Render returned nil or non-*Node")
	}
	if node.Type != "vbox" {
		t.Errorf("expected root type 'vbox', got %q", node.Type)
	}
	if node.Style.Border != "rounded" {
		t.Errorf("expected border 'rounded', got %q", node.Style.Border)
	}
	if node.Style.Width != 40 {
		t.Errorf("expected default width 40, got %d", node.Style.Width)
	}
	// Should have title + divider + message = 3 children
	if len(node.Children) != 3 {
		t.Fatalf("expected 3 children (title+divider+message), got %d", len(node.Children))
	}
	titleNode := node.Children[0]
	if titleNode.Content != "Confirm" {
		t.Errorf("title: got %q, want 'Confirm'", titleNode.Content)
	}
	if !titleNode.Style.Bold {
		t.Error("title should be bold")
	}
	msgNode := node.Children[2]
	if msgNode.Content != "Are you sure?" {
		t.Errorf("message: got %q, want 'Are you sure?'", msgNode.Content)
	}
	if !node.Focusable {
		t.Error("open dialog should be focusable")
	}
}

func TestDialogRenderOpenDefaultTitle(t *testing.T) {
	state := Dialog.NewState()
	props := map[string]any{"open": true}
	node := Dialog.Render(props, state).(*render.Node)
	titleNode := node.Children[0]
	if titleNode.Content != "Dialog" {
		t.Errorf("default title: got %q, want 'Dialog'", titleNode.Content)
	}
}

func TestDialogRenderOpenNoMessage(t *testing.T) {
	state := Dialog.NewState()
	props := map[string]any{"open": true, "title": "Info"}
	node := Dialog.Render(props, state).(*render.Node)
	// Should have title + divider = 2 children (no message)
	if len(node.Children) != 2 {
		t.Errorf("expected 2 children (title+divider, no message), got %d", len(node.Children))
	}
}

func TestDialogRenderOpenCustomWidth(t *testing.T) {
	state := Dialog.NewState()
	props := map[string]any{"open": true, "width": float64(60)}
	node := Dialog.Render(props, state).(*render.Node)
	if node.Style.Width != 60 {
		t.Errorf("expected width 60, got %d", node.Style.Width)
	}
}

func TestDialogEscapeFiresClose(t *testing.T) {
	state := Dialog.NewState()
	props := map[string]any{"open": true}
	evt := &render.WidgetEvent{Type: "keydown", Key: "Escape"}
	Dialog.OnEvent(props, state, evt)
	if evt.FireOnChange != "close" {
		t.Errorf("Escape should fire onChange 'close', got %v", evt.FireOnChange)
	}
}

func TestDialogIgnoresEventsWhenClosed(t *testing.T) {
	state := Dialog.NewState()
	props := map[string]any{"open": false}
	evt := &render.WidgetEvent{Type: "keydown", Key: "Escape"}
	changed := Dialog.OnEvent(props, state, evt)
	if changed {
		t.Error("closed dialog should not handle events")
	}
	if evt.FireOnChange != nil {
		t.Error("closed dialog should not fire onChange")
	}
}

func TestDialogThemeColors(t *testing.T) {
	state := Dialog.NewState()
	props := map[string]any{"open": true, "title": "Test", "message": "Body"}
	node := Dialog.Render(props, state).(*render.Node)

	th := CurrentTheme
	if node.Style.Background != th.Surface0 {
		t.Errorf("background: got %q, want %q", node.Style.Background, th.Surface0)
	}
	titleNode := node.Children[0]
	if titleNode.Style.Foreground != th.Primary {
		t.Errorf("title fg: got %q, want %q", titleNode.Style.Foreground, th.Primary)
	}
	dividerNode := node.Children[1]
	if dividerNode.Style.Foreground != th.Surface1 {
		t.Errorf("divider fg: got %q, want %q", dividerNode.Style.Foreground, th.Surface1)
	}
	msgNode := node.Children[2]
	if msgNode.Style.Foreground != th.Text {
		t.Errorf("message fg: got %q, want %q", msgNode.Style.Foreground, th.Text)
	}
}

func TestDialogDividerWidth(t *testing.T) {
	state := Dialog.NewState()
	props := map[string]any{"open": true, "width": float64(50)}
	node := Dialog.Render(props, state).(*render.Node)
	dividerNode := node.Children[1]
	// Divider should be width - 4 (border + padding)
	expected := strings.Repeat("─", 46)
	if dividerNode.Content != expected {
		t.Errorf("divider content length: got %d, want %d", len([]rune(dividerNode.Content)), 46)
	}
}

func TestDialogParentPointers(t *testing.T) {
	state := Dialog.NewState()
	props := map[string]any{"open": true, "title": "Test", "message": "Body"}
	node := Dialog.Render(props, state).(*render.Node)
	for i, ch := range node.Children {
		if ch.Parent != node {
			t.Errorf("child %d Parent not set correctly", i)
		}
	}
}

func TestDialogWidgetDefInterface(t *testing.T) {
	var w render.WidgetDef = Dialog
	if w.GetName() != "Dialog" {
		t.Errorf("GetName: got %q, want 'Dialog'", w.GetName())
	}
	s := w.GetNewState()
	if s == nil {
		t.Fatal("GetNewState returned nil")
	}
	props := map[string]any{"open": true}
	result := w.DoRender(props, s)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}
}
