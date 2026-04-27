package bridge

import (
	"strings"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/animation"
	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
	"github.com/akzj/lumina/pkg/lumina/v2/event"
	"github.com/akzj/lumina/pkg/lumina/v2/layout"
	"github.com/akzj/lumina/pkg/lumina/v2/paint"
	"github.com/akzj/lumina/pkg/lumina/v2/router"
)

// newTestBridge creates a Bridge with a fresh Lua state for testing.
func newTestBridge(t *testing.T) *Bridge {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	return NewBridge(L)
}

// newHookTestComponent creates a minimal component for hook tests.
func newHookTestComponent(id string) *component.Component {
	nopRender := func(state, props map[string]any) *layout.VNode { return nil }
	return component.NewComponent(id, id, buffer.Rect{W: 1, H: 1}, 0, nopRender)
}

// --- LuaTableToVNode tests ---

func TestBridge_LuaTableToVNode_BasicBox(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	// Push a Lua table: { type = "box", id = "root" }
	err := L.DoString(`return { type = "box", id = "root" }`)
	if err != nil {
		t.Fatal(err)
	}

	vn := b.LuaTableToVNode(-1)
	L.Pop(1)

	if vn.Type != "box" {
		t.Errorf("Type = %q, want %q", vn.Type, "box")
	}
	if vn.ID != "root" {
		t.Errorf("ID = %q, want %q", vn.ID, "root")
	}
}

func TestBridge_LuaTableToVNode_Text(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	err := L.DoString(`return { type = "text", content = "hello world" }`)
	if err != nil {
		t.Fatal(err)
	}

	vn := b.LuaTableToVNode(-1)
	L.Pop(1)

	if vn.Type != "text" {
		t.Errorf("Type = %q, want %q", vn.Type, "text")
	}
	if vn.Content != "hello world" {
		t.Errorf("Content = %q, want %q", vn.Content, "hello world")
	}
}

func TestBridge_LuaTableToVNode_DefaultType(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	// Table without type field → default to "box".
	err := L.DoString(`return { id = "no-type" }`)
	if err != nil {
		t.Fatal(err)
	}

	vn := b.LuaTableToVNode(-1)
	L.Pop(1)

	if vn.Type != "box" {
		t.Errorf("Type = %q, want %q", vn.Type, "box")
	}
}

func TestBridge_LuaTableToVNode_Children(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	err := L.DoString(`return {
		type = "vbox",
		id = "parent",
		children = {
			{ type = "text", content = "child1" },
			{ type = "text", content = "child2" },
			{ type = "box", id = "inner", children = {
				{ type = "text", content = "nested" }
			}}
		}
	}`)
	if err != nil {
		t.Fatal(err)
	}

	vn := b.LuaTableToVNode(-1)
	L.Pop(1)

	if vn.Type != "vbox" {
		t.Errorf("Type = %q, want %q", vn.Type, "vbox")
	}
	if len(vn.Children) != 3 {
		t.Fatalf("len(Children) = %d, want 3", len(vn.Children))
	}
	if vn.Children[0].Content != "child1" {
		t.Errorf("Children[0].Content = %q, want %q", vn.Children[0].Content, "child1")
	}
	if vn.Children[1].Content != "child2" {
		t.Errorf("Children[1].Content = %q, want %q", vn.Children[1].Content, "child2")
	}
	// Nested child.
	inner := vn.Children[2]
	if inner.ID != "inner" {
		t.Errorf("inner.ID = %q, want %q", inner.ID, "inner")
	}
	if len(inner.Children) != 1 {
		t.Fatalf("inner children = %d, want 1", len(inner.Children))
	}
	if inner.Children[0].Content != "nested" {
		t.Errorf("nested.Content = %q, want %q", inner.Children[0].Content, "nested")
	}
}

