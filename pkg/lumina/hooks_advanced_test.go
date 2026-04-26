package lumina

import (
	"strings"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

func hooksTestState(t *testing.T) *lua.State {
	t.Helper()
	ClearComponents()
	ClearContextValues()
	L := lua.NewState()
	Open(L)
	return L
}

// ============================================================
// Fragment Tests
// ============================================================

func TestFragment_RendersChildrenWithoutWrapper(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	err := L.DoString(`
		_tree = {
			type = "fragment",
			children = {
				{ type = "text", content = "Line 1" },
				{ type = "text", content = "Line 2" },
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if vnode.Type != "fragment" {
		t.Fatalf("expected type='fragment', got %q", vnode.Type)
	}
	if len(vnode.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(vnode.Children))
	}
	if vnode.Children[0].Content != "Line 1" {
		t.Fatalf("expected 'Line 1', got %q", vnode.Children[0].Content)
	}
	if vnode.Children[1].Content != "Line 2" {
		t.Fatalf("expected 'Line 2', got %q", vnode.Children[1].Content)
	}
}

func TestFragment_InsideVBox_ChildrenFlowAsSiblings(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	err := L.DoString(`
		_tree = {
			type = "vbox",
			children = {
				{ type = "text", content = "Before" },
				{
					type = "fragment",
					children = {
						{ type = "text", content = "Frag1" },
						{ type = "text", content = "Frag2" },
					}
				},
				{ type = "text", content = "After" },
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// vbox should have 3 children: text, fragment, text
	if len(vnode.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(vnode.Children))
	}

	// Fragment should have 2 children
	frag := vnode.Children[1]
	if frag.Type != "fragment" {
		t.Fatalf("expected fragment, got %q", frag.Type)
	}
	if len(frag.Children) != 2 {
		t.Fatalf("expected 2 fragment children, got %d", len(frag.Children))
	}

	// Render to frame — fragment children should be laid out
	frame := VNodeToFrame(vnode, 40, 10)
	content := frameToString(frame)
	if !strings.Contains(content, "Before") {
		t.Fatalf("expected 'Before' in frame")
	}
	if !strings.Contains(content, "Frag1") {
		t.Fatalf("expected 'Frag1' in frame")
	}
	if !strings.Contains(content, "Frag2") {
		t.Fatalf("expected 'Frag2' in frame")
	}
}

func TestFragment_Nested(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	err := L.DoString(`
		_tree = {
			type = "fragment",
			children = {
				{
					type = "fragment",
					children = {
						{ type = "text", content = "Deep1" },
						{ type = "text", content = "Deep2" },
					}
				},
				{ type = "text", content = "Outer" },
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	if vnode.Type != "fragment" {
		t.Fatalf("expected fragment, got %q", vnode.Type)
	}
	if len(vnode.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(vnode.Children))
	}
	inner := vnode.Children[0]
	if inner.Type != "fragment" {
		t.Fatalf("expected inner fragment, got %q", inner.Type)
	}
	if len(inner.Children) != 2 {
		t.Fatalf("expected 2 inner children, got %d", len(inner.Children))
	}
}

// ============================================================
// useLayoutEffect Tests
// ============================================================

func TestUseLayoutEffect_RunsAfterRender(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "le_parent",
		Type:  "LEParent",
		Name:  "LEParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		Comp = lumina.defineComponent({
			name = "LayoutEffectComp",
			render = function(self)
				_effect_ran = false
				lumina.useLayoutEffect(function()
					_effect_ran = true
				end, {})
				return { type = "text", content = "hello" }
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

	L.GetGlobal("_effect_ran")
	ran := L.ToBoolean(-1)
	L.Pop(1)

	if !ran {
		t.Fatal("expected useLayoutEffect to have run during render")
	}
}

func TestUseLayoutEffect_CleanupRunsBeforeNext(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "le_cleanup_parent",
		Type:  "LECleanupParent",
		Name:  "LECleanupParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	// Use counters instead of table.insert (simpler Lua, avoids potential table issues)
	err := L.DoString(`
		_effect_count = 0
		_cleanup_count = 0
		Comp = lumina.defineComponent({
			name = "LECleanupComp",
			render = function(self)
				lumina.useLayoutEffect(function()
					_effect_count = _effect_count + 1
					return function()
						_cleanup_count = _cleanup_count + 1
					end
				end) -- no deps = runs every render
				return { type = "text", content = "hi" }
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

	// After first render: effect ran once, no cleanup yet
	L.GetGlobal("_effect_count")
	ec1, _ := L.ToInteger(-1)
	L.Pop(1)
	L.GetGlobal("_cleanup_count")
	cc1, _ := L.ToInteger(-1)
	L.Pop(1)

	if ec1 != 1 {
		t.Fatalf("after render 1: expected effect_count=1, got %d", ec1)
	}
	if cc1 != 0 {
		t.Fatalf("after render 1: expected cleanup_count=0, got %d", cc1)
	}

	// Re-establish parent context for second render
	SetCurrentComponent(parent)

	err = L.DoString(`
		_tree2 = { type = "box", children = { lumina.createElement(Comp, { v = 2 }) } }
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}
	L.GetGlobal("_tree2")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// After second render: cleanup from first ran, then new effect ran
	L.GetGlobal("_effect_count")
	ec2, _ := L.ToInteger(-1)
	L.Pop(1)
	L.GetGlobal("_cleanup_count")
	cc2, _ := L.ToInteger(-1)
	L.Pop(1)

	if ec2 != 2 {
		t.Fatalf("after render 2: expected effect_count=2, got %d", ec2)
	}
	if cc2 != 1 {
		t.Fatalf("after render 2: expected cleanup_count=1 (cleanup ran before re-effect), got %d", cc2)
	}
}

func TestUseLayoutEffect_SkipsWhenDepsUnchanged(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "le_deps_parent",
		Type:  "LEDepsParent",
		Name:  "LEDepsParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		_run_count = 0
		Comp = lumina.defineComponent({
			name = "LEDepsComp",
			render = function(self)
				lumina.useLayoutEffect(function()
					_run_count = _run_count + 1
				end, { "stable" })
				return { type = "text", content = "hi" }
			end
		})

		_tree = { type = "box", children = { lumina.createElement(Comp, {}) } }
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// First render — should run
	L.GetGlobal("_tree")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Second render with same deps — should skip
	err = L.DoString(`
		_tree2 = { type = "box", children = { lumina.createElement(Comp, { v = 2 }) } }
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}
	L.GetGlobal("_tree2")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_run_count")
	rc, _ := L.ToInteger(-1)
	L.Pop(1)
	if rc != 1 {
		t.Fatalf("expected run_count=1 (deps unchanged), got %d", rc)
	}
}

// ============================================================
// useId Tests
// ============================================================

func TestUseId_ReturnsStableId(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "id_parent",
		Type:  "IdParent",
		Name:  "IdParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		Comp = lumina.defineComponent({
			name = "IdComp",
			render = function(self)
				_id1 = lumina.useId()
				return { type = "text", content = _id1 }
			end
		})

		_tree = { type = "box", children = { lumina.createElement(Comp, {}) } }
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// First render
	L.GetGlobal("_tree")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_id1")
	id1, _ := L.ToString(-1)
	L.Pop(1)

	if id1 == "" {
		t.Fatal("expected non-empty ID")
	}

	// Second render — ID should be the same
	err = L.DoString(`
		_tree2 = { type = "box", children = { lumina.createElement(Comp, { v = 2 }) } }
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}
	L.GetGlobal("_tree2")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_id1")
	id2, _ := L.ToString(-1)
	L.Pop(1)

	if id1 != id2 {
		t.Fatalf("expected stable ID across renders: %q != %q", id1, id2)
	}
}

func TestUseId_MultipleCallsReturnDifferentIds(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "multi_id_parent",
		Type:  "MultiIdParent",
		Name:  "MultiIdParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		Comp = lumina.defineComponent({
			name = "MultiIdComp",
			render = function(self)
				_idA = lumina.useId()
				_idB = lumina.useId()
				return { type = "text", content = _idA .. " " .. _idB }
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

	L.GetGlobal("_idA")
	idA, _ := L.ToString(-1)
	L.Pop(1)

	L.GetGlobal("_idB")
	idB, _ := L.ToString(-1)
	L.Pop(1)

	if idA == idB {
		t.Fatalf("expected different IDs, both are %q", idA)
	}
}

func TestUseId_DifferentComponentsDifferentPrefixes(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "diff_id_parent",
		Type:  "DiffIdParent",
		Name:  "DiffIdParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		local CompA = lumina.defineComponent({
			name = "CompA",
			render = function(self)
				_idA = lumina.useId()
				return { type = "text", content = _idA }
			end
		})
		local CompB = lumina.defineComponent({
			name = "CompB",
			render = function(self)
				_idB = lumina.useId()
				return { type = "text", content = _idB }
			end
		})

		_tree = {
			type = "box",
			children = {
				lumina.createElement(CompA, {}),
				lumina.createElement(CompB, {}),
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_idA")
	idA, _ := L.ToString(-1)
	L.Pop(1)

	L.GetGlobal("_idB")
	idB, _ := L.ToString(-1)
	L.Pop(1)

	if idA == idB {
		t.Fatalf("expected different IDs for different components, both are %q", idA)
	}
}

// ============================================================
// useImperativeHandle Tests
// ============================================================

func TestUseImperativeHandle_SetsRefCurrent(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "ih_parent",
		Type:  "IHParent",
		Name:  "IHParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		Comp = lumina.defineComponent({
			name = "IHComp",
			render = function(self)
				local ref = { current = nil }
				lumina.useImperativeHandle(ref, function()
					return {
						greet = function() return "hello" end,
						value = 42,
					}
				end, {})
				_ref = ref
				return { type = "text", content = "ih" }
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

	// Check ref.current.value
	err = L.DoString(`
		_handle_value = _ref.current.value
	`)
	if err != nil {
		t.Fatalf("check handle: %v", err)
	}

	L.GetGlobal("_handle_value")
	v, _ := L.ToInteger(-1)
	L.Pop(1)
	if v != 42 {
		t.Fatalf("expected handle value=42, got %d", v)
	}
}

func TestUseImperativeHandle_UpdatesOnDepsChange(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "ih_deps_parent",
		Type:  "IHDepsParent",
		Name:  "IHDepsParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		_create_count = 0
		_ref = { current = nil }
		Comp = lumina.defineComponent({
			name = "IHDepsComp",
			render = function(self)
				lumina.useImperativeHandle(_ref, function()
					_create_count = _create_count + 1
					return { count = _create_count }
				end, { self.dep or "default" })
				return { type = "text", content = "ih" }
			end
		})

		-- First render
		_tree = { type = "box", children = { lumina.createElement(Comp, {}) } }
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Same deps — should not re-create
	err = L.DoString(`
		_tree2 = { type = "box", children = { lumina.createElement(Comp, {}) } }
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}
	L.GetGlobal("_tree2")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_create_count")
	cc, _ := L.ToInteger(-1)
	L.Pop(1)
	if cc != 1 {
		t.Fatalf("expected create_count=1 (deps unchanged), got %d", cc)
	}

	// Different deps — should re-create
	err = L.DoString(`
		_tree3 = { type = "box", children = { lumina.createElement(Comp, { dep = "changed" }) } }
	`)
	if err != nil {
		t.Fatalf("tree3: %v", err)
	}
	L.GetGlobal("_tree3")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_create_count")
	cc2, _ := L.ToInteger(-1)
	L.Pop(1)
	if cc2 != 2 {
		t.Fatalf("expected create_count=2 (deps changed), got %d", cc2)
	}
}

// ============================================================
// Suspense + lazy Tests
// ============================================================

func TestSuspense_ShowsFallbackForPendingLazy(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "suspense_parent",
		Type:  "SuspenseParent",
		Name:  "SuspenseParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	// Create a Suspense boundary with a lazy child that has _lazy_status = "pending"
	err := L.DoString(`
		-- Simulate a pending lazy element
		_tree = {
			type = "box",
			children = {
				lumina.createElement(lumina.Suspense, {
					fallback = { type = "text", content = "Loading..." },
					children = {
						{ type = "box", _lazy_status = "pending" },
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

	found := findTextInVNode(vnode, "Loading...")
	if !found {
		t.Fatalf("expected 'Loading...' fallback, got: %s", describeVNodeTree(vnode))
	}
}

func TestSuspense_RendersComponentAfterLoad(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "suspense_loaded_parent",
		Type:  "SuspenseLoadedParent",
		Name:  "SuspenseLoadedParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		-- lazy that resolves immediately
		local HeavyComp = lumina.lazy(function()
			return lumina.defineComponent({
				name = "Heavy",
				render = function(self)
					return { type = "text", content = "Heavy loaded!" }
				end
			})
		end)

		_tree = {
			type = "box",
			children = {
				lumina.createElement(HeavyComp, {}),
			}
		}
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	vnode := LuaVNodeToVNode(L, -1)
	L.Pop(1)

	found := findTextInVNode(vnode, "Heavy loaded!")
	if !found {
		t.Fatalf("expected 'Heavy loaded!' after lazy resolve, got: %s", describeVNodeTree(vnode))
	}
}

func TestSuspense_NestedBoundaries(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "nested_suspense_parent",
		Type:  "NestedSuspenseParent",
		Name:  "NestedSuspenseParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		local Ready = lumina.defineComponent({
			name = "Ready",
			render = function(self)
				return { type = "text", content = "Ready!" }
			end
		})

		-- Outer Suspense wraps a ready component
		-- Inner Suspense wraps a pending component
		_tree = {
			type = "box",
			children = {
				lumina.createElement(lumina.Suspense, {
					fallback = { type = "text", content = "Outer loading..." },
					children = {
						lumina.createElement(Ready, {}),
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

	// Should show Ready (no pending children)
	found := findTextInVNode(vnode, "Ready!")
	if !found {
		t.Fatalf("expected 'Ready!' (no pending), got: %s", describeVNodeTree(vnode))
	}
	// Should NOT show fallback
	fallback := findTextInVNode(vnode, "Outer loading...")
	if fallback {
		t.Fatal("should not show fallback when no pending children")
	}
}

// ============================================================
// useTransition Tests
// ============================================================

func TestUseTransition_IsPendingDuringTransition(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "transition_parent",
		Type:  "TransitionParent",
		Name:  "TransitionParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		_was_pending = false
		Comp = lumina.defineComponent({
			name = "TransComp",
			render = function(self)
				local isPending, startTransition = lumina.useTransition()
				_isPending_before = isPending

				startTransition(function()
					_was_pending = true
					-- In real use, setState calls would go here
				end)

				return { type = "text", content = "trans" }
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

	L.GetGlobal("_was_pending")
	wp := L.ToBoolean(-1)
	L.Pop(1)

	if !wp {
		t.Fatal("expected transition callback to have been called")
	}
}

func TestUseTransition_IsPendingFalseAfterComplete(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "trans_done_parent",
		Type:  "TransDoneParent",
		Name:  "TransDoneParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		Comp = lumina.defineComponent({
			name = "TransDoneComp",
			render = function(self)
				local isPending, startTransition = lumina.useTransition()
				_isPending = isPending

				if not _started then
					_started = true
					startTransition(function()
						-- transition work
					end)
				end

				return { type = "text", content = "done" }
			end
		})

		_tree = { type = "box", children = { lumina.createElement(Comp, {}) } }
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// First render — starts and completes transition
	L.GetGlobal("_tree")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Second render — isPending should be false
	err = L.DoString(`
		_tree2 = { type = "box", children = { lumina.createElement(Comp, { v = 2 }) } }
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}
	L.GetGlobal("_tree2")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_isPending")
	ip := L.ToBoolean(-1)
	L.Pop(1)

	if ip {
		t.Fatal("expected isPending=false after transition completes")
	}
}

// ============================================================
// useDeferredValue Tests
// ============================================================

func TestUseDeferredValue_LagsBehindCurrentValue(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "deferred_parent",
		Type:  "DeferredParent",
		Name:  "DeferredParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		Comp = lumina.defineComponent({
			name = "DeferredComp",
			render = function(self)
				local val = self.value or "initial"
				_deferred = lumina.useDeferredValue(val)
				return { type = "text", content = "d:" .. _deferred }
			end
		})

		-- First render with "A"
		_tree = { type = "box", children = { lumina.createElement(Comp, { value = "A" }) } }
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_deferred")
	d1, _ := L.ToString(-1)
	L.Pop(1)

	// First render: deferred = "A" (no previous value)
	if d1 != "A" {
		t.Fatalf("first render: expected deferred='A', got %q", d1)
	}

	// Second render with "B"
	err = L.DoString(`
		_tree2 = { type = "box", children = { lumina.createElement(Comp, { value = "B" }) } }
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}
	L.GetGlobal("_tree2")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_deferred")
	d2, _ := L.ToString(-1)
	L.Pop(1)

	// Second render: deferred should lag — return "A" (previous value)
	if d2 != "A" {
		t.Fatalf("second render: expected deferred='A' (lagging), got %q", d2)
	}
}

func TestUseDeferredValue_CatchesUp(t *testing.T) {
	L := hooksTestState(t)
	defer L.Close()

	parent := &Component{
		ID:    "deferred_catchup_parent",
		Type:  "DeferredCatchupParent",
		Name:  "DeferredCatchupParent",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.components[parent.ID] = parent
	SetCurrentComponent(parent)
	defer func() {
		SetCurrentComponent(nil)
		delete(globalRegistry.components, parent.ID)
	}()

	err := L.DoString(`
		Comp = lumina.defineComponent({
			name = "DeferredCatchupComp",
			render = function(self)
				local val = self.value or "initial"
				_deferred = lumina.useDeferredValue(val)
				return { type = "text", content = "d:" .. _deferred }
			end
		})

		-- Render 1: "A"
		_tree1 = { type = "box", children = { lumina.createElement(Comp, { value = "A" }) } }
	`)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	L.GetGlobal("_tree1")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Render 2: "B" → deferred still "A"
	err = L.DoString(`
		_tree2 = { type = "box", children = { lumina.createElement(Comp, { value = "B" }) } }
	`)
	if err != nil {
		t.Fatalf("tree2: %v", err)
	}
	L.GetGlobal("_tree2")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	// Render 3: "B" again → deferred catches up to "B"
	err = L.DoString(`
		_tree3 = { type = "box", children = { lumina.createElement(Comp, { value = "B" }) } }
	`)
	if err != nil {
		t.Fatalf("tree3: %v", err)
	}
	L.GetGlobal("_tree3")
	LuaVNodeToVNode(L, -1)
	L.Pop(1)

	L.GetGlobal("_deferred")
	d3, _ := L.ToString(-1)
	L.Pop(1)

	if d3 != "B" {
		t.Fatalf("third render: expected deferred='B' (caught up), got %q", d3)
	}
}

// ============================================================
// Test Helpers
// ============================================================

func findTextInVNode(vnode *VNode, substr string) bool {
	if vnode == nil {
		return false
	}
	if vnode.Content != "" && strings.Contains(vnode.Content, substr) {
		return true
	}
	for _, child := range vnode.Children {
		if findTextInVNode(child, substr) {
			return true
		}
	}
	return false
}

func describeVNodeTree(vnode *VNode) string {
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
			s += describeVNodeTree(child)
		}
		s += "]"
	}
	return s
}

// frameToString extracts all non-space characters from a frame.
func frameToString(frame *Frame) string {
	var sb strings.Builder
	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			ch := frame.Cells[y][x].Char
			sb.WriteRune(ch)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}
