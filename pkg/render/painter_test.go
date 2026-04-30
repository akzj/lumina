package render

import "testing"

// --- CellBuffer Tests ---

func TestCellBuffer_SetGet(t *testing.T) {
	buf := NewCellBuffer(10, 5)
	buf.Set(3, 2, Cell{Ch: 'A', FG: "#FF0000", BG: "#000000", Bold: true})

	c := buf.Get(3, 2)
	if c.Ch != 'A' {
		t.Errorf("expected 'A', got %q", c.Ch)
	}
	if c.FG != "#FF0000" {
		t.Errorf("expected FG=#FF0000, got %q", c.FG)
	}
	if c.BG != "#000000" {
		t.Errorf("expected BG=#000000, got %q", c.BG)
	}
	if !c.Bold {
		t.Error("expected Bold=true")
	}
}

func TestCellBuffer_OutOfBounds(t *testing.T) {
	buf := NewCellBuffer(5, 5)
	// Set out of bounds — should not panic
	buf.Set(-1, 0, Cell{Ch: 'X'})
	buf.Set(0, -1, Cell{Ch: 'X'})
	buf.Set(5, 0, Cell{Ch: 'X'})
	buf.Set(0, 5, Cell{Ch: 'X'})

	// Get out of bounds — should return zero Cell
	c := buf.Get(-1, 0)
	if c.Ch != 0 {
		t.Errorf("expected zero cell for out-of-bounds Get, got %q", c.Ch)
	}
	c = buf.Get(5, 0)
	if c.Ch != 0 {
		t.Errorf("expected zero cell for out-of-bounds Get, got %q", c.Ch)
	}
}

func TestCellBuffer_Clear(t *testing.T) {
	buf := NewCellBuffer(3, 3)
	buf.Set(1, 1, Cell{Ch: 'Z', FG: "red"})
	buf.Clear()

	c := buf.Get(1, 1)
	if c.Ch != 0 || c.FG != "" {
		t.Errorf("expected zero cell after Clear, got Ch=%q FG=%q", c.Ch, c.FG)
	}
}

func TestCellBuffer_ClearRect(t *testing.T) {
	buf := NewCellBuffer(5, 5)
	// Fill entire buffer
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			buf.Set(x, y, Cell{Ch: '#'})
		}
	}

	// Clear a 2x2 rect at (1,1)
	buf.ClearRect(1, 1, 2, 2)

	// Inside rect should be cleared
	for y := 1; y <= 2; y++ {
		for x := 1; x <= 2; x++ {
			c := buf.Get(x, y)
			if c.Ch != 0 {
				t.Errorf("expected cleared cell at (%d,%d), got %q", x, y, c.Ch)
			}
		}
	}

	// Outside rect should still have '#'
	c := buf.Get(0, 0)
	if c.Ch != '#' {
		t.Errorf("expected '#' at (0,0), got %q", c.Ch)
	}
	c = buf.Get(4, 4)
	if c.Ch != '#' {
		t.Errorf("expected '#' at (4,4), got %q", c.Ch)
	}
}

func TestCellBuffer_Resize(t *testing.T) {
	buf := NewCellBuffer(3, 3)
	buf.Set(0, 0, Cell{Ch: 'A'})
	buf.Set(2, 2, Cell{Ch: 'B'})

	// Grow
	buf.Resize(5, 5)
	if buf.Width() != 5 || buf.Height() != 5 {
		t.Errorf("expected 5x5, got %dx%d", buf.Width(), buf.Height())
	}
	c := buf.Get(0, 0)
	if c.Ch != 'A' {
		t.Errorf("expected 'A' preserved at (0,0), got %q", c.Ch)
	}
	c = buf.Get(2, 2)
	if c.Ch != 'B' {
		t.Errorf("expected 'B' preserved at (2,2), got %q", c.Ch)
	}

	// Shrink
	buf.Resize(2, 2)
	if buf.Width() != 2 || buf.Height() != 2 {
		t.Errorf("expected 2x2, got %dx%d", buf.Width(), buf.Height())
	}
	c = buf.Get(0, 0)
	if c.Ch != 'A' {
		t.Errorf("expected 'A' preserved at (0,0) after shrink, got %q", c.Ch)
	}
	// (2,2) is now out of bounds
	c = buf.Get(2, 2)
	if c.Ch != 0 {
		t.Errorf("expected zero cell for out-of-bounds after shrink, got %q", c.Ch)
	}
}

