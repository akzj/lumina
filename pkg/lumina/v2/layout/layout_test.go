package layout

import "testing"

// --- Helpers ---

func makeNode(nodeType string, style Style, children ...*VNode) *VNode {
	n := NewVNode(nodeType)
	n.Style = style
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

func makeTextStyled(content string, style Style) *VNode {
	n := NewVNode("text")
	n.Content = content
	n.Style = style
	return n
}

func defaultStyle() Style {
	return Style{Justify: "start", Align: "stretch", Right: -1, Bottom: -1}
}

func withFlex(flex int) Style {
	s := defaultStyle()
	s.Flex = flex
	return s
}

func withHeight(h int) Style {
	s := defaultStyle()
	s.Height = h
	return s
}

func withWidth(w int) Style {
	s := defaultStyle()
	s.Width = w
	return s
}

// --- Test: Single Box ---

func TestLayout_SingleBox(t *testing.T) {
	root := makeNode("box", Style{Width: 10, Height: 5, Right: -1, Bottom: -1, Justify: "start", Align: "stretch"})
	ComputeLayout(root, 0, 0, 80, 24)

	if root.X != 0 || root.Y != 0 {
		t.Errorf("root position: got (%d,%d), want (0,0)", root.X, root.Y)
	}
	if root.W != 10 || root.H != 5 {
		t.Errorf("root size: got %dx%d, want 10x5", root.W, root.H)
	}
}

// --- Test: VBox Equal Flex ---

func TestLayout_VBox_EqualFlex(t *testing.T) {
	root := makeNode("vbox", defaultStyle(),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
	)
	ComputeLayout(root, 0, 0, 30, 12)

	for i, child := range root.Children {
		if child.W != 30 {
			t.Errorf("child[%d].W = %d, want 30", i, child.W)
		}
		if child.H != 4 {
			t.Errorf("child[%d].H = %d, want 4", i, child.H)
		}
		if child.Y != i*4 {
			t.Errorf("child[%d].Y = %d, want %d", i, child.Y, i*4)
		}
	}
}

// --- Test: HBox Equal Flex ---

func TestLayout_HBox_EqualFlex(t *testing.T) {
	root := makeNode("hbox", defaultStyle(),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
	)
	ComputeLayout(root, 0, 0, 30, 12)

	for i, child := range root.Children {
		if child.W != 10 {
			t.Errorf("child[%d].W = %d, want 10", i, child.W)
		}
		if child.H != 12 {
			t.Errorf("child[%d].H = %d, want 12", i, child.H)
		}
		if child.X != i*10 {
			t.Errorf("child[%d].X = %d, want %d", i, child.X, i*10)
		}
	}
}

// --- Test: VBox Fixed + Flex ---

func TestLayout_VBox_FixedAndFlex(t *testing.T) {
	root := makeNode("vbox", defaultStyle(),
		makeNode("box", withHeight(2)),
		makeNode("box", withFlex(1)),
	)
	ComputeLayout(root, 0, 0, 10, 10)

	c0 := root.Children[0]
	c1 := root.Children[1]

	if c0.H != 2 {
		t.Errorf("fixed child height = %d, want 2", c0.H)
	}
	if c1.H != 8 {
		t.Errorf("flex child height = %d, want 8", c1.H)
	}
	if c1.Y != 2 {
		t.Errorf("flex child Y = %d, want 2", c1.Y)
	}
}

// --- Test: Text Wrap ---

func TestLayout_Text_Wrap(t *testing.T) {
	// "hello world" = 11 chars, width=5 → ceil(11/5) = 3 lines
	root := makeNode("vbox", defaultStyle(),
		makeText("hello world"),
	)
	ComputeLayout(root, 0, 0, 5, 10)

	child := root.Children[0]
	if child.H != 3 {
		t.Errorf("text height = %d, want 3 (11 chars in width 5)", child.H)
	}
}

// --- Test: Padding ---

func TestLayout_Padding(t *testing.T) {
	s := defaultStyle()
	s.PaddingTop = 1
	s.PaddingBottom = 1
	s.PaddingLeft = 1
	s.PaddingRight = 1

	root := makeNode("vbox", s,
		makeNode("box", withFlex(1)),
	)
	ComputeLayout(root, 0, 0, 20, 10)

	child := root.Children[0]
	if child.X != 1 {
		t.Errorf("child.X = %d, want 1", child.X)
	}
	if child.Y != 1 {
		t.Errorf("child.Y = %d, want 1", child.Y)
	}
	if child.W != 18 {
		t.Errorf("child.W = %d, want 18", child.W)
	}
	if child.H != 8 {
		t.Errorf("child.H = %d, want 8", child.H)
	}
}

// --- Test: Padding / margin shorthand ---

func TestLayout_PaddingShorthand(t *testing.T) {
	s := defaultStyle()
	s.Padding = 1

	root := makeNode("vbox", s,
		makeNode("box", withFlex(1)),
	)
	ComputeLayout(root, 0, 0, 20, 10)

	child := root.Children[0]
	if child.X != 1 || child.Y != 1 || child.W != 18 || child.H != 8 {
		t.Errorf("padding shorthand child layout = (%d,%d) %dx%d, want (1,1) 18x8",
			child.X, child.Y, child.W, child.H)
	}
}

func TestLayout_MarginShorthand(t *testing.T) {
	childStyle := defaultStyle()
	childStyle.Flex = 1
	childStyle.Margin = 2

	root := makeNode("vbox", defaultStyle(),
		makeNode("box", childStyle),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	if child.X != 2 || child.Y != 2 || child.W != 76 || child.H != 20 {
		t.Errorf("margin shorthand child layout = (%d,%d) %dx%d, want (2,2) 76x20",
			child.X, child.Y, child.W, child.H)
	}
}

func TestLayout_PaddingShorthandDoesNotOverrideLonghands(t *testing.T) {
	s := defaultStyle()
	s.Padding = 2
	s.PaddingTop = 1
	s.PaddingLeft = 3

	root := makeNode("vbox", s,
		makeNode("box", withFlex(1)),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	// content: x = 3+1 border? no border. x = PaddingLeft = 3
	if child.X != 3 {
		t.Errorf("child.X = %d, want 3 (explicit paddingLeft)", child.X)
	}
	if child.Y != 1 {
		t.Errorf("child.Y = %d, want 1 (explicit paddingTop)", child.Y)
	}
	// W = 80 - 3 - 2(right from shorthand) = 75
	if child.W != 75 {
		t.Errorf("child.W = %d, want 75", child.W)
	}
	// H = 24 - 1 - 2 = 21
	if child.H != 21 {
		t.Errorf("child.H = %d, want 21", child.H)
	}
}

// --- Test: Gap ---

func TestLayout_Gap(t *testing.T) {
	s := defaultStyle()
	s.Gap = 1

	root := makeNode("vbox", s,
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	// Total gaps = 2, available = 24 - 2 = 22, each child = 22/3 = 7
	c0 := root.Children[0]
	c1 := root.Children[1]
	c2 := root.Children[2]

	if c0.H != 7 {
		t.Errorf("child[0].H = %d, want 7", c0.H)
	}
	if c1.Y != c0.Y+c0.H+1 {
		t.Errorf("child[1].Y = %d, want %d", c1.Y, c0.Y+c0.H+1)
	}
	if c2.Y != c1.Y+c1.H+1 {
		t.Errorf("child[2].Y = %d, want %d", c2.Y, c1.Y+c1.H+1)
	}
}

// --- Test: Absolute Positioning ---

func TestLayout_Absolute(t *testing.T) {
	absStyle := defaultStyle()
	absStyle.Position = "absolute"
	absStyle.Top = 2
	absStyle.Left = 3
	absStyle.Width = 5
	absStyle.Height = 4

	root := makeNode("box", defaultStyle(),
		makeNode("box", absStyle),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	if child.X != 3 {
		t.Errorf("absolute child.X = %d, want 3", child.X)
	}
	if child.Y != 2 {
		t.Errorf("absolute child.Y = %d, want 2", child.Y)
	}
	if child.W != 5 {
		t.Errorf("absolute child.W = %d, want 5", child.W)
	}
	if child.H != 4 {
		t.Errorf("absolute child.H = %d, want 4", child.H)
	}
}

// --- Test: Nested Layout ---

func TestLayout_Nested(t *testing.T) {
	root := makeNode("vbox", defaultStyle(),
		makeNode("hbox", withFlex(1),
			makeNode("box", withFlex(1)),
			makeNode("box", withFlex(1)),
		),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	hbox := root.Children[0]
	if hbox.W != 80 || hbox.H != 24 {
		t.Errorf("hbox size: %dx%d, want 80x24", hbox.W, hbox.H)
	}

	left := hbox.Children[0]
	right := hbox.Children[1]

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

// --- Test: Component Rect (offset) ---

func TestLayout_ComponentRect(t *testing.T) {
	root := makeNode("vbox", defaultStyle(),
		makeNode("box", withFlex(1)),
	)
	ComputeLayout(root, 10, 5, 30, 12)

	if root.X != 10 || root.Y != 5 {
		t.Errorf("root position: got (%d,%d), want (10,5)", root.X, root.Y)
	}
	if root.W != 30 || root.H != 12 {
		t.Errorf("root size: got %dx%d, want 30x12", root.W, root.H)
	}

	child := root.Children[0]
	if child.X != 10 {
		t.Errorf("child.X = %d, want 10", child.X)
	}
	if child.Y != 5 {
		t.Errorf("child.Y = %d, want 5", child.Y)
	}
}

// --- Test: Flex Distribution Ratio ---

func TestLayout_FlexDistribution(t *testing.T) {
	root := makeNode("vbox", defaultStyle(),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(2)),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]

	// flex=1 gets 24*1/3 = 8, flex=2 gets 24*2/3 = 16
	if c0.H != 8 {
		t.Errorf("flex=1 child height = %d, want 8", c0.H)
	}
	if c1.H != 16 {
		t.Errorf("flex=2 child height = %d, want 16", c1.H)
	}
}

// --- Test: HBox Flex Distribution ---

func TestLayout_HBox_FlexDistribution(t *testing.T) {
	root := makeNode("hbox", defaultStyle(),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(3)),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]

	if c0.W != 20 {
		t.Errorf("hbox flex=1 width = %d, want 20", c0.W)
	}
	if c1.W != 60 {
		t.Errorf("hbox flex=3 width = %d, want 60", c1.W)
	}
}

// --- Test: Border + Padding ---

func TestLayout_BorderAndPadding(t *testing.T) {
	s := defaultStyle()
	s.Border = "single"
	s.PaddingTop = 1
	s.PaddingBottom = 1
	s.PaddingLeft = 1
	s.PaddingRight = 1

	root := makeNode("box", s,
		makeNode("box", withFlex(1)),
	)
	ComputeLayout(root, 0, 0, 40, 20)

	child := root.Children[0]
	// Border=1 on each side + padding=1 on each side = 2 each
	if child.X != 2 {
		t.Errorf("child.X = %d, want 2", child.X)
	}
	if child.Y != 2 {
		t.Errorf("child.Y = %d, want 2", child.Y)
	}
	if child.W != 36 {
		t.Errorf("child.W = %d, want 36", child.W)
	}
	if child.H != 16 {
		t.Errorf("child.H = %d, want 16", child.H)
	}
}

// --- Test: Justify Center ---

func TestLayout_JustifyCenter(t *testing.T) {
	s := defaultStyle()
	s.Justify = "center"

	root := makeNode("vbox", s,
		makeNode("box", withHeight(4)),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	// Extra space = 24 - 4 = 20, centered: offset = 10
	if child.Y != 10 {
		t.Errorf("centered child.Y = %d, want 10", child.Y)
	}
}

// --- Test: Justify End ---

func TestLayout_JustifyEnd(t *testing.T) {
	s := defaultStyle()
	s.Justify = "end"

	root := makeNode("vbox", s,
		makeNode("box", withHeight(4)),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	if child.Y != 20 {
		t.Errorf("end-justified child.Y = %d, want 20", child.Y)
	}
}

// --- Test: Justify Space-Between ---

func TestLayout_JustifySpaceBetween(t *testing.T) {
	s := defaultStyle()
	s.Justify = "space-between"

	root := makeNode("vbox", s,
		makeNode("box", withHeight(2)),
		makeNode("box", withHeight(2)),
		makeNode("box", withHeight(2)),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c2 := root.Children[2]

	if c0.Y != 0 {
		t.Errorf("space-between child[0].Y = %d, want 0", c0.Y)
	}
	// Last child at bottom: Y = 24 - 2 = 22
	if c2.Y != 22 {
		t.Errorf("space-between child[2].Y = %d, want 22", c2.Y)
	}
}

// --- Test: Align Center ---

func TestLayout_AlignCenter(t *testing.T) {
	s := defaultStyle()
	s.Align = "center"

	childStyle := defaultStyle()
	childStyle.Width = 20
	childStyle.Height = 5

	root := makeNode("vbox", s,
		makeNode("box", childStyle),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	// Centered: (80 - 20) / 2 = 30
	if child.X != 30 {
		t.Errorf("align-center child.X = %d, want 30", child.X)
	}
	if child.W != 20 {
		t.Errorf("align-center child.W = %d, want 20", child.W)
	}
}

// --- Test: Margin ---

func TestLayout_Margin(t *testing.T) {
	childStyle := defaultStyle()
	childStyle.Flex = 1
	childStyle.MarginTop = 2
	childStyle.MarginBottom = 2
	childStyle.MarginLeft = 2
	childStyle.MarginRight = 2

	root := makeNode("vbox", defaultStyle(),
		makeNode("box", childStyle),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
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

// --- Test: MinWidth/MaxWidth ---

func TestLayout_MinMaxWidth(t *testing.T) {
	c0Style := defaultStyle()
	c0Style.Flex = 1
	c0Style.MinWidth = 20
	c0Style.MaxWidth = 30

	root := makeNode("hbox", defaultStyle(),
		makeNode("box", c0Style),
		makeNode("box", withFlex(3)),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	if c0.W < 20 {
		t.Errorf("child[0].W = %d, want >= 20 (minWidth)", c0.W)
	}
	if c0.W > 30 {
		t.Errorf("child[0].W = %d, want <= 30 (maxWidth)", c0.W)
	}
}

// --- Test: Container Stretches Without Flex (vbox) ---

func TestLayout_VBox_ContainerStretchesWithoutFlex(t *testing.T) {
	// A container child without explicit height or flex should get implicit flex=1
	root := makeNode("vbox", defaultStyle(),
		makeNode("vbox", defaultStyle(), // no flex, no height → implicit flex=1
			makeText("Hello"),
		),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	if child.H != 24 {
		t.Errorf("container without flex should stretch, got H=%d, want 24", child.H)
	}
}

// --- Test: Container Stretches Without Flex (hbox) ---

func TestLayout_HBox_ContainerStretchesWithoutFlex(t *testing.T) {
	sidebarStyle := defaultStyle()
	sidebarStyle.Width = 20

	root := makeNode("hbox", defaultStyle(),
		makeNode("vbox", sidebarStyle),
		makeNode("vbox", defaultStyle()), // no flex, no width → implicit flex=1
	)
	ComputeLayout(root, 0, 0, 120, 40)

	sidebar := root.Children[0]
	content := root.Children[1]

	if sidebar.W != 20 {
		t.Errorf("sidebar.W = %d, want 20", sidebar.W)
	}
	if content.W != 100 {
		t.Errorf("content.W = %d, want 100", content.W)
	}
}

// --- Test: Text Gets Minimum Height ---

func TestLayout_VBox_TextGetsMinimumHeight(t *testing.T) {
	root := makeNode("vbox", defaultStyle(),
		makeText("Line 1"),
		makeText("Line 2"),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	if root.Children[0].H != 1 {
		t.Errorf("text1 H = %d, want 1", root.Children[0].H)
	}
	if root.Children[1].H != 1 {
		t.Errorf("text2 H = %d, want 1", root.Children[1].H)
	}
}

// --- Test: Complex Nested (header + content + footer) ---

func TestLayout_ComplexNested(t *testing.T) {
	headerStyle := defaultStyle()
	headerStyle.Height = 1

	footerStyle := defaultStyle()
	footerStyle.Height = 1

	sidebarStyle := defaultStyle()
	sidebarStyle.Width = 20

	root := makeNode("vbox", defaultStyle(),
		makeTextStyled("Header", headerStyle),
		makeNode("hbox", withFlex(1),
			makeNode("box", sidebarStyle),
			makeNode("box", withFlex(1)),
		),
		makeTextStyled("Footer", footerStyle),
	)
	ComputeLayout(root, 0, 0, 80, 24)

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
	if main.X != 20 {
		t.Errorf("main.X = %d, want 20", main.X)
	}
	if main.W != 60 {
		t.Errorf("main.W = %d, want 60", main.W)
	}
}

// --- Test: Relative Positioning ---

func TestLayout_Relative(t *testing.T) {
	childStyle := defaultStyle()
	childStyle.Position = "relative"
	childStyle.Top = 3
	childStyle.Left = 5
	childStyle.Flex = 1

	root := makeNode("vbox", defaultStyle(),
		makeNode("box", childStyle),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	// Normal flow position is (0,0), then offset by (5,3)
	if child.X != 5 {
		t.Errorf("relative child.X = %d, want 5", child.X)
	}
	if child.Y != 3 {
		t.Errorf("relative child.Y = %d, want 3", child.Y)
	}
}

// --- Test: Clamp helper ---

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

// --- Test: StringWidth ---

func TestStringWidth(t *testing.T) {
	if got := stringWidth("hello"); got != 5 {
		t.Errorf("stringWidth(\"hello\") = %d, want 5", got)
	}
	// CJK character = 2 width
	if got := stringWidth("你好"); got != 4 {
		t.Errorf("stringWidth(\"你好\") = %d, want 4", got)
	}
	if got := stringWidth(""); got != 0 {
		t.Errorf("stringWidth(\"\") = %d, want 0", got)
	}
}

// --- Test: Empty Container ---

func TestLayout_EmptyContainer(t *testing.T) {
	root := makeNode("vbox", defaultStyle())
	ComputeLayout(root, 0, 0, 80, 24)
	if root.W != 80 || root.H != 24 {
		t.Errorf("empty container size = %dx%d, want 80x24", root.W, root.H)
	}
}

// --- Test: Asymmetric Padding ---

func TestLayout_AsymmetricPadding(t *testing.T) {
	s := defaultStyle()
	s.PaddingTop = 1
	s.PaddingBottom = 3
	s.PaddingLeft = 2
	s.PaddingRight = 4

	root := makeNode("vbox", s,
		makeNode("box", withFlex(1)),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	if child.X != 2 {
		t.Errorf("child.X = %d, want 2", child.X)
	}
	if child.Y != 1 {
		t.Errorf("child.Y = %d, want 1", child.Y)
	}
	if child.W != 74 {
		t.Errorf("child.W = %d, want 74", child.W)
	}
	if child.H != 20 {
		t.Errorf("child.H = %d, want 20", child.H)
	}
}

// --- Test: HBox with text natural width ---

func TestLayout_HBox_TextNaturalWidth(t *testing.T) {
	root := makeNode("hbox", defaultStyle(),
		makeText("A"),
		makeText("B"),
		makeText("C"),
	)
	ComputeLayout(root, 0, 0, 60, 10)

	// Each text "A", "B", "C" has natural width 1
	for i, child := range root.Children {
		if child.X != i {
			t.Errorf("child[%d].X = %d, want %d", i, child.X, i)
		}
	}
}

// --- Test: Fixed position ---

func TestLayout_Fixed(t *testing.T) {
	fixedStyle := defaultStyle()
	fixedStyle.Position = "fixed"
	fixedStyle.Top = 5
	fixedStyle.Left = 10
	fixedStyle.Width = 20
	fixedStyle.Height = 8

	root := makeNode("box", defaultStyle(),
		makeNode("box", fixedStyle),
	)
	ComputeLayout(root, 0, 0, 80, 24)

	child := root.Children[0]
	if child.X != 10 {
		t.Errorf("fixed child.X = %d, want 10", child.X)
	}
	if child.Y != 5 {
		t.Errorf("fixed child.Y = %d, want 5", child.Y)
	}
	if child.W != 20 {
		t.Errorf("fixed child.W = %d, want 20", child.W)
	}
	if child.H != 8 {
		t.Errorf("fixed child.H = %d, want 8", child.H)
	}
}
