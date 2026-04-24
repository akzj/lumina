package lumina

import (
	"testing"
)

// helper to create a VNode with style props
func makeNode(nodeType string, props map[string]any, children ...*VNode) *VNode {
	n := NewVNode(nodeType)
	for k, v := range props {
		n.Props[k] = v
	}
	for _, c := range children {
		n.AddChild(c)
	}
	return n
}

// helper to create a VNode with a style sub-table
func makeStyledNode(nodeType string, style map[string]any, children ...*VNode) *VNode {
	n := NewVNode(nodeType)
	n.Props["style"] = style
	for _, c := range children {
		n.AddChild(c)
	}
	return n
}

func makeText(content string) *VNode {
	n := NewVNode("text")
	n.Content = content
	return n
}

func makeStyledText(content string, style map[string]any) *VNode {
	n := NewVNode("text")
	n.Content = content
	n.Props["style"] = style
	return n
}

// --- Test: Basic VBox Layout ---

func TestVBoxBasicLayout(t *testing.T) {
	// 3 children in a vbox, 80x24. No flex → each gets 1 row (default).
	root := makeNode("vbox", nil,
		makeText("Line 1"),
		makeText("Line 2"),
		makeText("Line 3"),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	if root.X != 0 || root.Y != 0 || root.W != 80 || root.H != 24 {
		t.Errorf("root: got (%d,%d,%d,%d), want (0,0,80,24)", root.X, root.Y, root.W, root.H)
	}

	// Children should stack vertically, each getting 1 row (default no-flex)
	for i, child := range root.Children {
		if child.X != 0 {
			t.Errorf("child[%d].X = %d, want 0", i, child.X)
		}
		if child.Y != i {
			t.Errorf("child[%d].Y = %d, want %d", i, child.Y, i)
		}
		if child.W != 80 {
			t.Errorf("child[%d].W = %d, want 80", i, child.W)
		}
	}
}

// --- Test: Basic HBox Layout ---

func TestHBoxBasicLayout(t *testing.T) {
	// 3 children in hbox, no flex → each gets 1 col
	root := makeNode("hbox", nil,
		makeText("A"),
		makeText("B"),
		makeText("C"),
	)

	computeFlexLayout(root, 0, 0, 60, 10)

	// Children should be side by side
	for i, child := range root.Children {
		if child.X != i {
			t.Errorf("child[%d].X = %d, want %d", i, child.X, i)
		}
		if child.Y != 0 {
			t.Errorf("child[%d].Y = %d, want 0", i, child.Y)
		}
	}
}

// --- Test: Flex Distribution (int64 values) ---

func TestFlexDistribution(t *testing.T) {
	// vbox with 2 children: flex=1 and flex=2 → 1:2 ratio of 24 rows
	root := makeStyledNode("vbox", nil,
		makeStyledNode("box", map[string]any{"flex": int64(1)}),
		makeStyledNode("box", map[string]any{"flex": int64(2)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]

	// flex=1 gets 24*1/3 = 8, flex=2 gets 24*2/3 = 16
	if c0.H != 8 {
		t.Errorf("flex=1 child height = %d, want 8", c0.H)
	}
	if c1.H != 16 {
		t.Errorf("flex=2 child height = %d, want 16", c1.H)
	}
	if c0.Y != 0 {
		t.Errorf("flex=1 child Y = %d, want 0", c0.Y)
	}
	if c1.Y != 8 {
		t.Errorf("flex=2 child Y = %d, want 8", c1.Y)
	}
}

// --- Test: Fixed + Flex Mix ---

func TestFixedPlusFlex(t *testing.T) {
	// vbox 80x24: child1 has height=3, child2 has flex=1
	root := makeStyledNode("vbox", nil,
		makeStyledNode("box", map[string]any{"height": int64(3)}),
		makeStyledNode("box", map[string]any{"flex": int64(1)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]

	if c0.H != 3 {
		t.Errorf("fixed child height = %d, want 3", c0.H)
	}
	if c1.H != 21 {
		t.Errorf("flex child height = %d, want 21", c1.H)
	}
	if c0.Y != 0 {
		t.Errorf("fixed child Y = %d, want 0", c0.Y)
	}
	if c1.Y != 3 {
		t.Errorf("flex child Y = %d, want 3", c1.Y)
	}
}

// --- Test: Padding ---

func TestPadding(t *testing.T) {
	// vbox with padding=2, one child
	root := makeStyledNode("vbox", map[string]any{"padding": int64(2)},
		makeStyledNode("box", map[string]any{"flex": int64(1)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	child := root.Children[0]

	// Content area: x=2, y=2, w=76, h=20
	if child.X != 2 {
		t.Errorf("child.X = %d, want 2 (padding=2)", child.X)
	}
	if child.Y != 2 {
		t.Errorf("child.Y = %d, want 2 (padding=2)", child.Y)
	}
	if child.W != 76 {
		t.Errorf("child.W = %d, want 76 (80-2*2)", child.W)
	}
	if child.H != 20 {
		t.Errorf("child.H = %d, want 20 (24-2*2)", child.H)
	}
}

// --- Test: Gap ---

func TestGap(t *testing.T) {
	// vbox with gap=1, 3 flex children in 80x24
	root := makeStyledNode("vbox", map[string]any{"gap": int64(1)},
		makeStyledNode("box", map[string]any{"flex": int64(1)}),
		makeStyledNode("box", map[string]any{"flex": int64(1)}),
		makeStyledNode("box", map[string]any{"flex": int64(1)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	// Total gaps = 2 (between 3 children)
	// Available height = 24 - 2 = 22
	// Each child = 22/3 = 7 (integer division)
	c0 := root.Children[0]
	c1 := root.Children[1]
	c2 := root.Children[2]

	if c0.H != 7 {
		t.Errorf("child[0].H = %d, want 7", c0.H)
	}
	if c1.Y != c0.Y+c0.H+1 {
		t.Errorf("child[1].Y = %d, want %d (child[0] end + gap)", c1.Y, c0.Y+c0.H+1)
	}
	if c2.Y != c1.Y+c1.H+1 {
		t.Errorf("child[2].Y = %d, want %d (child[1] end + gap)", c2.Y, c1.Y+c1.H+1)
	}
}

// --- Test: Justify Center ---

func TestJustifyCenter(t *testing.T) {
	// vbox with justify=center, one child height=4 in 80x24
	root := makeStyledNode("vbox", map[string]any{"justify": "center"},
		makeStyledNode("box", map[string]any{"height": int64(4)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	// Extra space = 24 - 4 = 20, centered: offset = 10
	if child.Y != 10 {
		t.Errorf("centered child.Y = %d, want 10", child.Y)
	}
	if child.H != 4 {
		t.Errorf("centered child.H = %d, want 4", child.H)
	}
}

// --- Test: Justify End ---

func TestJustifyEnd(t *testing.T) {
	root := makeStyledNode("vbox", map[string]any{"justify": "end"},
		makeStyledNode("box", map[string]any{"height": int64(4)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	// Extra space = 24 - 4 = 20, end: offset = 20
	if child.Y != 20 {
		t.Errorf("end-justified child.Y = %d, want 20", child.Y)
	}
}

// --- Test: Justify Space-Between ---

func TestJustifySpaceBetween(t *testing.T) {
	root := makeStyledNode("vbox", map[string]any{"justify": "space-between"},
		makeStyledNode("box", map[string]any{"height": int64(2)}),
		makeStyledNode("box", map[string]any{"height": int64(2)}),
		makeStyledNode("box", map[string]any{"height": int64(2)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c2 := root.Children[2]

	// First child at top
	if c0.Y != 0 {
		t.Errorf("space-between child[0].Y = %d, want 0", c0.Y)
	}
	// Last child at bottom: 24 - 2 = 22
	if c2.Y != 22 {
		t.Errorf("space-between child[2].Y = %d, want 22 (24-2)", c2.Y)
	}
}

// --- Test: Align Center (cross axis) ---

func TestAlignCenter(t *testing.T) {
	// vbox with align=center, child has width=20 in 80-wide container
	root := makeStyledNode("vbox", map[string]any{"align": "center"},
		makeStyledNode("box", map[string]any{"width": int64(20), "height": int64(5)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	// Centered: (80 - 20) / 2 = 30
	if child.X != 30 {
		t.Errorf("align-center child.X = %d, want 30", child.X)
	}
	if child.W != 20 {
		t.Errorf("align-center child.W = %d, want 20", child.W)
	}
}

// --- Test: Nested Layout ---

func TestNestedLayout(t *testing.T) {
	// vbox containing hbox containing 2 text nodes
	root := makeStyledNode("vbox", nil,
		makeStyledNode("hbox", map[string]any{"flex": int64(1)},
			makeStyledText("Left", map[string]any{"flex": int64(1)}),
			makeStyledText("Right", map[string]any{"flex": int64(1)}),
		),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	hbox := root.Children[0]
	if hbox.W != 80 {
		t.Errorf("hbox.W = %d, want 80", hbox.W)
	}
	if hbox.H != 24 {
		t.Errorf("hbox.H = %d, want 24", hbox.H)
	}

	left := hbox.Children[0]
	right := hbox.Children[1]

	// Each gets half: 40
	if left.W != 40 {
		t.Errorf("left.W = %d, want 40", left.W)
	}
	if right.W != 40 {
		t.Errorf("right.W = %d, want 40", right.W)
	}
	if left.X != 0 {
		t.Errorf("left.X = %d, want 0", left.X)
	}
	if right.X != 40 {
		t.Errorf("right.X = %d, want 40", right.X)
	}
}

// --- Test: MinWidth/MaxWidth ---

func TestMinMaxWidth(t *testing.T) {
	// hbox: child with minWidth=20, maxWidth=30, flex=1 in 80-wide container
	root := makeStyledNode("hbox", nil,
		makeStyledNode("box", map[string]any{"flex": int64(1), "minWidth": int64(20), "maxWidth": int64(30)}),
		makeStyledNode("box", map[string]any{"flex": int64(3)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]

	// flex=1 out of 4 → 80/4 = 20, which is within [20, 30]
	if c0.W < 20 {
		t.Errorf("child[0].W = %d, want >= 20 (minWidth)", c0.W)
	}
	if c0.W > 30 {
		t.Errorf("child[0].W = %d, want <= 30 (maxWidth)", c0.W)
	}
	// Second child should get remaining space
	if c1.X != c0.X+c0.W {
		t.Errorf("child[1].X = %d, want %d", c1.X, c0.X+c0.W)
	}
}

// --- Test: Border + Padding ---

func TestBorderAndPadding(t *testing.T) {
	// box with border=single and padding=1
	root := makeStyledNode("box", map[string]any{
		"border":  "single",
		"padding": int64(1),
	},
		makeStyledNode("box", map[string]any{"flex": int64(1)}),
	)

	computeFlexLayout(root, 0, 0, 40, 20)

	child := root.Children[0]
	// Border takes 1 on each side, padding takes 1 on each side
	// Content area: x=2, y=2, w=36, h=16
	if child.X != 2 {
		t.Errorf("child.X = %d, want 2 (border+padding)", child.X)
	}
	if child.Y != 2 {
		t.Errorf("child.Y = %d, want 2 (border+padding)", child.Y)
	}
	if child.W != 36 {
		t.Errorf("child.W = %d, want 36 (40-2*2)", child.W)
	}
	if child.H != 16 {
		t.Errorf("child.H = %d, want 16 (20-2*2)", child.H)
	}
}

// --- Test: Text Wrapping ---

func TestTextWrapping(t *testing.T) {
	// Text node with content longer than available width
	root := makeStyledNode("vbox", nil,
		makeStyledText("Hello World! This is a long text that should wrap.", nil),
	)

	computeFlexLayout(root, 0, 0, 20, 10)

	child := root.Children[0]
	// 50 chars in 20-wide → ceil(50/20) = 3 lines
	if child.H != 3 {
		t.Errorf("wrapped text height = %d, want 3", child.H)
	}
}

// --- Test: int64 Fix Verification ---

func TestInt64PropsWork(t *testing.T) {
	// Verify that int64 values from Lua (via ToAny) are handled correctly
	root := makeNode("vbox", nil,
		makeNode("box", map[string]any{"flex": int64(1)}),
		makeNode("box", map[string]any{"flex": int64(2)}),
	)

	computeFlexLayout(root, 0, 0, 80, 30)

	c0 := root.Children[0]
	c1 := root.Children[1]

	// flex=1 → 10, flex=2 → 20
	if c0.H != 10 {
		t.Errorf("int64 flex=1 height = %d, want 10", c0.H)
	}
	if c1.H != 20 {
		t.Errorf("int64 flex=2 height = %d, want 20", c1.H)
	}
}

// --- Test: float64 Props (JSON-sourced) ---

func TestFloat64PropsWork(t *testing.T) {
	// Verify that float64 values (from JSON) are handled correctly
	root := makeNode("vbox", nil,
		makeNode("box", map[string]any{"flex": float64(1)}),
		makeNode("box", map[string]any{"flex": float64(2)}),
	)

	computeFlexLayout(root, 0, 0, 80, 30)

	c0 := root.Children[0]
	c1 := root.Children[1]

	if c0.H != 10 {
		t.Errorf("float64 flex=1 height = %d, want 10", c0.H)
	}
	if c1.H != 20 {
		t.Errorf("float64 flex=2 height = %d, want 20", c1.H)
	}
}

// --- Test: Backward Compatibility (top-level props) ---

func TestBackwardCompatTopLevelProps(t *testing.T) {
	// Old-style: flex, minHeight directly on props (not in style sub-table)
	root := makeNode("vbox", nil,
		makeNode("box", map[string]any{"minHeight": int64(3)}),
		makeNode("box", map[string]any{"flex": int64(1)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]

	if c0.H != 3 {
		t.Errorf("backward compat minHeight child = %d, want 3", c0.H)
	}
	if c1.H != 21 {
		t.Errorf("backward compat flex child = %d, want 21", c1.H)
	}
}

// --- Test: HBox Flex Distribution ---

func TestHBoxFlexDistribution(t *testing.T) {
	root := makeStyledNode("hbox", nil,
		makeStyledNode("box", map[string]any{"flex": int64(1)}),
		makeStyledNode("box", map[string]any{"flex": int64(3)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]

	// flex=1 gets 80*1/4 = 20, flex=3 gets 80*3/4 = 60
	if c0.W != 20 {
		t.Errorf("hbox flex=1 width = %d, want 20", c0.W)
	}
	if c1.W != 60 {
		t.Errorf("hbox flex=3 width = %d, want 60", c1.W)
	}
}

// --- Test: Margin ---

func TestMargin(t *testing.T) {
	root := makeStyledNode("vbox", nil,
		makeStyledNode("box", map[string]any{
			"flex":   int64(1),
			"margin": int64(2),
		}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	// Margin shrinks the node: x=2, y=2, w=76, h=20
	if child.X != 2 {
		t.Errorf("margin child.X = %d, want 2", child.X)
	}
	if child.Y != 2 {
		t.Errorf("margin child.Y = %d, want 2", child.Y)
	}
	if child.W != 76 {
		t.Errorf("margin child.W = %d, want 76", child.W)
	}
	if child.H != 20 {
		t.Errorf("margin child.H = %d, want 20", child.H)
	}
}

// --- Test: parseStyle helpers ---

func TestGetInt(t *testing.T) {
	m := map[string]any{
		"a": int64(42),
		"b": int(10),
		"c": float64(3.7),
		"d": "notanint",
	}

	if got := getInt(m, "a"); got != 42 {
		t.Errorf("getInt int64 = %d, want 42", got)
	}
	if got := getInt(m, "b"); got != 10 {
		t.Errorf("getInt int = %d, want 10", got)
	}
	if got := getInt(m, "c"); got != 3 {
		t.Errorf("getInt float64 = %d, want 3", got)
	}
	if got := getInt(m, "d"); got != 0 {
		t.Errorf("getInt string = %d, want 0", got)
	}
	if got := getInt(m, "missing"); got != 0 {
		t.Errorf("getInt missing = %d, want 0", got)
	}
}

func TestGetString(t *testing.T) {
	m := map[string]any{
		"a": "hello",
		"b": 42,
	}

	if got := getString(m, "a", "def"); got != "hello" {
		t.Errorf("getString = %q, want %q", got, "hello")
	}
	if got := getString(m, "b", "def"); got != "def" {
		t.Errorf("getString non-string = %q, want %q", got, "def")
	}
	if got := getString(m, "missing", "def"); got != "def" {
		t.Errorf("getString missing = %q, want %q", got, "def")
	}
}

func TestGetBool(t *testing.T) {
	m := map[string]any{
		"a": true,
		"b": false,
		"c": "true",
	}

	if got := getBool(m, "a"); !got {
		t.Error("getBool true = false")
	}
	if got := getBool(m, "b"); got {
		t.Error("getBool false = true")
	}
	if got := getBool(m, "c"); got {
		t.Error("getBool string should be false")
	}
	if got := getBool(m, "missing"); got {
		t.Error("getBool missing should be false")
	}
}

// --- Test: Clamp ---

func TestClamp(t *testing.T) {
	if got := clamp(5, 0, 10); got != 5 {
		t.Errorf("clamp(5,0,10) = %d, want 5", got)
	}
	if got := clamp(-1, 0, 10); got != 0 {
		t.Errorf("clamp(-1,0,10) = %d, want 0", got)
	}
	if got := clamp(15, 0, 10); got != 10 {
		t.Errorf("clamp(15,0,10) = %d, want 10", got)
	}
	if got := clamp(100, 5, 0); got != 100 {
		t.Errorf("clamp(100,5,0) = %d, want 100 (no upper bound)", got)
	}
}

// --- Test: VNodeToFrame integration ---

func TestVNodeToFrameIntegration(t *testing.T) {
	// Full integration: create a vbox with styled children, render to frame
	root := makeStyledNode("vbox", map[string]any{"background": "#000000"},
		makeStyledText("Hello", map[string]any{"foreground": "#FFFFFF"}),
		makeStyledNode("box", map[string]any{"flex": int64(1), "background": "#333333"}),
	)

	frame := VNodeToFrame(root, 40, 10)

	if frame.Width != 40 || frame.Height != 10 {
		t.Errorf("frame size = %dx%d, want 40x10", frame.Width, frame.Height)
	}

	// Check that background was rendered
	if frame.Cells[5][5].Background != "#333333" {
		t.Errorf("flex child background = %q, want #333333", frame.Cells[5][5].Background)
	}

	// Check that text was rendered
	if frame.Cells[0][0].Char != 'H' {
		t.Errorf("text cell[0][0] = %c, want 'H'", frame.Cells[0][0].Char)
	}
	if frame.Cells[0][4].Char != 'o' {
		t.Errorf("text cell[0][4] = %c, want 'o'", frame.Cells[0][4].Char)
	}
}

// --- Test: Empty container ---

func TestEmptyContainer(t *testing.T) {
	root := makeStyledNode("vbox", nil)
	computeFlexLayout(root, 0, 0, 80, 24)
	if root.W != 80 || root.H != 24 {
		t.Errorf("empty container size = %dx%d, want 80x24", root.W, root.H)
	}
}

// --- Test: Asymmetric Padding ---

func TestAsymmetricPadding(t *testing.T) {
	root := makeStyledNode("vbox", map[string]any{
		"paddingTop":    int64(1),
		"paddingBottom": int64(3),
		"paddingLeft":   int64(2),
		"paddingRight":  int64(4),
	},
		makeStyledNode("box", map[string]any{"flex": int64(1)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	// Content: x=2, y=1, w=80-2-4=74, h=24-1-3=20
	if child.X != 2 {
		t.Errorf("asymmetric padding child.X = %d, want 2", child.X)
	}
	if child.Y != 1 {
		t.Errorf("asymmetric padding child.Y = %d, want 1", child.Y)
	}
	if child.W != 74 {
		t.Errorf("asymmetric padding child.W = %d, want 74", child.W)
	}
	if child.H != 20 {
		t.Errorf("asymmetric padding child.H = %d, want 20", child.H)
	}
}

// --- Test: Complex nested layout (real-world-like) ---

func TestComplexNestedLayout(t *testing.T) {
	// Simulates a typical TUI: header, content area with sidebar + main, footer
	root := makeStyledNode("vbox", nil,
		// Header: fixed height 1
		makeStyledText("Header", map[string]any{"height": int64(1)}),
		// Content: flex=1, hbox with sidebar + main
		makeStyledNode("hbox", map[string]any{"flex": int64(1)},
			makeStyledNode("box", map[string]any{"width": int64(20)}), // sidebar
			makeStyledNode("box", map[string]any{"flex": int64(1)}),   // main
		),
		// Footer: fixed height 1
		makeStyledText("Footer", map[string]any{"height": int64(1)}),
	)

	computeFlexLayout(root, 0, 0, 80, 24)

	header := root.Children[0]
	content := root.Children[1]
	footer := root.Children[2]

	if header.Y != 0 || header.H != 1 {
		t.Errorf("header: Y=%d H=%d, want Y=0 H=1", header.Y, header.H)
	}
	if content.Y != 1 || content.H != 22 {
		t.Errorf("content: Y=%d H=%d, want Y=1 H=22", content.Y, content.H)
	}
	if footer.Y != 23 || footer.H != 1 {
		t.Errorf("footer: Y=%d H=%d, want Y=23 H=1", footer.Y, footer.H)
	}

	sidebar := content.Children[0]
	main := content.Children[1]

	if sidebar.W != 20 {
		t.Errorf("sidebar.W = %d, want 20", sidebar.W)
	}
	if sidebar.X != 0 {
		t.Errorf("sidebar.X = %d, want 0", sidebar.X)
	}
	if main.X != 20 {
		t.Errorf("main.X = %d, want 20", main.X)
	}
	if main.W != 60 {
		t.Errorf("main.W = %d, want 60", main.W)
	}
}
