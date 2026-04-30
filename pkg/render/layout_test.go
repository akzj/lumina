package render

import "testing"

// --- Helpers ---

func makeNode(nodeType string, s Style, children ...*Node) *Node {
	n := NewNode(nodeType)
	n.Style = s
	for _, c := range children {
		n.AddChild(c)
	}
	return n
}

func makeText(content string) *Node {
	n := NewNode("text")
	n.Content = content
	return n
}

func makeTextStyled(content string, s Style) *Node {
	n := NewNode("text")
	n.Content = content
	n.Style = s
	return n
}

func ds() Style {
	return Style{Justify: "start", Align: "stretch", Right: -1, Bottom: -1}
}

func withFlex(flex int) Style {
	s := ds()
	s.Flex = flex
	return s
}

func withHeight(h int) Style {
	s := ds()
	s.Height = h
	return s
}

func withWidth(w int) Style {
	s := ds()
	s.Width = w
	return s
}

// --- Test: LayoutFull single box ---

func TestLayout_SingleBox(t *testing.T) {
	root := makeNode("box", Style{Width: 10, Height: 5, Right: -1, Bottom: -1, Justify: "start", Align: "stretch"})
	LayoutFull(root, 0, 0, 80, 24)

	if root.X != 0 || root.Y != 0 {
		t.Errorf("root position: got (%d,%d), want (0,0)", root.X, root.Y)
	}
	if root.W != 10 || root.H != 5 {
		t.Errorf("root size: got %dx%d, want 10x5", root.W, root.H)
	}
}

// --- Test: SimpleVBox — 3 children with fixed height ---

func TestLayout_SimpleVBox(t *testing.T) {
	root := makeNode("vbox", ds(),
		makeNode("box", withHeight(3)),
		makeNode("box", withHeight(4)),
		makeNode("box", withHeight(5)),
	)
	LayoutFull(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]
	c2 := root.Children[2]

	if c0.Y != 0 || c0.H != 3 {
		t.Errorf("child[0]: Y=%d H=%d, want Y=0 H=3", c0.Y, c0.H)
	}
	if c1.Y != 3 || c1.H != 4 {
		t.Errorf("child[1]: Y=%d H=%d, want Y=3 H=4", c1.Y, c1.H)
	}
	if c2.Y != 7 || c2.H != 5 {
		t.Errorf("child[2]: Y=%d H=%d, want Y=7 H=5", c2.Y, c2.H)
	}
}

// --- Test: SimpleHBox — 3 children with fixed width ---

func TestLayout_SimpleHBox(t *testing.T) {
	root := makeNode("hbox", ds(),
		makeNode("box", withWidth(10)),
		makeNode("box", withWidth(20)),
		makeNode("box", withWidth(15)),
	)
	LayoutFull(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]
	c2 := root.Children[2]

	if c0.X != 0 || c0.W != 10 {
		t.Errorf("child[0]: X=%d W=%d, want X=0 W=10", c0.X, c0.W)
	}
	if c1.X != 10 || c1.W != 20 {
		t.Errorf("child[1]: X=%d W=%d, want X=10 W=20", c1.X, c1.W)
	}
	if c2.X != 30 || c2.W != 15 {
		t.Errorf("child[2]: X=%d W=%d, want X=30 W=15", c2.X, c2.W)
	}
}

// --- Test: Flex grow factors ---

func TestLayout_Flex(t *testing.T) {
	root := makeNode("vbox", ds(),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(2)),
	)
	LayoutFull(root, 0, 0, 80, 24)

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

// --- Test: VBox equal flex ---

