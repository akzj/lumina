package v2

import (
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
)

// =============================================================================
// Scenario: parent + 3 child components + middle curtain (z above children)
// =============================================================================
//
// 屏幕 60×12，列 x=0..59。`.` = 该格无子层字形，合成后多为 parent 的空格+#202020。
// Reconcile 子组件 z = parent.zIndex+1（本例 parent z=0 → 子 z=1）；curtain z=20。
//
// --- 合成层（概念）---
//   z=0  parent     整屏 box #202020
//   z=1  child-a    rect (0,0)-(19,9)   文本根 → 字形 'A' 仅在缓冲 (0,0) → 屏 (0,0)
//   z=1  child-b    rect (20,0)-(39,9) 初态；可移到 (20,8)
//   z=1  child-c    rect (40,0)-(59,9)
//   z=20 curtain   rect (0,5)-(59,7)  高 3 行，#606060，盖住三子中间一段
//
// --- 渲染图像：字符层（移动前 RenderAll）---
// 列标尺:    0         10        20        30        40        50        59
// y=0   A...................B...................C...................
// y=5   ============================================================  <- curtain 带内为底色条；'=' 可能因 flex 落在带内某列
// y=7   ============================================================  （同上，共 y=5,6,7 三行）
//
// --- 渲染图像：字符层（仅 child-b 移到 (20,8) + RenderDirty 后）---
// y=0   A...................#...................C...................
//       (#) 仅示意「(20,0) 已无 B」；实际该格为 parent 的合成空格，非字符 '#'
// y=5   ============================================================  （与移动前相同；a/c 仍被挡一段，测 (10,5) 背景 #606060）
// y=8   ....................B.......................................
//       'B' 在屏 (20,8) = 子缓冲 (0,0)，低于 curtain 下沿 y=7，完全可见
//
// 仅把 child-b 从 (20,0,20,10) 移到 (20,8,20,10) 后:
//   - child-a / child-c 不应再次执行 RenderFn（无 DirtyPaint）
//   - child-b 也不应再次 render（position-only move, same W/H → no re-render）
//   - 旧 B 区域由 parent 等下层补洞；新位置可见 'B'（recompose only）
//
// =============================================================================

