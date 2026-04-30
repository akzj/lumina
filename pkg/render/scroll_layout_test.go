package render

import "testing"

// setParentsRecursive sets Parent pointers for all descendants.
func setParentsRecursive(node *Node) {
	for _, child := range node.Children {
		child.Parent = node
		setParentsRecursive(child)
	}
}

func TestScrollContainer_SingleChildIntrinsicHeight(t *testing.T) {
	// A scroll container with a single vbox child that has 3 text children.
	// Previously the vbox child would get finalH=1, collapsing all content.
	root := &Node{
		Type:  "vbox",
		Style: Style{Overflow: "scroll"},
		Children: []*Node{
			{
				Type:  "vbox",
				Style: Style{},
				Children: []*Node{
					{Type: "text", Content: "Line 1"},
					{Type: "text", Content: "Line 2"},
					{Type: "text", Content: "Line 3"},
				},
			},
		},
	}
	setParentsRecursive(root)
	LayoutFull(root, 0, 0, 40, 10)

	vbox := root.Children[0]
	if vbox.H < 3 {
		t.Errorf("expected vbox height >= 3, got %d", vbox.H)
	}
	// Each text child should be visible (not all stacked at same Y)
	for i, child := range vbox.Children {
		if child.H < 1 {
			t.Errorf("text child %d has height %d, expected >= 1", i, child.H)
		}
	}
	// Text children should be at different Y positions
	if len(vbox.Children) >= 3 {
		y0 := vbox.Children[0].Y
		y1 := vbox.Children[1].Y
		y2 := vbox.Children[2].Y
		if y1 <= y0 || y2 <= y1 {
			t.Errorf("text children should have increasing Y: %d, %d, %d", y0, y1, y2)
		}
	}
}

func TestScrollContainer_MultipleChildrenNoHeight(t *testing.T) {
	// Multiple children without explicit height should each get intrinsic height.
	root := &Node{
		Type:  "vbox",
		Style: Style{Overflow: "scroll"},
		Children: []*Node{
			{
				Type: "vbox",
				Children: []*Node{
					{Type: "text", Content: "A"},
					{Type: "text", Content: "B"},
				},
			},
			{
				Type: "vbox",
				Children: []*Node{
					{Type: "text", Content: "C"},
					{Type: "text", Content: "D"},
					{Type: "text", Content: "E"},
				},
			},
		},
	}
	setParentsRecursive(root)
	LayoutFull(root, 0, 0, 40, 5)

	child0 := root.Children[0]
	child1 := root.Children[1]
	if child0.H < 2 {
		t.Errorf("first child expected height >= 2, got %d", child0.H)
	}
	if child1.H < 3 {
		t.Errorf("second child expected height >= 3, got %d", child1.H)
	}
	// Second child should start after first
	if child1.Y <= child0.Y {
		t.Errorf("second child Y (%d) should be > first child Y (%d)", child1.Y, child0.Y)
	}
}

func TestScrollContainer_ExplicitHeightStillWorks(t *testing.T) {
	// Children with explicit height should still use that height (not intrinsic).
	root := &Node{
		Type:  "vbox",
		Style: Style{Overflow: "scroll"},
		Children: []*Node{
			{Type: "vbox", Style: Style{Height: 5}},
			{Type: "vbox", Style: Style{Height: 3}},
		},
	}
	setParentsRecursive(root)
	LayoutFull(root, 0, 0, 40, 4)

	if root.Children[0].H != 5 {
		t.Errorf("expected height 5, got %d", root.Children[0].H)
	}
	if root.Children[1].H != 3 {
		t.Errorf("expected height 3, got %d", root.Children[1].H)
	}
}

func TestScrollContainer_ComponentChildNoExplicitHeight(t *testing.T) {
	// Component child without explicit height and without grafted height
	// should use intrinsic measurement instead of defaulting to 1.
	root := &Node{
		Type:  "vbox",
		Style: Style{Overflow: "scroll"},
		Children: []*Node{
			{
				Type: "component",
				Children: []*Node{
					{
						Type: "vbox",
						Children: []*Node{
							{Type: "text", Content: "Item 1"},
							{Type: "text", Content: "Item 2"},
							{Type: "text", Content: "Item 3"},
							{Type: "text", Content: "Item 4"},
						},
					},
				},
			},
		},
	}
	setParentsRecursive(root)
	LayoutFull(root, 0, 0, 40, 3)

	comp := root.Children[0]
	// Component should have measured intrinsic height >= 4 (4 text lines)
	if comp.H < 4 {
		t.Errorf("component expected height >= 4, got %d", comp.H)
	}
}

