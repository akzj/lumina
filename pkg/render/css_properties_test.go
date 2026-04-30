package render

import (
	"testing"
)

// --- display:none ---

func TestDisplayNone_Layout(t *testing.T) {
	// A child with display:none should get W=0, H=0 and siblings should fill the space.
	root := &Node{Type: "vbox", Style: Style{}}
	child1 := &Node{Type: "box", Style: Style{Height: 5}, Parent: root}
	child2 := &Node{Type: "box", Style: Style{Display: "none", Height: 5}, Parent: root}
	child3 := &Node{Type: "box", Style: Style{Height: 5}, Parent: root}
	root.Children = []*Node{child1, child2, child3}

	LayoutFull(root, 0, 0, 20, 20)

	if child2.W != 0 || child2.H != 0 {
		t.Errorf("display:none child should have W=0, H=0, got W=%d, H=%d", child2.W, child2.H)
	}
	if child1.H != 5 {
		t.Errorf("child1 should have H=5, got %d", child1.H)
	}
	if child3.H != 5 {
		t.Errorf("child3 should have H=5, got %d", child3.H)
	}
	// child3 should start right after child1 (no gap for display:none child)
	expectedY3 := child1.Y + child1.H
	if child3.Y != expectedY3 {
		t.Errorf("child3.Y should be %d (after child1), got %d", expectedY3, child3.Y)
	}
}

func TestDisplayNone_HBox(t *testing.T) {
	root := &Node{Type: "hbox", Style: Style{}}
	child1 := &Node{Type: "box", Style: Style{Width: 5}, Parent: root}
	child2 := &Node{Type: "box", Style: Style{Display: "none", Width: 5}, Parent: root}
	child3 := &Node{Type: "box", Style: Style{Width: 5}, Parent: root}
	root.Children = []*Node{child1, child2, child3}

	LayoutFull(root, 0, 0, 20, 10)

	if child2.W != 0 || child2.H != 0 {
		t.Errorf("display:none child in hbox should have W=0, H=0, got W=%d, H=%d", child2.W, child2.H)
	}
	// child3 should start right after child1
	expectedX3 := child1.X + child1.W
	if child3.X != expectedX3 {
		t.Errorf("child3.X should be %d, got %d", expectedX3, child3.X)
	}
}

func TestDisplayNone_Paint(t *testing.T) {
	// display:none node should not paint anything
	buf := NewCellBuffer(10, 5)
	node := &Node{
		Type:    "text",
		Content: "HIDDEN",
		Style:   Style{Display: "none", Foreground: "#FF0000"},
		X: 0, Y: 0, W: 10, H: 1,
	}
	paintNode(buf, node)

	// Buffer should be empty
	for x := 0; x < 10; x++ {
		c := buf.Get(x, 0)
		if c.Ch != 0 {
			t.Errorf("display:none node painted at x=%d, ch=%c", x, c.Ch)
		}
	}
}

func TestDisplayNone_HitTest(t *testing.T) {
	node := &Node{
		Type:  "box",
		Style: Style{Display: "none"},
		X: 0, Y: 0, W: 10, H: 10,
	}
	hit := hitTestWithOffset(node, 5, 5, 0)
	if hit != nil {
		t.Error("display:none node should not be hit-testable")
	}
}

// --- visibility:hidden ---

func TestVisibilityHidden_Layout(t *testing.T) {
	// visibility:hidden still takes up space
	root := &Node{Type: "vbox", Style: Style{}}
	child1 := &Node{Type: "box", Style: Style{Height: 5}, Parent: root}
	child2 := &Node{Type: "box", Style: Style{Visibility: "hidden", Height: 5}, Parent: root}
	child3 := &Node{Type: "box", Style: Style{Height: 5}, Parent: root}
	root.Children = []*Node{child1, child2, child3}

	LayoutFull(root, 0, 0, 20, 20)

	// child2 should still have its size
	if child2.H != 5 {
		t.Errorf("visibility:hidden child should still have H=5, got %d", child2.H)
	}
	// child3 should be after child2
	expectedY3 := child2.Y + child2.H
	if child3.Y != expectedY3 {
		t.Errorf("child3.Y should be %d, got %d", expectedY3, child3.Y)
	}
}

func TestVisibilityHidden_Paint(t *testing.T) {
	buf := NewCellBuffer(10, 5)
	node := &Node{
		Type:    "text",
		Content: "HIDDEN",
		Style:   Style{Visibility: "hidden", Foreground: "#FF0000"},
		X: 0, Y: 0, W: 10, H: 1,
	}
	paintNode(buf, node)

	// Buffer should be empty — visibility:hidden doesn't paint
	for x := 0; x < 6; x++ {
		c := buf.Get(x, 0)
		if c.Ch != 0 {
			t.Errorf("visibility:hidden node painted at x=%d, ch=%c", x, c.Ch)
		}
	}
}

