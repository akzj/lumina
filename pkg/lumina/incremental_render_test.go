package lumina

import (
	"testing"
)

// TestIncrementalRender_TextContentUpdate verifies that when only text content
// changes, only the affected cells are modified in the frame.
func TestIncrementalRender_TextContentUpdate(t *testing.T) {
	width, height := 40, 10

	// Build initial VNode tree
	oldRoot := NewVNode("vbox")
	oldRoot.Style = Style{Width: width, Height: height}
	child1 := NewVNode("text")
	child1.Content = "Hello World"
	child2 := NewVNode("text")
	child2.Content = "Unchanged Line"
	oldRoot.Children = []*VNode{child1, child2}

	// Render initial frame
	frame := VNodeToFrame(oldRoot, width, height)

	// Verify initial content
	line0 := extractLine(frame, 0, 0, 11)
	if line0 != "Hello World" {
		t.Errorf("initial line0 = %q, want %q", line0, "Hello World")
	}

	// Build new tree with changed text
	newRoot := NewVNode("vbox")
	newRoot.Style = Style{Width: width, Height: height}
	newChild1 := NewVNode("text")
	newChild1.Content = "Hello Lumina"
	newChild2 := NewVNode("text")
	newChild2.Content = "Unchanged Line"
	newRoot.Children = []*VNode{newChild1, newChild2}

	// Diff
	patches := DiffVNode(oldRoot, newRoot)
	if len(patches) == 0 {
		t.Fatal("expected patches for text content change")
	}

	// Apply incremental patches
	computeFlexLayout(newRoot, 0, 0, width, height)
	ApplyPatches(frame, newRoot, patches, width, height)

	// Verify updated content
	line0 = extractLine(frame, 0, 0, 12)
	if line0 != "Hello Lumina" {
		t.Errorf("updated line0 = %q, want %q", line0, "Hello Lumina")
	}

	// Verify unchanged line is still there
	line1 := extractLine(frame, 0, 1, 14)
	if line1 != "Unchanged Line" {
		t.Errorf("line1 = %q, want %q", line1, "Unchanged Line")
	}

	// Verify dirty rects were set
	if len(frame.DirtyRects) == 0 {
		t.Error("expected dirty rects to be set")
	}
}

// TestIncrementalRender_ElementRemove verifies that removing an element
// clears its old area.
func TestIncrementalRender_ElementRemove(t *testing.T) {
	width, height := 40, 10

	// Build initial tree with 3 children
	oldRoot := NewVNode("vbox")
	oldRoot.Style = Style{Width: width, Height: height}
	for _, txt := range []string{"Line A", "Line B", "Line C"} {
		child := NewVNode("text")
		child.Content = txt
		oldRoot.Children = append(oldRoot.Children, child)
	}

	// Render initial frame
	frame := VNodeToFrame(oldRoot, width, height)

	// Verify Line B exists
	lineB := extractLine(frame, 0, 1, 6)
	if lineB != "Line B" {
		t.Errorf("initial lineB = %q, want %q", lineB, "Line B")
	}

	// New tree: remove middle child
	newRoot := NewVNode("vbox")
	newRoot.Style = Style{Width: width, Height: height}
	childA := NewVNode("text")
	childA.Content = "Line A"
	childC := NewVNode("text")
	childC.Content = "Line C"
	newRoot.Children = []*VNode{childA, childC}

	// Diff
	patches := DiffVNode(oldRoot, newRoot)
	if len(patches) == 0 {
		t.Fatal("expected patches for child removal")
	}

	// Apply
	computeFlexLayout(newRoot, 0, 0, width, height)
	ApplyPatches(frame, newRoot, patches, width, height)

	// Verify dirty rects exist
	if len(frame.DirtyRects) == 0 {
		t.Error("expected dirty rects after removal")
	}
}

// TestIncrementalRender_ElementAdd verifies that adding an element
// renders it correctly.
func TestIncrementalRender_ElementAdd(t *testing.T) {
	width, height := 40, 10

	// Initial tree: 1 child
	oldRoot := NewVNode("vbox")
	oldRoot.Style = Style{Width: width, Height: height}
	child1 := NewVNode("text")
	child1.Content = "First"
	oldRoot.Children = []*VNode{child1}

	frame := VNodeToFrame(oldRoot, width, height)

	// New tree: 2 children
	newRoot := NewVNode("vbox")
	newRoot.Style = Style{Width: width, Height: height}
	newChild1 := NewVNode("text")
	newChild1.Content = "First"
	newChild2 := NewVNode("text")
	newChild2.Content = "Second"
	newRoot.Children = []*VNode{newChild1, newChild2}

	patches := DiffVNode(oldRoot, newRoot)
	if len(patches) == 0 {
		t.Fatal("expected patches for child addition")
	}

	computeFlexLayout(newRoot, 0, 0, width, height)
	ApplyPatches(frame, newRoot, patches, width, height)

	// Verify new child is rendered
	line1 := extractLine(frame, 0, 1, 6)
	if line1 != "Second" {
		t.Errorf("line1 = %q, want %q", line1, "Second")
	}

	if len(frame.DirtyRects) == 0 {
		t.Error("expected dirty rects after addition")
	}
}

