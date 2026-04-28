package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestButtonRenderDefault(t *testing.T) {
	state := Button.NewState()
	props := map[string]any{"label": "OK"}
	result := Button.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Button.Render returned nil or non-*Node")
	}
	if node.Type != "hbox" {
		t.Errorf("expected root type 'hbox', got %q", node.Type)
	}
	if len(node.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(node.Children))
	}
	text := node.Children[0]
	if text.Type != "text" {
		t.Errorf("expected child type 'text', got %q", text.Type)
	}
	if text.Content != "OK" {
		t.Errorf("expected content 'OK', got %q", text.Content)
	}
	if node.Style.Border != "rounded" {
		t.Errorf("expected border 'rounded', got %q", node.Style.Border)
	}
	if node.Style.Background != "#89B4FA" {
		t.Errorf("expected bg '#89B4FA', got %q", node.Style.Background)
	}
	if !node.Focusable {
		t.Error("expected Focusable=true")
	}
	if node.Disabled {
		t.Error("expected Disabled=false")
	}
}

func TestButtonRenderDisabled(t *testing.T) {
	state := Button.NewState()
	props := map[string]any{"label": "Nope", "disabled": true}
	result := Button.Render(props, state)
	node := result.(*render.Node)
	if node.Focusable {
		t.Error("disabled button should not be focusable")
	}
	if !node.Disabled {
		t.Error("disabled button should have Disabled=true")
	}
	if node.Style.Background != "#313244" {
		t.Errorf("disabled bg should be '#313244', got %q", node.Style.Background)
	}
	text := node.Children[0]
	if !text.Style.Dim {
		t.Error("disabled text should be dim")
	}
}

func TestButtonRenderVariants(t *testing.T) {
	tests := []struct {
		variant string
		wantBg  string
		wantFg  string
		border  string
	}{
		{"primary", "#89B4FA", "#1E1E2E", "rounded"},
		{"secondary", "#45475A", "#CDD6F4", "rounded"},
		{"outline", "", "#89B4FA", "rounded"},
		{"ghost", "", "#CDD6F4", "none"},
	}
	for _, tt := range tests {
		t.Run(tt.variant, func(t *testing.T) {
			state := Button.NewState()
			props := map[string]any{"label": "X", "variant": tt.variant}
			node := Button.Render(props, state).(*render.Node)
			if node.Style.Background != tt.wantBg {
				t.Errorf("bg: got %q, want %q", node.Style.Background, tt.wantBg)
			}
			text := node.Children[0]
			if text.Style.Foreground != tt.wantFg {
				t.Errorf("fg: got %q, want %q", text.Style.Foreground, tt.wantFg)
			}
			if node.Style.Border != tt.border {
				t.Errorf("border: got %q, want %q", node.Style.Border, tt.border)
			}
		})
	}
}

func TestButtonHoverState(t *testing.T) {
	state := Button.NewState().(*ButtonState)
	props := map[string]any{"label": "Hover"}

	// mouseenter → hovered
	changed := Button.OnEvent(props, state, &Event{Type: "mouseenter"})
	if !changed {
		t.Error("mouseenter should return true (state changed)")
	}
	if !state.Hovered {
		t.Error("expected Hovered=true after mouseenter")
	}

	// Render with hovered state
	node := Button.Render(props, state).(*render.Node)
	if node.Style.Background != "#B4BEFE" {
		t.Errorf("hovered bg should be '#B4BEFE', got %q", node.Style.Background)
	}

	// Duplicate mouseenter → no change
	changed = Button.OnEvent(props, state, &Event{Type: "mouseenter"})
	if changed {
		t.Error("duplicate mouseenter should return false")
	}

	// mouseleave → not hovered
	changed = Button.OnEvent(props, state, &Event{Type: "mouseleave"})
	if !changed {
		t.Error("mouseleave should return true")
	}
	if state.Hovered {
		t.Error("expected Hovered=false after mouseleave")
	}
}

func TestButtonPressState(t *testing.T) {
	state := Button.NewState().(*ButtonState)
	props := map[string]any{"label": "Press"}

	// mouseenter first
	Button.OnEvent(props, state, &Event{Type: "mouseenter"})

	// mousedown → pressed
	changed := Button.OnEvent(props, state, &Event{Type: "mousedown"})
	if !changed {
		t.Error("mousedown should return true")
	}
	if !state.Pressed {
		t.Error("expected Pressed=true")
	}

	// Render with pressed state
	node := Button.Render(props, state).(*render.Node)
	if node.Style.Background != "#74C7EC" {
		t.Errorf("pressed bg should be '#74C7EC', got %q", node.Style.Background)
	}

	// mouseup → not pressed
	changed = Button.OnEvent(props, state, &Event{Type: "mouseup"})
	if !changed {
		t.Error("mouseup should return true")
	}
	if state.Pressed {
		t.Error("expected Pressed=false after mouseup")
	}
}

func TestButtonDisabledIgnoresEvents(t *testing.T) {
	state := Button.NewState().(*ButtonState)
	props := map[string]any{"label": "Disabled", "disabled": true}

	changed := Button.OnEvent(props, state, &Event{Type: "mouseenter"})
	if changed {
		t.Error("disabled button should ignore mouseenter")
	}
	if state.Hovered {
		t.Error("disabled button should not become hovered")
	}
}

func TestButtonMouseleaveResetsPressed(t *testing.T) {
	state := Button.NewState().(*ButtonState)
	props := map[string]any{"label": "X"}

	Button.OnEvent(props, state, &Event{Type: "mouseenter"})
	Button.OnEvent(props, state, &Event{Type: "mousedown"})
	if !state.Pressed || !state.Hovered {
		t.Fatal("precondition: should be hovered+pressed")
	}

	// mouseleave resets both
	changed := Button.OnEvent(props, state, &Event{Type: "mouseleave"})
	if !changed {
		t.Error("mouseleave should return true")
	}
	if state.Hovered || state.Pressed {
		t.Error("mouseleave should reset both Hovered and Pressed")
	}
}

func TestWidgetDefInterface(t *testing.T) {
	// Verify Button satisfies render.WidgetDef interface methods
	var w interface {
		GetName() string
		GetNewState() any
		DoRender(props map[string]any, state any) any
		DoOnEvent(props map[string]any, state any, eventType, key string, x, y int) bool
	} = Button

	if w.GetName() != "Button" {
		t.Errorf("GetName() = %q, want 'Button'", w.GetName())
	}

	state := w.GetNewState()
	if state == nil {
		t.Fatal("GetNewState() returned nil")
	}

	result := w.DoRender(map[string]any{"label": "Test"}, state)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}

	changed := w.DoOnEvent(map[string]any{"label": "Test"}, state, "mouseenter", "", 0, 0)
	if !changed {
		t.Error("DoOnEvent mouseenter should return true")
	}
}

func TestAllWidgets(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatal("All() returned empty")
	}
	if all[0].Name != "Button" {
		t.Errorf("first widget should be Button, got %q", all[0].Name)
	}
}