func TestVisibilityHidden_HitTest(t *testing.T) {
	node := &Node{
		Type:  "box",
		Style: Style{Visibility: "hidden"},
		X: 0, Y: 0, W: 10, H: 10,
	}
	hit := hitTestWithOffset(node, 5, 5, 0)
	if hit != nil {
		t.Error("visibility:hidden node should not be hit-testable")
	}
}

// --- text decorations ---

func TestItalic_Cell(t *testing.T) {
	buf := NewCellBuffer(10, 1)
	node := &Node{
		Type:    "text",
		Content: "Hi",
		Style:   Style{Italic: true, Foreground: "#FFFFFF"},
		X: 0, Y: 0, W: 10, H: 1,
	}
	paintNode(buf, node)

	c := buf.Get(0, 0)
	if c.Ch != 'H' {
		t.Errorf("expected 'H', got %c", c.Ch)
	}
	if !c.Italic {
		t.Error("expected Italic=true on cell")
	}
}

func TestStrikethrough_Cell(t *testing.T) {
	buf := NewCellBuffer(10, 1)
	node := &Node{
		Type:    "text",
		Content: "X",
		Style:   Style{Strikethrough: true, Foreground: "#FFFFFF"},
		X: 0, Y: 0, W: 10, H: 1,
	}
	paintNode(buf, node)

	c := buf.Get(0, 0)
	if !c.Strikethrough {
		t.Error("expected Strikethrough=true on cell")
	}
}

func TestInverse_Cell(t *testing.T) {
	buf := NewCellBuffer(10, 1)
	node := &Node{
		Type:    "text",
		Content: "I",
		Style:   Style{Inverse: true, Foreground: "#FFFFFF"},
		X: 0, Y: 0, W: 10, H: 1,
	}
	paintNode(buf, node)

	c := buf.Get(0, 0)
	if !c.Inverse {
		t.Error("expected Inverse=true on cell")
	}
}

// --- textAlign ---

func TestTextAlign_Center(t *testing.T) {
	buf := NewCellBuffer(10, 1)
	node := &Node{
		Type:    "text",
		Content: "Hi",
		Style:   Style{TextAlign: "center", WhiteSpace: "nowrap"},
		X: 0, Y: 0, W: 10, H: 1,
	}
	paintNode(buf, node)

	// "Hi" is 2 chars wide, centered in 10 → offset = (10-2)/2 = 4
	c := buf.Get(4, 0)
	if c.Ch != 'H' {
		t.Errorf("expected 'H' at x=4 (centered), got %c", c.Ch)
	}
	c = buf.Get(5, 0)
	if c.Ch != 'i' {
		t.Errorf("expected 'i' at x=5 (centered), got %c", c.Ch)
	}
	// Position 0-3 should be empty
	c = buf.Get(0, 0)
	if c.Ch != 0 {
		t.Errorf("expected empty at x=0, got %c", c.Ch)
	}
}

func TestTextAlign_Right(t *testing.T) {
	buf := NewCellBuffer(10, 1)
	node := &Node{
		Type:    "text",
		Content: "Hi",
		Style:   Style{TextAlign: "right", WhiteSpace: "nowrap"},
		X: 0, Y: 0, W: 10, H: 1,
	}
	paintNode(buf, node)

	// "Hi" is 2 chars wide, right-aligned in 10 → offset = 10-2 = 8
	c := buf.Get(8, 0)
	if c.Ch != 'H' {
		t.Errorf("expected 'H' at x=8 (right-aligned), got %c", c.Ch)
	}
	c = buf.Get(9, 0)
	if c.Ch != 'i' {
		t.Errorf("expected 'i' at x=9 (right-aligned), got %c", c.Ch)
	}
}

func TestTextAlign_Center_WrapMode(t *testing.T) {
	// textAlign should also work in wrap mode (default whiteSpace)
	buf := NewCellBuffer(10, 1)
	node := &Node{
		Type:    "text",
		Content: "AB",
		Style:   Style{TextAlign: "center"},
		X: 0, Y: 0, W: 10, H: 1,
	}
	paintNode(buf, node)

	// "AB" is 2 chars wide, centered in 10 → offset = 4
	c := buf.Get(4, 0)
	if c.Ch != 'A' {
		t.Errorf("expected 'A' at x=4 (centered, wrap mode), got %c", c.Ch)
	}
}

// --- whiteSpace: nowrap ---

func TestWhiteSpace_Nowrap_Layout(t *testing.T) {
	// With nowrap, text that would wrap should still be height=1
	root := &Node{Type: "vbox", Style: Style{}}
	child := &Node{
		Type:    "text",
		Content: "This is a long line that exceeds width",
		Style:   Style{WhiteSpace: "nowrap"},
		Parent:  root,
	}
	root.Children = []*Node{child}

	LayoutFull(root, 0, 0, 10, 10)

	if child.H != 1 {
		t.Errorf("nowrap text should have H=1, got %d", child.H)
	}
}

