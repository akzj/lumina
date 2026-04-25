package lumina

import (
	"bytes"
	"strings"
	"testing"
)

// TestDoubleBuffer_FirstRenderFull verifies that the first Write() call
// produces a full-screen write (since there is no previous frame to diff against).
func TestDoubleBuffer_FirstRenderFull(t *testing.T) {
	buf := &bytes.Buffer{}
	adapter := NewANSIAdapter(buf)

	frame := NewFrame(10, 3)
	frame.Cells[0][0] = Cell{Char: 'A', Foreground: "#ff0000"}
	frame.Cells[1][5] = Cell{Char: 'B', Foreground: "#00ff00"}
	frame.Cells[2][9] = Cell{Char: 'C', Foreground: "#0000ff"}
	frame.MarkDirty()

	err := adapter.Write(frame)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Should contain all three characters
	if !strings.Contains(output, "A") {
		t.Error("first render should contain 'A'")
	}
	if !strings.Contains(output, "B") {
		t.Error("first render should contain 'B'")
	}
	if !strings.Contains(output, "C") {
		t.Error("first render should contain 'C'")
	}

	// Should contain cursor-hide escape
	if !strings.Contains(output, "\x1b[?25l") {
		t.Error("first render should hide cursor")
	}

	// Should have reasonable output length (full frame write)
	if len(output) < 50 {
		t.Errorf("first render output too short (%d bytes), expected full frame", len(output))
	}
}

// TestDoubleBuffer_UnchangedSkipped verifies that writing the same frame twice
// produces minimal or no output on the second write (only cursor-hide + position).
func TestDoubleBuffer_UnchangedSkipped(t *testing.T) {
	buf := &bytes.Buffer{}
	adapter := NewANSIAdapter(buf)

	frame := NewFrame(10, 3)
	frame.Cells[1][5] = Cell{Char: 'X', Foreground: "#ffffff"}
	frame.MarkDirty()

	// First write — full render
	err := adapter.Write(frame)
	if err != nil {
		t.Fatalf("first Write failed: %v", err)
	}
	firstLen := buf.Len()

	// Second write — same frame, should produce minimal output
	buf.Reset()
	frame2 := NewFrame(10, 3)
	frame2.Cells[1][5] = Cell{Char: 'X', Foreground: "#ffffff"}
	frame2.MarkDirty()

	err = adapter.Write(frame2)
	if err != nil {
		t.Fatalf("second Write failed: %v", err)
	}
	secondLen := buf.Len()

	// Second write should be much smaller (only cursor hide + position, no cell data)
	if secondLen >= firstLen {
		t.Errorf("second write (%d bytes) should be smaller than first (%d bytes)", secondLen, firstLen)
	}

	// Second write should NOT contain 'X' (cell unchanged)
	output := buf.String()
	if strings.Contains(output, "X") {
		t.Error("second write should not re-write unchanged cell 'X'")
	}
}

// TestDoubleBuffer_SingleCellChange verifies that changing one cell produces
// minimal output (just that cell's escape codes + character).
func TestDoubleBuffer_SingleCellChange(t *testing.T) {
	buf := &bytes.Buffer{}
	adapter := NewANSIAdapter(buf)

	// First frame
	frame1 := NewFrame(10, 3)
	frame1.Cells[0][0] = Cell{Char: 'A'}
	frame1.Cells[1][5] = Cell{Char: 'B'}
	frame1.MarkDirty()
	adapter.Write(frame1)

	// Second frame — change only cell (1,5) from 'B' to 'Z'
	buf.Reset()
	frame2 := NewFrame(10, 3)
	frame2.Cells[0][0] = Cell{Char: 'A'} // unchanged
	frame2.Cells[1][5] = Cell{Char: 'Z'} // changed!
	frame2.MarkDirty()

	err := adapter.Write(frame2)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Should contain the changed character
	if !strings.Contains(output, "Z") {
		t.Error("diff write should contain changed cell 'Z'")
	}

	// Should NOT contain the unchanged character
	if strings.Contains(output, "A") {
		t.Error("diff write should not contain unchanged cell 'A'")
	}

	// Should NOT contain the old character
	if strings.Contains(output, "B") {
		t.Error("diff write should not contain old cell 'B'")
	}
}

