package render

import "testing"

// TestScrollbarClick verifies that clicking on the scrollbar column changes ScrollY.
func TestScrollbarClick(t *testing.T) {
	// Create a scroll container: width=20, height=10, border, with tall content.
	root := &Node{
		Type: "vbox",
		X:    0,
		Y:    0,
		W:    20,
		H:    10,
		Style: Style{
			Overflow: "scroll",
			Border:   "single",
		},
		ScrollHeight: 50, // content is 50 rows tall
	}
	// Child that fills the content
	child := &Node{
		Type:   "vbox",
		X:      1,
		Y:      1,
		W:      18,
		H:      50,
		Parent: root,
	}
	root.Children = []*Node{child}

	// The scrollbar X should be at: node.X + node.W - bw - paddingRight - 1 = 0 + 20 - 1 - 0 - 1 = 18
	scrollbarX := 18

	// Inner Y range: Y+bw to Y+H-bw = 1 to 9 (exclusive) → valid: 1..8
	// visibleH = 8, totalH = 50, thumbSize = 8*8/50 = 1
	// With ScrollY=0, thumbPos=0, so clicking at y=5 (below thumb) should page down.

	e := &Engine{
		layers: []*Layer{{Root: root}},
	}

	// Click below thumb → page down
	handled := e.handleScrollbarClick(root, scrollbarX, 5)
	if !handled {
		t.Fatal("handleScrollbarClick returned false, expected true")
	}
	if root.ScrollY == 0 {
		t.Error("ScrollY should have changed after clicking below thumb")
	}
	if root.ScrollY < 0 {
		t.Error("ScrollY should not be negative")
	}
	t.Logf("After click below thumb: ScrollY=%d", root.ScrollY)
}

// TestScrollbarClick_MatchesPainter verifies click handler and painter agree on scrollbar X.
func TestScrollbarClick_MatchesPainter(t *testing.T) {
	// scrollbarX = node.X + node.W - bw - paddingRight - 1
	// With no border, no padding: X=0, W=20 → scrollbarX = 19
	// Painter draws at innerX2-1 = (0+20-0-0)-1 = 19 → matches!
	root := &Node{
		Type: "vbox",
		X:    0,
		Y:    0,
		W:    20,
		H:    10,
		Style: Style{
			Overflow: "scroll",
		},
		ScrollHeight: 50,
	}
	child := &Node{
		Type:   "vbox",
		X:      0,
		Y:      0,
		W:      20,
		H:      50,
		Parent: root,
	}
	root.Children = []*Node{child}

	e := &Engine{
		layers: []*Layer{{Root: root}},
	}

	// Click at x=19 SHOULD match (scrollbar column)
	handled := e.handleScrollbarClick(root, 19, 3)
	if !handled {
		t.Error("x=19 should be the scrollbar column for no-border W=20 node")
	}

	// Click at x=20 should NOT match (outside node)
	root.ScrollY = 0 // reset
	handled = e.handleScrollbarClick(root, 20, 3)
	if handled {
		t.Error("x=20 should NOT be the scrollbar column")
	}
}

// TestScrollbarDrag verifies mouseDown on thumb → mouseMove → mouseUp changes ScrollY proportionally.
func TestScrollbarDrag(t *testing.T) {
	// Create a scroll container with known geometry
	root := &Node{
		Type: "vbox",
		X:    0,
		Y:    0,
		W:    20,
		H:    12,
		Style: Style{
			Overflow: "scroll",
			Border:   "single",
		},
		ScrollHeight: 100, // tall content
		ScrollY:      0,
	}
	child := &Node{
		Type:   "vbox",
		X:      1,
		Y:      1,
		W:      18,
		H:      100,
		Parent: root,
	}
	root.Children = []*Node{child}

	e := &Engine{
		layers: []*Layer{{Root: root}},
	}

	// Geometry:
	// bw=1, innerY1=1, innerY2=11, visibleH=10
	// totalH=100, thumbSize=10*10/100=1, trackSpace=10-1=9
	// maxScroll = 100-10 = 90
	// With ScrollY=0, thumbPos=0 → thumb at y=1 (innerY1 + 0)

	// scrollbarX = 0 + 20 - 1 - 0 - 1 = 18
	scrollbarX := 18
	thumbY := 1 // innerY1 + thumbPos = 1 + 0

	// Start drag on thumb
	started := e.handleScrollbarDragStart(scrollbarX, thumbY)
	if !started {
		t.Fatal("handleScrollbarDragStart should have found thumb at (19, 1)")
	}
	if e.scrollbarDragNode != root {
		t.Fatal("scrollbarDragNode should be root")
	}

	// Drag down by 5 pixels
	e.HandleMouseMove(scrollbarX, thumbY+5)

	// Expected: dy=5, scrollDelta = 5 * 90 / 9 = 50
	expectedScrollY := 50
	if root.ScrollY != expectedScrollY {
		t.Errorf("after drag down 5: ScrollY=%d, want %d", root.ScrollY, expectedScrollY)
	}

	// Drag further down by 4 more (total 9 from start)
	e.HandleMouseMove(scrollbarX, thumbY+9)

	// Expected: dy=9, scrollDelta = 9 * 90 / 9 = 90
	expectedScrollY = 90
	if root.ScrollY != expectedScrollY {
		t.Errorf("after drag down 9: ScrollY=%d, want %d", root.ScrollY, expectedScrollY)
	}

	// Drag beyond max (should clamp)
	e.HandleMouseMove(scrollbarX, thumbY+20)
	if root.ScrollY != 90 {
		t.Errorf("after drag beyond max: ScrollY=%d, want 90 (clamped)", root.ScrollY)
	}

	// Release
	e.HandleMouseUp(scrollbarX, thumbY+20)
	if e.scrollbarDragNode != nil {
		t.Error("scrollbarDragNode should be nil after mouseUp")
	}
}

// TestScrollbarDrag_NegativeClamp verifies dragging above start clamps to 0.
func TestScrollbarDrag_NegativeClamp(t *testing.T) {
	root := &Node{
		Type: "vbox",
		X:    0,
		Y:    0,
		W:    20,
		H:    12,
		Style: Style{
			Overflow: "scroll",
			Border:   "single",
		},
		ScrollHeight: 100,
		ScrollY:      45, // start in the middle
	}
	child := &Node{
		Type:   "vbox",
		X:      1,
		Y:      1,
		W:      18,
		H:      100,
		Parent: root,
	}
	root.Children = []*Node{child}

	e := &Engine{
		layers: []*Layer{{Root: root}},
	}

	// Geometry: bw=1, visibleH=10, totalH=100, thumbSize=1, trackSpace=9, maxScroll=90
	// thumbPos = 45 * 9 / 90 = 4.5 → 4 (int division)
	// Thumb at y = innerY1 + thumbPos = 1 + 4 = 5
	// scrollbarX = 0 + 20 - 1 - 0 - 1 = 18
	scrollbarX := 18
	thumbY := 5

	started := e.handleScrollbarDragStart(scrollbarX, thumbY)
	if !started {
		t.Fatal("should start drag")
	}

	// Drag up beyond start
	e.HandleMouseMove(scrollbarX, thumbY-10)

	// dy = -10, scrollDelta = -10 * 90 / 9 = -100
	// newScrollY = 45 + (-100) = -55 → clamped to 0
	if root.ScrollY != 0 {
		t.Errorf("ScrollY=%d, want 0 (clamped)", root.ScrollY)
	}

	e.HandleMouseUp(scrollbarX, thumbY-10)
}
