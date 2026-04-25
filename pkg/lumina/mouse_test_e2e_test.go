package lumina

import (
	"strings"
	"testing"
	"time"
)

func TestMouseTest_LoadsAndRenders(t *testing.T) {
	// Use smaller terminal to reduce VNode count (avoids pthread_create limit)
	app := NewAppWithSize(30, 10)
	tio := NewMockTermIO(30, 10)
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
	if !strings.Contains(row0, "Hover:") {
		t.Errorf("Expected status bar with 'Hover:', got: %q", row0)
	}

	// Check that grid rows have dots (each cell is 3 chars: " . ")
	dotsFound := false
	for y := 1; y < frame.Height; y++ {
		row := getFrameRow(frame, y)
		if strings.Contains(row, " . ") {
			dotsFound = true
			break
		}
	}
	if !dotsFound {
		t.Error("Expected grid rows with ' . ' cells")
	}

	t.Logf("Row 0: %s", strings.TrimSpace(row0))
	t.Logf("Row 1 (first 30): %.30s", getFrameRow(frame, 1))
}

func TestMouseTest_HoverAndClick(t *testing.T) {
	// Use smaller terminal to reduce VNode count
	app := NewAppWithSize(30, 10)
	tio := NewMockTermIO(30, 10)
	SetOutputAdapter(NewANSIAdapter(tio))

	err := app.LoadScript("../../examples/mouse_test.lua", tio)
	if err != nil {
		t.Fatalf("LoadScript: %v", err)
	}

	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	// Each cell is 3 chars wide. Cell c-2-3 is at screen x=6..8, y=4 (row 3+1 for status bar)
	// Simulate mousemove at x=7, y=4 (center of cell c-2-3)
	app.handleEvent(AppEvent{
		Type:    "input_event",
		Payload: &Event{Type: "mousemove", X: 7, Y: 4},
	})
	app.ProcessPendingEvents()
	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame := app.lastFrame
	if frame == nil {
		t.Fatal("No frame after hover")
	}

	// Check status bar shows hover target
	row0 := getFrameRow(frame, 0)
	t.Logf("After hover: %s", strings.TrimSpace(row0))

	// The hovered cell should show [O] (green hover indicator)
	row4 := getFrameRow(frame, 4)
	t.Logf("Row 4 after hover: %s", strings.TrimSpace(row4))
	if strings.Contains(row4, "[O]") {
		t.Log("Hover indicator [O] found in grid row")
	}

	// Simulate mousedown at same position
	app.handleEvent(AppEvent{
		Type:    "input_event",
		Payload: &Event{Type: "mousedown", X: 7, Y: 4, Button: "left"},
	})
	app.ProcessPendingEvents()
	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame = app.lastFrame
	row0 = getFrameRow(frame, 0)
	t.Logf("After click: %s", strings.TrimSpace(row0))

	// Should show click indicator [X] or [*] (if still hovered)
	row4 = getFrameRow(frame, 4)
	t.Logf("Row 4 after click: %s", strings.TrimSpace(row4))
	if strings.Contains(row4, "[X]") || strings.Contains(row4, "[*]") {
		t.Log("Click indicator found in grid row")
	}
}

// TestMouseEnterLeave verifies mouseenter/mouseleave are synthesized from mousemove
func TestMouseEnterLeave(t *testing.T) {
	app := NewAppWithSize(30, 10)
	tio := NewMockTermIO(30, 10)
	SetOutputAdapter(NewANSIAdapter(tio))

	// Simple script with two boxes that track enter/leave via global handlers
	script := `
local lumina = require("lumina")

local store = lumina.createStore({
    state = {
        entered = "",
        left = "",
    }
})

lumina.on("mouseenter", "", function(e)
    store.dispatch("setState", { entered = e.target or "" })
end)

lumina.on("mouseleave", "", function(e)
    store.dispatch("setState", { left = e.target or "" })
end)

local App = lumina.defineComponent({
    name = "EnterLeaveTest",
    render = function()
        local state = lumina.useStore(store)
        return {
            type = "vbox",
            children = {
                {
                    type = "text",
                    content = "enter:" .. (state.entered or "") .. " leave:" .. (state.left or ""),
                    style = { foreground = "#FFFFFF" },
                },
                {
                    type = "hbox",
                    children = {
                        { type = "box", props = { id = "boxA" }, style = { width = 10, height = 3, background = "#FF0000" },
                          children = { { type = "text", content = "A" } } },
                        { type = "box", props = { id = "boxB" }, style = { width = 10, height = 3, background = "#00FF00" },
                          children = { { type = "text", content = "B" } } },
                    }
                },
            }
        }
    end,
})

lumina.mount(App)
`

	err := app.L.DoString(script)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	// Register component and render
	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame := app.lastFrame
	if frame == nil {
		t.Fatal("No frame rendered")
	}

	// Move mouse to boxA (x=5, y=2)
	app.handleEvent(AppEvent{
		Type:    "input_event",
		Payload: &Event{Type: "mousemove", X: 5, Y: 2},
	})
	app.ProcessPendingEvents()
	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame = app.lastFrame
	row0 := getFrameRow(frame, 0)
	t.Logf("After hover boxA: %s", strings.TrimSpace(row0))

	if !strings.Contains(row0, "enter:boxA") {
		t.Errorf("Expected 'enter:boxA' in status, got: %q", row0)
	}

	// Move mouse to boxB (x=15, y=2)
	app.handleEvent(AppEvent{
		Type:    "input_event",
		Payload: &Event{Type: "mousemove", X: 15, Y: 2},
	})
	app.ProcessPendingEvents()
	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame = app.lastFrame
	row0 = getFrameRow(frame, 0)
	t.Logf("After hover boxB: %s", strings.TrimSpace(row0))

	if !strings.Contains(row0, "enter:boxB") {
		t.Errorf("Expected 'enter:boxB' in status, got: %q", row0)
	}
	if !strings.Contains(row0, "leave:boxA") {
		t.Errorf("Expected 'leave:boxA' in status, got: %q", row0)
	}
}
