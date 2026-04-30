package render

import "testing"

// --- flexShrink tests ---

func TestFlexShrink_VBox(t *testing.T) {
	// Parent height=10, two children each want height=8, both have flexShrink=1
	// Total=16, overflow=6, each shrinks by 3 → each gets 5
	root := &Node{Type: "vbox", Style: Style{}, Children: []*Node{
		{Type: "vbox", Style: Style{Height: 8, FlexShrink: 1}},
		{Type: "vbox", Style: Style{Height: 8, FlexShrink: 1}},
	}}
	for _, c := range root.Children {
		c.Parent = root
	}
	LayoutFull(root, 0, 0, 20, 10)

	c0 := root.Children[0]
	c1 := root.Children[1]
	if c0.H != 5 {
		t.Errorf("child0 H = %d, want 5", c0.H)
	}
	if c1.H != 5 {
		t.Errorf("child1 H = %d, want 5", c1.H)
	}
}

func TestFlexShrink_HBox(t *testing.T) {
	// Parent width=20, two children each want width=15, both have flexShrink=1
	// Total=30, overflow=10, each shrinks by 5 → each gets 10
	root := &Node{Type: "hbox", Style: Style{}, Children: []*Node{
		{Type: "vbox", Style: Style{Width: 15, FlexShrink: 1}},
		{Type: "vbox", Style: Style{Width: 15, FlexShrink: 1}},
	}}
	for _, c := range root.Children {
		c.Parent = root
	}
	LayoutFull(root, 0, 0, 20, 10)

	c0 := root.Children[0]
	c1 := root.Children[1]
	if c0.W != 10 {
		t.Errorf("child0 W = %d, want 10", c0.W)
	}
	if c1.W != 10 {
		t.Errorf("child1 W = %d, want 10", c1.W)
	}
}

func TestFlexShrink_NoShrink(t *testing.T) {
	// Children without flexShrink don't shrink — they keep fixed size
	// even if that causes overflow. Both keep their explicit width.
	root := &Node{Type: "hbox", Style: Style{}, Children: []*Node{
		{Type: "vbox", Style: Style{Width: 15}},
		{Type: "vbox", Style: Style{Width: 15}},
	}}
	for _, c := range root.Children {
		c.Parent = root
	}
	LayoutFull(root, 0, 0, 20, 10)

	c0 := root.Children[0]
	c1 := root.Children[1]
	// Without flexShrink, children keep their fixed width (overflow allowed)
	if c0.W != 15 {
		t.Errorf("child0 W = %d, want 15 (no shrink)", c0.W)
	}
	if c1.W != 15 {
		t.Errorf("child1 W = %d, want 15 (no shrink, overflow)", c1.W)
	}
}

func TestFlexShrink_Proportional(t *testing.T) {
	// Unequal shrink factors: child0 shrink=2, child1 shrink=1
	// Parent height=10, child0 height=8, child1 height=8
	// Total=16, overflow=6, shrinkTotal=3
	// child0 shrinks by 6*2/3=4 → 8-4=4
	// child1 shrinks by 6*1/3=2 → 8-2=6
	root := &Node{Type: "vbox", Style: Style{}, Children: []*Node{
		{Type: "vbox", Style: Style{Height: 8, FlexShrink: 2}},
		{Type: "vbox", Style: Style{Height: 8, FlexShrink: 1}},
	}}
	for _, c := range root.Children {
		c.Parent = root
	}
	LayoutFull(root, 0, 0, 20, 10)

	c0 := root.Children[0]
	c1 := root.Children[1]
	if c0.H != 4 {
		t.Errorf("child0 H = %d, want 4", c0.H)
	}
	if c1.H != 6 {
		t.Errorf("child1 H = %d, want 6", c1.H)
	}
}

// --- flexBasis tests ---

func TestFlexBasis_VBox(t *testing.T) {
	// Parent height=20, child with flexBasis=5 and flex=1, another with flex=1
	// fixedTotal from basis child = 5, flexTotal = 2
	// remainH = 20 - 5 = 15
	// child0 (has basis): base=5 + 15*1/2=7 → finalH=12
	// child1 (no basis): base=0 + 15*1/2=7 → finalH=7 (but min 1 → 7)
	root := &Node{Type: "vbox", Style: Style{}, Children: []*Node{
		{Type: "vbox", Style: Style{Flex: 1, FlexBasis: 5}},
		{Type: "vbox", Style: Style{Flex: 1}},
	}}
	for _, c := range root.Children {
		c.Parent = root
	}
	LayoutFull(root, 0, 0, 20, 20)

	c0 := root.Children[0]
	c1 := root.Children[1]
	// c0: base=5 + 15*1/2 = 5+7 = 12
	// c1: base=0 + 15*1/2 = 7
	if c0.H != 12 {
		t.Errorf("child0 H = %d, want 12 (flexBasis=5 + grow)", c0.H)
	}
	if c1.H != 7 {
		t.Errorf("child1 H = %d, want 7 (no basis + grow)", c1.H)
	}
}

func TestFlexBasis_HBox(t *testing.T) {
	// Parent width=30, child with flexBasis=10 and flex=1, another with flex=1
	// fixedTotal = 10, flexTotal = 2
	// remainW = 30 - 10 = 20
	// child0: base=10 + 20*1/2=10 → 20
	// child1: base=0 + 20*1/2=10 → 10
	root := &Node{Type: "hbox", Style: Style{}, Children: []*Node{
		{Type: "vbox", Style: Style{Flex: 1, FlexBasis: 10}},
		{Type: "vbox", Style: Style{Flex: 1}},
	}}
	for _, c := range root.Children {
		c.Parent = root
	}
	LayoutFull(root, 0, 0, 30, 10)

	c0 := root.Children[0]
	c1 := root.Children[1]
	if c0.W != 20 {
		t.Errorf("child0 W = %d, want 20 (flexBasis=10 + grow)", c0.W)
	}
	if c1.W != 10 {
		t.Errorf("child1 W = %d, want 10 (no basis + grow)", c1.W)
	}
}

