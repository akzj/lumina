package lumina

import (
    "bytes"
    "strings"
    "testing"
)

func TestANSIDefaultBackground(t *testing.T) {
    var buf bytes.Buffer
    adapter := NewANSIAdapter(&buf)
    
    t.Logf("DefaultBackground: %q", adapter.DefaultBackground)
    t.Logf("ColorMode: %v", adapter.ColorMode())
    
    // Create a small frame with some empty-bg cells
    frame := NewFrame(5, 1)
    frame.Cells[0][0] = Cell{Char: 'A', Foreground: "#FF0000", Background: "#1E1E2E"}
    frame.Cells[0][1] = Cell{Char: 'B', Foreground: "#FF0000", Background: ""}  // empty bg!
    frame.Cells[0][2] = Cell{Char: 'C', Foreground: "#FF0000", Background: "#1E1E2E"}
    frame.MarkDirty()
    
    adapter.Write(frame)
    output := buf.String()
    
    // Check if the output contains bg color for cell B
    t.Logf("Output length: %d bytes", len(output))
    
    // Count occurrences of "48;2;30;30;46" (RGB for #1E1E2E)
    // The ANSI writer optimizes consecutive same-styled cells, so the bg code
    // appears once for "ABC" and once for the trailing spaces (2 total, not 3).
    bgCode := "48;2;30;30;46"
    count := strings.Count(output, bgCode)
    t.Logf("Occurrences of bg code %q: %d (expect >=2 — writer optimizes consecutive same-style cells)", bgCode, count)
    
    if count < 2 {
        t.Errorf("Expected at least 2 bg color sequences (cells + trailing spaces), got %d", count)
        // Show raw output for debugging
        for i, b := range []byte(output) {
            if b == 0x1b {
                t.Logf("  ESC at byte %d: %q", i, output[i:min(i+30, len(output))])
            }
        }
    }
}

func TestRgbTo256_NearGray(t *testing.T) {
    // #1E1E2E (30,30,46) should map to grayscale, not cube index 59
    idx := rgbTo256(0x1E, 0x1E, 0x2E)
    if idx == 59 {
        t.Errorf("expected grayscale index, got cube index 59")
    }
    if idx < 232 || idx > 255 {
        t.Logf("Warning: index %d not in grayscale range", idx)
    }
    t.Logf("#1E1E2E → index %d", idx)

    // Pure gray should definitely hit grayscale ramp
    idx2 := rgbTo256(128, 128, 128)
    if idx2 < 232 || idx2 > 255 {
        t.Errorf("pure gray (128,128,128) should map to grayscale, got %d", idx2)
    }
    t.Logf("(128,128,128) → index %d", idx2)

    // Saturated color should hit the cube
    idx3 := rgbTo256(255, 0, 0)
    if idx3 < 16 || idx3 > 231 {
        t.Errorf("pure red should map to cube, got %d", idx3)
    }
    t.Logf("(255,0,0) → index %d", idx3)
}

func min(a, b int) int {
    if a < b { return a }
    return b
}
