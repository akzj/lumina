package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

// ---------- helpers ----------

func compTestState(t *testing.T) *lua.State {
	t.Helper()
	L := lua.NewState()
	Open(L)
	return L
}

// ---------- createElement ----------

func TestCreateElement_ReturnsComponentTable(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Counter = lumina.defineComponent({
			name = "Counter",
			render = function(self)
				return { type = "text", content = "Count: 0" }
			end
		})
		_el = lumina.createElement(Counter, { initial = 5 })
	`)
	if err != nil {
		t.Fatalf("createElement: %v", err)
	}

	L.GetGlobal("_el")
	if L.Type(-1) != lua.TypeTable {
		t.Fatalf("expected table, got %s", L.TypeName(L.Type(-1)))
	}

	// Check type = "component"
	L.GetField(-1, "type")
	tp, _ := L.ToString(-1)
	L.Pop(1)
	if tp != "component" {
		t.Fatalf("expected type='component', got %q", tp)
	}

	// Check _factory is a table
	L.GetField(-1, "_factory")
	if L.Type(-1) != lua.TypeTable {
		t.Fatalf("expected _factory to be table, got %s", L.TypeName(L.Type(-1)))
	}
	L.Pop(1)

	// Check _props.initial = 5
	L.GetField(-1, "_props")
	if L.Type(-1) != lua.TypeTable {
		t.Fatalf("expected _props to be table, got %s", L.TypeName(L.Type(-1)))
	}
	L.GetField(-1, "initial")
	v, ok := L.ToInteger(-1)
	L.Pop(2) // pop initial + _props
	if !ok || v != 5 {
		t.Fatalf("expected _props.initial=5, got %v (ok=%v)", v, ok)
	}
	L.Pop(1) // pop _el
}

func TestCreateElement_DefaultProps(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	err := L.DoString(`
		local C = lumina.defineComponent({
			name = "C",
			render = function(self)
				return { type = "text", content = "hi" }
			end
		})
		_el = lumina.createElement(C)
	`)
	if err != nil {
		t.Fatalf("createElement no props: %v", err)
	}

	L.GetGlobal("_el")
	L.GetField(-1, "_props")
	if L.Type(-1) != lua.TypeTable {
		t.Fatalf("expected _props table even without args, got %s", L.TypeName(L.Type(-1)))
	}
	L.Pop(2)
}

// ---------- VNode struct fields ----------

func TestVNode_ComponentRefAndKey(t *testing.T) {
	vnode := NewVNode("box")
	if vnode.ComponentRef != nil {
		t.Fatal("ComponentRef should be nil by default")
	}
	if vnode.ComponentKey != "" {
		t.Fatal("ComponentKey should be empty by default")
	}

	comp := &Component{ID: "test", Type: "Test", Name: "Test"}
	vnode.ComponentRef = comp
	vnode.ComponentKey = "my-key"

	if vnode.ComponentRef.ID != "test" {
		t.Fatalf("expected ComponentRef.ID='test', got %q", vnode.ComponentRef.ID)
	}
	if vnode.ComponentKey != "my-key" {
		t.Fatalf("expected ComponentKey='my-key', got %q", vnode.ComponentKey)
	}
}

// ---------- Component tree ----------

func TestComponent_AddRemoveChild(t *testing.T) {
	parent := &Component{
		ID:    "parent",
		Type:  "Parent",
		Name:  "Parent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	child1 := &Component{
		ID:    "child1",
		Type:  "Child",
		Name:  "Child1",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	child2 := &Component{
		ID:    "child2",
		Type:  "Child",
		Name:  "Child2",
		Props: make(map[string]any),
		State: make(map[string]any),
	}

	parent.AddChild(child1)
	parent.AddChild(child2)

	children := parent.GetChildren()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if child1.Parent != parent {
		t.Fatal("child1.Parent should be parent")
	}
	if child2.Parent != parent {
		t.Fatal("child2.Parent should be parent")
	}

	parent.RemoveChild(child1)
	children = parent.GetChildren()
	if len(children) != 1 {
		t.Fatalf("expected 1 child after remove, got %d", len(children))
	}
	if children[0].ID != "child2" {
		t.Fatalf("expected remaining child to be child2, got %s", children[0].ID)
	}
	if child1.Parent != nil {
		t.Fatal("child1.Parent should be nil after remove")
	}
}

func TestComponent_UpdateProps(t *testing.T) {
	comp := &Component{
		ID:    "test",
		Type:  "Test",
		Name:  "Test",
		Props: map[string]any{"color": "red", "size": int64(10)},
		State: make(map[string]any),
	}

	// Same props → no change
	changed := comp.UpdateProps(map[string]any{"color": "red", "size": int64(10)})
	if changed {
		t.Fatal("expected no change for identical props")
	}
	if comp.Dirty.Load() {
		t.Fatal("should not be dirty for identical props")
	}

	// Different props → change
	changed = comp.UpdateProps(map[string]any{"color": "blue", "size": int64(10)})
	if !changed {
		t.Fatal("expected change for different props")
	}
	if !comp.Dirty.Load() {
		t.Fatal("should be dirty after props change")
	}
	if comp.Props["color"] != "blue" {
		t.Fatalf("expected color='blue', got %v", comp.Props["color"])
	}
}
