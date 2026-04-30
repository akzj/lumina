package render

import "testing"

// --- flex-wrap tests ---

func TestFlexWrap_HBox_Basic(t *testing.T) {
	// 3 children each width=40 in container width=100
	// With wrap: row1=[40,40], row2=[40]
	root := &Node{Type: "hbox", Style: Style{FlexWrap: "wrap"}}
	c1 := &Node{Type: "box", Style: Style{Width: 40, Height: 5}, Parent: root}
	c2 := &Node{Type: "box", Style: Style{Width: 40, Height: 5}, Parent: root}
	c3 := &Node{Type: "box", Style: Style{Width: 40, Height: 5}, Parent: root}
	root.Children = []*Node{c1, c2, c3}

	LayoutFull(root, 0, 0, 100, 20)

	// Row 1: c1 at x=0, c2 at x=40
	// Row 2: c3 at x=0, y=5
	if c1.X != 0 || c1.Y != 0 || c1.W != 40 {
		t.Errorf("c1: got X=%d Y=%d W=%d, want X=0 Y=0 W=40", c1.X, c1.Y, c1.W)
	}
	if c2.X != 40 || c2.Y != 0 || c2.W != 40 {
		t.Errorf("c2: got X=%d Y=%d W=%d, want X=40 Y=0 W=40", c2.X, c2.Y, c2.W)
	}
	if c3.X != 0 || c3.Y != 5 {
		t.Errorf("c3: got X=%d Y=%d, want X=0 Y=5", c3.X, c3.Y)
	}
}

func TestFlexWrap_HBox_WithGap(t *testing.T) {
	// 3 children each width=40, gap=10, container width=100
	// Row 1: 40 + 10 + 40 = 90 (fits), Row 2: 40
	root := &Node{Type: "hbox", Style: Style{FlexWrap: "wrap", Gap: 10}}
	c1 := &Node{Type: "box", Style: Style{Width: 40, Height: 3}, Parent: root}
	c2 := &Node{Type: "box", Style: Style{Width: 40, Height: 3}, Parent: root}
	c3 := &Node{Type: "box", Style: Style{Width: 40, Height: 3}, Parent: root}
	root.Children = []*Node{c1, c2, c3}

	LayoutFull(root, 0, 0, 100, 20)

	// Row 1: c1 at x=0, c2 at x=50 (40+10)
	// Row 2: c3 at x=0, y=3+10=13
	if c1.X != 0 || c1.Y != 0 {
		t.Errorf("c1: got X=%d Y=%d, want X=0 Y=0", c1.X, c1.Y)
	}
	if c2.X != 50 || c2.Y != 0 {
		t.Errorf("c2: got X=%d Y=%d, want X=50 Y=0", c2.X, c2.Y)
	}
	if c3.X != 0 || c3.Y != 13 {
		t.Errorf("c3: got X=%d Y=%d, want X=0 Y=13", c3.X, c3.Y)
	}
}

func TestFlexWrap_HBox_RowHeight(t *testing.T) {
	// Children in same row with different heights — row height = max
	root := &Node{Type: "hbox", Style: Style{FlexWrap: "wrap"}}
	c1 := &Node{Type: "box", Style: Style{Width: 30, Height: 3}, Parent: root}
	c2 := &Node{Type: "box", Style: Style{Width: 30, Height: 7}, Parent: root}
	c3 := &Node{Type: "box", Style: Style{Width: 30, Height: 2}, Parent: root}
	root.Children = []*Node{c1, c2, c3}

	LayoutFull(root, 0, 0, 100, 30)

	// All 3 fit in one row (30+30+30=90 <= 100)
	// Row height = max(3, 7, 2) = 7
	if c1.Y != 0 || c2.Y != 0 || c3.Y != 0 {
		t.Errorf("All should be in row 0: c1.Y=%d c2.Y=%d c3.Y=%d", c1.Y, c2.Y, c3.Y)
	}
	if c1.X != 0 || c2.X != 30 || c3.X != 60 {
		t.Errorf("X positions: c1=%d c2=%d c3=%d, want 0,30,60", c1.X, c2.X, c3.X)
	}
}

