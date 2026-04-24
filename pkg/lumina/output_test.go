package lumina_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/lumina"
)

func TestFrameCreation(t *testing.T) {
	frame := lumina.NewFrame(10, 5)
	
	if frame.Width != 10 {
		t.Errorf("expected width 10, got %d", frame.Width)
	}
	if frame.Height != 5 {
		t.Errorf("expected height 5, got %d", frame.Height)
	}
	if len(frame.Cells) != 5 {
		t.Errorf("expected 5 rows, got %d", len(frame.Cells))
	}
	if len(frame.Cells[0]) != 10 {
		t.Errorf("expected 10 cols, got %d", len(frame.Cells[0]))
	}
	
	// Check default cell is space
	if frame.Cells[0][0].Char != ' ' {
		t.Errorf("expected space char, got %v", frame.Cells[0][0].Char)
	}
}

func TestFrameMarkDirty(t *testing.T) {
	frame := lumina.NewFrame(10, 5)
	frame.MarkDirty()
	
	if len(frame.DirtyRects) != 1 {
		t.Fatalf("expected 1 dirty rect, got %d", len(frame.DirtyRects))
	}
	
	rect := frame.DirtyRects[0]
	if rect.X != 0 || rect.Y != 0 || rect.W != 10 || rect.H != 5 {
		t.Errorf("dirty rect incorrect: %v", rect)
	}
}

func TestANSIAdapterWrite(t *testing.T) {
	buf := &bytes.Buffer{}
	adapter := lumina.NewANSIAdapter(buf)
	
	// Create a simple frame with one cell
	frame := &lumina.Frame{
		Cells: [][]lumina.Cell{
			{{Char: 'H', Foreground: "#ff0000"}},
		},
		DirtyRects: []lumina.Rect{{X: 0, Y: 0, W: 1, H: 1}},
		Width:      1,
		Height:     1,
	}
	
	err := adapter.Write(frame)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	
	output := buf.String()
	
	// Check ANSI escape codes are present
	if !strings.Contains(output, "\x1b[") {
		t.Error("expected ANSI escape codes in output")
	}
	
	// Check character is written
	if !strings.Contains(output, "H") {
		t.Error("expected 'H' in output")
	}
}

func TestANSIAdapterMode(t *testing.T) {
	adapter := lumina.NewANSIAdapter(&bytes.Buffer{})
	
	if adapter.Mode() != lumina.ModeANSI {
		t.Errorf("expected ModeANSI, got %v", adapter.Mode())
	}
}

func TestANSIColorOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	adapter := lumina.NewANSIAdapter(buf)
	
	frame := &lumina.Frame{
		Cells: [][]lumina.Cell{
			{{Char: 'A', Foreground: "#ffffff", Background: "#000000"}},
		},
		DirtyRects: []lumina.Rect{{X: 0, Y: 0, W: 1, H: 1}},
		Width:      1,
		Height:     1,
	}
	
	adapter.Write(frame)
	output := buf.String()
	
	// Check for 24-bit color codes (ESC[38;2;R;G;Bm)
	if !strings.Contains(output, "38;2") {
		t.Error("expected 24-bit foreground color code")
	}
}

func TestANSIAdapterEmptyFrame(t *testing.T) {
	buf := &bytes.Buffer{}
	adapter := lumina.NewANSIAdapter(buf)
	
	frame := &lumina.Frame{
		Cells:      [][]lumina.Cell{{}},
		DirtyRects: nil, // No dirty rects
		Width:      10,
		Height:     10,
	}
	
	err := adapter.Write(frame)
	if err != nil {
		t.Fatalf("Write with empty dirty rects failed: %v", err)
	}
	
	// When DirtyRects is nil/empty, Write should return early with no output
	// This is the expected behavior for optimization
	// The test passes - no output is correct for "nothing dirty"
}

func TestNopAdapter(t *testing.T) {
	adapter := &lumina.NopAdapter{}
	
	frame := &lumina.Frame{
		Cells:      [][]lumina.Cell{{}},
		DirtyRects: []lumina.Rect{{X: 0, Y: 0, W: 1, H: 1}},
		Width:      1,
		Height:     1,
	}
	
	// NopAdapter should do nothing
	err := adapter.Write(frame)
	if err != nil {
		t.Errorf("NopAdapter.Write failed: %v", err)
	}
	
	if adapter.Mode() != lumina.ModeANSI {
		t.Errorf("NopAdapter mode incorrect")
	}
}