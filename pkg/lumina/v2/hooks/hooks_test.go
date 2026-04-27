package hooks

import (
	"strings"
	"testing"
)

// ---------- useState ----------

func TestUseState_InitialValue(t *testing.T) {
	h := NewHookContext("c1", nil)
	h.BeginRender()

	val, _ := h.UseState(42)
	if val != 42 {
		t.Fatalf("expected 42, got %v", val)
	}

	if err := h.EndRender(); err != nil {
		t.Fatal(err)
	}
}

func TestUseState_SetterUpdatesValue(t *testing.T) {
	h := NewHookContext("c1", nil)

	// Render 1: initialize
	h.BeginRender()
	_, setter := h.UseState("hello")
	h.EndRender()

	// Mutate between renders.
	setter("world")

	// Render 2: read updated value
	h.BeginRender()
	val, _ := h.UseState("hello") // initialValue ignored on re-render
	h.EndRender()

	if val != "world" {
		t.Fatalf("expected 'world', got %v", val)
	}
}

func TestUseState_MultipleStates(t *testing.T) {
	h := NewHookContext("c1", nil)

	h.BeginRender()
	v1, s1 := h.UseState(1)
	v2, s2 := h.UseState("a")
	v3, _ := h.UseState(true)
	h.EndRender()

	if v1 != 1 || v2 != "a" || v3 != true {
		t.Fatalf("initial values wrong: %v, %v, %v", v1, v2, v3)
	}

	s1(10)
	s2("b")

	h.BeginRender()
	v1, _ = h.UseState(1)
	v2, _ = h.UseState("a")
	v3, _ = h.UseState(true)
	h.EndRender()

	if v1 != 10 {
		t.Fatalf("expected 10, got %v", v1)
	}
	if v2 != "b" {
		t.Fatalf("expected 'b', got %v", v2)
	}
	if v3 != true {
		t.Fatalf("expected true, got %v", v3)
	}
}

func TestUseState_SetterTriggersOnDirty(t *testing.T) {
	called := 0
	h := NewHookContext("c1", func() { called++ })

	h.BeginRender()
	_, setter := h.UseState(0)
	h.EndRender()

	setter(1)
	setter(2)

	if called != 2 {
		t.Fatalf("expected onDirty called 2 times, got %d", called)
	}
}

// ---------- useRef ----------

func TestUseRef_InitialValue(t *testing.T) {
	h := NewHookContext("c1", nil)
	h.BeginRender()
	ref := h.UseRef(99)
	h.EndRender()

	if ref.Current != 99 {
		t.Fatalf("expected 99, got %v", ref.Current)
	}
}

func TestUseRef_PersistsAcrossRenders(t *testing.T) {
	h := NewHookContext("c1", nil)

	h.BeginRender()
	ref := h.UseRef("init")
	h.EndRender()

	ref.Current = "mutated"

	h.BeginRender()
	ref2 := h.UseRef("init")
	h.EndRender()

	if ref2.Current != "mutated" {
		t.Fatalf("expected 'mutated', got %v", ref2.Current)
	}
	if ref != ref2 {
		t.Fatal("expected same Ref pointer across renders")
	}
}

func TestUseRef_NoOnDirtyTrigger(t *testing.T) {
	called := 0
	h := NewHookContext("c1", func() { called++ })

	h.BeginRender()
	ref := h.UseRef(0)
	h.EndRender()

	ref.Current = 42 // direct mutation — should NOT trigger onDirty

	if called != 0 {
		t.Fatalf("expected onDirty NOT called, got %d", called)
	}
}

// ---------- useEffect ----------

func TestUseEffect_NoDeps_PendingEveryRender(t *testing.T) {
	h := NewHookContext("c1", nil)

	for i := 0; i < 3; i++ {
		h.BeginRender()
		eff := h.UseEffect(nil, false)
		h.EndRender()

		if !eff.IsPending() {
			t.Fatalf("render %d: expected pending", i)
		}
		eff.ClearPending()
	}
}

func TestUseEffect_EmptyDeps_PendingOnlyFirst(t *testing.T) {
	h := NewHookContext("c1", nil)

	// Render 1: should be pending.
	h.BeginRender()
	eff := h.UseEffect([]any{}, true)
	h.EndRender()

	if !eff.IsPending() {
		t.Fatal("render 1: expected pending")
	}
	eff.ClearPending()

	// Render 2: same deps → not pending.
	h.BeginRender()
	eff = h.UseEffect([]any{}, true)
	h.EndRender()

	if eff.IsPending() {
		t.Fatal("render 2: expected NOT pending")
	}
}