func TestFlexWrap_HBox_WrapReverse(t *testing.T) {
	// wrap-reverse reverses the order of rows
	root := &Node{Type: "hbox", Style: Style{FlexWrap: "wrap-reverse"}}
	c1 := &Node{Type: "box", Style: Style{Width: 60, Height: 5}, Parent: root}
	c2 := &Node{Type: "box", Style: Style{Width: 60, Height: 5}, Parent: root}
	c3 := &Node{Type: "box", Style: Style{Width: 60, Height: 5}, Parent: root}
	root.Children = []*Node{c1, c2, c3}

	LayoutFull(root, 0, 0, 100, 20)

	// Without reverse: row1=[c1,c2(overflow)], actually c1=60 fits, c2=60 doesn't fit with c1
	// Wait: 60+60=120 > 100, so row1=[c1], row2=[c2], row3=[c3]
	// With wrap-reverse: rows are reversed, so row order is [c3],[c2],[c1]
	// c3 at y=0, c2 at y=5, c1 at y=10
	if c3.Y != 0 {
		t.Errorf("c3 (wrap-reverse first row): got Y=%d, want Y=0", c3.Y)
	}
	if c2.Y != 5 {
		t.Errorf("c2 (wrap-reverse second row): got Y=%d, want Y=5", c2.Y)
	}
	if c1.Y != 10 {
		t.Errorf("c1 (wrap-reverse third row): got Y=%d, want Y=10", c1.Y)
	}
}

func TestFlexWrap_VBox_Basic(t *testing.T) {
	// 3 children each height=40 in container height=100
	// With wrap: col1=[c1,c2], col2=[c3]
	root := &Node{Type: "vbox", Style: Style{FlexWrap: "wrap"}}
	c1 := &Node{Type: "box", Style: Style{Width: 20, Height: 40}, Parent: root}
	c2 := &Node{Type: "box", Style: Style{Width: 20, Height: 40}, Parent: root}
	c3 := &Node{Type: "box", Style: Style{Width: 20, Height: 40}, Parent: root}
	root.Children = []*Node{c1, c2, c3}

	LayoutFull(root, 0, 0, 100, 100)

	// Col 1: c1 at y=0, c2 at y=40
	// Col 2: c3 at x=20, y=0
	if c1.X != 0 || c1.Y != 0 {
		t.Errorf("c1: got X=%d Y=%d, want X=0 Y=0", c1.X, c1.Y)
	}
	if c2.X != 0 || c2.Y != 40 {
		t.Errorf("c2: got X=%d Y=%d, want X=0 Y=40", c2.X, c2.Y)
	}
	if c3.X != 20 || c3.Y != 0 {
		t.Errorf("c3: got X=%d Y=%d, want X=20 Y=0", c3.X, c3.Y)
	}
}

func TestFlexWrap_NoWrap_Default(t *testing.T) {
	// Default behavior: no wrapping (all in one row)
	root := &Node{Type: "hbox"}
	c1 := &Node{Type: "box", Style: Style{Width: 60, Height: 5}, Parent: root}
	c2 := &Node{Type: "box", Style: Style{Width: 60, Height: 5}, Parent: root}
	root.Children = []*Node{c1, c2}

	LayoutFull(root, 0, 0, 100, 20)

	// Both in same row, c2 gets clamped to remaining space
	if c1.Y != 0 || c2.Y != 0 {
		t.Errorf("Both should be at Y=0: c1.Y=%d c2.Y=%d", c1.Y, c2.Y)
	}
	if c1.X != 0 {
		t.Errorf("c1.X: got %d, want 0", c1.X)
	}
	if c2.X != 60 {
		t.Errorf("c2.X: got %d, want 60", c2.X)
	}
}