// TestDoubleBuffer_ResizeFullRewrite verifies that when the frame dimensions change,
// a full rewrite is performed (not a diff).
func TestDoubleBuffer_ResizeFullRewrite(t *testing.T) {
	buf := &bytes.Buffer{}
	adapter := NewANSIAdapter(buf)

	// First frame: 10x3
	frame1 := NewFrame(10, 3)
	frame1.Cells[0][0] = Cell{Char: 'A'}
	frame1.MarkDirty()
	adapter.Write(frame1)

	// Second frame: different size (20x5)
	buf.Reset()
	frame2 := NewFrame(20, 5)
	frame2.Cells[0][0] = Cell{Char: 'A'} // same char at same position
	frame2.Cells[4][19] = Cell{Char: 'Z'}
	frame2.MarkDirty()

	err := adapter.Write(frame2)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Should contain both characters (full rewrite, not diff)
	if !strings.Contains(output, "A") {
		t.Error("resize rewrite should contain 'A'")
	}
	if !strings.Contains(output, "Z") {
		t.Error("resize rewrite should contain 'Z'")
	}

	// Output should be substantial (full frame, not just diff)
	if len(output) < 100 {
		t.Errorf("resize rewrite output too short (%d bytes), expected full frame", len(output))
	}
}

// TestCellEqual verifies the cell comparison function.
func TestCellEqual(t *testing.T) {
	a := Cell{Char: 'A', Foreground: "#ff0000", Background: "#000000", Bold: true}
	b := Cell{Char: 'A', Foreground: "#ff0000", Background: "#000000", Bold: true}

	if !cellEqual(a, b) {
		t.Error("identical cells should be equal")
	}

	// Different char
	c := b
	c.Char = 'B'
	if cellEqual(a, c) {
		t.Error("cells with different Char should not be equal")
	}

	// Different foreground
	d := b
	d.Foreground = "#00ff00"
	if cellEqual(a, d) {
		t.Error("cells with different Foreground should not be equal")
	}

	// Different background
	e := b
	e.Background = "#ffffff"
	if cellEqual(a, e) {
		t.Error("cells with different Background should not be equal")
	}

	// Different bold
	f := b
	f.Bold = false
	if cellEqual(a, f) {
		t.Error("cells with different Bold should not be equal")
	}

	// Different dim
	g := b
	g.Dim = true
	if cellEqual(a, g) {
		t.Error("cells with different Dim should not be equal")
	}

	// Different underline
	h := b
	h.Underline = true
	if cellEqual(a, h) {
		t.Error("cells with different Underline should not be equal")
	}

	// OwnerNode difference should NOT affect equality (visual comparison only)
	i := b
	i.OwnerNode = &VNode{Type: "test"}
	if !cellEqual(a, i) {
		t.Error("cells differing only in OwnerNode should be equal (visual comparison)")
	}
}

// TestFrameClone verifies Frame.Clone produces an independent deep copy.
func TestFrameClone(t *testing.T) {
	original := NewFrame(5, 3)
	original.Cells[0][0] = Cell{Char: 'A', Foreground: "#ff0000", OwnerNode: &VNode{Type: "test"}}
	original.Cells[1][2] = Cell{Char: 'B', Background: "#00ff00"}

	clone := original.Clone()

	// Same dimensions
	if clone.Width != original.Width || clone.Height != original.Height {
		t.Errorf("clone dimensions %dx%d != original %dx%d",
			clone.Width, clone.Height, original.Width, original.Height)
	}

	// Same cell content
	if clone.Cells[0][0].Char != 'A' {
		t.Error("clone should have same cell content")
	}
	if clone.Cells[1][2].Char != 'B' {
		t.Error("clone should have same cell content")
	}

	// OwnerNode should be nil in clone (prevent memory leaks)
	if clone.Cells[0][0].OwnerNode != nil {
		t.Error("clone should nil out OwnerNode pointers")
	}

	// Modifying clone should not affect original
	clone.Cells[0][0].Char = 'Z'
	if original.Cells[0][0].Char != 'A' {
		t.Error("modifying clone should not affect original")
	}
}