func TestUseEffect_DepsChanged_Pending(t *testing.T) {
	h := NewHookContext("c1", nil)

	// Render 1.
	h.BeginRender()
	eff := h.UseEffect([]any{1, "a"}, true)
	h.EndRender()
	if !eff.IsPending() {
		t.Fatal("render 1: expected pending")
	}
	eff.ClearPending()

	// Render 2: same deps.
	h.BeginRender()
	eff = h.UseEffect([]any{1, "a"}, true)
	h.EndRender()
	if eff.IsPending() {
		t.Fatal("render 2: expected NOT pending")
	}

	// Render 3: changed deps.
	h.BeginRender()
	eff = h.UseEffect([]any{2, "a"}, true)
	h.EndRender()
	if !eff.IsPending() {
		t.Fatal("render 3: expected pending")
	}
}

func TestUseEffect_Cleanup(t *testing.T) {
	h := NewHookContext("c1", nil)
	cleaned := false

	h.BeginRender()
	eff := h.UseEffect(nil, false)
	h.EndRender()

	eff.SetCleanup(func() { cleaned = true })
	eff.ClearPending()

	// Simulate running cleanup.
	eff.RunCleanup()
	if !cleaned {
		t.Fatal("expected cleanup to run")
	}
	// Cleanup should be cleared after running.
	if eff.Cleanup() != nil {
		t.Fatal("expected cleanup to be nil after RunCleanup")
	}
}

// ---------- useLayoutEffect ----------

func TestUseLayoutEffect_SeparateFromEffect(t *testing.T) {
	h := NewHookContext("c1", nil)

	h.BeginRender()
	eff := h.UseEffect([]any{1}, true)
	leff := h.UseLayoutEffect([]any{1}, true)
	h.EndRender()

	if !eff.IsPending() {
		t.Fatal("effect should be pending")
	}
	if !leff.IsPending() {
		t.Fatal("layout effect should be pending")
	}

	// Verify they appear in the correct lists.
	pending := h.PendingEffects()
	lpending := h.PendingLayoutEffects()

	if len(pending) != 1 {
		t.Fatalf("expected 1 pending effect, got %d", len(pending))
	}
	if len(lpending) != 1 {
		t.Fatalf("expected 1 pending layout effect, got %d", len(lpending))
	}
}

// ---------- useMemo ----------

func TestUseMemo_FirstCall_Stale(t *testing.T) {
	h := NewHookContext("c1", nil)
	h.BeginRender()
	m := h.UseMemo([]any{1}, true)
	h.EndRender()

	if !m.IsStale() {
		t.Fatal("expected stale on first call")
	}

	m.Set(100)
	if m.IsStale() {
		t.Fatal("expected not stale after Set")
	}
	if m.Value() != 100 {
		t.Fatalf("expected 100, got %v", m.Value())
	}
}

func TestUseMemo_SameDeps_Cached(t *testing.T) {
	h := NewHookContext("c1", nil)

	// Render 1.
	h.BeginRender()
	m := h.UseMemo([]any{"x"}, true)
	h.EndRender()
	m.Set("computed")

	// Render 2: same deps.
	h.BeginRender()
	m = h.UseMemo([]any{"x"}, true)
	h.EndRender()

	if m.IsStale() {
		t.Fatal("expected NOT stale with same deps")
	}
	if m.Value() != "computed" {
		t.Fatalf("expected 'computed', got %v", m.Value())
	}
}

func TestUseMemo_ChangedDeps_Stale(t *testing.T) {
	h := NewHookContext("c1", nil)

	// Render 1.
	h.BeginRender()
	m := h.UseMemo([]any{1}, true)
	h.EndRender()
	m.Set("v1")

	// Render 2: changed deps.
	h.BeginRender()
	m = h.UseMemo([]any{2}, true)
	h.EndRender()

	if !m.IsStale() {
		t.Fatal("expected stale with changed deps")
	}
}

func TestUseMemo_NoDeps_StaleEveryRender(t *testing.T) {
	h := NewHookContext("c1", nil)

	for i := 0; i < 3; i++ {
		h.BeginRender()
		m := h.UseMemo(nil, false)
		h.EndRender()

		if !m.IsStale() {
			t.Fatalf("render %d: expected stale with no deps", i)
		}
		m.Set(i)
	}
}

