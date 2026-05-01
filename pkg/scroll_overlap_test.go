package v2

import (
	"testing"

	"github.com/akzj/lumina/pkg/event"
)

// TestScrollOverlapBug reproduces: scroll a box (behind another) → content
// paints over the front box. After fix, front box should remain fully visible.
func TestScrollOverlapBug(t *testing.T) {
	app, ta, _ := newV2App(t, 80, 24)

	err := app.RunString(`
		lumina.createComponent({
			id = "test", name = "Test",
			render = function()
				local lines = {}
				for i = 1, 30 do
					lines[i] = lumina.createElement("text", {key = "l"..i}, string.format("%02d | Line content here", i))
				end
				return lumina.createElement("box", {
					style = {width = 80, height = 24, background = "#000000"}},
					-- Back box: scrollable editor-like content
					lumina.createElement("vbox", {
						id = "editor",
						style = {
							position = "absolute",
							left = 2, top = 1,
							width = 35, height = 12,
							overflow = "scroll",
							background = "#222222",
						},
					}, table.unpack(lines)),
					-- Front box: overlapping "palette" on top
					lumina.createElement("vbox", {
						id = "palette",
						style = {
							position = "absolute",
							left = 10, top = 3,
							width = 30, height = 10,
							background = "#444444",
						},
					},
						lumina.createElement("text", {}, "Palette"),
						lumina.createElement("text", {}, "Color palette content")
					)
				)
			end,
		})
	`)
	if err != nil {
		t.Fatalf("RunString failed: %v", err)
	}
	app.RenderAll()

	// Verify Palette is visible initially
	if !screenHasString(ta, "Palette") {
		t.Fatal("Palette should be visible initially")
	}

	// Scroll the editor area (behind the palette)
	for i := 0; i < 5; i++ {
		app.HandleEvent(&event.Event{Type: "scroll", X: 5, Y: 5, Key: "down"})
		app.RenderDirty()
	}

	// After scrolling, Palette should STILL be fully visible
	if !screenHasString(ta, "Palette") {
		t.Error("BUG: Palette title disappeared after scrolling editor behind it")
	}

	if !screenHasString(ta, "Color palette") {
		t.Error("BUG: Palette content disappeared after scrolling editor")
	}
}
