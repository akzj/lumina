package render

import (
	"testing"
)

func TestSpans_BasicRender(t *testing.T) {
	// Create a text node with spans
	node := &Node{
		Type: "text",
		X:    0, Y: 0, W: 20, H: 1,
		Spans: []Span{
			{Text: "Hi ", Foreground: "#FF0000"},
			{Text: "World", Foreground: "#00FF00", Bold: boolPtr(true)},
		},
		Style: Style{
			Foreground: "#FFFFFF",
			Right:      -1,
			Bottom:     -1,
		},
	}

	buf := NewCellBuffer(20, 1)
	paintText(buf, node)

	// Check "Hi " has red foreground
	cell0 := buf.Get(0, 0)
	if cell0.Ch != 'H' {
		t.Errorf("expected 'H' at (0,0), got %q", cell0.Ch)
	}
	if cell0.FG != "#FF0000" {
		t.Errorf("expected FG=#FF0000 at (0,0), got %q", cell0.FG)
	}
	if cell0.Bold {
		t.Errorf("expected Bold=false at (0,0)")
	}

	cell1 := buf.Get(1, 0)
	if cell1.Ch != 'i' {
		t.Errorf("expected 'i' at (1,0), got %q", cell1.Ch)
	}
	if cell1.FG != "#FF0000" {
		t.Errorf("expected FG=#FF0000 at (1,0), got %q", cell1.FG)
	}

	cell2 := buf.Get(2, 0)
	if cell2.Ch != ' ' {
		t.Errorf("expected ' ' at (2,0), got %q", cell2.Ch)
	}
	if cell2.FG != "#FF0000" {
		t.Errorf("expected FG=#FF0000 at (2,0), got %q", cell2.FG)
	}

	// Check "World" has green foreground and bold
	cell3 := buf.Get(3, 0)
	if cell3.Ch != 'W' {
		t.Errorf("expected 'W' at (3,0), got %q", cell3.Ch)
	}
	if cell3.FG != "#00FF00" {
		t.Errorf("expected FG=#00FF00 at (3,0), got %q", cell3.FG)
	}
	if !cell3.Bold {
		t.Errorf("expected Bold=true at (3,0)")
	}

	cell7 := buf.Get(7, 0)
	if cell7.Ch != 'd' {
		t.Errorf("expected 'd' at (7,0), got %q", cell7.Ch)
	}
	if cell7.FG != "#00FF00" {
		t.Errorf("expected FG=#00FF00 at (7,0), got %q", cell7.FG)
	}
	if !cell7.Bold {
		t.Errorf("expected Bold=true at (7,0)")
	}
}

func TestSpans_InheritNodeStyle(t *testing.T) {
	// Span without foreground should inherit from node style
	node := &Node{
		Type: "text",
		X:    0, Y: 0, W: 10, H: 1,
		Spans: []Span{
			{Text: "AB"},
			{Text: "CD", Foreground: "#FF0000"},
		},
		Style: Style{
			Foreground: "#AAAAAA",
			Bold:       true,
			Right:      -1,
			Bottom:     -1,
		},
	}

	buf := NewCellBuffer(10, 1)
	paintText(buf, node)

	// "A" should inherit node's foreground and bold
	cell0 := buf.Get(0, 0)
	if cell0.FG != "#AAAAAA" {
		t.Errorf("expected inherited FG=#AAAAAA, got %q", cell0.FG)
	}
	if !cell0.Bold {
		t.Errorf("expected inherited Bold=true")
	}

	// "C" should have span's foreground but inherit bold
	cell2 := buf.Get(2, 0)
	if cell2.FG != "#FF0000" {
		t.Errorf("expected FG=#FF0000, got %q", cell2.FG)
	}
	if !cell2.Bold {
		t.Errorf("expected inherited Bold=true for span without bold override")
	}
}

func TestSpans_Measurement(t *testing.T) {
	// Text node with spans should measure width from concatenated span texts
	node := &Node{
		Type: "text",
		Spans: []Span{
			{Text: "Hello"},
			{Text: " "},
			{Text: "World"},
		},
		Style: Style{
			WhiteSpace: "nowrap",
			Right:      -1,
			Bottom:     -1,
		},
	}

	w, h := measureText(node, 100, 0, 0)
	if w != 11 {
		t.Errorf("expected width=11, got %d", w)
	}
	if h != 1 {
		t.Errorf("expected height=1, got %d", h)
	}
}

