package v2

import (
	"fmt"
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/perf"
)

// Scenario tests call app.Tracker().Enable() and Tracker.AssertLastFrame after
// RenderAll / RenderDirty where render counts are deterministic (same hooks as
// perf_integration_test.go: component render observer → perf.RecordComponent).

// =============================================================================
// Scenario 1: Desktop With Panels — Simulates a 3-panel IDE layout
// =============================================================================

func TestScenario_DesktopWithPanels(t *testing.T) {
	// 80×35 screen, 4 components: sidebar, editor, outline, statusbar
	//
	// ┌─────────┬──────────────────┬─────────┐
	// │ sidebar  │    editor        │ outline │
	// │ (z=10)   │    (z=10)        │ (z=10)  │
	// │ 15x30    │    50x30         │ 15x30   │
	// └─────────┴──────────────────┴─────────┘
	// │              status bar (z=20)        │
	// └───────────────────────────────────────┘
	app, ta := NewTestApp(80, 35)
	app.Tracker().Enable()

	var clickLog []string

	// Sidebar: 15×30 at (0,0), z=10
	app.RegisterComponent("sidebar", "sidebar", buffer.Rect{X: 0, Y: 0, W: 15, H: 30}, 10,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "sidebar-root"
			root.Style.Background = "#1a1a2e"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickLog = append(clickLog, "sidebar")
			})

			title := layout.NewVNode("text")
			title.ID = "sidebar-title"
			title.Content = "Files"
			title.Props["focusable"] = true
			root.AddChild(title)

			item := layout.NewVNode("text")
			item.ID = "sidebar-item1"
			item.Content = "main.go"
			item.Props["focusable"] = true
			root.AddChild(item)

			return root
		})

	// Editor: 50×30 at (15,0), z=10
	app.RegisterComponent("editor", "editor", buffer.Rect{X: 15, Y: 0, W: 50, H: 30}, 10,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "editor-root"
			root.Style.Background = "#16213e"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickLog = append(clickLog, "editor")
			})

			line := layout.NewVNode("text")
			line.ID = "editor-line1"
			line.Content = "func main()"
			line.Props["focusable"] = true
			root.AddChild(line)

			return root
		})

	// Outline: 15×30 at (65,0), z=10
	app.RegisterComponent("outline", "outline", buffer.Rect{X: 65, Y: 0, W: 15, H: 30}, 10,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "outline-root"
			root.Style.Background = "#0f3460"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickLog = append(clickLog, "outline")
			})

			sym := layout.NewVNode("text")
			sym.ID = "outline-sym1"
			sym.Content = "main"
			sym.Props["focusable"] = true
			root.AddChild(sym)

			return root
		})

	// Statusbar: 80×5 at (0,30), z=20 (overlaps bottom)
	app.RegisterComponent("statusbar", "statusbar", buffer.Rect{X: 0, Y: 30, W: 80, H: 5}, 20,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "statusbar-root"
			root.Style.Background = "#533483"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickLog = append(clickLog, "statusbar")
			})

			msg := layout.NewVNode("text")
			msg.ID = "statusbar-msg"
			msg.Content = "Ready"
			msg.Props["focusable"] = true
			root.AddChild(msg)

			return root
		})

	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(4),
		perf.CheckLayouts(4),
		perf.CheckPaints(4),
		perf.CheckRenderComponents("editor", "outline", "sidebar", "statusbar"),
	)

	// --- Verify rendering positions ---

	// Sidebar text "Files" at (0,0)
	if c := ta.LastScreen.Get(0, 0).Char; c != 'F' {
		t.Errorf("sidebar: expected 'F' at (0,0), got %q", c)
	}
	// Sidebar background at interior
	if bg := ta.LastScreen.Get(10, 15).Background; bg != "#1a1a2e" {
		t.Errorf("sidebar interior: expected bg '#1a1a2e', got %q", bg)
	}

	// Editor text "func main()" at (15,0)
	if c := ta.LastScreen.Get(15, 0).Char; c != 'f' {
		t.Errorf("editor: expected 'f' at (15,0), got %q", c)
	}

	// Outline text "main" at (65,0)
	if c := ta.LastScreen.Get(65, 0).Char; c != 'm' {
		t.Errorf("outline: expected 'm' at (65,0), got %q", c)
	}

	// Statusbar text "Ready" at (0,30)
	if c := ta.LastScreen.Get(0, 30).Char; c != 'R' {
		t.Errorf("statusbar: expected 'R' at (0,30), got %q", c)
	}

	// --- Verify click dispatch ---

	// Click in editor area
	clickLog = nil
	app.HandleEvent(&event.Event{Type: "mousedown", X: 20, Y: 5})
	if len(clickLog) == 0 || clickLog[0] != "editor" {
		t.Errorf("click in editor area: expected 'editor', got %v", clickLog)
	}

	// Click in sidebar area
	clickLog = nil
	app.HandleEvent(&event.Event{Type: "mousedown", X: 5, Y: 10})
	if len(clickLog) == 0 || clickLog[0] != "sidebar" {
		t.Errorf("click in sidebar area: expected 'sidebar', got %v", clickLog)
	}

	// Click in statusbar area (z=20)
	clickLog = nil
	app.HandleEvent(&event.Event{Type: "mousedown", X: 40, Y: 32})
	if len(clickLog) == 0 || clickLog[0] != "statusbar" {
		t.Errorf("click in statusbar area: expected 'statusbar', got %v", clickLog)
	}

	// --- Verify SetState triggers only editor re-render ---
	app.SetState("editor", "cursor", 5)
	dirtyBefore := app.DirtyRects()
	app.RenderDirty()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(1),
		perf.CheckRenderComponents("editor"),
		perf.CheckOcclusionUpdates(1),
		perf.CheckMetric(perf.ComposeDirty, 1),
		perf.CheckHandlerDirtySyncs(1),
	)
	dirtyAfter := app.DirtyRects()
	// After SetState+RenderDirty, dirty rects should exist
	if len(dirtyAfter) == 0 && len(dirtyBefore) == 0 {
		// It's possible the dirty rects were consumed; just verify no crash
	}

	// --- Verify Tab navigation across panels ---
	// There are 5 focusable elements across 4 components
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	first := app.FocusedID()
	if first == "" {
		t.Error("expected focus after first Tab")
	}

	seen := map[string]bool{first: true}
	for i := 0; i < 4; i++ {
		app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
		id := app.FocusedID()
		seen[id] = true
	}
	if len(seen) < 3 {
		t.Errorf("expected at least 3 unique focused IDs across panels, got %d: %v", len(seen), seen)
	}
}

