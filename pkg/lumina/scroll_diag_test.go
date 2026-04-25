package lumina

import (
	"strings"
	"testing"
	"time"
)

// TestComponentLib_ScrollWorks verifies that scrolling actually shifts visible content.
func TestComponentLib_ScrollWorks(t *testing.T) {
	app := NewAppWithSize(120, 40)
	tio := NewMockTermIO(120, 40)
	SetOutputAdapter(NewANSIAdapter(tio))

	err := app.LoadScript("../../examples/components/main.lua", tio)
	if err != nil {
		t.Fatalf("LoadScript: %v", err)
	}

	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame := app.lastFrame
	if frame == nil {
		t.Fatal("No frame rendered")
	}

	// Verify viewport was created with correct dimensions
	vp := GetViewport("content-scroll")
	if vp.ContentH == 0 {
		t.Fatal("Viewport 'content-scroll' has ContentH=0 — not created")
	}
	t.Logf("Viewport: ContentH=%d, ViewH=%d, NeedsScroll=%v", vp.ContentH, vp.ViewH, vp.NeedsScroll())

	if !vp.NeedsScroll() {
		t.Error("Expected NeedsScroll=true (content taller than viewport)")
	}

	// Verify scrollbar is rendered (█ or │ on right edge)
	scrollbarFound := false
	for y := 0; y < frame.Height; y++ {
		cell := frame.Cells[y][frame.Width-1]
		if cell.Char == '█' || cell.Char == '│' {
			scrollbarFound = true
			break
		}
	}
	if !scrollbarFound {
		t.Error("Expected scrollbar on right edge")
	}

	// Verify "Supported Features" is NOT visible before scrolling
	for y := 0; y < frame.Height; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, "Supported Features") {
			t.Log("'Supported Features' visible before scroll — content fits in viewport")
			return // No scroll test needed
		}
	}
	t.Log("'Supported Features' not visible before scroll — testing scroll...")

	// Scroll down by 20 lines using luaScrollBy
	err = app.L.DoString(`lumina.scrollBy("content-scroll", 20)`)
	if err != nil {
		t.Fatalf("scrollBy failed: %v", err)
	}

	// Re-render
	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame = app.lastFrame
	if frame == nil {
		t.Fatal("No frame after scroll re-render")
	}

	// After scrolling, "Supported Features" should now be visible
	supportedFound := false
	for y := 0; y < frame.Height; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, "Supported Features") {
			supportedFound = true
			t.Logf("'Supported Features' visible at row %d after scroll", y)
			break
		}
	}
	if !supportedFound {
		t.Error("'Supported Features' should be visible after scrolling down 20 lines")
	}

	// "Button" title (at original row 2) should now be scrolled off-screen
	buttonVisible := false
	for y := 0; y < 5; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, "Button") {
			buttonVisible = true
			break
		}
	}
	if buttonVisible {
		t.Error("'Button' title should be scrolled off-screen after scrolling 20 lines")
	}

	// Verify scroll position updated
	vp = GetViewport("content-scroll")
	t.Logf("After scroll: ScrollY=%d", vp.ScrollY)
	if vp.ScrollY == 0 {
		t.Error("Expected ScrollY > 0 after scrolling down")
	}
}