func TestSpans_Wrapping(t *testing.T) {
	// Spans should wrap correctly
	node := &Node{
		Type: "text",
		X:    0, Y: 0, W: 5, H: 3,
		Spans: []Span{
			{Text: "AAABB", Foreground: "#FF0000"},
			{Text: "BCCCC", Foreground: "#00FF00"},
		},
		Style: Style{
			Foreground: "#FFFFFF",
			Right:      -1,
			Bottom:     -1,
		},
	}

	buf := NewCellBuffer(5, 3)
	paintText(buf, node)

	// Row 0: "AAABB" (first 5 chars)
	cell0 := buf.Get(0, 0)
	if cell0.Ch != 'A' || cell0.FG != "#FF0000" {
		t.Errorf("(0,0) expected A/#FF0000, got %c/%s", cell0.Ch, cell0.FG)
	}
	cell3 := buf.Get(3, 0)
	if cell3.Ch != 'B' || cell3.FG != "#FF0000" {
		t.Errorf("(3,0) expected B/#FF0000, got %c/%s", cell3.Ch, cell3.FG)
	}
	cell4 := buf.Get(4, 0)
	if cell4.Ch != 'B' || cell4.FG != "#FF0000" {
		t.Errorf("(4,0) expected B/#FF0000, got %c/%s", cell4.Ch, cell4.FG)
	}

	// Row 1: "BCCCC" — first char 'B' from first span (idx 5 of "AAABB" wait no...)
	// Actually: "AAABB" is 5 chars from first span, "BCCCC" is 5 chars from second span
	// Wrapping at width=5: row0="AAABB", row1="BCCCC"
	cell5 := buf.Get(0, 1)
	if cell5.Ch != 'B' || cell5.FG != "#00FF00" {
		t.Errorf("(0,1) expected B/#00FF00, got %c/%s", cell5.Ch, cell5.FG)
	}
	cell9 := buf.Get(4, 1)
	if cell9.Ch != 'C' || cell9.FG != "#00FF00" {
		t.Errorf("(4,1) expected C/#00FF00, got %c/%s", cell9.Ch, cell9.FG)
	}
}

func TestSpans_Ellipsis(t *testing.T) {
	// Spans with nowrap + ellipsis should truncate
	node := &Node{
		Type: "text",
		X:    0, Y: 0, W: 5, H: 1,
		Spans: []Span{
			{Text: "ABCDEFGH", Foreground: "#FF0000"},
		},
		Style: Style{
			WhiteSpace:   "nowrap",
			TextOverflow: "ellipsis",
			Right:        -1,
			Bottom:       -1,
		},
	}

	buf := NewCellBuffer(5, 1)
	paintText(buf, node)

	// Should render "ABCD…" (4 chars + ellipsis)
	cell0 := buf.Get(0, 0)
	if cell0.Ch != 'A' {
		t.Errorf("(0,0) expected 'A', got %q", cell0.Ch)
	}
	cell3 := buf.Get(3, 0)
	if cell3.Ch != 'D' {
		t.Errorf("(3,0) expected 'D', got %q", cell3.Ch)
	}
	cell4 := buf.Get(4, 0)
	if cell4.Ch != '…' {
		t.Errorf("(4,0) expected '…', got %q", cell4.Ch)
	}
}

func TestSpans_ContentIgnoredWhenSpansPresent(t *testing.T) {
	// When spans are present, Content should be ignored
	node := &Node{
		Type:    "text",
		Content: "IGNORED",
		X:       0, Y: 0, W: 10, H: 1,
		Spans: []Span{
			{Text: "USED"},
		},
		Style: Style{
			Right:  -1,
			Bottom: -1,
		},
	}

	buf := NewCellBuffer(10, 1)
	paintText(buf, node)

	cell0 := buf.Get(0, 0)
	if cell0.Ch != 'U' {
		t.Errorf("expected 'U' at (0,0) from spans, got %q", cell0.Ch)
	}
}

func boolPtr(b bool) *bool {
	return &b
}
