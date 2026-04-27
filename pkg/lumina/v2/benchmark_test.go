package v2

import (
	"fmt"
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/compositor"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
	"github.com/akzj/lumina/pkg/lumina/v2/paint"
)

// ─── No-op adapter (zero overhead, measures pure render path) ────────

type nopAdapter struct{}

func (nopAdapter) WriteFull(_ *buffer.Buffer) error                       { return nil }
func (nopAdapter) WriteDirty(_ *buffer.Buffer, _ []buffer.Rect) error     { return nil }
func (nopAdapter) Flush() error                                           { return nil }
func (nopAdapter) Close() error                                           { return nil }

var _ output.Adapter = nopAdapter{}

// ─── Constants ───────────────────────────────────────────────────────
// 4K resolution with 8×16 font: 3840/8 = 480 cols, 2160/16 = 135 rows = 64,800 cells
const (
	screenW = 480
	screenH = 135
)

// ─── Helper functions ────────────────────────────────────────────────

func makeFullScreenLayer(id string, w, h, z int) *compositor.Layer {
	buf := buffer.New(w, h)
	buf.Fill(buffer.Rect{X: 0, Y: 0, W: w, H: h}, buffer.Cell{Char: '·', Foreground: "#aaaaaa"})
	return &compositor.Layer{ID: id, Buffer: buf, Rect: buffer.Rect{X: 0, Y: 0, W: w, H: h}, ZIndex: z}
}

func makeDialogLayer(id string, x, y, w, h, z int) *compositor.Layer {
	buf := buffer.New(w, h)
	buf.Fill(buffer.Rect{X: 0, Y: 0, W: w, H: h}, buffer.Cell{Char: 'D', Foreground: "#ffffff", Background: "#333333"})
	return &compositor.Layer{ID: id, Buffer: buf, Rect: buffer.Rect{X: x, Y: y, W: w, H: h}, ZIndex: z}
}

func makeCellLayer(id string, x, y, z int) *compositor.Layer {
	buf := buffer.New(1, 1)
	buf.Set(0, 0, buffer.Cell{Char: '·', Foreground: "#aaaaaa"})
	return &compositor.Layer{ID: id, Buffer: buf, Rect: buffer.Rect{X: x, Y: y, W: 1, H: 1}, ZIndex: z}
}

var cellRenderFn = func(state, props map[string]any) *layout.VNode {
	vn := &layout.VNode{
		Type:    "text",
		Content: "·",
		Props:   make(map[string]any),
	}
	return vn
}

var dialogRenderFn = func(state, props map[string]any) *layout.VNode {
	vn := layout.NewVNode("vbox")
	vn.Style.Background = "#333333"
	vn.Style.Border = "single"
	for i := 0; i < 10; i++ {
		child := layout.NewVNode("text")
		child.Content = fmt.Sprintf("Dialog line %d with some content here", i)
		vn.AddChild(child)
	}
	return vn
}

var fullScreenRenderFn = func(state, props map[string]any) *layout.VNode {
	vn := layout.NewVNode("vbox")
	vn.Style.Background = "#1a1a1a"
	for i := 0; i < screenH; i++ {
		child := layout.NewVNode("text")
		child.Content = fmt.Sprintf("Line %03d: background content filling the entire screen width with text", i)
		vn.AddChild(child)
	}
	return vn
}

// ═══════════════════════════════════════════════════════════════════════
// GROUP 1: OcclusionMap Performance
// ═══════════════════════════════════════════════════════════════════════

func BenchmarkOcclusionMap_Build_FullScreen(b *testing.B) {
	om := compositor.NewOcclusionMap(screenW, screenH)
	bg := makeFullScreenLayer("bg", screenW, screenH, 0)
	layers := []*compositor.Layer{bg}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		om.Build(layers)
	}
}

func BenchmarkOcclusionMap_Build_FullScreen_10Dialogs(b *testing.B) {
	om := compositor.NewOcclusionMap(screenW, screenH)
	layers := make([]*compositor.Layer, 0, 11)
	layers = append(layers, makeFullScreenLayer("bg", screenW, screenH, 0))
	for i := 0; i < 10; i++ {
		x := (i % 5) * 80
		y := (i / 5) * 60
		layers = append(layers, makeDialogLayer(fmt.Sprintf("dlg-%d", i), x, y, 40, 20, i+1))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		om.Build(layers)
	}
}

