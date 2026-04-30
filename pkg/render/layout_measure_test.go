package render

import "testing"

func TestMeasure_TextNode_SingleLine(t *testing.T) {
	node := &Node{Type: "text", Content: "hello"}
	layoutViewportW = 80
	layoutViewportH = 24
	w, h := measure(node, Constraints{Width: 40, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
	if w != 40 {
		t.Errorf("text width=%d, want 40", w)
	}
	if h != 1 {
		t.Errorf("text height=%d, want 1", h)
	}
}

func TestMeasure_TextNode_Wrapping(t *testing.T) {
	// "abcdefghij" is 10 chars, should wrap to 2 lines at width 5
	node := &Node{Type: "text", Content: "abcdefghij"}
	layoutViewportW = 80
	layoutViewportH = 24
	w, h := measure(node, Constraints{Width: 5, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
	if w != 5 {
		t.Errorf("text width=%d, want 5", w)
	}
	if h != 2 {
		t.Errorf("text height=%d, want 2 (wrapped)", h)
	}
}

func TestMeasure_VBox_SumsChildHeights(t *testing.T) {
	root := &Node{
		Type: "vbox",
		Children: []*Node{
			{Type: "text", Content: "Line 1"},
			{Type: "text", Content: "Line 2"},
			{Type: "text", Content: "Line 3"},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
	}
	layoutViewportW = 80
	layoutViewportH = 24
	w, h := measure(root, Constraints{Width: 40, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
	if w != 40 {
		t.Errorf("vbox width=%d, want 40", w)
	}
	if h != 3 {
		t.Errorf("vbox height=%d, want 3 (sum of 3 text lines)", h)
	}
}

func TestMeasure_VBox_WithGap(t *testing.T) {
	root := &Node{
		Type:  "vbox",
		Style: Style{Gap: 1},
		Children: []*Node{
			{Type: "text", Content: "A"},
			{Type: "text", Content: "B"},
			{Type: "text", Content: "C"},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
	}
	layoutViewportW = 80
	layoutViewportH = 24
	_, h := measure(root, Constraints{Width: 40, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
	// 3 items + 2 gaps = 3 + 2 = 5
	if h != 5 {
		t.Errorf("vbox with gap height=%d, want 5", h)
	}
}

func TestMeasure_VBox_WithBorderAndPadding(t *testing.T) {
	root := &Node{
		Type:  "vbox",
		Style: Style{Border: "single", PaddingTop: 1, PaddingBottom: 1},
		Children: []*Node{
			{Type: "text", Content: "hello"},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
	}
	layoutViewportW = 80
	layoutViewportH = 24
	_, h := measure(root, Constraints{Width: 40, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
	// 1 text + 2 border + 2 padding = 5
	if h != 5 {
		t.Errorf("vbox with border+padding height=%d, want 5", h)
	}
}

func TestMeasure_ExplicitHeight_Overrides(t *testing.T) {
	root := &Node{
		Type:  "vbox",
		Style: Style{Height: 10},
		Children: []*Node{
			{Type: "text", Content: "A"},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
	}
	layoutViewportW = 80
	layoutViewportH = 24
	_, h := measure(root, Constraints{Width: 40, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
	if h != 10 {
		t.Errorf("explicit height=%d, want 10", h)
	}
}

func TestMeasure_Component_PassesThrough(t *testing.T) {
	inner := &Node{
		Type:  "vbox",
		Style: Style{Height: 5},
		Children: []*Node{
			{Type: "text", Content: "x"},
		},
	}
	comp := &Node{
		Type:     "component",
		Children: []*Node{inner},
	}
	inner.Parent = comp
	layoutViewportW = 80
	layoutViewportH = 24
	_, h := measure(comp, Constraints{Width: 40, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
	if h != 5 {
		t.Errorf("component measured height=%d, want 5 (from inner)", h)
	}
}

func TestMeasure_DisplayNone_ReturnsZero(t *testing.T) {
	root := &Node{
		Type:  "vbox",
		Style: Style{Display: "none"},
		Children: []*Node{
			{Type: "text", Content: "hidden"},
		},
	}
	layoutViewportW = 80
	layoutViewportH = 24
	w, h := measure(root, Constraints{Width: 40, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
	if w != 0 || h != 0 {
		t.Errorf("display:none measured=%dx%d, want 0x0", w, h)
	}
}

func TestMeasure_HBox_MaxChildHeight(t *testing.T) {
	root := &Node{
		Type: "hbox",
		Children: []*Node{
			{Type: "text", Content: "A"},
			{Type: "vbox", Style: Style{Height: 3}, Children: []*Node{{Type: "text", Content: "B"}}},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
		for _, gc := range c.Children {
			gc.Parent = c
		}
	}
	layoutViewportW = 80
	layoutViewportH = 24
	_, h := measure(root, Constraints{Width: 40, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
	if h != 3 {
		t.Errorf("hbox height=%d, want 3 (max of children)", h)
	}
}

func TestMeasure_MinHeight_Applied(t *testing.T) {
	root := &Node{
		Type:  "vbox",
		Style: Style{MinHeight: 10},
		Children: []*Node{
			{Type: "text", Content: "short"},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
	}
	layoutViewportW = 80
	layoutViewportH = 24
	_, h := measure(root, Constraints{Width: 40, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
	if h != 10 {
		t.Errorf("minHeight applied height=%d, want 10", h)
	}
}

func TestMeasure_MaxHeight_Applied(t *testing.T) {
	root := &Node{
		Type:  "vbox",
		Style: Style{MaxHeight: 2},
		Children: []*Node{
			{Type: "text", Content: "A"},
			{Type: "text", Content: "B"},
			{Type: "text", Content: "C"},
			{Type: "text", Content: "D"},
			{Type: "text", Content: "E"},
		},
	}
	for _, c := range root.Children {
		c.Parent = root
	}
	layoutViewportW = 80
	layoutViewportH = 24
	_, h := measure(root, Constraints{Width: 40, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
	if h != 2 {
		t.Errorf("maxHeight applied height=%d, want 2", h)
	}
}
