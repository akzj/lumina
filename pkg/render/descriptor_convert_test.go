package render

import "testing"

func TestDescriptorFromMap_Text(t *testing.T) {
	m := map[string]any{
		"type":    "text",
		"content": "hello",
	}
	desc := descriptorFromMap(m)
	if desc.Type != "text" {
		t.Errorf("type: got %q, want 'text'", desc.Type)
	}
	if desc.Content != "hello" {
		t.Errorf("content: got %q, want 'hello'", desc.Content)
	}
	if !desc.ContentSet {
		t.Error("ContentSet should be true when content is set")
	}
}

func TestDescriptorFromMap_DefaultType(t *testing.T) {
	m := map[string]any{
		"content": "no type",
	}
	desc := descriptorFromMap(m)
	if desc.Type != "box" {
		t.Errorf("type: got %q, want 'box' (default)", desc.Type)
	}
}

func TestDescriptorFromMap_VBoxWithChildren(t *testing.T) {
	m := map[string]any{
		"type": "vbox",
		"children": []any{
			map[string]any{"type": "text", "content": "child1"},
			map[string]any{"type": "text", "content": "child2"},
		},
	}
	desc := descriptorFromMap(m)
	if desc.Type != "vbox" {
		t.Errorf("type: got %q, want 'vbox'", desc.Type)
	}
	if len(desc.Children) != 2 {
		t.Fatalf("children: got %d, want 2", len(desc.Children))
	}
	if desc.Children[0].Content != "child1" {
		t.Errorf("child[0].content: got %q, want 'child1'", desc.Children[0].Content)
	}
	if desc.Children[1].Content != "child2" {
		t.Errorf("child[1].content: got %q, want 'child2'", desc.Children[1].Content)
	}
}

func TestDescriptorFromMap_WithStyle(t *testing.T) {
	m := map[string]any{
		"type": "box",
		"style": map[string]any{
			"width":      int64(40),
			"height":     int64(10),
			"background": "#ff0000",
			"border":     "rounded",
			"padding":    int64(2),
		},
	}
	desc := descriptorFromMap(m)
	if desc.Style.Width != 40 {
		t.Errorf("width: got %d, want 40", desc.Style.Width)
	}
	if desc.Style.Height != 10 {
		t.Errorf("height: got %d, want 10", desc.Style.Height)
	}
	if desc.Style.Background != "#ff0000" {
		t.Errorf("bg: got %q, want '#ff0000'", desc.Style.Background)
	}
	if desc.Style.Border != "rounded" {
		t.Errorf("border: got %q, want 'rounded'", desc.Style.Border)
	}
	if desc.Style.Padding != 2 {
		t.Errorf("padding: got %d, want 2", desc.Style.Padding)
	}
}

func TestDescriptorFromMap_StyleFloat64(t *testing.T) {
	// Lua numbers come as float64 via JSON/ToAny
	m := map[string]any{
		"type": "box",
		"style": map[string]any{
			"width":  float64(30),
			"height": float64(5),
		},
	}
	desc := descriptorFromMap(m)
	if desc.Style.Width != 30 {
		t.Errorf("width: got %d, want 30", desc.Style.Width)
	}
	if desc.Style.Height != 5 {
		t.Errorf("height: got %d, want 5", desc.Style.Height)
	}
}

func TestDescriptorFromMap_Component(t *testing.T) {
	m := map[string]any{
		"_factoryName": "MyWidget",
		"_props": map[string]any{
			"label": "test",
		},
	}
	desc := descriptorFromMap(m)
	if desc.Type != "component" {
		t.Errorf("type: got %q, want 'component'", desc.Type)
	}
	if desc.ComponentType != "MyWidget" {
		t.Errorf("componentType: got %q, want 'MyWidget'", desc.ComponentType)
	}
	if desc.ComponentProps["label"] != "test" {
		t.Errorf("componentProps[label]: got %v, want 'test'", desc.ComponentProps["label"])
	}
}

func TestDescriptorFromMap_TopLevelVisual(t *testing.T) {
	m := map[string]any{
		"type":       "text",
		"foreground": "#00ff00",
		"background": "#0000ff",
		"bold":       true,
		"dim":        true,
		"underline":  true,
	}
	desc := descriptorFromMap(m)
	if desc.Style.Foreground != "#00ff00" {
		t.Errorf("fg: got %q, want '#00ff00'", desc.Style.Foreground)
	}
	if desc.Style.Background != "#0000ff" {
		t.Errorf("bg: got %q, want '#0000ff'", desc.Style.Background)
	}
	if !desc.Style.Bold {
		t.Error("bold should be true")
	}
	if !desc.Style.Dim {
		t.Error("dim should be true")
	}
	if !desc.Style.Underline {
		t.Error("underline should be true")
	}
}

