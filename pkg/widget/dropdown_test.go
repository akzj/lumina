package widget

import (
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

// testDropdownItems returns standard dropdown items for testing.
func testDropdownItems() []any {
	return []any{
		map[string]any{"label": "Copy", "icon": "📋"},
		map[string]any{"label": "Paste", "icon": "📄"},
		map[string]any{"divider": true},
		map[string]any{"label": "Delete", "icon": "🗑"},
	}
}

func TestDropdownRenderClosed(t *testing.T) {
	state := Dropdown.NewState()
	props := map[string]any{"label": "Actions", "items": testDropdownItems()}
	result := Dropdown.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Dropdown.Render returned nil or non-*Node")
	}
	if node.Type != "box" {
		t.Errorf("expected root type 'box', got %q", node.Type)
	}
	if !node.Focusable {
		t.Error("expected Focusable=true")
	}
	// Closed: only trigger, no dropdown
	if len(node.Children) != 1 {
		t.Fatalf("closed: expected 1 child (trigger), got %d", len(node.Children))
	}
	trigger := node.Children[0]
	if trigger.Type != "hbox" {
		t.Errorf("trigger type: got %q, want 'hbox'", trigger.Type)
	}
	if len(trigger.Children) != 2 {
		t.Fatalf("trigger: expected 2 children (text + arrow), got %d", len(trigger.Children))
	}
	// Label text
	labelNode := trigger.Children[0]
	if !strings.Contains(labelNode.Content, "Actions") {
		t.Errorf("trigger label: got %q, want containing 'Actions'", labelNode.Content)
	}
	// Arrow ▼ when closed
	arrowNode := trigger.Children[1]
	if !strings.Contains(arrowNode.Content, "▼") {
		t.Errorf("arrow: got %q, want containing '▼'", arrowNode.Content)
	}
}

func TestDropdownRenderDefaultLabel(t *testing.T) {
	state := Dropdown.NewState()
	props := map[string]any{"items": testDropdownItems()}
	node := Dropdown.Render(props, state).(*render.Node)
	trigger := node.Children[0]
	labelNode := trigger.Children[0]
	if !strings.Contains(labelNode.Content, "Menu") {
		t.Errorf("default label: got %q, want containing 'Menu'", labelNode.Content)
	}
}

func TestDropdownRenderOpen(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	state.Highlighted = 0
	props := map[string]any{"label": "Actions", "items": testDropdownItems()}
	node := Dropdown.Render(props, state).(*render.Node)

	// Open: trigger + dropdown
	if len(node.Children) != 2 {
		t.Fatalf("open: expected 2 children, got %d", len(node.Children))
	}

	// Arrow ▲ when open
	trigger := node.Children[0]
	arrowNode := trigger.Children[1]
	if !strings.Contains(arrowNode.Content, "▲") {
		t.Errorf("open arrow: got %q, want containing '▲'", arrowNode.Content)
	}

	// Dropdown
	dropdown := node.Children[1]
	if dropdown.Type != "vbox" {
		t.Errorf("dropdown type: got %q, want 'vbox'", dropdown.Type)
	}
	if dropdown.Style.Position != "absolute" {
		t.Errorf("dropdown position: got %q, want 'absolute'", dropdown.Style.Position)
	}
	if len(dropdown.Children) != 4 {
		t.Fatalf("dropdown: expected 4 items, got %d", len(dropdown.Children))
	}

	// Highlighted item
	highlighted := dropdown.Children[0]
	if highlighted.Style.Background != CurrentTheme.Surface1 {
		t.Errorf("highlighted bg: got %q, want %q", highlighted.Style.Background, CurrentTheme.Surface1)
	}
}

func TestDropdownRenderOpenDivider(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	state.Highlighted = -1
	props := map[string]any{"items": testDropdownItems()}
	node := Dropdown.Render(props, state).(*render.Node)
	dropdown := node.Children[1]
	divider := dropdown.Children[2]
	if !strings.Contains(divider.Content, "─") {
		t.Errorf("divider should contain '─', got %q", divider.Content)
	}
	if !divider.Style.Dim {
		t.Error("divider should have Dim=true")
	}
}

