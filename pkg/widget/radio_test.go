package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestRadioRenderUnselected(t *testing.T) {
	state := Radio.NewState()
	props := map[string]any{"label": "Option A", "value": "a"}
	result := Radio.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Radio.Render returned nil or non-*Node")
	}
	if node.Type != "hbox" {
		t.Errorf("expected root type 'hbox', got %q", node.Type)
	}
	if len(node.Children) != 2 {
		t.Fatalf("expected 2 children (indicator + label), got %d", len(node.Children))
	}
	indicator := node.Children[0]
	if indicator.Content != "( )" {
		t.Errorf("unselected indicator: got %q, want '( )'", indicator.Content)
	}
	label := node.Children[1]
	if label.Content != " Option A" {
		t.Errorf("label: got %q, want ' Option A'", label.Content)
	}
	if !node.Focusable {
		t.Error("expected Focusable=true")
	}
}

func TestRadioRenderSelected(t *testing.T) {
	state := Radio.NewState().(*RadioState)
	state.Checked = true
	props := map[string]any{"label": "Option B", "value": "b"}
	node := Radio.Render(props, state).(*render.Node)
	indicator := node.Children[0]
	if indicator.Content != "(●)" {
		t.Errorf("selected indicator: got %q, want '(●)'", indicator.Content)
	}
}

func TestRadioControlledMode(t *testing.T) {
	state := Radio.NewState().(*RadioState)
	state.Checked = false
	props := map[string]any{"label": "Ctrl", "value": "x", "checked": true}
	node := Radio.Render(props, state).(*render.Node)
	indicator := node.Children[0]
	if indicator.Content != "(●)" {
		t.Errorf("controlled selected: got %q, want '(●)'", indicator.Content)
	}

	// Reverse: internal true, prop false
	state.Checked = true
	props["checked"] = false
	node = Radio.Render(props, state).(*render.Node)
	indicator = node.Children[0]
	if indicator.Content != "( )" {
		t.Errorf("controlled unselected: got %q, want '( )'", indicator.Content)
	}
}

func TestRadioClickSelects(t *testing.T) {
	state := Radio.NewState().(*RadioState)
	props := map[string]any{"label": "Choice", "value": "choice1"}

	// Click → selected, fires onChange with value
	evt := &render.WidgetEvent{Type: "click"}
	changed := Radio.OnEvent(props, state, evt)
	if !changed {
		t.Error("click should return true")
	}
	if !state.Checked {
		t.Error("expected Checked=true after click")
	}
	if evt.FireOnChange != "choice1" {
		t.Errorf("FireOnChange should be 'choice1', got %v", evt.FireOnChange)
	}
}

func TestRadioClickAlreadySelectedStillFiresOnChange(t *testing.T) {
	state := Radio.NewState().(*RadioState)
	state.Checked = true
	props := map[string]any{"label": "X", "value": "val"}

	// Click when already checked → still fires onChange (for group management)
	evt := &render.WidgetEvent{Type: "click"}
	changed := Radio.OnEvent(props, state, evt)
	if !changed {
		t.Error("click on already-selected radio should return true")
	}
	if evt.FireOnChange != "val" {
		t.Errorf("FireOnChange should be 'val', got %v", evt.FireOnChange)
	}
	// Should stay checked (never toggles off)
	if !state.Checked {
		t.Error("radio should never toggle off via click")
	}
}

func TestRadioKeydownSpace(t *testing.T) {
	state := Radio.NewState().(*RadioState)
	props := map[string]any{"label": "Key", "value": "kv"}

	evt := &render.WidgetEvent{Type: "keydown", Key: " "}
	changed := Radio.OnEvent(props, state, evt)
	if !changed {
		t.Error("Space keydown should select")
	}
	if !state.Checked {
		t.Error("expected Checked=true after Space")
	}
	if evt.FireOnChange != "kv" {
		t.Errorf("FireOnChange should be 'kv', got %v", evt.FireOnChange)
	}
}

func TestRadioKeydownOtherIgnored(t *testing.T) {
	state := Radio.NewState().(*RadioState)
	props := map[string]any{"label": "Key", "value": "x"}

	evt := &render.WidgetEvent{Type: "keydown", Key: "a"}
	changed := Radio.OnEvent(props, state, evt)
	if changed {
		t.Error("'a' keydown should be ignored")
	}
}

func TestRadioDisabled(t *testing.T) {
	state := Radio.NewState().(*RadioState)
	props := map[string]any{"label": "No", "value": "n", "disabled": true}

	node := Radio.Render(props, state).(*render.Node)
	if node.Focusable {
		t.Error("disabled radio should not be focusable")
	}
	if !node.Disabled {
		t.Error("disabled radio should have Disabled=true")
	}
	label := node.Children[1]
	if !label.Style.Dim {
		t.Error("disabled label should be dim")
	}

	evt := &render.WidgetEvent{Type: "click"}
	changed := Radio.OnEvent(props, state, evt)
	if changed {
		t.Error("disabled radio should ignore click")
	}
	if state.Checked {
		t.Error("disabled radio should not select")
	}
}

func TestRadioHover(t *testing.T) {
	state := Radio.NewState().(*RadioState)
	props := map[string]any{"label": "Hover", "value": "h"}

	changed := Radio.OnEvent(props, state, &render.WidgetEvent{Type: "mouseenter"})
	if !changed {
		t.Error("mouseenter should return true")
	}
	if !state.Hovered {
		t.Error("expected Hovered=true")
	}

	// Render with hover → indicator color changes
	node := Radio.Render(props, state).(*render.Node)
	indicator := node.Children[0]
	if indicator.Style.Foreground != "#B4BEFE" {
		t.Errorf("hovered indicator fg: got %q, want '#B4BEFE'", indicator.Style.Foreground)
	}

	changed = Radio.OnEvent(props, state, &render.WidgetEvent{Type: "mouseleave"})
	if !changed {
		t.Error("mouseleave should return true")
	}
	if state.Hovered {
		t.Error("expected Hovered=false")
	}
}

func TestRadioNoLabel(t *testing.T) {
	state := Radio.NewState()
	props := map[string]any{"value": "x"}
	node := Radio.Render(props, state).(*render.Node)
	if len(node.Children) != 1 {
		t.Errorf("no label: expected 1 child, got %d", len(node.Children))
	}
}

func TestRadioParentPointers(t *testing.T) {
	state := Radio.NewState()
	props := map[string]any{"label": "Test", "value": "t"}
	node := Radio.Render(props, state).(*render.Node)
	for i, ch := range node.Children {
		if ch.Parent != node {
			t.Errorf("child %d Parent not set", i)
		}
	}
}