func BenchmarkOcclusionMap_Build_FullScreen_100Windows(b *testing.B) {
	om := compositor.NewOcclusionMap(screenW, screenH)
	layers := make([]*compositor.Layer, 0, 101)
	layers = append(layers, makeFullScreenLayer("bg", screenW, screenH, 0))
	for i := 0; i < 100; i++ {
		x := (i % 48) * 10
		y := (i / 48) * 5
		layers = append(layers, makeDialogLayer(fmt.Sprintf("win-%d", i), x, y, 10, 5, i+1))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		om.Build(layers)
	}
}

func BenchmarkOcclusionMap_Build_1000Cells(b *testing.B) {
	om := compositor.NewOcclusionMap(screenW, screenH)
	layers := make([]*compositor.Layer, 0, 1000)
	for i := 0; i < 1000; i++ {
		x := i % screenW
		y := i / screenW
		layers = append(layers, makeCellLayer(fmt.Sprintf("cell-%d", i), x, y, 0))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		om.Build(layers)
	}
}

func BenchmarkOcclusionMap_Build_10000Cells(b *testing.B) {
	om := compositor.NewOcclusionMap(screenW, screenH)
	layers := make([]*compositor.Layer, 0, 10000)
	for i := 0; i < 10000; i++ {
		x := i % screenW
		y := i / screenW
		layers = append(layers, makeCellLayer(fmt.Sprintf("cell-%d", i), x, y, 0))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		om.Build(layers)
	}
}

func BenchmarkOcclusionMap_Owner(b *testing.B) {
	om := compositor.NewOcclusionMap(screenW, screenH)
	layers := make([]*compositor.Layer, 0, 11)
	layers = append(layers, makeFullScreenLayer("bg", screenW, screenH, 0))
	for i := 0; i < 10; i++ {
		x := (i % 5) * 80
		y := (i / 5) * 60
		layers = append(layers, makeDialogLayer(fmt.Sprintf("dlg-%d", i), x, y, 40, 20, i+1))
	}
	om.Build(layers)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = om.Owner(i%screenW, i%screenH)
	}
}

// ═══════════════════════════════════════════════════════════════════════
// GROUP 2: Compositor Performance
// ═══════════════════════════════════════════════════════════════════════

func BenchmarkCompositor_ComposeAll_FullScreen(b *testing.B) {
	c := compositor.NewCompositor(screenW, screenH)
	bg := makeFullScreenLayer("bg", screenW, screenH, 0)
	c.SetLayers([]*compositor.Layer{bg})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ComposeAll()
	}
}

func BenchmarkCompositor_ComposeAll_FullScreen_10Dialogs(b *testing.B) {
	c := compositor.NewCompositor(screenW, screenH)
	layers := make([]*compositor.Layer, 0, 11)
	layers = append(layers, makeFullScreenLayer("bg", screenW, screenH, 0))
	for i := 0; i < 10; i++ {
		x := (i % 5) * 80
		y := (i / 5) * 60
		layers = append(layers, makeDialogLayer(fmt.Sprintf("dlg-%d", i), x, y, 40, 20, i+1))
	}
	c.SetLayers(layers)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ComposeAll()
	}
}

func BenchmarkCompositor_ComposeDirty_SingleCell(b *testing.B) {
	c := compositor.NewCompositor(screenW, screenH)
	bg := makeFullScreenLayer("bg", screenW, screenH, 0)
	c.SetLayers([]*compositor.Layer{bg})
	c.ComposeAll()

	// A single dirty cell layer
	dirtyLayer := &compositor.Layer{
		ID:     "bg",
		Buffer: bg.Buffer,
		Rect:   bg.Rect,
		ZIndex: 0,
		DirtyRect: &buffer.Rect{X: 100, Y: 50, W: 1, H: 1},
	}
	dirtyLayers := []*compositor.Layer{dirtyLayer}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ComposeDirty(dirtyLayers)
	}
}

func BenchmarkCompositor_ComposeDirty_Dialog(b *testing.B) {
	c := compositor.NewCompositor(screenW, screenH)
	layers := make([]*compositor.Layer, 0, 11)
	bg := makeFullScreenLayer("bg", screenW, screenH, 0)
	layers = append(layers, bg)
	dlg := makeDialogLayer("dlg-0", 100, 30, 40, 20, 1)
	layers = append(layers, dlg)
	c.SetLayers(layers)
	c.ComposeAll()

	// Dialog content changes — entire dialog is dirty
	dirtyLayer := &compositor.Layer{
		ID:     "dlg-0",
		Buffer: dlg.Buffer,
		Rect:   dlg.Rect,
		ZIndex: 1,
	}
	dirtyLayers := []*compositor.Layer{dirtyLayer}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ComposeDirty(dirtyLayers)
	}
}