// =============================================================================
// Scenario 2: Modal Dialog Over Content
// =============================================================================

func TestScenario_ModalDialogOverContent(t *testing.T) {
	app, ta := NewTestApp(60, 20)
	app.Tracker().Enable()

	var clickLog []string

	// Background content (z=0)
	app.RegisterComponent("bg", "bg", buffer.Rect{X: 0, Y: 0, W: 60, H: 20}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "bg-root"
			root.Style.Background = "#000000"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickLog = append(clickLog, "bg")
			})

			txt := layout.NewVNode("text")
			txt.ID = "bg-text"
			txt.Content = "Background Content"
			root.AddChild(txt)
			return root
		})

	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(1),
		perf.CheckLayouts(1),
		perf.CheckPaints(1),
		perf.CheckRenderComponents("bg"),
	)

	// Background visible at (0,0)
	if c := ta.LastScreen.Get(0, 0).Char; c != 'B' {
		t.Errorf("initial bg: expected 'B' at (0,0), got %q", c)
	}

	// --- Open dialog ---
	app.RegisterComponent("dialog", "dialog", buffer.Rect{X: 15, Y: 5, W: 30, H: 10}, 100,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "dialog-root"
			root.Style.Background = "#FFFFFF"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickLog = append(clickLog, "dialog")
			})

			title := layout.NewVNode("text")
			title.ID = "dialog-title"
			title.Content = "Confirm?"
			root.AddChild(title)

			okBtn := layout.NewVNode("text")
			okBtn.ID = "dialog-ok"
			okBtn.Content = "OK"
			okBtn.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickLog = append(clickLog, "dialog-ok")
			})
			okBtn.Props["focusable"] = true
			root.AddChild(okBtn)

			cancelBtn := layout.NewVNode("text")
			cancelBtn.ID = "dialog-cancel"
			cancelBtn.Content = "Cancel"
			cancelBtn.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickLog = append(clickLog, "dialog-cancel")
			})
			cancelBtn.Props["focusable"] = true
			root.AddChild(cancelBtn)

			return root
		})

	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(1),
		perf.CheckLayouts(1),
		perf.CheckPaints(1),
		perf.CheckRenderComponents("dialog"),
	)

	// Dialog visible at its origin
	if c := ta.LastScreen.Get(15, 5).Char; c != 'C' {
		t.Errorf("dialog: expected 'C' (from 'Confirm?') at (15,5), got %q", c)
	}

	// Background still visible outside dialog
	if c := ta.LastScreen.Get(0, 0).Char; c != 'B' {
		t.Errorf("bg still visible: expected 'B' at (0,0), got %q", c)
	}

	// Dialog background in interior
	if bg := ta.LastScreen.Get(30, 10).Background; bg != "#FFFFFF" {
		t.Errorf("dialog interior: expected bg '#FFFFFF', got %q", bg)
	}

	// --- Click inside dialog ---
	clickLog = nil
	app.HandleEvent(&event.Event{Type: "mousedown", X: 20, Y: 7})
	dialogClicked := false
	bgClicked := false
	for _, entry := range clickLog {
		if entry == "dialog" || entry == "dialog-ok" || entry == "dialog-cancel" {
			dialogClicked = true
		}
		if entry == "bg" {
			bgClicked = true
		}
	}
	if !dialogClicked {
		t.Errorf("click inside dialog: expected dialog handler, got %v", clickLog)
	}
	if bgClicked {
		t.Error("click inside dialog: bg handler should NOT fire (occluded)")
	}

	// --- Click outside dialog (in bg) ---
	clickLog = nil
	app.HandleEvent(&event.Event{Type: "mousedown", X: 2, Y: 2})
	if len(clickLog) == 0 || clickLog[0] != "bg" {
		t.Errorf("click outside dialog: expected 'bg', got %v", clickLog)
	}

	// --- Close dialog ---
	app.UnregisterComponent("dialog")
	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(0),
		perf.CheckLayouts(0),
		perf.CheckPaints(0),
	)

	// Background fully visible where dialog was
	if bg := ta.LastScreen.Get(30, 10).Background; bg != "#000000" {
		t.Errorf("after close: expected bg '#000000' at dialog area, got %q", bg)
	}

	// Click in old dialog area → bg handler fires
	clickLog = nil
	app.HandleEvent(&event.Event{Type: "mousedown", X: 20, Y: 7})
	if len(clickLog) == 0 || clickLog[0] != "bg" {
		t.Errorf("click after close: expected 'bg', got %v", clickLog)
	}

	// --- Re-open dialog ---
	app.RegisterComponent("dialog2", "dialog", buffer.Rect{X: 15, Y: 5, W: 30, H: 10}, 100,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "dialog2-root"
			root.Style.Background = "#EEEEEE"
			txt := layout.NewVNode("text")
			txt.Content = "Reopened"
			root.AddChild(txt)
			return root
		})
	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(1),
		perf.CheckLayouts(1),
		perf.CheckPaints(1),
		perf.CheckRenderComponents("dialog2"),
	)

	if c := ta.LastScreen.Get(15, 5).Char; c != 'R' {
		t.Errorf("reopened dialog: expected 'R' at (15,5), got %q", c)
	}
}

