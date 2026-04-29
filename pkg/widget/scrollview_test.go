package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestScrollViewRenderForcesOverflowScroll(t *testing.T) {
	state := ScrollView.NewState()
	props := map[string]any{
		"style": map[string]any{
			"height":   10,
			"overflow": "hidden",
		},
	}
	node := ScrollView.Render(props, state).(*render.Node)
	if node.Style.Overflow != "scroll" {
		t.Errorf("expected overflow scroll, got %q", node.Style.Overflow)
	}
	if !node.Focusable {
		t.Error("expected Focusable root for keyboard scrolling")
	}
}

func TestScrollViewKeydownArrowDown(t *testing.T) {
	state := ScrollView.NewState()
	props := map[string]any{}
	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	if !ScrollView.OnEvent(props, state, evt) {
		t.Fatal("ArrowDown should be consumed")
	}
	if evt.ScrollBy != 1 {
		t.Errorf("ScrollBy: got %d, want 1", evt.ScrollBy)
	}
}

func TestScrollViewKeydownArrowUp(t *testing.T) {
	state := ScrollView.NewState()
	props := map[string]any{}
	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowUp"}
	if !ScrollView.OnEvent(props, state, evt) {
		t.Fatal("ArrowUp should be consumed")
	}
	if evt.ScrollBy != -1 {
		t.Errorf("ScrollBy: got %d, want -1", evt.ScrollBy)
	}
}