func BenchmarkCompositor_ComposeRects_WindowMove(b *testing.B) {
	c := compositor.NewCompositor(screenW, screenH)
	layers := make([]*compositor.Layer, 0, 2)
	layers = append(layers, makeFullScreenLayer("bg", screenW, screenH, 0))
	layers = append(layers, makeDialogLayer("dlg-0", 100, 30, 40, 20, 1))
	c.SetLayers(layers)
	c.ComposeAll()

	// Simulate window move: old rect + new rect
	rects := []buffer.Rect{
		{X: 100, Y: 30, W: 40, H: 20}, // old position
		{X: 150, Y: 40, W: 40, H: 20}, // new position
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ComposeRects(rects)
	}
}

// ═══════════════════════════════════════════════════════════════════════
// GROUP 3: Full Pipeline Performance (App level)
// ═══════════════════════════════════════════════════════════════════════

func BenchmarkRenderAll_FullScreen_1Component(b *testing.B) {
	app, _ := NewTestApp(screenW, screenH)
	app.RegisterComponent("bg", "background",
		buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}, 0, fullScreenRenderFn)
	app.RenderAll() // warm up
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mark dirty again for re-render
		app.SetState("bg", "tick", i)
		app.RenderAll()
	}
}

func BenchmarkRenderAll_FullScreen_10Dialogs(b *testing.B) {
	app, _ := NewTestApp(screenW, screenH)
	app.RegisterComponent("bg", "background",
		buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}, 0, fullScreenRenderFn)
	for i := 0; i < 10; i++ {
		x := (i % 5) * 80
		y := (i / 5) * 60
		app.RegisterComponent(fmt.Sprintf("dlg-%d", i), "dialog",
			buffer.Rect{X: x, Y: y, W: 40, H: 20}, i+1, dialogRenderFn)
	}
	app.RenderAll() // warm up
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.SetState("bg", "tick", i)
		for j := 0; j < 10; j++ {
			app.SetState(fmt.Sprintf("dlg-%d", j), "tick", i)
		}
		app.RenderAll()
	}
}

func BenchmarkRenderDirty_SingleCellHover_4K(b *testing.B) {
	app, _ := NewTestApp(screenW, screenH)

	// Register background
	app.RegisterComponent("bg", "background",
		buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}, 0, fullScreenRenderFn)

	// Register 100 small cell components (simulating interactive cells)
	for i := 0; i < 100; i++ {
		x := (i % 50) * 2
		y := (i / 50) * 2
		id := fmt.Sprintf("cell-%d", i)
		app.RegisterComponent(id, "cell",
			buffer.Rect{X: x, Y: y, W: 1, H: 1}, 1, cellRenderFn)
	}

	// Initial full render
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate hover: change state of one cell → RenderDirty
		cellID := fmt.Sprintf("cell-%d", i%100)
		app.SetState(cellID, "hover", true)
		app.RenderDirty()
	}
}

func BenchmarkRenderDirty_DialogContentUpdate(b *testing.B) {
	app, _ := NewTestApp(screenW, screenH)

	app.RegisterComponent("bg", "background",
		buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}, 0, fullScreenRenderFn)
	app.RegisterComponent("dlg-0", "dialog",
		buffer.Rect{X: 100, Y: 30, W: 40, H: 20}, 1, dialogRenderFn)

	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.SetState("dlg-0", "content", fmt.Sprintf("Updated content %d", i))
		app.RenderDirty()
	}
}

func BenchmarkHandleEvent_MouseMove_HitTest(b *testing.B) {
	app, _ := NewTestApp(screenW, screenH)

	// Register background with event handler
	bgRenderFn := func(state, props map[string]any) *layout.VNode {
		vn := layout.NewVNode("box")
		vn.ID = "bg-root"
		vn.Style.Background = "#1a1a1a"
		vn.Props["onMouseMove"] = event.EventHandler(func(e *event.Event) {})
		for i := 0; i < 10; i++ {
			child := layout.NewVNode("text")
			child.ID = fmt.Sprintf("text-%d", i)
			child.Content = fmt.Sprintf("Line %d", i)
			child.Props = make(map[string]any)
			child.Props["onClick"] = event.EventHandler(func(e *event.Event) {})
			vn.AddChild(child)
		}
		return vn
	}

	app.RegisterComponent("bg", "background",
		buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}, 0, bgRenderFn)
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.HandleEvent(&event.Event{
			Type: "mousemove",
			X:    i % screenW,
			Y:    i % screenH,
		})
	}
}