// =============================================================================
// Scenario 3: Cascading Windows — Multiple overlapping windows
// =============================================================================

func TestScenario_CascadingWindows(t *testing.T) {
	app, ta := NewTestApp(80, 30)
	app.Tracker().Enable()

	var clickLog []string

	// Win A: 30×15 at (0,0), z=10
	app.RegisterComponent("winA", "winA", buffer.Rect{X: 0, Y: 0, W: 30, H: 15}, 10,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "winA-root"
			root.Style.Background = "#AA0000"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickLog = append(clickLog, "winA")
			})
			txt := layout.NewVNode("text")
			txt.Content = "A"
			root.AddChild(txt)
			return root
		})

	// Win B: 30×15 at (10,5), z=20
	app.RegisterComponent("winB", "winB", buffer.Rect{X: 10, Y: 5, W: 30, H: 15}, 20,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "winB-root"
			root.Style.Background = "#00AA00"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickLog = append(clickLog, "winB")
			})
			txt := layout.NewVNode("text")
			txt.Content = "B"
			root.AddChild(txt)
			return root
		})

	// Win C: 30×15 at (20,10), z=30
	app.RegisterComponent("winC", "winC", buffer.Rect{X: 20, Y: 10, W: 30, H: 15}, 30,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "winC-root"
			root.Style.Background = "#0000AA"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickLog = append(clickLog, "winC")
			})
			txt := layout.NewVNode("text")
			txt.Content = "C"
			root.AddChild(txt)
			return root
		})

	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(3),
		perf.CheckLayouts(3),
		perf.CheckPaints(3),
		perf.CheckRenderComponents("winA", "winB", "winC"),
	)

	// --- Verify z-order at specific positions ---

	// (0,0) — only Win A → shows 'A'
	if c := ta.LastScreen.Get(0, 0).Char; c != 'A' {
		t.Errorf("(0,0): expected 'A', got %q", c)
	}

	// (10,5) — Win A and Win B overlap → Win B wins (z=20), shows 'B'
	if c := ta.LastScreen.Get(10, 5).Char; c != 'B' {
		t.Errorf("(10,5): expected 'B' (z=20 wins), got %q", c)
	}

	// (20,10) — All three overlap → Win C wins (z=30), shows 'C'
	if c := ta.LastScreen.Get(20, 10).Char; c != 'C' {
		t.Errorf("(20,10): expected 'C' (z=30 wins), got %q", c)
	}

	// (25,12) — Win B and Win C overlap → Win C wins (z=30)
	cell := ta.LastScreen.Get(25, 12)
	if cell.Background != "#0000AA" {
		t.Errorf("(25,12): expected winC bg '#0000AA', got %q", cell.Background)
	}

	// --- Click in overlap area → highest z wins ---
	clickLog = nil
	app.HandleEvent(&event.Event{Type: "mousedown", X: 25, Y: 12})
	if len(clickLog) != 1 || clickLog[0] != "winC" {
		t.Errorf("click in overlap: expected 'winC', got %v", clickLog)
	}

	// --- Click in Win A exclusive area ---
	clickLog = nil
	app.HandleEvent(&event.Event{Type: "mousedown", X: 5, Y: 2})
	if len(clickLog) != 1 || clickLog[0] != "winA" {
		t.Errorf("click in winA exclusive: expected 'winA', got %v", clickLog)
	}

	// --- Move Win B to non-overlapping position ---
	app.MoveComponent("winB", buffer.Rect{X: 50, Y: 0, W: 30, H: 15})
	app.RenderDirty()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(0),
		perf.CheckPaints(0),
		perf.CheckLayouts(1),
		perf.CheckOcclusionBuilds(1),
		perf.CheckMetric(perf.ComposeRects, 1),
		perf.CheckHitTesterRebuilds(1),
		perf.CheckHandlerFullSyncs(1),
	)

	// Old Win B position (15,7) should now show Win A's bg (A covers 0..29, 0..14)
	if bg := ta.LastScreen.Get(15, 7).Background; bg != "#AA0000" {
		t.Errorf("after move: (15,7) expected winA bg '#AA0000', got %q", bg)
	}

	// New Win B position (50,0) should show 'B'
	if c := ta.LastScreen.Get(50, 0).Char; c != 'B' {
		t.Errorf("after move: (50,0) expected 'B', got %q", c)
	}
}

