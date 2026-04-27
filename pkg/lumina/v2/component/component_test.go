package component

import (
	"testing"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/paint"
)

// simpleRender returns a RenderFunc that creates a box with a text child.
func simpleRender(text string) RenderFunc {
	return func(state map[string]any, props map[string]any) *layout.VNode {
		root := layout.NewVNode("box")
		root.ID = "root"
		child := layout.NewVNode("text")
		child.ID = "text1"
		child.Content = text
		root.AddChild(child)
		return root
	}
}

// newTestComponent creates a component with a buffer and rect for testing.
func newTestComponent(id, name string, w, h int, renderFn RenderFunc) *Component {
	return &Component{
		ID:         id,
		Name:       name,
		Buffer:     buffer.New(w, h),
		Rect:       buffer.Rect{X: 0, Y: 0, W: w, H: h},
		State:      make(map[string]any),
		Props:      make(map[string]any),
		RenderFn:   renderFn,
		ChildMap:   make(map[string]*Component),
		Handlers:   make(map[string]event.HandlerMap),
		DirtyPaint: true,
	}
}

func TestComponent_SetState_MarksDirty(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	comp := newTestComponent("c1", "test", 10, 5, simpleRender("hello"))
	comp.DirtyPaint = false
	mgr.Register(comp)

	mgr.SetState("c1", "count", 42)

	if !comp.DirtyPaint {
		t.Fatal("expected DirtyPaint=true after SetState")
	}
	if comp.State["count"] != 42 {
		t.Fatalf("expected state[count]=42, got %v", comp.State["count"])
	}
}

func TestComponent_Render_UpdatesBuffer(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	comp := newTestComponent("c1", "test", 20, 5, simpleRender("hello"))
	mgr.Register(comp)

	mgr.RenderDirty()

	if comp.DirtyPaint {
		t.Fatal("expected DirtyPaint=false after RenderDirty")
	}
	if comp.VNodeTree == nil {
		t.Fatal("expected VNodeTree to be set after render")
	}
	// The painter should have written something to the buffer.
	// Check that at least one cell is non-zero.
	found := false
	for y := 0; y < comp.Buffer.Height(); y++ {
		for x := 0; x < comp.Buffer.Width(); x++ {
			if !comp.Buffer.Get(x, y).Zero() {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("expected buffer to have non-zero cells after render")
	}
}

func TestComponent_RectUnchanged_NoRectChanged(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	comp := newTestComponent("c1", "test", 10, 5, simpleRender("hello"))
	comp.RectChanged = false
	mgr.Register(comp)

	// Just rendering shouldn't change RectChanged.
	mgr.RenderDirty()

	if comp.RectChanged {
		t.Fatal("expected RectChanged=false when rect didn't change")
	}
}

func TestComponent_RectChanged(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	comp := newTestComponent("c1", "test", 10, 5, simpleRender("hello"))
	mgr.Register(comp)

	// Simulate a rect change.
	comp.PrevRect = comp.Rect
	comp.Rect = buffer.Rect{X: 0, Y: 0, W: 20, H: 10}
	comp.RectChanged = true

	changed := mgr.GetRectChanged()
	if len(changed) != 1 || changed[0].ID != "c1" {
		t.Fatalf("expected 1 rect-changed component, got %d", len(changed))
	}
}

func TestComponent_GetDirtyPaint(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	c1 := newTestComponent("c1", "a", 10, 5, simpleRender("a"))
	c2 := newTestComponent("c2", "b", 10, 5, simpleRender("b"))
	c3 := newTestComponent("c3", "c", 10, 5, simpleRender("c"))

	c1.DirtyPaint = false
	c2.DirtyPaint = true
	c3.DirtyPaint = false

	mgr.Register(c1)
	mgr.Register(c2)
	mgr.Register(c3)

	dirty := mgr.GetDirtyPaint()
	if len(dirty) != 1 {
		t.Fatalf("expected 1 dirty component, got %d", len(dirty))
	}
	if dirty[0].ID != "c2" {
		t.Fatalf("expected dirty component c2, got %s", dirty[0].ID)
	}
}

func TestComponent_ParentChild(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	parent := newTestComponent("parent", "parent", 40, 20, simpleRender("parent"))
	child1 := newTestComponent("parent:child1", "child1", 20, 10, simpleRender("child1"))
	child2 := newTestComponent("parent:child2", "child2", 20, 10, simpleRender("child2"))

	child1.Parent = parent
	child2.Parent = parent
	parent.Children = []*Component{child1, child2}
	parent.ChildMap["child1"] = child1
	parent.ChildMap["child2"] = child2

	mgr.Register(parent)
	mgr.Register(child1)
	mgr.Register(child2)

	if len(parent.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(parent.Children))
	}
	if child1.Parent != parent {
		t.Fatal("child1.Parent should be parent")
	}
	if child2.Parent != parent {
		t.Fatal("child2.Parent should be parent")
	}
}

func TestComponent_AllLayers(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	c1 := newTestComponent("c1", "a", 10, 5, simpleRender("a"))
	c1.Rect = buffer.Rect{X: 5, Y: 3, W: 10, H: 5}
	c1.ZIndex = 2

	mgr.Register(c1)

	layers := mgr.AllLayers()
	if len(layers) != 1 {
		t.Fatalf("expected 1 layer, got %d", len(layers))
	}
	l := layers[0]
	if l.Layer.ID != "c1" {
		t.Fatalf("expected layer ID c1, got %s", l.Layer.ID)
	}
	if l.Layer.Rect != c1.Rect {
		t.Fatalf("expected layer rect %v, got %v", c1.Rect, l.Layer.Rect)
	}
	if l.Layer.ZIndex != 2 {
		t.Fatalf("expected ZIndex 2, got %d", l.Layer.ZIndex)
	}
}

func TestComponent_Reconcile_AddChild(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	parent := newTestComponent("p", "parent", 40, 20, simpleRender("parent"))
	mgr.Register(parent)

	descs := []ChildDescriptor{
		{Key: "a", Name: "childA", Props: map[string]any{"x": 1}, RenderFn: simpleRender("a")},
	}
	mgr.Reconcile(parent, descs)

	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children))
	}
	child := parent.Children[0]
	if child.ID != "p:a" {
		t.Fatalf("expected child ID p:a, got %s", child.ID)
	}
	if child.Parent != parent {
		t.Fatal("child.Parent should be parent")
	}
	if !child.DirtyPaint {
		t.Fatal("new child should be DirtyPaint=true")
	}
	if mgr.Get("p:a") == nil {
		t.Fatal("child should be registered in manager")
	}
}

