package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func TestPaginationRenderDefault(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(1), "totalPages": float64(1)}
	result := Pagination.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Pagination.Render returned nil or non-*Node")
	}
	if node.Type != "hbox" {
		t.Errorf("expected root type 'hbox', got %q", node.Type)
	}
	if !node.Focusable {
		t.Error("pagination should be focusable")
	}
	// Should have: ◀ + 1 page + ▶ = 3 children
	if len(node.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(node.Children))
	}
	// Prev button
	if node.Children[0].Content != " ◀ " {
		t.Errorf("prev button: got %q, want ' ◀ '", node.Children[0].Content)
	}
	// Page 1
	if node.Children[1].Content != " 1 " {
		t.Errorf("page button: got %q, want ' 1 '", node.Children[1].Content)
	}
	// Next button
	if node.Children[2].Content != " ▶ " {
		t.Errorf("next button: got %q, want ' ▶ '", node.Children[2].Content)
	}
}

func TestPaginationRenderMiddle(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(3), "totalPages": float64(5)}
	node := Pagination.Render(props, state).(*render.Node)

	// ◀ + 5 pages + ▶ = 7 children
	if len(node.Children) != 7 {
		t.Fatalf("expected 7 children, got %d", len(node.Children))
	}
}

func TestPaginationCurrentPageHighlighted(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(3), "totalPages": float64(5)}
	node := Pagination.Render(props, state).(*render.Node)

	// Page 3 is at index 3 (◀ + page1 + page2 + page3)
	page3 := node.Children[3]
	if page3.Content != " 3 " {
		t.Errorf("page 3 content: got %q, want ' 3 '", page3.Content)
	}
	if page3.Style.Background != CurrentTheme.Primary {
		t.Errorf("current page bg: got %q, want %q", page3.Style.Background, CurrentTheme.Primary)
	}
	if page3.Style.Foreground != CurrentTheme.PrimaryDark {
		t.Errorf("current page fg: got %q, want %q", page3.Style.Foreground, CurrentTheme.PrimaryDark)
	}
	if !page3.Style.Bold {
		t.Error("current page should be bold")
	}

	// Other pages should NOT have Primary bg
	page2 := node.Children[2]
	if page2.Style.Background == CurrentTheme.Primary {
		t.Error("non-current page should not have primary bg")
	}
}

func TestPaginationPrevDisabledOnFirstPage(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(1), "totalPages": float64(5)}
	node := Pagination.Render(props, state).(*render.Node)

	prev := node.Children[0]
	if prev.Style.Foreground != CurrentTheme.Muted {
		t.Errorf("prev on page 1 fg: got %q, want %q (muted)", prev.Style.Foreground, CurrentTheme.Muted)
	}
}

func TestPaginationNextDisabledOnLastPage(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(5), "totalPages": float64(5)}
	node := Pagination.Render(props, state).(*render.Node)

	next := node.Children[len(node.Children)-1]
	if next.Style.Foreground != CurrentTheme.Muted {
		t.Errorf("next on last page fg: got %q, want %q (muted)", next.Style.Foreground, CurrentTheme.Muted)
	}
}

func TestPaginationPrevEnabledOnMiddlePage(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(3), "totalPages": float64(5)}
	node := Pagination.Render(props, state).(*render.Node)

	prev := node.Children[0]
	if prev.Style.Foreground != CurrentTheme.Text {
		t.Errorf("prev on page 3 fg: got %q, want %q (text)", prev.Style.Foreground, CurrentTheme.Text)
	}
	next := node.Children[len(node.Children)-1]
	if next.Style.Foreground != CurrentTheme.Text {
		t.Errorf("next on page 3 fg: got %q, want %q (text)", next.Style.Foreground, CurrentTheme.Text)
	}
}

func TestPaginationKeydownArrowRight(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(2), "totalPages": float64(5)}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowRight"}
	changed := Pagination.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowRight should return true")
	}
	if evt.FireOnChange != 3 {
		t.Errorf("FireOnChange: got %v, want 3", evt.FireOnChange)
	}
}