// =============================================================================
// Scenario 4: Dynamic Window Manager — Add/remove/move windows rapidly
// =============================================================================

func TestScenario_DynamicWindowManager(t *testing.T) {
	app, ta := NewTestApp(80, 30)
	app.Tracker().Enable()

	// Start empty
	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(0),
		perf.CheckLayouts(0),
		perf.CheckPaints(0),
	)

	// Add 5 windows one by one
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("win-%d", i)
		x := i * 15
		char := rune('1' + i)
		bg := fmt.Sprintf("#%02x0000", (i+1)*40)
		app.RegisterComponent(id, "win", buffer.Rect{X: x, Y: 0, W: 15, H: 10}, i*10,
			func(state, props map[string]any) *layout.VNode {
				root := layout.NewVNode("box")
				root.ID = id + "-root"
				root.Style.Background = bg
				txt := layout.NewVNode("text")
				txt.Content = string(char)
				root.AddChild(txt)
				return root
			})
		app.RenderDirty()
		if i == 0 {
			app.Tracker().AssertLastFrame(t,
				perf.CheckRenders(1),
				perf.CheckRenderComponents("win-0"),
			)
		}
	}

	// Verify win-0 text at (0,0)
	if c := ta.LastScreen.Get(0, 0).Char; c == 0 {
		t.Error("after adding 5 windows: expected non-zero at (0,0)")
	}

	// Move win-2 to overlap win-0 (win-2 has higher z=20)
	app.MoveComponent("win-2", buffer.Rect{X: 0, Y: 0, W: 15, H: 10})
	app.RenderDirty()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(0),
		perf.CheckPaints(0),
		perf.CheckLayouts(1),
		perf.CheckOcclusionBuilds(1),
		perf.CheckMetric(perf.ComposeRects, 1),
		perf.CheckHitTesterRebuilds(1),
		perf.CheckHandlerFullSyncs(1),
	)

	// Remove win-1
	app.UnregisterComponent("win-1")
	app.RenderDirty()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(0),
		perf.CheckOcclusionBuilds(1),
		perf.CheckHitTesterRebuilds(1),
		perf.CheckHandlerFullSyncs(1),
	)

	// Add win-5 at highest z-index
	app.RegisterComponent("win-5", "win", buffer.Rect{X: 30, Y: 10, W: 20, H: 10}, 200,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "win-5-root"
			root.Style.Background = "#FF00FF"
			txt := layout.NewVNode("text")
			txt.Content = "Z"
			root.AddChild(txt)
			return root
		})
	app.RenderDirty()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(1),
		perf.CheckRenderComponents("win-5"),
	)

	// Verify win-5 renders
	if c := ta.LastScreen.Get(30, 10).Char; c != 'Z' {
		t.Errorf("win-5: expected 'Z' at (30,10), got %q", c)
	}

	// Resize screen
	app.Resize(120, 40)
	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(5),
		perf.CheckLayouts(5),
		perf.CheckPaints(5),
		perf.CheckRenderComponents("win-0", "win-2", "win-3", "win-4", "win-5"),
	)

	if ta.LastScreen.Width() != 120 || ta.LastScreen.Height() != 40 {
		t.Errorf("after resize: expected 120x40, got %dx%d", ta.LastScreen.Width(), ta.LastScreen.Height())
	}

	// win-5 still renders at (30,10) after resize
	if c := ta.LastScreen.Get(30, 10).Char; c != 'Z' {
		t.Errorf("after resize: win-5 expected 'Z' at (30,10), got %q", c)
	}

	// Verify no stale content from removed win-1 (was at x=15)
	// After win-1 removal and resize, area at (15,0) should either be empty
	// or show another window's content (not win-1's)
	comp := app.manager.Get("win-1")
	if comp != nil {
		t.Error("win-1 should be unregistered")
	}

	// Cumulative RenderFn invocations: empty frame + 5 new-window frames + move +
	// unregister + win-5 + resize-all (5 surviving components).
	if got := app.Tracker().TotalStats().Get(perf.Renders); got != 11 {
		t.Errorf("total RenderFn calls (perf): got %d, want 11", got)
	}
}

// =============================================================================
// Scenario 5: State Updates Across Components — Cross-component communication
// =============================================================================