// ---------- useCallback ----------

func TestUseCallback_IsAliasForUseMemo(t *testing.T) {
	h := NewHookContext("c1", nil)

	h.BeginRender()
	m := h.UseCallback([]any{1}, true)
	h.EndRender()

	if !m.IsStale() {
		t.Fatal("expected stale on first call")
	}
	fn := func() {}
	m.Set(fn)

	h.BeginRender()
	m2 := h.UseCallback([]any{1}, true)
	h.EndRender()

	if m2.IsStale() {
		t.Fatal("expected NOT stale with same deps")
	}
}

// ---------- useReducer ----------

func TestUseReducer_InitialState(t *testing.T) {
	reducer := func(state any, action any) any {
		return state.(int) + action.(int)
	}

	h := NewHookContext("c1", nil)
	h.BeginRender()
	state, _ := h.UseReducer(reducer, 10)
	h.EndRender()

	if state != 10 {
		t.Fatalf("expected 10, got %v", state)
	}
}

func TestUseReducer_DispatchUpdatesState(t *testing.T) {
	reducer := func(state any, action any) any {
		return state.(int) + action.(int)
	}

	h := NewHookContext("c1", nil)

	h.BeginRender()
	_, dispatch := h.UseReducer(reducer, 0)
	h.EndRender()

	dispatch(5)
	dispatch(3)

	h.BeginRender()
	state, _ := h.UseReducer(reducer, 0)
	h.EndRender()

	if state != 8 {
		t.Fatalf("expected 8, got %v", state)
	}
}

func TestUseReducer_DispatchTriggersOnDirty(t *testing.T) {
	called := 0
	reducer := func(state any, action any) any { return action }

	h := NewHookContext("c1", func() { called++ })
	h.BeginRender()
	_, dispatch := h.UseReducer(reducer, nil)
	h.EndRender()

	dispatch("action1")
	if called != 1 {
		t.Fatalf("expected onDirty called 1 time, got %d", called)
	}
}

// ---------- useContext ----------

func TestUseContext_DefaultValue(t *testing.T) {
	ctx := NewContext("default-theme")

	h := NewHookContext("c1", nil)
	h.BeginRender()
	val := h.UseContext(ctx, nil)
	h.EndRender()

	if val != "default-theme" {
		t.Fatalf("expected 'default-theme', got %v", val)
	}
}

func TestUseContext_ProviderValue(t *testing.T) {
	ctx := NewContext("default")

	provider := NewContextProvider(nil)
	provider.Set(ctx, "dark")

	h := NewHookContext("c1", nil)
	h.BeginRender()
	val := h.UseContext(ctx, provider)
	h.EndRender()

	if val != "dark" {
		t.Fatalf("expected 'dark', got %v", val)
	}
}

func TestUseContext_NestedProviders(t *testing.T) {
	ctx := NewContext("default")

	parent := NewContextProvider(nil)
	parent.Set(ctx, "light")

	child := NewContextProvider(parent)
	// child does NOT set ctx — should inherit from parent.

	h := NewHookContext("c1", nil)
	h.BeginRender()
	val := h.UseContext(ctx, child)
	h.EndRender()

	if val != "light" {
		t.Fatalf("expected 'light', got %v", val)
	}

	// Now override in child.
	child.Set(ctx, "dark")
	h.BeginRender()
	val = h.UseContext(ctx, child)
	h.EndRender()

	if val != "dark" {
		t.Fatalf("expected 'dark', got %v", val)
	}
}

// ---------- useId ----------

func TestUseId_StableAcrossRenders(t *testing.T) {
	h := NewHookContext("c1", nil)

	h.BeginRender()
	id1 := h.UseId()
	h.EndRender()

	h.BeginRender()
	id2 := h.UseId()
	h.EndRender()

	if id1 != id2 {
		t.Fatalf("expected stable ID, got %q and %q", id1, id2)
	}
}

func TestUseId_UniquePerPosition(t *testing.T) {
	h := NewHookContext("c1", nil)

	h.BeginRender()
	id1 := h.UseId()
	id2 := h.UseId()
	h.EndRender()

	if id1 == id2 {
		t.Fatalf("expected unique IDs, both are %q", id1)
	}
}

func TestUseId_ContainsComponentID(t *testing.T) {
	h := NewHookContext("my-comp", nil)

	h.BeginRender()
	id := h.UseId()
	h.EndRender()

	if !strings.HasPrefix(id, "my-comp:") {
		t.Fatalf("expected ID to start with 'my-comp:', got %q", id)
	}
}

