package render

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestHitTest_Basic(t *testing.T) {
	root := NewNode("box")
	root.X, root.Y, root.W, root.H = 0, 0, 80, 24

	child := NewNode("box")
	child.X, child.Y, child.W, child.H = 10, 5, 20, 10
	child.ID = "inner"
	root.AddChild(child)

	// Hit inside child
	hit := HitTest(root, 15, 8)
	if hit == nil || hit.ID != "inner" {
		t.Error("expected hit on inner box")
	}

	// Hit outside child but inside root
	hit = HitTest(root, 5, 3)
	if hit != root {
		t.Error("expected hit on root")
	}

	// Hit outside root
	hit = HitTest(root, 100, 100)
	if hit != nil {
		t.Error("expected nil hit")
	}
}

func TestHitTest_DeepNested(t *testing.T) {
	root := NewNode("box")
	root.X, root.Y, root.W, root.H = 0, 0, 80, 24

	mid := NewNode("box")
	mid.X, mid.Y, mid.W, mid.H = 5, 5, 30, 15
	root.AddChild(mid)

	leaf := NewNode("text")
	leaf.X, leaf.Y, leaf.W, leaf.H = 10, 8, 10, 1
	leaf.ID = "leaf"
	mid.AddChild(leaf)

	hit := HitTest(root, 12, 8)
	if hit == nil || hit.ID != "leaf" {
		t.Error("expected deepest node (leaf)")
	}
}

func TestHitTest_LastChildOnTop(t *testing.T) {
	root := NewNode("box")
	root.X, root.Y, root.W, root.H = 0, 0, 80, 24

	// Two overlapping children — last one should win
	child1 := NewNode("box")
	child1.X, child1.Y, child1.W, child1.H = 0, 0, 20, 10
	child1.ID = "first"
	root.AddChild(child1)

	child2 := NewNode("box")
	child2.X, child2.Y, child2.W, child2.H = 0, 0, 20, 10
	child2.ID = "second"
	root.AddChild(child2)

	hit := HitTest(root, 5, 5)
	if hit == nil || hit.ID != "second" {
		t.Error("expected last child (on top)")
	}
}

func TestHitTestWithHandler_Bubbling(t *testing.T) {
	root := NewNode("box")
	root.X, root.Y, root.W, root.H = 0, 0, 80, 24
	root.OnClick = 42 // has handler

	child := NewNode("text")
	child.X, child.Y, child.W, child.H = 5, 5, 10, 1
	// No onClick handler
	root.AddChild(child)

	// Click on child — should bubble up to root
	target := HitTestWithHandler(root, 7, 5, "click")
	if target != root {
		t.Error("expected bubbling to root with onClick handler")
	}
}

func TestHitTestWithHandler_NoHandler(t *testing.T) {
	root := NewNode("box")
	root.X, root.Y, root.W, root.H = 0, 0, 80, 24
	// No handlers at all

	child := NewNode("text")
	child.X, child.Y, child.W, child.H = 5, 5, 10, 1
	root.AddChild(child)

	target := HitTestWithHandler(root, 7, 5, "click")
	if target != nil {
		t.Error("expected nil when no handler exists")
	}
}

