package widget

import (
	"testing"

	"github.com/akzj/lumina/pkg/render"
)

// testOptions returns a standard set of options for testing.
func testOptions() []any {
	return []any{
		map[string]any{"label": "Apple", "value": "apple"},
		map[string]any{"label": "Banana", "value": "banana"},
		map[string]any{"label": "Cherry", "value": "cherry"},
	}
}

func TestSelectRenderDefault(t *testing.T) {
	state := Select.NewState()
	props := map[string]any{"options": testOptions()}
	result := Select.Render(props, state)
	node, ok := result.(*render.Node)
	if !ok || node == nil {
		t.Fatal("Select.Render returned nil or non-*Node")
	}
	if node.Type != "box" {
		t.Errorf("expected root type 'box', got %q", node.Type)
	}
	if node.Style.Border != "rounded" {
		t.Errorf("expected border 'rounded', got %q", node.Style.Border)
	}
	if !node.Focusable {
		t.Error("expected Focusable=true")
	}
	if node.Disabled {
		t.Error("expected Disabled=false")
	}
	// Should have 1 child (trigger hbox), no dropdown
	if len(node.Children) != 1 {
		t.Fatalf("closed select: expected 1 child (trigger), got %d", len(node.Children))
	}
	trigger := node.Children[0]
	if trigger.Type != "hbox" {
		t.Errorf("trigger type: got %q, want 'hbox'", trigger.Type)
	}
	if len(trigger.Children) != 2 {
		t.Fatalf("trigger: expected 2 children (text + arrow), got %d", len(trigger.Children))
	}
	// Default placeholder
	textNode := trigger.Children[0]
	if textNode.Content != "Select..." {
		t.Errorf("placeholder: got %q, want 'Select...'", textNode.Content)
	}
	// Arrow should be ▼ when closed
	arrowNode := trigger.Children[1]
	if arrowNode.Content != "▼" {
		t.Errorf("arrow: got %q, want '▼'", arrowNode.Content)
	}
}

func TestSelectRenderCustomPlaceholder(t *testing.T) {
	state := Select.NewState()
	props := map[string]any{"options": testOptions(), "placeholder": "Pick fruit..."}
	node := Select.Render(props, state).(*render.Node)
	trigger := node.Children[0]
	textNode := trigger.Children[0]
	if textNode.Content != "Pick fruit..." {
		t.Errorf("custom placeholder: got %q, want 'Pick fruit...'", textNode.Content)
	}
}

func TestSelectRenderDisabled(t *testing.T) {
	state := Select.NewState()
	props := map[string]any{"options": testOptions(), "disabled": true}
	node := Select.Render(props, state).(*render.Node)
	if node.Focusable {
		t.Error("disabled select should not be focusable")
	}
	if !node.Disabled {
		t.Error("disabled select should have Disabled=true")
	}
}

func TestSelectRenderWithSelectedValue(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Selected = 1 // "Banana"
	props := map[string]any{"options": testOptions()}
	node := Select.Render(props, state).(*render.Node)
	trigger := node.Children[0]
	textNode := trigger.Children[0]
	if textNode.Content != "Banana" {
		t.Errorf("selected display: got %q, want 'Banana'", textNode.Content)
	}
}

func TestSelectControlledMode(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Selected = 0 // internal: Apple
	// Controlled: value prop overrides
	props := map[string]any{"options": testOptions(), "value": "cherry"}
	node := Select.Render(props, state).(*render.Node)
	trigger := node.Children[0]
	textNode := trigger.Children[0]
	if textNode.Content != "Cherry" {
		t.Errorf("controlled display: got %q, want 'Cherry'", textNode.Content)
	}
}

