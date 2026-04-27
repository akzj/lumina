package v2

import (
	"fmt"
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// --- Test 1: Position-only move does NOT re-render ---

func TestRenderCount_MoveWithoutResize_NoRerender(t *testing.T) {
	app, ta := NewTestApp(60, 20)
	renderCounts := map[string]int{}

	// Register 3 components at non-overlapping positions.
	app.RegisterComponent("c1", "c1", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			renderCounts["c1"]++
			root := layout.NewVNode("box")
			root.Style.Background = "#111111"
			txt := layout.NewVNode("text")
			txt.Content = "A"
			root.AddChild(txt)
			return root
		})
	app.RegisterComponent("c2", "c2", buffer.Rect{X: 15, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			renderCounts["c2"]++
			root := layout.NewVNode("box")
			root.Style.Background = "#222222"
			txt := layout.NewVNode("text")
			txt.Content = "B"
			root.AddChild(txt)
			return root
		})
	app.RegisterComponent("c3", "c3", buffer.Rect{X: 30, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			renderCounts["c3"]++
			root := layout.NewVNode("box")
			root.Style.Background = "#333333"
			txt := layout.NewVNode("text")
			txt.Content = "C"
			root.AddChild(txt)
			return root
		})

	app.RenderAll()
	for _, id := range []string{"c1", "c2", "c3"} {
		if renderCounts[id] != 1 {
			t.Fatalf("after RenderAll: %s renderCount=%d, want 1", id, renderCounts[id])
		}
	}

	// Move c2 to a new position (same size).
	app.MoveComponent("c2", buffer.Rect{X: 45, Y: 10, W: 10, H: 5})
	app.RenderDirty()

	// c2 must NOT have been re-rendered.
	if renderCounts["c2"] != 1 {
		t.Errorf("after position-only move: c2 renderCount=%d, want 1", renderCounts["c2"])
	}
	// c1 and c3 must be untouched.
	if renderCounts["c1"] != 1 {
		t.Errorf("c1 renderCount=%d, want 1", renderCounts["c1"])
	}
	if renderCounts["c3"] != 1 {
		t.Errorf("c3 renderCount=%d, want 1", renderCounts["c3"])
	}

	// c2 should be visible at the new position.
	if c := ta.LastScreen.Get(45, 10).Char; c != 'B' {
		t.Errorf("expected 'B' at new position (45,10), got %q", c)
	}
	// Old position should show zero (no component there anymore).
	if c := ta.LastScreen.Get(15, 0).Char; c == 'B' {
		t.Errorf("old position (15,0) still shows 'B' after move")
	}
}

// --- Test 2: Move with resize DOES re-render ---

func TestRenderCount_MoveWithResize_Rerenders(t *testing.T) {
	app, ta := NewTestApp(40, 20)
	renderCount := 0

	app.RegisterComponent("c1", "c1", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			renderCount++
			root := layout.NewVNode("box")
			root.Style.Background = "#FF0000"
			txt := layout.NewVNode("text")
			txt.Content = "X"
			root.AddChild(txt)
			return root
		})

	app.RenderAll()
	if renderCount != 1 {
		t.Fatalf("after RenderAll: renderCount=%d, want 1", renderCount)
	}

	// Move with different size → must re-render.
	app.MoveComponent("c1", buffer.Rect{X: 5, Y: 5, W: 20, H: 10})
	app.RenderDirty()

	if renderCount != 2 {
		t.Errorf("after resize move: renderCount=%d, want 2", renderCount)
	}
	// Content should be at new position.
	if c := ta.LastScreen.Get(5, 5).Char; c != 'X' {
		t.Errorf("expected 'X' at (5,5), got %q", c)
	}
}

// --- Test 3: SetState only re-renders the dirty component ---

