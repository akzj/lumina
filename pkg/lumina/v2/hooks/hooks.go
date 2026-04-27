// Package hooks implements React-style hook state management in pure Go.
// It has zero Lua dependency — the Lua bridge wraps these hooks.
//
// Usage:
//
//	ctx := hooks.NewHookContext("comp-1", func() { /* mark dirty */ })
//	ctx.BeginRender()
//	val, setter := ctx.UseState("initial")
//	ref := ctx.UseRef(nil)
//	eff := ctx.UseEffect([]any{"dep1"}, true)
//	memo := ctx.UseMemo([]any{"dep1"}, true)
//	err := ctx.EndRender()
package hooks

import (
	"fmt"
	"sync/atomic"
)

// ---------- global ID counter for useId ----------

var globalIDCounter int64

func nextGlobalID() int64 {
	return atomic.AddInt64(&globalIDCounter, 1)
}

// ---------- HookContext ----------

// HookContext manages all hook state for a single component render cycle.
type HookContext struct {
	id      string
	onDirty func() // called when useState setter / useReducer dispatch fires

	callIdx   int // current hook call index (reset each render)
	prevCount int // hook count from previous render (-1 = first render)

	hooks []hookSlot // ordered list of all hook slots

	rendered bool // true between BeginRender and EndRender
}

// hookKind identifies the type of hook.
type hookKind int

const (
	kindState hookKind = iota
	kindRef
	kindEffect
	kindLayoutEffect
	kindMemo
	kindReducer
	kindID
)

// hookSlot is a tagged union for any hook type.
type hookSlot struct {
	kind    hookKind
	state   *stateSlot
	ref     *Ref
	effect  *Effect
	memo    *Memo
	reducer *reducerSlot
	id      string
}

// stateSlot stores one useState call.
type stateSlot struct {
	value any
}

// reducerSlot stores one useReducer call.
type reducerSlot struct {
	state   any
	reducer ReducerFunc
}

// NewHookContext creates a new hook context for a component.
// onDirty is called whenever a useState setter or useReducer dispatch fires.
// It may be nil.
func NewHookContext(componentID string, onDirty func()) *HookContext {
	return &HookContext{
		id:        componentID,
		onDirty:   onDirty,
		prevCount: -1, // first render
	}
}

// ID returns the component ID associated with this hook context.
func (h *HookContext) ID() string { return h.id }

// BeginRender resets the hook call index for a new render cycle.
func (h *HookContext) BeginRender() {
	h.callIdx = 0
	h.rendered = true
}

// EndRender finalizes the render cycle. It validates that the same number
// of hooks were called as in the previous render (React's rules of hooks).
func (h *HookContext) EndRender() error {
	h.rendered = false
	count := h.callIdx
	if h.prevCount >= 0 && count != h.prevCount {
		return fmt.Errorf("hooks: hook count changed between renders: was %d, now %d (component %s)", h.prevCount, count, h.id)
	}
	h.prevCount = count
	return nil
}

// ---------- useState ----------

// UseState returns (currentValue, setter). On the first call at this index,
// the slot is initialized with initialValue.
func (h *HookContext) UseState(initialValue any) (value any, setter func(any)) {
	idx := h.callIdx
	h.callIdx++

	if idx >= len(h.hooks) {
		slot := &stateSlot{value: initialValue}
		h.hooks = append(h.hooks, hookSlot{kind: kindState, state: slot})
	}

	slot := h.hooks[idx].state
	setter = func(newVal any) {
		slot.value = newVal
		if h.onDirty != nil {
			h.onDirty()
		}
	}
	return slot.value, setter
}

// ---------- useRef ----------

// Ref holds a mutable reference that persists across renders.
type Ref struct {
	Current any
}

// UseRef returns a persistent Ref. On the first call at this index,
// the ref is initialized with initialValue.
func (h *HookContext) UseRef(initialValue any) *Ref {
	idx := h.callIdx
	h.callIdx++

	if idx >= len(h.hooks) {
		ref := &Ref{Current: initialValue}
		h.hooks = append(h.hooks, hookSlot{kind: kindRef, ref: ref})
	}
	return h.hooks[idx].ref
}

// ---------- useEffect / useLayoutEffect ----------

// Effect represents a side effect with dependency tracking.
// The actual callback is NOT stored here — the bridge stores Lua function refs.
// HookContext only tracks deps and whether the effect needs to run (pending).
type Effect struct {
	deps    []any
	cleanup func()
	ran     bool
	pending bool
	layout  bool // true for useLayoutEffect
}

// IsPending returns true if the effect should run this cycle.
func (e *Effect) IsPending() bool { return e.pending }

// ClearPending marks the effect as no longer pending.
func (e *Effect) ClearPending() { e.pending = false }

// SetCleanup stores the cleanup function returned by the effect callback.
func (e *Effect) SetCleanup(fn func()) { e.cleanup = fn }

// Cleanup returns the current cleanup function, or nil.
func (e *Effect) Cleanup() func() { return e.cleanup }

// RunCleanup runs the cleanup function if set, then clears it.
func (e *Effect) RunCleanup() {
	if e.cleanup != nil {
		e.cleanup()
		e.cleanup = nil
	}
}

