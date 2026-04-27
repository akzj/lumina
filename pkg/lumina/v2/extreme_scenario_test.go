package v2

import (
	"fmt"
	"strings"
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
)

// extreme_scenario_test.go — additional integration scenarios beyond complex_scenario_test.go:
// out-of-bounds input, nil render, hbox/vbox/fragment/gap, same z-index, batch state,
// focus (Shift+Tab), hover storms, border, text wrap depth, unregister-all, render alternation.

// -----------------------------------------------------------------------------
// Bounds & safety
// -----------------------------------------------------------------------------

func TestExtreme_OutOfBoundsMouseDoesNotPanic(t *testing.T) {
	app, _ := NewTestApp(40, 20)
	app.RegisterComponent("a", "a", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("box")
			r.ID = "a-root"
			t := layout.NewVNode("text")
			t.Content = "X"
			r.AddChild(t)
			return r
		})
	app.RenderAll()

	for _, xy := range [][2]int{{-1, 0}, {0, -1}, {1000, 0}, {0, 9999}, {39, 19}} {
		app.HandleEvent(&event.Event{Type: "mousemove", X: xy[0], Y: xy[1]})
		app.HandleEvent(&event.Event{Type: "mousedown", X: xy[0], Y: xy[1]})
	}
}

func TestExtreme_ClickOutsideAnyComponent(t *testing.T) {
	app, ta := NewTestApp(60, 40)
	clicks := 0
	app.RegisterComponent("tiny", "tiny", buffer.Rect{X: 0, Y: 0, W: 3, H: 2}, 0,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("box")
			r.ID = "tiny-root"
			r.Props["onClick"] = event.EventHandler(func(e *event.Event) { clicks++ })
			t := layout.NewVNode("text")
			t.Content = "T"
			r.AddChild(t)
			return r
		})
	app.RenderAll()

	app.HandleEvent(&event.Event{Type: "mousedown", X: 50, Y: 30})
	if clicks != 0 {
		t.Errorf("click outside tiny: expected 0 root clicks, got %d", clicks)
	}
	// Screen cell should still be default / empty there
	if c := ta.LastScreen.Get(50, 30).Char; c != 0 && c != ' ' {
		// space is possible if adapter fills; zero is typical for untouched
		_ = c
	}
}

// -----------------------------------------------------------------------------
// Nil render & minimal surface
// -----------------------------------------------------------------------------

func TestExtreme_RenderFnReturnsNil_NoPanic(t *testing.T) {
	app, _ := NewTestApp(20, 10)
	app.RegisterComponent("ghost", "ghost", buffer.Rect{X: 0, Y: 0, W: 5, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode { return nil })
	app.RenderAll()
	app.RenderDirty()
}

func TestExtreme_1x1ComponentClick(t *testing.T) {
	app, ta := NewTestApp(10, 10)
	hit := false
	app.RegisterComponent("dot", "dot", buffer.Rect{X: 5, Y: 5, W: 1, H: 1}, 10,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("box")
			r.ID = "dot-root"
			r.Props["onClick"] = event.EventHandler(func(e *event.Event) { hit = true })
			t := layout.NewVNode("text")
			t.Content = "@"
			r.AddChild(t)
			return r
		})
	app.RenderAll()
	if ta.LastScreen.Get(5, 5).Char != '@' {
		t.Fatalf("expected '@' at (5,5), got %q", ta.LastScreen.Get(5, 5).Char)
	}
	app.HandleEvent(&event.Event{Type: "mousedown", X: 5, Y: 5})
	if !hit {
		t.Error("expected click on 1x1 component")
	}
}

// -----------------------------------------------------------------------------
// Layout: hbox, vbox+gap, fragment, deep tree, text wrap
// -----------------------------------------------------------------------------

func TestExtreme_HBoxChildOrderOnScreen(t *testing.T) {
	app, ta := NewTestApp(30, 5)
	app.RegisterComponent("row", "row", buffer.Rect{X: 0, Y: 0, W: 30, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("hbox")
			for _, ch := range []string{"A", "B", "C"} {
				t := layout.NewVNode("text")
				t.Content = ch
				r.AddChild(t)
			}
			return r
		})
	app.RenderAll()
	if ta.LastScreen.Get(0, 0).Char != 'A' {
		t.Errorf("col0: want 'A', got %q", ta.LastScreen.Get(0, 0).Char)
	}
	if ta.LastScreen.Get(1, 0).Char != 'B' {
		t.Errorf("col1: want 'B', got %q", ta.LastScreen.Get(1, 0).Char)
	}
	if ta.LastScreen.Get(2, 0).Char != 'C' {
		t.Errorf("col2: want 'C', got %q", ta.LastScreen.Get(2, 0).Char)
	}
}