func TestRenderCount_SetState_OnlyDirtyRerenders(t *testing.T) {
	app, _ := NewTestApp(60, 20)
	renderCounts := map[string]int{}

	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("c%d", i)
		x := (i - 1) * 12
		app.RegisterComponent(id, id, buffer.Rect{X: x, Y: 0, W: 10, H: 5}, 0,
			func(state, props map[string]any) *layout.VNode {
				renderCounts[id]++
				vn := layout.NewVNode("text")
				vn.Content = id
				return vn
			})
	}

	app.RenderAll()
	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("c%d", i)
		if renderCounts[id] != 1 {
			t.Fatalf("after RenderAll: %s renderCount=%d, want 1", id, renderCounts[id])
		}
	}

	// Only set state on c3.
	app.SetState("c3", "foo", "bar")
	app.RenderDirty()

	if renderCounts["c3"] != 2 {
		t.Errorf("c3 renderCount=%d, want 2", renderCounts["c3"])
	}
	for _, id := range []string{"c1", "c2", "c4", "c5"} {
		if renderCounts[id] != 1 {
			t.Errorf("%s renderCount=%d, want 1 (should not re-render)", id, renderCounts[id])
		}
	}
}

// --- Test 4: SetState on two components re-renders only those two ---

func TestRenderCount_SetState_TwoComponents(t *testing.T) {
	app, _ := NewTestApp(60, 20)
	renderCounts := map[string]int{}

	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("c%d", i)
		x := (i - 1) * 12
		app.RegisterComponent(id, id, buffer.Rect{X: x, Y: 0, W: 10, H: 5}, 0,
			func(state, props map[string]any) *layout.VNode {
				renderCounts[id]++
				vn := layout.NewVNode("text")
				vn.Content = id
				return vn
			})
	}

	app.RenderAll()

	app.SetState("c2", "x", 1)
	app.SetState("c4", "x", 2)
	app.RenderDirty()

	for _, id := range []string{"c2", "c4"} {
		if renderCounts[id] != 2 {
			t.Errorf("%s renderCount=%d, want 2", id, renderCounts[id])
		}
	}
	for _, id := range []string{"c1", "c3", "c5"} {
		if renderCounts[id] != 1 {
			t.Errorf("%s renderCount=%d, want 1", id, renderCounts[id])
		}
	}
}

// --- Test 5: Multiple position-only moves in one batch → no re-render ---

func TestRenderCount_MultipleMovesOneBatch(t *testing.T) {
	app, ta := NewTestApp(80, 20)
	renderCounts := map[string]int{}
	contents := []string{"A", "B", "C"}

	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("c%d", i)
		x := i * 15
		content := contents[i]
		app.RegisterComponent(id, id, buffer.Rect{X: x, Y: 0, W: 10, H: 5}, 0,
			func(state, props map[string]any) *layout.VNode {
				renderCounts[id]++
				root := layout.NewVNode("box")
				root.Style.Background = "#AAAAAA"
				txt := layout.NewVNode("text")
				txt.Content = content
				root.AddChild(txt)
				return root
			})
	}

	app.RenderAll()

	// Move all 3 to new positions (same size).
	app.MoveComponent("c0", buffer.Rect{X: 50, Y: 10, W: 10, H: 5})
	app.MoveComponent("c1", buffer.Rect{X: 60, Y: 10, W: 10, H: 5})
	app.MoveComponent("c2", buffer.Rect{X: 70, Y: 10, W: 10, H: 5})
	app.RenderDirty()

	// No re-renders.
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("c%d", i)
		if renderCounts[id] != 1 {
			t.Errorf("%s renderCount=%d, want 1", id, renderCounts[id])
		}
	}

	// All visible at new positions.
	for i, ch := range []rune{'A', 'B', 'C'} {
		x := 50 + i*10
		if c := ta.LastScreen.Get(x, 10).Char; c != ch {
			t.Errorf("expected %q at (%d,10), got %q", ch, x, c)
		}
	}
}

// --- Test 6: Move without RenderDirty doesn't update screen ---

