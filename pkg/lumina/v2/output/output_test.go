package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
)

func makeTestBuffer() *buffer.Buffer {
	b := buffer.New(3, 2)
	b.Set(0, 0, buffer.Cell{Char: 'A', Foreground: "#ff0000", Background: "#000000", Bold: true})
	b.Set(1, 0, buffer.Cell{Char: 'B', Foreground: "#00ff00"})
	b.Set(2, 0, buffer.Cell{Char: 'C', Foreground: "#0000ff", Underline: true})
	b.Set(0, 1, buffer.Cell{Char: 'D'})
	b.Set(1, 1, buffer.Cell{Char: 'E', Dim: true})
	b.Set(2, 1, buffer.Cell{Char: 'F'})
	return b
}

func TestTUIAdapter_WriteFull(t *testing.T) {
	var buf bytes.Buffer
	adapter := NewTUIAdapter(&buf)

	screen := makeTestBuffer()
	if err := adapter.WriteFull(screen); err != nil {
		t.Fatalf("WriteFull error: %v", err)
	}
	if err := adapter.Flush(); err != nil {
		t.Fatalf("Flush error: %v", err)
	}

	out := buf.String()

	// Must contain cursor moves for both rows (1-based).
	if !strings.Contains(out, "\033[1;1H") {
		t.Error("missing cursor move to row 1, col 1")
	}
	if !strings.Contains(out, "\033[2;1H") {
		t.Error("missing cursor move to row 2, col 1")
	}

	// Must contain 24-bit foreground color for red (#ff0000).
	if !strings.Contains(out, "\033[38;2;255;0;0m") {
		t.Error("missing red foreground color escape")
	}

	// Must contain the characters.
	if !strings.Contains(out, "A") || !strings.Contains(out, "F") {
		t.Error("missing expected characters in output")
	}

	// Must contain bold escape.
	if !strings.Contains(out, "\033[1m") {
		t.Error("missing bold escape sequence")
	}
}

func TestTUIAdapter_WriteDirty(t *testing.T) {
	var buf bytes.Buffer
	adapter := NewTUIAdapter(&buf)

	screen := makeTestBuffer()
	dirty := []buffer.Rect{{X: 1, Y: 0, W: 2, H: 1}} // only cols 1-2 of row 0
	if err := adapter.WriteDirty(screen, dirty); err != nil {
		t.Fatalf("WriteDirty error: %v", err)
	}
	if err := adapter.Flush(); err != nil {
		t.Fatalf("Flush error: %v", err)
	}

	out := buf.String()

	// Should position cursor at row 1, col 2 (1-based).
	if !strings.Contains(out, "\033[1;2H") {
		t.Error("missing cursor move to row 1, col 2")
	}

	// Should contain B and C but NOT the cursor move for row 2.
	if !strings.Contains(out, "B") {
		t.Error("missing character B in dirty output")
	}
	if !strings.Contains(out, "C") {
		t.Error("missing character C in dirty output")
	}
	if strings.Contains(out, "\033[2;1H") {
		t.Error("should not contain cursor move for row 2 (not dirty)")
	}
}

func TestTUIAdapter_ColorOptimization(t *testing.T) {
	var buf bytes.Buffer
	adapter := NewTUIAdapter(&buf)

	// Create a buffer where all cells have the same foreground color.
	screen := buffer.New(3, 1)
	screen.Set(0, 0, buffer.Cell{Char: 'X', Foreground: "#aabbcc"})
	screen.Set(1, 0, buffer.Cell{Char: 'Y', Foreground: "#aabbcc"})
	screen.Set(2, 0, buffer.Cell{Char: 'Z', Foreground: "#aabbcc"})

	if err := adapter.WriteFull(screen); err != nil {
		t.Fatalf("WriteFull error: %v", err)
	}
	if err := adapter.Flush(); err != nil {
		t.Fatalf("Flush error: %v", err)
	}

	out := buf.String()

	// The color \033[38;2;170;187;204m should appear only ONCE,
	// not three times (optimization: skip if same as current).
	colorEsc := "\033[38;2;170;187;204m"
	count := strings.Count(out, colorEsc)
	if count != 1 {
		t.Errorf("expected color escape to appear once, got %d times", count)
	}
}

