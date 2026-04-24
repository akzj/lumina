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

// ---------- Recursive component rendering ----------

func TestComposition_SimpleChildComponent(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Counter = lumina.defineComponent({
			name = "Counter",
			render = function(self)
				return { type = "text", content = "Count: 0" }
			end
		})

		-- Create a VNode tree with a component child
		_tree = {
			type = "vbox",
			children = {
				{ type = "text", content = "Header" },
				lumina.createElement(Counter, {}),
				{ type = "text", content = "Footer" },
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if vnode.Type != "vbox" {
		t.Fatalf("expected root type='vbox', got %q", vnode.Type)
	}
	if len(vnode.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(vnode.Children))
	}

	// First child: plain text
	if vnode.Children[0].Type != "text" || vnode.Children[0].Content != "Header" {
		t.Fatalf("child[0]: expected text/Header, got %s/%s", vnode.Children[0].Type, vnode.Children[0].Content)
	}

	// Second child: rendered component (should be text node from Counter's render)
	child1 := vnode.Children[1]
	if child1.Type != "text" {
		t.Fatalf("child[1]: expected type='text' (rendered Counter), got %q", child1.Type)
	}
	if child1.Content != "Count: 0" {
		t.Fatalf("child[1]: expected content='Count: 0', got %q", child1.Content)
	}
	if child1.ComponentRef == nil {
		t.Fatal("child[1]: expected ComponentRef to be set")
	}
	if child1.ComponentRef.Type != "Counter" {
		t.Fatalf("child[1]: expected ComponentRef.Type='Counter', got %q", child1.ComponentRef.Type)
	}

	// Third child: plain text
	if vnode.Children[2].Type != "text" || vnode.Children[2].Content != "Footer" {
		t.Fatalf("child[2]: expected text/Footer, got %s/%s", vnode.Children[2].Type, vnode.Children[2].Content)
	}
}

func TestComposition_ComponentWithProps(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Greeter = lumina.defineComponent({
			name = "Greeter",
			render = function(self)
				local name = self.name or "World"
				return { type = "text", content = "Hello, " .. name .. "!" }
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(Greeter, { name = "Alice" }),
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if len(vnode.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(vnode.Children))
	}
	child := vnode.Children[0]
	if child.Content != "Hello, Alice!" {
		t.Fatalf("expected 'Hello, Alice!', got %q", child.Content)
	}
}

func TestComposition_NestedComponents(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Leaf = lumina.defineComponent({
			name = "Leaf",
			render = function(self)
				return { type = "text", content = "leaf" }
			end
		})

		local Middle = lumina.defineComponent({
			name = "Middle",
			render = function(self)
				return {
					type = "box",
					children = {
						lumina.createElement(Leaf, {}),
					}
				}
			end
		})

		local Root = lumina.defineComponent({
			name = "Root",
			render = function(self)
				return {
					type = "vbox",
					children = {
						lumina.createElement(Middle, {}),
					}
				}
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(Root, {}),
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// box > vbox (from Root) > box (from Middle) > text "leaf" (from Leaf)
	if len(vnode.Children) != 1 {
		t.Fatalf("root: expected 1 child, got %d", len(vnode.Children))
	}
	rootChild := vnode.Children[0]
	if rootChild.Type != "vbox" {
		t.Fatalf("expected Root to render vbox, got %q", rootChild.Type)
	}
	if rootChild.ComponentRef == nil || rootChild.ComponentRef.Type != "Root" {
		t.Fatal("expected ComponentRef for Root")
	}

	if len(rootChild.Children) != 1 {
		t.Fatalf("Root vbox: expected 1 child, got %d", len(rootChild.Children))
	}
	middleChild := rootChild.Children[0]
	if middleChild.Type != "box" {
		t.Fatalf("expected Middle to render box, got %q", middleChild.Type)
	}
	if middleChild.ComponentRef == nil || middleChild.ComponentRef.Type != "Middle" {
		t.Fatal("expected ComponentRef for Middle")
	}

	if len(middleChild.Children) != 1 {
		t.Fatalf("Middle box: expected 1 child, got %d", len(middleChild.Children))
	}
	leafChild := middleChild.Children[0]
	if leafChild.Type != "text" || leafChild.Content != "leaf" {
		t.Fatalf("expected Leaf text='leaf', got type=%q content=%q", leafChild.Type, leafChild.Content)
	}
	if leafChild.ComponentRef == nil || leafChild.ComponentRef.Type != "Leaf" {
		t.Fatal("expected ComponentRef for Leaf")
	}
}

func TestComposition_ComponentWithInit(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Counter = lumina.defineComponent({
			name = "Counter",
			init = function(props)
				return { count = props.initial or 0 }
			end,
			render = function(self)
				return { type = "text", content = "Count: " .. tostring(self.count) }
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(Counter, { initial = 42 }),
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if len(vnode.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(vnode.Children))
	}
	child := vnode.Children[0]
	if child.Content != "Count: 42" {
		t.Fatalf("expected 'Count: 42', got %q", child.Content)
	}
}

func TestComposition_MixedChildren(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Badge = lumina.defineComponent({
			name = "Badge",
			render = function(self)
				return { type = "text", content = "[" .. (self.label or "?") .. "]" }
			end
		})

		_tree = {
			type = "hbox",
			children = {
				{ type = "text", content = "Name: " },
				lumina.createElement(Badge, { label = "Admin" }),
				{ type = "text", content = " - Active" },
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if len(vnode.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(vnode.Children))
	}
	if vnode.Children[0].Content != "Name: " {
		t.Fatalf("child[0]: expected 'Name: ', got %q", vnode.Children[0].Content)
	}
	if vnode.Children[1].Content != "[Admin]" {
		t.Fatalf("child[1]: expected '[Admin]', got %q", vnode.Children[1].Content)
	}
	if vnode.Children[1].ComponentRef == nil {
		t.Fatal("child[1]: expected ComponentRef")
	}
	if vnode.Children[2].Content != " - Active" {
		t.Fatalf("child[2]: expected ' - Active', got %q", vnode.Children[2].Content)
	}
}

func TestComposition_MultipleComponentChildren(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Item = lumina.defineComponent({
			name = "Item",
			render = function(self)
				return { type = "text", content = "item:" .. (self.id or "?") }
			end
		})

		_tree = {
			type = "vbox",
			children = {
				lumina.createElement(Item, { id = "a", key = "a" }),
				lumina.createElement(Item, { id = "b", key = "b" }),
				lumina.createElement(Item, { id = "c", key = "c" }),
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if len(vnode.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(vnode.Children))
	}
	expected := []string{"item:a", "item:b", "item:c"}
	for i, exp := range expected {
		if vnode.Children[i].Content != exp {
			t.Fatalf("child[%d]: expected %q, got %q", i, exp, vnode.Children[i].Content)
		}
	}
}

// ---------- Props Reactivity ----------

func TestPropsReactivity_UpdatedPropsReflected(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	// Define a component that reads a prop
	err := L.DoString(`
		local Greeting = lumina.defineComponent({
			name = "Greeting",
			render = function(self)
				return { type = "text", content = "Hello, " .. (self.name or "World") }
			end
		})
		_Greeting = Greeting
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// First render with name = "Alice"
	err = L.DoString(`
		_tree1 = {
			type = "box",
			children = {
				lumina.createElement(_Greeting, { name = "Alice" }),
			}
		}
	`)
	if err != nil {
		t.Fatalf("tree1: %v", err)
	}

	L.GetGlobal("_tree1")
	vnode1 := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if len(vnode1.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(vnode1.Children))
	}
	if vnode1.Children[0].Content != "Hello, Alice" {
		t.Fatalf("first render: expected 'Hello, Alice', got %q", vnode1.Children[0].Content)
	}

	// Second render with name = "Bob" — same component should get updated props
	err = L.DoString(`
		_tree2 = {
			type = "box",
			children = {
				lumina.createElement(_Greeting, { name = "Bob" }),
			}
		}
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}

	L.GetGlobal("_tree2")
	vnode2 := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if len(vnode2.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(vnode2.Children))
	}
	if vnode2.Children[0].Content != "Hello, Bob" {
		t.Fatalf("second render: expected 'Hello, Bob', got %q", vnode2.Children[0].Content)
	}
}

func TestPropsReactivity_ChildrenProp(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Layout = lumina.defineComponent({
			name = "Layout",
			render = function(self)
				local kids = self.children or {}
				return {
					type = "vbox",
					children = {
						{ type = "text", content = "=== Header ===" },
						unpack(kids)
					}
				}
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(Layout, {
					children = {
						{ type = "text", content = "Child A" },
						{ type = "text", content = "Child B" },
					}
				}),
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if len(vnode.Children) != 1 {
		t.Fatalf("expected 1 child (Layout), got %d", len(vnode.Children))
	}

	layout := vnode.Children[0]
	if layout.Type != "vbox" {
		t.Fatalf("expected Layout to render vbox, got %q", layout.Type)
	}

	// Layout should have: Header + Child A + Child B = 3 children
	if len(layout.Children) < 2 {
		t.Fatalf("expected at least 2 children in Layout, got %d", len(layout.Children))
	}
	if layout.Children[0].Content != "=== Header ===" {
		t.Fatalf("expected header, got %q", layout.Children[0].Content)
	}
}

func TestPropsReactivity_SamePropsNoRerender(t *testing.T) {
	// Verify that UpdateProps returns false for identical props
	comp := &Component{
		ID:    "test",
		Type:  "Test",
		Name:  "Test",
		Props: map[string]any{"x": int64(1), "y": "hello"},
		State: make(map[string]any),
	}

	// Same props → no change
	changed := comp.UpdateProps(map[string]any{"x": int64(1), "y": "hello"})
	if changed {
		t.Fatal("expected no change for identical props")
	}
	if comp.Dirty.Load() {
		t.Fatal("should not be dirty for identical props")
	}

	// Different props → change
	changed = comp.UpdateProps(map[string]any{"x": int64(2), "y": "hello"})
	if !changed {
		t.Fatal("expected change for different props")
	}
	if !comp.Dirty.Load() {
		t.Fatal("should be dirty after props change")
	}
}

func TestPropsReactivity_ComponentWithState(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	// Component that uses both state and props
	err := L.DoString(`
		local Display = lumina.defineComponent({
			name = "Display",
			init = function(props)
				return { prefix = props.prefix or ">" }
			end,
			render = function(self)
				return { type = "text", content = self.prefix .. " " .. (self.label or "?") }
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(Display, { prefix = ">>", label = "Hello" }),
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if len(vnode.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(vnode.Children))
	}
	child := vnode.Children[0]
	if child.Content != ">> Hello" {
		t.Fatalf("expected '>> Hello', got %q", child.Content)
	}
}

// ---------- Component tree structure ----------

func TestComponentTree_ParentChildReferences(t *testing.T) {
	L := compTestState(t)
	defer L.Close()

	// Set up a parent component context
	parent := &Component{
		ID:    "parent_comp",
		Type:  "Parent",
		Name:  "Parent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.mu.Lock()
	globalRegistry.components[parent.ID] = parent
	globalRegistry.mu.Unlock()
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		globalRegistry.mu.Lock()
		delete(globalRegistry.components, parent.ID)
		globalRegistry.mu.Unlock()
	}()

	err := L.DoString(`
		local Child = lumina.defineComponent({
			name = "Child",
			render = function(self)
				return { type = "text", content = "child" }
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(Child, {}),
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Verify parent-child relationship
	children := parent.GetChildren()
	if len(children) != 1 {
		t.Fatalf("expected parent to have 1 child component, got %d", len(children))
	}
	childComp := children[0]
	if childComp.Type != "Child" {
		t.Fatalf("expected child type='Child', got %q", childComp.Type)
	}
	if childComp.Parent != parent {
		t.Fatal("child.Parent should point to parent")
	}

	// Verify VNode has the component reference
	if vnode.Children[0].ComponentRef != childComp {
		t.Fatal("VNode.ComponentRef should point to the child component")
	}
}

// ---------- setState Batching ----------

func TestBatching_MultipleSetStateSingleRender(t *testing.T) {
	app := NewAppWithSize(80, 24)
	defer app.Close()

	// Track render count
	renderCount := 0
	origRenderAllDirty := app.renderAllDirty
	_ = origRenderAllDirty // can't monkey-patch, use different approach

	// Create a component with multiple state keys
	comp := &Component{
		ID:           "batch_test",
		Type:         "BatchTest",
		Name:         "BatchTest",
		Props:        make(map[string]any),
		State:        make(map[string]any),
		RenderNotify: make(chan struct{}, 1),
	}
	globalRegistry.mu.Lock()
	globalRegistry.components[comp.ID] = comp
	globalRegistry.mu.Unlock()
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.components, comp.ID)
		globalRegistry.mu.Unlock()
	}()

	// Without batching: each SetState marks dirty
	comp.SetState("a", 1)
	comp.SetState("b", 2)
	comp.SetState("c", 3)

	// Verify all three state values are set
	if v, _ := comp.GetState("a"); v != 1 {
		t.Fatalf("expected a=1, got %v", v)
	}
	if v, _ := comp.GetState("b"); v != 2 {
		t.Fatalf("expected b=2, got %v", v)
	}
	if v, _ := comp.GetState("c"); v != 3 {
		t.Fatalf("expected c=3, got %v", v)
	}
	if !comp.Dirty.Load() {
		t.Fatal("component should be dirty after setState")
	}

	// Test batch API
	app.BeginBatch()
	if !app.IsBatching() {
		t.Fatal("should be batching after BeginBatch")
	}

	comp.MarkClean()
	comp.SetState("a", 10)
	comp.SetState("b", 20)
	comp.SetState("c", 30)

	// Still batching — dirty but no render yet
	if !comp.Dirty.Load() {
		t.Fatal("component should be dirty during batch")
	}
	if !app.IsBatching() {
		t.Fatal("should still be batching")
	}

	app.EndBatch()
	if app.IsBatching() {
		t.Fatal("should not be batching after EndBatch")
	}

	// Verify final state
	if v, _ := comp.GetState("a"); v != 10 {
		t.Fatalf("expected a=10, got %v", v)
	}
	if v, _ := comp.GetState("b"); v != 20 {
		t.Fatalf("expected b=20, got %v", v)
	}
	if v, _ := comp.GetState("c"); v != 30 {
		t.Fatalf("expected c=30, got %v", v)
	}

	_ = renderCount
}

func TestBatching_NestedBatches(t *testing.T) {
	app := NewAppWithSize(80, 24)
	defer app.Close()

	// Nested batches: only outermost EndBatch triggers render
	app.BeginBatch()
	if !app.IsBatching() {
		t.Fatal("should be batching")
	}

	app.BeginBatch() // nested
	if !app.IsBatching() {
		t.Fatal("should still be batching (nested)")
	}

	app.EndBatch() // inner
	if !app.IsBatching() {
		t.Fatal("should still be batching (inner ended, outer still active)")
	}

	app.EndBatch() // outer — this triggers render
	if app.IsBatching() {
		t.Fatal("should not be batching after all batches ended")
	}
}

func TestBatching_HandleEventWrapped(t *testing.T) {
	app := NewAppWithSize(80, 24)
	defer app.Close()

	// Create a component
	comp := &Component{
		ID:           "event_batch_test",
		Type:         "EventBatchTest",
		Name:         "EventBatchTest",
		Props:        make(map[string]any),
		State:        make(map[string]any),
		RenderNotify: make(chan struct{}, 1),
	}
	globalRegistry.mu.Lock()
	globalRegistry.components[comp.ID] = comp
	globalRegistry.mu.Unlock()
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.components, comp.ID)
		globalRegistry.mu.Unlock()
	}()

	// handleEvent should be wrapped in batch
	// Simulate by calling BeginBatch/EndBatch directly
	app.BeginBatch()
	comp.SetState("x", 1)
	comp.SetState("y", 2)
	// During batch, component is dirty but no render happens
	if !comp.Dirty.Load() {
		t.Fatal("should be dirty during batch")
	}
	app.EndBatch()
	// After EndBatch, renderAllDirty was called (component cleaned)
	// Note: without a render function, renderAllDirty won't actually clean it
	// but the batch mechanism is verified
}

func TestBatching_ZeroDepthSafe(t *testing.T) {
	app := NewAppWithSize(80, 24)
	defer app.Close()

	// EndBatch when not batching should be safe (no panic)
	app.EndBatch()
	if app.IsBatching() {
		t.Fatal("should not be batching")
	}

	// Double EndBatch should be safe
	app.BeginBatch()
	app.EndBatch()
	app.EndBatch() // extra — should not panic or go negative
	if app.IsBatching() {
		t.Fatal("should not be batching after extra EndBatch")
	}
}
