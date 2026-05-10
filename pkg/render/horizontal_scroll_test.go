package render

import (
	"testing"
)

func TestHorizontalScroll_ScrollWidthComputed(t *testing.T) {
	// Create a vbox with overflow:scroll, width=20, containing an hbox child with width=50
	root := &Node{
		Type: "vbox",
		W:    20,
		H:    10,
		Style: Style{
			Overflow: "scroll",
		},
	}
	child := &Node{
		Type:   "hbox",
		W:      50,
		H:      5,
		Parent: root,
		Style:  Style{},
	}
	root.Children = []*Node{child}

	// Simulate layout setting ScrollWidth
	// (In real code, layout_vbox.go computes this)
	root.ScrollWidth = 50

	// Verify computeMaxScrollX
	maxSX := computeMaxScrollX(root)
	expectedMax := 50 - 20 // scrollWidth - contentW (no border/padding)
	if maxSX != expectedMax {
		t.Errorf("computeMaxScrollX = %d, want %d", maxSX, expectedMax)
	}
}

func TestHorizontalScroll_NowrapTextExpandsInScrollContainer(t *testing.T) {
	// Verify that layout_vbox gives nowrap text its intrinsic width in scroll containers
	root := &Node{
		Type: "vbox",
		W:    20,
		H:    10,
		Style: Style{
			Overflow: "scroll",
		},
	}
	// A text node with content wider than container (30 chars > 20 cols)
	textNode := &Node{
		Type:    "text",
		Content: "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123",
		Parent:  root,
		Style: Style{
			WhiteSpace: "nowrap",
		},
	}
	root.Children = []*Node{textNode}

	// Run layout
	computeFlex(root, 0, 0, 20, 10, 0)

	// The text node should have W > 20 (its intrinsic width = 30)
	if textNode.W <= 20 {
		t.Errorf("nowrap text W = %d, want > 20 (intrinsic width should expand in scroll container)", textNode.W)
	}
	if textNode.W != 30 {
		t.Errorf("nowrap text W = %d, want 30 (intrinsic width of 30-char string)", textNode.W)
	}

	// ScrollWidth should be >= 30
	if root.ScrollWidth < 30 {
		t.Errorf("root.ScrollWidth = %d, want >= 30", root.ScrollWidth)
	}
}

func TestHorizontalScroll_HandleScrollH(t *testing.T) {
	e := &Engine{}
	e.layers = append(e.layers, &Layer{})

	// Build a scroll container
	root := &Node{
		Type: "vbox",
		X:    0, Y: 0, W: 20, H: 10,
		Style: Style{
			Overflow: "scroll",
		},
		ScrollWidth: 50,
	}
	child := &Node{
		Type:   "text",
		X:      0, Y: 0, W: 50, H: 5,
		Parent: root,
	}
	root.Children = []*Node{child}
	e.layers[0].Root = root

	// Initial state
	if root.ScrollX != 0 {
		t.Fatalf("initial ScrollX = %d, want 0", root.ScrollX)
	}

	// Scroll right (positive delta)
	e.HandleScrollH(5, 5, 1)
	if root.ScrollX <= 0 {
		t.Errorf("after scroll right: ScrollX = %d, want > 0", root.ScrollX)
	}
	savedX := root.ScrollX

	// Scroll left (negative delta)
	e.HandleScrollH(5, 5, -1)
	if root.ScrollX >= savedX {
		t.Errorf("after scroll left: ScrollX = %d, want < %d", root.ScrollX, savedX)
	}
}

