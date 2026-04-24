package lumina

import (
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

// newTestState creates a fresh Lua state with lumina loaded and a test component active.
func newTestState(t *testing.T) (*lua.State, *Component) {
	t.Helper()
	L := lua.NewState()
	Open(L)

	comp := &Component{
		ID:    "test_comp",
		Type:  "TestComp",
		Name:  "TestComp",
		Props: make(map[string]any),
		State: make(map[string]any),
	}
	globalRegistry.mu.Lock()
	globalRegistry.components[comp.ID] = comp
	globalRegistry.mu.Unlock()

	SetCurrentComponent(comp)
	return L, comp
}

func cleanupTestState(L *lua.State, comp *Component) {
	SetCurrentComponent(nil)
	globalRegistry.mu.Lock()
	delete(globalRegistry.components, comp.ID)
	globalRegistry.mu.Unlock()
	L.Close()
}

func TestUseState_BasicGetSet(t *testing.T) {
	L, comp := newTestState(t)
	defer cleanupTestState(L, comp)

	// Call useState("count", 0)
	err := L.DoString(`
		local val, set = lumina.useState("count", 0)
		_test_val = val
	`)
	if err != nil {
		t.Fatalf("useState: %v", err)
	}

	L.GetGlobal("_test_val")
	v, ok := L.ToInteger(-1)
	L.Pop(1)
	if !ok || v != 0 {
		t.Fatalf("expected initial value 0, got %v (ok=%v)", v, ok)
	}

	// Verify state is stored on component.
	if comp.State["count"] != int64(0) {
		t.Fatalf("expected comp.State['count'] = 0, got %v", comp.State["count"])
	}
}

func TestUseEffect_RunsOnFirstRender(t *testing.T) {
	L, comp := newTestState(t)
	defer cleanupTestState(L, comp)

	comp.ResetHookIndex()
	err := L.DoString(`
		_effect_ran = false
		lumina.useEffect(function()
			_effect_ran = true
		end, {1})
	`)
	if err != nil {
		t.Fatalf("useEffect: %v", err)
	}

	L.GetGlobal("_effect_ran")
	if !L.ToBoolean(-1) {
		t.Fatal("expected effect to run on first render")
	}
	L.Pop(1)
}

func TestUseEffect_SkipsWhenDepsUnchanged(t *testing.T) {
	L, comp := newTestState(t)
	defer cleanupTestState(L, comp)

	// First render — effect runs.
	comp.ResetHookIndex()
	err := L.DoString(`
		_run_count = 0
		lumina.useEffect(function()
			_run_count = _run_count + 1
		end, {42})
	`)
	if err != nil {
		t.Fatalf("useEffect first: %v", err)
	}

	// Second render with same deps — effect should NOT run again.
	comp.ResetHookIndex()
	err = L.DoString(`
		lumina.useEffect(function()
			_run_count = _run_count + 1
		end, {42})
	`)
	if err != nil {
		t.Fatalf("useEffect second: %v", err)
	}

	L.GetGlobal("_run_count")
	v, _ := L.ToInteger(-1)
	L.Pop(1)
	if v != 1 {
		t.Fatalf("expected effect to run 1 time, ran %d times", v)
	}
}

func TestUseEffect_RerunsWhenDepsChange(t *testing.T) {
	L, comp := newTestState(t)
	defer cleanupTestState(L, comp)

	// First render with dep=1.
	comp.ResetHookIndex()
	err := L.DoString(`
		_run_count = 0
		lumina.useEffect(function()
			_run_count = _run_count + 1
		end, {1})
	`)
	if err != nil {
		t.Fatalf("useEffect first: %v", err)
	}

	// Second render with dep=2 — should re-run.
	comp.ResetHookIndex()
	err = L.DoString(`
		lumina.useEffect(function()
			_run_count = _run_count + 1
		end, {2})
	`)
	if err != nil {
		t.Fatalf("useEffect second: %v", err)
	}

	L.GetGlobal("_run_count")
	v, _ := L.ToInteger(-1)
	L.Pop(1)
	if v != 2 {
		t.Fatalf("expected effect to run 2 times, ran %d times", v)
	}
}

func TestUseEffect_Cleanup(t *testing.T) {
	L, comp := newTestState(t)
	defer cleanupTestState(L, comp)

	// First render — effect returns cleanup.
	comp.ResetHookIndex()
	err := L.DoString(`
		_cleanup_ran = false
		lumina.useEffect(function()
			return function()
				_cleanup_ran = true
			end
		end, {1})
	`)
	if err != nil {
		t.Fatalf("useEffect first: %v", err)
	}

	// Cleanup should NOT have run yet.
	L.GetGlobal("_cleanup_ran")
	if L.ToBoolean(-1) {
		t.Fatal("cleanup should not run on first render")
	}
	L.Pop(1)

	// Second render with changed deps — cleanup should run before new effect.
	comp.ResetHookIndex()
	err = L.DoString(`
		lumina.useEffect(function()
			return function() end
		end, {2})
	`)
	if err != nil {
		t.Fatalf("useEffect second: %v", err)
	}

	L.GetGlobal("_cleanup_ran")
	if !L.ToBoolean(-1) {
		t.Fatal("expected cleanup to run when deps changed")
	}
	L.Pop(1)
}

func TestUseMemo_CachesOnSameDeps(t *testing.T) {
	L, comp := newTestState(t)
	defer cleanupTestState(L, comp)

	// First render.
	comp.ResetHookIndex()
	err := L.DoString(`
		_compute_count = 0
		_memo_val = lumina.useMemo(function()
			_compute_count = _compute_count + 1
			return 42
		end, {1})
	`)
	if err != nil {
		t.Fatalf("useMemo first: %v", err)
	}

	// Second render with same deps — should NOT recompute.
	comp.ResetHookIndex()
	err = L.DoString(`
		_memo_val = lumina.useMemo(function()
			_compute_count = _compute_count + 1
			return 42
		end, {1})
	`)
	if err != nil {
		t.Fatalf("useMemo second: %v", err)
	}

	L.GetGlobal("_compute_count")
	v, _ := L.ToInteger(-1)
	L.Pop(1)
	if v != 1 {
		t.Fatalf("expected compute 1 time, got %d", v)
	}

	L.GetGlobal("_memo_val")
	mv, _ := L.ToInteger(-1)
	L.Pop(1)
	if mv != 42 {
		t.Fatalf("expected memo value 42, got %d", mv)
	}
}

func TestUseMemo_RecomputesOnDepsChange(t *testing.T) {
	L, comp := newTestState(t)
	defer cleanupTestState(L, comp)

	comp.ResetHookIndex()
	err := L.DoString(`
		_compute_count = 0
		_memo_val = lumina.useMemo(function()
			_compute_count = _compute_count + 1
			return 100
		end, {1})
	`)
	if err != nil {
		t.Fatalf("useMemo first: %v", err)
	}

	comp.ResetHookIndex()
	err = L.DoString(`
		_memo_val = lumina.useMemo(function()
			_compute_count = _compute_count + 1
			return 200
		end, {2})
	`)
	if err != nil {
		t.Fatalf("useMemo second: %v", err)
	}

	L.GetGlobal("_compute_count")
	v, _ := L.ToInteger(-1)
	L.Pop(1)
	if v != 2 {
		t.Fatalf("expected compute 2 times, got %d", v)
	}

	L.GetGlobal("_memo_val")
	mv, _ := L.ToInteger(-1)
	L.Pop(1)
	if mv != 200 {
		t.Fatalf("expected memo value 200, got %d", mv)
	}
}

func TestUseCallback_ReturnsCachedFunction(t *testing.T) {
	L, comp := newTestState(t)
	defer cleanupTestState(L, comp)

	comp.ResetHookIndex()
	err := L.DoString(`
		_cb = lumina.useCallback(function() return "hello" end, {1})
	`)
	if err != nil {
		t.Fatalf("useCallback: %v", err)
	}

	L.GetGlobal("_cb")
	if L.Type(-1) != lua.TypeFunction {
		t.Fatalf("expected function, got %s", L.TypeName(L.Type(-1)))
	}
	L.Pop(1)
}

func TestUseReducer_DispatchUpdatesState(t *testing.T) {
	L, comp := newTestState(t)
	defer cleanupTestState(L, comp)

	comp.ResetHookIndex()
	err := L.DoString(`
		local state, dispatch = lumina.useReducer(
			function(state, action)
				if action == "increment" then
					return state + 1
				end
				return state
			end,
			0
		)
		_reducer_state = state
		_dispatch = dispatch
	`)
	if err != nil {
		t.Fatalf("useReducer: %v", err)
	}

	L.GetGlobal("_reducer_state")
	v, _ := L.ToInteger(-1)
	L.Pop(1)
	if v != 0 {
		t.Fatalf("expected initial state 0, got %d", v)
	}

	// Dispatch an action.
	err = L.DoString(`_dispatch("increment")`)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	// Re-render to get updated state.
	comp.ResetHookIndex()
	err = L.DoString(`
		local state, dispatch = lumina.useReducer(
			function(state, action)
				if action == "increment" then
					return state + 1
				end
				return state
			end,
			0
		)
		_reducer_state = state
	`)
	if err != nil {
		t.Fatalf("useReducer re-render: %v", err)
	}

	L.GetGlobal("_reducer_state")
	v, _ = L.ToInteger(-1)
	L.Pop(1)
	if v != 1 {
		t.Fatalf("expected state 1 after dispatch, got %d", v)
	}
}

func TestCreateContext_UseContext(t *testing.T) {
	L, comp := newTestState(t)
	defer cleanupTestState(L, comp)
	defer ClearContextValues()

	err := L.DoString(`
		_ctx = lumina.createContext("default_value")
		_val = lumina.useContext(_ctx)
	`)
	if err != nil {
		t.Fatalf("createContext/useContext: %v", err)
	}

	L.GetGlobal("_val")
	v, ok := L.ToString(-1)
	L.Pop(1)
	if !ok || v != "default_value" {
		t.Fatalf("expected 'default_value', got '%v'", v)
	}

	// Set context value and read again.
	err = L.DoString(`
		lumina.setContextValue(_ctx, "updated_value")
		_val = lumina.useContext(_ctx)
	`)
	if err != nil {
		t.Fatalf("setContextValue/useContext: %v", err)
	}

	L.GetGlobal("_val")
	v, ok = L.ToString(-1)
	L.Pop(1)
	if !ok || v != "updated_value" {
		t.Fatalf("expected 'updated_value', got '%v'", v)
	}
}

func TestDepsEqual(t *testing.T) {
	if !depsEqual(nil, nil) {
		t.Fatal("nil == nil")
	}
	if !depsEqual([]any{}, []any{}) {
		t.Fatal("[] == []")
	}
	if !depsEqual([]any{int64(1), "hello"}, []any{int64(1), "hello"}) {
		t.Fatal("[1, 'hello'] == [1, 'hello']")
	}
	if depsEqual([]any{int64(1)}, []any{int64(2)}) {
		t.Fatal("[1] != [2]")
	}
	if depsEqual([]any{int64(1)}, []any{int64(1), int64(2)}) {
		t.Fatal("[1] != [1, 2]")
	}
}

func TestRunEffectCleanups(t *testing.T) {
	L, comp := newTestState(t)
	defer cleanupTestState(L, comp)

	comp.ResetHookIndex()
	err := L.DoString(`
		_cleanup_count = 0
		lumina.useEffect(function()
			return function()
				_cleanup_count = _cleanup_count + 1
			end
		end)
	`)
	if err != nil {
		t.Fatalf("useEffect: %v", err)
	}

	// Run cleanups (simulates unmount).
	RunEffectCleanups(L, comp)

	L.GetGlobal("_cleanup_count")
	v, _ := L.ToInteger(-1)
	L.Pop(1)
	if v != 1 {
		t.Fatalf("expected cleanup to run 1 time, ran %d", v)
	}
}