func TestScenario_ChildMoveUnderCurtain_OtherChildrenNotRerendered(t *testing.T) {
	const W, H = 60, 12

	app, ta := NewTestApp(W, H)

	renderCount := map[string]int{}

	parent := app.RegisterComponent("parent", "parent", buffer.Rect{X: 0, Y: 0, W: W, H: H}, 0,
		func(state, props map[string]any) *layout.VNode {
			renderCount["parent"]++
			root := layout.NewVNode("box")
			root.ID = "parent-root"
			root.Style.Background = "#202020"
			return root
		})

	childDescriptors := []component.ChildDescriptor{
		{
			Key: "a", Name: "childA", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
				renderCount["parent:a"]++
				ch := "A"
				if g, ok := state["glyph"].(string); ok && g != "" {
					ch = g
				}
				t := layout.NewVNode("text")
				t.ID = "a-t"
				t.Content = ch
				return t
			},
		},
		{
			Key: "b", Name: "childB", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
				renderCount["parent:b"]++
				t := layout.NewVNode("text")
				t.ID = "b-t"
				t.Content = "B"
				return t
			},
		},
		{
			Key: "c", Name: "childC", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
				renderCount["parent:c"]++
				t := layout.NewVNode("text")
				t.ID = "c-t"
				t.Content = "C"
				return t
			},
		},
	}

	app.manager.Reconcile(parent, childDescriptors)

	// Place children side-by-side (Reconcile starts them at 1×1).
	app.MoveComponent("parent:a", buffer.Rect{X: 0, Y: 0, W: 20, H: 10})
	app.MoveComponent("parent:b", buffer.Rect{X: 20, Y: 0, W: 20, H: 10})
	app.MoveComponent("parent:c", buffer.Rect{X: 40, Y: 0, W: 20, H: 10})

	// Horizontal curtain: rows y=5..7 (still occludes part of each child), full width.
	app.RegisterComponent("curtain", "curtain", buffer.Rect{X: 0, Y: 5, W: W, H: 3}, 20,
		func(state, props map[string]any) *layout.VNode {
			renderCount["curtain"]++
			root := layout.NewVNode("box")
			root.ID = "curtain-root"
			root.Style.Background = "#606060"
			t := layout.NewVNode("text")
			t.Content = "="
			t.ID = "curtain-t"
			root.AddChild(t)
			return root
		})

	app.RenderAll()

	// --- (1) Baseline render counts: each child + parent + curtain once ---
	if renderCount["parent"] != 1 {
		t.Fatalf("parent renders: want 1, got %d", renderCount["parent"])
	}
	for _, k := range []string{"parent:a", "parent:b", "parent:c", "curtain"} {
		if renderCount[k] != 1 {
			t.Fatalf("%s renders after initial RenderAll: want 1, got %d", k, renderCount[k])
		}
	}

	// --- (2) Visual baseline: A / B / C top row; curtain covers y=5 in x bands ---
	if ch := ta.LastScreen.Get(0, 0).Char; ch != 'A' {
		t.Fatalf("(0,0) want 'A', got %q", ch)
	}
	if ch := ta.LastScreen.Get(20, 0).Char; ch != 'B' {
		t.Fatalf("(20,0) want 'B', got %q", ch)
	}
	if ch := ta.LastScreen.Get(40, 0).Char; ch != 'C' {
		t.Fatalf("(40,0) want 'C', got %q", ch)
	}
	// Curtain paints '=' inside its buffer; exact x may vary with flex layout — use background.
	if bg := ta.LastScreen.Get(10, 5).Background; bg != "#606060" {
		t.Fatalf("curtain row y=5: want bg #606060, got %q", bg)
	}

	// --- (3) Move only child-b down so its first text row (y=8) clears the curtain band ---
	app.MoveComponent("parent:b", buffer.Rect{X: 20, Y: 8, W: 20, H: 10})
	app.RenderDirty()

	// --- (4) Unmoved children: no additional render ---
	if renderCount["parent:a"] != 1 || renderCount["parent:c"] != 1 {
		t.Errorf("unmoved children should not re-render: a=%d c=%d (want 1 each)",
			renderCount["parent:a"], renderCount["parent:c"])
	}
	if renderCount["parent"] != 1 {
		t.Errorf("parent should not re-render: got %d (want 1)", renderCount["parent"])
	}
	if renderCount["curtain"] != 1 {
		t.Errorf("curtain should not re-render: got %d (want 1)", renderCount["curtain"])
	}
	// Position-only move (same W, H) should NOT re-render — buffer content
	// is identical in component-local coordinates, only recompose is needed.
	if renderCount["parent:b"] != 1 {
		t.Errorf("moved child-b (position-only) should NOT re-render: got %d (want 1)", renderCount["parent:b"])
	}

	// --- (5) Composited outcome ---
	// B's first text row is at screen y=8 (below curtain rows 5–7).
	if ch := ta.LastScreen.Get(20, 8).Char; ch != 'B' {
		t.Errorf("(20,8) want 'B' after move, got %q", ch)
	}
	// A and C top rows unchanged.
	if ch := ta.LastScreen.Get(0, 0).Char; ch != 'A' {
		t.Errorf("(0,0) want 'A' preserved, got %q", ch)
	}
	if ch := ta.LastScreen.Get(40, 0).Char; ch != 'C' {
		t.Errorf("(40,0) want 'C' preserved, got %q", ch)
	}
	// Old B top-left (20,0): B layer no longer occupies → parent (or empty) wins; not 'B'.
	if ch := ta.LastScreen.Get(20, 0).Char; ch == 'B' {
		t.Error("(20,0) should not still show 'B' after B moved away")
	}
	// Curtain band still present over A at y=5.
	if bg := ta.LastScreen.Get(10, 5).Background; bg != "#606060" {
		t.Errorf("(10,5) curtain bg want #606060, got %q", bg)
	}

	// --- (6) Mark only child-a dirty: one RenderDirty → exactly one extra parent:a render ---
	app.SetState("parent:a", "glyph", "X")
	app.RenderDirty()

	if renderCount["parent:a"] != 2 {
		t.Fatalf("parent:a renders: want 2 (initial + one dirty pass), got %d", renderCount["parent:a"])
	}
	if renderCount["parent:b"] != 2 || renderCount["parent:c"] != 1 {
		t.Errorf("siblings should not re-render: b=%d (want 2) c=%d (want 1)",
			renderCount["parent:b"], renderCount["parent:c"])
	}
	if renderCount["parent"] != 1 || renderCount["curtain"] != 1 {
		t.Errorf("parent/curtain should not re-render: parent=%d curtain=%d (want 1 each)",
			renderCount["parent"], renderCount["curtain"])
	}
	if ch := ta.LastScreen.Get(0, 0).Char; ch != 'X' {
		t.Errorf("(0,0) want updated glyph 'X', got %q", ch)
	}
	if ch := ta.LastScreen.Get(20, 8).Char; ch != 'B' {
		t.Errorf("(20,8) want 'B' unchanged, got %q", ch)
	}
	if ch := ta.LastScreen.Get(40, 0).Char; ch != 'C' {
		t.Errorf("(40,0) want 'C' unchanged, got %q", ch)
	}
	if bg := ta.LastScreen.Get(10, 5).Background; bg != "#606060" {
		t.Errorf("(10,5) curtain bg want #606060, got %q", bg)
	}

	app.RenderDirty()
	if renderCount["parent:a"] != 2 {
		t.Errorf("no-op RenderDirty: parent:a render count should stay 2, got %d", renderCount["parent:a"])
	}
}