func TestDropdownClickOpens(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	props := map[string]any{"items": testDropdownItems()}

	evt := &render.WidgetEvent{Type: "click"}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("click should return true (opens dropdown)")
	}
	if !state.Open {
		t.Error("expected Open=true after click")
	}
	// Highlighted should be first selectable item (0)
	if state.Highlighted != 0 {
		t.Errorf("expected Highlighted=0, got %d", state.Highlighted)
	}
}

func TestDropdownClickCloses(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	props := map[string]any{"items": testDropdownItems()}

	// Click outside dropdown area → close
	evt := &render.WidgetEvent{Type: "click", X: 5, Y: 0, WidgetX: 0, WidgetY: 0}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("click should return true (closes dropdown)")
	}
	if state.Open {
		t.Error("expected Open=false after click to close")
	}
}

func TestDropdownClickOnItem(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	props := map[string]any{"items": testDropdownItems()}

	// Click on second item (index 1 = Paste)
	// Widget at Y=0, trigger height=3, option 0 at Y=3, option 1 at Y=4
	evt := &render.WidgetEvent{
		Type:    "click",
		X:       5,
		Y:       4,
		WidgetX: 0, WidgetY: 0,
	}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("click on item should return true")
	}
	if state.Open {
		t.Error("expected Open=false after selecting item")
	}
	if evt.FireOnChange != 2 {
		t.Errorf("FireOnChange: got %v, want 2 (1-based)", evt.FireOnChange)
	}
}

func TestDropdownClickOnDisabledItem(t *testing.T) {
	items := []any{
		map[string]any{"label": "OK"},
		map[string]any{"label": "Nope", "disabled": true},
	}
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	props := map[string]any{"items": items}

	// Click on disabled item (index 1)
	evt := &render.WidgetEvent{
		Type:    "click",
		X:       5,
		Y:       4,
		WidgetX: 0, WidgetY: 0,
	}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("click should still return true (closes dropdown)")
	}
	// Should NOT fire onChange for disabled item
	if evt.FireOnChange != nil {
		t.Errorf("FireOnChange should be nil for disabled item, got %v", evt.FireOnChange)
	}
}

func TestDropdownClickOnDivider(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	props := map[string]any{"items": testDropdownItems()}

	// Click on divider (index 2)
	evt := &render.WidgetEvent{
		Type:    "click",
		X:       5,
		Y:       5, // trigger(3) + items 0,1 + divider at index 2
		WidgetX: 0, WidgetY: 0,
	}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("click should return true")
	}
	if evt.FireOnChange != nil {
		t.Errorf("FireOnChange should be nil for divider, got %v", evt.FireOnChange)
	}
}

func TestDropdownKeydownEnterOpens(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	props := map[string]any{"items": testDropdownItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "Enter"}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("Enter should open dropdown")
	}
	if !state.Open {
		t.Error("expected Open=true after Enter")
	}
}

func TestDropdownKeydownSpaceOpens(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	props := map[string]any{"items": testDropdownItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: " "}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("Space should open dropdown")
	}
	if !state.Open {
		t.Error("expected Open=true after Space")
	}
}

func TestDropdownKeydownArrowDown(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	state.Highlighted = 0
	props := map[string]any{"items": testDropdownItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowDown should return true")
	}
	if state.Highlighted != 1 {
		t.Errorf("expected Highlighted=1, got %d", state.Highlighted)
	}
}

func TestDropdownKeydownArrowDownSkipsDivider(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	state.Highlighted = 1 // Paste
	props := map[string]any{"items": testDropdownItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	Dropdown.OnEvent(props, state, evt)
	// Should skip divider (index 2), land on Delete (index 3)
	if state.Highlighted != 3 {
		t.Errorf("expected Highlighted=3 (skip divider), got %d", state.Highlighted)
	}
}

func TestDropdownKeydownArrowUp(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	state.Highlighted = 1
	props := map[string]any{"items": testDropdownItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowUp"}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowUp should return true")
	}
	if state.Highlighted != 0 {
		t.Errorf("expected Highlighted=0, got %d", state.Highlighted)
	}
}