func TestDescriptorFromMap_Focusable(t *testing.T) {
	m := map[string]any{
		"type":      "box",
		"focusable": true,
		"disabled":  true,
	}
	desc := descriptorFromMap(m)
	if !desc.Focusable {
		t.Error("focusable should be true")
	}
	if !desc.Disabled {
		t.Error("disabled should be true")
	}
}

func TestConvertChildDescriptors_Nil(t *testing.T) {
	props := map[string]any{}
	nodes := convertChildDescriptors(props)
	if nodes != nil {
		t.Errorf("expected nil, got %v", nodes)
	}
}

func TestConvertChildDescriptors_Basic(t *testing.T) {
	props := map[string]any{
		"children": []any{
			map[string]any{"type": "text", "content": "hello"},
			map[string]any{"type": "text", "content": "world"},
		},
	}
	nodes := convertChildDescriptors(props)
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Content != "hello" {
		t.Errorf("node[0].Content: got %q, want 'hello'", nodes[0].Content)
	}
	if nodes[1].Content != "world" {
		t.Errorf("node[1].Content: got %q, want 'world'", nodes[1].Content)
	}
}

func TestConvertChildDescriptors_Nested(t *testing.T) {
	props := map[string]any{
		"children": []any{
			map[string]any{
				"type": "vbox",
				"children": []any{
					map[string]any{"type": "text", "content": "nested"},
				},
			},
		},
	}
	nodes := convertChildDescriptors(props)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Type != "vbox" {
		t.Errorf("node[0].Type: got %q, want 'vbox'", nodes[0].Type)
	}
	if len(nodes[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(nodes[0].Children))
	}
	if nodes[0].Children[0].Content != "nested" {
		t.Errorf("nested content: got %q, want 'nested'", nodes[0].Children[0].Content)
	}
}

func TestStyleFromMap_Margins(t *testing.T) {
	m := map[string]any{
		"margin":       int64(5),
		"marginTop":    int64(1),
		"marginBottom": int64(2),
		"marginLeft":   int64(3),
		"marginRight":  int64(4),
	}
	s := styleFromMap(m)
	if s.Margin != 5 {
		t.Errorf("margin: got %d, want 5", s.Margin)
	}
	if s.MarginTop != 1 {
		t.Errorf("marginTop: got %d, want 1", s.MarginTop)
	}
	if s.MarginBottom != 2 {
		t.Errorf("marginBottom: got %d, want 2", s.MarginBottom)
	}
	if s.MarginLeft != 3 {
		t.Errorf("marginLeft: got %d, want 3", s.MarginLeft)
	}
	if s.MarginRight != 4 {
		t.Errorf("marginRight: got %d, want 4", s.MarginRight)
	}
}

func TestStyleFromMap_Positioning(t *testing.T) {
	m := map[string]any{
		"position": "absolute",
		"top":      int64(10),
		"left":     int64(20),
		"right":    int64(30),
		"bottom":   int64(40),
		"zIndex":   int64(5),
	}
	s := styleFromMap(m)
	if s.Position != "absolute" {
		t.Errorf("position: got %q, want 'absolute'", s.Position)
	}
	if s.Top != 10 {
		t.Errorf("top: got %d, want 10", s.Top)
	}
	if s.Left != 20 {
		t.Errorf("left: got %d, want 20", s.Left)
	}
	if s.Right != 30 {
		t.Errorf("right: got %d, want 30", s.Right)
	}
	if s.Bottom != 40 {
		t.Errorf("bottom: got %d, want 40", s.Bottom)
	}
	if s.ZIndex != 5 {
		t.Errorf("zIndex: got %d, want 5", s.ZIndex)
	}
}

func TestStyleFromMap_DefaultRightBottom(t *testing.T) {
	m := map[string]any{
		"width": int64(10),
	}
	s := styleFromMap(m)
	if s.Right != -1 {
		t.Errorf("right default: got %d, want -1", s.Right)
	}
	if s.Bottom != -1 {
		t.Errorf("bottom default: got %d, want -1", s.Bottom)
	}
}
