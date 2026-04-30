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

func TestMeasure_HBox_FlexDistributesWidth(t *testing.T) {
	// Simulates: Body hbox with sidebar (width=26) + main (flex=1)
	// Total width = 100. Main should get 100-26 = 74.
	sidebar := &Node{Type: "vbox", Style: Style{Width: 26}}
	main := &Node{
		Type:  "vbox",
		Style: Style{Flex: 1},
		Children: []*Node{
			{Type: "text", Content: "content"},
		},
	}
	body := &Node{
		Type:     "hbox",
		Children: []*Node{sidebar, main},
	}
	for _, c := range body.Children {
		c.Parent = body
	}
	for _, c := range main.Children {
		c.Parent = main
	}
	layoutViewportW = 100
	layoutViewportH = 40
	measure(body, Constraints{Width: 100, WidthMode: SizeModeExact, Height: 40, HeightMode: SizeModeExact})

	// Main should get 74 (100 - 26), not 100
	if main.MeasuredW != 74 {
		t.Errorf("main.MeasuredW=%d, want 74 (100-26)", main.MeasuredW)
	}
	// Sidebar should get 26
	if sidebar.MeasuredW != 26 {
		t.Errorf("sidebar.MeasuredW=%d, want 26", sidebar.MeasuredW)
	}
}

func TestMeasure_ButtonPage_CardHeight(t *testing.T) {
	// Simulates the Button page layout:
	// Root (vbox, 100x40)
	//   Body (hbox)
	//     Sidebar (vbox, width=26)
	//     Main (component, flex=1)  → grafts ScrollVBox
	//       ScrollVBox (vbox, overflow=scroll)
	//         InnerVBox (vbox)
	//           Card1 (component) → grafts CardBox
	//             CardBox (vbox, border=rounded, padding=1)
	//               Title (text)
	//               ButtonRow (hbox, wrap, gap=1)
	//                 Button1..7 (component) → grafts ButtonBox
	//                   ButtonBox (hbox, width=13, height=3, border=rounded)

	// Build the tree bottom-up
	makeButton := func(label string) *Node {
		btnInner := &Node{Type: "hbox", Style: Style{Width: 13, Height: 3, Border: "rounded"},
			Children: []*Node{{Type: "text", Content: label}}}
		btnInner.Children[0].Parent = btnInner
		btnComp := &Node{Type: "component", Children: []*Node{btnInner}}
		btnInner.Parent = btnComp
		return btnComp
	}

	buttons := make([]*Node, 7)
	for i := 0; i < 7; i++ {
		buttons[i] = makeButton("Btn")
	}

	buttonRow := &Node{
		Type:     "hbox",
		Style:    Style{FlexWrap: "wrap", Gap: 1},
		Children: buttons,
	}
	for _, b := range buttons {
		b.Parent = buttonRow
	}

	title := &Node{Type: "text", Content: "Default Buttons"}
	cardBox := &Node{
		Type:     "vbox",
		Style:    Style{Border: "rounded", PaddingTop: 1, PaddingBottom: 1, PaddingLeft: 1, PaddingRight: 1},
		Children: []*Node{title, buttonRow},
	}
	title.Parent = cardBox
	buttonRow.Parent = cardBox

	cardComp := &Node{Type: "component", Children: []*Node{cardBox}}
	cardBox.Parent = cardComp

	innerVBox := &Node{Type: "vbox", Children: []*Node{cardComp}}
	cardComp.Parent = innerVBox

	scrollVBox := &Node{
		Type:     "vbox",
		Style:    Style{Overflow: "scroll"},
		Children: []*Node{innerVBox},
	}
	innerVBox.Parent = scrollVBox

	mainComp := &Node{Type: "component", Style: Style{Flex: 1}, Children: []*Node{scrollVBox}}
	scrollVBox.Parent = mainComp

	sidebar := &Node{Type: "vbox", Style: Style{Width: 26}}

	body := &Node{Type: "hbox", Children: []*Node{sidebar, mainComp}}
	sidebar.Parent = body
	mainComp.Parent = body

	root := &Node{Type: "vbox", Children: []*Node{body}}
	body.Parent = root

	layoutViewportW = 100
	layoutViewportH = 40
	measure(root, Constraints{Width: 100, WidthMode: SizeModeExact, Height: 40, HeightMode: SizeModeExact})

	// Main should get 74 (100 - 26)
	if mainComp.MeasuredW != 74 {
		t.Errorf("mainComp.MeasuredW=%d, want 74", mainComp.MeasuredW)
	}

	// ScrollVBox should get 74 (passes through from component)
	if scrollVBox.MeasuredW != 74 {
		t.Errorf("scrollVBox.MeasuredW=%d, want 74", scrollVBox.MeasuredW)
	}

	// Card's measured height should be reasonable (not 32!)
	// With width=74, scrollbar=1, so inner=73, card border=2, padding=2, content=69
	// ButtonRow: 7 buttons × 13w = 91, but content width is only 67 (73-2border-2padding-1scrollbar... let's check)
	// Actually: scroll content width = 74 - 2*0border - 0padding - 1scrollbar = 73
	// innerVBox gets 73
	// cardBox outer = 73, content = 73 - 2border - 2padding = 67
	// buttonRow content = 67
	// Each button = 13w. 67/14 = 4 per row (13+1gap = 14, 4*14-1=55, 5th starts at 56+13=69 > 67)
	// Actually: 13+1+13+1+13+1+13 = 55 (4 buttons), next would be 55+1+13=69 > 67
	// So 4 per row, 2 rows (4+3). Row height = 3. Total = 3+1+3 = 7
	// Card height = 7 + 2border + 2padding + 1(title) = 12
	// That's the correct height!

	if cardComp.MeasuredH > 15 {
		t.Errorf("cardComp.MeasuredH=%d, want <= 15 (should be ~12, not 32)", cardComp.MeasuredH)
	}
	if cardComp.MeasuredH < 8 {
		t.Errorf("cardComp.MeasuredH=%d, want >= 8 (at least title + 1 row of buttons)", cardComp.MeasuredH)
	}
	t.Logf("Card measured: W=%d H=%d", cardComp.MeasuredW, cardComp.MeasuredH)
	t.Logf("ButtonRow measured: W=%d H=%d", buttonRow.MeasuredW, buttonRow.MeasuredH)
	t.Logf("ScrollVBox measured: W=%d H=%d", scrollVBox.MeasuredW, scrollVBox.MeasuredH)
}