func TestLayout_VBox_EqualFlex(t *testing.T) {
	root := makeNode("vbox", ds(),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
	)
	LayoutFull(root, 0, 0, 30, 12)

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

// --- Test: HBox equal flex ---

func TestLayout_HBox_EqualFlex(t *testing.T) {
	root := makeNode("hbox", ds(),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
	)
	LayoutFull(root, 0, 0, 30, 12)

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
	root := makeNode("vbox", ds(),
		makeNode("box", withHeight(2)),
		makeNode("box", withFlex(1)),
	)
	LayoutFull(root, 0, 0, 10, 10)

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

// --- Test: Padding ---

func TestLayout_Padding(t *testing.T) {
	s := ds()
	s.PaddingTop = 1
	s.PaddingBottom = 1
	s.PaddingLeft = 1
	s.PaddingRight = 1

	root := makeNode("vbox", s,
		makeNode("box", withFlex(1)),
	)
	LayoutFull(root, 0, 0, 20, 10)

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

// --- Test: Padding shorthand ---

func TestLayout_PaddingShorthand(t *testing.T) {
	s := ds()
	s.Padding = 1

	root := makeNode("vbox", s,
		makeNode("box", withFlex(1)),
	)
	LayoutFull(root, 0, 0, 20, 10)

	child := root.Children[0]
	if child.X != 1 || child.Y != 1 || child.W != 18 || child.H != 8 {
		t.Errorf("padding shorthand child layout = (%d,%d) %dx%d, want (1,1) 18x8",
			child.X, child.Y, child.W, child.H)
	}
}

// --- Test: Border ---

func TestLayout_Border(t *testing.T) {
	s := ds()
	s.Border = "single"

	root := makeNode("box", s,
		makeNode("box", withFlex(1)),
	)
	LayoutFull(root, 0, 0, 40, 20)

	child := root.Children[0]
	// Border=1 on each side
	if child.X != 1 {
		t.Errorf("child.X = %d, want 1", child.X)
	}
	if child.Y != 1 {
		t.Errorf("child.Y = %d, want 1", child.Y)
	}
	if child.W != 38 {
		t.Errorf("child.W = %d, want 38", child.W)
	}
	if child.H != 18 {
		t.Errorf("child.H = %d, want 18", child.H)
	}
}

// --- Test: Border + Padding ---

func TestLayout_BorderAndPadding(t *testing.T) {
	s := ds()
	s.Border = "single"
	s.PaddingTop = 1
	s.PaddingBottom = 1
	s.PaddingLeft = 1
	s.PaddingRight = 1

	root := makeNode("box", s,
		makeNode("box", withFlex(1)),
	)
	LayoutFull(root, 0, 0, 40, 20)

	child := root.Children[0]
	// Border=1 + Padding=1 on each side = 2 each
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

// --- Test: Gap ---

func TestLayout_Gap(t *testing.T) {
	s := ds()
	s.Gap = 1

	root := makeNode("vbox", s,
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
	)
	LayoutFull(root, 0, 0, 80, 24)

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

// --- Test: Margin ---

func TestLayout_Margin(t *testing.T) {
	childStyle := ds()
	childStyle.Flex = 1
	childStyle.MarginTop = 2
	childStyle.MarginBottom = 2
	childStyle.MarginLeft = 2
	childStyle.MarginRight = 2

	root := makeNode("vbox", ds(),
		makeNode("box", childStyle),
	)
	LayoutFull(root, 0, 0, 80, 24)

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

// --- Test: Margin shorthand ---

func TestLayout_MarginShorthand(t *testing.T) {
	childStyle := ds()
	childStyle.Flex = 1
	childStyle.Margin = 2

	root := makeNode("vbox", ds(),
		makeNode("box", childStyle),
	)
	LayoutFull(root, 0, 0, 80, 24)

	child := root.Children[0]
	if child.X != 2 || child.Y != 2 || child.W != 76 || child.H != 20 {
		t.Errorf("margin shorthand child layout = (%d,%d) %dx%d, want (2,2) 76x20",
			child.X, child.Y, child.W, child.H)
	}
}

// --- Test: Justify Center ---

func TestLayout_Justify_Center(t *testing.T) {
	s := ds()
	s.Justify = "center"

	root := makeNode("vbox", s,
		makeNode("box", withHeight(4)),
	)
	LayoutFull(root, 0, 0, 80, 24)

	child := root.Children[0]
	// Extra space = 24 - 4 = 20, centered: offset = 10
	if child.Y != 10 {
		t.Errorf("centered child.Y = %d, want 10", child.Y)
	}
}

// --- Test: Justify End ---

func TestLayout_Justify_End(t *testing.T) {
	s := ds()
	s.Justify = "end"

	root := makeNode("vbox", s,
		makeNode("box", withHeight(4)),
	)
	LayoutFull(root, 0, 0, 80, 24)

	child := root.Children[0]
	if child.Y != 20 {
		t.Errorf("end-justified child.Y = %d, want 20", child.Y)
	}
}

// --- Test: Justify Space-Between ---

func TestLayout_Justify_SpaceBetween(t *testing.T) {
	s := ds()
	s.Justify = "space-between"

	root := makeNode("vbox", s,
		makeNode("box", withHeight(2)),
		makeNode("box", withHeight(2)),
		makeNode("box", withHeight(2)),
	)
	LayoutFull(root, 0, 0, 80, 24)

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

func TestLayout_Align_Center(t *testing.T) {
	s := ds()
	s.Align = "center"

	childStyle := ds()
	childStyle.Width = 20
	childStyle.Height = 5

	root := makeNode("vbox", s,
		makeNode("box", childStyle),
	)
	LayoutFull(root, 0, 0, 80, 24)

	child := root.Children[0]
	// Centered: (80 - 20) / 2 = 30
	if child.X != 30 {
		t.Errorf("align-center child.X = %d, want 30", child.X)
	}
	if child.W != 20 {
		t.Errorf("align-center child.W = %d, want 20", child.W)
	}
}

// --- Test: Nested layout ---

func TestLayout_Nested(t *testing.T) {
	headerStyle := ds()
	headerStyle.Height = 1

	footerStyle := ds()
	footerStyle.Height = 1

	sidebarStyle := ds()
	sidebarStyle.Width = 20

	root := makeNode("vbox", ds(),
		makeTextStyled("Header", headerStyle),
		makeNode("hbox", withFlex(1),
			makeNode("box", sidebarStyle),
			makeNode("box", withFlex(1)),
		),
		makeTextStyled("Footer", footerStyle),
	)
	LayoutFull(root, 0, 0, 80, 24)

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

// --- Test: Text node ---

func TestLayout_Text(t *testing.T) {
	root := makeNode("vbox", ds(),
		makeText("hello world"),
	)
	LayoutFull(root, 0, 0, 5, 10)

	child := root.Children[0]
	// "hello world" = 11 chars, width=5 → ceil(11/5) = 3 lines
	if child.H != 3 {
		t.Errorf("text height = %d, want 3 (11 chars in width 5)", child.H)
	}
}

// --- Test: Text in hbox uses natural width ---

func TestLayout_HBox_TextNaturalWidth(t *testing.T) {
	root := makeNode("hbox", ds(),
		makeText("A"),
		makeText("B"),
		makeText("C"),
	)
	LayoutFull(root, 0, 0, 60, 10)

	for i, child := range root.Children {
		if child.X != i {
			t.Errorf("child[%d].X = %d, want %d", i, child.X, i)
		}
	}
}

// --- Test: Absolute positioning ---

func TestLayout_AbsolutePosition(t *testing.T) {
	absStyle := ds()
	absStyle.Position = "absolute"
	absStyle.Top = 2
	absStyle.Left = 3
	absStyle.Width = 5
	absStyle.Height = 4

	root := makeNode("box", ds(),
		makeNode("box", absStyle),
	)
	LayoutFull(root, 0, 0, 80, 24)

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

// --- Test: Fixed positioning ---

func TestLayout_FixedPosition(t *testing.T) {
	fixedStyle := ds()
	fixedStyle.Position = "fixed"
	fixedStyle.Top = 5
	fixedStyle.Left = 10
	fixedStyle.Width = 20
	fixedStyle.Height = 8

	root := makeNode("box", ds(),
		makeNode("box", fixedStyle),
	)
	LayoutFull(root, 0, 0, 80, 24)

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

// --- Test: MinMax size ---

func TestLayout_MinMaxSize(t *testing.T) {
	c0Style := ds()
	c0Style.Flex = 1
	c0Style.MinWidth = 20
	c0Style.MaxWidth = 30

	root := makeNode("hbox", ds(),
		makeNode("box", c0Style),
		makeNode("box", withFlex(3)),
	)
	LayoutFull(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	if c0.W < 20 {
		t.Errorf("child[0].W = %d, want >= 20 (minWidth)", c0.W)
	}
	if c0.W > 30 {
		t.Errorf("child[0].W = %d, want <= 30 (maxWidth)", c0.W)
	}
}

// --- Test: Relative positioning ---

func TestLayout_Relative(t *testing.T) {
	childStyle := ds()
	childStyle.Position = "relative"
	childStyle.Top = 3
	childStyle.Left = 5
	childStyle.Flex = 1

	root := makeNode("vbox", ds(),
		makeNode("box", childStyle),
	)
	LayoutFull(root, 0, 0, 80, 24)

	child := root.Children[0]
	if child.X != 5 {
		t.Errorf("relative child.X = %d, want 5", child.X)
	}
	if child.Y != 3 {
		t.Errorf("relative child.Y = %d, want 3", child.Y)
	}
}

// --- Test: Offset positioning ---

func TestLayout_ComponentRect(t *testing.T) {
	root := makeNode("vbox", ds(),
		makeNode("box", withFlex(1)),
	)
	LayoutFull(root, 10, 5, 30, 12)

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

// --- Test: HBox flex distribution ---

func TestLayout_HBox_FlexDistribution(t *testing.T) {
	root := makeNode("hbox", ds(),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(3)),
	)
	LayoutFull(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]

	if c0.W != 20 {
		t.Errorf("hbox flex=1 width = %d, want 20", c0.W)
	}
	if c1.W != 60 {
		t.Errorf("hbox flex=3 width = %d, want 60", c1.W)
	}
}

// --- Test: Empty container ---

func TestLayout_EmptyContainer(t *testing.T) {
	root := makeNode("vbox", ds())
	LayoutFull(root, 0, 0, 80, 24)
	if root.W != 80 || root.H != 24 {
		t.Errorf("empty container size = %dx%d, want 80x24", root.W, root.H)
	}
}

// --- Test: Nil root ---

func TestLayout_NilRoot(t *testing.T) {
	// Should not panic
	LayoutFull(nil, 0, 0, 80, 24)
	LayoutIncremental(nil)
}

// --- Test: Container stretches without flex (vbox) ---

func TestLayout_VBox_ContainerStretchesWithoutFlex(t *testing.T) {
	root := makeNode("vbox", ds(),
		makeNode("vbox", ds(),
			makeText("Hello"),
		),
	)
	LayoutFull(root, 0, 0, 80, 24)

	child := root.Children[0]
	if child.H != 24 {
		t.Errorf("container without flex should stretch, got H=%d, want 24", child.H)
	}
}

// --- Test: Container stretches without flex (hbox) ---

func TestLayout_HBox_ContainerStretchesWithoutFlex(t *testing.T) {
	sidebarStyle := ds()
	sidebarStyle.Width = 20

	root := makeNode("hbox", ds(),
		makeNode("vbox", sidebarStyle),
		makeNode("vbox", ds()),
	)
	LayoutFull(root, 0, 0, 120, 40)

	sidebar := root.Children[0]
	content := root.Children[1]

	if sidebar.W != 20 {
		t.Errorf("sidebar.W = %d, want 20", sidebar.W)
	}
	if content.W != 100 {
		t.Errorf("content.W = %d, want 100", content.W)
	}
}

// --- Test: LayoutFull clears LayoutDirty ---

func TestLayout_FullClearsLayoutDirty(t *testing.T) {
	root := makeNode("vbox", ds(),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
	)
	root.LayoutDirty = true
	root.Children[0].LayoutDirty = true
	root.Children[1].LayoutDirty = true

	LayoutFull(root, 0, 0, 80, 24)

	if root.LayoutDirty {
		t.Error("root.LayoutDirty should be false after LayoutFull")
	}
	if root.Children[0].LayoutDirty {
		t.Error("child[0].LayoutDirty should be false after LayoutFull")
	}
	if root.Children[1].LayoutDirty {
		t.Error("child[1].LayoutDirty should be false after LayoutFull")
	}
}

// --- Test: Incremental — no dirty → positions unchanged ---

func TestLayout_Incremental_NoDirty(t *testing.T) {
	root := makeNode("vbox", ds(),
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(1)),
	)

	// First, do a full layout
	LayoutFull(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]

	// Save positions
	c0X, c0Y, c0W, c0H := c0.X, c0.Y, c0.W, c0.H
	c1X, c1Y, c1W, c1H := c1.X, c1.Y, c1.W, c1.H

	// Clear PaintDirty
	c0.PaintDirty = false
	c1.PaintDirty = false

	// Run incremental with nothing dirty
	LayoutIncremental(root)

	// Positions should be unchanged
	if c0.X != c0X || c0.Y != c0Y || c0.W != c0W || c0.H != c0H {
		t.Errorf("child[0] position changed after no-dirty incremental")
	}
	if c1.X != c1X || c1.Y != c1Y || c1.W != c1W || c1.H != c1H {
		t.Errorf("child[1] position changed after no-dirty incremental")
	}
	// PaintDirty should NOT have been set
	if c0.PaintDirty {
		t.Error("child[0].PaintDirty should be false after no-dirty incremental")
	}
	if c1.PaintDirty {
		t.Error("child[1].PaintDirty should be false after no-dirty incremental")
	}
}

// --- Test: Incremental — only dirty subtree recomputed ---

func TestLayout_Incremental_SubtreeOnly(t *testing.T) {
	root := makeNode("vbox", ds(),
		makeNode("box", withFlex(1)),
		makeNode("vbox", withFlex(1),
			makeNode("box", withFlex(1)),
			makeNode("box", withFlex(1)),
		),
	)

	// Full layout
	LayoutFull(root, 0, 0, 80, 24)

	// Save all positions
	c0 := root.Children[0]
	c1 := root.Children[1]
	c1c0 := c1.Children[0]
	c1c1 := c1.Children[1]

	c0X, c0Y := c0.X, c0.Y
	c1c0X, c1c0Y, c1c0W, c1c0H := c1c0.X, c1c0.Y, c1c0.W, c1c0.H
	c1c1X, c1c1Y, c1c1W, c1c1H := c1c1.X, c1c1.Y, c1c1.W, c1c1.H

	// Clear all dirty
	c0.PaintDirty = false
	c1.PaintDirty = false
	c1c0.PaintDirty = false
	c1c1.PaintDirty = false

	// Mark only c1 as layout dirty
	c1.LayoutDirty = true

	// Run incremental
	LayoutIncremental(root)

	// c0 should NOT have been touched (not dirty, not a descendant of dirty)
	if c0.X != c0X || c0.Y != c0Y {
		t.Error("child[0] position changed — should not have been recomputed")
	}
	if c0.PaintDirty {
		t.Error("child[0].PaintDirty should be false — not recomputed")
	}

	// c1's children should have been recomputed (same values since container didn't change)
	if c1c0.X != c1c0X || c1c0.Y != c1c0Y || c1c0.W != c1c0W || c1c0.H != c1c0H {
		t.Error("c1.child[0] position changed unexpectedly")
	}
	if c1c1.X != c1c1X || c1c1.Y != c1c1Y || c1c1.W != c1c1W || c1c1.H != c1c1H {
		t.Error("c1.child[1] position changed unexpectedly")
	}

	// LayoutDirty should be cleared
	if c1.LayoutDirty {
		t.Error("c1.LayoutDirty should be false after incremental")
	}
}

// --- Test: Incremental — position change marks PaintDirty ---

func TestLayout_Incremental_MarksPaintDirty(t *testing.T) {
	root := makeNode("vbox", ds(),
		makeNode("box", withHeight(5)),
		makeNode("box", withFlex(1)),
	)

	// Full layout
	LayoutFull(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]

	// Verify initial layout
	if c0.H != 5 {
		t.Fatalf("c0.H = %d, want 5", c0.H)
	}
	if c1.Y != 5 {
		t.Fatalf("c1.Y = %d, want 5", c1.Y)
	}

	// Clear PaintDirty
	c0.PaintDirty = false
	c1.PaintDirty = false
	root.PaintDirty = false

	// Change c0's height → mark parent LayoutDirty
	c0.Style.Height = 10
	root.LayoutDirty = true

	// Run incremental
	LayoutIncremental(root)

	// c0 should now be H=10
	if c0.H != 10 {
		t.Errorf("c0.H = %d, want 10 after style change", c0.H)
	}

	// c1 should have moved down
	if c1.Y != 10 {
		t.Errorf("c1.Y = %d, want 10 after c0 grew", c1.Y)
	}

	// c1 should be PaintDirty because its position changed
	if !c1.PaintDirty {
		t.Error("c1.PaintDirty should be true — position changed")
	}
}

// --- Test: Incremental produces same result as full ---

func TestLayout_Incremental_MatchesFull(t *testing.T) {
	// Build a complex tree
	buildTree := func() *Node {
		return makeNode("vbox", ds(),
			makeTextStyled("Header", withHeight(1)),
			makeNode("hbox", withFlex(1),
				makeNode("vbox", withWidth(20),
					makeText("Sidebar"),
				),
				makeNode("vbox", withFlex(1),
					makeNode("box", withFlex(1)),
					makeNode("box", withFlex(2)),
				),
			),
			makeTextStyled("Footer", withHeight(1)),
		)
	}

	// Full layout
	full := buildTree()
	LayoutFull(full, 0, 0, 80, 24)

	// Incremental: do full first, then mark everything dirty and run incremental
	incr := buildTree()
	LayoutFull(incr, 0, 0, 80, 24)

	// Mark all dirty
	markAllDirty(incr)
	LayoutIncremental(incr)

	// Compare all positions
	compareNodes(t, "root", full, incr)
}

func markAllDirty(node *Node) {
	node.LayoutDirty = true
	for _, child := range node.Children {
		markAllDirty(child)
	}
}

func compareNodes(t *testing.T, path string, a, b *Node) {
	t.Helper()
	if a.X != b.X || a.Y != b.Y || a.W != b.W || a.H != b.H {
		t.Errorf("%s: full=(%d,%d,%d,%d) incr=(%d,%d,%d,%d)",
			path, a.X, a.Y, a.W, a.H, b.X, b.Y, b.W, b.H)
	}
	if len(a.Children) != len(b.Children) {
		t.Errorf("%s: children count mismatch: full=%d incr=%d",
			path, len(a.Children), len(b.Children))
		return
	}
	for i := range a.Children {
		compareNodes(t, path+"."+a.Children[i].Type, a.Children[i], b.Children[i])
	}
}

// --- Test: StringWidth ---

func TestLayout_StringWidth(t *testing.T) {
	if got := stringWidth("hello"); got != 5 {
		t.Errorf("stringWidth(\"hello\") = %d, want 5", got)
	}
	if got := stringWidth("你好"); got != 4 {
		t.Errorf("stringWidth(\"你好\") = %d, want 4", got)
	}
	if got := stringWidth(""); got != 0 {
		t.Errorf("stringWidth(\"\") = %d, want 0", got)
	}
}

// --- Test: Clamp ---

func TestLayout_Clamp(t *testing.T) {
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

// --- Test: Asymmetric Padding ---

func TestLayout_AsymmetricPadding(t *testing.T) {
	s := ds()
	s.PaddingTop = 1
	s.PaddingBottom = 3
	s.PaddingLeft = 2
	s.PaddingRight = 4

	root := makeNode("vbox", s,
		makeNode("box", withFlex(1)),
	)
	LayoutFull(root, 0, 0, 80, 24)

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

// --- Test: VBox overflow=scroll — children get natural heights, overflow container ---

func TestLayout_VBox_OverflowScroll(t *testing.T) {
	// Container: height=10, overflow=scroll, 20 children each with height=3
	s := ds()
	s.Overflow = "scroll"
	s.Height = 10
	s.Width = 20

	var children []*Node
	for i := 0; i < 20; i++ {
		children = append(children, makeNode("box", withHeight(3)))
	}

	root := makeNode("vbox", s, children...)
	LayoutFull(root, 0, 0, 80, 24)

	// Root should be constrained to its declared size
	if root.W != 20 || root.H != 10 {
		t.Errorf("root size: got %dx%d, want 20x10", root.W, root.H)
	}

	// Children should be positioned sequentially, overflowing the container
	for i, child := range root.Children {
		expectedY := i * 3
		if child.Y != expectedY {
			t.Errorf("child[%d].Y = %d, want %d", i, child.Y, expectedY)
		}
		if child.H != 3 {
			t.Errorf("child[%d].H = %d, want 3", i, child.H)
		}
	}

	// ScrollHeight should be set to total content height
	expectedScrollH := 20 * 3 // 20 children × 3 height each
	if root.ScrollHeight != expectedScrollH {
		t.Errorf("ScrollHeight = %d, want %d", root.ScrollHeight, expectedScrollH)
	}
}

func TestLayout_VBox_OverflowScroll_WithGap(t *testing.T) {
	s := ds()
	s.Overflow = "scroll"
	s.Height = 10
	s.Width = 20
	s.Gap = 1

	root := makeNode("vbox", s,
		makeNode("box", withHeight(3)),
		makeNode("box", withHeight(3)),
		makeNode("box", withHeight(3)),
	)
	LayoutFull(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]
	c2 := root.Children[2]

	if c0.Y != 0 || c0.H != 3 {
		t.Errorf("child[0]: Y=%d H=%d, want Y=0 H=3", c0.Y, c0.H)
	}
	if c1.Y != 4 || c1.H != 3 {
		t.Errorf("child[1]: Y=%d H=%d, want Y=4 H=3", c1.Y, c1.H)
	}
	if c2.Y != 8 || c2.H != 3 {
		t.Errorf("child[2]: Y=%d H=%d, want Y=8 H=3", c2.Y, c2.H)
	}

	// ScrollHeight = last child bottom - contentY = (8+3) - 0 = 11
	if root.ScrollHeight != 11 {
		t.Errorf("ScrollHeight = %d, want 11", root.ScrollHeight)
	}
}

func TestLayout_VBox_OverflowScroll_FlexChildrenGetHeight1(t *testing.T) {
	// In scroll containers, flex children should get height 1 (not distributed)
	s := ds()
	s.Overflow = "scroll"
	s.Height = 10
	s.Width = 20

	root := makeNode("vbox", s,
		makeNode("box", withFlex(1)),
		makeNode("box", withFlex(2)),
		makeNode("box", withHeight(5)),
	)
	LayoutFull(root, 0, 0, 80, 24)

	c0 := root.Children[0]
	c1 := root.Children[1]
	c2 := root.Children[2]

	// Flex children get natural height 1
	if c0.H != 1 {
		t.Errorf("flex=1 child H = %d, want 1 (natural in scroll)", c0.H)
	}
	if c1.H != 1 {
		t.Errorf("flex=2 child H = %d, want 1 (natural in scroll)", c1.H)
	}
	// Fixed height child keeps its height
	if c2.H != 5 {
		t.Errorf("fixed child H = %d, want 5", c2.H)
	}

	// Sequential positioning
	if c0.Y != 0 {
		t.Errorf("child[0].Y = %d, want 0", c0.Y)
	}
	if c1.Y != 1 {
		t.Errorf("child[1].Y = %d, want 1", c1.Y)
	}
	if c2.Y != 2 {
		t.Errorf("child[2].Y = %d, want 2", c2.Y)
	}
}

func TestLayout_VBox_OverflowScroll_WithBorderAndPadding(t *testing.T) {
	s := ds()
	s.Overflow = "scroll"
	s.Height = 20
	s.Width = 30
	s.Border = "single"
	s.PaddingTop = 1
	s.PaddingBottom = 1
	s.PaddingLeft = 1
	s.PaddingRight = 1

	root := makeNode("vbox", s,
		makeNode("box", withHeight(5)),
		makeNode("box", withHeight(5)),
		makeNode("box", withHeight(5)),
	)
	LayoutFull(root, 0, 0, 80, 24)

	// Content area: border=1 + padding=1 = 2 on each side
	// contentX=2, contentY=2, contentW=30-4-1(scrollbar)=25, contentH=20-4=16
	c0 := root.Children[0]
	c1 := root.Children[1]
	c2 := root.Children[2]

	if c0.X != 2 {
		t.Errorf("child[0].X = %d, want 2", c0.X)
	}
	if c0.Y != 2 {
		t.Errorf("child[0].Y = %d, want 2", c0.Y)
	}
	if c1.Y != 7 {
		t.Errorf("child[1].Y = %d, want 7", c1.Y)
	}
	if c2.Y != 12 {
		t.Errorf("child[2].Y = %d, want 12", c2.Y)
	}

	// ScrollHeight = (12+5) - 2 = 15
	if root.ScrollHeight != 15 {
		t.Errorf("ScrollHeight = %d, want 15", root.ScrollHeight)
	}
}

func TestLayout_VBox_OverflowScroll_ScrollbarReservesColumn(t *testing.T) {
	s := ds()
	s.Overflow = "scroll"
	s.Height = 10
	s.Width = 20

	root := makeNode("vbox", s,
		makeNode("box", withHeight(3)),
	)
	LayoutFull(root, 0, 0, 80, 24)

	// Child width should be 19 (20 - 1 for scrollbar)
	child := root.Children[0]
	if child.W != 19 {
		t.Errorf("child.W = %d, want 19 (scrollbar reserves 1 column)", child.W)
	}
}

func TestLayoutHBoxWrap_BorderedButtonNotClampedToWidth1(t *testing.T) {
	// Narrow row tail used to clamp last flex item to w=1; paintBorder needs w>=2.
	inner := makeNode("hbox", Style{Height: 3, Border: "rounded", BorderColor: "#F5C842", PaddingLeft: 1, PaddingRight: 1}, makeText("XX"))
	comp := makeNode("component", ds(), inner)
	root := makeNode("hbox", Style{FlexWrap: "wrap"}, comp)
	layoutViewportW = 80
	layoutViewportH = 24
	normalizeSpacingInTree(root)
	computeFlex(root, 0, 0, 3, 5, 0)

	if inner.W < 2 {
		t.Fatalf("grafted button hbox W=%d, want >= 2 for border paint", inner.W)
	}
}

func TestLayoutHBoxNowrap_GrowsWhenGraftedChildTallerThanSlot(t *testing.T) {
	// Non-wrap hbox (e.g. Lux SplitButton): row may pass a short cross-axis h while
	// grafted inner hbox uses style.Height=3. Outer hbox must grow so border + fill
	// are not shorter than text backgrounds.
	inner := makeNode("hbox", Style{Height: 3, Width: 14, Border: "rounded"},
		makeTextStyled("Action", Style{Height: 3, Flex: 1, PaddingLeft: 1}),
		makeTextStyled("▼", Style{Height: 3, Width: 3}),
	)
	comp := makeNode("component", ds(), inner)
	root := makeNode("hbox", ds(), comp)
	setParentsRecursive(root)
	layoutViewportW = 50
	layoutViewportH = 24
	normalizeSpacingInTree(root)
	computeFlex(root, 0, 0, 50, 1, 0)

	if inner.H < 3 {
		t.Fatalf("inner split hbox H=%d, want >= 3", inner.H)
	}
	if comp.H < 3 {
		t.Fatalf("component placeholder H=%d, want >= 3", comp.H)
	}
}

func TestComponentPlaceholderSyncsSizeToGraftedRoot(t *testing.T) {
	// Flex-wrap can pass a tiny cross-axis h (e.g. 1) into computeFlex for the row;
	// LuxButton still lays out its grafted hbox at style.Height=3. The placeholder
	// must pick up that height so row stacking and border paint see the real box.
	inner := makeNode("hbox", Style{Height: 3, Width: 10, Border: "rounded", BorderColor: "#F5C842"}, makeText("Ok"))
	comp := makeNode("component", ds(), inner)
	root := makeNode("hbox", Style{FlexWrap: "wrap"}, comp)
	layoutViewportW = 50
	layoutViewportH = 24
	normalizeSpacingInTree(root)
	computeFlex(root, 0, 0, 50, 1, 0)

	if comp.H < 3 {
		t.Fatalf("component placeholder H=%d, want >= 3 after graft sync", comp.H)
	}
}
func TestComponentPlaceholderShrinksToGraftedRootWhenSlotTaller(t *testing.T) {
	// Parent may pass a larger cross-axis size than the grafted UI needs; the
	// placeholder must shrink to the real root so row height and paint bounds match.
	inner := makeNode("hbox", Style{Height: 3, Width: 8, Border: "rounded"}, makeText("Ok"))
	comp := makeNode("component", ds(), inner)
	root := makeNode("hbox", Style{FlexWrap: "wrap"}, comp)
	layoutViewportW = 50
	layoutViewportH = 24
	normalizeSpacingInTree(root)
	computeFlex(root, 0, 0, 50, 6, 0)

	if comp.H != inner.H {
		t.Fatalf("placeholder H=%d, inner H=%d, want equal after shrink sync", comp.H, inner.H)
	}
	if comp.W != inner.W {
		t.Fatalf("placeholder W=%d, inner W=%d, want equal after shrink sync", comp.W, inner.W)
	}
}

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		r    rune
		want int
	}{
		{'A', 1},    // ASCII
		{'z', 1},    // ASCII lowercase
		{'中', 2},   // CJK Unified Ideograph
		{'あ', 2},   // Hiragana
		{'한', 2},   // Korean Hangul
		{'é', 1},    // Latin extended
		{'\x00', 0}, // null
		{'\n', 0},   // newline (control char)
		{'\t', 0},   // tab (control char)
	}
	for _, tt := range tests {
		got := runeWidth(tt.r)
		if got != tt.want {
			t.Errorf("runeWidth(%q) = %d, want %d", tt.r, got, tt.want)
		}
	}
}
