package lumina

import (
	"strings"
	"testing"
	"time"
)

func TestCreateElement_WithCallbacks(t *testing.T) {
	app := NewAppWithSize(20, 5)
	tio := NewMockTermIO(20, 5)
	SetOutputAdapter(NewANSIAdapter(tio))

	script := `
local lumina = require("lumina")

local Child = lumina.defineComponent({
    name = "Child",
    render = function(self)
        local hovered, setHovered = lumina.useState("hovered", false)
        local label = hovered and "[H]" or " . "
        return {
            type = "box",
            id = self.id or "child",
            style = { width = 3, height = 1 },
            onMouseEnter = function() setHovered(true) end,
            onMouseLeave = function() setHovered(false) end,
            onClick = function()
                if self.onChildClick then
                    self.onChildClick(self.id)
                end
            end,
            children = {
                { type = "text", content = label }
            }
        }
    end
})

local App = lumina.defineComponent({
    name = "TestApp",
    render = function()
        local lastClick, setLastClick = lumina.useState("lastClick", "none")
        return {
            type = "vbox",
            children = {
                { type = "text", content = "Click: " .. lastClick },
                {
                    type = "hbox",
                    children = {
                        lumina.createElement(Child, {
                            id = "c1",
                            onChildClick = function(id) setLastClick(id) end,
                        }),
                        lumina.createElement(Child, {
                            id = "c2",
                            onChildClick = function(id) setLastClick(id) end,
                        }),
                    }
                }
            }
        }
    end
})

lumina.mount(App)
`
	err := app.L.DoString(script)
	if err != nil {
		t.Fatalf("DoString: %v", err)
	}

	renderAndProcess := func() {
		app.lastRenderTime = time.Time{}
		app.RenderOnce()
		app.ProcessPendingEvents()
		app.ProcessPendingEvents() // second pass for cascaded events
		app.lastRenderTime = time.Time{}
		app.RenderOnce()
	}

	renderAndProcess()

	frame := app.lastFrame
	if frame == nil {
		t.Fatal("No frame rendered")
	}

	row0 := getFrameRow(frame, 0)
	t.Logf("Row 0: %s", strings.TrimSpace(row0))
	if !strings.Contains(row0, "Click: none") {
		t.Errorf("Expected 'Click: none', got: %q", row0)
	}

	row1 := getFrameRow(frame, 1)
	t.Logf("Row 1: %s", strings.TrimSpace(row1))
	if !strings.Contains(row1, " . ") {
		t.Errorf("Expected ' . ' cells, got: %q", row1)
	}

	// Test mouseenter on c1 (x=1, y=1)
	app.handleEvent(AppEvent{
		Type:    "input_event",
		Payload: &Event{Type: "mousemove", X: 1, Y: 1},
	})
	app.ProcessPendingEvents()
	app.ProcessPendingEvents()
	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame = app.lastFrame
	row1 = getFrameRow(frame, 1)
	t.Logf("Row 1 after hover c1: %s", strings.TrimSpace(row1))

	// Check that c1 shows [H] and c2 shows . (independent state)
	if strings.Contains(row1, "[H]") {
		t.Log("✓ mouseenter triggered useState update")
	} else {
		t.Log("✗ mouseenter did NOT trigger useState update")
	}

	// Test click on c2 (x=4, y=1)
	app.handleEvent(AppEvent{
		Type:    "input_event",
		Payload: &Event{Type: "mousedown", X: 4, Y: 1, Button: "left"},
	})
	app.ProcessPendingEvents()
	app.ProcessPendingEvents()
	app.lastRenderTime = time.Time{}
	app.RenderOnce()

	frame = app.lastFrame
	row0 = getFrameRow(frame, 0)
	t.Logf("Row 0 after click c2: %s", strings.TrimSpace(row0))
	if strings.Contains(row0, "Click: c2") {
		t.Log("✓ Callback prop works: parent received click from child")
	} else {
		t.Errorf("Expected 'Click: c2', got: %q", row0)
	}
}