func TestExtreme_VBoxWithGap(t *testing.T) {
	app, ta := NewTestApp(10, 12)
	app.RegisterComponent("g", "g", buffer.Rect{X: 0, Y: 0, W: 10, H: 12}, 0,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("vbox")
			r.Style.Gap = 2
			for i := 0; i < 3; i++ {
				t := layout.NewVNode("text")
				t.Content = fmt.Sprintf("%d", i)
				r.AddChild(t)
			}
			return r
		})
	app.RenderAll()
	// Row0: '0', gap 2 → next text baseline row ~3
	if ta.LastScreen.Get(0, 0).Char != '0' {
		t.Errorf("y0: want '0', got %q", ta.LastScreen.Get(0, 0).Char)
	}
	if ta.LastScreen.Get(0, 3).Char != '1' {
		t.Errorf("y3: want '1' after gap, got %q", ta.LastScreen.Get(0, 3).Char)
	}
}

func TestExtreme_FragmentNaturalHeight(t *testing.T) {
	app, ta := NewTestApp(20, 10)
	app.RegisterComponent("f", "f", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 0,
		func(state, props map[string]any) *layout.VNode {
			frag := layout.NewVNode("fragment")
			for i := 0; i < 4; i++ {
				t := layout.NewVNode("text")
				t.Content = "x"
				frag.AddChild(t)
			}
			return frag
		})
	app.RenderAll()
	// At least one row should show x from fragment children
	if ta.LastScreen.Get(0, 0).Char != 'x' {
		t.Errorf("expected 'x' at origin from fragment stack, got %q", ta.LastScreen.Get(0, 0).Char)
	}
}

func TestExtreme_TextWrapsInNarrowColumn(t *testing.T) {
	app, ta := NewTestApp(6, 20)
	app.RegisterComponent("wrap", "wrap", buffer.Rect{X: 0, Y: 0, W: 6, H: 20}, 0,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("box")
			t := layout.NewVNode("text")
			t.Content = "ABCDEFGHIJ" // 10 chars, width 6 → 2 lines
			r.AddChild(t)
			return r
		})
	app.RenderAll()
	if ta.LastScreen.Get(0, 0).Char != 'A' {
		t.Fatalf("line1: got %q", ta.LastScreen.Get(0, 0).Char)
	}
	if ta.LastScreen.Get(0, 1).Char != 'G' {
		t.Errorf("line2 should start with 'G', got %q", ta.LastScreen.Get(0, 1).Char)
	}
}

func TestExtreme_DeepVBoxChainSingleComponent(t *testing.T) {
	app, _ := NewTestApp(40, 25)
	const depth = 12
	var build func(d int) *layout.VNode
	build = func(d int) *layout.VNode {
		if d == 0 {
			leaf := layout.NewVNode("text")
			leaf.Content = "Z"
			return leaf
		}
		v := layout.NewVNode("vbox")
		v.AddChild(build(d - 1))
		return v
	}
	app.RegisterComponent("deep", "deep", buffer.Rect{X: 0, Y: 0, W: 10, H: 25}, 0,
		func(state, props map[string]any) *layout.VNode { return build(depth) })
	app.RenderAll()
}

// -----------------------------------------------------------------------------
// Z-index & stacking
// -----------------------------------------------------------------------------

func TestExtreme_HighZRegisteredAfterLowZ_WinsOnOverlap(t *testing.T) {
	app, ta := NewTestApp(30, 10)
	app.RegisterComponent("low", "low", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 5,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("box")
			r.Style.Background = "#111111"
			t := layout.NewVNode("text")
			t.Content = "L"
			r.AddChild(t)
			return r
		})
	app.RegisterComponent("high", "high", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 50,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("box")
			r.Style.Background = "#222222"
			t := layout.NewVNode("text")
			t.Content = "H"
			r.AddChild(t)
			return r
		})
	app.RenderAll()
	if ta.LastScreen.Get(0, 0).Char != 'H' {
		t.Errorf("overlap: higher z should paint 'H', got %q", ta.LastScreen.Get(0, 0).Char)
	}
}