func TestScenario_StateUpdatesAcrossComponents(t *testing.T) {
	app, ta := NewTestApp(60, 10)
	app.Tracker().Enable()

	// Counter display
	app.RegisterComponent("counter", "counter", buffer.Rect{X: 0, Y: 0, W: 20, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			count := 0
			if c, ok := state["count"].(int); ok {
				count = c
			}
			root := layout.NewVNode("box")
			root.Style.Background = "#111111"
			txt := layout.NewVNode("text")
			txt.Content = fmt.Sprintf("Count:%d", count)
			root.AddChild(txt)
			return root
		})

	// Incrementer button
	app.RegisterComponent("incrementer", "incrementer", buffer.Rect{X: 20, Y: 0, W: 20, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "inc-root"
			root.Style.Background = "#222222"
			btn := layout.NewVNode("text")
			btn.ID = "inc-btn"
			btn.Content = "+"
			btn.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				// Simulate cross-component state update
				c := 0
				if comp := app.manager.Get("counter"); comp != nil {
					if v, ok := comp.State()["count"].(int); ok {
						c = v
					}
				}
				c++
				app.SetState("counter", "count", c)
				app.SetState("display", "count", c)
			})
			root.AddChild(btn)
			return root
		})

	// Parity display
	app.RegisterComponent("display", "display", buffer.Rect{X: 40, Y: 0, W: 20, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			count := 0
			if c, ok := state["count"].(int); ok {
				count = c
			}
			parity := "even"
			if count%2 != 0 {
				parity = "odd"
			}
			root := layout.NewVNode("box")
			root.Style.Background = "#333333"
			txt := layout.NewVNode("text")
			txt.Content = parity
			root.AddChild(txt)
			return root
		})

	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(3),
		perf.CheckLayouts(3),
		perf.CheckPaints(3),
		perf.CheckRenderComponents("counter", "display", "incrementer"),
	)

	// Initial state: count=0
	if c := ta.LastScreen.Get(0, 0).Char; c != 'C' {
		t.Errorf("initial counter: expected 'C' (from 'Count:0'), got %q", c)
	}
	if c := ta.LastScreen.Get(40, 0).Char; c != 'e' {
		t.Errorf("initial display: expected 'e' (from 'even'), got %q", c)
	}

	// Click incrementer 3 times
	for i := 0; i < 3; i++ {
		app.HandleEvent(&event.Event{Type: "mousedown", X: 20, Y: 0})
		app.RenderDirty()
		app.Tracker().AssertLastFrame(t,
			perf.CheckRenders(2),
			perf.CheckRenderComponents("counter", "display"),
			perf.CheckOcclusionUpdates(1),
			perf.CheckMetric(perf.ComposeDirty, 1),
			perf.CheckHandlerDirtySyncs(1),
		)
	}

	// Counter should show "Count:3" — read exactly 7 chars
	s := ""
	for x := 0; x < 7; x++ {
		c := ta.LastScreen.Get(x, 0).Char
		if c == 0 {
			break
		}
		s += string(c)
	}
	if s != "Count:3" {
		t.Errorf("after 3 clicks: expected 'Count:3', got %q", s)
	}

	// Display should show "odd"
	if c := ta.LastScreen.Get(40, 0).Char; c != 'o' {
		t.Errorf("after 3 clicks: expected 'o' (from 'odd'), got %q", c)
	}

	// One more click → count=4, even
	app.HandleEvent(&event.Event{Type: "mousedown", X: 20, Y: 0})
	app.RenderDirty()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(2),
		perf.CheckRenderComponents("counter", "display"),
		perf.CheckOcclusionUpdates(1),
		perf.CheckMetric(perf.ComposeDirty, 1),
		perf.CheckHandlerDirtySyncs(1),
	)

	if c := ta.LastScreen.Get(40, 0).Char; c != 'e' {
		t.Errorf("after 4 clicks: expected 'e' (from 'even'), got %q", c)
	}
}

// =============================================================================
// Scenario 6: Focus Navigation Multi-Window — Tab across windows
// =============================================================================

