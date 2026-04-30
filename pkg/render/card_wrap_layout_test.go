package render

import "testing"

// Regression: Card + flex-wrap row + fixed-height LuxButton-like components under a
// scroll shell must measure the row tall enough; otherwise the card border clips
// through the middle of buttons (Raised / Outlined showcases).
func TestScrollContentStack_CardFlexWrapRowHeight(t *testing.T) {
	btn := func(key string) *Node {
		inner := makeNode("hbox", Style{
			Height: 3, PaddingLeft: 2, PaddingRight: 2,
			Border: "single", Right: -1, Bottom: -1,
			Justify: "center", Align: "center",
		}, makeText("Lbl"))
		comp := makeNode("component", ds(), inner)
		comp.Key = key
		return comp
	}
	wrap := makeNode("hbox", Style{
		FlexWrap:     "wrap",
		WidthPercent: 100,
		Gap:          1,
		Right:        -1, Bottom: -1,
	}, btn("a"), btn("b"))
	cardBody := makeNode("box", Style{
		Border: "rounded", Padding: 1, Right: -1, Bottom: -1,
	}, makeTextStyled("Raised", withHeight(1)), wrap)
	cardComp := makeNode("component", ds(), cardBody)

	scroll := makeNode("vbox", Style{Overflow: "scroll", Width: 60, Height: 30, Right: -1, Bottom: -1},
		makeNode("vbox", Style{Flex: 1, WidthPercent: 100, Right: -1, Bottom: -1}, cardComp),
	)
	setParentsRecursive(scroll)
	LayoutFull(scroll, 0, 0, 80, 24)

	title := cardBody.Children[0]
	row := cardBody.Children[1]
	if row.Type != "hbox" {
		t.Fatalf("expected wrap hbox second child, got %q", row.Type)
	}
	if row.H < 3 {
		t.Fatalf("wrap row outer H=%d, want >= 3 (button row)", row.H)
	}
	// Title + row (+ gaps/padding/border) must fit inside the card body.
	innerBottom := title.Y + title.H
	if row.Y < innerBottom {
		t.Fatalf("row Y=%d overlaps title bottom %d", row.Y, innerBottom)
	}
	cardBottom := cardBody.Y + cardBody.H
	for _, ch := range row.Children {
		b := ch.Y + ch.H
		if b > cardBottom {
			t.Fatalf("button extends below card: bottom=%d cardBottom=%d", b, cardBottom)
		}
	}
}
