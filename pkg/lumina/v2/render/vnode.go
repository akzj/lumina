package render

import "encoding/json"

// VNode is the JSON-serializable representation of a Node tree.
// Used for testing and debugging — contains both semantic and layout info.
type VNode struct {
	Type    string         `json:"type"`
	ID      string         `json:"id,omitempty"`
	Key     string         `json:"key,omitempty"`
	Content string         `json:"content,omitempty"`
	Style   map[string]any `json:"style,omitempty"`

	// Layout (absolute positions, same as internal engine coordinates)
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`

	// Scroll state (only for overflow=scroll containers)
	ScrollY      int `json:"scrollY,omitempty"`
	ScrollHeight int `json:"scrollHeight,omitempty"`

	// Children
	Children []*VNode `json:"children,omitempty"`
}

// NodeToVNode converts a render.Node tree to a VNode tree.
func NodeToVNode(node *Node) *VNode {
	if node == nil {
		return nil
	}
	vn := &VNode{
		Type:    node.Type,
		ID:      node.ID,
		Key:     node.Key,
		Content: node.Content,
		X:       node.X,
		Y:       node.Y,
		W:       node.W,
		H:       node.H,
	}

	// Scroll state
	if node.Style.Overflow == "scroll" {
		vn.ScrollY = node.ScrollY
		vn.ScrollHeight = node.ScrollHeight
	}

	// Style: only output non-zero values to reduce noise
	vn.Style = styleToMap(node.Style)

	// Children (skip component wrappers — inline their children)
	if len(node.Children) > 0 {
		vn.Children = make([]*VNode, len(node.Children))
		for i, child := range node.Children {
			vn.Children[i] = NodeToVNode(child)
		}
	}

	return vn
}

// styleToMap converts a Style struct to a map with only non-zero values.
func styleToMap(s Style) map[string]any {
	m := make(map[string]any)

	if s.Width > 0 {
		m["width"] = s.Width
	}
	if s.Height > 0 {
		m["height"] = s.Height
	}
	if s.MinWidth > 0 {
		m["minWidth"] = s.MinWidth
	}
	if s.MaxWidth > 0 {
		m["maxWidth"] = s.MaxWidth
	}
	if s.MinHeight > 0 {
		m["minHeight"] = s.MinHeight
	}
	if s.MaxHeight > 0 {
		m["maxHeight"] = s.MaxHeight
	}
	if s.Flex > 0 {
		m["flex"] = s.Flex
	}
	if s.PaddingTop > 0 {
		m["paddingTop"] = s.PaddingTop
	}
	if s.PaddingBottom > 0 {
		m["paddingBottom"] = s.PaddingBottom
	}
	if s.PaddingLeft > 0 {
		m["paddingLeft"] = s.PaddingLeft
	}
	if s.PaddingRight > 0 {
		m["paddingRight"] = s.PaddingRight
	}
	if s.MarginTop > 0 {
		m["marginTop"] = s.MarginTop
	}
	if s.MarginBottom > 0 {
		m["marginBottom"] = s.MarginBottom
	}
	if s.MarginLeft > 0 {
		m["marginLeft"] = s.MarginLeft
	}
	if s.MarginRight > 0 {
		m["marginRight"] = s.MarginRight
	}
	if s.Gap > 0 {
		m["gap"] = s.Gap
	}
	if s.Justify != "" {
		m["justify"] = s.Justify
	}
	if s.Align != "" {
		m["align"] = s.Align
	}
	if s.Border != "" && s.Border != "none" {
		m["border"] = s.Border
	}
	if s.Foreground != "" {
		m["foreground"] = s.Foreground
	}
	if s.Background != "" {
		m["background"] = s.Background
	}
	if s.Bold {
		m["bold"] = true
	}
	if s.Dim {
		m["dim"] = true
	}
	if s.Underline {
		m["underline"] = true
	}
	if s.Overflow != "" {
		m["overflow"] = s.Overflow
	}
	if s.Position != "" {
		m["position"] = s.Position
	}
	if s.Top != 0 {
		m["top"] = s.Top
	}
	if s.Left != 0 {
		m["left"] = s.Left
	}
	if s.Right >= 0 {
		m["right"] = s.Right
	}
	if s.Bottom >= 0 {
		m["bottom"] = s.Bottom
	}
	if s.ZIndex != 0 {
		m["zIndex"] = s.ZIndex
	}

	if len(m) == 0 {
		return nil
	}
	return m
}

// VNodeJSON returns the JSON representation of a VNode tree.
func VNodeJSON(node *Node) ([]byte, error) {
	vn := NodeToVNode(node)
	return json.MarshalIndent(vn, "", "  ")
}