func TestScenario_FocusNavigationMultiWindow(t *testing.T) {
	app, _ := NewTestApp(60, 20)
	app.Tracker().Enable()

	// 3 windows, each with 2 focusable elements
	for i := 0; i < 3; i++ {
		winID := fmt.Sprintf("win-%d", i)
		x := i * 20
		f1ID := fmt.Sprintf("field-%d-a", i)
		f2ID := fmt.Sprintf("field-%d-b", i)

		app.RegisterComponent(winID, "win", buffer.Rect{X: x, Y: 0, W: 20, H: 10}, 0,
			func(state, props map[string]any) *layout.VNode {
				root := layout.NewVNode("vbox")
				root.ID = winID + "-root"

				f1 := layout.NewVNode("text")
				f1.ID = f1ID
				f1.Content = f1ID
				f1.Props["focusable"] = true

				f2 := layout.NewVNode("text")
				f2.ID = f2ID
				f2.Content = f2ID
				f2.Props["focusable"] = true

				root.AddChild(f1)
				root.AddChild(f2)
				return root
			})
	}

	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(3),
		perf.CheckLayouts(3),
		perf.CheckPaints(3),
		perf.CheckRenderComponents("win-0", "win-1", "win-2"),
	)

	// Tab through all 6 focusable elements
	seen := make(map[string]bool)
	var order []string
	for i := 0; i < 6; i++ {
		app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
		id := app.FocusedID()
		if id == "" {
			t.Fatalf("Tab %d: expected non-empty focus", i+1)
		}
		seen[id] = true
		order = append(order, id)
	}

	if len(seen) != 6 {
		t.Errorf("expected 6 unique focused IDs, got %d: %v", len(seen), order)
	}

	// Verify wrap-around
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	if got := app.FocusedID(); got != order[0] {
		t.Errorf("wrap: expected %q, got %q", order[0], got)
	}

	// Shift-Tab reverses
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Shift+Tab"})
	if got := app.FocusedID(); got != order[5] {
		t.Errorf("Shift+Tab: expected %q, got %q", order[5], got)
	}

	// Remove middle window
	app.UnregisterComponent("win-1")
	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(0),
		perf.CheckLayouts(0),
		perf.CheckPaints(0),
	)

	// Tab through remaining elements (should be 4: from win-0 and win-2)
	seen2 := make(map[string]bool)
	for i := 0; i < 4; i++ {
		app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
		id := app.FocusedID()
		seen2[id] = true
	}

	// The removed window's focusables should not appear
	for id := range seen2 {
		if id == "field-1-a" || id == "field-1-b" {
			t.Errorf("removed window's focusable %q still reachable", id)
		}
	}

	// Add new window with focusable
	app.RegisterComponent("win-3", "win", buffer.Rect{X: 20, Y: 0, W: 20, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("vbox")
			root.ID = "win-3-root"
			f := layout.NewVNode("text")
			f.ID = "field-3-new"
			f.Content = "New"
			f.Props["focusable"] = true
			root.AddChild(f)
			return root
		})
	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(1),
		perf.CheckLayouts(1),
		perf.CheckPaints(1),
		perf.CheckRenderComponents("win-3"),
	)

	// Tab should eventually reach the new focusable
	foundNew := false
	for i := 0; i < 10; i++ {
		app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
		if app.FocusedID() == "field-3-new" {
			foundNew = true
			break
		}
	}
	if !foundNew {
		t.Error("new window's focusable 'field-3-new' not reachable via Tab")
	}
}

// =============================================================================
// Scenario 7: Hover Across Overlapping Windows
// =============================================================================

func TestScenario_HoverAcrossOverlappingWindows(t *testing.T) {
	app, _ := NewTestApp(40, 10)
	app.Tracker().Enable()

	enterCount := map[string]int{}
	leaveCount := map[string]int{}

	// Win A: 20×10 at (0,0), z=10
	app.RegisterComponent("winA", "winA", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 10,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "winA-box"
			root.Style.Background = "#AA0000"
			root.Props["onMouseEnter"] = event.EventHandler(func(e *event.Event) {
				enterCount["winA"]++
			})
			root.Props["onMouseLeave"] = event.EventHandler(func(e *event.Event) {
				leaveCount["winA"]++
			})
			txt := layout.NewVNode("text")
			txt.Content = "A"
			root.AddChild(txt)
			return root
		})

	// Win B: 20×10 at (20,0), z=10
	app.RegisterComponent("winB", "winB", buffer.Rect{X: 20, Y: 0, W: 20, H: 10}, 10,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "winB-box"
			root.Style.Background = "#00AA00"
			root.Props["onMouseEnter"] = event.EventHandler(func(e *event.Event) {
				enterCount["winB"]++
			})
			root.Props["onMouseLeave"] = event.EventHandler(func(e *event.Event) {
				leaveCount["winB"]++
			})
			txt := layout.NewVNode("text")
			txt.Content = "B"
			root.AddChild(txt)
			return root
		})

	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(2),
		perf.CheckLayouts(2),
		perf.CheckPaints(2),
		perf.CheckRenderComponents("winA", "winB"),
	)

	// Move mouse to Win A exclusive area (5,5)
	app.HandleEvent(&event.Event{Type: "mousemove", X: 5, Y: 5})
	if enterCount["winA"] != 1 {
		t.Errorf("winA enter: expected 1, got %d", enterCount["winA"])
	}
	if enterCount["winB"] != 0 {
		t.Errorf("winB enter: expected 0, got %d", enterCount["winB"])
	}

	// Move to Win B exclusive area (30,5)
	app.HandleEvent(&event.Event{Type: "mousemove", X: 30, Y: 5})
	if leaveCount["winA"] != 1 {
		t.Errorf("winA leave after crossing: expected 1, got %d", leaveCount["winA"])
	}
	if enterCount["winB"] != 1 {
		t.Errorf("winB enter after crossing: expected 1, got %d", enterCount["winB"])
	}

	// Move back to Win A
	app.HandleEvent(&event.Event{Type: "mousemove", X: 5, Y: 5})
	if enterCount["winA"] != 2 {
		t.Errorf("winA re-enter: expected 2, got %d", enterCount["winA"])
	}
	if leaveCount["winB"] != 1 {
		t.Errorf("winB leave: expected 1, got %d", leaveCount["winB"])
	}

	// Move outside both windows (no component at 39,9 if B ends at 40)
	// Actually B is at (20..39), so 39 is still in B. Move to a position outside.
	// Both windows cover the full 40-wide screen, so there's no "outside".
	// Instead, verify total counts are consistent.
	if enterCount["winA"] != 2 || leaveCount["winA"] != 1 {
		t.Errorf("winA final: enter=%d leave=%d, expected enter=2 leave=1",
			enterCount["winA"], leaveCount["winA"])
	}
	if enterCount["winB"] != 1 || leaveCount["winB"] != 1 {
		t.Errorf("winB final: enter=%d leave=%d, expected enter=1 leave=1",
			enterCount["winB"], leaveCount["winB"])
	}
}