func TestSelectRenderOpen(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	state.Highlighted = 1
	props := map[string]any{"options": testOptions()}
	node := Select.Render(props, state).(*render.Node)

	// Should have 2 children: trigger + dropdown
	if len(node.Children) != 2 {
		t.Fatalf("open select: expected 2 children, got %d", len(node.Children))
	}

	// Arrow should be ▲ when open
	trigger := node.Children[0]
	arrowNode := trigger.Children[1]
	if arrowNode.Content != "▲" {
		t.Errorf("open arrow: got %q, want '▲'", arrowNode.Content)
	}

	// Dropdown
	dropdown := node.Children[1]
	if dropdown.Type != "vbox" {
		t.Errorf("dropdown type: got %q, want 'vbox'", dropdown.Type)
	}
	if dropdown.Style.Position != "absolute" {
		t.Errorf("dropdown position: got %q, want 'absolute'", dropdown.Style.Position)
	}
	if len(dropdown.Children) != 3 {
		t.Fatalf("dropdown: expected 3 option nodes, got %d", len(dropdown.Children))
	}

	// Highlighted option (index 1) should have background
	highlighted := dropdown.Children[1]
	if highlighted.Style.Background != "#45475A" {
		t.Errorf("highlighted bg: got %q, want '#45475A'", highlighted.Style.Background)
	}

	// Non-highlighted options should have no background
	if dropdown.Children[0].Style.Background != "" {
		t.Errorf("non-highlighted bg should be empty, got %q", dropdown.Children[0].Style.Background)
	}
}

func TestSelectRenderSelectedCheckmark(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	state.Selected = 2 // Cherry
	state.Highlighted = 0
	props := map[string]any{"options": testOptions()}
	node := Select.Render(props, state).(*render.Node)
	dropdown := node.Children[1]
	selectedOpt := dropdown.Children[2]
	if selectedOpt.Content != "Cherry ✓" {
		t.Errorf("selected option content: got %q, want 'Cherry ✓'", selectedOpt.Content)
	}
	if selectedOpt.Style.Foreground != "#89B4FA" {
		t.Errorf("selected option fg: got %q, want '#89B4FA'", selectedOpt.Style.Foreground)
	}
}

func TestSelectClickOpens(t *testing.T) {
	state := Select.NewState().(*SelectState)
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "click"}
	changed := Select.OnEvent(props, state, evt)
	if !changed {
		t.Error("click should return true (opens dropdown)")
	}
	if !state.Open {
		t.Error("expected Open=true after click")
	}
	if state.Highlighted != 0 {
		t.Errorf("expected Highlighted=0, got %d", state.Highlighted)
	}
}

func TestSelectClickOpensWithPreviousSelection(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Selected = 2
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "click"}
	Select.OnEvent(props, state, evt)
	if state.Highlighted != 2 {
		t.Errorf("expected Highlighted=2 (matches Selected), got %d", state.Highlighted)
	}
}

func TestSelectClickClosesWhenOpen(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	props := map[string]any{"options": testOptions()}

	// Click outside dropdown area (on trigger) → close
	evt := &render.WidgetEvent{Type: "click", X: 5, Y: 0, WidgetX: 0, WidgetY: 0}
	changed := Select.OnEvent(props, state, evt)
	if !changed {
		t.Error("click should return true (closes dropdown)")
	}
	if state.Open {
		t.Error("expected Open=false after click to close")
	}
}

func TestSelectClickOnOption(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	props := map[string]any{"options": testOptions()}

	// Click on second option (index 1)
	// Widget at Y=0, trigger height=3 (border+content+border), dropdown border=1
	// Option 0 at Y=4, Option 1 at Y=5, Option 2 at Y=6
	evt := &render.WidgetEvent{
		Type:    "click",
		X:       5,
		Y:       5,
		WidgetX: 0, WidgetY: 0,
		WidgetW: 20, WidgetH: 3,
	}
	changed := Select.OnEvent(props, state, evt)
	if !changed {
		t.Error("click on option should return true")
	}
	if state.Open {
		t.Error("expected Open=false after selecting option")
	}
	if state.Selected != 1 {
		t.Errorf("expected Selected=1, got %d", state.Selected)
	}
	if evt.FireOnChange != "banana" {
		t.Errorf("FireOnChange: got %v, want 'banana'", evt.FireOnChange)
	}
}

func TestSelectKeydownSpaceOpens(t *testing.T) {
	state := Select.NewState().(*SelectState)
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "keydown", Key: " "}
	changed := Select.OnEvent(props, state, evt)
	if !changed {
		t.Error("Space should open dropdown")
	}
	if !state.Open {
		t.Error("expected Open=true after Space")
	}
}

func TestSelectKeydownEnterOpens(t *testing.T) {
	state := Select.NewState().(*SelectState)
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "Enter"}
	changed := Select.OnEvent(props, state, evt)
	if !changed {
		t.Error("Enter should open dropdown")
	}
	if !state.Open {
		t.Error("expected Open=true after Enter")
	}
}

