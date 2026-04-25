package lumina

import (
	"testing"
)

// TestCompositor_SingleOverlay verifies an overlay is composited onto the base frame.
func TestCompositor_SingleOverlay(t *testing.T) {
	// Base frame: 10x5 with 'B' at (0,0)
	base := NewFrame(10, 5)
	base.Cells[0][0] = Cell{Char: 'B', Foreground: "#ffffff"}
	base.MarkDirty()

	// Overlay: 3x1 with "Hi!" at position (2,1)
	ovNode := NewVNode("hbox")
	textChild := NewVNode("text")
	textChild.Content = "Hi!"
	ovNode.AddChild(textChild)

	ov := &Overlay{
		ID:      "test-ov",
		VNode:   ovNode,
		X:       2,
		Y:       1,
		W:       3,
		H:       1,
		ZIndex:  10,
		Visible: true,
	}

	compositor := NewCompositor(10, 5)
	result := compositor.Compose(base, []*Overlay{ov})

	// Base cell should still be there
	if result.Cells[0][0].Char != 'B' {
		t.Errorf("base cell at (0,0) should be 'B', got '%c'", result.Cells[0][0].Char)
	}

	// Overlay text should appear at (2,1)
	if result.Cells[1][2].Char != 'H' {
		t.Errorf("overlay cell at (2,1) should be 'H', got '%c'", result.Cells[1][2].Char)
	}
	if result.Cells[1][3].Char != 'i' {
		t.Errorf("overlay cell at (3,1) should be 'i', got '%c'", result.Cells[1][3].Char)
	}
	if result.Cells[1][4].Char != '!' {
		t.Errorf("overlay cell at (4,1) should be '!', got '%c'", result.Cells[1][4].Char)
	}
}

// TestCompositor_TransparencyPreservesBase verifies that transparent cells
// in an overlay don't overwrite the base layer.
func TestCompositor_TransparencyPreservesBase(t *testing.T) {
	// Base frame: fill row 0 with 'X'
	base := NewFrame(10, 3)
	for x := 0; x < 10; x++ {
		base.Cells[0][x] = Cell{Char: 'X', Foreground: "#ff0000"}
	}

	// Create overlay frame manually (to test compositeFrame directly)
	ovFrame := NewFrame(5, 1)
	// Only write to positions 1 and 3, leaving 0, 2, 4 transparent
	ovFrame.Cells[0][1] = Cell{Char: 'A', Transparent: false}
	ovFrame.Cells[0][3] = Cell{Char: 'B', Transparent: false}

	// Composite at (2, 0) — overlay spans base columns 2-6
	compositeFrame(base, ovFrame, 2, 0)

	// Position 2 (overlay x=0): transparent → base 'X' preserved
	if base.Cells[0][2].Char != 'X' {
		t.Errorf("transparent overlay cell should preserve base 'X' at (2,0), got '%c'", base.Cells[0][2].Char)
	}

	// Position 3 (overlay x=1): opaque 'A' → overwrites base
	if base.Cells[0][3].Char != 'A' {
		t.Errorf("opaque overlay cell should write 'A' at (3,0), got '%c'", base.Cells[0][3].Char)
	}

	// Position 4 (overlay x=2): transparent → base 'X' preserved
	if base.Cells[0][4].Char != 'X' {
		t.Errorf("transparent overlay cell should preserve base 'X' at (4,0), got '%c'", base.Cells[0][4].Char)
	}

	// Position 5 (overlay x=3): opaque 'B' → overwrites base
	if base.Cells[0][5].Char != 'B' {
		t.Errorf("opaque overlay cell should write 'B' at (5,0), got '%c'", base.Cells[0][5].Char)
	}

	// Position 6 (overlay x=4): transparent → base 'X' preserved
	if base.Cells[0][6].Char != 'X' {
		t.Errorf("transparent overlay cell should preserve base 'X' at (6,0), got '%c'", base.Cells[0][6].Char)
	}
}