func TestRenderCount_MoveDoesNotTriggerRenderWithoutRenderDirty(t *testing.T) {
	app, ta := NewTestApp(40, 10)
	renderCount := 0

	app.RegisterComponent("c1", "c1", buffer.Rect{X: 0, Y: 0, W: 10, H: 5}, 0,
		func(state, props map[string]any) *layout.VNode {
			renderCount++
			root := layout.NewVNode("box")
			root.Style.Background = "#FF0000"
			txt := layout.NewVNode("text")
			txt.Content = "M"
			root.AddChild(txt)
			return root
		})

	app.RenderAll()
	writesBefore := ta.WriteCount

	// Move but don't call RenderDirty.
	app.MoveComponent("c1", buffer.Rect{X: 20, Y: 0, W: 10, H: 5})

	// Screen should not have been updated.
	if ta.WriteCount != writesBefore {
		t.Errorf("screen updated without RenderDirty: WriteCount=%d, want %d", ta.WriteCount, writesBefore)
	}

	// Now call RenderDirty — screen should update.
	app.RenderDirty()
	if ta.WriteCount == writesBefore {
		t.Errorf("screen NOT updated after RenderDirty: WriteCount still %d", ta.WriteCount)
	}

	// Content at new position.
	if c := ta.LastScreen.Get(20, 0).Char; c != 'M' {
		t.Errorf("expected 'M' at (20,0), got %q", c)
	}
	// renderFn should NOT have been called again (position-only move).
	if renderCount != 1 {
		t.Errorf("renderCount=%d, want 1", renderCount)
	}
}

// --- Test 7: Move + SetState → only state-dirty re-renders ---

func TestRenderCount_MoveAndSetState_OnlyStateDirtyRerenders(t *testing.T) {
	app, _ := NewTestApp(60, 20)
	renderCounts := map[string]int{}

	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("c%d", i)
		x := (i - 1) * 15
		app.RegisterComponent(id, id, buffer.Rect{X: x, Y: 0, W: 10, H: 5}, 0,
			func(state, props map[string]any) *layout.VNode {
				renderCounts[id]++
				vn := layout.NewVNode("text")
				vn.Content = id
				return vn
			})
	}

	app.RenderAll()

	// Move c1 (same size), SetState on c2, leave c3 untouched.
	app.MoveComponent("c1", buffer.Rect{X: 45, Y: 10, W: 10, H: 5})
	app.SetState("c2", "val", 42)
	app.RenderDirty()

	if renderCounts["c1"] != 1 {
		t.Errorf("c1 (moved only) renderCount=%d, want 1", renderCounts["c1"])
	}
	if renderCounts["c2"] != 2 {
		t.Errorf("c2 (state changed) renderCount=%d, want 2", renderCounts["c2"])
	}
	if renderCounts["c3"] != 1 {
		t.Errorf("c3 (untouched) renderCount=%d, want 1", renderCounts["c3"])
	}
}

// --- Test 8: Multi-window move — screen correctness without re-render ---

func TestRenderCount_MultiWindowMove_ScreenCorrectness(t *testing.T) {
	app, ta := NewTestApp(80, 30)
	renderCounts := map[string]int{}
	chars := []string{"1", "2", "3", "4", "5"}

	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("w%d", i)
		content := chars[i]
		app.RegisterComponent(id, id, buffer.Rect{X: i * 12, Y: 0, W: 10, H: 5}, 0,
			func(state, props map[string]any) *layout.VNode {
				renderCounts[id]++
				root := layout.NewVNode("box")
				root.Style.Background = "#AAAAAA"
				txt := layout.NewVNode("text")
				txt.Content = content
				root.AddChild(txt)
				return root
			})
	}

	app.RenderAll()

	// Verify initial positions.
	for i, ch := range []rune{'1', '2', '3', '4', '5'} {
		x := i * 12
		if c := ta.LastScreen.Get(x, 0).Char; c != ch {
			t.Errorf("initial: expected %q at (%d,0), got %q", ch, x, c)
		}
	}

	// Move all 5 to new non-overlapping positions (same size).
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("w%d", i)
		app.MoveComponent(id, buffer.Rect{X: i * 12, Y: 15, W: 10, H: 5})
	}
	app.RenderDirty()

	// All should be at new positions.
	for i, ch := range []rune{'1', '2', '3', '4', '5'} {
		x := i * 12
		if c := ta.LastScreen.Get(x, 15).Char; c != ch {
			t.Errorf("moved: expected %q at (%d,15), got %q", ch, x, c)
		}
	}

	// Old positions should be empty.
	for i := 0; i < 5; i++ {
		x := i * 12
		if c := ta.LastScreen.Get(x, 0).Char; c != 0 {
			t.Errorf("old position (%d,0) not cleared: got %q", x, c)
		}
	}

	// No re-renders.
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("w%d", i)
		if renderCounts[id] != 1 {
			t.Errorf("%s renderCount=%d, want 1", id, renderCounts[id])
		}
	}
}

