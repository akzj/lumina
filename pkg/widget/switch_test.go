package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestSwitchRenderOff(t *testing.T) {
	state := Switch.NewState()
	props := map[string]any{"label": "Notifications"}
	result := Switch.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Switch.Render returned nil or non-*Node")
	}
	if node.Type != "hbox" {
		t.Errorf("expected root type 'hbox', got %q", node.Type)
	}
	if len(node.Children) != 2 {
		t.Fatalf("expected 2 children (track + label), got %d", len(node.Children))
	}
	track := node.Children[0]
	if track.Content != "[○  ]" {
		t.Errorf("off track: got %q, want '[○  ]'", track.Content)
	}
	if track.Style.Background != "#45475A" {
		t.Errorf("off track bg: got %q, want '#45475A'", track.Style.Background)
	}
	label := node.Children[1]
	if label.Content != " Notifications" {
		t.Errorf("label: got %q, want ' Notifications'", label.Content)
	}
	if !node.Focusable {
		t.Error("expected Focusable=true")
	}
}

func TestSwitchRenderOn(t *testing.T) {
	state := Switch.NewState().(*SwitchState)
	state.Checked = true
	props := map[string]any{"label": "Active"}
	node := Switch.Render(props, state).(*render.Node)
	track := node.Children[0]
	if track.Content != "[  ●]" {
		t.Errorf("on track: got %q, want '[  ●]'", track.Content)
	}
	if track.Style.Background != "#89B4FA" {
		t.Errorf("on track bg: got %q, want '#89B4FA'", track.Style.Background)
	}
}

func TestSwitchControlledMode(t *testing.T) {
	state := Switch.NewState().(*SwitchState)
	state.Checked = false
	props := map[string]any{"label": "Ctrl", "checked": true}
	node := Switch.Render(props, state).(*render.Node)
	track := node.Children[0]
	if track.Content != "[  ●]" {
		t.Errorf("controlled on: got %q, want '[  ●]'", track.Content)
	}
}

func TestSwitchClickToggles(t *testing.T) {
	state := Switch.NewState().(*SwitchState)
	props := map[string]any{"label": "Toggle"}

	// Click → on
	evt := &render.WidgetEvent{Type: "click"}
	changed := Switch.OnEvent(props, state, evt)
	if !changed {
		t.Error("click should return true")
	}
	if !state.Checked {
		t.Error("expected Checked=true after click")
	}
	if evt.FireOnChange != true {
		t.Errorf("FireOnChange should be true, got %v", evt.FireOnChange)
	}

	// Click again → off
	evt2 := &render.WidgetEvent{Type: "click"}
	changed = Switch.OnEvent(props, state, evt2)
	if !changed {
		t.Error("second click should return true")
	}
	if state.Checked {
		t.Error("expected Checked=false after second click")
	}
	if evt2.FireOnChange != false {
		t.Errorf("FireOnChange should be false, got %v", evt2.FireOnChange)
	}
}

func TestSwitchKeydownSpace(t *testing.T) {
	state := Switch.NewState().(*SwitchState)
	props := map[string]any{"label": "Key"}

	evt := &render.WidgetEvent{Type: "keydown", Key: " "}
	changed := Switch.OnEvent(props, state, evt)
	if !changed {
		t.Error("Space keydown should toggle")
	}
	if !state.Checked {
		t.Error("expected Checked=true after Space")
	}
	if evt.FireOnChange != true {
		t.Errorf("FireOnChange should be true, got %v", evt.FireOnChange)
	}
}

func TestSwitchDisabled(t *testing.T) {
	state := Switch.NewState().(*SwitchState)
	props := map[string]any{"label": "Off", "disabled": true}

	node := Switch.Render(props, state).(*render.Node)
	if node.Focusable {
		t.Error("disabled switch should not be focusable")
	}
	if !node.Disabled {
		t.Error("disabled switch should have Disabled=true")
	}
	track := node.Children[0]
	if track.Style.Background != "#313244" {
		t.Errorf("disabled track bg: got %q, want '#313244'", track.Style.Background)
	}
	label := node.Children[1]
	if !label.Style.Dim {
		t.Error("disabled label should be dim")
	}

	evt := &render.WidgetEvent{Type: "click"}
	changed := Switch.OnEvent(props, state, evt)
	if changed {
		t.Error("disabled switch should ignore click")
	}
}

func TestSwitchHover(t *testing.T) {
	state := Switch.NewState().(*SwitchState)
	props := map[string]any{"label": "Hover"}

	changed := Switch.OnEvent(props, state, &render.WidgetEvent{Type: "mouseenter"})
	if !changed {
		t.Error("mouseenter should return true")
	}
	if !state.Hovered {
		t.Error("expected Hovered=true")
	}

	// Render off + hovered
	node := Switch.Render(props, state).(*render.Node)
	track := node.Children[0]
	if track.Style.Background != "#585B70" {
		t.Errorf("hovered off track bg: got %q, want '#585B70'", track.Style.Background)
	}

	// Render on + hovered
	state.Checked = true
	node = Switch.Render(props, state).(*render.Node)
	track = node.Children[0]
	if track.Style.Background != "#B4BEFE" {
		t.Errorf("hovered on track bg: got %q, want '#B4BEFE'", track.Style.Background)
	}

	changed = Switch.OnEvent(props, state, &render.WidgetEvent{Type: "mouseleave"})
	if !changed {
		t.Error("mouseleave should return true")
	}
	if state.Hovered {
		t.Error("expected Hovered=false")
	}
}

func TestSwitchNoLabel(t *testing.T) {
	state := Switch.NewState()
	props := map[string]any{}
	node := Switch.Render(props, state).(*render.Node)
	if len(node.Children) != 1 {
		t.Errorf("no label: expected 1 child, got %d", len(node.Children))
	}
}

func TestSwitchParentPointers(t *testing.T) {
	state := Switch.NewState()
	props := map[string]any{"label": "Test"}
	node := Switch.Render(props, state).(*render.Node)
	for i, ch := range node.Children {
		if ch.Parent != node {
			t.Errorf("child %d Parent not set", i)
		}
	}
}