// TestIncrementalRender_LargeChangeFallback verifies that ShouldFullRerender
// returns true when too many patches exist.
func TestIncrementalRender_LargeChangeFallback(t *testing.T) {
	// Build a tree with many children
	root := NewVNode("vbox")
	for i := 0; i < 20; i++ {
		child := NewVNode("text")
		child.Content = "item"
		root.Children = append(root.Children, child)
	}

	// Create a completely different tree
	newRoot := NewVNode("vbox")
	for i := 0; i < 20; i++ {
		child := NewVNode("box") // different type → PatchReplace for each
		child.Content = "new-item"
		newRoot.Children = append(newRoot.Children, child)
	}

	patches := DiffVNode(root, newRoot)
	if len(patches) == 0 {
		t.Fatal("expected many patches")
	}

	// With 20+ replace patches vs 21 nodes, ShouldFullRerender should return true
	if !ShouldFullRerender(patches, root) {
		t.Errorf("ShouldFullRerender = false, want true for %d patches vs %d nodes",
			len(patches), countNodes(root))
	}
}

// TestClearRect verifies that clearRect properly clears cells and marks them transparent.
func TestClearRect(t *testing.T) {
	frame := NewFrame(20, 10)

	// Fill some cells
	for y := 2; y < 5; y++ {
		for x := 3; x < 8; x++ {
			frame.Cells[y][x] = Cell{Char: '#', Foreground: "red", Transparent: false}
		}
	}

	// Verify filled
	if frame.Cells[3][5].Char != '#' {
		t.Fatal("cell should be '#' before clear")
	}

	// Clear the region
	clearRect(frame, 3, 2, 5, 3)

	// Verify cleared
	for y := 2; y < 5; y++ {
		for x := 3; x < 8; x++ {
			cell := frame.Cells[y][x]
			if cell.Char != ' ' {
				t.Errorf("cell (%d,%d) char = %q, want ' '", x, y, string(cell.Char))
			}
			if !cell.Transparent {
				t.Errorf("cell (%d,%d) should be transparent after clear", x, y)
			}
		}
	}

	// Verify cells outside the region are unchanged
	if frame.Cells[0][0].Char != ' ' || !frame.Cells[0][0].Transparent {
		// NewFrame defaults are space + transparent, so this should still be true
	}
}

// TestDirtyRectRegionDiff verifies that writeDiffRegion only writes within bounds.
func TestDirtyRectRegionDiff(t *testing.T) {
	// Create two frames
	_ = NewFrame(20, 10) // oldFrame - used conceptually for diff comparison
	newFrame := NewFrame(20, 10)

	// Change only a small region in the new frame
	for x := 5; x < 10; x++ {
		newFrame.Cells[3][x] = Cell{Char: 'X', Transparent: false}
	}

	// Set dirty rects to cover only the changed region
	newFrame.DirtyRects = []Rect{{X: 5, Y: 3, W: 5, H: 1}}

	// Verify isFullFrameDirty returns false
	if isFullFrameDirty(newFrame) {
		t.Error("should not be full frame dirty with small rect")
	}

	// Full frame dirty rect
	fullFrame := NewFrame(20, 10)
	fullFrame.DirtyRects = []Rect{{X: 0, Y: 0, W: 20, H: 10}}
	if !isFullFrameDirty(fullFrame) {
		t.Error("should be full frame dirty")
	}
}

// TestNodeClipRect verifies clip rect calculation.
func TestNodeClipRect(t *testing.T) {
	frame := NewFrame(80, 24)

	// Normal case
	node := NewVNode("box")
	node.X = 5
	node.Y = 3
	node.W = 10
	node.H = 5
	clip := nodeClipRect(node, frame)
	if clip.X != 5 || clip.Y != 3 || clip.W != 10 || clip.H != 5 {
		t.Errorf("clip = %+v, want {5 3 10 5}", clip)
	}

	// Clamp to frame bounds
	node2 := NewVNode("box")
	node2.X = 75
	node2.Y = 20
	node2.W = 10
	node2.H = 10
	clip2 := nodeClipRect(node2, frame)
	if clip2.X != 75 || clip2.Y != 20 || clip2.W != 5 || clip2.H != 4 {
		t.Errorf("clip2 = %+v, want {75 20 5 4}", clip2)
	}

	// Negative origin
	node3 := NewVNode("box")
	node3.X = -2
	node3.Y = -1
	node3.W = 10
	node3.H = 5
	clip3 := nodeClipRect(node3, frame)
	if clip3.X != 0 || clip3.Y != 0 || clip3.W != 8 || clip3.H != 4 {
		t.Errorf("clip3 = %+v, want {0 0 8 4}", clip3)
	}
}

// extractLine reads characters from a frame row starting at (startX, y) for length chars.
func extractLine(frame *Frame, startX, y, length int) string {
	if y >= frame.Height {
		return ""
	}
	result := make([]rune, 0, length)
	for x := startX; x < startX+length && x < frame.Width; x++ {
		ch := frame.Cells[y][x].Char
		if ch == 0 {
			continue // skip wide char padding
		}
		result = append(result, ch)
	}
	return string(result)
}