// --- Painter Tests ---

func TestPaintFull_TextNode(t *testing.T) {
	buf := NewCellBuffer(20, 5)
	node := &Node{
		Type:    "text",
		Content: "Hello",
		X: 2, Y: 1, W: 10, H: 1,
		Style: Style{Foreground: "#FFFFFF", Right: -1, Bottom: -1},
	}

	PaintFull(buf, node)

	expected := "Hello"
	for i, ch := range expected {
		c := buf.Get(2+i, 1)
		if c.Ch != ch {
			t.Errorf("at (%d,1): expected %q, got %q", 2+i, ch, c.Ch)
		}
		if c.FG != "#FFFFFF" {
			t.Errorf("at (%d,1): expected FG=#FFFFFF, got %q", 2+i, c.FG)
		}
	}

	// Cell before text should be empty
	c := buf.Get(1, 1)
	if c.Ch != 0 {
		t.Errorf("expected empty cell at (1,1), got %q", c.Ch)
	}
}

func TestPaintFull_BoxWithBackground(t *testing.T) {
	buf := NewCellBuffer(10, 5)
	node := &Node{
		Type: "box",
		X: 0, Y: 0, W: 5, H: 3,
		Style: Style{Background: "#1E1E2E", Right: -1, Bottom: -1},
	}

	PaintFull(buf, node)

	// All cells in the box region should have the background
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			c := buf.Get(x, y)
			if c.BG != "#1E1E2E" {
				t.Errorf("at (%d,%d): expected BG=#1E1E2E, got %q", x, y, c.BG)
			}
			if c.Ch != ' ' {
				t.Errorf("at (%d,%d): expected space for bg fill, got %q", x, y, c.Ch)
			}
		}
	}

	// Outside the box should be empty
	c := buf.Get(5, 0)
	if c.BG != "" {
		t.Errorf("expected empty BG outside box, got %q", c.BG)
	}
}

func TestPaintFull_Border(t *testing.T) {
	buf := NewCellBuffer(10, 5)
	node := &Node{
		Type: "box",
		X: 0, Y: 0, W: 5, H: 3,
		Style: Style{Border: "single", Right: -1, Bottom: -1},
	}

	PaintFull(buf, node)

	// Check corners
	if c := buf.Get(0, 0); c.Ch != '┌' {
		t.Errorf("top-left corner: expected '┌', got %q", c.Ch)
	}
	if c := buf.Get(4, 0); c.Ch != '┐' {
		t.Errorf("top-right corner: expected '┐', got %q", c.Ch)
	}
	if c := buf.Get(0, 2); c.Ch != '└' {
		t.Errorf("bottom-left corner: expected '└', got %q", c.Ch)
	}
	if c := buf.Get(4, 2); c.Ch != '┘' {
		t.Errorf("bottom-right corner: expected '┘', got %q", c.Ch)
	}

	// Top edge
	for x := 1; x < 4; x++ {
		if c := buf.Get(x, 0); c.Ch != '─' {
			t.Errorf("top edge at (%d,0): expected '─', got %q", x, c.Ch)
		}
	}

	// Left edge
	if c := buf.Get(0, 1); c.Ch != '│' {
		t.Errorf("left edge at (0,1): expected '│', got %q", c.Ch)
	}
}

