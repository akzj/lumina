package render

import (
	"testing"
)

func TestHighlight_Go(t *testing.T) {
	code := `package main

func main() {
	println("hello")
}`
	lines := Highlight(code, "go", "monokai")
	if len(lines) == 0 {
		t.Fatal("expected non-empty lines")
	}
	// First line should have "package" keyword
	if len(lines[0].Spans) == 0 {
		t.Fatal("expected spans on first line")
	}
	found := false
	for _, span := range lines[0].Spans {
		if span.Text == "package" {
			found = true
			if span.Foreground == "" {
				t.Error("expected foreground color for keyword 'package'")
			}
			break
		}
	}
	if !found {
		t.Errorf("expected 'package' keyword in first line spans, got: %+v", lines[0].Spans)
	}
}

func TestHighlight_EmptyCode(t *testing.T) {
	lines := Highlight("", "go", "")
	// Empty code should return empty or minimal result
	if len(lines) > 1 {
		t.Errorf("expected 0 or 1 lines for empty code, got %d", len(lines))
	}
}

func TestHighlight_UnknownLanguage(t *testing.T) {
	code := "hello world"
	lines := Highlight(code, "nonexistent_language_xyz", "")
	// Should fallback gracefully
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
}

func TestHighlight_NilLanguage(t *testing.T) {
	code := "x = 1\ny = 2\n"
	lines := Highlight(code, "", "")
	if len(lines) < 2 {
		t.Errorf("expected at least 2 lines, got %d", len(lines))
	}
}

func TestHighlight_MultiLine(t *testing.T) {
	code := "line1\nline2\nline3"
	lines := Highlight(code, "", "")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}
