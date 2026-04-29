package widget

import (
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

func testListItems() []any {
	return []any{"Apple", "Banana", "Cherry"}
}

func TestListRenderEmpty(t *testing.T) {
	state := List.NewState()
	props := map[string]any{}
	result := List.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("List.Render returned nil or non-*Node")
	}
	if node.Type != "vbox" {
		t.Errorf("expected root type 'vbox', got %q", node.Type)
	}
	if node.Style.Border != "rounded" {
		t.Errorf("expected border 'rounded', got %q", node.Style.Border)
	}
	if len(node.Children) != 0 {
		t.Errorf("empty list: expected 0 children, got %d", len(node.Children))
	}
}

func TestListRenderItems(t *testing.T) {
	state := List.NewState()
	props := map[string]any{"items": testListItems()}
	node := List.Render(props, state).(*render.Node)

	if len(node.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(node.Children))
	}

	// All items should have "  " prefix (not selected)
	for i, child := range node.Children {
		if !strings.HasPrefix(child.Content, "  ") {
			t.Errorf("item %d: expected '  ' prefix, got %q", i, child.Content)
		}
	}
}

func TestListRenderSelectedItem(t *testing.T) {
	state := List.NewState().(*ListState)
	state.SelectedIndex = 1
	props := map[string]any{"items": testListItems()}
	node := List.Render(props, state).(*render.Node)

	// Selected item should have "● " prefix and primary color
	selected := node.Children[1]
	if !strings.HasPrefix(selected.Content, "● ") {
		t.Errorf("selected item: expected '● ' prefix, got %q", selected.Content)
	}
	if selected.Style.Foreground != CurrentTheme.Primary {
		t.Errorf("selected item fg: got %q, want %q", selected.Style.Foreground, CurrentTheme.Primary)
	}

	// Non-selected should have "  " prefix
	nonSelected := node.Children[0]
	if !strings.HasPrefix(nonSelected.Content, "  ") {
		t.Errorf("non-selected item: expected '  ' prefix, got %q", nonSelected.Content)
	}
	if nonSelected.Style.Foreground != CurrentTheme.Text {
		t.Errorf("non-selected item fg: got %q, want %q", nonSelected.Style.Foreground, CurrentTheme.Text)
	}
}

func TestListRenderShowIndex(t *testing.T) {
	state := List.NewState()
	props := map[string]any{"items": testListItems(), "showIndex": true}
	node := List.Render(props, state).(*render.Node)

	// Items should have "1. ", "2. ", "3. " format
	expected := []string{"  1. Apple", "  2. Banana", "  3. Cherry"}
	for i, child := range node.Children {
		if child.Content != expected[i] {
			t.Errorf("item %d: got %q, want %q", i, child.Content, expected[i])
		}
	}
}

func TestListKeydownArrowDown(t *testing.T) {
	state := List.NewState().(*ListState)
	state.SelectedIndex = 0
	props := map[string]any{"items": testListItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	changed := List.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowDown should return true")
	}
	if state.SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex=1, got %d", state.SelectedIndex)
	}
	if evt.FireOnChange != 2 {
		t.Errorf("FireOnChange: got %v, want 2 (1-based)", evt.FireOnChange)
	}
}

func TestListKeydownArrowDownWraps(t *testing.T) {
	state := List.NewState().(*ListState)
	state.SelectedIndex = 2 // last item
	props := map[string]any{"items": testListItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	List.OnEvent(props, state, evt)
	if state.SelectedIndex != 0 {
		t.Errorf("ArrowDown should wrap to 0, got %d", state.SelectedIndex)
	}
}

func TestListKeydownArrowUp(t *testing.T) {
	state := List.NewState().(*ListState)
	state.SelectedIndex = 2
	props := map[string]any{"items": testListItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowUp"}
	changed := List.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowUp should return true")
	}
	if state.SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex=1, got %d", state.SelectedIndex)
	}
}

func TestListKeydownArrowUpWraps(t *testing.T) {
	state := List.NewState().(*ListState)
	state.SelectedIndex = 0
	props := map[string]any{"items": testListItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowUp"}
	List.OnEvent(props, state, evt)
	if state.SelectedIndex != 2 {
		t.Errorf("ArrowUp should wrap to 2, got %d", state.SelectedIndex)
	}
}

func TestListKeydownJ(t *testing.T) {
	state := List.NewState().(*ListState)
	state.SelectedIndex = 0
	props := map[string]any{"items": testListItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "j"}
	changed := List.OnEvent(props, state, evt)
	if !changed {
		t.Error("j should return true")
	}
	if state.SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex=1, got %d", state.SelectedIndex)
	}
}

func TestListKeydownK(t *testing.T) {
	state := List.NewState().(*ListState)
	state.SelectedIndex = 2
	props := map[string]any{"items": testListItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "k"}
	changed := List.OnEvent(props, state, evt)
	if !changed {
		t.Error("k should return true")
	}
	if state.SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex=1, got %d", state.SelectedIndex)
	}
}

