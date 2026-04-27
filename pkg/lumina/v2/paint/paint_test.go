package paint

import (
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

func TestPaint_Box_Background(t *testing.T) {
	// A 5x3 box with background "#FF0000" → all cells have bg="#FF0000"
	node := layout.NewVNode("box")
	node.Style.Background = "#FF0000"
	node.X, node.Y, node.W, node.H = 0, 0, 5, 3

	buf := buffer.New(5, 3)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			c := buf.Get(x, y)
			if c.Background != "#FF0000" {
				t.Errorf("cell(%d,%d) bg=%q, want #FF0000", x, y, c.Background)
			}
			if c.Char != ' ' {
				t.Errorf("cell(%d,%d) char=%q, want ' '", x, y, c.Char)
			}
		}
	}
}

func TestPaint_Text_Simple(t *testing.T) {
	// text "Hi" → cells[0]='H', cells[1]='i'
	node := layout.NewVNode("text")
	node.Content = "Hi"
	node.X, node.Y, node.W, node.H = 0, 0, 10, 1

	buf := buffer.New(10, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	if c := buf.Get(0, 0); c.Char != 'H' {
		t.Errorf("cell(0,0) char=%q, want 'H'", c.Char)
	}
	if c := buf.Get(1, 0); c.Char != 'i' {
		t.Errorf("cell(1,0) char=%q, want 'i'", c.Char)
	}
}

func TestPaint_Text_Foreground(t *testing.T) {
	// text with fg="#00FF00" → cells have correct fg
	node := layout.NewVNode("text")
	node.Content = "AB"
	node.Style.Foreground = "#00FF00"
	node.X, node.Y, node.W, node.H = 0, 0, 10, 1

	buf := buffer.New(10, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	for x := 0; x < 2; x++ {
		c := buf.Get(x, 0)
		if c.Foreground != "#00FF00" {
			t.Errorf("cell(%d,0) fg=%q, want #00FF00", x, c.Foreground)
		}
	}
}

func TestPaint_Box_Border_Single(t *testing.T) {
	// border="single" → ┌┐└┘─│
	node := layout.NewVNode("box")
	node.Style.Border = "single"
	node.X, node.Y, node.W, node.H = 0, 0, 4, 3

	buf := buffer.New(4, 3)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	// Corners
	assertChar(t, buf, 0, 0, '┌')
	assertChar(t, buf, 3, 0, '┐')
	assertChar(t, buf, 0, 2, '└')
	assertChar(t, buf, 3, 2, '┘')

	// Horizontal lines
	assertChar(t, buf, 1, 0, '─')
	assertChar(t, buf, 2, 0, '─')
	assertChar(t, buf, 1, 2, '─')
	assertChar(t, buf, 2, 2, '─')

	// Vertical lines
	assertChar(t, buf, 0, 1, '│')
	assertChar(t, buf, 3, 1, '│')
}

func TestPaint_Box_Border_Rounded(t *testing.T) {
	// border="rounded" → ╭╮╰╯─│
	node := layout.NewVNode("box")
	node.Style.Border = "rounded"
	node.X, node.Y, node.W, node.H = 0, 0, 4, 3

	buf := buffer.New(4, 3)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	assertChar(t, buf, 0, 0, '╭')
	assertChar(t, buf, 3, 0, '╮')
	assertChar(t, buf, 0, 2, '╰')
	assertChar(t, buf, 3, 2, '╯')
	assertChar(t, buf, 1, 0, '─')
	assertChar(t, buf, 0, 1, '│')
}

func TestPaint_Box_Border_Double(t *testing.T) {
	node := layout.NewVNode("box")
	node.Style.Border = "double"
	node.X, node.Y, node.W, node.H = 0, 0, 4, 3

	buf := buffer.New(4, 3)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	assertChar(t, buf, 0, 0, '╔')
	assertChar(t, buf, 3, 0, '╗')
	assertChar(t, buf, 0, 2, '╚')
	assertChar(t, buf, 3, 2, '╝')
	assertChar(t, buf, 1, 0, '═')
	assertChar(t, buf, 0, 1, '║')
}

func TestPaint_Nested(t *testing.T) {
	// box > text → text painted inside box area
	box := layout.NewVNode("box")
	box.Style.Border = "single"
	box.Style.Background = "#000000"
	box.X, box.Y, box.W, box.H = 0, 0, 10, 3

	txt := layout.NewVNode("text")
	txt.Content = "Hi"
	txt.X, txt.Y, txt.W, txt.H = 1, 1, 8, 1 // inside border
	box.AddChild(txt)

	buf := buffer.New(10, 3)
	p := NewPainter()
	p.Paint(buf, box, 0, 0)

	// Border should be present
	assertChar(t, buf, 0, 0, '┌')
	// Text should be inside
	assertChar(t, buf, 1, 1, 'H')
	assertChar(t, buf, 2, 1, 'i')
}

func TestPaint_VBox_Children(t *testing.T) {
	// vbox with 2 text children → stacked vertically
	vbox := layout.NewVNode("vbox")
	vbox.X, vbox.Y, vbox.W, vbox.H = 0, 0, 10, 2

	t1 := layout.NewVNode("text")
	t1.Content = "AA"
	t1.X, t1.Y, t1.W, t1.H = 0, 0, 10, 1

	t2 := layout.NewVNode("text")
	t2.Content = "BB"
	t2.X, t2.Y, t2.W, t2.H = 0, 1, 10, 1

	vbox.AddChild(t1)
	vbox.AddChild(t2)

	buf := buffer.New(10, 2)
	p := NewPainter()
	p.Paint(buf, vbox, 0, 0)

	assertChar(t, buf, 0, 0, 'A')
	assertChar(t, buf, 1, 0, 'A')
	assertChar(t, buf, 0, 1, 'B')
	assertChar(t, buf, 1, 1, 'B')
}

func TestPaint_Offset(t *testing.T) {
	// VNode at abs (10,5), offset=(10,5) → painted at buffer (0,0)
	node := layout.NewVNode("text")
	node.Content = "X"
	node.X, node.Y, node.W, node.H = 10, 5, 5, 1

	buf := buffer.New(5, 1)
	p := NewPainter()
	p.Paint(buf, node, 10, 5)

	assertChar(t, buf, 0, 0, 'X')
	// Ensure nothing at (10,5) in the small buffer (would be out of bounds, should be zero)
}

func TestPaint_Clip(t *testing.T) {
	// VNode extends beyond buffer → clipped, no panic
	node := layout.NewVNode("text")
	node.Content = "Hello World, this is a long text that goes beyond the buffer"
	node.X, node.Y, node.W, node.H = 0, 0, 100, 1

	buf := buffer.New(5, 1) // only 5 wide
	p := NewPainter()
	p.Paint(buf, node, 0, 0) // should not panic

	// First 5 chars should be painted
	assertChar(t, buf, 0, 0, 'H')
	assertChar(t, buf, 4, 0, 'o')
}

func TestPaint_Fragment(t *testing.T) {
	// Fragment just recurses children without painting itself
	frag := layout.NewVNode("fragment")
	frag.X, frag.Y, frag.W, frag.H = 0, 0, 10, 1

	txt := layout.NewVNode("text")
	txt.Content = "Z"
	txt.X, txt.Y, txt.W, txt.H = 0, 0, 10, 1
	frag.AddChild(txt)

	buf := buffer.New(10, 1)
	p := NewPainter()
	p.Paint(buf, frag, 0, 0)

	assertChar(t, buf, 0, 0, 'Z')
}

func TestPaint_TextWrap(t *testing.T) {
	// Text wraps at available width
	node := layout.NewVNode("text")
	node.Content = "ABCDE"
	node.X, node.Y, node.W, node.H = 0, 0, 3, 3

	buf := buffer.New(3, 3)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	// Row 0: A B C
	assertChar(t, buf, 0, 0, 'A')
	assertChar(t, buf, 1, 0, 'B')
	assertChar(t, buf, 2, 0, 'C')
	// Row 1: D E
	assertChar(t, buf, 0, 1, 'D')
	assertChar(t, buf, 1, 1, 'E')
}

func TestPaint_TextNewline(t *testing.T) {
	node := layout.NewVNode("text")
	node.Content = "AB\nCD"
	node.X, node.Y, node.W, node.H = 0, 0, 10, 2

	buf := buffer.New(10, 2)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	assertChar(t, buf, 0, 0, 'A')
	assertChar(t, buf, 1, 0, 'B')
	assertChar(t, buf, 0, 1, 'C')
	assertChar(t, buf, 1, 1, 'D')
}

func TestPaint_NilInputs(t *testing.T) {
	p := NewPainter()
	// Should not panic
	p.Paint(nil, nil, 0, 0)
	p.Paint(buffer.New(1, 1), nil, 0, 0)
	p.Paint(nil, layout.NewVNode("box"), 0, 0)
}

func TestPaint_TextBoldDimUnderline(t *testing.T) {
	node := layout.NewVNode("text")
	node.Content = "X"
	node.Style.Bold = true
	node.Style.Dim = true
	node.Style.Underline = true
	node.Style.Foreground = "#FFFFFF"
	node.X, node.Y, node.W, node.H = 0, 0, 5, 1

	buf := buffer.New(5, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	c := buf.Get(0, 0)
	if !c.Bold {
		t.Error("expected Bold")
	}
	if !c.Dim {
		t.Error("expected Dim")
	}
	if !c.Underline {
		t.Error("expected Underline")
	}
}

func TestPaint_WideChar(t *testing.T) {
	// CJK character '中' should be 2 columns wide
	node := layout.NewVNode("text")
	node.Content = "中"
	node.X, node.Y, node.W, node.H = 0, 0, 5, 1

	buf := buffer.New(5, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	c := buf.Get(0, 0)
	if c.Char != '中' {
		t.Errorf("cell(0,0) char=%q, want '中'", c.Char)
	}
	// Padding cell at x=1
	c1 := buf.Get(1, 0)
	if c1.Char != 0 {
		t.Errorf("cell(1,0) char=%q, want 0 (padding)", c1.Char)
	}
}

func TestPaint_TextInheritsContainerBackground(t *testing.T) {
	// Box with background "#FF0000" + text child without explicit background.
	// Text cells should inherit the box's background.
	box := layout.NewVNode("box")
	box.Style.Background = "#FF0000"
	box.X, box.Y, box.W, box.H = 0, 0, 10, 5

	txt := layout.NewVNode("text")
	txt.Content = "AB"
	txt.X, txt.Y, txt.W, txt.H = 0, 0, 10, 1
	box.AddChild(txt)

	buf := buffer.New(10, 5)
	p := NewPainter()
	p.Paint(buf, box, 0, 0)

	// Text cells should have the box's background
	cellA := buf.Get(0, 0)
	if cellA.Char != 'A' {
		t.Errorf("expected 'A' at (0,0), got %q", cellA.Char)
	}
	if cellA.Background != "#FF0000" {
		t.Errorf("text cell bg: got %q, want '#FF0000' (should inherit from parent box)", cellA.Background)
	}

	cellB := buf.Get(1, 0)
	if cellB.Char != 'B' {
		t.Errorf("expected 'B' at (1,0), got %q", cellB.Char)
	}
	if cellB.Background != "#FF0000" {
		t.Errorf("text cell bg: got %q, want '#FF0000'", cellB.Background)
	}

	// Empty cell in box area should also have background
	cellEmpty := buf.Get(5, 3)
	if cellEmpty.Background != "#FF0000" {
		t.Errorf("empty cell bg: got %q, want '#FF0000'", cellEmpty.Background)
	}
}

func assertChar(t *testing.T, buf *buffer.Buffer, x, y int, want rune) {
	t.Helper()
	c := buf.Get(x, y)
	if c.Char != want {
		t.Errorf("cell(%d,%d) char=%q (0x%X), want %q (0x%X)", x, y, c.Char, c.Char, want, want)
	}
}

// --- Input paint tests ---

func TestPaint_InputBasic(t *testing.T) {
	// Input with value "Hello" → cells show H,e,l,l,o
	node := layout.NewVNode("input")
	node.Content = "Hello"
	node.Style.Foreground = "#CDD6F4"
	node.X, node.Y, node.W, node.H = 0, 0, 20, 1

	buf := buffer.New(20, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	assertChar(t, buf, 0, 0, 'H')
	assertChar(t, buf, 1, 0, 'e')
	assertChar(t, buf, 2, 0, 'l')
	assertChar(t, buf, 3, 0, 'l')
	assertChar(t, buf, 4, 0, 'o')

	// Verify foreground color
	c := buf.Get(0, 0)
	if c.Foreground != "#CDD6F4" {
		t.Errorf("foreground=%q, want #CDD6F4", c.Foreground)
	}
}

func TestPaint_InputPlaceholder(t *testing.T) {
	// Empty input with placeholder → shows placeholder text, dimmed
	node := layout.NewVNode("input")
	node.Content = "" // empty value
	node.Props["placeholder"] = "Type here"
	node.Style.Foreground = "#CDD6F4"
	node.X, node.Y, node.W, node.H = 0, 0, 20, 1

	buf := buffer.New(20, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	// Should show placeholder text
	assertChar(t, buf, 0, 0, 'T')
	assertChar(t, buf, 1, 0, 'y')
	assertChar(t, buf, 2, 0, 'p')
	assertChar(t, buf, 3, 0, 'e')

	// Placeholder text should be dimmed
	c := buf.Get(0, 0)
	if !c.Dim {
		t.Error("placeholder text should be Dim=true")
	}
	// Placeholder uses muted foreground color
	if c.Foreground != "#6C7086" {
		t.Errorf("placeholder fg=%q, want #6C7086", c.Foreground)
	}
}

func TestPaint_InputCursor(t *testing.T) {
	// Focused input with cursor at position 3 → cell at position 3 has cursor bg
	node := layout.NewVNode("input")
	node.Content = "Hello"
	node.Props["focused"] = true
	node.Props["cursorPos"] = 3
	node.Style.Foreground = "#CDD6F4"
	node.Style.Background = "#313244"
	node.X, node.Y, node.W, node.H = 0, 0, 20, 1

	buf := buffer.New(20, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	// Text should still be rendered
	assertChar(t, buf, 0, 0, 'H')
	assertChar(t, buf, 3, 0, 'l')

	// Cell at cursor position 3 should have cursor background
	cursorCell := buf.Get(3, 0)
	if cursorCell.Background != "#585B70" {
		t.Errorf("cursor cell bg=%q, want #585B70", cursorCell.Background)
	}

	// Cell NOT at cursor should have normal background
	normalCell := buf.Get(0, 0)
	if normalCell.Background != "#313244" {
		t.Errorf("normal cell bg=%q, want #313244", normalCell.Background)
	}
}

func TestPaint_InputCursorAtEnd(t *testing.T) {
	// Focused input with cursor at end of text → cursor block after text
	node := layout.NewVNode("input")
	node.Content = "Hi"
	node.Props["focused"] = true
	// cursorPos defaults to len(value) = 2
	node.Style.Foreground = "#CDD6F4"
	node.X, node.Y, node.W, node.H = 0, 0, 20, 1

	buf := buffer.New(20, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	assertChar(t, buf, 0, 0, 'H')
	assertChar(t, buf, 1, 0, 'i')

	// Cursor at end: position 2 should be a space with cursor bg
	cursorCell := buf.Get(2, 0)
	if cursorCell.Char != ' ' {
		t.Errorf("cursor-at-end char=%q, want ' '", cursorCell.Char)
	}
	if cursorCell.Background != "#585B70" {
		t.Errorf("cursor-at-end bg=%q, want #585B70", cursorCell.Background)
	}
}

func TestPaint_InputWithBorder(t *testing.T) {
	// Input with border → text inside border area
	node := layout.NewVNode("input")
	node.Content = "AB"
	node.Style.Border = "single"
	node.Style.Background = "#313244"
	node.X, node.Y, node.W, node.H = 0, 0, 10, 3

	buf := buffer.New(10, 3)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	// Border corners
	assertChar(t, buf, 0, 0, '┌')
	assertChar(t, buf, 9, 0, '┐')

	// Text inside border (at x=1, y=1)
	assertChar(t, buf, 1, 1, 'A')
	assertChar(t, buf, 2, 1, 'B')
}

func TestPaint_InputWithBackground(t *testing.T) {
	// Input with background fills entire area
	node := layout.NewVNode("input")
	node.Content = "X"
	node.Style.Background = "#FF0000"
	node.X, node.Y, node.W, node.H = 0, 0, 10, 1

	buf := buffer.New(10, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	// Text cell should have background
	c := buf.Get(0, 0)
	if c.Background != "#FF0000" {
		t.Errorf("text cell bg=%q, want #FF0000", c.Background)
	}

	// Empty cell should also have background (from Fill)
	c2 := buf.Get(5, 0)
	if c2.Background != "#FF0000" {
		t.Errorf("empty cell bg=%q, want #FF0000", c2.Background)
	}
}

func TestPaint_InputEmpty_NoCursor(t *testing.T) {
	// Empty unfocused input → nothing visible (no cursor, no placeholder)
	node := layout.NewVNode("input")
	node.Content = ""
	node.X, node.Y, node.W, node.H = 0, 0, 10, 1

	buf := buffer.New(10, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	// All cells should be zero (nothing rendered)
	for x := 0; x < 10; x++ {
		c := buf.Get(x, 0)
		if c.Char != 0 && c.Char != ' ' {
			t.Errorf("cell(%d,0) char=%q, want 0 or ' '", x, c.Char)
		}
	}
}

// --- Scroll container paint tests ---

func TestPaint_ScrollContainer_Clips(t *testing.T) {
	// A scroll container of height 3 with 5 text children (each 1 row).
	// Only the first 3 should be visible (scrollY=0).
	container := layout.NewVNode("vbox")
	container.Style.Overflow = "scroll"
	container.X, container.Y, container.W, container.H = 0, 0, 11, 3 // 10 content + 1 scrollbar

	// Create 5 children, each 1 row tall, at positions 0-4.
	for i := 0; i < 5; i++ {
		child := layout.NewVNode("text")
		child.Content = string(rune('A' + i)) // A, B, C, D, E
		child.X, child.Y, child.W, child.H = 0, i, 10, 1
		container.AddChild(child)
	}

	buf := buffer.New(11, 3)
	p := NewPainter()
	p.Paint(buf, container, 0, 0)

	// Visible: A(row0), B(row1), C(row2)
	assertChar(t, buf, 0, 0, 'A')
	assertChar(t, buf, 0, 1, 'B')
	assertChar(t, buf, 0, 2, 'C')
}

func TestPaint_ScrollContainer_Offset(t *testing.T) {
	// Same container but scrollY=2 → shows C, D, E.
	container := layout.NewVNode("vbox")
	container.Style.Overflow = "scroll"
	container.ScrollY = 2
	container.X, container.Y, container.W, container.H = 0, 0, 11, 3

	for i := 0; i < 5; i++ {
		child := layout.NewVNode("text")
		child.Content = string(rune('A' + i))
		child.X, child.Y, child.W, child.H = 0, i, 10, 1
		container.AddChild(child)
	}

	buf := buffer.New(11, 3)
	p := NewPainter()
	p.Paint(buf, container, 0, 0)

	// Visible: C(row0), D(row1), E(row2)
	assertChar(t, buf, 0, 0, 'C')
	assertChar(t, buf, 0, 1, 'D')
	assertChar(t, buf, 0, 2, 'E')
}

func TestPaint_ScrollContainer_WithBorder(t *testing.T) {
	// Scroll container with border: 12w x 5h, border=single.
	// Content area = 10w x 3h (minus 2 for border, minus 1 for scrollbar = 9w).
	container := layout.NewVNode("vbox")
	container.Style.Overflow = "scroll"
	container.Style.Border = "single"
	container.X, container.Y, container.W, container.H = 0, 0, 12, 5

	// 6 children at absolute positions inside border.
	for i := 0; i < 6; i++ {
		child := layout.NewVNode("text")
		child.Content = string(rune('A' + i))
		child.X, child.Y, child.W, child.H = 1, 1+i, 9, 1 // inside border
		container.AddChild(child)
	}

	buf := buffer.New(12, 5)
	p := NewPainter()
	p.Paint(buf, container, 0, 0)

	// Border should be present.
	assertChar(t, buf, 0, 0, '┌')
	assertChar(t, buf, 11, 0, '┐')

	// Inside border, content area starts at (1,1), visible height=3.
	// Visible: A(row1), B(row2), C(row3)
	assertChar(t, buf, 1, 1, 'A')
	assertChar(t, buf, 1, 2, 'B')
	assertChar(t, buf, 1, 3, 'C')
}

func TestPaint_Scrollbar_Position(t *testing.T) {
	// Container: 11w x 4h, 8 children → scrollbar should appear.
	container := layout.NewVNode("vbox")
	container.Style.Overflow = "scroll"
	container.X, container.Y, container.W, container.H = 0, 0, 11, 4

	for i := 0; i < 8; i++ {
		child := layout.NewVNode("text")
		child.Content = string(rune('A' + i))
		child.X, child.Y, child.W, child.H = 0, i, 10, 1
		container.AddChild(child)
	}

	buf := buffer.New(11, 4)
	p := NewPainter()
	p.Paint(buf, container, 0, 0)

	// Scrollbar is in column 10 (contentW=10, scrollbar at x=10).
	// With scrollY=0, thumb should be at the top.
	// Thumb height = viewH * viewH / totalH = 4*4/8 = 2
	// Thumb at y=0.
	scrollbarX := 10
	thumbCell := buf.Get(scrollbarX, 0)
	if thumbCell.Char != '█' {
		t.Errorf("scrollbar thumb at (10,0) char=%q, want '█'", thumbCell.Char)
	}
	thumbCell2 := buf.Get(scrollbarX, 1)
	if thumbCell2.Char != '█' {
		t.Errorf("scrollbar thumb at (10,1) char=%q, want '█'", thumbCell2.Char)
	}
	// Track below thumb.
	trackCell := buf.Get(scrollbarX, 2)
	if trackCell.Char != ' ' {
		t.Errorf("scrollbar track at (10,2) char=%q, want ' '", trackCell.Char)
	}
}

func TestPaint_Scrollbar_ContentFits(t *testing.T) {
	// When content fits (2 children, 4 rows), scrollbar track should still show
	// but no thumb needed (content <= viewH).
	container := layout.NewVNode("vbox")
	container.Style.Overflow = "scroll"
	container.X, container.Y, container.W, container.H = 0, 0, 11, 4

	for i := 0; i < 2; i++ {
		child := layout.NewVNode("text")
		child.Content = string(rune('A' + i))
		child.X, child.Y, child.W, child.H = 0, i, 10, 1
		container.AddChild(child)
	}

	buf := buffer.New(11, 4)
	p := NewPainter()
	p.Paint(buf, container, 0, 0)

	// Content fits → scrollbar track with no thumb (all spaces).
	scrollbarX := 10
	for dy := 0; dy < 4; dy++ {
		cell := buf.Get(scrollbarX, dy)
		if cell.Char != ' ' {
			t.Errorf("scrollbar at (%d,%d) char=%q, want ' ' (content fits)", scrollbarX, dy, cell.Char)
		}
	}
}

func TestPaint_ScrollContainer_ScrollYClamped(t *testing.T) {
	// scrollY larger than max → should be clamped, show last rows.
	container := layout.NewVNode("vbox")
	container.Style.Overflow = "scroll"
	container.ScrollY = 100 // way beyond content
	container.X, container.Y, container.W, container.H = 0, 0, 11, 3

	for i := 0; i < 5; i++ {
		child := layout.NewVNode("text")
		child.Content = string(rune('A' + i))
		child.X, child.Y, child.W, child.H = 0, i, 10, 1
		container.AddChild(child)
	}

	buf := buffer.New(11, 3)
	p := NewPainter()
	p.Paint(buf, container, 0, 0)

	// maxScroll = 5-3 = 2, so clamped to scrollY=2 → shows C, D, E.
	assertChar(t, buf, 0, 0, 'C')
	assertChar(t, buf, 0, 1, 'D')
	assertChar(t, buf, 0, 2, 'E')
}

func TestPaint_ScrollContainer_WithBackground(t *testing.T) {
	// Scroll container with background → background fills entire container.
	container := layout.NewVNode("vbox")
	container.Style.Overflow = "scroll"
	container.Style.Background = "#FF0000"
	container.X, container.Y, container.W, container.H = 0, 0, 11, 3

	child := layout.NewVNode("text")
	child.Content = "Hi"
	child.X, child.Y, child.W, child.H = 0, 0, 10, 1
	container.AddChild(child)

	buf := buffer.New(11, 3)
	p := NewPainter()
	p.Paint(buf, container, 0, 0)

	// Background should fill the whole container (including scrollbar area).
	for y := 0; y < 3; y++ {
		for x := 0; x < 10; x++ {
			c := buf.Get(x, y)
			if c.Background != "#FF0000" {
				t.Errorf("cell(%d,%d) bg=%q, want #FF0000", x, y, c.Background)
			}
		}
	}

	// Text should be visible.
	assertChar(t, buf, 0, 0, 'H')
	assertChar(t, buf, 1, 0, 'i')
}

func TestPaint_ScrollContainer_Scrollbar_BottomPosition(t *testing.T) {
	// When scrolled to bottom, thumb should be at the bottom.
	container := layout.NewVNode("vbox")
	container.Style.Overflow = "scroll"
	container.ScrollY = 4 // maxScroll = 8-4 = 4
	container.X, container.Y, container.W, container.H = 0, 0, 11, 4

	for i := 0; i < 8; i++ {
		child := layout.NewVNode("text")
		child.Content = string(rune('A' + i))
		child.X, child.Y, child.W, child.H = 0, i, 10, 1
		container.AddChild(child)
	}

	buf := buffer.New(11, 4)
	p := NewPainter()
	p.Paint(buf, container, 0, 0)

	// Thumb height = 4*4/8 = 2.
	// scrollY=4, scrollRange=4, trackRange=4-2=2.
	// thumbY = 4*2/4 = 2 → thumb at rows 2,3.
	scrollbarX := 10
	trackCell := buf.Get(scrollbarX, 0)
	if trackCell.Char != ' ' {
		t.Errorf("scrollbar track at (10,0) char=%q, want ' '", trackCell.Char)
	}
	thumbCell := buf.Get(scrollbarX, 2)
	if thumbCell.Char != '█' {
		t.Errorf("scrollbar thumb at (10,2) char=%q, want '█'", thumbCell.Char)
	}
	thumbCell2 := buf.Get(scrollbarX, 3)
	if thumbCell2.Char != '█' {
		t.Errorf("scrollbar thumb at (10,3) char=%q, want '█'", thumbCell2.Char)
	}
}

func TestPaint_CJKText_WideFlagSet(t *testing.T) {
	// Paint Chinese text "中文" → each CJK char occupies 2 columns.
	// cell(0,0) should have Char='中', Wide=true
	// cell(2,0) should have Char='文', Wide=true
	node := layout.NewVNode("text")
	node.Content = "中文"
	node.X, node.Y, node.W, node.H = 0, 0, 10, 1

	buf := buffer.New(10, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	c0 := buf.Get(0, 0)
	if c0.Char != '中' {
		t.Errorf("cell(0,0) char=%q, want '中'", c0.Char)
	}
	if !c0.Wide {
		t.Error("cell(0,0) Wide=false, want true for CJK char '中'")
	}

	c2 := buf.Get(2, 0)
	if c2.Char != '文' {
		t.Errorf("cell(2,0) char=%q, want '文'", c2.Char)
	}
	if !c2.Wide {
		t.Error("cell(2,0) Wide=false, want true for CJK char '文'")
	}
}

func TestPaint_CJKText_PaddingCell(t *testing.T) {
	// Paint "中A" → cell(0,0)='中' Wide=true, cell(1,0)=padding (Char=0),
	// cell(2,0)='A' Wide=false
	node := layout.NewVNode("text")
	node.Content = "中A"
	node.X, node.Y, node.W, node.H = 0, 0, 10, 1

	buf := buffer.New(10, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	c0 := buf.Get(0, 0)
	if c0.Char != '中' || !c0.Wide {
		t.Errorf("cell(0,0) char=%q wide=%v, want '中' true", c0.Char, c0.Wide)
	}

	// Padding cell at (1,0)
	c1 := buf.Get(1, 0)
	if c1.Char != 0 {
		t.Errorf("cell(1,0) char=%d, want 0 (padding cell)", c1.Char)
	}
	if c1.Wide {
		t.Error("cell(1,0) Wide=true, want false for padding cell")
	}

	// ASCII 'A' at column 2
	c2 := buf.Get(2, 0)
	if c2.Char != 'A' {
		t.Errorf("cell(2,0) char=%q, want 'A'", c2.Char)
	}
	if c2.Wide {
		t.Error("cell(2,0) Wide=true, want false for ASCII char")
	}
}

func TestPaint_MixedCJKASCII_Alignment(t *testing.T) {
	// Paint "A中B" → A at col 0, 中 at col 1-2, B at col 3
	node := layout.NewVNode("text")
	node.Content = "A中B"
	node.X, node.Y, node.W, node.H = 0, 0, 10, 1

	buf := buffer.New(10, 1)
	p := NewPainter()
	p.Paint(buf, node, 0, 0)

	if c := buf.Get(0, 0); c.Char != 'A' || c.Wide {
		t.Errorf("cell(0,0) char=%q wide=%v, want 'A' false", c.Char, c.Wide)
	}
	if c := buf.Get(1, 0); c.Char != '中' || !c.Wide {
		t.Errorf("cell(1,0) char=%q wide=%v, want '中' true", c.Char, c.Wide)
	}
	if c := buf.Get(2, 0); c.Char != 0 {
		t.Errorf("cell(2,0) char=%d, want 0 (padding)", c.Char)
	}
	if c := buf.Get(3, 0); c.Char != 'B' || c.Wide {
		t.Errorf("cell(3,0) char=%q wide=%v, want 'B' false", c.Char, c.Wide)
	}
}
