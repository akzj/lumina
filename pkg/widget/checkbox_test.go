package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestCheckboxRenderUnchecked(t *testing.T) {
	state := Checkbox.NewState()
	props := map[string]any{"label": "Accept"}
	result := Checkbox.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Checkbox.Render returned nil or non-*Node")
	}
	if node.Type != "hbox" {
		t.Errorf("expected root type 'hbox', got %q", node.Type)
	}
	if len(node.Children) != 2 {
		t.Fatalf("expected 2 children (indicator + label), got %d", len(node.Children))
	}
	indicator := node.Children[0]
	if indicator.Content != "[ ]" {
		t.Errorf("unchecked indicator: got %q, want '[ ]'", indicator.Content)
	}
	label := node.Children[1]
	if label.Content != " Accept" {
		t.Errorf("label: got %q, want ' Accept'", label.Content)
	}
	if !node.Focusable {
		t.Error("expected Focusable=true")
	}
	if node.Disabled {
		t.Error("expected Disabled=false")
	}
}

func TestCheckboxRenderChecked(t *testing.T) {
	state := Checkbox.NewState().(*CheckboxState)
	state.Checked = true
	props := map[string]any{"label": "Done"}
	node := Checkbox.Render(props, state).(*render.Node)
	indicator := node.Children[0]
	if indicator.Content != "[x]" {
		t.Errorf("checked indicator: got %q, want '[x]'", indicator.Content)
	}
}

func TestCheckboxControlledMode(t *testing.T) {
	state := Checkbox.NewState().(*CheckboxState)
	state.Checked = false
	// Controlled mode: props["checked"] overrides internal state
	props := map[string]any{"label": "Ctrl", "checked": true}
	node := Checkbox.Render(props, state).(*render.Node)
	indicator := node.Children[0]
	if indicator.Content != "[x]" {
		t.Errorf("controlled checked: got %q, want '[x]'", indicator.Content)
	}

	// Reverse: internal state true but prop false
	state.Checked = true
	props["checked"] = false
	node = Checkbox.Render(props, state).(*render.Node)
	indicator = node.Children[0]
	if indicator.Content != "[ ]" {
		t.Errorf("controlled unchecked: got %q, want '[ ]'", indicator.Content)
	}
}

func TestCheckboxClickToggles(t *testing.T) {
	state := Checkbox.NewState().(*CheckboxState)
	props := map[string]any{"label": "Toggle"}

	// Click → checked
	evt := &render.WidgetEvent{Type: "click"}
	changed := Checkbox.OnEvent(props, state, evt)
	if !changed {
		t.Error("click should return true")
	}
	if !state.Checked {
		t.Error("expected Checked=true after click")
	}
	if evt.FireOnChange != true {
		t.Errorf("FireOnChange should be true, got %v", evt.FireOnChange)
	}

	// Click again → unchecked
	evt2 := &render.WidgetEvent{Type: "click"}
	changed = Checkbox.OnEvent(props, state, evt2)
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

func TestCheckboxKeydownSpace(t *testing.T) {
	state := Checkbox.NewState().(*CheckboxState)
	props := map[string]any{"label": "Key"}

	evt := &render.WidgetEvent{Type: "keydown", Key: " "}
	changed := Checkbox.OnEvent(props, state, evt)
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

func TestCheckboxKeydownEnter(t *testing.T) {
	state := Checkbox.NewState().(*CheckboxState)
	props := map[string]any{"label": "Key"}

	evt := &render.WidgetEvent{Type: "keydown", Key: "Enter"}
	changed := Checkbox.OnEvent(props, state, evt)
	if !changed {
		t.Error("Enter keydown should toggle")
	}
	if !state.Checked {
		t.Error("expected Checked=true after Enter")
	}
}

func TestCheckboxKeydownOtherIgnored(t *testing.T) {
	state := Checkbox.NewState().(*CheckboxState)
	props := map[string]any{"label": "Key"}

	evt := &render.WidgetEvent{Type: "keydown", Key: "a"}
	changed := Checkbox.OnEvent(props, state, evt)
	if changed {
		t.Error("'a' keydown should be ignored")
	}
}

func TestCheckboxDisabled(t *testing.T) {
	state := Checkbox.NewState().(*CheckboxState)
	props := map[string]any{"label": "No", "disabled": true}

	// Render: not focusable, disabled
	node := Checkbox.Render(props, state).(*render.Node)
	if node.Focusable {
		t.Error("disabled checkbox should not be focusable")
	}
	if !node.Disabled {
		t.Error("disabled checkbox should have Disabled=true")
	}
	// Label should be dim
	label := node.Children[1]
	if !label.Style.Dim {
		t.Error("disabled label should be dim")
	}

	// Events ignored
	evt := &render.WidgetEvent{Type: "click"}
	changed := Checkbox.OnEvent(props, state, evt)
	if changed {
		t.Error("disabled checkbox should ignore click")
	}
	if state.Checked {
		t.Error("disabled checkbox should not toggle")
	}
}

func TestCheckboxHover(t *testing.T) {
	state := Checkbox.NewState().(*CheckboxState)
	props := map[string]any{"label": "Hover"}

	// mouseenter
	changed := Checkbox.OnEvent(props, state, &render.WidgetEvent{Type: "mouseenter"})
	if !changed {
		t.Error("mouseenter should return true")
	}
	if !state.Hovered {
		t.Error("expected Hovered=true")
	}

	// Duplicate mouseenter
	changed = Checkbox.OnEvent(props, state, &render.WidgetEvent{Type: "mouseenter"})
	if changed {
		t.Error("duplicate mouseenter should return false")
	}

	// Render with hover → indicator color changes
	node := Checkbox.Render(props, state).(*render.Node)
	indicator := node.Children[0]
	if indicator.Style.Foreground != "#B4BEFE" {
		t.Errorf("hovered indicator fg: got %q, want '#B4BEFE'", indicator.Style.Foreground)
	}

	// mouseleave
	changed = Checkbox.OnEvent(props, state, &render.WidgetEvent{Type: "mouseleave"})
	if !changed {
		t.Error("mouseleave should return true")
	}
	if state.Hovered {
		t.Error("expected Hovered=false")
	}
}

func TestCheckboxNoLabel(t *testing.T) {
	state := Checkbox.NewState()
	props := map[string]any{}
	node := Checkbox.Render(props, state).(*render.Node)
	if len(node.Children) != 1 {
		t.Errorf("no label: expected 1 child, got %d", len(node.Children))
	}
}

func TestCheckboxParentPointers(t *testing.T) {
	state := Checkbox.NewState()
	props := map[string]any{"label": "Test"}
	node := Checkbox.Render(props, state).(*render.Node)
	for i, ch := range node.Children {
		if ch.Parent != node {
			t.Errorf("child %d Parent not set", i)
		}
	}
}
