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

func assertChar(t *testing.T, buf *buffer.Buffer, x, y int, want rune) {
	t.Helper()
	c := buf.Get(x, y)
	if c.Char != want {
		t.Errorf("cell(%d,%d) char=%q (0x%X), want %q (0x%X)", x, y, c.Char, c.Char, want, want)
	}
}
