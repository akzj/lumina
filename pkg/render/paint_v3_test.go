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

// --- Comparison Tests: V3 vs V2 ---

func TestV3PaintMatchesV2_SimpleText(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 20, 5

	txt := NewNode("text")
	txt.Content = "Hello"
	txt.Style.Foreground = "#FFFFFF"
	txt.Parent = root
	root.Children = []*Node{txt}

	LayoutFull(root, 0, 0, 20, 5)

	buf2 := NewCellBuffer(20, 5)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(20, 5)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_BoxWithBackground(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 10, 5

	child := NewNode("box")
	child.Style.Background = "#FF0000"
	child.Parent = root
	root.Children = []*Node{child}

	LayoutFull(root, 0, 0, 10, 5)

	buf2 := NewCellBuffer(10, 5)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(10, 5)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_Border(t *testing.T) {
	root := NewNode("box")
	root.Style.Border = "rounded"
	root.Style.Background = "#1E1E2E"
	root.Style.Foreground = "#FFFFFF"
	root.X, root.Y, root.W, root.H = 0, 0, 10, 5

	LayoutFull(root, 0, 0, 10, 5)

	buf2 := NewCellBuffer(10, 5)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(10, 5)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_NestedBoxText(t *testing.T) {
	root := NewNode("vbox")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 20, 10

	box := NewNode("box")
	box.Style.Background = "#FF0000"
	box.Style.Border = "single"
	box.Style.Foreground = "#FFFFFF"
	box.Parent = root

	txt := NewNode("text")
	txt.Content = "Inside"
	txt.Style.Foreground = "#00FF00"
	txt.Parent = box
	box.Children = []*Node{txt}
	root.Children = []*Node{box}

	LayoutFull(root, 0, 0, 20, 10)

	buf2 := NewCellBuffer(20, 10)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(20, 10)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_OverflowHidden(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.Style.Overflow = "hidden"
	root.X, root.Y, root.W, root.H = 0, 0, 10, 5

	// Child that overflows parent
	txt := NewNode("text")
	txt.Content = "This is a very long text that overflows the container"
	txt.Style.Foreground = "#FFFFFF"
	txt.Parent = root
	root.Children = []*Node{txt}

	LayoutFull(root, 0, 0, 10, 5)

	buf2 := NewCellBuffer(10, 5)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(10, 5)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_ScrollContainer(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.Style.Overflow = "scroll"
	root.Style.Scrollbar = "none" // simplify: no scrollbar
	root.X, root.Y, root.W, root.H = 0, 0, 15, 5
	root.ScrollY = 2
	root.ScrollHeight = 20

	// Several children that form a tall scrollable list
	for i := 0; i < 10; i++ {
		txt := NewNode("text")
		txt.Content = fmt.Sprintf("Line %d", i)
		txt.Style.Foreground = "#FFFFFF"
		txt.Parent = root
		root.Children = append(root.Children, txt)
	}

	LayoutFull(root, 0, 0, 15, 5)

	buf2 := NewCellBuffer(15, 5)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(15, 5)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_MultilineTextWrap(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 10, 5

	txt := NewNode("text")
	txt.Content = "Hello World, this wraps around"
	txt.Style.Foreground = "#FFFFFF"
	txt.Parent = root
	root.Children = []*Node{txt}

	LayoutFull(root, 0, 0, 10, 5)

	buf2 := NewCellBuffer(10, 5)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(10, 5)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_TextAlign(t *testing.T) {
	for _, align := range []string{"center", "right"} {
		t.Run(align, func(t *testing.T) {
			root := NewNode("box")
			root.Style.Background = "#1E1E2E"
			root.X, root.Y, root.W, root.H = 0, 0, 20, 3

			txt := NewNode("text")
			txt.Content = "Hi"
			txt.Style.Foreground = "#FFFFFF"
			txt.Style.TextAlign = align
			txt.Parent = root
			root.Children = []*Node{txt}

			LayoutFull(root, 0, 0, 20, 3)

			buf2 := NewCellBuffer(20, 3)
			PaintFull(buf2, root)

			buf3 := NewCellBuffer(20, 3)
			PaintFullV3(buf3, root)

			diffs := compareBufs(buf2, buf3, 10)
			if len(diffs) > 0 {
				t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
				for _, d := range diffs {
					t.Errorf("  %s", d)
				}
				t.Logf("V2:\n%s", dumpBuf(buf2))
				t.Logf("V3:\n%s", dumpBuf(buf3))
			}
		})
	}
}

func TestV3PaintMatchesV2_TextNowrapEllipsis(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 10, 3

	txt := NewNode("text")
	txt.Content = "Very long text that gets truncated"
	txt.Style.Foreground = "#FFFFFF"
	txt.Style.WhiteSpace = "nowrap"
	txt.Style.TextOverflow = "ellipsis"
	txt.Parent = root
	root.Children = []*Node{txt}

	LayoutFull(root, 0, 0, 10, 3)

	buf2 := NewCellBuffer(10, 3)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(10, 3)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_TextSpans(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 20, 3

	txt := NewNode("text")
	txt.Style.Foreground = "#FFFFFF"
	boolTrue := true
	txt.Spans = []Span{
		{Text: "Hello ", Foreground: "#FF0000"},
		{Text: "World", Foreground: "#00FF00", Bold: &boolTrue},
	}
	txt.Parent = root
	root.Children = []*Node{txt}

	LayoutFull(root, 0, 0, 20, 3)

	buf2 := NewCellBuffer(20, 3)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(20, 3)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_ZIndex(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 20, 10

	// Two overlapping boxes with different z-index
	box1 := NewNode("box")
	box1.Style.Background = "#FF0000"
	box1.Style.ZIndex = 1
	box1.Parent = root

	box2 := NewNode("box")
	box2.Style.Background = "#00FF00"
	box2.Style.ZIndex = 2
	box2.Parent = root

	root.Children = []*Node{box1, box2}

	LayoutFull(root, 0, 0, 20, 10)

	// Manually overlap them
	box1.X, box1.Y, box1.W, box1.H = 0, 0, 10, 5
	box2.X, box2.Y, box2.W, box2.H = 5, 2, 10, 5

	buf2 := NewCellBuffer(20, 10)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(20, 10)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_ComponentTransparent(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 20, 5

	comp := NewNode("component")
	comp.Parent = root

	txt := NewNode("text")
	txt.Content = "Via component"
	txt.Style.Foreground = "#FFFFFF"
	txt.Parent = comp
	comp.Children = []*Node{txt}
	root.Children = []*Node{comp}

	LayoutFull(root, 0, 0, 20, 5)

	buf2 := NewCellBuffer(20, 5)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(20, 5)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_DisplayNone(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 20, 5

	hidden := NewNode("box")
	hidden.Style.Display = "none"
	hidden.Style.Background = "#FF0000"
	hidden.Parent = root

	visible := NewNode("text")
	visible.Content = "Visible"
	visible.Style.Foreground = "#FFFFFF"
	visible.Parent = root

	root.Children = []*Node{hidden, visible}

	LayoutFull(root, 0, 0, 20, 5)

	buf2 := NewCellBuffer(20, 5)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(20, 5)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
	}
}

func TestV3PaintMatchesV2_BorderWithPadding(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.Style.Border = "single"
	root.Style.Foreground = "#FFFFFF"
	root.Style.PaddingLeft = 1
	root.Style.PaddingRight = 1
	root.Style.PaddingTop = 1
	root.Style.PaddingBottom = 1
	root.X, root.Y, root.W, root.H = 0, 0, 15, 8

	txt := NewNode("text")
	txt.Content = "Padded"
	txt.Style.Foreground = "#00FF00"
	txt.Parent = root
	root.Children = []*Node{txt}

	LayoutFull(root, 0, 0, 15, 8)

	buf2 := NewCellBuffer(15, 8)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(15, 8)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
}

func TestV3PaintMatchesV2_SplitPaneLayout(t *testing.T) {
	// Simulate a split pane: root with two side-by-side boxes
	root := NewNode("hbox")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 30, 10

	left := NewNode("box")
	left.Style.Background = "#FF0000"
	left.Style.Border = "single"
	left.Style.Foreground = "#FFFFFF"
	left.Parent = root

	right := NewNode("box")
	right.Style.Background = "#0000FF"
	right.Style.Border = "single"
	right.Style.Foreground = "#FFFFFF"
	right.Parent = root

	leftTxt := NewNode("text")
	leftTxt.Content = "Left"
	leftTxt.Style.Foreground = "#FFFFFF"
	leftTxt.Parent = left
	left.Children = []*Node{leftTxt}

	rightTxt := NewNode("text")
	rightTxt.Content = "Right"
	rightTxt.Style.Foreground = "#FFFFFF"
	rightTxt.Parent = right
	right.Children = []*Node{rightTxt}

	root.Children = []*Node{left, right}

	LayoutFull(root, 0, 0, 30, 10)

	buf2 := NewCellBuffer(30, 10)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(30, 10)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
		t.Logf("V2:\n%s", dumpBuf(buf2))
		t.Logf("V3:\n%s", dumpBuf(buf3))
	}
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

func TestV3PaintMatchesV2_HoverStyle(t *testing.T) {
	root := NewNode("box")
	root.Style.Background = "#1E1E2E"
	root.X, root.Y, root.W, root.H = 0, 0, 10, 3

	btn := NewNode("box")
	btn.Style.Background = "#333333"
	btn.Style.Foreground = "#FFFFFF"
	btn.HoverStyle = &Style{Background: "#555555"}
	btn.Hovered = true
	btn.Parent = root
	root.Children = []*Node{btn}

	LayoutFull(root, 0, 0, 10, 3)

	buf2 := NewCellBuffer(10, 3)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(10, 3)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
	}
}

func TestV3PaintMatchesV2_DoubleBorder(t *testing.T) {
	root := NewNode("box")
	root.Style.Border = "double"
	root.Style.Background = "#1E1E2E"
	root.Style.Foreground = "#FFFFFF"
	root.X, root.Y, root.W, root.H = 0, 0, 10, 5

	LayoutFull(root, 0, 0, 10, 5)

	buf2 := NewCellBuffer(10, 5)
	PaintFull(buf2, root)

	buf3 := NewCellBuffer(10, 5)
	PaintFullV3(buf3, root)

	diffs := compareBufs(buf2, buf3, 10)
	if len(diffs) > 0 {
		t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
		for _, d := range diffs {
			t.Errorf("  %s", d)
		}
	}
}

func TestV3PaintMatchesV2_SplitPaneResizeLargeText(t *testing.T) {
	// This is the exact scenario that causes rendering tearing in V2's dirty paint.
	// V3 full paint should produce identical output to V2 full paint.
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

	// Test both initial layout and after resize
	for _, tc := range []struct {
		name   string
		rightW int
	}{
		{"initial_36", 36},
		{"resized_58", 58},
		{"resized_20", 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := build(tc.rightW)
			LayoutFull(root, 0, 0, bufW, bufH)

			// Set scroll to middle
			chatScroll := findSplitResizeNodeByID(root, "chat-scroll")
			if chatScroll != nil {
				chatScroll.ScrollY = computeMaxScrollY(chatScroll) / 2
			}

			buf2 := NewCellBuffer(bufW, bufH)
			PaintFull(buf2, root)

			buf3 := NewCellBuffer(bufW, bufH)
			PaintFullV3(buf3, root)

			diffs := compareBufs(buf2, buf3, 20)
			if len(diffs) > 0 {
				t.Errorf("V3 differs from V2 (%d diffs):", len(diffs))
				for _, d := range diffs {
					t.Errorf("  %s", d)
				}
			}

			// Also verify: no text bleeds past pane boundaries
			// Pane2 (center) ends where divider2 starts
			// Find divider2 X position
			for _, child := range root.Children {
				if child.Style.Width == 1 && child.X > 33 { // divider2
					divX := child.X
					// Check that V3 buffer has divider bg at divX for all rows
					for y := 0; y < bufH; y++ {
						c := buf3.Get(divX, y)
						if c.BG != "" && c.BG != "#38605c" && c.BG != "#120d1a" && c.BG != "#1a1626" {
							t.Errorf("unexpected content at divider x=%d y=%d: ch=%c bg=%s", divX, y, c.Ch, c.BG)
						}
					}
					break
				}
			}
		})
	}
}
