package v2

import (
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/perf"
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
// 执行次数由 perf.Tracker（RenderObserver → RecordComponent）断言，与手写计数解耦。
//
// =============================================================================

func TestScenario_ChildMoveUnderCurtain_OtherChildrenNotRerendered(t *testing.T) {
	const W, H = 60, 12

	app, ta := NewTestApp(W, H)
	app.Tracker().Enable()

	parent := app.RegisterComponent("parent", "parent", buffer.Rect{X: 0, Y: 0, W: W, H: H}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "parent-root"
			root.Style.Background = "#202020"
			return root
		})

	childDescriptors := []component.ChildDescriptor{
		{
			Key: "a", Name: "childA", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
				ch := "A"
				if g, ok := state["glyph"].(string); ok && g != "" {
					ch = g
				}
				tn := layout.NewVNode("text")
				tn.ID = "a-t"
				tn.Content = ch
				return tn
			},
		},
		{
			Key: "b", Name: "childB", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
				tn := layout.NewVNode("text")
				tn.ID = "b-t"
				tn.Content = "B"
				return tn
			},
		},
		{
			Key: "c", Name: "childC", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
				tn := layout.NewVNode("text")
				tn.ID = "c-t"
				tn.Content = "C"
				return tn
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

	// --- (1) Baseline: five components each render once (perf Renders == 5).
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(5),
		perf.CheckLayouts(5),
		perf.CheckPaints(5),
		perf.CheckRenderComponents("curtain", "parent", "parent:a", "parent:b", "parent:c"),
	)

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
	if bg := ta.LastScreen.Get(10, 5).Background; bg != "#606060" {
		t.Fatalf("curtain row y=5: want bg #606060, got %q", bg)
	}

	// --- (3) Move only child-b down (position-only); recompose without re-render ---
	app.MoveComponent("parent:b", buffer.Rect{X: 20, Y: 8, W: 20, H: 10})
	app.RenderDirty()

	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(0),
		perf.CheckPaints(0),
		perf.CheckLayouts(1), // vnode abs coords for moved layer (no renderFn)
		perf.CheckOcclusionBuilds(1),
		perf.CheckMetric(perf.ComposeRects, 1),
		perf.CheckHitTesterRebuilds(1),
		perf.CheckHandlerFullSyncs(1),
	)

	// --- (4) Composited outcome ---
	if ch := ta.LastScreen.Get(20, 8).Char; ch != 'B' {
		t.Errorf("(20,8) want 'B' after move, got %q", ch)
	}
	if ch := ta.LastScreen.Get(0, 0).Char; ch != 'A' {
		t.Errorf("(0,0) want 'A' preserved, got %q", ch)
	}
	if ch := ta.LastScreen.Get(40, 0).Char; ch != 'C' {
		t.Errorf("(40,0) want 'C' preserved, got %q", ch)
	}
	if ch := ta.LastScreen.Get(20, 0).Char; ch == 'B' {
		t.Error("(20,0) should not still show 'B' after B moved away")
	}
	if bg := ta.LastScreen.Get(10, 5).Background; bg != "#606060" {
		t.Errorf("(10,5) curtain bg want #606060, got %q", bg)
	}

	// --- (5) Mark only child-a dirty: exactly one component render ---
	app.SetState("parent:a", "glyph", "X")
	app.RenderDirty()

	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(1),
		perf.CheckLayouts(1),
		perf.CheckPaints(1),
		perf.CheckRenderComponents("parent:a"),
		perf.CheckOcclusionUpdates(1),
		perf.CheckMetric(perf.ComposeDirty, 1),
		perf.CheckHandlerDirtySyncs(1),
	)

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
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(0),
		perf.CheckPaints(0),
		perf.CheckLayouts(0),
	)
}

// TestScenario_ParentA_NestedReconcile_NoSiblingRerender checks that Reconcile
// on parent:a to add a nested child component does not mark siblings (parent:b)
// or the root parent dirty — only the new subtree renders once.
func TestScenario_ParentA_NestedReconcile_NoSiblingRerender(t *testing.T) {
	const W, H = 40, 12

	app, ta := NewTestApp(W, H)
	app.Tracker().Enable()

	parent := app.RegisterComponent("parent", "parent", buffer.Rect{X: 0, Y: 0, W: W, H: H}, 0,
		func(state, props map[string]any) *layout.VNode {
			root := layout.NewVNode("box")
			root.ID = "parent-root"
			root.Style.Background = "#101010"
			return root
		})

	top := []component.ChildDescriptor{
		{
			Key: "a", Name: "childA", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
				tn := layout.NewVNode("text")
				tn.ID = "a-t"
				tn.Content = "A"
				return tn
			},
		},
		{
			Key: "b", Name: "childB", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
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

	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(3),
		perf.CheckLayouts(3),
		perf.CheckPaints(3),
		perf.CheckRenderComponents("parent", "parent:a", "parent:b"),
	)

	childA := app.manager.Get("parent:a")
	if childA == nil {
		t.Fatal("parent:a not found")
	}

	app.manager.Reconcile(childA, []component.ChildDescriptor{
		{
			Key: "sub", Name: "nestedUnderA", Props: nil,
			RenderFn: func(state, props map[string]any) *layout.VNode {
				tn := layout.NewVNode("text")
				tn.ID = "sub-t"
				tn.Content = "D"
				return tn
			},
		},
	})

	app.MoveComponent("parent:a:sub", buffer.Rect{X: 3, Y: 2, W: 1, H: 1})
	app.RenderDirty()

	// New nested layer renders once; manager layout+paint for it, then App
	// re-layouts vnode abs coords for rect-changed layer without a second render.
	app.Tracker().AssertLastFrame(t,
		perf.CheckRenders(1),
		perf.CheckLayouts(2),
		perf.CheckPaints(1),
		perf.CheckRenderComponents("parent:a:sub"),
		perf.CheckOcclusionBuilds(1),
		perf.CheckMetric(perf.ComposeRects, 1),
		perf.CheckHitTesterRebuilds(1),
		perf.CheckHandlerFullSyncs(1),
	)

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
