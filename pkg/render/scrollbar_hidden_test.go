package render

import "testing"

// TestScrollbarHidden_NoPaint verifies no scrollbar is painted when scrollbar="none".
func TestScrollbarHidden_NoPaint(t *testing.T) {
	// Create a scroll container with scrollbar="none"
	root := &Node{
		Type: "vbox",
		X:    0,
		Y:    0,
		W:    20,
		H:    10,
		Style: Style{
			Overflow:  "scroll",
			Scrollbar: "none",
			Border:    "single",
		},
		ScrollHeight: 50,
		ScrollY:      10,
	}
	child := &Node{
		Type:   "text",
		X:      1,
		Y:      1,
		W:      18,
		H:      50,
		Parent: root,
		Content: "Hello",
	}
	root.Children = []*Node{child}

	// Paint into a buffer
	buf := NewCellBuffer(20, 10)
	paintNode(buf, root)

	// The scrollbar column for a bordered 20-wide node is at x=18
	// (node.X + node.W - bw - paddingRight - 1 = 0 + 20 - 1 - 0 - 1 = 18)
	scrollbarX := 18

	// Check that no scrollbar characters (█ or ░) are painted at scrollbarX
	for y := 1; y < 9; y++ {
		cell := buf.Get(scrollbarX, y)
		if cell.Ch == '█' || cell.Ch == '░' {
			t.Errorf("scrollbar character found at (%d,%d): %c — should be hidden", scrollbarX, y, cell.Ch)
		}
	}
}

// TestScrollbarHidden_StillScrolls verifies scrolling still works with scrollbar="none".
func TestScrollbarHidden_StillScrolls(t *testing.T) {
	root := &Node{
		Type: "vbox",
		X:    0,
		Y:    0,
		W:    20,
		H:    10,
		Style: Style{
			Overflow:  "scroll",
			Scrollbar: "none",
			Border:    "single",
		},
		ScrollHeight: 50,
		ScrollY:      0,
	}
	child := &Node{
		Type:   "vbox",
		X:      1,
		Y:      1,
		W:      18,
		H:      50,
		Parent: root,
	}
	root.Children = []*Node{child}

	e := &Engine{
		layers: []*Layer{{Root: root}},
	}

	// Scroll down — should still work
	e.autoScroll(root, 1)
	if root.ScrollY == 0 {
		t.Error("ScrollY should have changed after autoScroll with scrollbar=none")
	}
	t.Logf("ScrollY after autoScroll: %d", root.ScrollY)
}

// TestScrollbarHidden_FullWidthForChildren verifies layout gives full contentW to children.
func TestScrollbarHidden_FullWidthForChildren(t *testing.T) {
	// With scrollbar="none", no column is reserved for the scrollbar.
	// A bordered vbox with W=20: contentW = 20 - 2*1 = 18, layoutW should be 18 (not 17).
	root := &Node{
		Type: "vbox",
		W:    20,
		H:    10,
		Style: Style{
			Overflow:  "scroll",
			Scrollbar: "none",
			Border:    "single",
		},
	}
	textNode := &Node{
		Type:    "text",
		Content: "Hello World",
		Parent:  root,
		Style:   Style{},
	}
	root.Children = []*Node{textNode}

	// Run layout
	computeFlex(root, 0, 0, 20, 10, 0)

	// Child should get full contentW = 18 (not 17 which would be if scrollbar was reserved)
	if textNode.W != 18 {
		t.Errorf("text node W = %d, want 18 (full content width, no scrollbar reserved)", textNode.W)
	}
}

// TestScrollbarHidden_ClickIgnored verifies scrollbar click is ignored when hidden.
func TestScrollbarHidden_ClickIgnored(t *testing.T) {
	root := &Node{
		Type: "vbox",
		X:    0,
		Y:    0,
		W:    20,
		H:    10,
		Style: Style{
			Overflow:  "scroll",
			Scrollbar: "none",
			Border:    "single",
		},
		ScrollHeight: 50,
		ScrollY:      0,
	}
	child := &Node{
		Type:   "vbox",
		X:      1,
		Y:      1,
		W:      18,
		H:      50,
		Parent: root,
	}
	root.Children = []*Node{child}

	e := &Engine{
		layers: []*Layer{{Root: root}},
	}

	// Click at scrollbar column — should be ignored
	scrollbarX := 18
	handled := e.handleScrollbarClick(root, scrollbarX, 5)
	if handled {
		t.Error("scrollbar click should be ignored when scrollbar=none")
	}

	// Drag should also not start
	started := e.handleScrollbarDragStart(scrollbarX, 5)
	if started {
		t.Error("scrollbar drag should not start when scrollbar=none")
	}
}