func TestScrollContainer_NestedScrollDoesNotInfiniteLoop(t *testing.T) {
	// Ensure a scroll container inside a scroll container doesn't cause issues.
	root := &Node{
		Type:  "vbox",
		Style: Style{Overflow: "scroll"},
		Children: []*Node{
			{
				Type:  "vbox",
				Style: Style{Overflow: "scroll"},
				Children: []*Node{
					{Type: "text", Content: "Nested 1"},
					{Type: "text", Content: "Nested 2"},
				},
			},
		},
	}
	setParentsRecursive(root)
	// Should complete without hanging
	LayoutFull(root, 0, 0, 40, 5)

	inner := root.Children[0]
	if inner.H < 2 {
		t.Errorf("inner scroll container expected height >= 2, got %d", inner.H)
	}
}

func TestScrollContainer_MixedExplicitAndIntrinsic(t *testing.T) {
	// Mix of children with and without explicit heights.
	root := &Node{
		Type:  "vbox",
		Style: Style{Overflow: "scroll"},
		Children: []*Node{
			{Type: "vbox", Style: Style{Height: 3}},
			{
				Type: "vbox",
				Children: []*Node{
					{Type: "text", Content: "A"},
					{Type: "text", Content: "B"},
				},
			},
			{Type: "vbox", Style: Style{Height: 2}},
		},
	}
	setParentsRecursive(root)
	LayoutFull(root, 0, 0, 40, 4)

	c0 := root.Children[0]
	c1 := root.Children[1]
	c2 := root.Children[2]

	if c0.H != 3 {
		t.Errorf("first child (explicit h=3) got %d", c0.H)
	}
	if c1.H < 2 {
		t.Errorf("second child (intrinsic, 2 texts) expected >= 2, got %d", c1.H)
	}
	if c2.H != 2 {
		t.Errorf("third child (explicit h=2) got %d", c2.H)
	}

	// Verify ordering: c0 at top, c1 after c0, c2 after c1
	if c1.Y < c0.Y+c0.H {
		t.Errorf("c1.Y (%d) should be >= c0.Y+c0.H (%d)", c1.Y, c0.Y+c0.H)
	}
	if c2.Y < c1.Y+c1.H {
		t.Errorf("c2.Y (%d) should be >= c1.Y+c1.H (%d)", c2.Y, c1.Y+c1.H)
	}
}

func TestScrollContainer_ScrollHeightReflectsIntrinsic(t *testing.T) {
	// ScrollHeight should reflect the total intrinsic content height.
	root := &Node{
		Type:  "vbox",
		Style: Style{Overflow: "scroll"},
		Children: []*Node{
			{
				Type: "vbox",
				Children: []*Node{
					{Type: "text", Content: "L1"},
					{Type: "text", Content: "L2"},
					{Type: "text", Content: "L3"},
					{Type: "text", Content: "L4"},
					{Type: "text", Content: "L5"},
				},
			},
		},
	}
	setParentsRecursive(root)
	LayoutFull(root, 0, 0, 40, 3)

	// ScrollHeight should be >= 5 (5 text lines in the child)
	if root.ScrollHeight < 5 {
		t.Errorf("ScrollHeight expected >= 5, got %d", root.ScrollHeight)
	}
}

func TestScrollContainer_InnerVBoxStacksManyPanels(t *testing.T) {
	// Shell pattern: scroll > inner vbox (flex=1) > many panels. Intrinsic measure
	// used ~99999h on the inner vbox; children must not each flex-grow to split that.
	root := &Node{
		Type:  "vbox",
		Style: Style{Overflow: "scroll"},
		Children: []*Node{
			{
				Type:  "vbox",
				Style: Style{Flex: 1, Width: 40},
				Children: []*Node{
					{Type: "vbox", Style: Style{Height: 4}, Children: []*Node{{Type: "text", Content: "A"}}},
					{Type: "vbox", Style: Style{Height: 4}, Children: []*Node{{Type: "text", Content: "B"}}},
					{Type: "vbox", Style: Style{Height: 4}, Children: []*Node{{Type: "text", Content: "C"}}},
				},
			},
		},
	}
	setParentsRecursive(root)
	LayoutFull(root, 0, 0, 40, 8)

	inner := root.Children[0]
	wantMin := 4 + 4 + 4 // three h=4 panels (no gap on inner vbox in this tree)
	if inner.H < wantMin {
		t.Errorf("inner vbox height got %d, want >= %d (stacked panels, not flex-stretched)", inner.H, wantMin)
	}
	if root.ScrollHeight < wantMin {
		t.Errorf("ScrollHeight got %d, want >= %d", root.ScrollHeight, wantMin)
	}
	// Stacked vertically (not three equal flex slices of the viewport).
	c0, c1, c2 := inner.Children[0], inner.Children[1], inner.Children[2]
	if c1.Y != c0.Y+c0.H {
		t.Errorf("second panel Y: got %d want %d (first Y+H)", c1.Y, c0.Y+c0.H)
	}
	if c2.Y != c1.Y+c1.H {
		t.Errorf("third panel Y: got %d want %d", c2.Y, c1.Y+c1.H)
	}
}