// --- alignSelf tests ---

func TestAlignSelf_Override_VBox(t *testing.T) {
	// Parent align=start (default stretch), child0 has alignSelf=center, child1 has alignSelf=end
	// Both children have explicit width=6 in a container width=20
	root := &Node{Type: "vbox", Style: Style{Align: "start"}, Children: []*Node{
		{Type: "text", Style: Style{Width: 6, AlignSelf: "center"}, Content: "hello"},
		{Type: "text", Style: Style{Width: 6, AlignSelf: "end"}, Content: "world"},
	}}
	for _, c := range root.Children {
		c.Parent = root
	}
	LayoutFull(root, 0, 0, 20, 10)

	c0 := root.Children[0]
	c1 := root.Children[1]
	// center: X = (20-6)/2 = 7
	if c0.X != 7 {
		t.Errorf("child0 (alignSelf=center) X = %d, want 7", c0.X)
	}
	// end: X = 20-6 = 14
	if c1.X != 14 {
		t.Errorf("child1 (alignSelf=end) X = %d, want 14", c1.X)
	}
}

func TestAlignSelf_Override_HBox(t *testing.T) {
	// Parent align=start, child0 has alignSelf=center with height=4, container height=10
	root := &Node{Type: "hbox", Style: Style{Align: "start"}, Children: []*Node{
		{Type: "vbox", Style: Style{Width: 5, Height: 4, AlignSelf: "center"}},
		{Type: "vbox", Style: Style{Width: 5, Height: 4, AlignSelf: "end"}},
	}}
	for _, c := range root.Children {
		c.Parent = root
	}
	LayoutFull(root, 0, 0, 20, 10)

	c0 := root.Children[0]
	c1 := root.Children[1]
	// center: Y = (10-4)/2 = 3
	if c0.Y != 3 {
		t.Errorf("child0 (alignSelf=center) Y = %d, want 3", c0.Y)
	}
	// end: Y = 10-4 = 6
	if c1.Y != 6 {
		t.Errorf("child1 (alignSelf=end) Y = %d, want 6", c1.Y)
	}
}

// --- order tests ---

func TestOrder_VBox(t *testing.T) {
	// Three children with order=2,0,1 should be positioned in order 0,1,2
	// i.e., child1 (order=0) at top, child2 (order=1) middle, child0 (order=2) bottom
	root := &Node{Type: "vbox", Style: Style{}, Children: []*Node{
		{Type: "text", Style: Style{Order: 2}, Content: "A"},
		{Type: "text", Style: Style{Order: 0}, Content: "B"},
		{Type: "text", Style: Style{Order: 1}, Content: "C"},
	}}
	for _, c := range root.Children {
		c.Parent = root
	}
	LayoutFull(root, 0, 0, 20, 10)

	// Expected layout order: B(order=0) → C(order=1) → A(order=2)
	c0 := root.Children[0] // order=2, should be last (Y=2)
	c1 := root.Children[1] // order=0, should be first (Y=0)
	c2 := root.Children[2] // order=1, should be middle (Y=1)

	if c1.Y != 0 {
		t.Errorf("child1 (order=0) Y = %d, want 0", c1.Y)
	}
	if c2.Y != 1 {
		t.Errorf("child2 (order=1) Y = %d, want 1", c2.Y)
	}
	if c0.Y != 2 {
		t.Errorf("child0 (order=2) Y = %d, want 2", c0.Y)
	}
}

func TestOrder_HBox(t *testing.T) {
	// Three children with order=2,0,1, each width=5, parent width=15
	root := &Node{Type: "hbox", Style: Style{}, Children: []*Node{
		{Type: "vbox", Style: Style{Width: 5, Order: 2}},
		{Type: "vbox", Style: Style{Width: 5, Order: 0}},
		{Type: "vbox", Style: Style{Width: 5, Order: 1}},
	}}
	for _, c := range root.Children {
		c.Parent = root
	}
	LayoutFull(root, 0, 0, 15, 10)

	c0 := root.Children[0] // order=2, should be last (X=10)
	c1 := root.Children[1] // order=0, should be first (X=0)
	c2 := root.Children[2] // order=1, should be middle (X=5)

	if c1.X != 0 {
		t.Errorf("child1 (order=0) X = %d, want 0", c1.X)
	}
	if c2.X != 5 {
		t.Errorf("child2 (order=1) X = %d, want 5", c2.X)
	}
	if c0.X != 10 {
		t.Errorf("child0 (order=2) X = %d, want 10", c0.X)
	}
}

func TestOrder_DefaultPreserved(t *testing.T) {
	// All order=0 (default), original array order preserved
	root := &Node{Type: "vbox", Style: Style{}, Children: []*Node{
		{Type: "text", Style: Style{}, Content: "first"},
		{Type: "text", Style: Style{}, Content: "second"},
		{Type: "text", Style: Style{}, Content: "third"},
	}}
	for _, c := range root.Children {
		c.Parent = root
	}
	LayoutFull(root, 0, 0, 20, 10)

	c0 := root.Children[0]
	c1 := root.Children[1]
	c2 := root.Children[2]
	if c0.Y != 0 {
		t.Errorf("child0 Y = %d, want 0", c0.Y)
	}
	if c1.Y != 1 {
		t.Errorf("child1 Y = %d, want 1", c1.Y)
	}
	if c2.Y != 2 {
		t.Errorf("child2 Y = %d, want 2", c2.Y)
	}
}