// --- Grid layout tests ---

func TestGrid_Basic(t *testing.T) {
	// 2x2 grid with "1fr 1fr" columns, 4 children
	root := &Node{
		Type: "box",
		Style: Style{
			Display:             "grid",
			GridTemplateColumns: "1fr 1fr",
		},
	}
	c1 := &Node{Type: "box", Parent: root}
	c2 := &Node{Type: "box", Parent: root}
	c3 := &Node{Type: "box", Parent: root}
	c4 := &Node{Type: "box", Parent: root}
	root.Children = []*Node{c1, c2, c3, c4}

	LayoutFull(root, 0, 0, 100, 20)

	// 2 columns of 50 each, 2 rows
	if c1.X != 0 || c1.W != 50 {
		t.Errorf("c1: got X=%d W=%d, want X=0 W=50", c1.X, c1.W)
	}
	if c2.X != 50 || c2.W != 50 {
		t.Errorf("c2: got X=%d W=%d, want X=50 W=50", c2.X, c2.W)
	}
	if c3.X != 0 {
		t.Errorf("c3: got X=%d, want X=0", c3.X)
	}
	if c4.X != 50 {
		t.Errorf("c4: got X=%d, want X=50", c4.X)
	}
	// c3 and c4 should be in row 2 (below c1 and c2)
	if c3.Y <= c1.Y {
		t.Errorf("c3.Y=%d should be > c1.Y=%d", c3.Y, c1.Y)
	}
	if c4.Y <= c2.Y {
		t.Errorf("c4.Y=%d should be > c2.Y=%d", c4.Y, c2.Y)
	}
}

func TestGrid_FrAndPx(t *testing.T) {
	// "100 1fr 2fr" in width=300 → [100, 66, 133]
	root := &Node{
		Type: "box",
		Style: Style{
			Display:             "grid",
			GridTemplateColumns: "100 1fr 2fr",
		},
	}
	c1 := &Node{Type: "box", Parent: root}
	c2 := &Node{Type: "box", Parent: root}
	c3 := &Node{Type: "box", Parent: root}
	root.Children = []*Node{c1, c2, c3}

	LayoutFull(root, 0, 0, 300, 10)

	// 300 - 100 = 200 remaining, split 1:2 → 66, 133
	if c1.X != 0 || c1.W != 100 {
		t.Errorf("c1: got X=%d W=%d, want X=0 W=100", c1.X, c1.W)
	}
	if c2.X != 100 {
		t.Errorf("c2: got X=%d, want X=100", c2.X)
	}
	// c2.W should be ~66 (200*1/3)
	if c2.W < 60 || c2.W > 70 {
		t.Errorf("c2.W: got %d, want ~66", c2.W)
	}
	// c3.W should be ~133 (200*2/3)
	if c3.W < 125 || c3.W > 140 {
		t.Errorf("c3.W: got %d, want ~133", c3.W)
	}
}

func TestGrid_ExplicitPlacement(t *testing.T) {
	// Child with gridColumn="2", gridRow="1" placed in correct cell
	root := &Node{
		Type: "box",
		Style: Style{
			Display:             "grid",
			GridTemplateColumns: "1fr 1fr 1fr",
		},
	}
	// c1: auto-placed at (1,1)
	c1 := &Node{Type: "box", Parent: root}
	// c2: explicitly placed at column 3, row 1
	c2 := &Node{Type: "box", Style: Style{GridColumnStart: 3, GridColumnEnd: 4, GridRowStart: 1, GridRowEnd: 2}, Parent: root}
	root.Children = []*Node{c1, c2}

	LayoutFull(root, 0, 0, 90, 10)

	// 3 columns of 30 each
	if c1.X != 0 || c1.W != 30 {
		t.Errorf("c1: got X=%d W=%d, want X=0 W=30", c1.X, c1.W)
	}
	// c2 at column 3 → x=60
	if c2.X != 60 || c2.W != 30 {
		t.Errorf("c2: got X=%d W=%d, want X=60 W=30", c2.X, c2.W)
	}
}

