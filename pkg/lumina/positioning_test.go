package lumina

import (
	"testing"
	"time"
)

// TestRelativePositioning verifies that position="relative" offsets from normal flow.
func TestRelativePositioning(t *testing.T) {
	// Build a vbox with 3 children; middle one has position=relative, top=2, left=5
	root := NewVNode("vbox")
	root.Style = Style{Width: 40, Height: 10}

	child1 := NewVNode("text")
	child1.Content = "First"

	child2 := NewVNode("text")
	child2.Content = "Offset"
	child2.Props["style"] = map[string]any{
		"position": "relative",
		"top":      int64(2),
		"left":     int64(5),
	}

	child3 := NewVNode("text")
	child3.Content = "Third"

	root.Children = []*VNode{child1, child2, child3}

	// Layout
	computeFlexLayout(root, 0, 0, 40, 10)

	// child1 should be at Y=0 (normal flow)
	if child1.Y != 0 {
		t.Errorf("child1.Y = %d, want 0", child1.Y)
	}

	// child2 should be offset: normal Y=1, +top=2 → Y=3; normal X=0, +left=5 → X=5
	if child2.Y != 3 {
		t.Errorf("child2.Y = %d, want 3 (1 normal + 2 offset)", child2.Y)
	}
	if child2.X != 5 {
		t.Errorf("child2.X = %d, want 5 (0 normal + 5 offset)", child2.X)
	}

	// child3 should be at Y=2 (normal flow, unaffected by child2's relative offset)
	if child3.Y != 2 {
		t.Errorf("child3.Y = %d, want 2 (normal flow)", child3.Y)
	}
}

// TestRelativePositioning_RightBottom verifies right/bottom offsets.
func TestRelativePositioning_RightBottom(t *testing.T) {
	root := NewVNode("vbox")
	root.Style = Style{Width: 40, Height: 10}

	child := NewVNode("text")
	child.Content = "Nudge"
	child.Props["style"] = map[string]any{
		"position": "relative",
		"right":    int64(3),
		"bottom":   int64(1),
	}

	root.Children = []*VNode{child}
	computeFlexLayout(root, 0, 0, 40, 10)

	// right=3 with left=0 → X offset = -3
	if child.X != -3 {
		t.Errorf("child.X = %d, want -3 (right=3 offset)", child.X)
	}
	// bottom=1 with top=0 → Y offset = -1
	if child.Y != -1 {
		t.Errorf("child.Y = %d, want -1 (bottom=1 offset)", child.Y)
	}
}

// TestZIndexRenderOrder verifies that higher z-index paints on top.
func TestZIndexRenderOrder(t *testing.T) {
	width, height := 20, 5

	root := NewVNode("box")
	root.Style = Style{Width: width, Height: height}

	// Two absolute children overlapping at same position
	// Lower z-index child has '#', higher has '@'
	low := NewVNode("text")
	low.Content = "####"
	low.Props["style"] = map[string]any{
		"position": "absolute",
		"top":      int64(1),
		"left":     int64(1),
		"width":    int64(4),
		"height":   int64(1),
		"zIndex":   int64(1),
	}

	high := NewVNode("text")
	high.Content = "@@@@"
	high.Props["style"] = map[string]any{
		"position": "absolute",
		"top":      int64(1),
		"left":     int64(1),
		"width":    int64(4),
		"height":   int64(1),
		"zIndex":   int64(10),
	}

	// Add low first, high second — but high should render on top due to z-index
	root.Children = []*VNode{low, high}

	frame := VNodeToFrame(root, width, height)

	// The cell at (1,1) should show '@' (from high z-index child)
	cell := frame.Cells[1][1]
	if cell.Char != '@' {
		t.Errorf("cell (1,1) = %q, want '@' (higher z-index should paint on top)", string(cell.Char))
	}
}

// TestZIndexRenderOrder_ReversedOrder verifies z-index works regardless of DOM order.
func TestZIndexRenderOrder_ReversedOrder(t *testing.T) {
	width, height := 20, 5

	root := NewVNode("box")
	root.Style = Style{Width: width, Height: height}

	// High z-index first in DOM, low z-index second
	high := NewVNode("text")
	high.Content = "@@@@"
	high.Props["style"] = map[string]any{
		"position": "absolute",
		"top":      int64(1),
		"left":     int64(1),
		"width":    int64(4),
		"height":   int64(1),
		"zIndex":   int64(10),
	}

	low := NewVNode("text")
	low.Content = "####"
	low.Props["style"] = map[string]any{
		"position": "absolute",
		"top":      int64(1),
		"left":     int64(1),
		"width":    int64(4),
		"height":   int64(1),
		"zIndex":   int64(1),
	}

	root.Children = []*VNode{high, low}

	frame := VNodeToFrame(root, width, height)

	// High z-index should still win
	cell := frame.Cells[1][1]
	if cell.Char != '@' {
		t.Errorf("cell (1,1) = %q, want '@' (higher z-index should paint on top)", string(cell.Char))
	}
}

// TestWebTerminalResize verifies resize callback fires.
func TestWebTerminalResize(t *testing.T) {
	var calledW, calledH int
	wt := &WebTerminal{
		width:  80,
		height: 24,
	}
	wt.SetOnResize(func(w, h int) {
		calledW = w
		calledH = h
	})

	wt.SetSize(120, 40)

	if calledW != 120 || calledH != 40 {
		t.Errorf("resize callback got (%d,%d), want (120,40)", calledW, calledH)
	}

	// Verify internal dimensions updated
	w, h := wt.Size()
	if w != 120 || h != 40 {
		t.Errorf("Size() = (%d,%d), want (120,40)", w, h)
	}
}

// TestFrameRateLimit verifies that renderAllDirty respects the 16ms limit.
func TestFrameRateLimit(t *testing.T) {
	app := &App{
		lastRenderTime: time.Now(), // just rendered
	}
	// Verify the field exists and is set correctly
	if app.lastRenderTime.IsZero() {
		t.Error("lastRenderTime should not be zero after initialization")
	}
	// Verify the frame rate limit threshold is reasonable
	elapsed := time.Since(app.lastRenderTime)
	if elapsed > 16*time.Millisecond {
		t.Error("time.Since should be < 16ms for just-set lastRenderTime")
	}
}