func TestDropdownKeydownEnterSelects(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	state.Highlighted = 1 // Paste
	props := map[string]any{"items": testDropdownItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "Enter"}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("Enter should select item")
	}
	if state.Open {
		t.Error("expected Open=false after Enter select")
	}
	if evt.FireOnChange != 2 {
		t.Errorf("FireOnChange: got %v, want 2 (1-based)", evt.FireOnChange)
	}
}

func TestDropdownKeydownEscapeCloses(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	props := map[string]any{"items": testDropdownItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "Escape"}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("Escape should return true")
	}
	if state.Open {
		t.Error("expected Open=false after Escape")
	}
}

func TestDropdownKeydownEscapeIgnoredWhenClosed(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	props := map[string]any{"items": testDropdownItems()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "Escape"}
	changed := Dropdown.OnEvent(props, state, evt)
	if changed {
		t.Error("Escape when closed should return false")
	}
}

func TestDropdownBlurCloses(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	props := map[string]any{"items": testDropdownItems()}

	evt := &render.WidgetEvent{Type: "blur"}
	changed := Dropdown.OnEvent(props, state, evt)
	if !changed {
		t.Error("blur should return true when open")
	}
	if state.Open {
		t.Error("expected Open=false after blur")
	}
}

func TestDropdownBlurIgnoredWhenClosed(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	props := map[string]any{"items": testDropdownItems()}

	evt := &render.WidgetEvent{Type: "blur"}
	changed := Dropdown.OnEvent(props, state, evt)
	if changed {
		t.Error("blur when closed should return false")
	}
}

func TestDropdownHover(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	props := map[string]any{"items": testDropdownItems()}

	changed := Dropdown.OnEvent(props, state, &render.WidgetEvent{Type: "mouseenter"})
	if !changed {
		t.Error("mouseenter should return true")
	}
	if !state.Hovered {
		t.Error("expected Hovered=true")
	}

	// Duplicate
	changed = Dropdown.OnEvent(props, state, &render.WidgetEvent{Type: "mouseenter"})
	if changed {
		t.Error("duplicate mouseenter should return false")
	}

	changed = Dropdown.OnEvent(props, state, &render.WidgetEvent{Type: "mouseleave"})
	if !changed {
		t.Error("mouseleave should return true")
	}
	if state.Hovered {
		t.Error("expected Hovered=false")
	}
}

func TestDropdownParentPointers(t *testing.T) {
	state := Dropdown.NewState()
	props := map[string]any{"items": testDropdownItems()}
	node := Dropdown.Render(props, state).(*render.Node)
	checkParents(t, node)
}

func TestDropdownParentPointersOpen(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	props := map[string]any{"items": testDropdownItems()}
	node := Dropdown.Render(props, state).(*render.Node)
	checkParents(t, node)
}

func TestDropdownOpenNoItems(t *testing.T) {
	state := Dropdown.NewState().(*DropdownState)
	state.Open = true
	props := map[string]any{"label": "Empty"}
	node := Dropdown.Render(props, state).(*render.Node)
	// No items → only trigger, no dropdown
	if len(node.Children) != 1 {
		t.Errorf("open with no items: expected 1 child, got %d", len(node.Children))
	}
}

func TestDropdownWidgetDefInterface(t *testing.T) {
	var w interface {
		GetName() string
		GetNewState() any
		DoRender(props map[string]any, state any) any
		DoOnEvent(props map[string]any, state any, event *render.WidgetEvent) bool
	} = Dropdown

	if w.GetName() != "Dropdown" {
		t.Errorf("GetName() = %q, want 'Dropdown'", w.GetName())
	}

	state := w.GetNewState()
	if state == nil {
		t.Fatal("GetNewState() returned nil")
	}

	props := map[string]any{"items": testDropdownItems()}
	result := w.DoRender(props, state)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}
}
