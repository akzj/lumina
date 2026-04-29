package widget

import (
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

// testMenuItems returns standard menu items for testing.
func testMenuItems() []any {
	return []any{
		map[string]any{"label": "Files", "icon": "📁"},
		map[string]any{"label": "Settings", "icon": "⚙"},
		map[string]any{"divider": true},
		map[string]any{"label": "Stats", "disabled": true, "icon": "📊"},
		map[string]any{"label": "Help", "icon": "❓"},
	}
}

func TestMenuRenderDefault(t *testing.T) {
	state := Menu.NewState()
	props := map[string]any{"items": testMenuItems()}
	result := Menu.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Menu.Render returned nil or non-*Node")
	}
	if node.Type != "vbox" {
		t.Errorf("expected root type 'vbox', got %q", node.Type)
	}
	if node.Style.Border != "rounded" {
		t.Errorf("expected border 'rounded', got %q", node.Style.Border)
	}
	if !node.Focusable {
		t.Error("expected Focusable=true")
	}
	if len(node.Children) != 5 {
		t.Fatalf("expected 5 children, got %d", len(node.Children))
	}
}

func TestMenuRenderSelected(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = 0
	props := map[string]any{"items": testMenuItems()}
	node := Menu.Render(props, state).(*render.Node)

	selected := node.Children[0]
	if !strings.HasPrefix(selected.Content, "▸ ") {
		t.Errorf("selected item should start with '▸ ', got %q", selected.Content)
	}
	if selected.Style.Foreground != CurrentTheme.Primary {
		t.Errorf("selected fg: got %q, want %q", selected.Style.Foreground, CurrentTheme.Primary)
	}
	if selected.Style.Background != CurrentTheme.Surface0 {
		t.Errorf("selected bg: got %q, want %q", selected.Style.Background, CurrentTheme.Surface0)
	}
}

func TestMenuRenderDivider(t *testing.T) {
	state := Menu.NewState()
	props := map[string]any{"items": testMenuItems()}
	node := Menu.Render(props, state).(*render.Node)

	divider := node.Children[2]
	if !strings.Contains(divider.Content, "─") {
		t.Errorf("divider should contain '─', got %q", divider.Content)
	}
	if !divider.Style.Dim {
		t.Error("divider should have Dim=true")
	}
}

func TestMenuRenderDisabled(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = -1 // nothing selected
	props := map[string]any{"items": testMenuItems()}
	node := Menu.Render(props, state).(*render.Node)

	disabled := node.Children[3] // Stats is disabled
	if disabled.Style.Foreground != CurrentTheme.Muted {
		t.Errorf("disabled fg: got %q, want %q", disabled.Style.Foreground, CurrentTheme.Muted)
	}
}

func TestMenuRenderIcon(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = -1
	props := map[string]any{"items": testMenuItems()}
	node := Menu.Render(props, state).(*render.Node)

	// First item (not selected, has icon "📁")
	first := node.Children[0]
	// When selected (index 0 is default), prefix is "▸ "
	// Set state to something else to test non-selected icon rendering
	state.SelectedIndex = 4
	node = Menu.Render(props, state).(*render.Node)
	first = node.Children[0]
	if !strings.Contains(first.Content, "📁") {
		t.Errorf("expected icon '📁' in content, got %q", first.Content)
	}
	if !strings.Contains(first.Content, "Files") {
		t.Errorf("expected label 'Files' in content, got %q", first.Content)
	}
}

func TestMenuRenderCompact(t *testing.T) {
	state := Menu.NewState()
	props := map[string]any{"items": testMenuItems(), "compact": true}
	node := Menu.Render(props, state).(*render.Node)
	if node.Style.Border != "" {
		t.Errorf("compact menu should have no border, got %q", node.Style.Border)
	}
}

func TestMenuControlledSelected(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = 0
	props := map[string]any{"items": testMenuItems(), "selected": float64(4)}
	node := Menu.Render(props, state).(*render.Node)

	// Item 4 (Help) should be selected
	help := node.Children[4]
	if !strings.HasPrefix(help.Content, "▸ ") {
		t.Errorf("controlled selected item should start with '▸ ', got %q", help.Content)
	}
	// Item 0 (Files) should NOT be selected (overridden by prop)
	files := node.Children[0]
	if strings.HasPrefix(files.Content, "▸ ") {
		t.Error("item 0 should not be selected when controlled prop says 4")
	}
}

func TestMenuKeydownArrowDown(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = 0
	props := map[string]any{"items": testMenuItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	changed := Menu.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowDown should return true")
	}
	if state.SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex=1, got %d", state.SelectedIndex)
	}
}

func TestMenuKeydownArrowDownSkipsDivider(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = 1 // Settings
	props := map[string]any{"items": testMenuItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	changed := Menu.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowDown should return true")
	}
	// Should skip divider (index 2) and disabled (index 3), land on Help (index 4)
	if state.SelectedIndex != 4 {
		t.Errorf("expected SelectedIndex=4 (skip divider+disabled), got %d", state.SelectedIndex)
	}
}

func TestMenuKeydownArrowDownSkipsDisabled(t *testing.T) {
	// Only disabled items after current
	items := []any{
		map[string]any{"label": "A"},
		map[string]any{"label": "B", "disabled": true},
		map[string]any{"label": "C"},
	}
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = 0
	props := map[string]any{"items": items}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	Menu.OnEvent(props, state, evt)
	if state.SelectedIndex != 2 {
		t.Errorf("expected SelectedIndex=2 (skip disabled), got %d", state.SelectedIndex)
	}
}