// --- Test 9: Overlapping move — recompose z-order without re-render ---

func TestRenderCount_MoveOverlapping_Recompose(t *testing.T) {
	app, ta := NewTestApp(60, 20)
	renderCounts := map[string]int{}

	// Window A at z=10.
	app.RegisterComponent("wA", "wA", buffer.Rect{X: 0, Y: 0, W: 20, H: 10}, 10,
		func(state, props map[string]any) *layout.VNode {
			renderCounts["wA"]++
			root := layout.NewVNode("box")
			root.Style.Background = "#AA0000"
			txt := layout.NewVNode("text")
			txt.Content = "A"
			root.AddChild(txt)
			return root
		})

	// Window B at z=20 (higher).
	app.RegisterComponent("wB", "wB", buffer.Rect{X: 30, Y: 0, W: 20, H: 10}, 20,
		func(state, props map[string]any) *layout.VNode {
			renderCounts["wB"]++
			root := layout.NewVNode("box")
			root.Style.Background = "#00BB00"
			txt := layout.NewVNode("text")
			txt.Content = "B"
			root.AddChild(txt)
			return root
		})

	app.RenderAll()

	// Before overlap: A at (0,0), B at (30,0).
	if c := ta.LastScreen.Get(0, 0).Char; c != 'A' {
		t.Fatalf("initial: expected 'A' at (0,0), got %q", c)
	}
	if c := ta.LastScreen.Get(30, 0).Char; c != 'B' {
		t.Fatalf("initial: expected 'B' at (30,0), got %q", c)
	}

	// Move B to overlap A at (5,0).
	app.MoveComponent("wB", buffer.Rect{X: 5, Y: 0, W: 20, H: 10})
	app.RenderDirty()

	// Overlap area: B has higher z, so should show 'B' at (5,0).
	if c := ta.LastScreen.Get(5, 0).Char; c != 'B' {
		t.Errorf("overlap: expected 'B' at (5,0), got %q", c)
	}
	// A's non-overlapped area (0,0) should still show 'A'.
	if c := ta.LastScreen.Get(0, 0).Char; c != 'A' {
		t.Errorf("non-overlap: expected 'A' at (0,0), got %q", c)
	}

	// Neither was re-rendered.
	if renderCounts["wA"] != 1 {
		t.Errorf("wA renderCount=%d, want 1", renderCounts["wA"])
	}
	if renderCounts["wB"] != 1 {
		t.Errorf("wB renderCount=%d, want 1", renderCounts["wB"])
	}
}

// --- Test 10: No dirty → RenderDirty is a no-op ---

func TestRenderCount_NoDirty_NoOp(t *testing.T) {
	app, ta := NewTestApp(30, 10)
	renderCounts := map[string]int{}

	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("c%d", i)
		app.RegisterComponent(id, id, buffer.Rect{X: i * 10, Y: 0, W: 10, H: 5}, 0,
			func(state, props map[string]any) *layout.VNode {
				renderCounts[id]++
				vn := layout.NewVNode("text")
				vn.Content = id
				return vn
			})
	}

	app.RenderAll()
	writesBefore := ta.WriteCount

	// Nothing dirty — RenderDirty should be a no-op.
	app.RenderDirty()

	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("c%d", i)
		if renderCounts[id] != 1 {
			t.Errorf("%s renderCount=%d, want 1", id, renderCounts[id])
		}
	}

	if ta.WriteCount != writesBefore {
		t.Errorf("WriteCount changed from %d to %d on no-op RenderDirty", writesBefore, ta.WriteCount)
	}

	if rects := app.DirtyRects(); len(rects) != 0 {
		t.Errorf("DirtyRects=%d, want 0", len(rects))
	}
}
