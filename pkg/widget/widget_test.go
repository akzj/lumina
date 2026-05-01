package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestWidgetDefInterface(t *testing.T) {
	// Verify Checkbox satisfies render.WidgetDef interface methods
	var w interface {
		GetName() string
		GetNewState() any
		DoRender(props map[string]any, state any) any
		DoOnEvent(props map[string]any, state any, event *render.WidgetEvent) bool
	} = Checkbox

	if w.GetName() != "Checkbox" {
		t.Errorf("GetName() = %q, want 'Checkbox'", w.GetName())
	}

	state := w.GetNewState()
	if state == nil {
		t.Fatal("GetNewState() returned nil")
	}

	result := w.DoRender(map[string]any{"label": "Test"}, state)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}

	changed := w.DoOnEvent(map[string]any{"label": "Test"}, state, &render.WidgetEvent{Type: "mouseenter"})
	// Checkbox may or may not change on mouseenter, just verify no panic
	_ = changed
}

func TestAllWidgets(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatal("All() returned empty")
	}
	if all[0].Name != "Checkbox" {
		t.Errorf("first widget should be Checkbox, got %q", all[0].Name)
	}
}