// UseEffect registers a side effect. If deps changed (or first run, or no deps),
// the returned Effect has pending=true.
func (h *HookContext) UseEffect(deps []any, hasDeps bool) *Effect {
	return h.useEffectInternal(deps, hasDeps, false)
}

// UseLayoutEffect registers a layout effect (runs synchronously after render).
func (h *HookContext) UseLayoutEffect(deps []any, hasDeps bool) *Effect {
	return h.useEffectInternal(deps, hasDeps, true)
}

func (h *HookContext) useEffectInternal(deps []any, hasDeps bool, isLayout bool) *Effect {
	idx := h.callIdx
	h.callIdx++

	kind := kindEffect
	if isLayout {
		kind = kindLayoutEffect
	}

	if idx >= len(h.hooks) {
		eff := &Effect{
			deps:    deps,
			pending: true,
			ran:     true, // mark as ran on first creation
			layout:  isLayout,
		}
		h.hooks = append(h.hooks, hookSlot{kind: hookKind(kind), effect: eff})
		return eff
	}

	eff := h.hooks[idx].effect
	if !hasDeps {
		// No deps → run every render.
		eff.pending = true
	} else {
		eff.pending = !depsEqual(eff.deps, deps)
	}

	if eff.pending {
		eff.deps = deps
	}
	return eff
}

// ---------- useMemo ----------

// Memo caches a computed value.
type Memo struct {
	deps     []any
	value    any
	hasValue bool
	stale    bool
}

// Value returns the cached value.
func (m *Memo) Value() any { return m.value }

// IsStale returns true if deps changed and the value needs recomputing.
func (m *Memo) IsStale() bool { return m.stale }

// Set stores a new computed value and clears the stale flag.
func (m *Memo) Set(value any) {
	m.value = value
	m.hasValue = true
	m.stale = false
}

// UseMemo returns a Memo. If deps changed, Memo.IsStale() is true and the
// caller should recompute and call Memo.Set(newValue).
func (h *HookContext) UseMemo(deps []any, hasDeps bool) *Memo {
	idx := h.callIdx
	h.callIdx++

	if idx >= len(h.hooks) {
		m := &Memo{deps: deps, stale: true}
		h.hooks = append(h.hooks, hookSlot{kind: kindMemo, memo: m})
		return m
	}

	m := h.hooks[idx].memo
	if !m.hasValue {
		m.stale = true
	} else if !hasDeps {
		m.stale = true
	} else {
		m.stale = !depsEqual(m.deps, deps)
	}

	if m.stale {
		m.deps = deps
	}
	return m
}

// UseCallback is sugar for UseMemo that caches a function reference.
func (h *HookContext) UseCallback(deps []any, hasDeps bool) *Memo {
	return h.UseMemo(deps, hasDeps)
}

// ---------- useReducer ----------

// ReducerFunc is the reducer signature: (state, action) → newState.
type ReducerFunc func(state any, action any) any

// UseReducer returns (currentState, dispatch). On the first call at this
// index, the slot is initialized with initialState.
func (h *HookContext) UseReducer(reducer ReducerFunc, initialState any) (state any, dispatch func(action any)) {
	idx := h.callIdx
	h.callIdx++

	if idx >= len(h.hooks) {
		slot := &reducerSlot{
			state:   initialState,
			reducer: reducer,
		}
		h.hooks = append(h.hooks, hookSlot{kind: kindReducer, reducer: slot})
	}

	slot := h.hooks[idx].reducer
	dispatch = func(action any) {
		slot.state = slot.reducer(slot.state, action)
		if h.onDirty != nil {
			h.onDirty()
		}
	}
	return slot.state, dispatch
}

// ---------- useId ----------

// UseId returns a stable unique ID for this hook call position.
// The ID is generated once (first render) and persists across re-renders.
func (h *HookContext) UseId() string {
	idx := h.callIdx
	h.callIdx++

	if idx >= len(h.hooks) {
		id := fmt.Sprintf("%s:hook-%d", h.id, nextGlobalID())
		h.hooks = append(h.hooks, hookSlot{kind: kindID, id: id})
	}
	return h.hooks[idx].id
}

// ---------- Pending Effects ----------

// PendingEffects returns all non-layout effects that need to run this cycle.
func (h *HookContext) PendingEffects() []*Effect {
	var out []*Effect
	for _, s := range h.hooks {
		if s.kind == kindEffect && s.effect.pending {
			out = append(out, s.effect)
		}
	}
	return out
}

// PendingLayoutEffects returns layout effects that need to run this cycle.
func (h *HookContext) PendingLayoutEffects() []*Effect {
	var out []*Effect
	for _, s := range h.hooks {
		if s.kind == kindLayoutEffect && s.effect.pending {
			out = append(out, s.effect)
		}
	}
	return out
}

// ---------- Destroy ----------

// Destroy runs all effect cleanups (called when component unmounts).
func (h *HookContext) Destroy() {
	for _, s := range h.hooks {
		if s.kind == kindEffect || s.kind == kindLayoutEffect {
			s.effect.RunCleanup()
		}
	}
}

// ---------- Utilities ----------

// depsEqual compares two dependency slices for shallow equality.
func depsEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
