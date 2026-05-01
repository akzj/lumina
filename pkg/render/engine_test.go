package render

import (
	"strings"
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

func TestEngine_ReadDescriptor_TextOnClick(t *testing.T) {
	e, L := newTestEngine(t)
	err := L.DoString(`
		t = lumina.createElement("text", {
			foreground = "#89B4FA",
			bold = true,
			onClick = function() end,
		}, "  [ OK ]  ")
	`)
	if err != nil {
		t.Fatal(err)
	}
	L.GetGlobal("t")
	desc := e.readDescriptor(L, -1)
	L.Pop(1)
	if desc.Type != "text" {
		t.Fatalf("type: got %q want text", desc.Type)
	}
	if desc.OnClick == 0 {
		t.Fatalf("readDescriptor lost text onClick (same shape as list_dialog OK button)")
	}
}

func findDescTextOnClick(d Descriptor, sub string) (onClick LuaRef, ok bool) {
	if d.Type == "text" && strings.Contains(d.Content, sub) {
		return d.OnClick, true
	}
	for _, ch := range d.Children {
		if ref, found := findDescTextOnClick(ch, sub); found {
			return ref, true
		}
	}
	return 0, false
}

func TestEngine_ReadDescriptor_LuxDialogLikeVBoxOnClick(t *testing.T) {
	e, L := newTestEngine(t)
	err := L.DoString(`
		local ok = lumina.createElement("text", {
			foreground = "#89B4FA",
			bold = true,
			onClick = function() end,
		}, "  [ OK ]  ")
		local actionsSlot = { children = { ok } }
		dlgLike = lumina.createElement("vbox", {
			style = {
				border = "rounded",
				padding = 1,
				width = 40,
				background = "#313244",
			},
		},
			lumina.createElement("text", { foreground = "#89B4FA", bold = true }, "T"),
			lumina.createElement("text", { foreground = "#6C7086", dim = true }, string.rep("-", 36)),
			lumina.createElement("text", { foreground = "#CDD6F4" }, "Body"),
			lumina.createElement("text", { foreground = "#6C7086", dim = true }, string.rep("-", 36)),
			lumina.createElement("hbox", { style = { gap = 1 } }, table.unpack(actionsSlot.children))
		)
	`)
	if err != nil {
		t.Fatal(err)
	}
	L.GetGlobal("dlgLike")
	desc := e.readDescriptor(L, -1)
	L.Pop(1)
	ref, ok := findDescTextOnClick(desc, "[ OK ]")
	if !ok {
		t.Fatal("descriptor tree missing OK text")
	}
	if ref == 0 {
		t.Fatal("LuxDialog-like vbox lost OK onClick at readDescriptor stage")
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
	e.MarkNeedsRender()
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
	e.MarkNeedsRender()
	e.RenderDirty()

	if len(comp.RootNode.Children) != 1 {
		t.Fatalf("after remove: children = %d, want 1", len(comp.RootNode.Children))
	}
}

func TestEngine_ComponentCleanup_RemovedFromMap(t *testing.T) {
	// Bug #3: Child components should be removed from e.components when
	// the parent stops rendering them.
	e, L := newTestEngine(t)

	err := L.DoString(`
		Cell = lumina.defineComponent("Cell", function(props)
			return lumina.createElement("text", {}, "cell")
		end)

		show_cells = true
		lumina.createComponent({
			id = "root",
			name = "Root",
			render = function(props)
				if show_cells then
					return lumina.createElement("hbox", {id = "row"},
						lumina.createElement(Cell, {key = "a", id = "a"}),
						lumina.createElement(Cell, {key = "b", id = "b"})
					)
				else
					return lumina.createElement("hbox", {id = "row"})
				end
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Should have root + 2 child components = 3 total
	allComps := e.AllComponents()
	if len(allComps) != 3 {
		t.Fatalf("initial: expected 3 components, got %d", len(allComps))
	}

	// Verify child components exist
	if e.GetComponent("root:a") == nil {
		t.Fatal("child 'root:a' not found")
	}
	if e.GetComponent("root:b") == nil {
		t.Fatal("child 'root:b' not found")
	}

	// Toggle off — remove all child components
	err = L.DoString(`show_cells = false`)
	if err != nil {
		t.Fatal(err)
	}
	e.GetComponent("root").Dirty = true
	e.MarkNeedsRender()
	e.RenderDirty()

	// Should have only root component = 1 total
	allComps = e.AllComponents()
	if len(allComps) != 1 {
		t.Errorf("after removal: expected 1 component, got %d", len(allComps))
		for id := range allComps {
			t.Logf("  remaining: %s", id)
		}
	}

	// Child components should be gone
	if e.GetComponent("root:a") != nil {
		t.Error("child 'root:a' should have been removed")
	}
	if e.GetComponent("root:b") != nil {
		t.Error("child 'root:b' should have been removed")
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
	e.MarkNeedsRender()
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

func TestEngine_GraftWalk_SkipWhenAlreadyGrafted(t *testing.T) {
	e, L := newTestEngine(t)

	// Define a child component
	err := L.DoString(`
		local Cell = lumina.defineComponent("Cell", function(props)
			return lumina.createElement("text", {}, "cell")
		end)

		lumina.createComponent({
			id = "root",
			name = "Root",
			render = function(props)
				return lumina.createElement("box", {},
					lumina.createElement(Cell, {key = "c1"}),
					lumina.createElement(Cell, {key = "c2"})
				)
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Initial full render + graft
	e.RenderAll()

	root := e.Root().RootNode
	if root == nil {
		t.Fatal("root node nil after RenderAll")
	}

	// Find component placeholder nodes and verify they are grafted
	var compNodes []*Node
	var findCompNodes func(n *Node)
	findCompNodes = func(n *Node) {
		if n.Type == "component" {
			compNodes = append(compNodes, n)
		}
		for _, ch := range n.Children {
			findCompNodes(ch)
		}
	}
	findCompNodes(root)
	if len(compNodes) < 2 {
		t.Fatalf("expected at least 2 component nodes, got %d", len(compNodes))
	}

	// Clear all dirty flags to simulate idle state
	clearLayoutDirty(root)
	clearPaintDirty(root)

	// Verify no node is dirty
	if hasAnyDirty(root) {
		t.Fatal("expected no dirty nodes after clearing")
	}

	// Call RenderDirty (idle frame) — should NOT mark anything dirty
	e.RenderDirty()

	// After idle RenderDirty, nothing should be dirty (graftWalk should skip)
	for i, cn := range compNodes {
		if cn.LayoutDirty {
			t.Errorf("component node %d: LayoutDirty=true after idle RenderDirty", i)
		}
		if cn.PaintDirty {
			t.Errorf("component node %d: PaintDirty=true after idle RenderDirty", i)
		}
	}
}

func TestEngine_IdleFrame_NoPaintWork(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "test",
			name = "Test",
			render = function(props)
				return lumina.createElement("box", {
					style = {background = "#000"},
				},
					lumina.createElement("text", {}, "Hello")
				)
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Run idle frame
	e.RenderDirty()

	// Check that no paint work was done
	stats := e.Buffer().Stats()
	if stats.WriteCount != 0 {
		t.Errorf("idle frame: WriteCount=%d, want 0", stats.WriteCount)
	}
	if stats.ClearCount != 0 {
		t.Errorf("idle frame: ClearCount=%d, want 0", stats.ClearCount)
	}
}

func BenchmarkEngine_IdleFrame(b *testing.B) {
	L := lua.NewState()
	defer L.Close()
	e := NewEngine(L, 80, 24)
	e.RegisterLuaAPI()

	// Create a component with many child components (simulates stress_test.lua)
	err := L.DoString(`
		local Cell = lumina.defineComponent("Cell", function(props)
			return lumina.createElement("text", {
				style = {foreground = "#FFF"},
			}, "X")
		end)

		lumina.createComponent({
			id = "root",
			name = "Root",
			render = function(props)
				local children = {}
				for i = 1, 100 do
					children[i] = lumina.createElement(Cell, {key = "c" .. i})
				end
				return lumina.createElement("box", {}, table.unpack(children))
			end,
		})
	`)
	if err != nil {
		b.Fatal(err)
	}

	e.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.RenderDirty()
	}
}

func BenchmarkEngine_IdleFrame_WithSubComponents(b *testing.B) {
	L := lua.NewState()
	defer L.Close()
	e := NewEngine(L, 160, 48)
	e.RegisterLuaAPI()

	err := L.DoString(`
		local Cell = lumina.defineComponent("Cell", function(props)
			return lumina.createElement("box", {
				style = {width = 3, height = 1, background = "#333"},
			}, lumina.createElement("text", {}, "X"))
		end)

		lumina.createComponent({
			id = "root",
			name = "Root",
			render = function(props)
				local rows = {}
				for r = 1, 20 do
					local cells = {}
					for c = 1, 40 do
						cells[c] = lumina.createElement(Cell, {key = "c" .. r .. "_" .. c})
					end
					rows[r] = lumina.createElement("hbox", {key = "row" .. r}, table.unpack(cells))
				end
				return lumina.createElement("box", {}, table.unpack(rows))
			end,
		})
	`)
	if err != nil {
		b.Fatal(err)
	}

	e.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.RenderDirty()
	}
}

func TestHoverLeaveRestoresBackground(t *testing.T) {
	// Simulates stress_test.lua: 80 Cell components in an hbox row (width=1 each).
	// Each Cell has its own hover state. Hovering cell 0, then cell 1, then
	// moving away should restore backgrounds correctly.
	e, L := newTestEngine(t)

	err := L.DoString(`
		local theme_bg = "#1E1E2E"
		local theme_hoverBg = "#313244"

		Cell = lumina.defineComponent("Cell", function(props)
			local hovered, setHovered = lumina.useState("h", false)

			local bg
			if hovered then
				bg = theme_hoverBg
			else
				bg = theme_bg
			end

			return lumina.createElement("box", {
				style = {width = 1, height = 1, background = bg},
				onMouseEnter = function() setHovered(true) end,
				onMouseLeave = function() setHovered(false) end,
			}, lumina.createElement("text", {
				style = {foreground = "#FFF"},
			}, hovered and "H" or "."))
		end)

		lumina.createComponent({
			id = "root",
			name = "Root",
			render = function(props)
				local cells = {}
				for x = 0, 79 do
					local id = tostring(x)
					cells[#cells + 1] = lumina.createElement(Cell, {key = id, id = id})
				end
				return lumina.createElement("hbox", {
					id = "row",
					style = {height = 1},
				}, table.unpack(cells))
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Verify initial state: first few cells should have theme_bg
	for x := 0; x < 3; x++ {
		cell := e.buffer.Get(x, 0)
		if cell.BG != "#1E1E2E" {
			t.Errorf("initial: cell(%d,0).BG = %q, want '#1E1E2E'", x, cell.BG)
		}
		if cell.Ch != '.' {
			t.Errorf("initial: cell(%d,0).Ch = %q, want '.'", x, string(cell.Ch))
		}
	}

	// Hover cell 0
	e.HandleMouseMove(0, 0)
	e.RenderDirty()

	cell0 := e.buffer.Get(0, 0)
	if cell0.BG != "#313244" {
		t.Errorf("after hover c0: cell(0,0).BG = %q, want '#313244'", cell0.BG)
	}
	if cell0.Ch != 'H' {
		t.Errorf("after hover c0: cell(0,0).Ch = %q, want 'H'", string(cell0.Ch))
	}

	// Move hover to cell 1 → cell 0 should restore, cell 1 should highlight
	e.HandleMouseMove(1, 0)
	e.RenderDirty()

	cell0After := e.buffer.Get(0, 0)
	if cell0After.BG != "#1E1E2E" {
		t.Errorf("after hover c1: cell(0,0).BG = %q, want '#1E1E2E' (restored)", cell0After.BG)
	}
	if cell0After.Ch != '.' {
		t.Errorf("after hover c1: cell(0,0).Ch = %q, want '.' (restored)", string(cell0After.Ch))
	}

	cell1 := e.buffer.Get(1, 0)
	if cell1.BG != "#313244" {
		t.Errorf("after hover c1: cell(1,0).BG = %q, want '#313244'", cell1.BG)
	}
	if cell1.Ch != 'H' {
		t.Errorf("after hover c1: cell(1,0).Ch = %q, want 'H'", string(cell1.Ch))
	}

	// Move hover away (outside all cells) → cell 1 should restore
	e.HandleMouseMove(5, 5)
	e.RenderDirty()

	cell1After := e.buffer.Get(1, 0)
	if cell1After.BG != "#1E1E2E" {
		t.Errorf("after hover away: cell(1,0).BG = %q, want '#1E1E2E' (restored)", cell1After.BG)
	}
	if cell1After.Ch != '.' {
		t.Errorf("after hover away: cell(1,0).Ch = %q, want '.' (restored)", string(cell1After.Ch))
	}
}

func TestReconcileChildComponents_KeyFallback(t *testing.T) {
	// Test that reconcileChildComponents uses Key as fallback when ID is empty.
	// This simulates createElement(Cell, {key = "c0"}) WITHOUT an id prop.
	e, L := newTestEngine(t)

	err := L.DoString(`
		Cell = lumina.defineComponent("Cell", function(props)
			return lumina.createElement("text", {}, "cell")
		end)

		lumina.createComponent({
			id = "root",
			name = "Root",
			render = function(props)
				return lumina.createElement("hbox", {id = "row"},
					lumina.createElement(Cell, {key = "a"}),
					lumina.createElement(Cell, {key = "b"}),
					lumina.createElement(Cell, {key = "c"})
				)
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Should have 3 separate child components (not just 1)
	root := e.GetComponent("root")
	if root == nil {
		t.Fatal("root component nil")
	}
	if len(root.Children) != 3 {
		t.Errorf("expected 3 child components, got %d", len(root.Children))
		for i, ch := range root.Children {
			t.Logf("  child %d: ID=%q Type=%q", i, ch.ID, ch.Type)
		}
	}
}

func TestReconcileChildComponents_SurvivesParentRerender(t *testing.T) {
	// Verify that child components are found (not recreated) when the parent re-renders.
	// This tests the FindChild/AddChild key consistency fix.
	e, L := newTestEngine(t)

	err := L.DoString(`
		Cell = lumina.defineComponent("Cell", function(props)
			local val, setVal = lumina.useState("v", 0)
			return lumina.createElement("text", {}, tostring(val))
		end)

		lumina.createComponent({
			id = "root",
			name = "Root",
			render = function(props)
				local count, setCount = lumina.useState("n", 0)
				return lumina.createElement("hbox", {id = "row"},
					lumina.createElement(Cell, {key = "a", id = "a"}),
					lumina.createElement(Cell, {key = "b", id = "b"})
				)
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	root := e.GetComponent("root")
	if root == nil {
		t.Fatal("root component nil")
	}
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 child components, got %d", len(root.Children))
	}

	// Set state on child "a" to give it unique state
	childA := e.GetComponent("root:a")
	if childA == nil {
		t.Fatal("child component 'root:a' not found")
	}
	e.SetState("root:a", "v", int64(42))

	// Force parent re-render
	root.Dirty = true
	e.MarkNeedsRender()
	e.RenderDirty()

	// Child count should still be 2 (not 4 due to duplicates)
	if len(root.Children) != 2 {
		t.Errorf("after parent re-render: expected 2 child components, got %d", len(root.Children))
		for i, ch := range root.Children {
			t.Logf("  child %d: ID=%q Type=%q", i, ch.ID, ch.Type)
		}
	}

	// Child A's state should be preserved (same component, not recreated)
	childAAfter := e.GetComponent("root:a")
	if childAAfter == nil {
		t.Fatal("child 'root:a' not found after parent re-render")
	}
	if childAAfter != childA {
		t.Error("child 'root:a' was recreated (different pointer) — should be reused")
	}
	if childAAfter.State["v"] != int64(42) {
		t.Errorf("child 'root:a' state lost: v=%v, want 42", childAAfter.State["v"])
	}
}

// --- NeedsRender flag tests ---

func TestEngine_NeedsRender_FalseWhenIdle(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "idle",
			name = "Idle",
			render = function(props)
				return lumina.createElement("text", {}, "hello")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// After RenderAll, needsRender should be false
	if e.NeedsRender() {
		t.Error("expected NeedsRender=false after RenderAll")
	}

	// RenderDirty should be a no-op
	e.RenderDirty()

	if e.NeedsRender() {
		t.Error("expected NeedsRender=false after idle RenderDirty")
	}

	// Verify no paint work was done
	stats := e.Buffer().Stats()
	if stats.WriteCount != 0 {
		t.Errorf("expected 0 writes on idle frame, got %d", stats.WriteCount)
	}
	if stats.ClearCount != 0 {
		t.Errorf("expected 0 clears on idle frame, got %d", stats.ClearCount)
	}
}

func TestEngine_NeedsRender_TrueAfterSetState(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "state_test",
			name = "StateTest",
			render = function(props)
				local val = lumina.useState("x", 0)
				return lumina.createElement("text", {}, tostring(val))
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	e.RenderDirty() // clear any residual

	if e.NeedsRender() {
		t.Error("expected NeedsRender=false before SetState")
	}

	// SetState should mark needsRender
	e.SetState("state_test", "x", int64(42))
	if !e.NeedsRender() {
		t.Error("expected NeedsRender=true after SetState")
	}

	// RenderDirty should process the change and clear the flag
	e.RenderDirty()
	if e.NeedsRender() {
		t.Error("expected NeedsRender=false after RenderDirty processed the change")
	}
}

func TestEngine_NeedsRender_TrueAfterResize(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "resize_test",
			name = "ResizeTest",
			render = function(props)
				return lumina.createElement("box", {})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	e.RenderDirty()

	if e.NeedsRender() {
		t.Error("expected NeedsRender=false before Resize")
	}

	e.Resize(100, 50)
	if !e.NeedsRender() {
		t.Error("expected NeedsRender=true after Resize")
	}
}

func TestEngine_NeedsRender_SetStateSameValue_NoRender(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "same_val",
			name = "SameVal",
			render = function(props)
				local val = lumina.useState("x", 10)
				return lumina.createElement("text", {}, tostring(val))
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	e.RenderDirty()

	// SetState with the same value should NOT mark needsRender
	e.SetState("same_val", "x", int64(10))
	if e.NeedsRender() {
		t.Error("expected NeedsRender=false when SetState with same value")
	}
}

func TestEngine_CreateLayer_RemoveLayer(t *testing.T) {
	e, L := newTestEngine(t)

	// Create a component so the engine has a main layer
	err := L.DoString(`
		lumina.createComponent({
			id = "main",
			name = "Main",
			render = function(props)
				return lumina.createElement("text", {}, "Main content")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}
	e.RenderAll()

	// Create a layer via Lua API
	err = L.DoString(`
		lumina.createLayer("overlay1",
			lumina.createElement("vbox", {
				id = "layer-root",
				style = {width = 20, height = 5, background = "#313244"},
			},
				lumina.createElement("text", {id = "item1"}, "Item 1"),
				lumina.createElement("text", {id = "item2"}, "Item 2")
			),
			{ modal = true }
		)
	`)
	if err != nil {
		t.Fatalf("createLayer failed: %v", err)
	}

	// Verify the layer was created
	if len(e.layers) < 2 {
		t.Fatalf("expected at least 2 layers, got %d", len(e.layers))
	}

	var overlayLayer *Layer
	for _, l := range e.layers {
		if l.ID == "overlay1" {
			overlayLayer = l
			break
		}
	}
	if overlayLayer == nil {
		t.Fatal("overlay1 layer not found")
	}
	if !overlayLayer.Modal {
		t.Error("expected modal=true")
	}
	if overlayLayer.Root == nil {
		t.Fatal("layer root is nil")
	}
	if overlayLayer.Root.Type != "vbox" {
		t.Errorf("expected root type 'vbox', got %q", overlayLayer.Root.Type)
	}
	if overlayLayer.Root.ID != "layer-root" {
		t.Errorf("expected root ID 'layer-root', got %q", overlayLayer.Root.ID)
	}
	if len(overlayLayer.Root.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(overlayLayer.Root.Children))
	}

	// Remove the layer via Lua API
	err = L.DoString(`lumina.removeLayer("overlay1")`)
	if err != nil {
		t.Fatalf("removeLayer failed: %v", err)
	}

	// Verify the layer was removed
	for _, l := range e.layers {
		if l.ID == "overlay1" {
			t.Fatal("overlay1 layer should have been removed")
		}
	}
}

func TestEngine_CreateLayer_NonModal(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "main",
			name = "Main",
			render = function(props)
				return lumina.createElement("text", {}, "Main")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}
	e.RenderAll()

	// Create non-modal layer (no options table)
	err = L.DoString(`
		lumina.createLayer("tooltip",
			lumina.createElement("text", {id = "tip"}, "Tooltip text")
		)
	`)
	if err != nil {
		t.Fatalf("createLayer (no options) failed: %v", err)
	}

	var tipLayer *Layer
	for _, l := range e.layers {
		if l.ID == "tooltip" {
			tipLayer = l
			break
		}
	}
	if tipLayer == nil {
		t.Fatal("tooltip layer not found")
	}
	if tipLayer.Modal {
		t.Error("expected modal=false for tooltip layer")
	}
	if tipLayer.Root == nil || tipLayer.Root.Content != "Tooltip text" {
		t.Error("tooltip layer content mismatch")
	}
}

func TestEngine_RemoveLayer_NonExistent(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "main",
			name = "Main",
			render = function(props)
				return lumina.createElement("text", {}, "Main")
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}
	e.RenderAll()

	initialLayers := len(e.layers)

	// Removing a non-existent layer should not panic or error
	err = L.DoString(`lumina.removeLayer("does-not-exist")`)
	if err != nil {
		t.Fatalf("removeLayer of non-existent layer should not error: %v", err)
	}

	if len(e.layers) != initialLayers {
		t.Error("layer count should not change when removing non-existent layer")
	}
}