func TestBridge_LuaTableToVNode_Style(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	err := L.DoString(`return {
		type = "box",
		style = {
			width = 100,
			height = 50,
			flex = 1,
			padding = 2,
			paddingTop = 3,
			margin = 1,
			gap = 4,
			justify = "center",
			align = "end",
			border = "single",
			foreground = "red",
			background = "blue",
			bold = true,
			underline = true,
			overflow = "hidden",
			position = "absolute",
			top = 10,
			left = 20,
			zIndex = 5,
		}
	}`)
	if err != nil {
		t.Fatal(err)
	}

	vn := b.LuaTableToVNode(-1)
	L.Pop(1)
	s := vn.Style

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"Width", s.Width, 100},
		{"Height", s.Height, 50},
		{"Flex", s.Flex, 1},
		{"Padding", s.Padding, 2},
		{"PaddingTop", s.PaddingTop, 3},
		{"Margin", s.Margin, 1},
		{"Gap", s.Gap, 4},
		{"Justify", s.Justify, "center"},
		{"Align", s.Align, "end"},
		{"Border", s.Border, "single"},
		{"Foreground", s.Foreground, "red"},
		{"Background", s.Background, "blue"},
		{"Bold", s.Bold, true},
		{"Underline", s.Underline, true},
		{"Overflow", s.Overflow, "hidden"},
		{"Position", s.Position, "absolute"},
		{"Top", s.Top, 10},
		{"Left", s.Left, 20},
		{"ZIndex", s.ZIndex, 5},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("Style.%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestBridge_LuaTableToVNode_StyleShorthand(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	// Test fg/bg shorthand.
	err := L.DoString(`return {
		type = "box",
		style = { fg = "green", bg = "black" }
	}`)
	if err != nil {
		t.Fatal(err)
	}

	vn := b.LuaTableToVNode(-1)
	L.Pop(1)

	if vn.Style.Foreground != "green" {
		t.Errorf("Foreground = %q, want %q", vn.Style.Foreground, "green")
	}
	if vn.Style.Background != "black" {
		t.Errorf("Background = %q, want %q", vn.Style.Background, "black")
	}
}

func TestBridge_LuaTableToVNode_Props(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	err := L.DoString(`return {
		type = "box",
		id = "btn",
		label = "Click me",
		tabIndex = 3,
	}`)
	if err != nil {
		t.Fatal(err)
	}

	vn := b.LuaTableToVNode(-1)
	L.Pop(1)

	if vn.Props["label"] != "Click me" {
		t.Errorf("Props[label] = %v, want %q", vn.Props["label"], "Click me")
	}
	// Lua integers come through as int64.
	if vn.Props["tabIndex"] != int64(3) {
		t.Errorf("Props[tabIndex] = %v (%T), want int64(3)", vn.Props["tabIndex"], vn.Props["tabIndex"])
	}
}

func TestBridge_LuaTableToVNode_FunctionProps(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	err := L.DoString(`return {
		type = "box",
		id = "btn",
		onClick = function(e) end,
	}`)
	if err != nil {
		t.Fatal(err)
	}

	vn := b.LuaTableToVNode(-1)
	L.Pop(1)

	ref, ok := vn.Props["onClick"].(int)
	if !ok || ref <= 0 {
		t.Errorf("Props[onClick] should be a positive int ref, got %v (%T)", vn.Props["onClick"], vn.Props["onClick"])
	}

	// Cleanup tracked refs.
	b.ReleaseRefs()
}

// --- WrapRenderFn tests ---