func TestSelectKeydownArrowDown(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	state.Highlighted = 0
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	changed := Select.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowDown should return true")
	}
	if state.Highlighted != 1 {
		t.Errorf("expected Highlighted=1, got %d", state.Highlighted)
	}
}

func TestSelectKeydownArrowDownWraps(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	state.Highlighted = 2 // last option
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowDown"}
	Select.OnEvent(props, state, evt)
	if state.Highlighted != 0 {
		t.Errorf("ArrowDown should wrap to 0, got %d", state.Highlighted)
	}
}

func TestSelectKeydownArrowUp(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	state.Highlighted = 1
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowUp"}
	changed := Select.OnEvent(props, state, evt)
	if !changed {
		t.Error("ArrowUp should return true")
	}
	if state.Highlighted != 0 {
		t.Errorf("expected Highlighted=0, got %d", state.Highlighted)
	}
}

func TestSelectKeydownArrowUpWraps(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	state.Highlighted = 0
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "ArrowUp"}
	Select.OnEvent(props, state, evt)
	if state.Highlighted != 2 {
		t.Errorf("ArrowUp should wrap to 2, got %d", state.Highlighted)
	}
}

func TestSelectKeydownEnterSelectsOption(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	state.Highlighted = 1
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "Enter"}
	changed := Select.OnEvent(props, state, evt)
	if !changed {
		t.Error("Enter should select option")
	}
	if state.Open {
		t.Error("expected Open=false after Enter select")
	}
	if state.Selected != 1 {
		t.Errorf("expected Selected=1, got %d", state.Selected)
	}
	if evt.FireOnChange != "banana" {
		t.Errorf("FireOnChange: got %v, want 'banana'", evt.FireOnChange)
	}
}

func TestSelectKeydownEscapeCloses(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "Escape"}
	changed := Select.OnEvent(props, state, evt)
	if !changed {
		t.Error("Escape should return true")
	}
	if state.Open {
		t.Error("expected Open=false after Escape")
	}
}

func TestSelectKeydownEscapeIgnoredWhenClosed(t *testing.T) {
	state := Select.NewState().(*SelectState)
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "keydown", Key: "Escape"}
	changed := Select.OnEvent(props, state, evt)
	if changed {
		t.Error("Escape when closed should return false")
	}
}

func TestSelectDisabledIgnoresEvents(t *testing.T) {
	state := Select.NewState().(*SelectState)
	props := map[string]any{"options": testOptions(), "disabled": true}

	events := []string{"click", "mouseenter", "mouseleave", "keydown"}
	for _, evtType := range events {
		evt := &render.WidgetEvent{Type: evtType, Key: " "}
		changed := Select.OnEvent(props, state, evt)
		if changed {
			t.Errorf("disabled select should ignore %q", evtType)
		}
	}
	if state.Open {
		t.Error("disabled select should not open")
	}
}

func TestSelectHover(t *testing.T) {
	state := Select.NewState().(*SelectState)
	props := map[string]any{"options": testOptions()}

	changed := Select.OnEvent(props, state, &render.WidgetEvent{Type: "mouseenter"})
	if !changed {
		t.Error("mouseenter should return true")
	}
	if !state.Hovered {
		t.Error("expected Hovered=true")
	}

	// Duplicate mouseenter
	changed = Select.OnEvent(props, state, &render.WidgetEvent{Type: "mouseenter"})
	if changed {
		t.Error("duplicate mouseenter should return false")
	}

	changed = Select.OnEvent(props, state, &render.WidgetEvent{Type: "mouseleave"})
	if !changed {
		t.Error("mouseleave should return true")
	}
	if state.Hovered {
		t.Error("expected Hovered=false")
	}
}

func TestSelectBlurClosesDropdown(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "blur"}
	changed := Select.OnEvent(props, state, evt)
	if !changed {
		t.Error("blur should return true when open")
	}
	if state.Open {
		t.Error("expected Open=false after blur")
	}
}

func TestSelectBlurIgnoredWhenClosed(t *testing.T) {
	state := Select.NewState().(*SelectState)
	props := map[string]any{"options": testOptions()}

	evt := &render.WidgetEvent{Type: "blur"}
	changed := Select.OnEvent(props, state, evt)
	if changed {
		t.Error("blur when closed should return false")
	}
}