func TestGrid_AutoPlacement(t *testing.T) {
	// 6 children in a 3-column grid, auto-placed
	root := &Node{
		Type: "box",
		Style: Style{
			Display:             "grid",
			GridTemplateColumns: "1fr 1fr 1fr",
		},
	}
	children := make([]*Node, 6)
	for i := range children {
		children[i] = &Node{Type: "box", Parent: root}
	}
	root.Children = children

	LayoutFull(root, 0, 0, 90, 20)

	// 3 columns of 30 each, 2 rows
	// Row 1: children[0..2] at y=0
	// Row 2: children[3..5] at y > 0
	for i := 0; i < 3; i++ {
		expectedX := i * 30
		if children[i].X != expectedX {
			t.Errorf("children[%d].X: got %d, want %d", i, children[i].X, expectedX)
		}
	}
	for i := 3; i < 6; i++ {
		expectedX := (i - 3) * 30
		if children[i].X != expectedX {
			t.Errorf("children[%d].X: got %d, want %d", i, children[i].X, expectedX)
		}
		if children[i].Y <= children[0].Y {
			t.Errorf("children[%d].Y=%d should be > children[0].Y=%d", i, children[i].Y, children[0].Y)
		}
	}
}

func TestGrid_Gap(t *testing.T) {
	// 2x2 grid with gap=10
	root := &Node{
		Type: "box",
		Style: Style{
			Display:             "grid",
			GridTemplateColumns: "1fr 1fr",
			Gap:                 10,
		},
	}
	c1 := &Node{Type: "box", Parent: root}
	c2 := &Node{Type: "box", Parent: root}
	c3 := &Node{Type: "box", Parent: root}
	c4 := &Node{Type: "box", Parent: root}
	root.Children = []*Node{c1, c2, c3, c4}

	LayoutFull(root, 0, 0, 110, 30)

	// 110 - 10 (gap) = 100, split 1:1 → 50 each
	// c1 at x=0, c2 at x=60 (50+10)
	if c1.X != 0 || c1.W != 50 {
		t.Errorf("c1: got X=%d W=%d, want X=0 W=50", c1.X, c1.W)
	}
	if c2.X != 60 || c2.W != 50 {
		t.Errorf("c2: got X=%d W=%d, want X=60 W=50", c2.X, c2.W)
	}
	// Row 2 should be offset by row height + gap
	if c3.Y <= c1.Y {
		t.Errorf("c3.Y=%d should be > c1.Y=%d", c3.Y, c1.Y)
	}
}

func TestGrid_Span(t *testing.T) {
	// Child spanning multiple columns: gridColumn="1 / 3"
	root := &Node{
		Type: "box",
		Style: Style{
			Display:             "grid",
			GridTemplateColumns: "1fr 1fr 1fr",
		},
	}
	// c1 spans columns 1-3 (first two columns)
	c1 := &Node{Type: "box", Style: Style{GridColumn: "1 / 3"}, Parent: root}
	// c2 auto-placed
	c2 := &Node{Type: "box", Parent: root}
	root.Children = []*Node{c1, c2}

	LayoutFull(root, 0, 0, 90, 10)

	// 3 columns of 30 each
	// c1 spans cols 1-2 → width = 30 + 30 = 60
	if c1.X != 0 || c1.W != 60 {
		t.Errorf("c1: got X=%d W=%d, want X=0 W=60", c1.X, c1.W)
	}
	// c2 auto-placed: c1 occupies (row1,col1) and (row1,col2), so c2 goes to (row1,col3)
	if c2.X != 60 || c2.W != 30 {
		t.Errorf("c2: got X=%d W=%d, want X=60 W=30", c2.X, c2.W)
	}
}