func TestBridge_WrapRenderFn(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	// Create a Lua render function that returns a VNode table.
	err := L.DoString(`
		function myRender(state, props)
			return {
				type = "text",
				content = props.label or "default",
			}
		end
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Store function as registry ref.
	L.GetGlobal("myRender")
	ref := L.Ref(lua.RegistryIndex)

	renderFn := b.WrapRenderFn(ref)

	// Call the render function.
	vn := renderFn(
		map[string]any{},
		map[string]any{"label": "Hello"},
	)

	if vn.Type != "text" {
		t.Errorf("Type = %q, want %q", vn.Type, "text")
	}
	if vn.Content != "Hello" {
		t.Errorf("Content = %q, want %q", vn.Content, "Hello")
	}
}

func TestBridge_WrapRenderFn_InvalidRef(t *testing.T) {
	b := newTestBridge(t)

	// Use a ref that doesn't point to a function.
	renderFn := b.WrapRenderFn(lua.RefNil)
	vn := renderFn(nil, nil)

	if vn.Type != "box" {
		t.Errorf("Type = %q, want fallback %q", vn.Type, "box")
	}
}

// --- ExtractHandlers tests ---

func TestBridge_ExtractHandlers(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	// Create a VNode tree with onClick handler stored as a ref.
	err := L.DoString(`return {
		type = "box",
		id = "btn",
		onClick = function(e) end,
		children = {
			{ type = "text", id = "label", onMouseEnter = function(e) end },
		}
	}`)
	if err != nil {
		t.Fatal(err)
	}

	root := b.LuaTableToVNode(-1)
	L.Pop(1)

	handlers := b.ExtractHandlers(root)

	// btn should have "click" handler.
	if hm, ok := handlers["btn"]; !ok {
		t.Error("handlers[btn] not found")
	} else if _, ok := hm["click"]; !ok {
		t.Error("handlers[btn][click] not found")
	}

	// label should have "mouseenter" handler.
	if hm, ok := handlers["label"]; !ok {
		t.Error("handlers[label] not found")
	} else if _, ok := hm["mouseenter"]; !ok {
		t.Error("handlers[label][mouseenter] not found")
	}

	b.ReleaseRefs()
}

func TestBridge_ExtractHandlers_NoID(t *testing.T) {
	b := newTestBridge(t)

	// VNode without ID should not produce handlers.
	root := &layout.VNode{
		Type:  "box",
		Props: map[string]any{"onClick": 42},
	}

	handlers := b.ExtractHandlers(root)
	if len(handlers) != 0 {
		t.Errorf("expected no handlers for VNode without ID, got %d", len(handlers))
	}
}

func TestBridge_ExtractHandlers_CallHandler(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	// Create a handler that sets a global variable when called.
	err := L.DoString(`
		handler_called = false
		return {
			type = "box",
			id = "btn",
			onClick = function(e)
				handler_called = true
			end,
		}
	`)
	if err != nil {
		t.Fatal(err)
	}

	root := b.LuaTableToVNode(-1)
	L.Pop(1)

	handlers := b.ExtractHandlers(root)
	clickHandler := handlers["btn"]["click"]

	// Call the handler.
	clickHandler(&event.Event{Type: "click", X: 5, Y: 10, Target: "btn"})

	// Check that the Lua global was set.
	if !L.GetFieldBool(-10002, "handler_called") {
		// Try via GetGlobal.
		L.GetGlobal("handler_called")
		result := L.ToBoolean(-1)
		L.Pop(1)
		if !result {
			t.Error("handler was not called (handler_called is still false)")
		}
	}

	b.ReleaseRefs()
}

// --- RegisterHooks tests ---

func TestBridge_RegisterHooks(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	b.RegisterHooks()

	// Check that all lumina hooks exist.
	err := L.DoString(`
		assert(type(lumina) == "table", "lumina should be a table")
		assert(type(lumina.useState) == "function", "useState should be a function")
		assert(type(lumina.useEffect) == "function", "useEffect should be a function")
		assert(type(lumina.useMemo) == "function", "useMemo should be a function")
		assert(type(lumina.createElement) == "function", "createElement should be a function")
		assert(type(lumina.useCallback) == "function", "useCallback should be a function")
		assert(type(lumina.useRef) == "function", "useRef should be a function")
		assert(type(lumina.useReducer) == "function", "useReducer should be a function")
		assert(type(lumina.useId) == "function", "useId should be a function")
		assert(type(lumina.useLayoutEffect) == "function", "useLayoutEffect should be a function")
		assert(type(lumina.useAnimation) == "function", "useAnimation should be a function")
		assert(type(lumina.navigate) == "function", "navigate should be a function")
		assert(type(lumina.back) == "function", "back should be a function")
		assert(type(lumina.useRoute) == "function", "useRoute should be a function")
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBridge_UseState(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	// Create a component for the hook to operate on.
	comp := newHookTestComponent("test-comp")
	b.SetCurrentComponent(comp)
	b.RegisterHooks()

	// Call useState from Lua.
	err := L.DoString(`
		val, setter = lumina.useState("count", 0)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Check initial value.
	L.GetGlobal("val")
	val := L.ToAny(-1)
	L.Pop(1)
	if val != int64(0) {
		t.Errorf("initial value = %v (%T), want int64(0)", val, val)
	}

	// Call setter.
	err = L.DoString(`setter(42)`)
	if err != nil {
		t.Fatal(err)
	}

	// Verify state was updated.
	if comp.State()["count"] != int64(42) {
		t.Errorf("State[count] = %v, want int64(42)", comp.State()["count"])
	}
	if !comp.IsDirtyPaint() {
		t.Error("DirtyPaint should be true after setState")
	}
}

func TestBridge_UseState_Persists(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	comp := newHookTestComponent("test-comp")
	b.SetCurrentComponent(comp)
	b.RegisterHooks()

	// First call: initialize.
	err := L.DoString(`val1, _ = lumina.useState("x", 10)`)
	if err != nil {
		t.Fatal(err)
	}

	// Manually set state (simulating setter call).
	comp.SetState("x", int64(99))

	// Second call: should return persisted value, not initial.
	err = L.DoString(`val2, _ = lumina.useState("x", 10)`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("val2")
	val := L.ToAny(-1)
	L.Pop(1)
	if val != int64(99) {
		t.Errorf("persisted value = %v, want int64(99)", val)
	}
}

func TestBridge_UseEffect(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	comp := newHookTestComponent("test-comp")
	b.SetCurrentComponent(comp)
	b.ResetHookIndices()
	b.RegisterHooks()

	// Effect that sets a global and returns cleanup.
	err := L.DoString(`
		effect_ran = 0
		cleanup_ran = 0
		lumina.useEffect(function()
			effect_ran = effect_ran + 1
			return function()
				cleanup_ran = cleanup_ran + 1
			end
		end, {1})
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("effect_ran")
	effectRan := L.ToAny(-1)
	L.Pop(1)
	if effectRan != int64(1) {
		t.Errorf("effect_ran = %v, want 1", effectRan)
	}

	// Run again with same deps — should NOT run.
	b.ResetHookIndices()
	err = L.DoString(`
		lumina.useEffect(function()
			effect_ran = effect_ran + 1
			return function()
				cleanup_ran = cleanup_ran + 1
			end
		end, {1})
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("effect_ran")
	effectRan = L.ToAny(-1)
	L.Pop(1)
	if effectRan != int64(1) {
		t.Errorf("effect_ran = %v after same deps, want 1 (should not re-run)", effectRan)
	}

	// Run with different deps — should run and call cleanup.
	b.ResetHookIndices()
	err = L.DoString(`
		lumina.useEffect(function()
			effect_ran = effect_ran + 1
			return function()
				cleanup_ran = cleanup_ran + 1
			end
		end, {2})
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("effect_ran")
	effectRan = L.ToAny(-1)
	L.Pop(1)
	if effectRan != int64(2) {
		t.Errorf("effect_ran = %v after changed deps, want 2", effectRan)
	}

	L.GetGlobal("cleanup_ran")
	cleanupRan := L.ToAny(-1)
	L.Pop(1)
	if cleanupRan != int64(1) {
		t.Errorf("cleanup_ran = %v, want 1", cleanupRan)
	}
}

func TestBridge_UseMemo(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	comp := newHookTestComponent("test-comp")
	b.SetCurrentComponent(comp)
	b.ResetHookIndices()
	b.RegisterHooks()

	// Memoize a computation.
	err := L.DoString(`
		compute_count = 0
		result = lumina.useMemo(function()
			compute_count = compute_count + 1
			return 42
		end, {1})
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("result")
	result := L.ToAny(-1)
	L.Pop(1)
	if result != int64(42) {
		t.Errorf("result = %v, want 42", result)
	}

	L.GetGlobal("compute_count")
	count := L.ToAny(-1)
	L.Pop(1)
	if count != int64(1) {
		t.Errorf("compute_count = %v, want 1", count)
	}

	// Same deps — should return cached value without recomputing.
	b.ResetHookIndices()
	err = L.DoString(`
		result2 = lumina.useMemo(function()
			compute_count = compute_count + 1
			return 99
		end, {1})
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("result2")
	result2 := L.ToAny(-1)
	L.Pop(1)
	if result2 != int64(42) {
		t.Errorf("result2 = %v, want cached 42", result2)
	}

	L.GetGlobal("compute_count")
	count = L.ToAny(-1)
	L.Pop(1)
	if count != int64(1) {
		t.Errorf("compute_count = %v after same deps, want 1 (no recompute)", count)
	}

	// Changed deps — should recompute.
	b.ResetHookIndices()
	err = L.DoString(`
		result3 = lumina.useMemo(function()
			compute_count = compute_count + 1
			return 99
		end, {2})
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("result3")
	result3 := L.ToAny(-1)
	L.Pop(1)
	if result3 != int64(99) {
		t.Errorf("result3 = %v, want recomputed 99", result3)
	}
}

func TestBridge_CreateElement(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	b.RegisterHooks()

	err := L.DoString(`
		local el = lumina.createElement("button", { id = "btn1", label = "OK" },
			{ type = "text", content = "Click" }
		)
		result_type = el.type
		result_id = el.id
		result_label = el.label
		-- Check children
		result_child_type = el.children[1].type
		result_child_content = el.children[1].content
	`)
	if err != nil {
		t.Fatal(err)
	}

	checks := map[string]string{
		"result_type":          "button",
		"result_id":            "btn1",
		"result_label":         "OK",
		"result_child_type":    "text",
		"result_child_content": "Click",
	}
	for global, want := range checks {
		L.GetGlobal(global)
		got, _ := L.ToString(-1)
		L.Pop(1)
		if got != want {
			t.Errorf("%s = %q, want %q", global, got, want)
		}
	}
}

// --- ReleaseRefs tests ---

func TestBridge_ReleaseRefs(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	// Create some Lua functions and track their refs.
	err := L.DoString(`return function() end`)
	if err != nil {
		t.Fatal(err)
	}
	ref1 := L.Ref(lua.RegistryIndex)
	b.TrackRef(ref1)

	err = L.DoString(`return function() end`)
	if err != nil {
		t.Fatal(err)
	}
	ref2 := L.Ref(lua.RegistryIndex)
	b.TrackRef(ref2)

	if len(b.refs) != 2 {
		t.Errorf("refs count = %d, want 2", len(b.refs))
	}

	b.ReleaseRefs()

	if len(b.refs) != 0 {
		t.Errorf("refs count after release = %d, want 0", len(b.refs))
	}

	// Verify refs are actually released — RawGetI should return nil.
	L.RawGetI(lua.RegistryIndex, int64(ref1))
	if L.IsFunction(-1) {
		t.Error("ref1 should be released but still points to a function")
	}
	L.Pop(1)
}

// --- Integration test ---

func TestBridge_FullRenderCycle(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	// Set up a component with render function.
	p := paint.NewPainter()
	mgr := component.NewManager(p)
	b.SetManager(mgr)
	b.RegisterHooks()

	// Define Lua render function.
	err := L.DoString(`
		function render(state, props)
			return {
				type = "vbox",
				id = "root",
				style = { width = 80, height = 24 },
				children = {
					{
						type = "text",
						id = "title",
						content = "Hello",
						style = { bold = true },
					},
					{
						type = "box",
						id = "btn",
						onClick = function(e) end,
						style = { width = 20, height = 3, border = "single" },
					},
				}
			}
		end
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Store render function ref.
	L.GetGlobal("render")
	renderRef := L.Ref(lua.RegistryIndex)

	// Wrap as Go RenderFunc.
	renderFn := b.WrapRenderFn(renderRef)

	// Call render.
	vn := renderFn(map[string]any{}, map[string]any{})

	// Verify tree structure.
	if vn.Type != "vbox" {
		t.Errorf("root type = %q, want %q", vn.Type, "vbox")
	}
	if len(vn.Children) != 2 {
		t.Fatalf("root children = %d, want 2", len(vn.Children))
	}

	title := vn.Children[0]
	if title.Content != "Hello" {
		t.Errorf("title content = %q, want %q", title.Content, "Hello")
	}
	if !title.Style.Bold {
		t.Error("title.Style.Bold should be true")
	}

	btn := vn.Children[1]
	if btn.Style.Width != 20 {
		t.Errorf("btn width = %d, want 20", btn.Style.Width)
	}
	if btn.Style.Border != "single" {
		t.Errorf("btn border = %q, want %q", btn.Style.Border, "single")
	}

	// Extract handlers.
	handlers := b.ExtractHandlers(vn)
	if _, ok := handlers["btn"]; !ok {
		t.Error("btn should have handlers")
	} else if _, ok := handlers["btn"]["click"]; !ok {
		t.Error("btn should have click handler")
	}

	// Release refs.
	b.ReleaseRefs()
}

// --- Utility tests ---

func TestDepsEqual(t *testing.T) {
	tests := []struct {
		a, b []any
		want bool
	}{
		{nil, nil, true},
		{[]any{}, []any{}, true},
		{[]any{int64(1)}, []any{int64(1)}, true},
		{[]any{int64(1)}, []any{int64(2)}, false},
		{[]any{"a"}, []any{"a"}, true},
		{[]any{"a"}, []any{"b"}, false},
		{[]any{int64(1), "a"}, []any{int64(1), "a"}, true},
		{[]any{int64(1)}, []any{int64(1), "a"}, false},
	}
	for i, tt := range tests {
		got := depsEqual(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("case %d: depsEqual(%v, %v) = %v, want %v", i, tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIsEventProp(t *testing.T) {
	eventProps := []string{
		"onClick", "onChange", "onFocus", "onBlur",
		"onKeyDown", "onKeyUp", "onSubmit", "onScroll",
		"onMouseDown", "onMouseUp", "onMouseMove",
		"onMouseEnter", "onMouseLeave",
		"onDragOver", "onDrop", "onWheel",
		"onInput", "onResize", "onContextMenu",
	}
	for _, p := range eventProps {
		if !isEventProp(p) {
			t.Errorf("isEventProp(%q) = false, want true", p)
		}
	}

	nonEventProps := []string{"id", "label", "style", "children", "type", "content"}
	for _, p := range nonEventProps {
		if isEventProp(p) {
			t.Errorf("isEventProp(%q) = true, want false", p)
		}
	}
}

func TestMapPropToEvent(t *testing.T) {
	tests := []struct {
		prop string
		want string
	}{
		{"onClick", "click"},
		{"onMouseEnter", "mouseenter"},
		{"onKeyDown", "keydown"},
		{"onContextMenu", "contextmenu"},
	}
	for _, tt := range tests {
		got := mapPropToEvent(tt.prop)
		if got != tt.want {
			t.Errorf("mapPropToEvent(%q) = %q, want %q", tt.prop, got, tt.want)
		}
	}
}

// --- New Hook Tests ---

// newHookTestBridge creates a bridge with component + HookContext ready.
func newHookTestBridge(t *testing.T) (*Bridge, *component.Component) {
	t.Helper()
	L := lua.NewState()
	t.Cleanup(func() { L.Close() })
	b := NewBridge(L)
	comp := newHookTestComponent("hook-test-comp")
	b.SetCurrentComponent(comp)
	b.RegisterHooks()
	// Initialize HookContext via BeginComponentRender.
	b.BeginComponentRender(comp)
	return b, comp
}

func TestBridge_UseCallback(t *testing.T) {
	b, comp := newHookTestBridge(t)
	L := b.L

	// First call: cache a callback.
	err := L.DoString(`
		compute_count = 0
		cb = lumina.useCallback(function()
			compute_count = compute_count + 1
		end, {1})
	`)
	if err != nil {
		t.Fatal(err)
	}

	// cb should be a function.
	L.GetGlobal("cb")
	if !L.IsFunction(-1) {
		t.Error("useCallback should return a function")
	}
	L.Pop(1)

	// Call the callback.
	err = L.DoString(`cb()`)
	if err != nil {
		t.Fatal(err)
	}
	L.GetGlobal("compute_count")
	count := L.ToAny(-1)
	L.Pop(1)
	if count != int64(1) {
		t.Errorf("compute_count = %v, want 1", count)
	}

	// End + begin new render with same deps — should return cached.
	b.EndComponentRender()
	b.BeginComponentRender(comp)

	err = L.DoString(`
		cb2 = lumina.useCallback(function()
			compute_count = compute_count + 100
		end, {1})
		cb2()
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("compute_count")
	count = L.ToAny(-1)
	L.Pop(1)
	// Same deps → cached function → increment by 1, not 100.
	if count != int64(2) {
		t.Errorf("compute_count = %v, want 2 (cached function)", count)
	}
}

func TestBridge_UseCallback_Caching(t *testing.T) {
	b, comp := newHookTestBridge(t)
	L := b.L

	// First render: cache a callback.
	err := L.DoString(`
		call_count = 0
		cb1 = lumina.useCallback(function()
			call_count = call_count + 1
		end, {1})
	`)
	if err != nil {
		t.Fatal(err)
	}

	b.EndComponentRender()

	// Second render, same deps: should return same cached function.
	b.BeginComponentRender(comp)
	err = L.DoString(`
		cb2 = lumina.useCallback(function()
			call_count = call_count + 100
		end, {1})
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Call cb2 — if cached, it should increment by 1 (old fn), not 100.
	err = L.DoString(`cb2()`)
	if err != nil {
		t.Fatal(err)
	}
	L.GetGlobal("call_count")
	count := L.ToAny(-1)
	L.Pop(1)
	if count != int64(1) {
		t.Errorf("call_count = %v, want 1 (cached function should be used)", count)
	}

	b.EndComponentRender()

	// Third render, different deps: should use new function.
	b.BeginComponentRender(comp)
	err = L.DoString(`
		cb3 = lumina.useCallback(function()
			call_count = call_count + 100
		end, {2})
		cb3()
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("call_count")
	count = L.ToAny(-1)
	L.Pop(1)
	if count != int64(101) {
		t.Errorf("call_count = %v, want 101 (new function after deps change)", count)
	}
}

func TestBridge_UseRef(t *testing.T) {
	b, comp := newHookTestBridge(t)
	L := b.L

	err := L.DoString(`
		ref = lumina.useRef(42)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Check that ref.current == 42.
	L.GetGlobal("ref")
	if !L.IsTable(-1) {
		t.Fatal("useRef should return a table")
	}
	L.GetField(-1, "current")
	val := L.ToAny(-1)
	L.Pop(2)
	if val != int64(42) {
		t.Errorf("ref.current = %v, want 42", val)
	}

	b.EndComponentRender()

	// Second render: ref should persist.
	b.BeginComponentRender(comp)
	err = L.DoString(`
		ref2 = lumina.useRef(99)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Should still be 42 (initial value persists).
	L.GetGlobal("ref2")
	L.GetField(-1, "current")
	val = L.ToAny(-1)
	L.Pop(2)
	if val != int64(42) {
		t.Errorf("ref2.current = %v, want 42 (should persist)", val)
	}
}

func TestBridge_UseReducer(t *testing.T) {
	b, _ := newHookTestBridge(t)
	L := b.L

	err := L.DoString(`
		state, dispatch = lumina.useReducer(function(state, action)
			if action == "increment" then
				return state + 1
			elseif action == "decrement" then
				return state - 1
			end
			return state
		end, 0)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Check initial state.
	L.GetGlobal("state")
	state := L.ToAny(-1)
	L.Pop(1)
	if state != int64(0) {
		t.Errorf("initial state = %v, want 0", state)
	}

	// Dispatch increment.
	err = L.DoString(`dispatch("increment")`)
	if err != nil {
		t.Fatal(err)
	}

	// Dispatch again.
	err = L.DoString(`dispatch("increment")`)
	if err != nil {
		t.Fatal(err)
	}

	// Note: useReducer state is in the HookContext, not visible until next render.
	// After dispatching, the component should be dirty.
}

func TestBridge_UseId(t *testing.T) {
	b, comp := newHookTestBridge(t)
	L := b.L

	err := L.DoString(`id1 = lumina.useId()`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("id1")
	id1, _ := L.ToString(-1)
	L.Pop(1)

	if id1 == "" {
		t.Error("useId should return a non-empty string")
	}
	if !strings.Contains(id1, "hook-test-comp") {
		t.Errorf("useId = %q, should contain component ID", id1)
	}

	b.EndComponentRender()

	// Second render: same position should return same ID.
	b.BeginComponentRender(comp)
	err = L.DoString(`id2 = lumina.useId()`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("id2")
	id2, _ := L.ToString(-1)
	L.Pop(1)

	if id1 != id2 {
		t.Errorf("useId not stable: first=%q, second=%q", id1, id2)
	}
}

func TestBridge_UseLayoutEffect(t *testing.T) {
	b, comp := newHookTestBridge(t)
	L := b.L

	err := L.DoString(`
		layout_effect_ran = 0
		layout_cleanup_ran = 0
		lumina.useLayoutEffect(function()
			layout_effect_ran = layout_effect_ran + 1
			return function()
				layout_cleanup_ran = layout_cleanup_ran + 1
			end
		end, {1})
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("layout_effect_ran")
	ran := L.ToAny(-1)
	L.Pop(1)
	if ran != int64(1) {
		t.Errorf("layout_effect_ran = %v, want 1", ran)
	}

	b.EndComponentRender()

	// Same deps — should NOT run.
	b.BeginComponentRender(comp)
	err = L.DoString(`
		lumina.useLayoutEffect(function()
			layout_effect_ran = layout_effect_ran + 1
			return function()
				layout_cleanup_ran = layout_cleanup_ran + 1
			end
		end, {1})
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("layout_effect_ran")
	ran = L.ToAny(-1)
	L.Pop(1)
	if ran != int64(1) {
		t.Errorf("layout_effect_ran = %v after same deps, want 1", ran)
	}

	b.EndComponentRender()

	// Changed deps — should run and call cleanup.
	b.BeginComponentRender(comp)
	err = L.DoString(`
		lumina.useLayoutEffect(function()
			layout_effect_ran = layout_effect_ran + 1
			return function()
				layout_cleanup_ran = layout_cleanup_ran + 1
			end
		end, {2})
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("layout_effect_ran")
	ran = L.ToAny(-1)
	L.Pop(1)
	if ran != int64(2) {
		t.Errorf("layout_effect_ran = %v after changed deps, want 2", ran)
	}

	L.GetGlobal("layout_cleanup_ran")
	cleanup := L.ToAny(-1)
	L.Pop(1)
	if cleanup != int64(1) {
		t.Errorf("layout_cleanup_ran = %v, want 1", cleanup)
	}
}

// --- Animation Hook Tests ---

func TestBridge_UseAnimation(t *testing.T) {
	b, _ := newHookTestBridge(t)
	L := b.L

	mgr := animation.NewManager()
	b.SetAnimationManager(mgr)
	// Re-register hooks so animation hook picks up the manager.
	b.RegisterHooks()

	err := L.DoString(`
		anim = lumina.useAnimation({
			id = "fade",
			from = 0,
			to = 1,
			duration = 500,
			easing = "linear"
		})
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Check returned table has value, start, stop.
	L.GetGlobal("anim")
	if !L.IsTable(-1) {
		t.Fatal("useAnimation should return a table")
	}

	L.GetField(-1, "value")
	val := L.ToAny(-1)
	L.Pop(1)
	// Initial value should be 0 (from).
	if v, ok := val.(float64); !ok || v != 0 {
		t.Errorf("anim.value = %v, want 0.0", val)
	}

	L.GetField(-1, "start")
	if !L.IsFunction(-1) {
		t.Error("anim.start should be a function")
	}
	L.Pop(1)

	L.GetField(-1, "stop")
	if !L.IsFunction(-1) {
		t.Error("anim.stop should be a function")
	}
	L.Pop(1)
	L.Pop(1) // pop anim table

	// Call start.
	err = L.DoString(`anim.start()`)
	if err != nil {
		t.Fatal(err)
	}

	// Animation should now be running.
	if mgr.Count() != 1 {
		t.Errorf("animation count = %d, want 1", mgr.Count())
	}

	// Call stop.
	err = L.DoString(`anim.stop()`)
	if err != nil {
		t.Fatal(err)
	}

	if mgr.Count() != 0 {
		t.Errorf("animation count after stop = %d, want 0", mgr.Count())
	}
}

func TestBridge_UseAnimation_NoManager(t *testing.T) {
	b := newTestBridge(t)
	L := b.L

	comp := newHookTestComponent("anim-test")
	b.SetCurrentComponent(comp)
	b.RegisterHooks()

	// Should error when no animation manager is set.
	err := L.DoString(`
		local ok, err = pcall(function()
			lumina.useAnimation({ id = "test", from = 0, to = 1, duration = 100 })
		end)
		anim_error = not ok
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("anim_error")
	hasError := L.ToBoolean(-1)
	L.Pop(1)
	if !hasError {
		t.Error("useAnimation without manager should error")
	}
}

// --- Router Hook Tests ---

func TestBridge_Navigate(t *testing.T) {
	b, _ := newHookTestBridge(t)
	L := b.L

	r := router.New()
	r.AddRoute("/users/:id")
	b.SetRouter(r)
	b.RegisterHooks()

	err := L.DoString(`lumina.navigate("/users/42")`)
	if err != nil {
		t.Fatal(err)
	}

	if r.CurrentPath() != "/users/42" {
		t.Errorf("CurrentPath = %q, want %q", r.CurrentPath(), "/users/42")
	}

	params := r.Params()
	if params["id"] != "42" {
		t.Errorf("Params[id] = %q, want %q", params["id"], "42")
	}
}

func TestBridge_Back(t *testing.T) {
	b, _ := newHookTestBridge(t)
	L := b.L

	r := router.New()
	r.AddRoute("/")
	r.AddRoute("/about")
	b.SetRouter(r)
	b.RegisterHooks()

	// Navigate to /about, then back.
	err := L.DoString(`
		lumina.navigate("/about")
		result = lumina.back()
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("result")
	result := L.ToBoolean(-1)
	L.Pop(1)
	if !result {
		t.Error("back() should return true when history exists")
	}

	if r.CurrentPath() != "/" {
		t.Errorf("CurrentPath after back = %q, want %q", r.CurrentPath(), "/")
	}

	// Back again with empty history.
	err = L.DoString(`result2 = lumina.back()`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("result2")
	result2 := L.ToBoolean(-1)
	L.Pop(1)
	if result2 {
		t.Error("back() should return false when history is empty")
	}
}

func TestBridge_UseRoute(t *testing.T) {
	b, _ := newHookTestBridge(t)
	L := b.L

	r := router.New()
	r.AddRoute("/users/:id")
	b.SetRouter(r)
	b.RegisterHooks()

	r.Navigate("/users/99")

	err := L.DoString(`
		route = lumina.useRoute()
		route_path = route.path
		route_id = route.params.id
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("route_path")
	path, _ := L.ToString(-1)
	L.Pop(1)
	if path != "/users/99" {
		t.Errorf("route.path = %q, want %q", path, "/users/99")
	}

	L.GetGlobal("route_id")
	id, _ := L.ToString(-1)
	L.Pop(1)
	if id != "99" {
		t.Errorf("route.params.id = %q, want %q", id, "99")
	}
}

func TestBridge_UseRoute_NoRouter(t *testing.T) {
	b, _ := newHookTestBridge(t)
	L := b.L

	// No router set — should return default.
	err := L.DoString(`
		route = lumina.useRoute()
		route_path = route.path
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("route_path")
	path, _ := L.ToString(-1)
	L.Pop(1)
	if path != "/" {
		t.Errorf("default route.path = %q, want %q", path, "/")
	}
}

func TestBridge_Navigate_NoRouter(t *testing.T) {
	b, _ := newHookTestBridge(t)
	L := b.L

	// Navigate without router should error.
	err := L.DoString(`
		local ok, _ = pcall(function()
			lumina.navigate("/test")
		end)
		nav_error = not ok
	`)
	if err != nil {
		t.Fatal(err)
	}

	L.GetGlobal("nav_error")
	hasError := L.ToBoolean(-1)
	L.Pop(1)
	if !hasError {
		t.Error("navigate without router should error")
	}
}

// --- HookContext Lifecycle Tests ---

func TestBridge_BeginEndComponentRender(t *testing.T) {
	b := newTestBridge(t)
	comp := newHookTestComponent("lifecycle-comp")

	b.BeginComponentRender(comp)

	if b.CurrentComponent() != comp {
		t.Error("CurrentComponent should be set after BeginComponentRender")
	}

	err := b.EndComponentRender()
	if err != nil {
		t.Errorf("EndComponentRender error: %v", err)
	}

	if b.CurrentComponent() != nil {
		t.Error("CurrentComponent should be nil after EndComponentRender")
	}
}

func TestBridge_DestroyComponent(t *testing.T) {
	b := newTestBridge(t)
	comp := newHookTestComponent("destroy-comp")

	// Create a hook context.
	b.BeginComponentRender(comp)
	_ = b.GetHookContext(comp)
	b.EndComponentRender()

	// Verify context exists.
	if _, ok := b.hookContexts["destroy-comp"]; !ok {
		t.Error("hookContext should exist before destroy")
	}

	b.DestroyComponent("destroy-comp")

	if _, ok := b.hookContexts["destroy-comp"]; ok {
		t.Error("hookContext should be removed after DestroyComponent")
	}
}
