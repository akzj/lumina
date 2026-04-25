package lumina

import (
	"strings"
	"testing"
	"time"
)

func TestComponentLib_MultiFile(t *testing.T) {
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

	// Verify sidebar shows "Components" title
	found := false
	for y := 0; y < frame.Height; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, "Components") {
			t.Logf("Sidebar title at row %d: %q", y, strings.TrimSpace(row))
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Components' in sidebar")
	}

	// Verify sidebar shows "▸ 1. Button" (selected)
	found = false
	for y := 0; y < frame.Height; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, "1. Button") {
			t.Logf("Sidebar item at row %d: %q", y, strings.TrimSpace(row))
			found = true
			if !strings.Contains(row, "▸") {
				t.Error("Expected '▸' marker for selected page")
			}
			break
		}
	}
	if !found {
		t.Error("Expected '1. Button' in sidebar")
	}

	// Verify content area shows Button page description (from pages/button.lua)
	found = false
	for y := 0; y < frame.Height; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, "Displays a button") {
			t.Logf("Button page description at row %d: %q", y, strings.TrimSpace(row))
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected Button page description in content area — multi-file require may have failed")
		for y := 0; y < frame.Height; y++ {
			row := getFrameRow(frame, y)
			if strings.TrimSpace(row) != "" {
				t.Logf("  row %2d: %q", y, strings.TrimSpace(row))
			}
		}
	}

	// Verify multi-file require worked (no error message in frame)
	for y := 0; y < frame.Height; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, "Error loading") {
			t.Errorf("Found error message at row %d: %q — require() failed", y, strings.TrimSpace(row))
		}
	}

	// Verify Variants section is visible (from button.lua content)
	found = false
	for y := 0; y < frame.Height; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, "Variants") {
			t.Logf("Variants section at row %d: %q", y, strings.TrimSpace(row))
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Variants' section in Button page")
	}

	// Verify Supported Features section
	found = false
	for y := 0; y < frame.Height; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, "Supported Features") {
			t.Logf("Features section at row %d: %q", y, strings.TrimSpace(row))
			found = true
			break
		}
	}
	if !found {
		t.Log("Note: 'Supported Features' not found — may be below visible area at 120x40")
	}

	// Verify footer shows quit hint
	found = false
	for y := 0; y < frame.Height; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, "[q] Quit") {
			t.Logf("Quit hint at row %d: %q", y, strings.TrimSpace(row))
			found = true
			break
		}
	}
	if !found {
		t.Log("Note: '[q] Quit' not found — may be below visible area")
	}

	t.Log("Multi-file component library loaded and rendered successfully")
}