// TestCompositor_ZIndexOrder verifies that higher z-index overlays overwrite lower ones.
func TestCompositor_ZIndexOrder(t *testing.T) {
	base := NewFrame(10, 3)

	// Two overlays at the same position, different z-index
	// Lower z-index writes 'L', higher writes 'H'
	lowNode := NewVNode("text")
	lowNode.Content = "LLL"

	highNode := NewVNode("text")
	highNode.Content = "HHH"

	overlays := []*Overlay{
		{
			ID: "low", VNode: lowNode,
			X: 0, Y: 0, W: 3, H: 1,
			ZIndex: 1, Visible: true,
		},
		{
			ID: "high", VNode: highNode,
			X: 0, Y: 0, W: 3, H: 1,
			ZIndex: 10, Visible: true,
		},
	}

	compositor := NewCompositor(10, 3)
	result := compositor.Compose(base, overlays)

	// Higher z-index should win
	if result.Cells[0][0].Char != 'H' {
		t.Errorf("higher z-index overlay should win, got '%c' instead of 'H'", result.Cells[0][0].Char)
	}
	if result.Cells[0][1].Char != 'H' {
		t.Errorf("higher z-index overlay should win, got '%c' instead of 'H'", result.Cells[0][1].Char)
	}
}

// TestCompositor_ModalDimsBase verifies that a modal overlay dims the base layer.
func TestCompositor_ModalDimsBase(t *testing.T) {
	base := NewFrame(10, 3)
	base.Cells[0][0] = Cell{Char: 'A', Foreground: "#ffffff", Dim: false}
	base.Cells[1][5] = Cell{Char: 'B', Foreground: "#ffffff", Dim: false}

	modalNode := NewVNode("text")
	modalNode.Content = "M"

	overlays := []*Overlay{
		{
			ID: "modal", VNode: modalNode,
			X: 4, Y: 1, W: 1, H: 1,
			ZIndex: 100, Visible: true, Modal: true,
		},
	}

	compositor := NewCompositor(10, 3)
	result := compositor.Compose(base, overlays)

	// Base cells should be dimmed
	if !result.Cells[0][0].Dim {
		t.Error("base cell at (0,0) should be dimmed by modal overlay")
	}
	if !result.Cells[1][5].Dim {
		t.Error("base cell at (5,1) should be dimmed by modal overlay")
	}

	// Modal content should be present
	if result.Cells[1][4].Char != 'M' {
		t.Errorf("modal overlay content should be at (4,1), got '%c'", result.Cells[1][4].Char)
	}
}

// TestCompositor_OverlayClipped verifies overlays at screen edges are clipped.
func TestCompositor_OverlayClipped(t *testing.T) {
	base := NewFrame(10, 5)

	// Create overlay frame manually at edge of screen
	ovFrame := NewFrame(5, 3)
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			ovFrame.Cells[y][x] = Cell{Char: 'O', Transparent: false}
		}
	}

	// Place at (8, 3) — only 2 columns and 2 rows visible
	compositeFrame(base, ovFrame, 8, 3)

	// (8,3) and (9,3) should have 'O'
	if base.Cells[3][8].Char != 'O' {
		t.Errorf("expected 'O' at (8,3), got '%c'", base.Cells[3][8].Char)
	}
	if base.Cells[3][9].Char != 'O' {
		t.Errorf("expected 'O' at (9,3), got '%c'", base.Cells[3][9].Char)
	}
	// (8,4) and (9,4) should have 'O'
	if base.Cells[4][8].Char != 'O' {
		t.Errorf("expected 'O' at (8,4), got '%c'", base.Cells[4][8].Char)
	}
	if base.Cells[4][9].Char != 'O' {
		t.Errorf("expected 'O' at (9,4), got '%c'", base.Cells[4][9].Char)
	}

	// Cells outside the frame should be unchanged (transparent space)
	if base.Cells[0][0].Char != ' ' || !base.Cells[0][0].Transparent {
		t.Error("cells outside overlay should be unchanged")
	}
}

// TestCompositeFrame_NegativeOffset verifies overlays with negative offsets are clipped.
func TestCompositeFrame_NegativeOffset(t *testing.T) {
	base := NewFrame(10, 5)

	ovFrame := NewFrame(5, 3)
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			ovFrame.Cells[y][x] = Cell{Char: rune('A' + x), Transparent: false}
		}
	}

	// Place at (-2, -1) — only partial overlap
	compositeFrame(base, ovFrame, -2, -1)

	// overlay x=2,y=1 maps to base (0,0)
	if base.Cells[0][0].Char != 'C' {
		t.Errorf("expected 'C' at (0,0) from negative offset overlay, got '%c'", base.Cells[0][0].Char)
	}
	// overlay x=3,y=1 maps to base (1,0)
	if base.Cells[0][1].Char != 'D' {
		t.Errorf("expected 'D' at (1,0) from negative offset overlay, got '%c'", base.Cells[0][1].Char)
	}
	// overlay x=4,y=1 maps to base (2,0)
	if base.Cells[0][2].Char != 'E' {
		t.Errorf("expected 'E' at (2,0) from negative offset overlay, got '%c'", base.Cells[0][2].Char)
	}
}