// ---------- Hook order validation ----------

func TestEndRender_DetectsDifferentHookCount(t *testing.T) {
	h := NewHookContext("c1", nil)

	// Render 1: 2 hooks.
	h.BeginRender()
	h.UseState(1)
	h.UseRef(nil)
	if err := h.EndRender(); err != nil {
		t.Fatal(err)
	}

	// Render 2: 3 hooks — should error.
	h.BeginRender()
	h.UseState(1)
	h.UseRef(nil)
	h.UseState(2)
	err := h.EndRender()
	if err == nil {
		t.Fatal("expected error for different hook count")
	}
	if !strings.Contains(err.Error(), "hook count changed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEndRender_SameCountOK(t *testing.T) {
	h := NewHookContext("c1", nil)

	for i := 0; i < 5; i++ {
		h.BeginRender()
		h.UseState(0)
		h.UseRef(nil)
		if err := h.EndRender(); err != nil {
			t.Fatalf("render %d: unexpected error: %v", i, err)
		}
	}
}

// ---------- Destroy ----------

func TestDestroy_RunsAllCleanups(t *testing.T) {
	h := NewHookContext("c1", nil)
	cleaned := make([]int, 0)

	h.BeginRender()
	e1 := h.UseEffect(nil, false)
	e2 := h.UseLayoutEffect(nil, false)
	h.EndRender()

	e1.SetCleanup(func() { cleaned = append(cleaned, 1) })
	e2.SetCleanup(func() { cleaned = append(cleaned, 2) })

	h.Destroy()

	if len(cleaned) != 2 {
		t.Fatalf("expected 2 cleanups, got %d", len(cleaned))
	}
	if cleaned[0] != 1 || cleaned[1] != 2 {
		t.Fatalf("expected [1, 2], got %v", cleaned)
	}
}

// ---------- Multiple render cycles ----------

func TestMultipleRenderCycles_StatePersists(t *testing.T) {
	h := NewHookContext("c1", nil)

	// Render 1.
	h.BeginRender()
	val, setter := h.UseState(0)
	h.EndRender()
	if val != 0 {
		t.Fatalf("expected 0, got %v", val)
	}

	setter(1)

	// Render 2.
	h.BeginRender()
	val, setter = h.UseState(0)
	h.EndRender()
	if val != 1 {
		t.Fatalf("expected 1, got %v", val)
	}

	setter(2)

	// Render 3.
	h.BeginRender()
	val, _ = h.UseState(0)
	h.EndRender()
	if val != 2 {
		t.Fatalf("expected 2, got %v", val)
	}
}

func TestMultipleRenderCycles_EffectsReEvaluate(t *testing.T) {
	h := NewHookContext("c1", nil)

	// Render 1: deps=[1] → pending.
	h.BeginRender()
	eff := h.UseEffect([]any{1}, true)
	h.EndRender()
	if !eff.IsPending() {
		t.Fatal("render 1: expected pending")
	}
	eff.ClearPending()

	// Render 2: deps=[1] → not pending.
	h.BeginRender()
	eff = h.UseEffect([]any{1}, true)
	h.EndRender()
	if eff.IsPending() {
		t.Fatal("render 2: expected NOT pending")
	}

	// Render 3: deps=[2] → pending.
	h.BeginRender()
	eff = h.UseEffect([]any{2}, true)
	h.EndRender()
	if !eff.IsPending() {
		t.Fatal("render 3: expected pending")
	}
}

func TestMultipleRenderCycles_MemoReEvaluates(t *testing.T) {
	h := NewHookContext("c1", nil)

	// Render 1.
	h.BeginRender()
	m := h.UseMemo([]any{"a"}, true)
	h.EndRender()
	if !m.IsStale() {
		t.Fatal("render 1: expected stale")
	}
	m.Set("computed-a")

	// Render 2: same deps.
	h.BeginRender()
	m = h.UseMemo([]any{"a"}, true)
	h.EndRender()
	if m.IsStale() {
		t.Fatal("render 2: expected NOT stale")
	}
	if m.Value() != "computed-a" {
		t.Fatalf("render 2: expected 'computed-a', got %v", m.Value())
	}

	// Render 3: changed deps.
	h.BeginRender()
	m = h.UseMemo([]any{"b"}, true)
	h.EndRender()
	if !m.IsStale() {
		t.Fatal("render 3: expected stale")
	}
}

// ---------- Mixed hooks in one component ----------

func TestMixedHooks_CorrectOrdering(t *testing.T) {
	dirtyCount := 0
	h := NewHookContext("c1", func() { dirtyCount++ })

	// Render 1: state, ref, effect, memo, reducer, id
	h.BeginRender()
	v, setter := h.UseState(10)
	ref := h.UseRef("hello")
	eff := h.UseEffect([]any{1}, true)
	memo := h.UseMemo([]any{1}, true)
	rState, dispatch := h.UseReducer(func(s, a any) any { return s.(int) + a.(int) }, 100)
	id := h.UseId()
	h.EndRender()

	if v != 10 {
		t.Fatalf("state: expected 10, got %v", v)
	}
	if ref.Current != "hello" {
		t.Fatalf("ref: expected 'hello', got %v", ref.Current)
	}
	if !eff.IsPending() {
		t.Fatal("effect: expected pending")
	}
	if !memo.IsStale() {
		t.Fatal("memo: expected stale")
	}
	if rState != 100 {
		t.Fatalf("reducer: expected 100, got %v", rState)
	}
	if id == "" {
		t.Fatal("id: expected non-empty")
	}

	// Mutate.
	setter(20)
	ref.Current = "world"
	eff.ClearPending()
	memo.Set("cached")
	dispatch(50)

	if dirtyCount != 2 { // setter + dispatch
		t.Fatalf("expected 2 dirty calls, got %d", dirtyCount)
	}

	// Render 2: same hooks, same order.
	h.BeginRender()
	v, _ = h.UseState(10)
	ref2 := h.UseRef("hello")
	eff = h.UseEffect([]any{1}, true)
	memo = h.UseMemo([]any{1}, true)
	rState, _ = h.UseReducer(func(s, a any) any { return s.(int) + a.(int) }, 100)
	id2 := h.UseId()
	err := h.EndRender()

	if err != nil {
		t.Fatal(err)
	}
	if v != 20 {
		t.Fatalf("state: expected 20, got %v", v)
	}
	if ref2.Current != "world" {
		t.Fatalf("ref: expected 'world', got %v", ref2.Current)
	}
	if eff.IsPending() {
		t.Fatal("effect: expected NOT pending (same deps)")
	}
	if memo.IsStale() {
		t.Fatal("memo: expected NOT stale (same deps)")
	}
	if rState != 150 {
		t.Fatalf("reducer: expected 150, got %v", rState)
	}
	if id2 != id {
		t.Fatalf("id: expected stable %q, got %q", id, id2)
	}
}

// ---------- HookContext.ID ----------

func TestHookContext_ID(t *testing.T) {
	h := NewHookContext("my-component", nil)
	if h.ID() != "my-component" {
		t.Fatalf("expected 'my-component', got %q", h.ID())
	}
}

// ---------- Edge cases ----------

func TestUseState_NilInitialValue(t *testing.T) {
	h := NewHookContext("c1", nil)
	h.BeginRender()
	val, setter := h.UseState(nil)
	h.EndRender()

	if val != nil {
		t.Fatalf("expected nil, got %v", val)
	}

	setter("now-set")
	h.BeginRender()
	val, _ = h.UseState(nil)
	h.EndRender()

	if val != "now-set" {
		t.Fatalf("expected 'now-set', got %v", val)
	}
}

func TestUseEffect_DepsLengthChange(t *testing.T) {
	h := NewHookContext("c1", nil)

	h.BeginRender()
	eff := h.UseEffect([]any{1, 2}, true)
	h.EndRender()
	eff.ClearPending()

	// Different length deps → pending.
	h.BeginRender()
	eff = h.UseEffect([]any{1, 2, 3}, true)
	h.EndRender()

	if !eff.IsPending() {
		t.Fatal("expected pending when deps length changes")
	}
}

func TestPendingEffects_Empty(t *testing.T) {
	h := NewHookContext("c1", nil)
	h.BeginRender()
	h.EndRender()

	if len(h.PendingEffects()) != 0 {
		t.Fatal("expected no pending effects")
	}
	if len(h.PendingLayoutEffects()) != 0 {
		t.Fatal("expected no pending layout effects")
	}
}

func TestDestroy_NoEffects(t *testing.T) {
	h := NewHookContext("c1", nil)
	h.BeginRender()
	h.UseState(0)
	h.EndRender()

	// Should not panic.
	h.Destroy()
}