// ═══════════════════════════════════════════════════════════════════════
// GROUP 3b: Pure Render Path (no-op adapter — no output overhead)
// ═══════════════════════════════════════════════════════════════════════

func BenchmarkRenderDirty_SingleCellHover_4K_NopAdapter(b *testing.B) {
	app := NewApp(screenW, screenH, nopAdapter{})

	app.RegisterComponent("bg", "background",
		buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}, 0, fullScreenRenderFn)
	for i := 0; i < 100; i++ {
		x := (i % 50) * 2
		y := (i / 50) * 2
		id := fmt.Sprintf("cell-%d", i)
		app.RegisterComponent(id, "cell",
			buffer.Rect{X: x, Y: y, W: 1, H: 1}, 1, cellRenderFn)
	}
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cellID := fmt.Sprintf("cell-%d", i%100)
		app.SetState(cellID, "hover", true)
		app.RenderDirty()
	}
}

func BenchmarkRenderDirty_DialogContentUpdate_NopAdapter(b *testing.B) {
	app := NewApp(screenW, screenH, nopAdapter{})

	app.RegisterComponent("bg", "background",
		buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}, 0, fullScreenRenderFn)
	app.RegisterComponent("dlg-0", "dialog",
		buffer.Rect{X: 100, Y: 30, W: 40, H: 20}, 1, dialogRenderFn)
	app.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.SetState("dlg-0", "content", fmt.Sprintf("Updated content %d", i))
		app.RenderDirty()
	}
}

func BenchmarkRenderAll_FullScreen_1Component_NopAdapter(b *testing.B) {
	app := NewApp(screenW, screenH, nopAdapter{})
	app.RegisterComponent("bg", "background",
		buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}, 0, fullScreenRenderFn)
	app.RenderAll()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.SetState("bg", "tick", i)
		app.RenderAll()
	}
}

// ═══════════════════════════════════════════════════════════════════════
// GROUP 4: Memory Benchmarks
// ═══════════════════════════════════════════════════════════════════════

func BenchmarkMemory_1000_Cells(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		app, _ := NewTestApp(screenW, screenH)
		for j := 0; j < 1000; j++ {
			x := j % screenW
			y := j / screenW
			app.RegisterComponent(fmt.Sprintf("cell-%d", j), "cell",
				buffer.Rect{X: x, Y: y, W: 1, H: 1}, 0, cellRenderFn)
		}
		app.RenderAll()
	}
}

func BenchmarkMemory_10000_Cells(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		app, _ := NewTestApp(screenW, screenH)
		for j := 0; j < 10000; j++ {
			x := j % screenW
			y := j / screenW
			app.RegisterComponent(fmt.Sprintf("cell-%d", j), "cell",
				buffer.Rect{X: x, Y: y, W: 1, H: 1}, 0, cellRenderFn)
		}
		app.RenderAll()
	}
}

func BenchmarkMemory_10_Dialogs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		app, _ := NewTestApp(screenW, screenH)
		app.RegisterComponent("bg", "background",
			buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}, 0, fullScreenRenderFn)
		for j := 0; j < 10; j++ {
			x := (j % 5) * 80
			y := (j / 5) * 60
			app.RegisterComponent(fmt.Sprintf("dlg-%d", j), "dialog",
				buffer.Rect{X: x, Y: y, W: 40, H: 20}, j+1, dialogRenderFn)
		}
		app.RenderAll()
	}
}

// ═══════════════════════════════════════════════════════════════════════
// GROUP 5: Layout + Paint Performance
// ═══════════════════════════════════════════════════════════════════════

func BenchmarkLayout_DeepTree_100Children(b *testing.B) {
	root := layout.NewVNode("vbox")
	for i := 0; i < 100; i++ {
		child := layout.NewVNode("text")
		child.Content = fmt.Sprintf("Item %d with some text content", i)
		root.AddChild(child)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layout.ComputeLayout(root, 0, 0, screenW, screenH)
	}
}

