package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

// hooksTestStateRC creates a fresh Lua state for react_complete tests.
func hooksTestStateRC(t *testing.T) *lua.State {
	t.Helper()
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	return L
}

// -----------------------------------------------------------------------
// useSyncExternalStore tests
// -----------------------------------------------------------------------

func TestUseSyncExternalStore_ReturnsSnapshot(t *testing.T) {
	L := hooksTestStateRC(t)
	defer L.Close()

	parent := &Component{
		ID: "ses_parent", Type: "SESParent", Name: "SESParent",
		Props: make(map[string]any), State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		_store_value = 42
		Comp = lumina.defineComponent({
			name = "SESComp",
			render = function(self)
				local snapshot = lumina.useSyncExternalStore(
					function(callback) return function() end end,
					function() return _store_value end
				)
				_got_snapshot = snapshot
				return { type = "text", content = tostring(snapshot) }
			end
		})
		_tree = { type = "box", children = { lumina.createElement(Comp, {}) } }
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_got_snapshot")
	val, ok := L.ToInteger(-1)
	L.Pop(1)
	if !ok || val != 42 {
		t.Fatalf("expected snapshot=42, got %v (ok=%v)", val, ok)
	}
}

func TestUseSyncExternalStore_SubscribeCalled(t *testing.T) {
	L := hooksTestStateRC(t)
	defer L.Close()

	parent := &Component{
		ID: "ses_sub_parent", Type: "SESSubParent", Name: "SESSubParent",
		Props: make(map[string]any), State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		_subscribe_called = false
		_unsubscribe_fn = nil
		Comp = lumina.defineComponent({
			name = "SESSubComp",
			render = function(self)
				local snapshot = lumina.useSyncExternalStore(
					function(callback)
						_subscribe_called = true
						return function() _unsubscribe_fn = true end
					end,
					function() return 99 end
				)
				return { type = "text", content = tostring(snapshot) }
			end
		})
		_tree = { type = "box", children = { lumina.createElement(Comp, {}) } }
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_subscribe_called")
	if !L.ToBoolean(-1) {
		t.Fatal("expected subscribe to be called")
	}
	L.Pop(1)
}

func TestUseSyncExternalStore_SnapshotUpdatesOnRerender(t *testing.T) {
	L := hooksTestStateRC(t)
	defer L.Close()

	parent := &Component{
		ID: "ses_update_parent", Type: "SESUpdateParent", Name: "SESUpdateParent",
		Props: make(map[string]any), State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		_store_val = 10
		Comp = lumina.defineComponent({
			name = "SESUpdateComp",
			render = function(self)
				local snapshot = lumina.useSyncExternalStore(
					function(callback) return function() end end,
					function() return _store_val end
				)
				_got = snapshot
				return { type = "text", content = tostring(snapshot) }
			end
		})
		_tree = { type = "box", children = { lumina.createElement(Comp, {}) } }
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// First render: snapshot = 10
	L.GetGlobal("_got")
	v1, _ := L.ToInteger(-1)
	L.Pop(1)
	if v1 != 10 {
		t.Fatalf("render 1: expected 10, got %d", v1)
	}

	// Change store value and re-render
	err = L.DoString(`_store_val = 20`)
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	SetCurrentComponent(parent)
	err = L.DoString(`_tree2 = { type = "box", children = { lumina.createElement(Comp, { v = 2 }) } }`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}
	L.GetGlobal("_tree2")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_got")
	v2, _ := L.ToInteger(-1)
	L.Pop(1)
	if v2 != 20 {
		t.Fatalf("render 2: expected 20, got %d", v2)
	}
}

// -----------------------------------------------------------------------
// useDebugValue tests
// -----------------------------------------------------------------------

func TestUseDebugValue_StoresLabel(t *testing.T) {
	L := hooksTestStateRC(t)
	defer L.Close()

	parent := &Component{
		ID: "dv_parent", Type: "DVParent", Name: "DVParent",
		Props: make(map[string]any), State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		Comp = lumina.defineComponent({
			name = "DVComp",
			render = function(self)
				lumina.useDebugValue("my-debug-label")
				return { type = "text", content = "debug" }
			end
		})
		_tree = { type = "box", children = { lumina.createElement(Comp, {}) } }
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Find the child component and check debug values
	if len(parent.ChildComps) == 0 {
		t.Fatal("expected child component")
	}
	child := parent.ChildComps[0]
	debugVals := child.GetDebugValues()
	if len(debugVals) != 1 || debugVals[0] != "my-debug-label" {
		t.Fatalf("expected debug value 'my-debug-label', got %v", debugVals)
	}
}

// -----------------------------------------------------------------------
// Profiler tests
// -----------------------------------------------------------------------

func TestProfiler_OnRenderCalled(t *testing.T) {
	L := hooksTestStateRC(t)
	defer L.Close()

	// Use raw profiler VNode (not through createElement) since onRender is a Lua function
	err := L.DoString(`
		_profiler_called = false
		_profiler_id = ""
		_profiler_phase = ""
		_tree = {
			type = "profiler",
			id = "test-profiler",
			onRender = function(id, phase, actualDuration, baseDuration, startTime, commitTime)
				_profiler_called = true
				_profiler_id = id
				_profiler_phase = phase
			end,
			children = {
				{ type = "text", content = "profiled content" },
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Check that onRender was called
	L.GetGlobal("_profiler_called")
	if !L.ToBoolean(-1) {
		t.Fatal("expected onRender to be called")
	}
	L.Pop(1)

	L.GetGlobal("_profiler_id")
	pid, _ := L.ToString(-1)
	L.Pop(1)
	if pid != "test-profiler" {
		t.Fatalf("expected profiler id 'test-profiler', got %q", pid)
	}

	L.GetGlobal("_profiler_phase")
	phase, _ := L.ToString(-1)
	L.Pop(1)
	if phase != "mount" {
		t.Fatalf("expected phase 'mount', got %q", phase)
	}

	// Children should be rendered
	if len(vnode.Children) == 0 {
		t.Fatal("expected profiler to render children")
	}
}

func TestProfiler_ChildrenRendered(t *testing.T) {
	L := hooksTestStateRC(t)
	defer L.Close()

	err := L.DoString(`
		_tree = {
			type = "profiler",
			id = "nav",
			onRender = function() end,
			children = {
				{ type = "text", content = "child1" },
				{ type = "text", content = "child2" },
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if len(vnode.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(vnode.Children))
	}
	if vnode.Children[0].Content != "child1" {
		t.Fatalf("expected child1, got %q", vnode.Children[0].Content)
	}
	if vnode.Children[1].Content != "child2" {
		t.Fatalf("expected child2, got %q", vnode.Children[1].Content)
	}
}

// -----------------------------------------------------------------------
// StrictMode tests
// -----------------------------------------------------------------------

func TestStrictMode_DoubleRenders(t *testing.T) {
	L := hooksTestStateRC(t)
	defer L.Close()

	parent := &Component{
		ID: "sm_parent", Type: "SMParent", Name: "SMParent",
		Props: make(map[string]any), State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		_render_count = 0
		Comp = lumina.defineComponent({
			name = "SMComp",
			render = function(self)
				_render_count = _render_count + 1
				return { type = "text", content = "strict" }
			end
		})
		_tree = lumina.createElement(lumina.StrictMode, {
			children = {
				lumina.createElement(Comp, {}),
			}
		})
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// StrictMode double-renders: expect render_count >= 2
	L.GetGlobal("_render_count")
	rc, _ := L.ToInteger(-1)
	L.Pop(1)
	if rc < 2 {
		t.Fatalf("expected render_count >= 2 (StrictMode double-render), got %d", rc)
	}

	// Children should still be rendered correctly
	if len(vnode.Children) == 0 {
		t.Fatal("expected StrictMode to render children")
	}
}

func TestStrictMode_ChildrenRenderedCorrectly(t *testing.T) {
	L := hooksTestStateRC(t)
	defer L.Close()

	err := L.DoString(`
		_tree = {
			type = "strictmode",
			children = {
				{ type = "text", content = "strict-child" },
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if len(vnode.Children) == 0 {
		t.Fatal("expected children")
	}
	if vnode.Children[0].Content != "strict-child" {
		t.Fatalf("expected 'strict-child', got %q", vnode.Children[0].Content)
	}
}

// -----------------------------------------------------------------------
// Event Capture Phase tests
// -----------------------------------------------------------------------

func TestEventCapture_CaptureBeforeBubble(t *testing.T) {
	ClearComponents()
	eb := NewEventBus()

	// Build a VNode tree: root → parent → child
	root := NewVNode("box")
	root.Props["id"] = "root"
	parentNode := NewVNode("box")
	parentNode.Props["id"] = "parent"
	childNode := NewVNode("box")
	childNode.Props["id"] = "child"
	parentNode.AddChild(childNode)
	root.AddChild(parentNode)

	tree := BuildVNodeTree(root)
	eb.SetVNodeTree(tree)

	var order []string

	// Register capture handler on root
	eb.OnCapture("click", "root", func(e *Event) {
		order = append(order, "root-capture")
	})

	// Register bubble handler on root
	eb.On("click", "root", func(e *Event) {
		order = append(order, "root-bubble")
	})

	// Register capture handler on parent
	eb.OnCapture("click", "parent", func(e *Event) {
		order = append(order, "parent-capture")
	})

	// Register bubble handler on parent
	eb.On("click", "parent", func(e *Event) {
		order = append(order, "parent-bubble")
	})

	// Register target handler on child
	eb.On("click", "child", func(e *Event) {
		order = append(order, "child-target")
	})

	// Emit click on child
	eb.Emit(&Event{Type: "click", Target: "child"})

	// Expected order: capture (root→parent), target (child), bubble (parent→root)
	expected := []string{"root-capture", "parent-capture", "child-target", "parent-bubble", "root-bubble"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d events, got %d: %v", len(expected), len(order), order)
	}
	for i, exp := range expected {
		if order[i] != exp {
			t.Fatalf("order[%d]: expected %q, got %q (full: %v)", i, exp, order[i], order)
		}
	}
}

func TestEventCapture_StopPropagationInCapture(t *testing.T) {
	ClearComponents()
	eb := NewEventBus()

	root := NewVNode("box")
	root.Props["id"] = "root"
	childNode := NewVNode("box")
	childNode.Props["id"] = "child"
	root.AddChild(childNode)

	tree := BuildVNodeTree(root)
	eb.SetVNodeTree(tree)

	var order []string

	// Register capture handler on root that stops propagation
	eb.OnCapture("click", "root", func(e *Event) {
		order = append(order, "root-capture")
		e.StopPropagation()
	})

	// Register target handler on child (should NOT fire)
	eb.On("click", "child", func(e *Event) {
		order = append(order, "child-target")
	})

	// Register bubble handler on root (should NOT fire)
	eb.On("click", "root", func(e *Event) {
		order = append(order, "root-bubble")
	})

	eb.Emit(&Event{Type: "click", Target: "child"})

	// Only root-capture should fire
	if len(order) != 1 || order[0] != "root-capture" {
		t.Fatalf("expected only ['root-capture'], got %v", order)
	}
}