func TestPaintFull_NestedBoxText(t *testing.T) {
	buf := NewCellBuffer(20, 10)
	parent := &Node{
		Type: "box",
		X: 0, Y: 0, W: 20, H: 10,
		Style: Style{Background: "#111111", Right: -1, Bottom: -1},
	}
	child := &Node{
		Type:    "text",
		Content: "Hi",
		X: 2, Y: 1, W: 5, H: 1,
		Style: Style{Foreground: "#FFFFFF", Right: -1, Bottom: -1},
	}
	parent.AddChild(child)

	PaintFull(buf, parent)

	// Parent background should be set
	c := buf.Get(0, 0)
	if c.BG != "#111111" {
		t.Errorf("expected parent BG at (0,0), got %q", c.BG)
	}

	// Text should be painted on top
	c = buf.Get(2, 1)
	if c.Ch != 'H' {
		t.Errorf("expected 'H' at (2,1), got %q", c.Ch)
	}
	c = buf.Get(3, 1)
	if c.Ch != 'i' {
		t.Errorf("expected 'i' at (3,1), got %q", c.Ch)
	}
}

func TestPaintDirty_OnlyDirtyNodes(t *testing.T) {
	buf := NewCellBuffer(20, 5)

	// Create two text nodes
	root := &Node{
		Type: "box",
		X: 0, Y: 0, W: 20, H: 5,
		Style: Style{Right: -1, Bottom: -1},
	}
	text1 := &Node{
		Type:    "text",
		Content: "AAA",
		X: 0, Y: 0, W: 10, H: 1,
		Style: Style{Foreground: "#FF0000", Right: -1, Bottom: -1},
	}
	text2 := &Node{
		Type:    "text",
		Content: "BBB",
		X: 0, Y: 1, W: 10, H: 1,
		Style: Style{Foreground: "#00FF00", Right: -1, Bottom: -1},
	}
	root.AddChild(text1)
	root.AddChild(text2)

	// Full paint first
	PaintFull(buf, root)

	// Now change text1 content and mark dirty
	text1.Content = "XXX"
	text1.PaintDirty = true

	// Paint dirty only
	PaintDirty(buf, root)

	// text1 should be updated
	c := buf.Get(0, 0)
	if c.Ch != 'X' {
		t.Errorf("expected 'X' at (0,0) after dirty paint, got %q", c.Ch)
	}

	// text2 should still be there (not cleared)
	c = buf.Get(0, 1)
	if c.Ch != 'B' {
		t.Errorf("expected 'B' at (0,1) preserved, got %q", c.Ch)
	}
}

func TestPaintDirty_ClearsFlag(t *testing.T) {
	buf := NewCellBuffer(10, 5)
	node := &Node{
		Type:       "text",
		Content:    "Test",
		X: 0, Y: 0, W: 10, H: 1,
		PaintDirty: true,
		Style:      Style{Right: -1, Bottom: -1},
	}

	PaintDirty(buf, node)

	if node.PaintDirty {
		t.Error("expected PaintDirty to be cleared after PaintDirty()")
	}
}

func TestPaintDirty_DeepDirty(t *testing.T) {
	buf := NewCellBuffer(20, 10)

	root := &Node{Type: "box", X: 0, Y: 0, W: 20, H: 10, Style: Style{Right: -1, Bottom: -1}}
	mid := &Node{Type: "box", X: 0, Y: 0, W: 20, H: 10, Style: Style{Right: -1, Bottom: -1}}
	leaf := &Node{
		Type:       "text",
		Content:    "Deep",
		X: 5, Y: 5, W: 10, H: 1,
		PaintDirty: true,
		Style:      Style{Foreground: "#AABBCC", Right: -1, Bottom: -1},
	}
	root.AddChild(mid)
	mid.AddChild(leaf)

	PaintDirty(buf, root)

	c := buf.Get(5, 5)
	if c.Ch != 'D' {
		t.Errorf("expected 'D' at (5,5), got %q", c.Ch)
	}
	if leaf.PaintDirty {
		t.Error("expected leaf PaintDirty cleared")
	}
}