func TestWhiteSpace_Nowrap_Paint(t *testing.T) {
	// Nowrap text should clip at the edge, not wrap
	buf := NewCellBuffer(5, 2)
	node := &Node{
		Type:    "text",
		Content: "ABCDEFGH",
		Style:   Style{WhiteSpace: "nowrap"},
		X: 0, Y: 0, W: 5, H: 2,
	}
	paintNode(buf, node)

	// Only first 5 chars should be visible on row 0
	for i, expected := range "ABCDE" {
		c := buf.Get(i, 0)
		if c.Ch != expected {
			t.Errorf("x=%d: expected %c, got %c", i, expected, c.Ch)
		}
	}
	// Row 1 should be empty (no wrapping)
	c := buf.Get(0, 1)
	if c.Ch != 0 {
		t.Errorf("row 1 should be empty with nowrap, got %c", c.Ch)
	}
}

// --- textOverflow: ellipsis ---

func TestTextOverflow_Ellipsis(t *testing.T) {
	buf := NewCellBuffer(5, 1)
	node := &Node{
		Type:    "text",
		Content: "ABCDEFGH",
		Style:   Style{WhiteSpace: "nowrap", TextOverflow: "ellipsis"},
		X: 0, Y: 0, W: 5, H: 1,
	}
	paintNode(buf, node)

	// Should show "ABCD…" (4 chars + ellipsis)
	for i, expected := range "ABCD" {
		c := buf.Get(i, 0)
		if c.Ch != expected {
			t.Errorf("x=%d: expected %c, got %c", i, expected, c.Ch)
		}
	}
	c := buf.Get(4, 0)
	if c.Ch != '…' {
		t.Errorf("x=4: expected '…', got %c (0x%x)", c.Ch, c.Ch)
	}
}

func TestTextOverflow_Ellipsis_ShortText(t *testing.T) {
	// Text shorter than width should NOT get ellipsis
	buf := NewCellBuffer(10, 1)
	node := &Node{
		Type:    "text",
		Content: "Hi",
		Style:   Style{WhiteSpace: "nowrap", TextOverflow: "ellipsis"},
		X: 0, Y: 0, W: 10, H: 1,
	}
	paintNode(buf, node)

	c := buf.Get(0, 0)
	if c.Ch != 'H' {
		t.Errorf("expected 'H', got %c", c.Ch)
	}
	c = buf.Get(1, 0)
	if c.Ch != 'i' {
		t.Errorf("expected 'i', got %c", c.Ch)
	}
	// No ellipsis
	c = buf.Get(2, 0)
	if c.Ch == '…' {
		t.Error("short text should not have ellipsis")
	}
}

// --- borderColor ---

func TestBorderColor(t *testing.T) {
	buf := NewCellBuffer(10, 5)
	node := &Node{
		Type:  "box",
		Style: Style{Border: "single", BorderColor: "#00FF00", Foreground: "#FF0000", Background: "#000000"},
		X: 0, Y: 0, W: 10, H: 5,
	}
	paintNode(buf, node)

	// Border should use BorderColor (#00FF00), not Foreground (#FF0000)
	c := buf.Get(0, 0) // top-left corner
	if c.FG != "#00FF00" {
		t.Errorf("border should use borderColor #00FF00, got %s", c.FG)
	}
	// Verify it's a border char
	if c.Ch != '┌' {
		t.Errorf("expected top-left border '┌', got %c", c.Ch)
	}
}

func TestBorderColor_Fallback(t *testing.T) {
	// When borderColor is empty, should fall back to Foreground
	buf := NewCellBuffer(10, 5)
	node := &Node{
		Type:  "box",
		Style: Style{Border: "single", Foreground: "#FF0000", Background: "#000000"},
		X: 0, Y: 0, W: 10, H: 5,
	}
	paintNode(buf, node)

	c := buf.Get(0, 0)
	if c.FG != "#FF0000" {
		t.Errorf("border without borderColor should use Foreground #FF0000, got %s", c.FG)
	}
}

// --- Combined: display:none with flex siblings ---

func TestDisplayNone_FlexSiblings(t *testing.T) {
	// With display:none, remaining siblings should get more space
	root := &Node{Type: "vbox", Style: Style{}}
	child1 := &Node{Type: "box", Style: Style{Flex: 1}, Parent: root}
	child2 := &Node{Type: "box", Style: Style{Display: "none", Flex: 1}, Parent: root}
	child3 := &Node{Type: "box", Style: Style{Flex: 1}, Parent: root}
	root.Children = []*Node{child1, child2, child3}

	LayoutFull(root, 0, 0, 20, 20)

	// Two visible flex children should split 20 rows equally = 10 each
	if child1.H != 10 {
		t.Errorf("child1 should get H=10 (half of 20), got %d", child1.H)
	}
	if child3.H != 10 {
		t.Errorf("child3 should get H=10 (half of 20), got %d", child3.H)
	}
	if child2.W != 0 || child2.H != 0 {
		t.Errorf("display:none child should have W=0, H=0, got W=%d, H=%d", child2.W, child2.H)
	}
}
