package render

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func newTestEngine(t *testing.T) (*Engine, *lua.State) {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	e := NewEngine(L, 80, 24)
	e.RegisterLuaAPI()
	return e, L
}

func TestEngine_SimpleRender(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "test",
			name = "Test",
			render = function(props)
				return lumina.createElement("text", {
					id = "t1",
					style = {foreground = "#FFF"},
				}, "Hello")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	comp := e.GetComponent("test")
	if comp == nil {
		t.Fatal("component not registered")
	}
	if comp.RootNode == nil {
		t.Fatal("RootNode nil after render")
	}
	if comp.RootNode.Content != "Hello" {
		t.Errorf("expected 'Hello', got %q", comp.RootNode.Content)
	}
	if comp.RootNode.Style.Foreground != "#FFF" {
		t.Errorf("expected foreground '#FFF', got %q", comp.RootNode.Style.Foreground)
	}
}

func TestEngine_Reconcile_ContentChange(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "counter",
			name = "Counter",
			render = function(props)
				local count, setCount = lumina.useState("count", 0)
				return lumina.createElement("text", {id="val"}, tostring(count))
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	comp := e.GetComponent("counter")
	if comp == nil {
		t.Fatal("component not registered")
	}
	if comp.RootNode == nil {
		t.Fatal("RootNode nil after render")
	}
	if comp.RootNode.Content != "0" {
		t.Fatalf("initial render: expected '0', got %q", comp.RootNode.Content)
	}

	// Change state → re-render → reconcile (update in place)
	oldNode := comp.RootNode
	e.SetState("counter", "count", int64(42))
	e.RenderDirty()

	if comp.RootNode != oldNode {
		t.Error("RootNode pointer changed — should be reconciled in-place")
	}
	if comp.RootNode.Content != "42" {
		t.Errorf("after setState: expected '42', got %q", comp.RootNode.Content)
	}
}