func TestPaintDirty_ParentDirty(t *testing.T) {
	buf := NewCellBuffer(20, 10)

	parent := &Node{
		Type:       "box",
		X: 0, Y: 0, W: 20, H: 10,
		PaintDirty: true,
		Style:      Style{Background: "#222222", Right: -1, Bottom: -1},
	}
	child := &Node{
		Type:    "text",
		Content: "Child",
		X: 1, Y: 1, W: 10, H: 1,
		Style:   Style{Foreground: "#FFFFFF", Right: -1, Bottom: -1},
	}
	parent.AddChild(child)

	PaintDirty(buf, parent)

	// Parent bg should be painted
	c := buf.Get(0, 0)
	if c.BG != "#222222" {
		t.Errorf("expected parent BG at (0,0), got %q", c.BG)
	}

	// Child text should also be painted (parent dirty paints all children)
	c = buf.Get(1, 1)
	if c.Ch != 'C' {
		t.Errorf("expected 'C' at (1,1), got %q", c.Ch)
	}

	if parent.PaintDirty {
		t.Error("expected parent PaintDirty cleared")
	}
}

func TestPaintFull_VBoxChildren(t *testing.T) {
	buf := NewCellBuffer(20, 10)

	root := &Node{Type: "vbox", X: 0, Y: 0, W: 20, H: 10, Style: Style{Right: -1, Bottom: -1}}
	t1 := &Node{
		Type: "text", Content: "Line1",
		X: 0, Y: 0, W: 10, H: 1,
		Style: Style{Right: -1, Bottom: -1},
	}
	t2 := &Node{
		Type: "text", Content: "Line2",
		X: 0, Y: 1, W: 10, H: 1,
		Style: Style{Right: -1, Bottom: -1},
	}
	root.AddChild(t1)
	root.AddChild(t2)

	PaintFull(buf, root)

	c := buf.Get(0, 0)
	if c.Ch != 'L' {
		t.Errorf("expected 'L' at (0,0), got %q", c.Ch)
	}
	c = buf.Get(0, 1)
	if c.Ch != 'L' {
		t.Errorf("expected 'L' at (0,1), got %q", c.Ch)
	}
	c = buf.Get(4, 1)
	if c.Ch != '2' {
		t.Errorf("expected '2' at (4,1), got %q", c.Ch)
	}
}

func TestPaintFull_MultilineText(t *testing.T) {
	buf := NewCellBuffer(20, 10)
	node := &Node{
		Type:    "text",
		Content: "AB\nCD",
		X: 1, Y: 2, W: 10, H: 3,
		Style: Style{Right: -1, Bottom: -1},
	}

	PaintFull(buf, node)

	if c := buf.Get(1, 2); c.Ch != 'A' {
		t.Errorf("expected 'A' at (1,2), got %q", c.Ch)
	}
	if c := buf.Get(2, 2); c.Ch != 'B' {
		t.Errorf("expected 'B' at (2,2), got %q", c.Ch)
	}
	if c := buf.Get(1, 3); c.Ch != 'C' {
		t.Errorf("expected 'C' at (1,3), got %q", c.Ch)
	}
	if c := buf.Get(2, 3); c.Ch != 'D' {
		t.Errorf("expected 'D' at (2,3), got %q", c.Ch)
	}
}

func TestPaintFull_NilInputs(t *testing.T) {
	// Should not panic
	PaintFull(nil, nil)
	PaintFull(NewCellBuffer(5, 5), nil)
	PaintFull(nil, &Node{Type: "box"})

	PaintDirty(nil, nil)
	PaintDirty(NewCellBuffer(5, 5), nil)
	PaintDirty(nil, &Node{Type: "box"})
}

func TestPaintFull_RoundedBorder(t *testing.T) {
	buf := NewCellBuffer(10, 5)
	node := &Node{
		Type: "box",
		X: 0, Y: 0, W: 6, H: 4,
		Style: Style{Border: "rounded", Right: -1, Bottom: -1},
	}

	PaintFull(buf, node)

	if c := buf.Get(0, 0); c.Ch != '╭' {
		t.Errorf("top-left: expected '╭', got %q", c.Ch)
	}
	if c := buf.Get(5, 0); c.Ch != '╮' {
		t.Errorf("top-right: expected '╮', got %q", c.Ch)
	}
	if c := buf.Get(0, 3); c.Ch != '╰' {
		t.Errorf("bottom-left: expected '╰', got %q", c.Ch)
	}
	if c := buf.Get(5, 3); c.Ch != '╯' {
		t.Errorf("bottom-right: expected '╯', got %q", c.Ch)
	}
}