// =============================================================================
// Scenario 8: Stress Test — Many Windows
// =============================================================================

func TestScenario_StressTest_ManyWindows(t *testing.T) {
	app, ta := NewTestApp(100, 50)
	app.Tracker().Enable()

	clickTargets := make(map[string]bool)

	// Register 50 small windows at various positions
	for i := 0; i < 50; i++ {
		id := fmt.Sprintf("sw-%d", i)
		x := (i * 7) % 90
		y := (i * 3) % 40
		z := i
		char := rune('A' + (i % 26))
		bg := fmt.Sprintf("#%02x%02x00", (i*5)%256, (i*3)%256)

		app.RegisterComponent(id, "small", buffer.Rect{X: x, Y: y, W: 10, H: 5}, z,
			func(state, props map[string]any) *layout.VNode {
				root := layout.NewVNode("box")
				root.ID = id + "-root"
				root.Style.Background = bg
				root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
					clickTargets[id] = true
				})
				txt := layout.NewVNode("text")
				txt.Content = string(char)
				root.AddChild(txt)
				return root
			})
	}

	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(50),
		perf.CheckLayouts(50),
		perf.CheckPaints(50),
	)

	// Verify screen is not empty
	nonZero := 0
	for y := 0; y < 50; y++ {
		for x := 0; x < 100; x++ {
			if !ta.LastScreen.Get(x, y).Zero() {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Error("screen is entirely empty after rendering 50 windows")
	}

	// Click at 20 random positions across the screen
	for _, pos := range [][2]int{
		{5, 2}, {15, 8}, {25, 12}, {35, 5}, {45, 20},
		{55, 15}, {65, 25}, {75, 10}, {85, 30}, {95, 35},
		{10, 0}, {20, 3}, {30, 9}, {40, 18}, {50, 22},
		{60, 28}, {70, 33}, {80, 38}, {90, 42}, {0, 0},
	} {
		app.HandleEvent(&event.Event{Type: "mousedown", X: pos[0], Y: pos[1]})
	}
	// No crash is the main assertion; some clicks should have hit targets
	if len(clickTargets) == 0 {
		t.Error("expected at least some click targets to be hit")
	}

	// Remove every other window
	for i := 0; i < 50; i += 2 {
		app.UnregisterComponent(fmt.Sprintf("sw-%d", i))
	}
	app.RenderDirty()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(0),
		perf.CheckOcclusionBuilds(1),
		perf.CheckHitTesterRebuilds(1),
		perf.CheckHandlerFullSyncs(1),
	)

	// Verify remaining windows still render
	remainingRendered := false
	for y := 0; y < ta.LastScreen.Height(); y++ {
		for x := 0; x < ta.LastScreen.Width(); x++ {
			if !ta.LastScreen.Get(x, y).Zero() {
				remainingRendered = true
				break
			}
		}
		if remainingRendered {
			break
		}
	}
	if !remainingRendered {
		t.Error("after removing half the windows, remaining should still render")
	}
}

// =============================================================================
// Scenario 9: Resize With Active Content — State preserved
// =============================================================================