func TestListNotSelectableIgnoresKeys(t *testing.T) {
	state := List.NewState().(*ListState)
	props := map[string]any{
		"items":      testListItems(),
		"selectable": false,
	}

	keys := []string{"ArrowDown", "ArrowUp", "j", "k", "Enter"}
	for _, key := range keys {
		evt := &render.WidgetEvent{Type: "keydown", Key: key}
		changed := List.OnEvent(props, state, evt)
		if changed {
			t.Errorf("non-selectable list should ignore key %q", key)
		}
	}
}

func TestListItemsAsStrings(t *testing.T) {
	props := map[string]any{
		"items": []any{"one", "two", "three"},
	}
	items := readListItems(props)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0] != "one" || items[1] != "two" || items[2] != "three" {
		t.Errorf("items: got %v", items)
	}
}

func TestListItemsAsMaps(t *testing.T) {
	props := map[string]any{
		"items": []any{
			map[string]any{"label": "First"},
			map[string]any{"label": "Second"},
		},
	}
	items := readListItems(props)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0] != "First" || items[1] != "Second" {
		t.Errorf("items: got %v", items)
	}
}

func TestListThemeColors(t *testing.T) {
	state := List.NewState().(*ListState)
	state.SelectedIndex = 0
	props := map[string]any{"items": testListItems()}
	node := List.Render(props, state).(*render.Node)

	// Selected item should use Primary color
	if node.Children[0].Style.Foreground != CurrentTheme.Primary {
		t.Errorf("selected fg: got %q, want %q", node.Children[0].Style.Foreground, CurrentTheme.Primary)
	}
	// Non-selected should use Text color
	if node.Children[1].Style.Foreground != CurrentTheme.Text {
		t.Errorf("non-selected fg: got %q, want %q", node.Children[1].Style.Foreground, CurrentTheme.Text)
	}
}

func TestListParentPointers(t *testing.T) {
	state := List.NewState()
	props := map[string]any{"items": testListItems()}
	node := List.Render(props, state).(*render.Node)
	checkParents(t, node)
}

func TestListHoveredItem(t *testing.T) {
	state := List.NewState().(*ListState)
	state.HoveredIndex = 1
	props := map[string]any{"items": testListItems()}
	node := List.Render(props, state).(*render.Node)

	if node.Children[1].Style.Background != CurrentTheme.Surface1 {
		t.Errorf("hovered item bg: got %q, want %q", node.Children[1].Style.Background, CurrentTheme.Surface1)
	}
	// Non-hovered should have no background
	if node.Children[0].Style.Background != "" {
		t.Errorf("non-hovered item bg: got %q, want empty", node.Children[0].Style.Background)
	}
}

func TestListMouseleaveResetsHover(t *testing.T) {
	state := List.NewState().(*ListState)
	state.HoveredIndex = 1
	props := map[string]any{"items": testListItems()}

	evt := &render.WidgetEvent{Type: "mouseleave"}
	changed := List.OnEvent(props, state, evt)
	if !changed {
		t.Error("mouseleave should return true when hovered")
	}
	if state.HoveredIndex != -1 {
		t.Errorf("expected HoveredIndex=-1, got %d", state.HoveredIndex)
	}
}

func TestListSelectableDefaultTrue(t *testing.T) {
	state := List.NewState()
	props := map[string]any{"items": testListItems()}
	node := List.Render(props, state).(*render.Node)
	if !node.Focusable {
		t.Error("list should be focusable by default (selectable=true)")
	}
}

func TestListSelectableFalseNotFocusable(t *testing.T) {
	state := List.NewState()
	props := map[string]any{"items": testListItems(), "selectable": false}
	node := List.Render(props, state).(*render.Node)
	if node.Focusable {
		t.Error("list with selectable=false should not be focusable")
	}
}

func TestListWidgetDefInterface(t *testing.T) {
	var w interface {
		GetName() string
		GetNewState() any
		DoRender(props map[string]any, state any) any
		DoOnEvent(props map[string]any, state any, event *render.WidgetEvent) bool
	} = List

	if w.GetName() != "List" {
		t.Errorf("GetName() = %q, want 'List'", w.GetName())
	}

	state := w.GetNewState()
	if state == nil {
		t.Fatal("GetNewState() returned nil")
	}

	props := map[string]any{"items": testListItems()}
	result := w.DoRender(props, state)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}
}

func TestReadListItemsNil(t *testing.T) {
	items := readListItems(map[string]any{})
	if items != nil {
		t.Errorf("expected nil for missing items, got %v", items)
	}
}

func TestReadListItemsInvalidType(t *testing.T) {
	items := readListItems(map[string]any{"items": "not a list"})
	if items != nil {
		t.Errorf("expected nil for invalid items type, got %v", items)
	}
}