func TestHorizontalScroll_ClampToMax(t *testing.T) {
	e := &Engine{}
	e.layers = append(e.layers, &Layer{})

	root := &Node{
		Type: "vbox",
		X:    0, Y: 0, W: 20, H: 10,
		Style: Style{
			Overflow: "scroll",
		},
		ScrollWidth: 50,
	}
	child := &Node{
		Type:   "text",
		X:      0, Y: 0, W: 50, H: 5,
		Parent: root,
	}
	root.Children = []*Node{child}
	e.layers[0].Root = root

	// Scroll way past max
	for i := 0; i < 100; i++ {
		e.HandleScrollH(5, 5, 1)
	}

	maxSX := computeMaxScrollX(root)
	if root.ScrollX > maxSX {
		t.Errorf("ScrollX = %d, exceeds max %d", root.ScrollX, maxSX)
	}
	if root.ScrollX != maxSX {
		t.Errorf("ScrollX = %d, want max %d (should clamp)", root.ScrollX, maxSX)
	}
}

func TestHorizontalScroll_ClampToZero(t *testing.T) {
	e := &Engine{}
	e.layers = append(e.layers, &Layer{})

	root := &Node{
		Type: "vbox",
		X:    0, Y: 0, W: 20, H: 10,
		Style: Style{
			Overflow: "scroll",
		},
		ScrollWidth: 50,
	}
	child := &Node{
		Type:   "text",
		X:      0, Y: 0, W: 50, H: 5,
		Parent: root,
	}
	root.Children = []*Node{child}
	e.layers[0].Root = root

	// Scroll left past zero
	for i := 0; i < 100; i++ {
		e.HandleScrollH(5, 5, -1)
	}

	if root.ScrollX != 0 {
		t.Errorf("ScrollX = %d, want 0 (should clamp to zero)", root.ScrollX)
	}
}

func TestHorizontalScroll_PaintOffset(t *testing.T) {
	// Create a scroll container with a text child, scroll right, verify paint clips correctly
	root := &Node{
		Type: "vbox",
		X:    0, Y: 0, W: 20, H: 5,
		Style: Style{
			Overflow: "scroll",
		},
		ScrollWidth: 50,
		ScrollX:     10, // scrolled 10 columns right
	}
	textNode := &Node{
		Type:    "text",
		X:       0, Y: 0, W: 50, H: 1,
		Parent:  root,
		Content: "ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890123456789ABCD",
	}
	root.Children = []*Node{textNode}

	buf := NewCellBuffer(20, 5)
	PaintFullV3(buf, root)

	// After scrolling 10 right, the first visible char should be 'K' (index 10)
	cell := buf.Get(0, 0)
	if cell.Ch != 'K' {
		t.Errorf("after scrollX=10, cell(0,0) = '%c', want 'K'", cell.Ch)
	}

	// Last visible char at col 19 should be at original position 29 = '4' (index 29 in the content)
	// Content: ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890123456789ABCD
	// Index 29 = '3' (A=0..Z=25, 0=26, 1=27, 2=28, 3=29)
	cell19 := buf.Get(19, 0)
	if cell19.Ch != '3' {
		t.Errorf("after scrollX=10, cell(19,0) = '%c', want '3'", cell19.Ch)
	}
}

func TestHorizontalScroll_ScrollNodeByIDH(t *testing.T) {
	e := &Engine{}
	e.layers = append(e.layers, &Layer{})

	root := &Node{
		Type: "vbox",
		ID:   "scroll-box",
		X:    0, Y: 0, W: 20, H: 10,
		Style: Style{
			Overflow: "scroll",
		},
		ScrollWidth: 50,
	}
	child := &Node{
		Type:   "text",
		X:      0, Y: 0, W: 50, H: 5,
		Parent: root,
	}
	root.Children = []*Node{child}
	e.layers[0].Root = root

	// Scroll right by 5
	newSX := e.ScrollNodeByIDH("scroll-box", 5)
	if newSX != 5 {
		t.Errorf("ScrollNodeByIDH returned %d, want 5", newSX)
	}
	if root.ScrollX != 5 {
		t.Errorf("root.ScrollX = %d, want 5", root.ScrollX)
	}

	// Scroll left by 3
	newSX = e.ScrollNodeByIDH("scroll-box", -3)
	if newSX != 2 {
		t.Errorf("ScrollNodeByIDH returned %d, want 2", newSX)
	}
}
