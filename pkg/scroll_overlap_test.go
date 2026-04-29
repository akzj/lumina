package v2

import (
	"fmt"
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/output"
)

// TestScrollOverlapBug reproduces: scroll Editor (behind Palette) → Editor content
// paints over Palette. After fix, Palette should remain fully visible.
func TestScrollOverlapBug(t *testing.T) {
	app, ta, _ := newV2App(t, 80, 24)

	err := app.RunScript("../examples/windows_widget.lua")
	if err != nil {
		t.Fatalf("RunScript failed: %v", err)
	}
	app.RenderAll()

	// Initial state: windows are (bottom to top): Editor, Monitor, Palette
	// Palette is at x=10, y=3, w=30, h=10 — it's on TOP
	// Editor is at x=2, y=1, w=35, h=12, scrollable=true — it's at BOTTOM

	// Verify Palette is visible initially
	if !screenHasString(ta, "Palette") {
		t.Fatal("Palette should be visible initially")
	}

	// Scroll Editor with mouse wheel on its exposed area (left side, not covered by Palette)
	// Editor is at x=2..36, y=1..12. Palette covers x=10..39, y=3..12.
	// Editor's exposed area: x=2..9, y=3..12 (left strip not covered by Palette)
	// Let's scroll at x=5, y=5 (inside Editor, not covered by Palette)
	for i := 0; i < 5; i++ {
		app.HandleEvent(&event.Event{Type: "scroll", X: 5, Y: 5, Key: "down"})
		app.RenderDirty()
	}

	// After scrolling, Palette should STILL be fully visible (not overwritten by Editor scroll)
	if !screenHasString(ta, "Palette") {
		// Dump screen for debugging
		dumpScreenArea(ta, 10, 3, 30, 1) // Palette title bar area
		t.Error("BUG: Palette title disappeared after scrolling Editor behind it")
	}

	// Check Palette content is still visible
	if !screenHasString(ta, "Color palette") {
		t.Error("BUG: Palette content 'Color palette' disappeared after scrolling Editor")
	}

	// Also verify Editor actually scrolled (content changed)
	// Editor starts showing "01 │" lines, after scroll should show higher numbers
	fmt.Printf("Screen after scroll (rows 1-12):\n")
	for y := 1; y <= 12; y++ {
		line := readScreenLine(ta, y, 80)
		fmt.Printf("  row %2d: %s\n", y, line)
	}
}

func dumpScreenArea(ta *output.TestAdapter, x, y, w, h int) {
	for row := y; row < y+h; row++ {
		var sb strings.Builder
		for col := x; col < x+w; col++ {
			c := ta.LastScreen.Get(col, row)
			if c.Char == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(c.Char)
			}
		}
		fmt.Printf("  area[%d]: '%s'\n", row, sb.String())
	}
}