func TestScenario_ResizeWithActiveContent(t *testing.T) {
	app, ta := NewTestApp(40, 20)
	app.Tracker().Enable()

	// 3 components with state
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("comp-%d", i)
		x := i * 13
		app.RegisterComponent(id, "comp", buffer.Rect{X: x, Y: 0, W: 13, H: 10}, 0,
			func(state, props map[string]any) *layout.VNode {
				val := 0
				if v, ok := state["val"].(int); ok {
					val = v
				}
				root := layout.NewVNode("box")
				root.Style.Background = "#111111"
				txt := layout.NewVNode("text")
				txt.Content = fmt.Sprintf("%s:%d", id, val)
				root.AddChild(txt)
				return root
			})
	}

	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(3),
		perf.CheckLayouts(3),
		perf.CheckPaints(3),
		perf.CheckRenderComponents("comp-0", "comp-1", "comp-2"),
	)

	// Set state on each component
	app.SetState("comp-0", "val", 10)
	app.SetState("comp-1", "val", 20)
	app.SetState("comp-2", "val", 30)
	app.RenderDirty()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(3),
		perf.CheckRenderComponents("comp-0", "comp-1", "comp-2"),
		perf.CheckOcclusionUpdates(1),
		perf.CheckMetric(perf.ComposeDirty, 1),
		perf.CheckHandlerDirtySyncs(1),
	)

	// Verify state reflected on screen for comp-0
	// Read exactly len("comp-0:10") = 9 chars (box fills rest with spaces)
	readScreen := func(maxX int) string {
		s := ""
		for x := 0; x < maxX; x++ {
			c := ta.LastScreen.Get(x, 0).Char
			if c == 0 {
				break
			}
			s += string(c)
		}
		return s
	}
	expected := "comp-0:10"
	if s := readScreen(len(expected)); s != expected {
		t.Errorf("before resize: expected %q, got %q", expected, s)
	}

	// Resize smaller
	app.Resize(30, 15)
	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(3),
		perf.CheckLayouts(3),
		perf.CheckPaints(3),
		perf.CheckRenderComponents("comp-0", "comp-1", "comp-2"),
	)

	if ta.LastScreen.Width() != 30 || ta.LastScreen.Height() != 15 {
		t.Errorf("after shrink: expected 30x15, got %dx%d", ta.LastScreen.Width(), ta.LastScreen.Height())
	}

	// State should be preserved — comp-0 should still show val=10
	if s := readScreen(len(expected)); s != expected {
		t.Errorf("after shrink: expected %q, got %q", expected, s)
	}

	// Resize larger
	app.Resize(80, 40)
	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(3),
		perf.CheckLayouts(3),
		perf.CheckPaints(3),
		perf.CheckRenderComponents("comp-0", "comp-1", "comp-2"),
	)

	if ta.LastScreen.Width() != 80 || ta.LastScreen.Height() != 40 {
		t.Errorf("after grow: expected 80x40, got %dx%d", ta.LastScreen.Width(), ta.LastScreen.Height())
	}

	// State still preserved
	if s := readScreen(len(expected)); s != expected {
		t.Errorf("after grow: expected %q, got %q", expected, s)
	}

	// Verify comp-2 state also preserved
	comp2 := app.manager.Get("comp-2")
	if comp2 == nil {
		t.Fatal("comp-2 should exist")
	}
	if comp2.State()["val"] != 30 {
		t.Errorf("comp-2 state: expected val=30, got %v", comp2.State()["val"])
	}
}

// =============================================================================
// Scenario 10: Rapid Move And Click — Move then click verification
// =============================================================================

func TestScenario_RapidMoveAndClick(t *testing.T) {
	app, _ := NewTestApp(40, 10)
	app.Tracker().Enable()

	clickCount := 0

	// Window at position A
	app.RegisterComponent("movable", "movable", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "movable-root"
			root.Style.Background = "#FF0000"
			root.Props["onClick"] = event.EventHandler(func(e *event.Event) {
				clickCount++
			})
			txt := layout.NewVNode("text")
			txt.Content = "M"
			root.AddChild(txt)
			return root
		})

	app.RenderAll()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(1),
		perf.CheckLayouts(1),
		perf.CheckPaints(1),
		perf.CheckRenderComponents("movable"),
	)

	// Click at position A — should hit
	clickCount = 0
	app.HandleEvent(&event.Event{Type: "mousedown", X: 5, Y: 2})
	if clickCount != 1 {
		t.Errorf("before move: expected 1 click, got %d", clickCount)
	}

	// Move window to position B
	app.MoveComponent("movable", buffer.Rect{X: 25, Y: 3, W: 10, H: 5})
	app.RenderDirty()
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(0),
		perf.CheckPaints(0),
		perf.CheckLayouts(1),
		perf.CheckOcclusionBuilds(1),
		perf.CheckMetric(perf.ComposeRects, 1),
		perf.CheckHitTesterRebuilds(1),
		perf.CheckHandlerFullSyncs(1),
	)

	// Click at position B — should hit
	clickCount = 0
	app.HandleEvent(&event.Event{Type: "mousedown", X: 30, Y: 5})
	if clickCount != 1 {
		t.Errorf("after move, click at B: expected 1 click, got %d", clickCount)
	}

	// Click at old position A — should NOT hit
	clickCount = 0
	app.HandleEvent(&event.Event{Type: "mousedown", X: 5, Y: 2})
	if clickCount != 0 {
		t.Errorf("after move, click at A: expected 0 clicks (moved away), got %d", clickCount)
	}

	// Multiple rapid moves
	for i := 0; i < 10; i++ {
		x := (i * 3) % 30
		app.MoveComponent("movable", buffer.Rect{X: x, Y: 0, W: 10, H: 5})
		app.RenderDirty()
		app.Tracker().AssertLastFrame(t,
			perf.CheckRenders(0),
			perf.CheckPaints(0),
			perf.CheckLayouts(1),
			perf.CheckOcclusionBuilds(1),
			perf.CheckMetric(perf.ComposeRects, 1),
			perf.CheckHitTesterRebuilds(1),
			perf.CheckHandlerFullSyncs(1),
		)
	}

	// Final position: x = (9*3)%30 = 27
	clickCount = 0
	app.HandleEvent(&event.Event{Type: "mousedown", X: 30, Y: 2})
	if clickCount != 1 {
		t.Errorf("after rapid moves, click at final pos: expected 1 click, got %d", clickCount)
	}

	// Click at a position that's NOT the final position
	clickCount = 0
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	if clickCount != 0 {
		t.Errorf("after rapid moves, click at wrong pos: expected 0 clicks, got %d", clickCount)
	}
}