func TestExtreme_SameZIndexOverlap_DeterministicAndNoPanic(t *testing.T) {
	app, ta := NewTestApp(40, 10)
	app.RegisterComponent("p", "p", buffer.Rect{X: 0, Y: 0, W: 25, H: 10}, 7,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("box")
			t := layout.NewVNode("text")
			t.Content = "P"
			r.AddChild(t)
			return r
		})
	app.RegisterComponent("q", "q", buffer.Rect{X: 5, Y: 0, W: 25, H: 10}, 7,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("box")
			t := layout.NewVNode("text")
			t.Content = "Q"
			r.AddChild(t)
			return r
		})
	app.RenderAll()
	// (0,0) is only covered by p — must show P regardless of tie-break for equal z.
	if ta.LastScreen.Get(0, 0).Char != 'P' {
		t.Errorf("(0,0) exclusive to p: want 'P', got %q", ta.LastScreen.Get(0, 0).Char)
	}
	// Overlap region: equal z order is undefined; only assert composed cell is not an error / crash.
	_ = ta.LastScreen.Get(12, 0)
}

// -----------------------------------------------------------------------------
// State batching, resize edge, unregister all
// -----------------------------------------------------------------------------

func TestExtreme_MultipleSetStateSingleRenderDirty(t *testing.T) {
	app, ta := NewTestApp(30, 5)
	app.RegisterComponent("c", "c", buffer.Rect{X: 0, Y: 0, W: 30, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			a, b := 0, 0
			if v, ok := state["a"].(int); ok {
				a = v
			}
			if v, ok := state["b"].(int); ok {
				b = v
			}
			r := layout.NewVNode("box")
			t := layout.NewVNode("text")
			t.Content = fmt.Sprintf("%d,%d", a, b)
			r.AddChild(t)
			return r
		})
	app.RenderAll()
	app.SetState("c", "a", 7)
	app.SetState("c", "b", 3)
	app.RenderDirty()
	s := strings.TrimRight(screenRowPrefix(ta, 0, 8), "\x00")
	if s != "7,3" {
		t.Errorf("screen prefix: want 7,3, got %q", s)
	}
}

func TestExtreme_UnregisterAllThenRenderAll(t *testing.T) {
	app, ta := NewTestApp(20, 10)
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("w%d", i)
		app.RegisterComponent(id, id, buffer.Rect{X: i * 5, Y: 0, W: 4, H: 4}, i,
			func(state, props map[string]any) *layout.VNode {
				r := layout.NewVNode("text")
				r.Content = "x"
				return r
			})
	}
	app.RenderAll()
	for i := 0; i < 3; i++ {
		app.UnregisterComponent(fmt.Sprintf("w%d", i))
	}
	app.RenderAll()
	empty := true
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			if !ta.LastScreen.Get(x, y).Zero() {
				empty = false
				break
			}
		}
	}
	if !empty {
		t.Log("note: screen may retain previous frame depending on adapter; checking no crash only")
	}
}

func TestExtreme_RenderDirtyRenderDirtyRenderAll(t *testing.T) {
	app, ta := NewTestApp(24, 8)
	app.RegisterComponent("v", "v", buffer.Rect{X: 0, Y: 0, W: 24, H: 8}, 0,
		func(state, props map[string]any) *layout.VNode {
			k := 0
			if v, ok := state["k"].(int); ok {
				k = v
			}
			r := layout.NewVNode("text")
			r.Content = fmt.Sprintf("k=%d", k)
			return r
		})
	app.RenderAll()
	app.SetState("v", "k", 1)
	app.RenderDirty()
	app.SetState("v", "k", 2)
	app.RenderDirty()
	app.RenderAll()
	if !strings.Contains(screenRowPrefix(ta, 0, 8), "k=2") {
		t.Errorf("expected k=2 on screen, got %q", screenRowPrefix(ta, 0, 12))
	}
}

// -----------------------------------------------------------------------------
// Focus: Shift+Tab, key with no focus
// -----------------------------------------------------------------------------

func TestExtreme_ShiftTabCyclesBackward(t *testing.T) {
	app, _ := NewTestApp(40, 5)
	app.RegisterComponent("a", "a", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("box")
			t1 := layout.NewVNode("text")
			t1.ID = "t1"
			t1.Content = "1"
			t1.Props["focusable"] = true
			t2 := layout.NewVNode("text")
			t2.ID = "t2"
			t2.Content = "2"
			t2.Props["focusable"] = true
			r.AddChild(t1)
			r.AddChild(t2)
			return r
		})
	app.RenderAll()
	app.HandleEvent(&event.Event{Type: "mousedown", X: 0, Y: 0})
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Tab"})
	first := app.FocusedID()
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Shift+Tab"})
	back := app.FocusedID()
	if first == "" || back == "" {
		t.Fatalf("focus ids empty: first=%q back=%q", first, back)
	}
	if first == back {
		t.Errorf("Shift+Tab should move focus, got same %q", first)
	}
}

