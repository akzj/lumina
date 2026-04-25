package lumina

import (
	"strings"
	"testing"
	"time"
)

func TestMouseTest_LoadsAndRenders(t *testing.T) {
	app := NewAppWithSize(120, 40)
	tio := NewMockTermIO(120, 40)
	SetOutputAdapter(NewANSIAdapter(tio))

	err := app.LoadScript("../../examples/mouse_test.lua", tio)
	if err != nil {
		t.Fatalf("LoadScript: %v", err)
	}

	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame := app.lastFrame
	if frame == nil {
		t.Fatal("No frame rendered")
	}

	// Check status bar at row 0
	row0 := getFrameRow(frame, 0)
	if !strings.Contains(row0, "Mouse:") {
		t.Errorf("Expected status bar with 'Mouse:', got: %q", row0)
	}

	// Check that canvas rows have dots (ASCII '.')
	dotsFound := false
	for y := 1; y < 10; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, "...") {
			dotsFound = true
			break
		}
	}
	if !dotsFound {
		t.Error("Expected canvas rows with '.' dots")
	}

	t.Logf("Row 0: %s", strings.TrimSpace(getFrameRow(frame, 0)))
	t.Logf("Row 1 (first 60): %.60s...", getFrameRow(frame, 1))
}

func TestMouseTest_HoverAndClick(t *testing.T) {
	app := NewAppWithSize(120, 40)
	tio := NewMockTermIO(120, 40)
	SetOutputAdapter(NewANSIAdapter(tio))

	err := app.LoadScript("../../examples/mouse_test.lua", tio)
	if err != nil {
		t.Fatalf("LoadScript: %v", err)
	}

	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	// Simulate mousemove at (50, 10)
	app.handleEvent(AppEvent{
		Type:    "input_event",
		Payload: &Event{Type: "mousemove", X: 50, Y: 10},
	})
	// Process lua_callback events posted by event handlers
	app.ProcessPendingEvents()
	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame := app.lastFrame
	if frame == nil {
		t.Fatal("No frame after hover")
	}

	// Check status bar shows hover coordinates
	row0 := getFrameRow(frame, 0)
	if !strings.Contains(row0, "50") || !strings.Contains(row0, "10") {
		t.Errorf("Expected hover coords (50, 10) in status bar, got: %q", row0)
	}

	// Simulate mousedown at (30, 5)
	app.handleEvent(AppEvent{
		Type:    "input_event",
		Payload: &Event{Type: "mousedown", X: 30, Y: 5, Button: "left"},
	})
	app.ProcessPendingEvents()
	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame = app.lastFrame
	row0 = getFrameRow(frame, 0)
	if !strings.Contains(row0, "30") {
		t.Errorf("Expected click X=30 in status bar, got: %q", row0)
	}
	if !strings.Contains(row0, "Total: 1") {
		t.Errorf("Expected 'Total: 1' in status bar, got: %q", row0)
	}
}