func TestComponent_Reconcile_RemoveChild(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	parent := newTestComponent("p", "parent", 40, 20, simpleRender("parent"))
	mgr.Register(parent)

	// First reconcile: add two children.
	descs := []ChildDescriptor{
		{Key: "a", Name: "childA", Props: map[string]any{"x": 1}, RenderFn: simpleRender("a")},
		{Key: "b", Name: "childB", Props: map[string]any{"y": 2}, RenderFn: simpleRender("b")},
	}
	mgr.Reconcile(parent, descs)

	if len(parent.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(parent.Children))
	}

	// Second reconcile: remove child "a".
	descs2 := []ChildDescriptor{
		{Key: "b", Name: "childB", Props: map[string]any{"y": 2}, RenderFn: simpleRender("b")},
	}
	mgr.Reconcile(parent, descs2)

	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child after removal, got %d", len(parent.Children))
	}
	if parent.Children[0].ID != "p:b" {
		t.Fatalf("expected remaining child p:b, got %s", parent.Children[0].ID)
	}
	if mgr.Get("p:a") != nil {
		t.Fatal("removed child should be unregistered from manager")
	}
}

func TestComponent_Reconcile_UpdateChild(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	parent := newTestComponent("p", "parent", 40, 20, simpleRender("parent"))
	mgr.Register(parent)

	descs := []ChildDescriptor{
		{Key: "a", Name: "childA", Props: map[string]any{"x": 1}, RenderFn: simpleRender("a")},
	}
	mgr.Reconcile(parent, descs)

	child := parent.Children[0]
	child.DirtyPaint = false // clear dirty from creation

	// Reconcile with new props.
	descs2 := []ChildDescriptor{
		{Key: "a", Name: "childA", Props: map[string]any{"x": 99}, RenderFn: simpleRender("a")},
	}
	mgr.Reconcile(parent, descs2)

	if !child.DirtyPaint {
		t.Fatal("child should be dirty after props change")
	}
	if child.Props["x"] != 99 {
		t.Fatalf("expected props[x]=99, got %v", child.Props["x"])
	}
}

func TestComponent_Reconcile_StableKeys(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	parent := newTestComponent("p", "parent", 40, 20, simpleRender("parent"))
	mgr.Register(parent)

	descs := []ChildDescriptor{
		{Key: "a", Name: "childA", Props: map[string]any{"x": 1}, RenderFn: simpleRender("a")},
		{Key: "b", Name: "childB", Props: map[string]any{"y": 2}, RenderFn: simpleRender("b")},
	}
	mgr.Reconcile(parent, descs)

	childA := parent.ChildMap["a"]
	childB := parent.ChildMap["b"]

	// Reorder: b first, then a.
	descs2 := []ChildDescriptor{
		{Key: "b", Name: "childB", Props: map[string]any{"y": 2}, RenderFn: simpleRender("b")},
		{Key: "a", Name: "childA", Props: map[string]any{"x": 1}, RenderFn: simpleRender("a")},
	}
	mgr.Reconcile(parent, descs2)

	// Same instances should be reused.
	if parent.ChildMap["a"] != childA {
		t.Fatal("child 'a' instance should be stable across reconciliation")
	}
	if parent.ChildMap["b"] != childB {
		t.Fatal("child 'b' instance should be stable across reconciliation")
	}
	// Order should reflect new descriptors.
	if parent.Children[0] != childB {
		t.Fatal("first child should be 'b' after reorder")
	}
	if parent.Children[1] != childA {
		t.Fatal("second child should be 'a' after reorder")
	}
}

func TestComponent_ExtractHandlers(t *testing.T) {
	var clicked bool
	renderFn := func(state map[string]any, props map[string]any) *layout.VNode {
		root := layout.NewVNode("box")
		root.ID = "root"
		btn := layout.NewVNode("box")
		btn.ID = "btn1"
		btn.Props["onClick"] = event.EventHandler(func(e *event.Event) {
			clicked = true
		})
		btn.Props["focusable"] = true
		root.AddChild(btn)
		return root
	}

	comp := newTestComponent("c1", "test", 20, 10, renderFn)
	comp.VNodeTree = comp.RenderFn(comp.State, comp.Props)
	comp.ExtractHandlers()

	// Check handler extracted.
	hm, ok := comp.Handlers["btn1"]
	if !ok {
		t.Fatal("expected handlers for btn1")
	}
	handler, ok := hm["click"]
	if !ok {
		t.Fatal("expected 'click' handler for btn1")
	}
	handler(&event.Event{Type: "click"})
	if !clicked {
		t.Fatal("click handler was not called")
	}

	// Check focusable extracted.
	if len(comp.Focusables) != 1 || comp.Focusables[0] != "btn1" {
		t.Fatalf("expected focusables=[btn1], got %v", comp.Focusables)
	}
}