func TestEngine_NoRerenderWhenClean(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		render_count = 0
		lumina.createComponent({
			id = "static",
			name = "Static",
			render = function(props)
				render_count = render_count + 1
				return lumina.createElement("text", {}, "static")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Check render was called once
	L.GetGlobal("render_count")
	count := L.ToAny(-1)
	L.Pop(1)
	if count != int64(1) {
		t.Fatalf("expected render_count=1 after RenderAll, got %v", count)
	}

	// RenderDirty should NOT call Lua (component is clean)
	e.RenderDirty()

	L.GetGlobal("render_count")
	count = L.ToAny(-1)
	L.Pop(1)
	if count != int64(1) {
		t.Errorf("expected render_count=1 after RenderDirty (clean), got %v", count)
	}
}

func TestEngine_UseState_SetterMarksDirty(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "stateful",
			name = "Stateful",
			render = function(props)
				local val, setVal = lumina.useState("x", 10)
				-- Store setter globally so we can call it from Go
				global_setter = setVal
				return lumina.createElement("text", {}, tostring(val))
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	comp := e.GetComponent("stateful")
	if comp.RootNode.Content != "10" {
		t.Fatalf("initial: expected '10', got %q", comp.RootNode.Content)
	}

	// Call setter from Lua
	err = L.DoString(`global_setter(99)`)
	if err != nil {
		t.Fatal(err)
	}

	if !comp.Dirty {
		t.Error("component should be dirty after setState")
	}

	e.RenderDirty()
	if comp.RootNode.Content != "99" {
		t.Errorf("after setter: expected '99', got %q", comp.RootNode.Content)
	}
}

func TestEngine_CreateElement_WithChildren(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "tree",
			name = "Tree",
			render = function(props)
				return lumina.createElement("vbox", {id="root"},
					lumina.createElement("text", {id="t1"}, "AAA"),
					lumina.createElement("text", {id="t2"}, "BBB")
				)
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	comp := e.GetComponent("tree")
	if comp.RootNode == nil {
		t.Fatal("RootNode nil")
	}
	if comp.RootNode.Type != "vbox" {
		t.Errorf("root type = %q, want 'vbox'", comp.RootNode.Type)
	}
	if len(comp.RootNode.Children) != 2 {
		t.Fatalf("children count = %d, want 2", len(comp.RootNode.Children))
	}
	if comp.RootNode.Children[0].Content != "AAA" {
		t.Errorf("child 0 content = %q, want 'AAA'", comp.RootNode.Children[0].Content)
	}
	if comp.RootNode.Children[1].Content != "BBB" {
		t.Errorf("child 1 content = %q, want 'BBB'", comp.RootNode.Children[1].Content)
	}
}

func TestEngine_CreateElement_StyleProps(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "styled",
			name = "Styled",
			render = function(props)
				return lumina.createElement("box", {
					id = "root",
					style = {
						width = 40,
						height = 10,
						background = "#333",
						border = "single",
						padding = 1,
					},
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	comp := e.GetComponent("styled")
	s := comp.RootNode.Style

	if s.Width != 40 {
		t.Errorf("Width = %d, want 40", s.Width)
	}
	if s.Height != 10 {
		t.Errorf("Height = %d, want 10", s.Height)
	}
	if s.Background != "#333" {
		t.Errorf("Background = %q, want '#333'", s.Background)
	}
	if s.Border != "single" {
		t.Errorf("Border = %q, want 'single'", s.Border)
	}
	if s.Padding != 1 {
		t.Errorf("Padding = %d, want 1", s.Padding)
	}
}

func TestEngine_EventHandlerRef(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "evts",
			name = "Events",
			render = function(props)
				return lumina.createElement("box", {
					id = "btn",
					onClick = function(ev) end,
					onMouseEnter = function(ev) end,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	comp := e.GetComponent("evts")
	node := comp.RootNode

	if node.OnClick == 0 {
		t.Error("OnClick should be a non-zero Lua ref")
	}
	if node.OnMouseEnter == 0 {
		t.Error("OnMouseEnter should be a non-zero Lua ref")
	}
}

func TestEngine_DefineComponent(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		Cell = lumina.defineComponent("Cell", function(props)
			return lumina.createElement("text", {}, props.label or "?")
		end)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Verify factory was registered
	if _, ok := e.factories["Cell"]; !ok {
		t.Error("Cell factory not registered")
	}

	// Verify returned table has _isFactory
	L.GetGlobal("Cell")
	if !L.IsTable(-1) {
		t.Fatal("defineComponent should return a table")
	}
	L.GetField(-1, "_isFactory")
	if !L.ToBoolean(-1) {
		t.Error("returned table should have _isFactory=true")
	}
	L.Pop(2)
}

func TestEngine_Resize(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "resize",
			name = "Resize",
			render = function(props)
				return lumina.createElement("box", {id = "root"})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Resize
	e.Resize(120, 40)

	if e.buffer.Width() != 120 || e.buffer.Height() != 40 {
		t.Errorf("buffer size = %dx%d, want 120x40", e.buffer.Width(), e.buffer.Height())
	}

	comp := e.GetComponent("resize")
	if comp.RootNode != nil && !comp.RootNode.LayoutDirty {
		t.Error("root node should be layout dirty after resize")
	}
}

func TestEngine_LayoutAndPaint(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "layout",
			name = "Layout",
			render = function(props)
				return lumina.createElement("vbox", {
					id = "root",
					style = { background = "#111" },
				},
					lumina.createElement("text", {id="t1"}, "Row1"),
					lumina.createElement("text", {id="t2"}, "Row2")
				)
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Verify layout was computed
	comp := e.GetComponent("layout")
	root := comp.RootNode
	if root.W != 80 || root.H != 24 {
		t.Errorf("root size = %dx%d, want 80x24", root.W, root.H)
	}

	// Verify buffer has background painted.
	// Cell (0,0) is covered by the text "Row1" (which has no BG), so check a cell
	// in the background area below the text children.
	cell := e.buffer.Get(5, 5)
	if cell.BG != "#111" {
		t.Errorf("cell(5,5).BG = %q, want '#111'", cell.BG)
	}
	// Also verify text was painted at (0,0)
	cell00 := e.buffer.Get(0, 0)
	if cell00.Ch != 'R' {
		t.Errorf("cell(0,0).Ch = %q, want 'R' (from 'Row1')", string(cell00.Ch))
	}
}

func TestEngine_Reconcile_ChildAddRemove(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		show_extra = false
		lumina.createComponent({
			id = "dynamic",
			name = "Dynamic",
			render = function(props)
				local children = {
					lumina.createElement("text", {key="a"}, "A"),
				}
				if show_extra then
					children[#children+1] = lumina.createElement("text", {key="b"}, "B")
				end
				return lumina.createElement("vbox", {id="root", children = children})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	comp := e.GetComponent("dynamic")
	if len(comp.RootNode.Children) != 1 {
		t.Fatalf("initial children = %d, want 1", len(comp.RootNode.Children))
	}

	// Add a child
	err = L.DoString(`show_extra = true`)
	if err != nil {
		t.Fatal(err)
	}
	comp.Dirty = true
	e.RenderDirty()

	if len(comp.RootNode.Children) != 2 {
		t.Fatalf("after add: children = %d, want 2", len(comp.RootNode.Children))
	}
	if comp.RootNode.Children[1].Content != "B" {
		t.Errorf("new child content = %q, want 'B'", comp.RootNode.Children[1].Content)
	}

	// Remove the child
	err = L.DoString(`show_extra = false`)
	if err != nil {
		t.Fatal(err)
	}
	comp.Dirty = true
	e.RenderDirty()

	if len(comp.RootNode.Children) != 1 {
		t.Fatalf("after remove: children = %d, want 1", len(comp.RootNode.Children))
	}
}

func TestEngine_TopLevelStyleFields(t *testing.T) {
	e, L := newTestEngine(t)

	// Test style fields at top level (not in style sub-table)
	err := L.DoString(`
		lumina.createComponent({
			id = "toplevel",
			name = "TopLevel",
			render = function(props)
				return lumina.createElement("box", {
					id = "root",
					foreground = "#FFF",
					background = "#000",
					bold = true,
					border = "rounded",
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	comp := e.GetComponent("toplevel")
	s := comp.RootNode.Style

	if s.Foreground != "#FFF" {
		t.Errorf("Foreground = %q, want '#FFF'", s.Foreground)
	}
	if s.Background != "#000" {
		t.Errorf("Background = %q, want '#000'", s.Background)
	}
	if !s.Bold {
		t.Error("Bold should be true")
	}
	if s.Border != "rounded" {
		t.Errorf("Border = %q, want 'rounded'", s.Border)
	}
}

func TestEngine_UseState_Persists(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "persist",
			name = "Persist",
			render = function(props)
				local val, setVal = lumina.useState("x", 100)
				return lumina.createElement("text", {}, tostring(val))
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	comp := e.GetComponent("persist")
	if comp.RootNode.Content != "100" {
		t.Fatalf("initial: expected '100', got %q", comp.RootNode.Content)
	}

	// Set state and re-render
	e.SetState("persist", "x", int64(200))
	e.RenderDirty()

	if comp.RootNode.Content != "200" {
		t.Errorf("after setState: expected '200', got %q", comp.RootNode.Content)
	}

	// Re-render again without changing state — value should persist
	comp.Dirty = true
	e.RenderDirty()

	if comp.RootNode.Content != "200" {
		t.Errorf("after re-render: expected '200' (persisted), got %q", comp.RootNode.Content)
	}
}

func TestEngine_CreateElement_StringContent(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "strtest",
			name = "StringTest",
			render = function(props)
				return lumina.createElement("text", {id="t"}, "hello", " ", "world")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	comp := e.GetComponent("strtest")
	if comp.RootNode.Content != "hello world" {
		t.Errorf("content = %q, want 'hello world'", comp.RootNode.Content)
	}
}

func TestEngine_ReadDescriptor_StringChild(t *testing.T) {
	e, L := newTestEngine(t)

	// Test that string children inside a children array become text nodes
	err := L.DoString(`
		lumina.createComponent({
			id = "strchild",
			name = "StringChild",
			render = function(props)
				return {
					type = "vbox",
					id = "root",
					children = {
						"plain string",
						{ type = "text", content = "table child" },
					}
				}
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	comp := e.GetComponent("strchild")
	if comp.RootNode == nil {
		t.Fatal("RootNode nil")
	}
	if len(comp.RootNode.Children) != 2 {
		t.Fatalf("children = %d, want 2", len(comp.RootNode.Children))
	}
	if comp.RootNode.Children[0].Type != "text" {
		t.Errorf("child 0 type = %q, want 'text'", comp.RootNode.Children[0].Type)
	}
	if comp.RootNode.Children[0].Content != "plain string" {
		t.Errorf("child 0 content = %q, want 'plain string'", comp.RootNode.Children[0].Content)
	}
	if comp.RootNode.Children[1].Content != "table child" {
		t.Errorf("child 1 content = %q, want 'table child'", comp.RootNode.Children[1].Content)
	}
}