func TestEngine_HandleClick(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		clicked = ""
		lumina.createComponent({
			id = "click",
			name = "Click",
			render = function(props)
				return lumina.createElement("box", {
					id = "btn",
					style = {width = 20, height = 5},
					onClick = function() clicked = "btn" end,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	e.HandleClick(10, 2)

	L.GetGlobal("clicked")
	clicked, _ := L.ToString(-1)
	L.Pop(1)

	if clicked != "btn" {
		t.Errorf("expected clicked='btn', got %q", clicked)
	}
}

func TestEngine_HandleClick_StateChange_Reconcile(t *testing.T) {
	// Full cycle: click → setState → RenderDirty → reconcile → verify content
	e, L := newTestEngine(t)

	err := L.DoString(`
		lumina.createComponent({
			id = "counter",
			name = "Counter",
			render = function(props)
				local count, setCount = lumina.useState("c", 0)
				return lumina.createElement("box", {
					style = {width = 80, height = 24},
					onClick = function() setCount(count + 1) end,
				}, lumina.createElement("text", {id="val"}, tostring(count)))
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Verify initial
	comp := e.GetComponent("counter")
	if comp == nil || comp.RootNode == nil {
		t.Fatal("component or root node is nil")
	}
	if len(comp.RootNode.Children) == 0 {
		t.Fatal("no children on root node")
	}
	if comp.RootNode.Children[0].Content != "0" {
		t.Fatalf("initial: expected '0', got %q", comp.RootNode.Children[0].Content)
	}

	// Click → setState(count+1) → RenderDirty → reconcile
	e.HandleClick(10, 10)
	e.RenderDirty()

	if comp.RootNode.Children[0].Content != "1" {
		t.Errorf("after click: expected '1', got %q", comp.RootNode.Children[0].Content)
	}

	// Click again
	e.HandleClick(10, 10)
	e.RenderDirty()

	if comp.RootNode.Children[0].Content != "2" {
		t.Errorf("after 2nd click: expected '2', got %q", comp.RootNode.Children[0].Content)
	}
}

func TestEngine_HandleMouseMove_HoverTracking(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		entered = ""
		left = ""

		lumina.createComponent({
			id = "hover",
			name = "Hover",
			render = function(props)
				return lumina.createElement("vbox", {style={width=80, height=24}},
					lumina.createElement("box", {
						id = "a",
						style = {width = 40, height = 12},
						onMouseEnter = function() entered = "a" end,
						onMouseLeave = function() left = "a" end,
					}),
					lumina.createElement("box", {
						id = "b",
						style = {width = 40, height = 12},
						onMouseEnter = function() entered = "b" end,
						onMouseLeave = function() left = "b" end,
					})
				)
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Move to box "a"
	e.HandleMouseMove(5, 3)
	// Move to box "b" — should fire leave on a, enter on b
	e.HandleMouseMove(5, 15)

	// Verify via Lua globals
	L.GetGlobal("entered")
	enteredVal, _ := L.ToString(-1)
	L.Pop(1)
	L.GetGlobal("left")
	leftVal, _ := L.ToString(-1)
	L.Pop(1)

	if enteredVal != "b" {
		t.Errorf("expected entered='b', got %q", enteredVal)
	}
	if leftVal != "a" {
		t.Errorf("expected left='a', got %q", leftVal)
	}
}

func TestEngine_HoverDoesNotRerenderCleanComponent(t *testing.T) {
	// Hover over a node that doesn't change state → no re-render
	e, L := newTestEngine(t)

	err := L.DoString(`
		renderCount = 0
		lumina.createComponent({
			id = "static",
			name = "Static",
			render = function(props)
				renderCount = renderCount + 1
				return lumina.createElement("box", {
					style = {width = 80, height = 24},
					onMouseEnter = function() end,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Move mouse around — should NOT trigger re-render (handler doesn't call setState)
	e.HandleMouseMove(10, 10)
	e.HandleMouseMove(20, 10)
	e.HandleMouseMove(30, 10)
	e.RenderDirty() // should be no-op

	L.GetGlobal("renderCount")
	count, _ := L.ToInteger(-1)
	L.Pop(1)

	if count != 1 {
		t.Errorf("expected 1 render (initial only), got %d", count)
	}
}

func TestEngine_HandleKeyDown(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		lastKey = ""
		lumina.createComponent({
			id = "keys",
			name = "Keys",
			render = function(props)
				return lumina.createElement("box", {
					style = {width = 80, height = 24},
					onKeyDown = function(ev) lastKey = ev.key end,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	e.HandleKeyDown("q")

	L.GetGlobal("lastKey")
	key, _ := L.ToString(-1)
	L.Pop(1)

	if key != "q" {
		t.Errorf("expected lastKey='q', got %q", key)
	}
}

func TestEngine_HandleScroll(t *testing.T) {
	e, L := newTestEngine(t)

	err := L.DoString(`
		scrollDelta = 0
		lumina.createComponent({
			id = "scroll",
			name = "Scroll",
			render = function(props)
				return lumina.createElement("box", {
					style = {width = 80, height = 24},
					onScroll = function(ev) scrollDelta = ev.delta end,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()
	e.HandleScroll(10, 10, -3)

	L.GetGlobal("scrollDelta")
	delta, _ := L.ToInteger(-1)
	L.Pop(1)

	if delta != -3 {
		t.Errorf("expected scrollDelta=-3, got %d", delta)
	}
}

// Benchmarks

func BenchmarkHitTest_DeepTree(b *testing.B) {
	// Build a tree: root → 10 children → 10 grandchildren each = 100 leaves
	root := NewNode("box")
	root.X, root.Y, root.W, root.H = 0, 0, 80, 24

	for i := 0; i < 10; i++ {
		row := NewNode("hbox")
		row.X = 0
		row.Y = i * 2
		row.W = 80
		row.H = 2
		root.AddChild(row)
		for j := 0; j < 10; j++ {
			cell := NewNode("box")
			cell.X = j * 8
			cell.Y = i * 2
			cell.W = 8
			cell.H = 2
			row.AddChild(cell)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HitTest(root, (i*7)%80, (i*3)%20)
	}
}

func BenchmarkEngine_HoverCycle(b *testing.B) {
	L := lua.NewState()
	defer L.Close()
	e := NewEngine(L, 80, 24)
	e.RegisterLuaAPI()

	// Create a grid with hover handlers
	err := L.DoString(`
		lumina.createComponent({
			id = "stress",
			name = "Stress",
			render = function(props)
				local hovered, setHovered = lumina.useState("h", "")
				local children = {}
				for y = 0, 22 do
					local row = {}
					for x = 0, 79 do
						local id = x..","..y
						row[#row+1] = lumina.createElement("box", {
							id = id,
							key = id,
							style = {width=1, height=1},
							onMouseEnter = function() setHovered(id) end,
							onMouseLeave = function() setHovered("") end,
						})
					end
					children[#children+1] = lumina.createElement("hbox", {key=tostring(y)}, table.unpack(row))
				end
				return lumina.createElement("vbox", {
					style = {width=80, height=24},
				}, table.unpack(children))
			end,
		})
	`)
	if err != nil {
		b.Fatal(err)
	}

	e.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % 80
		y := i % 23
		e.HandleMouseMove(x, y)
		e.RenderDirty()
	}
}

func BenchmarkEngine_RenderDirty_NoChange(b *testing.B) {
	L := lua.NewState()
	defer L.Close()
	e := NewEngine(L, 80, 24)
	e.RegisterLuaAPI()

	_ = L.DoString(`
		lumina.createComponent({
			id = "static",
			name = "Static",
			render = function(props)
				return lumina.createElement("text", {}, "hello")
			end,
		})
	`)
	e.RenderAll()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.RenderDirty() // no-op — nothing dirty
	}
}
