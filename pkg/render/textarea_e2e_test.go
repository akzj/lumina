package render

import (
	"strings"
	"testing"
)

// TestTextareaEnterInsertsNewlineAndGrows verifies that pressing Enter in a
// textarea inserts a newline, moves cursor to the new line, and triggers
// MarkLayoutDirty so the node can grow in height.
func TestTextareaEnterInsertsNewlineAndGrows(t *testing.T) {
	e := &Engine{}
	node := &Node{
		Type:      "textarea",
		Focusable: true,
		Focused:   true,
		Content:   "hello",
		CursorPos: 5, // at end
		Style: Style{
			Flex:      1,
			MinHeight: 1,
			MaxHeight: 8,
		},
	}
	e.focusedNode = node

	// Simulate Enter key
	handled := e.HandleInputKeyDown("Enter")
	if !handled {
		t.Fatal("Enter should be handled by textarea")
	}

	// Verify newline was inserted
	if !strings.Contains(node.Content, "\n") {
		t.Fatalf("Expected newline in content, got %q", node.Content)
	}
	if node.Content != "hello\n" {
		t.Fatalf("Expected 'hello\\n', got %q", node.Content)
	}

	// Verify cursor moved past the newline
	if node.CursorPos != 6 {
		t.Fatalf("Expected cursor at 6, got %d", node.CursorPos)
	}

	// Verify layout was marked dirty (for auto-height)
	if !node.LayoutDirty {
		t.Fatal("MarkLayoutDirty should have been called after Enter in textarea")
	}

	// Verify cursor line/col computation
	runes := []rune(node.Content)
	line, col := computeCursorLineCol(runes, node.CursorPos)
	if line != 1 || col != 0 {
		t.Fatalf("Expected cursor at line=1, col=0, got line=%d, col=%d", line, col)
	}
}

// TestTextareaCtrlJSubmits verifies Ctrl+J triggers onSubmit.
func TestTextareaCtrlJSubmits(t *testing.T) {
	e := &Engine{}
	node := &Node{
		Type:      "textarea",
		Focusable: true,
		Focused:   true,
		Content:   "hello\nworld",
		CursorPos: 11,
	}
	e.focusedNode = node

	// Ctrl+J should be handled (submit trigger)
	handled := e.HandleInputKeyDown("Ctrl+J")
	if !handled {
		t.Fatal("Ctrl+J should be handled (submit)")
	}
}

// TestTextareaBackspaceNewlineMarksLayout verifies that deleting a newline
// marks layout dirty so the textarea can shrink.
func TestTextareaBackspaceNewlineMarksLayout(t *testing.T) {
	e := &Engine{}
	node := &Node{
		Type:      "textarea",
		Focusable: true,
		Focused:   true,
		Content:   "hello\nworld",
		CursorPos: 6, // right after the \n
	}
	e.focusedNode = node

	handled := e.HandleInputKeyDown("Backspace")
	if !handled {
		t.Fatal("Backspace should be handled")
	}

	if node.Content != "helloworld" {
		t.Fatalf("Expected 'helloworld', got %q", node.Content)
	}

	if !node.LayoutDirty {
		t.Fatal("MarkLayoutDirty should have been called after deleting newline")
	}
}

// TestTextareaPaintMultiLine verifies that paintTextarea renders multiple lines.
func TestTextareaPaintMultiLine(t *testing.T) {
	buf := NewCellBuffer(40, 5)
	node := &Node{
		Type:      "textarea",
		Content:   "line1\nline2\nline3",
		CursorPos: 6, // start of line2
		X:         0,
		Y:         0,
		W:         40,
		H:         3,
		Focused:   true,
		Style: Style{
			Foreground: "#FFFFFF",
			Background: "#000000",
		},
	}
	buf.CursorBlinkOn = true

	paintTextarea(buf, node)

	// Check line 1
	row0 := ""
	for x := 0; x < 5; x++ {
		c := buf.Get(x, 0)
		if c.Ch != 0 {
			row0 += string(c.Ch)
		}
	}
	if !strings.HasPrefix(row0, "line1") {
		t.Fatalf("Row 0 should start with 'line1', got %q", row0)
	}

	// Check line 2
	row1 := ""
	for x := 0; x < 5; x++ {
		c := buf.Get(x, 1)
		if c.Ch != 0 {
			row1 += string(c.Ch)
		}
	}
	if !strings.HasPrefix(row1, "line2") {
		t.Fatalf("Row 1 should start with 'line2', got %q", row1)
	}

	// Check cursor is at (0, 1) — start of line2
	runes := []rune(node.Content)
	line, col := computeCursorLineCol(runes, node.CursorPos)
	if line != 1 || col != 0 {
		t.Fatalf("Cursor should be at line=1, col=0, got line=%d, col=%d", line, col)
	}
}
