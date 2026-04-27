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
// Within the component package, we can access private fields directly.
func newTestComponent(id, name string, w, h int, renderFn RenderFunc) *Component {
	return &Component{
		id:         id,
		name:       name,
		buf:        buffer.New(w, h),
		rect:       buffer.Rect{X: 0, Y: 0, W: w, H: h},
		state:      make(map[string]any),
		props:      make(map[string]any),
		renderFn:   renderFn,
		childMap:   make(map[string]*Component),
		handlers:   make(map[string]HandlerMap),
		dirtyPaint: true,
		hookStore:  make(map[string]any),
	}
}

func TestComponent_SetState_MarksDirty(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	comp := newTestComponent("c1", "test", 10, 5, simpleRender("hello"))
	comp.dirtyPaint = false
	mgr.Register(comp)

	mgr.SetState("c1", "count", 42)

	if !comp.dirtyPaint {
		t.Fatal("expected DirtyPaint=true after SetState")
	}
	if comp.state["count"] != 42 {
		t.Fatalf("expected state[count]=42, got %v", comp.state["count"])
	}
}

func TestComponent_Render_UpdatesBuffer(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	comp := newTestComponent("c1", "test", 20, 5, simpleRender("hello"))
	mgr.Register(comp)

	mgr.RenderDirty()

	if comp.dirtyPaint {
		t.Fatal("expected DirtyPaint=false after RenderDirty")
	}
	if comp.vnodeTree == nil {
		t.Fatal("expected VNodeTree to be set after render")
	}
	// The painter should have written something to the buffer.
	// Check that at least one cell is non-zero.
	found := false
	for y := 0; y < comp.buf.Height(); y++ {
		for x := 0; x < comp.buf.Width(); x++ {
			if !comp.buf.Get(x, y).Zero() {
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
	comp.rectChanged = false
	mgr.Register(comp)

	// Just rendering shouldn't change RectChanged.
	mgr.RenderDirty()

	if comp.rectChanged {
		t.Fatal("expected RectChanged=false when rect didn't change")
	}
}

func TestComponent_RectChanged(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	comp := newTestComponent("c1", "test", 10, 5, simpleRender("hello"))
	mgr.Register(comp)

	// Simulate a rect change.
	comp.prevRect = comp.rect
	comp.rect = buffer.Rect{X: 0, Y: 0, W: 20, H: 10}
	comp.rectChanged = true

	changed := mgr.GetRectChanged()
	if len(changed) != 1 || changed[0].ID() != "c1" {
		t.Fatalf("expected 1 rect-changed component, got %d", len(changed))
	}
}

func TestComponent_GetDirtyPaint(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	c1 := newTestComponent("c1", "a", 10, 5, simpleRender("a"))
	c2 := newTestComponent("c2", "b", 10, 5, simpleRender("b"))
	c3 := newTestComponent("c3", "c", 10, 5, simpleRender("c"))

	c1.dirtyPaint = false
	c2.dirtyPaint = true
	c3.dirtyPaint = false

	mgr.Register(c1)
	mgr.Register(c2)
	mgr.Register(c3)

	dirty := mgr.GetDirtyPaint()
	if len(dirty) != 1 {
		t.Fatalf("expected 1 dirty component, got %d", len(dirty))
	}
	if dirty[0].ID() != "c2" {
		t.Fatalf("expected dirty component c2, got %s", dirty[0].ID())
	}
}

func TestComponent_ParentChild(t *testing.T) {
	mgr := NewManager(paint.NewPainter())
	parent := newTestComponent("parent", "parent", 40, 20, simpleRender("parent"))
	child1 := newTestComponent("parent:child1", "child1", 20, 10, simpleRender("child1"))
	child2 := newTestComponent("parent:child2", "child2", 20, 10, simpleRender("child2"))

	child1.parent = parent
	child2.parent = parent
	parent.children = []*Component{child1, child2}
	parent.childMap["child1"] = child1
	parent.childMap["child2"] = child2

	mgr.Register(parent)
	mgr.Register(child1)
	mgr.Register(child2)

	if len(parent.children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(parent.children))
	}
	if child1.parent != parent {
		t.Fatal("child1.Parent should be parent")
	}
	if child2.parent != parent {
		t.Fatal("child2.Parent should be parent")
	}
}

func TestComponent_AllLayers_Removed(t *testing.T) {
	// AllLayers() has been removed from Manager.
	// This test verifies that the accessor methods work correctly
	// so app.go can build layers itself.
	mgr := NewManager(paint.NewPainter())
	c1 := newTestComponent("c1", "a", 10, 5, simpleRender("a"))
	c1.rect = buffer.Rect{X: 5, Y: 3, W: 10, H: 5}
	c1.zIndex = 2

	mgr.Register(c1)

	all := mgr.GetAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 component, got %d", len(all))
	}
	comp := all[0]
	if comp.ID() != "c1" {
		t.Fatalf("expected ID c1, got %s", comp.ID())
	}
	if comp.Rect() != c1.rect {
		t.Fatalf("expected rect %v, got %v", c1.rect, comp.Rect())
	}
	if comp.ZIndex() != 2 {
		t.Fatalf("expected ZIndex 2, got %d", comp.ZIndex())
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

	if len(parent.children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.children))
	}
	child := parent.children[0]
	if child.id != "p:a" {
		t.Fatalf("expected child ID p:a, got %s", child.id)
	}
	if child.parent != parent {
		t.Fatal("child.Parent should be parent")
	}
	if !child.dirtyPaint {
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

	if len(parent.children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(parent.children))
	}

	// Second reconcile: remove child "a".
	descs2 := []ChildDescriptor{
		{Key: "b", Name: "childB", Props: map[string]any{"y": 2}, RenderFn: simpleRender("b")},
	}
	mgr.Reconcile(parent, descs2)

	if len(parent.children) != 1 {
		t.Fatalf("expected 1 child after removal, got %d", len(parent.children))
	}
	if parent.children[0].id != "p:b" {
		t.Fatalf("expected remaining child p:b, got %s", parent.children[0].id)
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

	child := parent.children[0]
	child.dirtyPaint = false // clear dirty from creation

	// Reconcile with new props.
	descs2 := []ChildDescriptor{
		{Key: "a", Name: "childA", Props: map[string]any{"x": 99}, RenderFn: simpleRender("a")},
	}
	mgr.Reconcile(parent, descs2)

	if !child.dirtyPaint {
		t.Fatal("child should be dirty after props change")
	}
	if child.props["x"] != 99 {
		t.Fatalf("expected props[x]=99, got %v", child.props["x"])
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

	childA := parent.childMap["a"]
	childB := parent.childMap["b"]

	// Reorder: b first, then a.
	descs2 := []ChildDescriptor{
		{Key: "b", Name: "childB", Props: map[string]any{"y": 2}, RenderFn: simpleRender("b")},
		{Key: "a", Name: "childA", Props: map[string]any{"x": 1}, RenderFn: simpleRender("a")},
	}
	mgr.Reconcile(parent, descs2)

	// Same instances should be reused.
	if parent.childMap["a"] != childA {
		t.Fatal("child 'a' instance should be stable across reconciliation")
	}
	if parent.childMap["b"] != childB {
		t.Fatal("child 'b' instance should be stable across reconciliation")
	}
	// Order should reflect new descriptors.
	if parent.children[0] != childB {
		t.Fatal("first child should be 'b' after reorder")
	}
	if parent.children[1] != childA {
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
	comp.vnodeTree = comp.renderFn(comp.state, comp.props)
	comp.ExtractHandlers()

	// Check handler extracted.
	hm, ok := comp.handlers["btn1"]
	if !ok {
		t.Fatal("expected handlers for btn1")
	}
	handler, ok := hm["click"]
	if !ok {
		t.Fatal("expected 'click' handler for btn1")
	}
	// Type-assert back to event.EventHandler (same as app.go does).
	if h, ok := handler.(event.EventHandler); ok {
		h(&event.Event{Type: "click"})
	} else {
		t.Fatal("handler should be type-assertable to event.EventHandler")
	}
	if !clicked {
		t.Fatal("click handler was not called")
	}

	// Check focusable extracted.
	if len(comp.focusables) != 1 || comp.focusables[0] != "btn1" {
		t.Fatalf("expected focusables=[btn1], got %v", comp.focusables)
	}
}