// TestScenario_ParentA_NestedReconcile_NoSiblingRerender checks that Reconcile
// on parent:a to add a nested child component does not mark siblings (parent:b)
// or the root parent dirty — only the new subtree renders once.
func TestScenario_ParentA_NestedReconcile_NoSiblingRerender(t *testing.T) {
	const W, H = 40, 12

	app, ta := NewTestApp(W, H)

	renderCount := map[string]int{}

	parent := app.RegisterComponent("parent", "parent", buffer.Rect{X: 0, Y: 0, W: W, H: H}, 0,
		func(state, props map[string]any) *layout.VNode {
			renderCount["parent"]++
			root := layout.NewVNode("box")
			root.ID = "parent-root"
			root.Style.Background = "#101010"
			return root
		})

	top := []component.ChildDescriptor{
		{
			Key: "a", Name: "childA", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
				renderCount["parent:a"]++
				tn := layout.NewVNode("text")
				tn.ID = "a-t"
				tn.Content = "A"
				return tn
			},
		},
		{
			Key: "b", Name: "childB", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
				renderCount["parent:b"]++
				tn := layout.NewVNode("text")
				tn.ID = "b-t"
				tn.Content = "B"
				return tn
			},
		},
	}

	app.manager.Reconcile(parent, top)
	app.MoveComponent("parent:a", buffer.Rect{X: 0, Y: 0, W: 20, H: 10})
	app.MoveComponent("parent:b", buffer.Rect{X: 20, Y: 0, W: 20, H: 10})
	app.RenderAll()

	if renderCount["parent"] != 1 || renderCount["parent:a"] != 1 || renderCount["parent:b"] != 1 {
		t.Fatalf("baseline renders want parent=1 a=1 b=1, got parent=%d a=%d b=%d",
			renderCount["parent"], renderCount["parent:a"], renderCount["parent:b"])
	}

	childA := app.manager.Get("parent:a")
	if childA == nil {
		t.Fatal("parent:a not found")
	}

	app.manager.Reconcile(childA, []component.ChildDescriptor{
		{
			Key: "sub", Name: "nestedUnderA", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
				renderCount["parent:a:sub"]++
				tn := layout.NewVNode("text")
				tn.ID = "sub-t"
				tn.Content = "D"
				return tn
			},
		},
	})

	// New layer + rect change → compositor rebuild; no need to touch sibling buffers.
	app.MoveComponent("parent:a:sub", buffer.Rect{X: 3, Y: 2, W: 1, H: 1})
	app.RenderDirty()

	if renderCount["parent:a:sub"] != 1 {
		t.Fatalf("nested parent:a:sub renders want 1, got %d", renderCount["parent:a:sub"])
	}
	if renderCount["parent:a"] != 1 || renderCount["parent:b"] != 1 || renderCount["parent"] != 1 {
		t.Errorf("sibling/root must not re-render: a=%d b=%d parent=%d (want 1 each)",
			renderCount["parent:a"], renderCount["parent:b"], renderCount["parent"])
	}

	if ch := ta.LastScreen.Get(0, 0).Char; ch != 'A' {
		t.Errorf("(0,0) want 'A', got %q", ch)
	}
	if ch := ta.LastScreen.Get(20, 0).Char; ch != 'B' {
		t.Errorf("(20,0) want 'B', got %q", ch)
	}
	if ch := ta.LastScreen.Get(3, 2).Char; ch != 'D' {
		t.Errorf("(3,2) nested want 'D', got %q", ch)
	}
}