func TestSelectParentPointers(t *testing.T) {
	state := Select.NewState()
	props := map[string]any{"options": testOptions()}
	node := Select.Render(props, state).(*render.Node)
	checkParents(t, node)
}

func TestSelectParentPointersOpen(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	props := map[string]any{"options": testOptions()}
	node := Select.Render(props, state).(*render.Node)
	checkParents(t, node)
}

func checkParents(t *testing.T, node *render.Node) {
	t.Helper()
	for i, ch := range node.Children {
		if ch.Parent != node {
			t.Errorf("child %d (%s) Parent not set correctly", i, ch.Type)
		}
		checkParents(t, ch)
	}
}

func TestSelectNoOptions(t *testing.T) {
	state := Select.NewState()
	props := map[string]any{}
	node := Select.Render(props, state).(*render.Node)
	if node == nil {
		t.Fatal("Select.Render with no options returned nil")
	}
	// Should still render trigger with placeholder
	trigger := node.Children[0]
	textNode := trigger.Children[0]
	if textNode.Content != "Select..." {
		t.Errorf("no options placeholder: got %q, want 'Select...'", textNode.Content)
	}
}

func TestSelectOpenNoOptions(t *testing.T) {
	state := Select.NewState().(*SelectState)
	state.Open = true
	props := map[string]any{}
	node := Select.Render(props, state).(*render.Node)
	// With no options, should only have trigger (no dropdown)
	if len(node.Children) != 1 {
		t.Errorf("open with no options: expected 1 child, got %d", len(node.Children))
	}
}

func TestReadOptions(t *testing.T) {
	props := map[string]any{
		"options": []any{
			map[string]any{"label": "A", "value": "a"},
			map[string]any{"label": "B", "value": "b"},
		},
	}
	opts := readOptions(props)
	if len(opts) != 2 {
		t.Fatalf("expected 2 options, got %d", len(opts))
	}
	if opts[0].Label != "A" || opts[0].Value != "a" {
		t.Errorf("option 0: got %+v", opts[0])
	}
	if opts[1].Label != "B" || opts[1].Value != "b" {
		t.Errorf("option 1: got %+v", opts[1])
	}
}

func TestReadOptionsNil(t *testing.T) {
	opts := readOptions(map[string]any{})
	if opts != nil {
		t.Errorf("expected nil for missing options, got %v", opts)
	}
}

func TestReadOptionsInvalidType(t *testing.T) {
	opts := readOptions(map[string]any{"options": "not a list"})
	if opts != nil {
		t.Errorf("expected nil for invalid options type, got %v", opts)
	}
}

func TestFindSelectedIndex(t *testing.T) {
	opts := []SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
		{Label: "C", Value: "c"},
	}
	if idx := findSelectedIndex(opts, "b"); idx != 1 {
		t.Errorf("expected 1, got %d", idx)
	}
	if idx := findSelectedIndex(opts, "z"); idx != -1 {
		t.Errorf("expected -1 for missing, got %d", idx)
	}
}

func TestSelectWidgetDefInterface(t *testing.T) {
	var w interface {
		GetName() string
		GetNewState() any
		DoRender(props map[string]any, state any) any
		DoOnEvent(props map[string]any, state any, event *render.WidgetEvent) bool
	} = Select

	if w.GetName() != "Select" {
		t.Errorf("GetName() = %q, want 'Select'", w.GetName())
	}

	state := w.GetNewState()
	if state == nil {
		t.Fatal("GetNewState() returned nil")
	}

	props := map[string]any{"options": testOptions()}
	result := w.DoRender(props, state)
	if result == nil {
		t.Fatal("DoRender returned nil")
	}

	changed := w.DoOnEvent(props, state, &render.WidgetEvent{Type: "click"})
	if !changed {
		t.Error("DoOnEvent click should return true (opens dropdown)")
	}
}

func TestAllWidgetsIncludesLabelAndSelect(t *testing.T) {
	all := All()
	names := make(map[string]bool)
	for _, w := range all {
		names[w.Name] = true
	}
	if !names["Label"] {
		t.Error("All() should include Label")
	}
	if !names["Select"] {
		t.Error("All() should include Select")
	}
	if len(all) != 6 {
		t.Errorf("expected 6 widgets, got %d", len(all))
	}
}
