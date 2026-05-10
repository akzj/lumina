package render

import (
	"fmt"
	"strings"
	"testing"
)

// compareBufs compares two CellBuffers cell-by-cell. Returns nil if identical,
// or a list of differences (up to maxDiffs).
func compareBufs(a, b *CellBuffer, maxDiffs int) []string {
	var diffs []string
	if a.Width() != b.Width() || a.Height() != b.Height() {
		return []string{fmt.Sprintf("size mismatch: %dx%d vs %dx%d", a.Width(), a.Height(), b.Width(), b.Height())}
	}
	for y := 0; y < a.Height(); y++ {
		for x := 0; x < a.Width(); x++ {
			ca := a.Get(x, y)
			cb := b.Get(x, y)
			if ca != cb {
				diffs = append(diffs, fmt.Sprintf("(%d,%d): v2={Ch:%q FG:%s BG:%s Bold:%v Wide:%v Dim:%v} v3={Ch:%q FG:%s BG:%s Bold:%v Wide:%v Dim:%v}",
					x, y, ca.Ch, ca.FG, ca.BG, ca.Bold, ca.Wide, ca.Dim,
					cb.Ch, cb.FG, cb.BG, cb.Bold, cb.Wide, cb.Dim))
				if len(diffs) >= maxDiffs {
					return diffs
				}
			}
		}
	}
	return diffs
}

// dumpBuf returns a visual string representation of the buffer for debugging.
func dumpBuf(buf *CellBuffer) string {
	s := ""
	for y := 0; y < buf.Height(); y++ {
		for x := 0; x < buf.Width(); x++ {
			c := buf.Get(x, y)
			if c.Ch == 0 {
				s += "."
			} else if c.Wide {
				s += "W"
			} else {
				s += string(c.Ch)
			}
		}
		s += "\n"
	}
	return s
}

// --- ClipRegion unit tests ---

func TestV3ClipNeverExpands(t *testing.T) {
	c := FullClip(100, 50)
	if c.Empty() {
		t.Error("full clip should not be empty")
	}

	// Narrow to a smaller region
	c2 := c.Narrow(10, 10, 20, 20)
	x1, y1, x2, y2 := c2.Bounds()
	if x1 != 10 || y1 != 10 || x2 != 30 || y2 != 30 {
		t.Errorf("expected (10,10,30,30), got (%d,%d,%d,%d)", x1, y1, x2, y2)
	}

	// Try to "expand" — should stay at intersection
	c3 := c2.Narrow(0, 0, 100, 100)
	x1, y1, x2, y2 = c3.Bounds()
	if x1 != 10 || y1 != 10 || x2 != 30 || y2 != 30 {
		t.Errorf("clip should not expand: expected (10,10,30,30), got (%d,%d,%d,%d)", x1, y1, x2, y2)
	}

	// Narrow further
	c4 := c3.Narrow(15, 15, 5, 5)
	x1, y1, x2, y2 = c4.Bounds()
	if x1 != 15 || y1 != 15 || x2 != 20 || y2 != 20 {
		t.Errorf("expected (15,15,20,20), got (%d,%d,%d,%d)", x1, y1, x2, y2)
	}

	// Narrow to empty
	c5 := c4.Narrow(50, 50, 10, 10)
	if !c5.Empty() {
		t.Error("expected empty clip after narrowing to non-overlapping region")
	}
}

func TestV3CellWriterClipping(t *testing.T) {
	buf := NewCellBuffer(10, 10)
	w := NewCellWriter(buf)

	// Narrow to 5x5 region at (2,2)
	w2 := w.WithClip(2, 2, 5, 5)

	// Write inside clip — should work
	w2.SetChar(3, 3, 'A', "#FFF", "#000", false)
	c := buf.Get(3, 3)
	if c.Ch != 'A' {
		t.Errorf("expected 'A' at (3,3), got %q", c.Ch)
	}

	// Write outside clip — should be silently discarded
	w2.SetChar(0, 0, 'B', "#FFF", "#000", false)
	c = buf.Get(0, 0)
	if c.Ch != 0 {
		t.Errorf("expected zero cell at (0,0) outside clip, got %q", c.Ch)
	}

	// Write at clip edge (inclusive start)
	w2.SetChar(2, 2, 'C', "#FFF", "#000", false)
	c = buf.Get(2, 2)
	if c.Ch != 'C' {
		t.Errorf("expected 'C' at (2,2), got %q", c.Ch)
	}

	// Write at clip edge (exclusive end)
	w2.SetChar(7, 7, 'D', "#FFF", "#000", false)
	c = buf.Get(7, 7)
	if c.Ch != 0 {
		t.Errorf("expected zero cell at (7,7) at exclusive end, got %q", c.Ch)
	}
}