func TestMenuKeydownArrowUp(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = 4 // Help
	props := map[string]any{"items": testMenuItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowUp"}
	changed := Menu.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowUp should return true")
	}
	// Should skip disabled (index 3) and divider (index 2), land on Settings (index 1)
	if state.SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex=1, got %d", state.SelectedIndex)
	}
}

func TestMenuKeydownArrowUpWraps(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = 0 // Files
	props := map[string]any{"items": testMenuItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowUp"}
	changed := Menu.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowUp should return true")
	}
	// Should wrap to last selectable: Help (index 4)
	if state.SelectedIndex != 4 {
		t.Errorf("expected SelectedIndex=4 (wrap), got %d", state.SelectedIndex)
	}
}

func TestMenuKeydownEnter(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = 1 // Settings
	props := map[string]any{"items": testMenuItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "Enter"}
	changed := Menu.OnEvent(props, state, evt)
	if changed {
		t.Error("Enter should return false (no state change)")
	}
	if evt.FireOnChange != 2 {
		t.Errorf("FireOnChange: got %v, want 2 (1-based)", evt.FireOnChange)
	}
}

func TestMenuKeydownEnterOnDisabled(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = 3 // Stats (disabled)
	props := map[string]any{"items": testMenuItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "Enter"}
	Menu.OnEvent(props, state, evt)
	if evt.FireOnChange != nil {
		t.Errorf("Enter on disabled should not fire onChange, got %v", evt.FireOnChange)
	}
}

func TestMenuMouseleave(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.HoveredIndex = 1
	props := map[string]any{"items": testMenuItems()}

	evt := &render.WidgetEvent{Type: "mouseleave"}
	changed := Menu.OnEvent(props, state, evt)
	if !changed {
		t.Error("mouseleave should return true")
	}
	if state.HoveredIndex != -1 {
		t.Errorf("expected HoveredIndex=-1, got %d", state.HoveredIndex)
	}
}

func TestMenuMouseleaveNoop(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.HoveredIndex = -1
	props := map[string]any{"items": testMenuItems()}

	evt := &render.WidgetEvent{Type: "mouseleave"}
	changed := Menu.OnEvent(props, state, evt)
	if changed {
		t.Error("mouseleave with no hover should return false")
	}
}

func TestMenuParentPointers(t *testing.T) {
	state := Menu.NewState()
	props := map[string]any{"items": testMenuItems()}
	node := Menu.Render(props, state).(*render.Node)
	checkParents(t, node)
}

func TestMenuEmptyItems(t *testing.T) {
	state := Menu.NewState()
	props := map[string]any{}
	result := Menu.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Menu.Render with no items returned nil")
	}
	if len(node.Children) != 0 {
		t.Errorf("expected 0 children for empty items, got %d", len(node.Children))
	}
}

func TestMenuKeydownVimBindings(t *testing.T) {
	state := Menu.NewState().(*MenuState)
	state.SelectedIndex = 0
	items := []any{
		map[string]any{"label": "A"},
		map[string]any{"label": "B"},
		map[string]any{"label": "C"},
	}
	props := map[string]any{"items": items}

	// j = down
	evt := &render.WidgetEvent{Type: "keydown", Key: "j"}
	Menu.OnEvent(props, state, evt)
	if state.SelectedIndex != 1 {
		t.Errorf("j: expected SelectedIndex=1, got %d", state.SelectedIndex)
	}

	// k = up
	evt = &render.WidgetEvent{Type: "keydown", Key: "k"}
	Menu.OnEvent(props, state, evt)
	if state.SelectedIndex != 0 {
		t.Errorf("k: expected SelectedIndex=0, got %d", state.SelectedIndex)
	}
}

func TestReadMenuItems(t *testing.T) {
	props := map[string]any{
		"items": []any{
			map[string]any{"label": "A", "icon": "📁"},
			map[string]any{"label": "B", "disabled": true},
			map[string]any{"divider": true},
		},
	}
	items := readMenuItems(props)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].Label != "A" || items[0].Icon != "📁" {
		t.Errorf("item 0: got %+v", items[0])
	}
	if !items[1].Disabled {
		t.Error("item 1 should be disabled")
	}
	if !items[2].Divider {
		t.Error("item 2 should be divider")
	}
}

func TestReadMenuItemsNil(t *testing.T) {
	items := readMenuItems(map[string]any{})
	if items != nil {
		t.Errorf("expected nil for missing items, got %v", items)
	}
}

func TestFindNextMenuItem(t *testing.T) {
	items := []MenuItem{
		{Label: "A"},
		{Divider: true},
		{Label: "B", Disabled: true},
		{Label: "C"},
	}
	// From 0, forward → should skip divider and disabled, land on 3
	next := findNextMenuItem(items, 0, 1)
	if next != 3 {
		t.Errorf("forward from 0: expected 3, got %d", next)
	}
	// From 3, backward → should skip disabled and divider, land on 0
	next = findNextMenuItem(items, 3, -1)
	if next != 0 {
		t.Errorf("backward from 3: expected 0, got %d", next)
	}
	// From 3, forward → should wrap to 0
	next = findNextMenuItem(items, 3, 1)
	if next != 0 {
		t.Errorf("forward from 3 (wrap): expected 0, got %d", next)
	}
}

func TestMenuWidgetDefInterface(t *testing.T) {
	var w interface {
		GetName() string
		GetNewState() any
		DoRender(props map[string]any, state any) any
		DoOnEvent(props map[string]any, state any, event *render.WidgetEvent) bool
	} = Menu

	if w.GetName() != "Menu" {
		t.Errorf("GetName() = %q, want 'Menu'", w.GetName())
	}

	state := w.GetNewState()
	if state == nil {
		t.Fatal("GetNewState() returned nil")
	}

	props := map[string]any{"items": testMenuItems()}
	result := w.DoRender(props, state)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}
}
