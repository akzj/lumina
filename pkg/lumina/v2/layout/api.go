// Package layout provides the flexbox-based layout engine for Lumina v2.
// It computes absolute screen positions for a VNode tree.
package layout

// Style holds layout and visual properties for a VNode.
type Style struct {
	// Sizing
	Width, Height        int // fixed size (0 = auto)
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
	Flex                 int // flex grow factor

	// Spacing
	Padding                    int
	PaddingTop, PaddingBottom  int
	PaddingLeft, PaddingRight  int
	Margin                     int
	MarginTop, MarginBottom    int
	MarginLeft, MarginRight    int
	Gap                        int

	// Alignment
	Justify string // "start", "center", "end", "space-between", "space-around"
	Align   string // "stretch", "start", "center", "end"

	// Visual (used by paint, stored here for convenience)
	Border     string // "none", "single", "double", "rounded"
	Foreground string
	Background string
	Bold       bool
	Dim        bool
	Underline  bool

	// Overflow
	Overflow string // "hidden" (default), "scroll" (future)

	// Positioning
	Position             string // "", "relative", "absolute", "fixed"
	Top, Left            int
	Right, Bottom        int // -1 = unset
	ZIndex               int
}

// VNode is a virtual DOM node — a drawing instruction.
type VNode struct {
	Type     string         // "box", "vbox", "hbox", "text", "input", "textarea", "fragment"
	ID       string         // unique identifier (for events, hit-test)
	Props    map[string]any // arbitrary properties
	Style    Style          // layout + visual style
	Children []*VNode       // child nodes
	Content  string         // for text nodes

	// Layout results (set by ComputeLayout)
	X, Y, W, H int
}

// NewVNode creates a new VNode of the given type with initialized fields.
func NewVNode(nodeType string) *VNode {
	return &VNode{
		Type:  nodeType,
		Props: make(map[string]any),
		Style: Style{
			Justify: "start",
			Align:   "stretch",
			Right:   -1,
			Bottom:  -1,
		},
	}
}

// AddChild appends a child VNode.
func (v *VNode) AddChild(child *VNode) {
	v.Children = append(v.Children, child)
}

// ComputeLayout computes absolute positions for the VNode tree.
// The root is positioned at (x, y) with available size (w, h).
// All descendant positions are set as absolute screen coordinates.
// Padding and Margin shorthand fields are expanded into per-side values first
// (non-zero longhands are left unchanged).
func ComputeLayout(root *VNode, x, y, w, h int) {
	normalizeSpacingInTree(root)
	computeFlexLayout(root, x, y, w, h)
}