func TestScrollContainer_InnerVBoxStacksManyComponentPanels(t *testing.T) {
	// Same as InnerVBoxStacksManyPanels but children are component placeholders
	// with grafted roots (mirrors Lux Card / defineComponent).
	panel := func(label string) *Node {
		return &Node{
			Type: "component",
			Children: []*Node{
				{
					Type:     "vbox",
					Style:    Style{Height: 4},
					Children: []*Node{{Type: "text", Content: label}},
				},
			},
		}
	}
	root := &Node{
		Type:  "vbox",
		Style: Style{Overflow: "scroll"},
		Children: []*Node{
			{
				Type:  "vbox",
				Style: Style{Flex: 1, Width: 40},
				Children: []*Node{
					panel("A"),
					panel("B"),
					panel("C"),
				},
			},
		},
	}
	setParentsRecursive(root)
	LayoutFull(root, 0, 0, 40, 8)

	inner := root.Children[0]
	wantMin := 4 + 4 + 4
	if inner.H < wantMin {
		t.Errorf("inner vbox height got %d, want >= %d", inner.H, wantMin)
	}
	if root.ScrollHeight < wantMin {
		t.Errorf("ScrollHeight got %d, want >= %d", root.ScrollHeight, wantMin)
	}
	c0, c1, c2 := inner.Children[0], inner.Children[1], inner.Children[2]
	if c1.Y != c0.Y+c0.H {
		t.Errorf("second panel Y: got %d want %d", c1.Y, c0.Y+c0.H)
	}
	if c2.Y != c1.Y+c1.H {
		t.Errorf("third panel Y: got %d want %d", c2.Y, c1.Y+c1.H)
	}
}

func TestScrollContainer_BoxBetweenScrollAndInner_StacksPanels(t *testing.T) {
	// Real shells may insert a non-scroll wrapper between overflow:scroll and the
	// content column; inner vbox must still use natural stacking (not flex-split).
	root := &Node{
		Type:  "vbox",
		Style: Style{Overflow: "scroll"},
		Children: []*Node{
			{
				Type:  "box",
				Style: Style{},
				Children: []*Node{
					{
						Type:  "vbox",
						Style: Style{Width: 40},
						Children: []*Node{
							{Type: "vbox", Style: Style{Height: 4}, Children: []*Node{{Type: "text", Content: "A"}}},
							{Type: "vbox", Style: Style{Height: 4}, Children: []*Node{{Type: "text", Content: "B"}}},
							{Type: "vbox", Style: Style{Height: 4}, Children: []*Node{{Type: "text", Content: "C"}}},
						},
					},
				},
			},
		},
	}
	setParentsRecursive(root)
	LayoutFull(root, 0, 0, 40, 8)

	inner := root.Children[0].Children[0]
	wantMin := 4 + 4 + 4
	if inner.H < wantMin {
		t.Errorf("inner vbox height got %d, want >= %d", inner.H, wantMin)
	}
	if root.ScrollHeight < wantMin {
		t.Errorf("ScrollHeight got %d, want >= %d", root.ScrollHeight, wantMin)
	}
	c0, c1, c2 := inner.Children[0], inner.Children[1], inner.Children[2]
	if c1.Y != c0.Y+c0.H {
		t.Errorf("second panel Y: got %d want %d", c1.Y, c0.Y+c0.H)
	}
	if c2.Y != c1.Y+c1.H {
		t.Errorf("third panel Y: got %d want %d", c2.Y, c1.Y+c1.H)
	}
}

func TestScrollContainer_InnerVBoxMinHeightDoesNotTruncateStack(t *testing.T) {
	// Pixel minHeight on the scroll content column must floor intrinsic height, not
	// replace it (otherwise only the first card fits and the rest look "unrendered").
	root := &Node{
		Type:  "vbox",
		Style: Style{Overflow: "scroll"},
		Children: []*Node{
			{
				Type:  "vbox",
				Style: Style{Flex: 1, Width: 40, MinHeight: 3},
				Children: []*Node{
					{Type: "vbox", Style: Style{Height: 4}, Children: []*Node{{Type: "text", Content: "A"}}},
					{Type: "vbox", Style: Style{Height: 4}, Children: []*Node{{Type: "text", Content: "B"}}},
					{Type: "vbox", Style: Style{Height: 4}, Children: []*Node{{Type: "text", Content: "C"}}},
				},
			},
		},
	}
	setParentsRecursive(root)
	LayoutFull(root, 0, 0, 40, 8)

	inner := root.Children[0]
	wantMin := 4 + 4 + 4
	if inner.H < wantMin {
		t.Errorf("inner vbox height got %d, want >= %d (minHeight=3 must not replace intrinsic)", inner.H, wantMin)
	}
	if root.ScrollHeight < wantMin {
		t.Errorf("ScrollHeight got %d, want >= %d", root.ScrollHeight, wantMin)
	}
}