func TestGrid_GridColumnGap_GridRowGap(t *testing.T) {
	// Separate column and row gaps
	root := &Node{
		Type: "box",
		Style: Style{
			Display:             "grid",
			GridTemplateColumns: "1fr 1fr",
			GridColumnGap:       5,
			GridRowGap:          10,
		},
	}
	c1 := &Node{Type: "box", Parent: root}
	c2 := &Node{Type: "box", Parent: root}
	c3 := &Node{Type: "box", Parent: root}
	c4 := &Node{Type: "box", Parent: root}
	root.Children = []*Node{c1, c2, c3, c4}

	LayoutFull(root, 0, 0, 105, 30)

	// 105 - 5 (colGap) = 100, split 1:1 → 50 each
	if c1.X != 0 || c1.W != 50 {
		t.Errorf("c1: got X=%d W=%d, want X=0 W=50", c1.X, c1.W)
	}
	if c2.X != 55 || c2.W != 50 {
		t.Errorf("c2: got X=%d W=%d, want X=55 W=50", c2.X, c2.W)
	}
	// Row gap = 10, so c3.Y = c1.H + 10
	rowGapDiff := c3.Y - (c1.Y + c1.H)
	if rowGapDiff != 10 {
		t.Errorf("row gap: c3.Y - (c1.Y+c1.H) = %d, want 10", rowGapDiff)
	}
}

func TestGrid_GridRowTemplate(t *testing.T) {
	// Explicit row template
	root := &Node{
		Type: "box",
		Style: Style{
			Display:             "grid",
			GridTemplateColumns: "1fr 1fr",
			GridTemplateRows:    "5 10",
		},
	}
	c1 := &Node{Type: "box", Parent: root}
	c2 := &Node{Type: "box", Parent: root}
	c3 := &Node{Type: "box", Parent: root}
	c4 := &Node{Type: "box", Parent: root}
	root.Children = []*Node{c1, c2, c3, c4}

	LayoutFull(root, 0, 0, 100, 30)

	// Row 1 height=5, Row 2 height=10
	if c1.H != 5 {
		t.Errorf("c1.H: got %d, want 5", c1.H)
	}
	if c3.H != 10 {
		t.Errorf("c3.H: got %d, want 10", c3.H)
	}
	if c3.Y != 5 {
		t.Errorf("c3.Y: got %d, want 5", c3.Y)
	}
}

func TestParseGridTemplate(t *testing.T) {
	tests := []struct {
		template string
		avail    int
		gap      int
		want     []int
	}{
		{"1fr 1fr", 100, 0, []int{50, 50}},
		{"1fr 2fr", 90, 0, []int{30, 60}},
		{"50 1fr", 100, 0, []int{50, 50}},
		{"1fr 1fr 1fr", 90, 0, []int{30, 30, 30}},
		{"50 1fr 2fr", 300, 0, []int{50, 83, 166}},
		{"1fr 1fr", 110, 10, []int{50, 50}},
	}

	for _, tt := range tests {
		got := parseGridTemplate(tt.template, tt.avail, tt.gap)
		if len(got) != len(tt.want) {
			t.Errorf("parseGridTemplate(%q, %d, %d): got %v, want %v", tt.template, tt.avail, tt.gap, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseGridTemplate(%q, %d, %d)[%d]: got %d, want %d", tt.template, tt.avail, tt.gap, i, got[i], tt.want[i])
			}
		}
	}
}

func TestParseGridSpan(t *testing.T) {
	tests := []struct {
		input     string
		wantStart int
		wantEnd   int
	}{
		{"1", 1, 2},
		{"3", 3, 4},
		{"1 / 3", 1, 3},
		{"2 / 5", 2, 5},
		{"", 0, 0},
	}

	for _, tt := range tests {
		start, end := parseGridSpan(tt.input)
		if start != tt.wantStart || end != tt.wantEnd {
			t.Errorf("parseGridSpan(%q): got (%d, %d), want (%d, %d)", tt.input, start, end, tt.wantStart, tt.wantEnd)
		}
	}
}
