package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestWidgetDefInterface(t *testing.T) {
	// Verify Label satisfies render.WidgetDef interface methods
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
	// Label has no state (returns nil) — this is valid

	result := w.DoRender(map[string]any{"text": "Test"}, state)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}

	changed := w.DoOnEvent(map[string]any{"text": "Test"}, state, &render.WidgetEvent{Type: "mouseenter"})
	// Label may or may not change on mouseenter, just verify no panic
	_ = changed
}

func TestAllWidgets(t *testing.T) {
	all := All()
	if len(all) != 9 {
		t.Fatalf("All() should return 9 widgets, got %d", len(all))
	}
	if all[0].Name != "Label" {
		t.Errorf("first widget should be Label, got %q", all[0].Name)
	}
}