func TestPaginationKeydownArrowLeft(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(3), "totalPages": float64(5)}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowLeft"}
	changed := Pagination.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowLeft should return true")
	}
	if evt.FireOnChange != 2 {
		t.Errorf("FireOnChange: got %v, want 2", evt.FireOnChange)
	}
}

func TestPaginationKeydownArrowLeftAtStart(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(1), "totalPages": float64(5)}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowLeft"}
	changed := Pagination.OnEvent(props, state, evt)
	if changed {
		t.Error("ArrowLeft at page 1 should return false")
	}
}

func TestPaginationKeydownArrowRightAtEnd(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(5), "totalPages": float64(5)}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowRight"}
	changed := Pagination.OnEvent(props, state, evt)
	if changed {
		t.Error("ArrowRight at last page should return false")
	}
}

func TestPaginationKeydownH(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(3), "totalPages": float64(5)}

	evt := &render.WidgetEvent{Type: "keydown", Key: "h"}
	changed := Pagination.OnEvent(props, state, evt)
	if !changed {
		t.Error("h should return true")
	}
	if evt.FireOnChange != 2 {
		t.Errorf("FireOnChange: got %v, want 2", evt.FireOnChange)
	}
}

func TestPaginationKeydownL(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(3), "totalPages": float64(5)}

	evt := &render.WidgetEvent{Type: "keydown", Key: "l"}
	changed := Pagination.OnEvent(props, state, evt)
	if !changed {
		t.Error("l should return true")
	}
	if evt.FireOnChange != 4 {
		t.Errorf("FireOnChange: got %v, want 4", evt.FireOnChange)
	}
}

func TestPaginationMaxVisible(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{
		"page":       float64(5),
		"totalPages": float64(10),
		"maxVisible": float64(3),
	}
	node := Pagination.Render(props, state).(*render.Node)

	// ◀ + 3 pages + ▶ = 5 children
	if len(node.Children) != 5 {
		t.Fatalf("expected 5 children, got %d", len(node.Children))
	}
}

func TestPaginationCalcPageRange(t *testing.T) {
	tests := []struct {
		name       string
		current    int
		total      int
		maxVisible int
		wantStart  int
		wantEnd    int
	}{
		{"all visible", 3, 5, 5, 1, 5},
		{"total < maxVisible", 2, 3, 5, 1, 3},
		{"middle of range", 5, 10, 5, 3, 7},
		{"near start", 2, 10, 5, 1, 5},
		{"near end", 9, 10, 5, 6, 10},
		{"at start", 1, 10, 5, 1, 5},
		{"at end", 10, 10, 5, 6, 10},
		{"single page", 1, 1, 5, 1, 1},
		{"even maxVisible", 5, 10, 4, 3, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := calcPageRange(tt.current, tt.total, tt.maxVisible)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("calcPageRange(%d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.current, tt.total, tt.maxVisible, start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestPaginationThemeColors(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(1), "totalPages": float64(3)}
	node := Pagination.Render(props, state).(*render.Node)

	// Current page (index 1) should use Primary
	page1 := node.Children[1]
	if page1.Style.Background != CurrentTheme.Primary {
		t.Errorf("current page bg: got %q, want %q", page1.Style.Background, CurrentTheme.Primary)
	}
}

func TestPaginationParentPointers(t *testing.T) {
	state := Pagination.NewState()
	props := map[string]any{"page": float64(2), "totalPages": float64(5)}
	node := Pagination.Render(props, state).(*render.Node)
	checkParents(t, node)
}

func TestPaginationWidgetDefInterface(t *testing.T) {
	var w interface {
		GetName() string
		GetNewState() any
		DoRender(props map[string]any, state any) any
		DoOnEvent(props map[string]any, state any, event *render.WidgetEvent) bool
	} = Pagination

	if w.GetName() != "Pagination" {
		t.Errorf("GetName() = %q, want 'Pagination'", w.GetName())
	}

	state := w.GetNewState()
	if state == nil {
		t.Fatal("GetNewState() returned nil")
	}

	props := map[string]any{"page": float64(1), "totalPages": float64(5)}
	result := w.DoRender(props, state)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}
}