// Bordered descendants of overflow:scroll are painted via paintBoxClipped; borders
// must not be omitted (regression: Lux outlined buttons in scrollable main).
func TestPaintFull_ScrollVBox_DrawsNestedRoundedBorder(t *testing.T) {
	buf := NewCellBuffer(24, 10)
	inner := makeNode("hbox", Style{
		Border: "rounded", BorderColor: "#F5C842", Foreground: "#F5C842",
		Height: 3, PaddingLeft: 1, PaddingRight: 1, Right: -1, Bottom: -1,
	}, makeText("OK"))
	root := makeNode("vbox", Style{
		Overflow: "scroll", Background: "#0B1220", Right: -1, Bottom: -1,
	}, inner)
	LayoutFull(root, 0, 0, 24, 10)
	PaintFull(buf, root)

	found := false
	for py := 0; py < 10; py++ {
		for px := 0; px < 24; px++ {
			if buf.Get(px, py).Ch == '╭' {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected at least one '╭' from nested bordered hbox inside scroll vbox")
	}
}

func TestPaintFull_DoubleBorder(t *testing.T) {
	buf := NewCellBuffer(10, 5)
	node := &Node{
		Type: "box",
		X: 0, Y: 0, W: 4, H: 3,
		Style: Style{Border: "double", Right: -1, Bottom: -1},
	}

	PaintFull(buf, node)

	if c := buf.Get(0, 0); c.Ch != '╔' {
		t.Errorf("top-left: expected '╔', got %q", c.Ch)
	}
	if c := buf.Get(3, 0); c.Ch != '╗' {
		t.Errorf("top-right: expected '╗', got %q", c.Ch)
	}
	if c := buf.Get(1, 0); c.Ch != '═' {
		t.Errorf("top edge: expected '═', got %q", c.Ch)
	}
	if c := buf.Get(0, 1); c.Ch != '║' {
		t.Errorf("left edge: expected '║', got %q", c.Ch)
	}
}

func TestCellBuffer_SetChar(t *testing.T) {
	buf := NewCellBuffer(5, 5)
	buf.SetChar(2, 3, 'Q', "#AABB", "#CCDD", true)

	c := buf.Get(2, 3)
	if c.Ch != 'Q' || c.FG != "#AABB" || c.BG != "#CCDD" || !c.Bold {
		t.Errorf("SetChar mismatch: got %+v", c)
	}
}

// --- CellBuffer Stats Tests ---

func TestCellBuffer_Stats_ResetStats(t *testing.T) {
	buf := NewCellBuffer(10, 5)
	buf.SetChar(3, 2, 'A', "#FFF", "#000", false)
	buf.SetChar(4, 2, 'B', "#FFF", "#000", false)

	s := buf.Stats()
	if s.WriteCount != 2 {
		t.Errorf("WriteCount: got %d, want 2", s.WriteCount)
	}
	if s.DirtyX != 3 || s.DirtyY != 2 || s.DirtyW != 2 || s.DirtyH != 1 {
		t.Errorf("DirtyRect: got (%d,%d,%d,%d), want (3,2,2,1)", s.DirtyX, s.DirtyY, s.DirtyW, s.DirtyH)
	}

	buf.ResetStats()
	s2 := buf.Stats()
	if s2.WriteCount != 0 || s2.ClearCount != 0 {
		t.Errorf("after ResetStats: WriteCount=%d ClearCount=%d, want 0,0", s2.WriteCount, s2.ClearCount)
	}
	if s2.DirtyW != 0 || s2.DirtyH != 0 {
		t.Errorf("after ResetStats: DirtyRect should be zero, got (%d,%d,%d,%d)", s2.DirtyX, s2.DirtyY, s2.DirtyW, s2.DirtyH)
	}
}

func TestCellBuffer_Stats_Set(t *testing.T) {
	buf := NewCellBuffer(10, 5)
	buf.Set(0, 0, Cell{Ch: 'X'})
	buf.Set(9, 4, Cell{Ch: 'Y'})

	s := buf.Stats()
	if s.WriteCount != 2 {
		t.Errorf("WriteCount: got %d, want 2", s.WriteCount)
	}
	if s.DirtyX != 0 || s.DirtyY != 0 || s.DirtyW != 10 || s.DirtyH != 5 {
		t.Errorf("DirtyRect: got (%d,%d,%d,%d), want (0,0,10,5)", s.DirtyX, s.DirtyY, s.DirtyW, s.DirtyH)
	}
}

func TestCellBuffer_Stats_ClearRect(t *testing.T) {
	buf := NewCellBuffer(10, 5)
	buf.ClearRect(2, 1, 3, 2)

	s := buf.Stats()
	if s.ClearCount != 6 {
		t.Errorf("ClearCount: got %d, want 6", s.ClearCount)
	}
	if s.DirtyX != 2 || s.DirtyY != 1 || s.DirtyW != 3 || s.DirtyH != 2 {
		t.Errorf("DirtyRect: got (%d,%d,%d,%d), want (2,1,3,2)", s.DirtyX, s.DirtyY, s.DirtyW, s.DirtyH)
	}
}

func TestCellBuffer_Stats_Clear(t *testing.T) {
	buf := NewCellBuffer(4, 3)
	buf.Clear()

	s := buf.Stats()
	if s.ClearCount != 12 {
		t.Errorf("ClearCount: got %d, want 12", s.ClearCount)
	}
	if s.DirtyX != 0 || s.DirtyY != 0 || s.DirtyW != 4 || s.DirtyH != 3 {
		t.Errorf("DirtyRect: got (%d,%d,%d,%d), want (0,0,4,3)", s.DirtyX, s.DirtyY, s.DirtyW, s.DirtyH)
	}
}

func TestCellBuffer_Stats_OutOfBounds_NoTrack(t *testing.T) {
	buf := NewCellBuffer(5, 5)
	buf.Set(-1, -1, Cell{Ch: 'X'})
	buf.Set(10, 10, Cell{Ch: 'Y'})
	buf.SetChar(-1, 0, 'Z', "", "", false)

	s := buf.Stats()
	if s.WriteCount != 0 {
		t.Errorf("out-of-bounds writes should not be counted, got WriteCount=%d", s.WriteCount)
	}
	if s.DirtyW != 0 || s.DirtyH != 0 {
		t.Errorf("out-of-bounds should not expand dirty rect, got W=%d H=%d", s.DirtyW, s.DirtyH)
	}
}

func TestCellBuffer_Stats_MixedWritesAndClears(t *testing.T) {
	buf := NewCellBuffer(10, 10)
	// Write 3 cells
	buf.SetChar(1, 1, 'A', "", "", false)
	buf.SetChar(2, 1, 'B', "", "", false)
	buf.SetChar(3, 1, 'C', "", "", false)
	// Clear a 2x2 rect
	buf.ClearRect(5, 5, 2, 2)

	s := buf.Stats()
	if s.WriteCount != 3 {
		t.Errorf("WriteCount: got %d, want 3", s.WriteCount)
	}
	if s.ClearCount != 4 {
		t.Errorf("ClearCount: got %d, want 4", s.ClearCount)
	}
	// Dirty bounding box should span both regions
	if s.DirtyX != 1 || s.DirtyY != 1 || s.DirtyW != 6 || s.DirtyH != 6 {
		t.Errorf("DirtyRect: got (%d,%d,%d,%d), want (1,1,6,6)", s.DirtyX, s.DirtyY, s.DirtyW, s.DirtyH)
	}
}

func TestPaintDirty_ParentDirtyClearsChildFlags(t *testing.T) {
	buf := NewCellBuffer(20, 10)

	parent := &Node{
		Type:       "box",
		X: 0, Y: 0, W: 20, H: 10,
		PaintDirty: true,
		Style:      Style{Background: "#222222", Right: -1, Bottom: -1},
	}
	child := &Node{
		Type:       "text",
		Content:    "Child",
		X: 1, Y: 1, W: 10, H: 1,
		PaintDirty: true,
		Style:      Style{Foreground: "#FFFFFF", Right: -1, Bottom: -1},
	}
	grandchild := &Node{
		Type:       "text",
		Content:    "GC",
		X: 2, Y: 2, W: 5, H: 1,
		PaintDirty: true,
		Style:      Style{Foreground: "#AAAAAA", Right: -1, Bottom: -1},
	}
	parent.AddChild(child)
	child.AddChild(grandchild)

	PaintDirty(buf, parent)

	// Parent dirty flag should be cleared
	if parent.PaintDirty {
		t.Error("expected parent PaintDirty cleared")
	}
	// Child's PaintDirty should also be cleared (painted by parent's paintNode)
	if child.PaintDirty {
		t.Error("expected child PaintDirty cleared when parent was dirty")
	}
	// Grandchild's PaintDirty should also be cleared
	if grandchild.PaintDirty {
		t.Error("expected grandchild PaintDirty cleared when parent was dirty")
	}
}

func TestPaintText_CJKWideChars(t *testing.T) {
	buf := NewCellBuffer(20, 5)
	node := &Node{
		Type:    "text",
		Content: "你好",
		X: 0, Y: 0, W: 10, H: 1,
		Style: Style{Foreground: "#FFFFFF", Right: -1, Bottom: -1},
	}

	PaintFull(buf, node)

	// '你' should be at (0,0) and occupy 2 columns
	c := buf.Get(0, 0)
	if c.Ch != '你' {
		t.Errorf("at (0,0): expected '你', got %q", c.Ch)
	}
	// (1,0) should be the padding cell (Wide=true)
	c = buf.Get(1, 0)
	if !c.Wide {
		t.Error("at (1,0): expected Wide=true for padding cell")
	}

	// '好' should be at (2,0)
	c = buf.Get(2, 0)
	if c.Ch != '好' {
		t.Errorf("at (2,0): expected '好', got %q", c.Ch)
	}
	// (3,0) should be the padding cell
	c = buf.Get(3, 0)
	if !c.Wide {
		t.Error("at (3,0): expected Wide=true for padding cell")
	}

	// (4,0) should be empty
	c = buf.Get(4, 0)
	if c.Ch != 0 {
		t.Errorf("at (4,0): expected empty, got %q", c.Ch)
	}
}

func TestPaintText_MixedASCIIAndCJK(t *testing.T) {
	buf := NewCellBuffer(20, 5)
	node := &Node{
		Type:    "text",
		Content: "A你B",
		X: 0, Y: 0, W: 10, H: 1,
		Style: Style{Foreground: "#FFFFFF", Right: -1, Bottom: -1},
	}

	PaintFull(buf, node)

	// 'A' at (0,0) — 1 column
	if c := buf.Get(0, 0); c.Ch != 'A' {
		t.Errorf("at (0,0): expected 'A', got %q", c.Ch)
	}
	// '你' at (1,0) — 2 columns
	if c := buf.Get(1, 0); c.Ch != '你' {
		t.Errorf("at (1,0): expected '你', got %q", c.Ch)
	}
	// padding at (2,0)
	if c := buf.Get(2, 0); !c.Wide {
		t.Error("at (2,0): expected Wide=true")
	}
	// 'B' at (3,0)
	if c := buf.Get(3, 0); c.Ch != 'B' {
		t.Errorf("at (3,0): expected 'B', got %q", c.Ch)
	}
}

func TestPaintInput_SingleLineDoesNotWrapIntoNextRow(t *testing.T) {
	buf := NewCellBuffer(24, 3)
	for y := 0; y < 3; y++ {
		for x := 0; x < 24; x++ {
			buf.Set(x, y, Cell{Ch: '.', BG: "#000001"})
		}
	}
	n := NewNode("input")
	n.Content = "姐姐斤斤计较"
	n.X, n.Y, n.W, n.H = 0, 0, 8, 1
	n.Style = Style{Foreground: "#FFFFFF", Background: "#222222"}

	paintInput(buf, n)

	for x := 0; x < 24; x++ {
		c := buf.Get(x, 1)
		if c.Ch != '.' {
			t.Fatalf("y=1 col %d: got Ch=%q (input must not wrap past H=1)", x, c.Ch)
		}
	}
}
