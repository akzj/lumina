package lumina

import (
	"strings"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func advTestState(t *testing.T) *lua.State {
	t.Helper()
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	return L
}

// ============================================================
// Phase 13D: Error Boundaries
// ============================================================

func TestErrorBoundary_CatchesChildRenderError(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Broken = lumina.defineComponent({
			name = "Broken",
			render = function(self)
				error("render exploded!")
			end
		})

		local SafeArea = lumina.createErrorBoundary({
			fallback = function(err)
				return { type = "text", content = "Caught: " .. err }
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(SafeArea, {
					children = {
						lumina.createElement(Broken, {}),
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
		t.Fatalf("expected 1 child, got %d", len(vnode.Children))
	}

	// The SafeArea should have caught the error and rendered fallback
	child := vnode.Children[0]
	// Walk into the boundary's rendered output
	found := findTextContent(child, "Caught:")
	if found == "" {
		t.Fatalf("expected fallback with 'Caught:' text, got tree: %s", describeVNode(child))
	}
}

func TestErrorBoundary_FallbackReceivesErrorMessage(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Broken = lumina.defineComponent({
			name = "Broken",
			render = function(self)
				error("specific error message")
			end
		})

		local SafeArea = lumina.createErrorBoundary({
			fallback = function(err)
				return { type = "text", content = "Error: " .. err }
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(SafeArea, {
					children = {
						lumina.createElement(Broken, {}),
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

	found := findTextContent(vnode, "specific error message")
	if found == "" {
		t.Fatalf("expected error message in fallback, got: %s", describeVNode(vnode))
	}
}

func TestErrorBoundary_NoBoundaryShowsError(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Broken = lumina.defineComponent({
			name = "Broken",
			render = function(self)
				error("no boundary here")
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(Broken, {}),
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Without boundary, should get an error text node
	found := findTextContent(vnode, "Render error:")
	if found == "" {
		// Also acceptable: "no boundary here" in the error
		found = findTextContent(vnode, "no boundary here")
	}
	if found == "" {
		t.Fatalf("expected error text without boundary, got: %s", describeVNode(vnode))
	}
}

func TestErrorBoundary_NestedBoundariesInnerCatchesFirst(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Broken = lumina.defineComponent({
			name = "Broken",
			render = function(self)
				error("inner error")
			end
		})

		local InnerBoundary = lumina.createErrorBoundary({
			fallback = function(err)
				return { type = "text", content = "INNER: " .. err }
			end
		})

		local OuterBoundary = lumina.createErrorBoundary({
			fallback = function(err)
				return { type = "text", content = "OUTER: " .. err }
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(OuterBoundary, {
					children = {
						lumina.createElement(InnerBoundary, {
							children = {
								lumina.createElement(Broken, {}),
							}
						}),
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

	// Inner boundary should catch first
	found := findTextContent(vnode, "INNER:")
	if found == "" {
		t.Fatalf("expected inner boundary to catch, got: %s", describeVNode(vnode))
	}
	// Should NOT see outer boundary
	outer := findTextContent(vnode, "OUTER:")
	if outer != "" {
		t.Fatalf("outer boundary should not have caught, but found: %s", outer)
	}
}

func TestErrorBoundary_NoErrorNoFallback(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	err := L.DoString(`
		local Good = lumina.defineComponent({
			name = "Good",
			render = function(self)
				return { type = "text", content = "all good" }
			end
		})

		local SafeArea = lumina.createErrorBoundary({
			fallback = function(err)
				return { type = "text", content = "ERROR: " .. err }
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(SafeArea, {
					children = {
						lumina.createElement(Good, {}),
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

	// Should see "all good", not "ERROR:"
	good := findTextContent(vnode, "all good")
	if good == "" {
		t.Fatalf("expected 'all good', got: %s", describeVNode(vnode))
	}
	errText := findTextContent(vnode, "ERROR:")
	if errText != "" {
		t.Fatalf("should not see fallback when no error, but found: %s", errText)
	}
}

// ============================================================
// Phase 13E: React.memo
// ============================================================

func TestMemo_SkipsRenderOnSameProps(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	// Set up a parent component context so child components can be found across renders
	parent := &Component{
		ID:    "memo_parent",
		Type:  "MemoParent",
		Name:  "MemoParent",
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
		_render_count = 0
		local Counter = lumina.defineComponent({
			name = "MemoCounter",
			render = function(self)
				_render_count = _render_count + 1
				return { type = "text", content = "count:" .. _render_count }
			end
		})
		_MemoCounter = lumina.memo(Counter)
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// First render
	err = L.DoString(`
		_tree1 = {
			type = "box",
			children = {
				lumina.createElement(_MemoCounter, { value = 1 }),
			}
		}
	`)
	if err != nil {
		t.Fatalf("tree1: %v", err)
	}

	L.GetGlobal("_tree1")
	vnode1 := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if findTextContent(vnode1, "count:1") == "" {
		t.Fatalf("first render should show count:1, got: %s", describeVNode(vnode1))
	}

	// Second render with SAME props — should reuse cached VNode
	err = L.DoString(`
		_tree2 = {
			type = "box",
			children = {
				lumina.createElement(_MemoCounter, { value = 1 }),
			}
		}
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}

	L.GetGlobal("_tree2")
	vnode2 := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Should still show count:1 (not count:2) because render was skipped
	if findTextContent(vnode2, "count:1") == "" {
		t.Fatalf("memo should skip render on same props, got: %s", describeVNode(vnode2))
	}

	// Verify render count is still 1
	L.GetGlobal("_render_count")
	rc, _ := L.ToInteger(-1)
	L.Pop(1)
	if rc != 1 {
		t.Fatalf("expected render_count=1 (memo skipped), got %d", rc)
	}
}

func TestMemo_RerendersOnChangedProps(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	err := L.DoString(`
		_render_count = 0
		local Counter = lumina.defineComponent({
			name = "MemoCounter2",
			render = function(self)
				_render_count = _render_count + 1
				return { type = "text", content = "v:" .. (self.value or "?") }
			end
		})
		_MemoCounter = lumina.memo(Counter)
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// First render
	err = L.DoString(`
		_tree1 = { type = "box", children = { lumina.createElement(_MemoCounter, { value = "A" }) } }
	`)
	if err != nil {
		t.Fatalf("tree1: %v", err)
	}
	L.GetGlobal("_tree1")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Second render with DIFFERENT props — should re-render
	err = L.DoString(`
		_tree2 = { type = "box", children = { lumina.createElement(_MemoCounter, { value = "B" }) } }
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}
	L.GetGlobal("_tree2")
	vnode2 := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if findTextContent(vnode2, "v:B") == "" {
		t.Fatalf("memo should re-render on changed props, got: %s", describeVNode(vnode2))
	}

	L.GetGlobal("_render_count")
	rc, _ := L.ToInteger(-1)
	L.Pop(1)
	if rc != 2 {
		t.Fatalf("expected render_count=2, got %d", rc)
	}
}

func TestMemo_NonMemoAlwaysRerenders(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	err := L.DoString(`
		_render_count = 0
		local Counter = lumina.defineComponent({
			name = "PlainCounter",
			render = function(self)
				_render_count = _render_count + 1
				return { type = "text", content = "n:" .. _render_count }
			end
		})
		_Counter = Counter  -- NOT wrapped in memo
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// First render
	err = L.DoString(`
		_tree1 = { type = "box", children = { lumina.createElement(_Counter, { x = 1 }) } }
	`)
	if err != nil {
		t.Fatalf("tree1: %v", err)
	}
	L.GetGlobal("_tree1")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Second render with same props — should still re-render (no memo)
	err = L.DoString(`
		_tree2 = { type = "box", children = { lumina.createElement(_Counter, { x = 1 }) } }
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}
	L.GetGlobal("_tree2")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_render_count")
	rc, _ := L.ToInteger(-1)
	L.Pop(1)
	if rc < 2 {
		t.Fatalf("non-memo component should always re-render, got render_count=%d", rc)
	}
}

func TestMemo_ComplexPropsComparison(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "memo_complex_parent",
		Type:  "MemoComplexParent",
		Name:  "MemoComplexParent",
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
		_render_count = 0
		local C = lumina.defineComponent({
			name = "MemoComplex",
			render = function(self)
				_render_count = _render_count + 1
				return { type = "text", content = "r:" .. _render_count }
			end
		})
		_MC = lumina.memo(C)
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Render with multiple props
	err = L.DoString(`
		_tree1 = { type = "box", children = { lumina.createElement(_MC, { a = 1, b = "hello", c = true }) } }
	`)
	if err != nil {
		t.Fatalf("tree1: %v", err)
	}
	L.GetGlobal("_tree1")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Same props → skip
	err = L.DoString(`
		_tree2 = { type = "box", children = { lumina.createElement(_MC, { a = 1, b = "hello", c = true }) } }
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}
	L.GetGlobal("_tree2")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_render_count")
	rc, _ := L.ToInteger(-1)
	L.Pop(1)
	if rc != 1 {
		t.Fatalf("expected render_count=1 with same complex props, got %d", rc)
	}
}

// ============================================================
// Phase 13F: Event Bubbling
// ============================================================

func TestEventBubbling_BubblesFromChildToParent(t *testing.T) {
	// Build a VNode tree
	root := NewVNode("box")
	root.Props["id"] = "root"

	parent := NewVNode("box")
	parent.Props["id"] = "parent"

	child := NewVNode("box")
	child.Props["id"] = "child"

	parent.AddChild(child)
	root.AddChild(parent)

	tree := BuildVNodeTree(root)

	// Set up event bus with the tree
	eb := &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
		vnodeTree: tree,
	}

	parentClicked := false
	eb.On("click", "parent", func(e *Event) {
		parentClicked = true
	})

	// Emit click on child — should bubble to parent
	eb.Emit(&Event{Type: "click", Target: "child"})

	if !parentClicked {
		t.Fatal("expected click to bubble from child to parent")
	}
}

func TestEventBubbling_StopPropagation(t *testing.T) {
	root := NewVNode("box")
	root.Props["id"] = "root"

	parent := NewVNode("box")
	parent.Props["id"] = "parent"

	child := NewVNode("box")
	child.Props["id"] = "child"

	middle := NewVNode("box")
	middle.Props["id"] = "middle"

	parent.AddChild(middle)
	middle.AddChild(child)
	root.AddChild(parent)

	tree := BuildVNodeTree(root)

	eb := &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
		vnodeTree: tree,
	}

	middleClicked := false
	parentClicked := false

	eb.On("click", "middle", func(e *Event) {
		middleClicked = true
		e.StopPropagation()
	})
	eb.On("click", "parent", func(e *Event) {
		parentClicked = true
	})

	eb.Emit(&Event{Type: "click", Target: "child"})

	if !middleClicked {
		t.Fatal("expected middle to receive bubbled click")
	}
	if parentClicked {
		t.Fatal("expected StopPropagation to prevent parent from receiving click")
	}
}

func TestEventBubbling_DirectHandlerFiresFirst(t *testing.T) {
	root := NewVNode("box")
	root.Props["id"] = "root"

	child := NewVNode("box")
	child.Props["id"] = "child"

	root.AddChild(child)

	tree := BuildVNodeTree(root)

	eb := &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
		vnodeTree: tree,
	}

	order := []string{}

	eb.On("click", "child", func(e *Event) {
		order = append(order, "direct")
	})
	eb.On("click", "root", func(e *Event) {
		order = append(order, "bubbled")
	})

	eb.Emit(&Event{Type: "click", Target: "child"})

	if len(order) != 2 {
		t.Fatalf("expected 2 handlers fired, got %d", len(order))
	}
	if order[0] != "direct" || order[1] != "bubbled" {
		t.Fatalf("expected [direct, bubbled], got %v", order)
	}
}

func TestEventBubbling_NoHandlerOnIntermediate(t *testing.T) {
	root := NewVNode("box")
	root.Props["id"] = "root"

	middle := NewVNode("box")
	middle.Props["id"] = "middle"
	// No handler on middle

	child := NewVNode("box")
	child.Props["id"] = "child"

	root.AddChild(middle)
	middle.AddChild(child)

	tree := BuildVNodeTree(root)

	eb := &EventBus{
		handlers:  make(map[string][]eventHandler),
		shortcuts: make(map[string]eventHandler),
		vnodeTree: tree,
	}

	rootClicked := false
	eb.On("click", "root", func(e *Event) {
		rootClicked = true
	})

	eb.Emit(&Event{Type: "click", Target: "child"})

	if !rootClicked {
		t.Fatal("expected click to bubble through middle (no handler) to root")
	}
}

// ============================================================
// Phase 13G: Context Tree
// ============================================================

func TestContextTree_ChildReadsFromParent(t *testing.T) {
	L := advTestState(t)
	defer L.Close()
	defer ClearContextValues()

	// Create parent component that provides a context value
	parent := &Component{
		ID:            "ctx_parent",
		Type:          "CtxParent",
		Name:          "CtxParent",
		Props:         make(map[string]any),
		State:         make(map[string]any),
		ContextValues: make(map[int64]any),
	}
	child := &Component{
		ID:    "ctx_child",
		Type:  "CtxChild",
		Name:  "CtxChild",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	parent.AddChild(child)

	globalRegistry.mu.Lock()
	globalRegistry.components[parent.ID] = parent
	globalRegistry.components[child.ID] = child
	globalRegistry.mu.Unlock()
	defer func() {
		globalRegistry.mu.Lock()
		delete(globalRegistry.components, parent.ID)
		delete(globalRegistry.components, child.ID)
		globalRegistry.mu.Unlock()
	}()

	// Create a context
	ctx := NewContext("default")

	// Set value on parent
	parent.ContextValues[ctx.ID] = "from_parent"

	// Resolve from child should find parent's value
	val, found := resolveContextFromTree(child, ctx.ID)
	if !found {
		t.Fatal("expected to find context value from parent")
	}
	if val != "from_parent" {
		t.Fatalf("expected 'from_parent', got %v", val)
	}
}

func TestContextTree_NestedProvidersInnerOverrides(t *testing.T) {
	grandparent := &Component{
		ID:            "gp",
		Type:          "GP",
		Name:          "GP",
		Props:         make(map[string]any),
		State:         make(map[string]any),
		ContextValues: make(map[int64]any),
	}
	parent := &Component{
		ID:            "p",
		Type:          "P",
		Name:          "P",
		Props:         make(map[string]any),
		State:         make(map[string]any),
		ContextValues: make(map[int64]any),
	}
	child := &Component{
		ID:    "c",
		Type:  "C",
		Name:  "C",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	grandparent.AddChild(parent)
	parent.AddChild(child)

	ctx := NewContext("default")

	// Both grandparent and parent provide values
	grandparent.ContextValues[ctx.ID] = "gp_value"
	parent.ContextValues[ctx.ID] = "parent_value"

	// Child should see parent's value (nearest provider)
	val, found := resolveContextFromTree(child, ctx.ID)
	if !found {
		t.Fatal("expected to find context value")
	}
	if val != "parent_value" {
		t.Fatalf("expected 'parent_value' (inner overrides), got %v", val)
	}
}

func TestContextTree_NoProviderUsesDefault(t *testing.T) {
	child := &Component{
		ID:    "orphan",
		Type:  "Orphan",
		Name:  "Orphan",
		Props: make(map[string]any),
		State: make(map[string]any),
	}

	ctx := NewContext("the_default")

	// No provider in tree — should not find
	val, found := resolveContextFromTree(child, ctx.ID)
	if found {
		t.Fatalf("expected not found (no provider), got %v", val)
	}

	// GetContextValue should return default
	globalVal := GetContextValue(ctx)
	if globalVal != "the_default" {
		t.Fatalf("expected default 'the_default', got %v", globalVal)
	}
}

func TestContextTree_UseContextViaLua(t *testing.T) {
	L := advTestState(t)
	defer L.Close()
	defer ClearContextValues()

	// Set up a component with context
	comp := &Component{
		ID:            "lua_ctx_comp",
		Type:          "LuaCtxComp",
		Name:          "LuaCtxComp",
		Props:         make(map[string]any),
		State:         make(map[string]any),
		ContextValues: make(map[int64]any),
	}
	globalRegistry.mu.Lock()
	globalRegistry.components[comp.ID] = comp
	globalRegistry.mu.Unlock()
	SetCurrentComponent(comp)
	defer func() {
		SetCurrentComponent(nil)
		globalRegistry.mu.Lock()
		delete(globalRegistry.components, comp.ID)
		globalRegistry.mu.Unlock()
	}()

	// Create context and set value on comp
	err := L.DoString(`
		_ctx = lumina.createContext("fallback")
		lumina.setContextValue(_ctx, "tree_value")
		_val = lumina.useContext(_ctx)
	`)
	if err != nil {
		t.Fatalf("context via Lua: %v", err)
	}

	L.GetGlobal("_val")
	v, ok := L.ToString(-1)
	L.Pop(1)
	if !ok || v != "tree_value" {
		t.Fatalf("expected 'tree_value', got %q", v)
	}
}

// ============================================================
// Phase 13H: Portals + forwardRef
// ============================================================

func TestPortal_CreatesPortalVNode(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	err := L.DoString(`
		_portal = lumina.createPortal(
			{ type = "text", content = "modal content" },
			"root"
		)
	`)
	if err != nil {
		t.Fatalf("createPortal: %v", err)
	}

	L.GetGlobal("_portal")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if !vnode.IsPortal {
		t.Fatal("expected IsPortal=true")
	}
	if vnode.PortalTarget != "root" {
		t.Fatalf("expected PortalTarget='root', got %q", vnode.PortalTarget)
	}
	// Portal content should be the text node
	if vnode.Type != "text" || vnode.Content != "modal content" {
		t.Fatalf("expected portal content 'modal content', got type=%q content=%q", vnode.Type, vnode.Content)
	}
}

func TestForwardRef_PassesRefToChild(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	err := L.DoString(`
		local FancyInput = lumina.forwardRef(function(props, ref)
			return { type = "input", id = ref, value = props.value or "" }
		end)

		_tree = {
			type = "box",
			children = {
				lumina.createElement(FancyInput, { ref = "my-input", value = "hello" }),
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
	if child.Type != "input" {
		t.Fatalf("expected type='input', got %q", child.Type)
	}
	if id, ok := child.Props["id"].(string); !ok || id != "my-input" {
		t.Fatalf("expected id='my-input', got %v", child.Props["id"])
	}
}

func TestForwardRef_ParentAccessesChildViaRef(t *testing.T) {
	L := advTestState(t)
	defer L.Close()

	err := L.DoString(`
		local FancyInput = lumina.forwardRef(function(props, ref)
			return { type = "input", id = ref, value = props.value or "" }
		end)

		-- Parent uses a ref to identify the child
		_tree = {
			type = "vbox",
			children = {
				lumina.createElement(FancyInput, { ref = "email-input", value = "test@example.com" }),
				lumina.createElement(FancyInput, { ref = "name-input", value = "Alice" }),
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

	// First child should have id="email-input"
	child0 := vnode.Children[0]
	if id, ok := child0.Props["id"].(string); !ok || id != "email-input" {
		t.Fatalf("child[0]: expected id='email-input', got %v", child0.Props["id"])
	}

	// Second child should have id="name-input"
	child1 := vnode.Children[1]
	if id, ok := child1.Props["id"].(string); !ok || id != "name-input" {
		t.Fatalf("child[1]: expected id='name-input', got %v", child1.Props["id"])
	}
}

// ============================================================
// Test Helpers
// ============================================================

// findTextContent searches a VNode tree for a text node containing the given substring.
// Returns the full content if found, empty string if not.
func findTextContent(vnode *VNode, substr string) string {
	if vnode == nil {
		return ""
	}
	if vnode.Content != "" && strings.Contains(vnode.Content, substr) {
		return vnode.Content
	}
	for _, child := range vnode.Children {
		if found := findTextContent(child, substr); found != "" {
			return found
		}
	}
	return ""
}

// describeVNode returns a debug string for a VNode tree.
func describeVNode(vnode *VNode) string {
	if vnode == nil {
		return "<nil>"
	}
	s := vnode.Type
	if vnode.Content != "" {
		s += "(" + vnode.Content + ")"
	}
	if len(vnode.Children) > 0 {
		s += "["
		for i, child := range vnode.Children {
			if i > 0 {
				s += ", "
			}
			s += describeVNode(child)
		}
		s += "]"
	}
	return s
}