func TestV3CellWriterOffset(t *testing.T) {
	buf := NewCellBuffer(20, 20)
	w := NewCellWriter(buf)

	// Apply offset (simulating scroll)
	w2 := w.WithOffset(5, 10)

	// Write at logical (0,0) — should appear at screen (5,10)
	w2.SetChar(0, 0, 'X', "#FFF", "#000", false)
	c := buf.Get(5, 10)
	if c.Ch != 'X' {
		t.Errorf("expected 'X' at screen (5,10), got %q", c.Ch)
	}

	// Verify nothing at logical position
	c = buf.Get(0, 0)
	if c.Ch != 0 {
		t.Errorf("expected zero at (0,0), got %q", c.Ch)
	}
}


func TestV3DirtyPaint_BasicDirtyNode(t *testing.T) {
	// Paint full, mark a leaf dirty, dirty paint, compare with full repaint
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 20, 5

	txt := NewNode("text")
	txt.Content = "Hello"
	txt.Style.Foreground = "#FFFFFF"
	txt.Parent = root
	root.Children = []*Node{txt}

	LayoutFull(root, 0, 0, 20, 5)

	// Full paint first
	buf := NewCellBuffer(20, 5)
	PaintFullV3(buf, root)

	// Change text content and mark dirty
	txt.Content = "World"
	txt.PaintDirty = true

	// Dirty paint
	PaintDirtyV3(buf, root)

	// Compare with a fresh full paint of the updated tree
	expected := NewCellBuffer(20, 5)
	PaintFullV3(expected, root)

	diffs := compareBufs(expected, buf, 10)
	if len(diffs) > 0 {
		t.Errorf("Dirty paint differs from full paint (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("Expected:\n%s", dumpBuf(expected))
		t.Logf("Got:\n%s", dumpBuf(buf))
	}
}

func TestV3DirtyPaint_ScrollContainerChild(t *testing.T) {
	// A dirty node inside a scroll container — must be clipped correctly
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.Style.Overflow = "scroll"
	root.Style.Scrollbar = "none"
	root.X, root.Y, root.W, root.H = 0, 0, 15, 5
	root.ScrollY = 2
	root.ScrollHeight = 20

	for i := 0; i < 10; i++ {
		txt := NewNode("text")
		txt.Content = fmt.Sprintf("Line %d", i)
		txt.Style.Foreground = "#FFFFFF"
		txt.Parent = root
		root.Children = append(root.Children, txt)
	}

	LayoutFull(root, 0, 0, 15, 5)

	// Full paint
	buf := NewCellBuffer(15, 5)
	PaintFullV3(buf, root)

	// Mark one child dirty (one that's partially visible)
	root.Children[3].Content = "CHANGED"
	root.Children[3].PaintDirty = true

	// Dirty paint
	PaintDirtyV3(buf, root)

	// Compare with fresh full paint
	expected := NewCellBuffer(15, 5)
	PaintFullV3(expected, root)

	diffs := compareBufs(expected, buf, 10)
	if len(diffs) > 0 {
		t.Errorf("Dirty paint differs from full paint (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("Expected:\n%s", dumpBuf(expected))
		t.Logf("Got:\n%s", dumpBuf(buf))
	}
}

func TestV3DirtyPaint_NestedOverflowHidden(t *testing.T) {
	// Dirty node inside overflow:hidden inside overflow:scroll
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.Style.Overflow = "scroll"
	root.Style.Scrollbar = "none"
	root.X, root.Y, root.W, root.H = 0, 0, 20, 10
	root.ScrollY = 1
	root.ScrollHeight = 30

	hidden := NewNode("box")
	hidden.Style.Overflow = "hidden"
	hidden.Style.Background = "#333333"
	hidden.Parent = root
	root.Children = []*Node{hidden}

	// Child that overflows the hidden container
	txt := NewNode("text")
	txt.Content = "This is a long text that should be clipped by hidden parent"
	txt.Style.Foreground = "#FFFFFF"
	txt.Parent = hidden
	hidden.Children = []*Node{txt}

	LayoutFull(root, 0, 0, 20, 10)

	// Full paint
	buf := NewCellBuffer(20, 10)
	PaintFullV3(buf, root)

	// Mark the text dirty
	txt.Content = "UPDATED long text that should also be clipped properly"
	txt.PaintDirty = true

	// Dirty paint
	PaintDirtyV3(buf, root)

	// Compare with fresh full paint
	expected := NewCellBuffer(20, 10)
	PaintFullV3(expected, root)

	diffs := compareBufs(expected, buf, 10)
	if len(diffs) > 0 {
		t.Errorf("Dirty paint differs from full paint (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("Expected:\n%s", dumpBuf(expected))
		t.Logf("Got:\n%s", dumpBuf(buf))
	}
}

func TestV3DirtyPaint_SplitPaneResize(t *testing.T) {
	// The exact tearing scenario:
	// 1. Full paint with rightW=36
	// 2. Change rightW to 37, relayout (which marks dirty)
	// 3. Dirty paint
	// 4. Compare with full repaint of rightW=37
	const bufW, bufH = 180, 40

	build := func(rightW int) *Node {
		leftRows := make([]*Node, 12)
		for i := range leftRows {
			leftRows[i] = &Node{
				Type:    "text",
				Content: "● agent-row-" + string(rune('a'+i)) + " · leaf",
				Style:   Style{Foreground: "#ffffff", WidthPercent: 100, Height: 1},
			}
		}
		leftScroll := &Node{Type: "vbox", Style: Style{
			Overflow: "scroll", Flex: 1, WidthPercent: 100, Background: "#120d1a",
		}, Children: leftRows}
		pane1 := &Node{Type: "vbox", ID: "split-pane-1", Style: Style{
			Width: 32, HeightPercent: 100, Overflow: "hidden", Background: "#120d1a",
		}, Children: []*Node{leftScroll}}

		divider1 := &Node{Type: "vbox", Style: Style{Width: 1, Background: "#38605c"}}
		divider2 := &Node{Type: "vbox", Style: Style{Width: 1, Background: "#38605c"}}

		bigText := strings.Repeat("Rewrite a1c08a842bb505d407544f8bbbd6b0121ccd309 (179/893) (4 seconds passed, remaining 15 predicted)    ", 700)
		rows := make([]*Node, 18)
		for i := range rows {
			content := "normal row"
			if i == 8 {
				content = bigText
			}
			rows[i] = &Node{Type: "text", Content: content, Style: Style{Foreground: "#cdd6f4", WidthPercent: 100}}
		}
		chatScroll := &Node{Type: "vbox", ID: "chat-scroll", Style: Style{
			Overflow: "scroll", Flex: 1, WidthPercent: 100, Background: "#1a1626",
		}, Children: rows}
		pane2 := &Node{Type: "vbox", ID: "split-pane-2", Style: Style{
			Flex: 1, HeightPercent: 100, Overflow: "hidden", Background: "#1a1626",
		}, Children: []*Node{chatScroll}}

		pane3 := &Node{Type: "vbox", ID: "split-pane-3", Style: Style{
			Width: rightW, HeightPercent: 100, Overflow: "hidden", Background: "#120d1a",
		}, Children: []*Node{
			{Type: "text", Content: "LLM Configuration", Style: Style{Height: 1, WidthPercent: 100}},
			{Type: "text", Content: "Provider : anthropic", Style: Style{Height: 1, WidthPercent: 100}},
		}}

		root := &Node{Type: "hbox", Style: Style{WidthPercent: 100, HeightPercent: 100},
			Children: []*Node{pane1, divider1, pane2, divider2, pane3}}
		setParentsRecursive(root)
		return root
	}

	// Step 1: Full paint with rightW=36
	root := build(36)
	LayoutFull(root, 0, 0, bufW, bufH)
	chatScroll := findNodeByID(root, "chat-scroll")
	if chatScroll != nil {
		chatScroll.ScrollY = computeMaxScrollY(chatScroll) / 2
	}
	buf := NewCellBuffer(bufW, bufH)
	PaintFullV3(buf, root)

	// Step 2: Resize pane3 to 37, relayout
	pane3 := findNodeByID(root, "split-pane-3")
	if pane3 == nil {
		t.Fatal("could not find split-pane-3")
	}
	pane3.Style.Width = 37
	LayoutFull(root, 0, 0, bufW, bufH)
	// Scroll may have changed
	if chatScroll != nil {
		maxSY := computeMaxScrollY(chatScroll)
		if chatScroll.ScrollY > maxSY {
			chatScroll.ScrollY = maxSY
		}
	}

	// Step 3: Dirty paint
	PaintDirtyV3(buf, root)

	// Step 4: Compare with fresh full paint of rightW=37
	expected := NewCellBuffer(bufW, bufH)
	PaintFullV3(expected, root)

	diffs := compareBufs(expected, buf, 20)
	if len(diffs) > 0 {
		t.Errorf("V3 dirty paint after resize differs from full paint (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
	}

	// Boundary check: no text bleeds past pane dividers
	for _, child := range root.Children {
		if child.Style.Width == 1 && child.X > 33 { // divider2
			divX := child.X
			for y := 0; y < bufH; y++ {
				c := buf.Get(divX, y)
				if c.BG != "" && c.BG != "#38605c" && c.BG != "#120d1a" && c.BG != "#1a1626" {
					t.Errorf("text bleed at divider x=%d y=%d: ch=%c bg=%s", divX, y, c.Ch, c.BG)
				}
			}
			break
		}
	}
}
