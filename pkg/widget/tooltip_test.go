package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestTooltipRenderDefault(t *testing.T) {
	state := Tooltip.NewState()
	props := map[string]any{"label": "Hover me", "text": "Help text"}
	result := Tooltip.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Tooltip.Render returned nil or non-*Node")
	}
	if node.Type != "vbox" {
		t.Errorf("expected root type 'vbox', got %q", node.Type)
	}
	// Not hovered → only trigger text (1 child)
	if len(node.Children) != 1 {
		t.Fatalf("expected 1 child (trigger only), got %d", len(node.Children))
	}
	trigger := node.Children[0]
	if trigger.Content != "Hover me" {
		t.Errorf("trigger content: got %q, want 'Hover me'", trigger.Content)
	}
}

func TestTooltipRenderHovered(t *testing.T) {
	state := Tooltip.NewState().(*TooltipState)
	state.Hovered = true
	props := map[string]any{"label": "Hover me", "text": "Help text"}
	node := Tooltip.Render(props, state).(*render.Node)

	// Hovered → tooltip text + trigger text (2 children)
	if len(node.Children) != 2 {
		t.Fatalf("expected 2 children (tooltip+trigger), got %d", len(node.Children))
	}
	tipNode := node.Children[0]
	if tipNode.Content != " Help text " {
		t.Errorf("tooltip content: got %q, want ' Help text '", tipNode.Content)
	}
	trigger := node.Children[1]
	if trigger.Content != "Hover me" {
		t.Errorf("trigger content: got %q, want 'Hover me'", trigger.Content)
	}
}

func TestTooltipMouseEnter(t *testing.T) {
	state := Tooltip.NewState().(*TooltipState)
	props := map[string]any{"label": "L", "text": "T"}

	evt := &render.WidgetEvent{Type: "mouseenter"}
	changed := Tooltip.OnEvent(props, state, evt)
	if !changed {
		t.Error("mouseenter should return true")
	}
	if !state.Hovered {
		t.Error("expected Hovered=true after mouseenter")
	}

	// Duplicate mouseenter should return false
	changed = Tooltip.OnEvent(props, state, &render.WidgetEvent{Type: "mouseenter"})
	if changed {
		t.Error("duplicate mouseenter should return false")
	}
}

func TestTooltipMouseLeave(t *testing.T) {
	state := Tooltip.NewState().(*TooltipState)
	state.Hovered = true
	props := map[string]any{"label": "L", "text": "T"}

	evt := &render.WidgetEvent{Type: "mouseleave"}
	changed := Tooltip.OnEvent(props, state, evt)
	if !changed {
		t.Error("mouseleave should return true")
	}
	if state.Hovered {
		t.Error("expected Hovered=false after mouseleave")
	}

	// Duplicate mouseleave should return false
	changed = Tooltip.OnEvent(props, state, &render.WidgetEvent{Type: "mouseleave"})
	if changed {
		t.Error("duplicate mouseleave should return false")
	}
}

func TestTooltipCustomLabel(t *testing.T) {
	state := Tooltip.NewState()
	props := map[string]any{"label": "Click here"}
	node := Tooltip.Render(props, state).(*render.Node)
	trigger := node.Children[0]
	if trigger.Content != "Click here" {
		t.Errorf("label: got %q, want 'Click here'", trigger.Content)
	}
}

func TestTooltipCustomText(t *testing.T) {
	state := Tooltip.NewState().(*TooltipState)
	state.Hovered = true
	props := map[string]any{"label": "L", "text": "Custom tooltip"}
	node := Tooltip.Render(props, state).(*render.Node)
	tipNode := node.Children[0]
	if tipNode.Content != " Custom tooltip " {
		t.Errorf("tooltip text: got %q, want ' Custom tooltip '", tipNode.Content)
	}
}

func TestTooltipDefaultProps(t *testing.T) {
	state := Tooltip.NewState().(*TooltipState)
	state.Hovered = true
	props := map[string]any{} // no label or text
	node := Tooltip.Render(props, state).(*render.Node)
	tipNode := node.Children[0]
	if tipNode.Content != " Tooltip " {
		t.Errorf("default tooltip text: got %q, want ' Tooltip '", tipNode.Content)
	}
	trigger := node.Children[1]
	if trigger.Content != "hover me" {
		t.Errorf("default label: got %q, want 'hover me'", trigger.Content)
	}
}

func TestTooltipThemeColors(t *testing.T) {
	th := CurrentTheme

	// Not hovered → text color
	state := Tooltip.NewState()
	props := map[string]any{"label": "L", "text": "T"}
	node := Tooltip.Render(props, state).(*render.Node)
	trigger := node.Children[0]
	if trigger.Style.Foreground != th.Text {
		t.Errorf("trigger fg (not hovered): got %q, want %q", trigger.Style.Foreground, th.Text)
	}

	// Hovered → primary color on trigger, surface1 bg on tooltip
	stateH := Tooltip.NewState().(*TooltipState)
	stateH.Hovered = true
	node = Tooltip.Render(props, stateH).(*render.Node)
	tipNode := node.Children[0]
	if tipNode.Style.Background != th.Surface1 {
		t.Errorf("tooltip bg: got %q, want %q", tipNode.Style.Background, th.Surface1)
	}
	if tipNode.Style.Foreground != th.Text {
		t.Errorf("tooltip fg: got %q, want %q", tipNode.Style.Foreground, th.Text)
	}
	triggerH := node.Children[1]
	if triggerH.Style.Foreground != th.Primary {
		t.Errorf("trigger fg (hovered): got %q, want %q", triggerH.Style.Foreground, th.Primary)
	}
}

func TestTooltipParentPointers(t *testing.T) {
	state := Tooltip.NewState().(*TooltipState)
	state.Hovered = true
	props := map[string]any{"label": "L", "text": "T"}
	node := Tooltip.Render(props, state).(*render.Node)
	for i, ch := range node.Children {
		if ch.Parent != node {
			t.Errorf("child %d Parent not set correctly", i)
		}
	}
}

func TestTooltipIgnoresOtherEvents(t *testing.T) {
	state := Tooltip.NewState().(*TooltipState)
	props := map[string]any{"label": "L", "text": "T"}

	for _, evtType := range []string{"click", "keydown", "focus", "blur"} {
		changed := Tooltip.OnEvent(props, state, &render.WidgetEvent{Type: evtType})
		if changed {
			t.Errorf("event %q should return false", evtType)
		}
	}
}

func TestTooltipWidgetDefInterface(t *testing.T) {
	var w render.WidgetDef = Tooltip
	if w.GetName() != "Tooltip" {
		t.Errorf("GetName: got %q, want 'Tooltip'", w.GetName())
	}
	s := w.GetNewState()
	if s == nil {
		t.Fatal("GetNewState returned nil")
	}
	props := map[string]any{"label": "L", "text": "T"}
	result := w.DoRender(props, s)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}
}
