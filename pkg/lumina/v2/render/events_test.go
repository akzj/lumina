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


func TestEngine_StaleHoveredNode_AfterRemoval(t *testing.T) {
	// Bug #2: hoveredNode points to an orphaned node after reconcile removes it.
	// Should not panic and should clear the stale pointer.
	e, L := newTestEngine(t)

	err := L.DoString(`
		show_child = true
		lumina.createComponent({
			id = "stale",
			name = "Stale",
			render = function(props)
				if show_child then
					return lumina.createElement("vbox", {style={width=80, height=24}},
						lumina.createElement("box", {
							key = "target",
							id = "target",
							style = {width = 40, height = 12},
							onMouseEnter = function() end,
							onMouseLeave = function() end,
						}),
						lumina.createElement("box", {
							key = "other",
							style = {width = 40, height = 12},
						})
					)
				else
					return lumina.createElement("vbox", {style={width=80, height=24}},
						lumina.createElement("box", {
							key = "other",
							style = {width = 40, height = 12},
						})
					)
				end
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Hover over the "target" box
	e.HandleMouseMove(5, 3)

	// Remove the target box by toggling show_child
	err = L.DoString(`show_child = false`)
	if err != nil {
		t.Fatal(err)
	}
	e.GetComponent("stale").Dirty = true
	e.MarkNeedsRender()
	e.RenderDirty()

	// Now hoveredNode points to a removed node. Moving the mouse should not panic.
	e.HandleMouseMove(5, 15) // move to a different position
	// If we get here without panic, the stale pointer was handled correctly.
}

func TestEngine_StaleFocusedNode_AfterRemoval(t *testing.T) {
	// Bug #2: focusedNode points to an orphaned node after reconcile removes it.
	e, L := newTestEngine(t)

	err := L.DoString(`
		show_input = true
		lumina.createComponent({
			id = "stale_focus",
			name = "StaleFocus",
			render = function(props)
				if show_input then
					return lumina.createElement("vbox", {style={width=80, height=24}},
						lumina.createElement("input", {
							key = "inp",
							id = "inp",
							style = {width = 40, height = 1},
						})
					)
				else
					return lumina.createElement("vbox", {style={width=80, height=24}},
						lumina.createElement("text", {key = "txt"}, "no input")
					)
				end
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	// Focus the input by clicking it
	e.HandleClick(5, 0)
	if e.FocusedNode() == nil {
		t.Fatal("expected input to be focused after click")
	}

	// Remove the input
	err = L.DoString(`show_input = false`)
	if err != nil {
		t.Fatal(err)
	}
	e.GetComponent("stale_focus").Dirty = true
	e.MarkNeedsRender()
	e.RenderDirty()

	// Now try to type — should not panic, focused node should be cleared
	e.HandleKeyDown("a")
	// If we get here without panic, the stale pointer was handled correctly.
}

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

// --- Auto-scroll tests ---

func TestEngine_AutoScroll(t *testing.T) {
	e, L := newTestEngine(t)

	// Create a vbox with overflow=scroll, height=5, containing 20 text children (each h=1)
	err := L.DoString(`
		local children = {}
		for i = 1, 20 do
			children[i] = lumina.createElement("text", {style = {height = 1}}, "Line " .. i)
		end
		lumina.createComponent({
			id = "auto_scroll",
			name = "AutoScroll",
			render = function(props)
				return lumina.createElement("vbox", {
					style = {overflow = "scroll", height = 5, width = 20},
					children = children,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	root := e.Root()
	if root == nil || root.RootNode == nil {
		t.Fatal("no root node")
	}

	// Find the scroll container (root's child with overflow=scroll)
	scrollNode := root.RootNode
	if scrollNode.Style.Overflow != "scroll" {
		// It might be nested
		if len(scrollNode.Children) > 0 {
			scrollNode = scrollNode.Children[0]
		}
	}
	if scrollNode.Style.Overflow != "scroll" {
		t.Fatal("could not find scroll container")
	}

	// Initial scrollY should be 0
	if scrollNode.ScrollY != 0 {
		t.Errorf("initial ScrollY = %d, want 0", scrollNode.ScrollY)
	}

	// Scroll down (delta=1 means scroll down)
	e.HandleScroll(5, 2, 1)

	if scrollNode.ScrollY <= 0 {
		t.Errorf("after scroll down: ScrollY = %d, want > 0", scrollNode.ScrollY)
	}

	savedY := scrollNode.ScrollY

	// Scroll up (delta=-1 means scroll up)
	e.HandleScroll(5, 2, -1)

	if scrollNode.ScrollY >= savedY {
		t.Errorf("after scroll up: ScrollY = %d, want < %d", scrollNode.ScrollY, savedY)
	}

	// Scroll up past 0 — should clamp at 0
	for i := 0; i < 20; i++ {
		e.HandleScroll(5, 2, -1)
	}
	if scrollNode.ScrollY != 0 {
		t.Errorf("after many scroll ups: ScrollY = %d, want 0", scrollNode.ScrollY)
	}

	// Scroll down past max — should clamp
	for i := 0; i < 100; i++ {
		e.HandleScroll(5, 2, 1)
	}
	maxScroll := computeMaxScrollY(scrollNode)
	if scrollNode.ScrollY != maxScroll {
		t.Errorf("after many scroll downs: ScrollY = %d, want maxScroll=%d", scrollNode.ScrollY, maxScroll)
	}
}

func TestEngine_AutoScroll_CustomHandlerPriority(t *testing.T) {
	e, L := newTestEngine(t)

	// Create a node with BOTH overflow=scroll AND onScroll handler
	err := L.DoString(`
		scroll_called = false
		local children = {}
		for i = 1, 20 do
			children[i] = lumina.createElement("text", {style = {height = 1}}, "Line " .. i)
		end
		lumina.createComponent({
			id = "priority_test",
			name = "PriorityTest",
			render = function(props)
				return lumina.createElement("vbox", {
					style = {overflow = "scroll", height = 5, width = 20},
					onScroll = function(e)
						scroll_called = true
					end,
					children = children,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	root := e.Root()
	if root == nil || root.RootNode == nil {
		t.Fatal("no root node")
	}

	// Scroll — custom handler should be called, NOT auto-scroll
	e.HandleScroll(5, 2, 1)

	// Verify Lua handler was called
	L.GetGlobal("scroll_called")
	called := L.ToBoolean(-1)
	L.Pop(1)
	if !called {
		t.Error("expected custom onScroll handler to be called")
	}

	// Verify auto-scroll did NOT change ScrollY
	scrollNode := root.RootNode
	if len(scrollNode.Children) > 0 && scrollNode.Children[0].Style.Overflow == "scroll" {
		scrollNode = scrollNode.Children[0]
	}
	if scrollNode.ScrollY != 0 {
		t.Errorf("auto-scroll should NOT have fired; ScrollY = %d, want 0", scrollNode.ScrollY)
	}
}

func TestEngine_AutoScroll_PreservedAcrossReRender(t *testing.T) {
	e, L := newTestEngine(t)

	// Create a scroll container with re-render trigger (useState counter)
	err := L.DoString(`
		lumina.createComponent({
			id = "scroll_preserve",
			name = "ScrollPreserve",
			render = function(props)
				local count, setCount = lumina.useState("count", 0)
				local children = {}
				for i = 1, 20 do
					children[i] = lumina.createElement("text", {style = {height = 1}}, "Line " .. i)
				end
				return lumina.createElement("vbox", {
					style = {overflow = "scroll", height = 5, width = 20},
					onKeyDown = function(e)
						if e.key == "r" then setCount(count + 1) end
					end,
					children = children,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	root := e.Root()
	if root == nil || root.RootNode == nil {
		t.Fatal("no root node")
	}

	// Find the scroll container
	scrollNode := root.RootNode
	if scrollNode.Style.Overflow != "scroll" {
		if len(scrollNode.Children) > 0 {
			scrollNode = scrollNode.Children[0]
		}
	}
	if scrollNode.Style.Overflow != "scroll" {
		t.Fatal("could not find scroll container")
	}

	// Auto-scroll down
	e.HandleScroll(5, 2, 1)
	e.HandleScroll(5, 2, 1)
	scrollYBefore := scrollNode.ScrollY
	if scrollYBefore <= 0 {
		t.Fatalf("expected scrollY > 0 after scrolling, got %d", scrollYBefore)
	}

	// Trigger a re-render via setState (simulated by key event)
	e.HandleKeyDown("r")
	e.RenderDirty()

	// scrollY should be preserved (Lua doesn't set scrollY, so reconciler should not reset it)
	if scrollNode.ScrollY != scrollYBefore {
		t.Errorf("scrollY changed after re-render: got %d, want %d", scrollNode.ScrollY, scrollYBefore)
	}
}

func TestEngine_AutoScroll_ClampWhenContentShrinks(t *testing.T) {
	e, L := newTestEngine(t)

	// Create a scroll container whose content count depends on state
	err := L.DoString(`
		lumina.createComponent({
			id = "scroll_clamp",
			name = "ScrollClamp",
			render = function(props)
				local small, setSmall = lumina.useState("small", false)
				local count = 20
				if small then count = 3 end
				local children = {}
				for i = 1, count do
					children[i] = lumina.createElement("text", {style = {height = 1}}, "Line " .. i)
				end
				return lumina.createElement("vbox", {
					style = {overflow = "scroll", height = 5, width = 20},
					onKeyDown = function(e)
						if e.key == "s" then setSmall(true) end
					end,
					children = children,
				})
			end,
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	e.RenderAll()

	root := e.Root()
	if root == nil || root.RootNode == nil {
		t.Fatal("no root node")
	}

	// Find the scroll container
	scrollNode := root.RootNode
	if scrollNode.Style.Overflow != "scroll" {
		if len(scrollNode.Children) > 0 {
			scrollNode = scrollNode.Children[0]
		}
	}
	if scrollNode.Style.Overflow != "scroll" {
		t.Fatal("could not find scroll container")
	}

	// Scroll down a lot
	for i := 0; i < 20; i++ {
		e.HandleScroll(5, 2, 1)
	}
	if scrollNode.ScrollY <= 0 {
		t.Fatalf("expected scrollY > 0 after scrolling, got %d", scrollNode.ScrollY)
	}

	// Shrink content: trigger setState to reduce children from 20 to 3
	e.HandleKeyDown("s")
	e.RenderDirty()

	// After paint, scrollY should be clamped to new maxScroll (3 - 5 = 0 since content fits)
	e.RenderAll() // ensure layout + paint runs
	maxScroll := computeMaxScrollY(scrollNode)
	if scrollNode.ScrollY > maxScroll {
		t.Errorf("scrollY not clamped: got %d, maxScroll=%d", scrollNode.ScrollY, maxScroll)
	}
	if maxScroll == 0 && scrollNode.ScrollY != 0 {
		t.Errorf("content fits in container, scrollY should be 0, got %d", scrollNode.ScrollY)
	}
}

func TestHitTest_ScrollOffset(t *testing.T) {
	// Build a scroll container manually with children at known positions
	parent := NewNode("vbox")
	parent.X = 0
	parent.Y = 0
	parent.W = 20
	parent.H = 5
	parent.Style.Overflow = "scroll"
	parent.ScrollY = 10 // scrolled down 10 rows

	// 20 children, each height=1, at Y=0..19
	for i := 0; i < 20; i++ {
		child := NewNode("text")
		child.X = 0
		child.Y = i
		child.W = 20
		child.H = 1
		child.Content = "item"
		child.Parent = parent
		parent.Children = append(parent.Children, child)
	}

	// Screen Y=2 with scrollY=10 should hit child at layout Y=12 (item index 12)
	hit := HitTest(parent, 5, 2)
	if hit == nil {
		t.Fatal("expected a hit, got nil")
	}
	// The hit should be the child at Y=12 (index 12)
	if hit.Y != 12 {
		t.Errorf("HitTest with scroll offset: hit Y=%d, want 12", hit.Y)
	}

	// Screen Y=0 with scrollY=10 should hit child at layout Y=10
	hit = HitTest(parent, 5, 0)
	if hit == nil {
		t.Fatal("expected a hit at Y=0, got nil")
	}
	if hit.Y != 10 {
		t.Errorf("HitTest at top: hit Y=%d, want 10", hit.Y)
	}

	// Screen Y=4 with scrollY=10 should hit child at layout Y=14
	hit = HitTest(parent, 5, 4)
	if hit == nil {
		t.Fatal("expected a hit at Y=4, got nil")
	}
	if hit.Y != 14 {
		t.Errorf("HitTest at bottom: hit Y=%d, want 14", hit.Y)
	}
}