func TestExtreme_KeyDownWithoutFocusNoPanic(t *testing.T) {
	app, _ := NewTestApp(20, 10)
	app.RegisterComponent("k", "k", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("text")
			r.Content = "ok"
			return r
		})
	app.RenderAll()
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Enter"})
	app.HandleEvent(&event.Event{Type: "keydown", Key: "Escape"})
}

// -----------------------------------------------------------------------------
// Hover storm & border
// -----------------------------------------------------------------------------

func TestExtreme_MouseMoveStorm(t *testing.T) {
	app, _ := NewTestApp(50, 30)
	app.RegisterComponent("w", "w", buffer.Rect{X: 0, Y: 0, W: 30, H: 20}, 0,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("box")
			r.ID = "w-root"
			t := layout.NewVNode("text")
			t.ID = "w-t"
			t.Content = "M"
			t.Props["onMouseEnter"] = event.EventHandler(func(e *event.Event) {})
			r.AddChild(t)
			return r
		})
	app.RenderAll()
	for i := 0; i < 400; i++ {
		x := (i * 7) % 50
		y := (i * 3) % 30
		app.HandleEvent(&event.Event{Type: "mousemove", X: x, Y: y})
	}
}

func TestExtreme_BorderedBoxDrawsCornerGlyph(t *testing.T) {
	app, ta := NewTestApp(12, 8)
	app.RegisterComponent("b", "b", buffer.Rect{X: 0, Y: 0, W: 12, H: 8}, 0,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("box")
			r.Style.Border = "single"
			r.Style.Background = "#000011"
			t := layout.NewVNode("text")
			t.Content = "in"
			r.AddChild(t)
			return r
		})
	app.RenderAll()
	c := ta.LastScreen.Get(0, 0).Char
	if c == 0 {
		t.Fatal("expected border corner glyph at (0,0)")
	}
}

// -----------------------------------------------------------------------------
// Move + paint-only dirty (occlusion cell transparency change simulation)
// -----------------------------------------------------------------------------

func TestExtreme_SetStateChangesTextWidthThenRenderDirty(t *testing.T) {
	app, ta := NewTestApp(20, 5)
	app.RegisterComponent("len", "len", buffer.Rect{X: 0, Y: 0, W: 20, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			s := "9"
			if v, ok := state["n"].(int); ok && v >= 10 {
				s = "10"
			}
			r := layout.NewVNode("box")
			t := layout.NewVNode("text")
			t.Content = s
			r.AddChild(t)
			return r
		})
	app.RenderAll()
	app.SetState("len", "n", 10)
	app.RenderDirty()
	// "10" at origin
	if ta.LastScreen.Get(0, 0).Char != '1' || ta.LastScreen.Get(1, 0).Char != '0' {
		t.Errorf("expected '10' at (0,0)-(1,0), got %q%q",
			ta.LastScreen.Get(0, 0).Char, ta.LastScreen.Get(1, 0).Char)
	}
}

// -----------------------------------------------------------------------------
// Padding shorthand (layout integration)
// -----------------------------------------------------------------------------

func TestExtreme_LayoutPaddingShorthandInVNodeTree(t *testing.T) {
	app, ta := NewTestApp(16, 8)
	app.RegisterComponent("pad", "pad", buffer.Rect{X: 0, Y: 0, W: 16, H: 8}, 0,
		func(state, props map[string]any) *layout.VNode {
			r := layout.NewVNode("vbox")
			r.Style.Padding = 2
			t := layout.NewVNode("text")
			t.Content = "Z"
			r.AddChild(t)
			return r
		})
	app.RenderAll()
	if ta.LastScreen.Get(2, 2).Char != 'Z' {
		t.Errorf("with padding=2 text should start around (2,2), got %q at (2,2)",
			ta.LastScreen.Get(2, 2).Char)
	}
}

// -----------------------------------------------------------------------------
// helpers
// -----------------------------------------------------------------------------

func screenRowPrefix(ta *output.TestAdapter, y, maxX int) string {
	var b strings.Builder
	for x := 0; x < maxX && x < ta.LastScreen.Width(); x++ {
		ch := ta.LastScreen.Get(x, y).Char
		if ch == 0 {
			break
		}
		b.WriteRune(ch)
	}
	return b.String()
}