func TestJSONAdapter_WriteFull(t *testing.T) {
	var buf bytes.Buffer
	adapter := NewJSONAdapter(&buf)

	screen := makeTestBuffer()
	if err := adapter.WriteFull(screen); err != nil {
		t.Fatalf("WriteFull error: %v", err)
	}

	var result RenderResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON decode error: %v", err)
	}

	if result.Width != 3 || result.Height != 2 {
		t.Errorf("expected 3x2, got %dx%d", result.Width, result.Height)
	}

	if len(result.Cells) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Cells))
	}
	if len(result.Cells[0]) != 3 {
		t.Fatalf("expected 3 cols, got %d", len(result.Cells[0]))
	}

	// Check first cell.
	c := result.Cells[0][0]
	if c.Char != "A" {
		t.Errorf("expected char 'A', got %q", c.Char)
	}
	if c.Fg != "#ff0000" {
		t.Errorf("expected fg '#ff0000', got %q", c.Fg)
	}
	if c.Bg != "#000000" {
		t.Errorf("expected bg '#000000', got %q", c.Bg)
	}
	if !c.Bold {
		t.Error("expected bold=true")
	}

	// DirtyRects should be nil/empty for WriteFull.
	if len(result.DirtyRects) != 0 {
		t.Errorf("expected no dirty rects for WriteFull, got %d", len(result.DirtyRects))
	}
}

func TestJSONAdapter_DirtyRects(t *testing.T) {
	var buf bytes.Buffer
	adapter := NewJSONAdapter(&buf)

	screen := makeTestBuffer()
	dirty := []buffer.Rect{{X: 0, Y: 0, W: 2, H: 1}, {X: 1, Y: 1, W: 1, H: 1}}
	if err := adapter.WriteDirty(screen, dirty); err != nil {
		t.Fatalf("WriteDirty error: %v", err)
	}

	var result RenderResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON decode error: %v", err)
	}

	if len(result.DirtyRects) != 2 {
		t.Fatalf("expected 2 dirty rects, got %d", len(result.DirtyRects))
	}

	dr := result.DirtyRects[0]
	if dr.X != 0 || dr.Y != 0 || dr.W != 2 || dr.H != 1 {
		t.Errorf("dirty rect 0 mismatch: %+v", dr)
	}

	dr = result.DirtyRects[1]
	if dr.X != 1 || dr.Y != 1 || dr.W != 1 || dr.H != 1 {
		t.Errorf("dirty rect 1 mismatch: %+v", dr)
	}
}

func TestTestAdapter_Captures(t *testing.T) {
	adapter := NewTestAdapter()

	screen := makeTestBuffer()
	if err := adapter.WriteFull(screen); err != nil {
		t.Fatalf("WriteFull error: %v", err)
	}

	if adapter.WriteCount != 1 {
		t.Errorf("expected WriteCount=1, got %d", adapter.WriteCount)
	}

	if adapter.LastScreen == nil {
		t.Fatal("LastScreen is nil after WriteFull")
	}

	// Verify it's a clone, not the same pointer.
	if adapter.LastScreen == screen {
		t.Error("LastScreen should be a clone, not the same pointer")
	}

	// Verify content matches.
	c := adapter.LastScreen.Get(0, 0)
	if c.Char != 'A' {
		t.Errorf("expected char 'A', got %c", c.Char)
	}

	// Modify original — clone should be unaffected.
	screen.Set(0, 0, buffer.Cell{Char: 'Z'})
	c2 := adapter.LastScreen.Get(0, 0)
	if c2.Char != 'A' {
		t.Error("clone was affected by modification of original")
	}
}

func TestTestAdapter_DirtyRects(t *testing.T) {
	adapter := NewTestAdapter()

	screen := makeTestBuffer()
	dirty := []buffer.Rect{{X: 0, Y: 0, W: 1, H: 1}}
	if err := adapter.WriteDirty(screen, dirty); err != nil {
		t.Fatalf("WriteDirty error: %v", err)
	}

	if adapter.WriteCount != 1 {
		t.Errorf("expected WriteCount=1, got %d", adapter.WriteCount)
	}

	if len(adapter.DirtyRects) != 1 {
		t.Fatalf("expected 1 dirty rect, got %d", len(adapter.DirtyRects))
	}

	dr := adapter.DirtyRects[0]
	if dr.X != 0 || dr.Y != 0 || dr.W != 1 || dr.H != 1 {
		t.Errorf("dirty rect mismatch: %+v", dr)
	}

	// WriteFull should clear DirtyRects.
	if err := adapter.WriteFull(screen); err != nil {
		t.Fatalf("WriteFull error: %v", err)
	}
	if adapter.DirtyRects != nil {
		t.Error("DirtyRects should be nil after WriteFull")
	}
	if adapter.WriteCount != 2 {
		t.Errorf("expected WriteCount=2, got %d", adapter.WriteCount)
	}
}