func BenchmarkLayout_WideTree_10Levels(b *testing.B) {
	// Build a 10-level deep tree: each level has 3 children
	var buildTree func(depth int) *layout.VNode
	buildTree = func(depth int) *layout.VNode {
		if depth == 0 {
			vn := layout.NewVNode("text")
			vn.Content = "leaf"
			return vn
		}
		vn := layout.NewVNode("vbox")
		for i := 0; i < 3; i++ {
			vn.AddChild(buildTree(depth - 1))
		}
		return vn
	}
	root := buildTree(10) // 3^10 = 59,049 leaf nodes
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layout.ComputeLayout(root, 0, 0, screenW, screenH)
	}
}

func BenchmarkPaint_FullScreen(b *testing.B) {
	root := layout.NewVNode("vbox")
	root.Style.Background = "#1a1a1a"
	for i := 0; i < screenH; i++ {
		child := layout.NewVNode("text")
		child.Content = fmt.Sprintf("Line %03d: filling screen with text content for paint benchmark", i)
		root.AddChild(child)
	}
	layout.ComputeLayout(root, 0, 0, screenW, screenH)

	p := paint.NewPainter()
	buf := buffer.New(screenW, screenH)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
		p.Paint(buf, root, 0, 0)
	}
}

func BenchmarkPaint_SingleCell(b *testing.B) {
	root := layout.NewVNode("text")
	root.Content = "·"
	layout.ComputeLayout(root, 0, 0, 1, 1)

	p := paint.NewPainter()
	buf := buffer.New(1, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
		p.Paint(buf, root, 0, 0)
	}
}

// ═══════════════════════════════════════════════════════════════════════
// GROUP 6: Buffer Operations
// ═══════════════════════════════════════════════════════════════════════

func BenchmarkBuffer_New_4K(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = buffer.New(screenW, screenH)
	}
}

func BenchmarkBuffer_Fill_4K(b *testing.B) {
	buf := buffer.New(screenW, screenH)
	cell := buffer.Cell{Char: '·', Foreground: "#aaaaaa"}
	rect := buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Fill(rect, cell)
	}
}

func BenchmarkBuffer_Clear_4K(b *testing.B) {
	buf := buffer.New(screenW, screenH)
	buf.Fill(buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}, buffer.Cell{Char: 'X'})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
	}
}

func BenchmarkBuffer_Blit_Dialog(b *testing.B) {
	dst := buffer.New(screenW, screenH)
	src := buffer.New(40, 20)
	src.Fill(buffer.Rect{X: 0, Y: 0, W: 40, H: 20}, buffer.Cell{Char: 'D', Foreground: "#ffffff"})
	clip := buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.Blit(dst, src, 100, 30, clip)
	}
}

// ═══════════════════════════════════════════════════════════════════════
// GROUP 7: Component Manager
// ═══════════════════════════════════════════════════════════════════════

func BenchmarkManager_RenderDirty_1Component(b *testing.B) {
	p := paint.NewPainter()
	mgr := component.NewManager(p)
	comp := &component.Component{
		ID:         "bg",
		Name:       "background",
		Buffer:     buffer.New(screenW, screenH),
		Rect:       buffer.Rect{X: 0, Y: 0, W: screenW, H: screenH},
		ZIndex:     0,
		DirtyPaint: true,
		State:      make(map[string]any),
		Props:      make(map[string]any),
		RenderFn:   fullScreenRenderFn,
		ChildMap:   make(map[string]*component.Component),
		Handlers:   make(map[string]event.HandlerMap),
	}
	mgr.Register(comp)
	mgr.RenderDirty() // warm up

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.DirtyPaint = true
		mgr.RenderDirty()
	}
}

func BenchmarkManager_RenderDirty_100CellComponents(b *testing.B) {
	p := paint.NewPainter()
	mgr := component.NewManager(p)
	comps := make([]*component.Component, 100)
	for i := 0; i < 100; i++ {
		comp := &component.Component{
			ID:         fmt.Sprintf("cell-%d", i),
			Name:       "cell",
			Buffer:     buffer.New(1, 1),
			Rect:       buffer.Rect{X: i % screenW, Y: i / screenW, W: 1, H: 1},
			ZIndex:     0,
			DirtyPaint: true,
			State:      make(map[string]any),
			Props:      make(map[string]any),
			RenderFn:   cellRenderFn,
			ChildMap:   make(map[string]*component.Component),
			Handlers:   make(map[string]event.HandlerMap),
		}
		mgr.Register(comp)
		comps[i] = comp
	}
	mgr.RenderDirty() // warm up

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mark just 1 cell dirty (hover scenario)
		comps[i%100].DirtyPaint = true
		mgr.RenderDirty()
	}
}
