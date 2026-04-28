package render

import (
	"fmt"

	"github.com/akzj/go-lua/pkg/lua"
)

// --- Hook helpers ---

// luaTableToDeps reads a Lua table at idx and returns a Go slice for dependency comparison.
// Returns (deps, hasDeps). If the arg is nil/none, hasDeps=false.
func luaTableToDeps(L *lua.State, idx int) ([]any, bool) {
	if L.GetTop() < idx || L.IsNoneOrNil(idx) {
		return nil, false
	}
	absIdx := L.AbsIndex(idx)
	if !L.IsTable(absIdx) {
		return nil, false
	}
	var result []any
	n := int(L.RawLen(absIdx))
	for i := 1; i <= n; i++ {
		L.RawGetI(absIdx, int64(i))
		result = append(result, L.ToAny(-1))
		L.Pop(1)
	}
	return result, true
}

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

// getHookSlot returns the hook slot at the current hookIdx for the component,
// creating it if this is the first render. Returns (slot, isNew).
// Panics (Lua error) if the hook kind doesn't match (ordering violation).
func getHookSlot(L *lua.State, comp *Component, kind hookKind) (*hookSlot, bool) {
	idx := comp.hookIdx
	comp.hookIdx++

	if idx < len(comp.hookSlots) {
		slot := comp.hookSlots[idx]
		if slot.kind != kind {
			kindNames := map[hookKind]string{
				hookEffect: "useEffect",
				hookMemo:   "useMemo/useCallback",
				hookRef:    "useRef",
			}
			L.PushString(fmt.Sprintf("hook ordering violation at index %d: expected %s, got %s",
				idx, kindNames[slot.kind], kindNames[kind]))
			L.Error()
			return nil, false
		}
		return slot, false
	}

	// New slot
	slot := &hookSlot{kind: kind}
	comp.hookSlots = append(comp.hookSlots, slot)
	return slot, true
}

// --- useEffect ---

// luaUseEffect implements lumina.useEffect(callback, deps?)
// callback: function that may return a cleanup function
// deps: optional table of dependency values
//   - nil/absent = run every render
//   - {} (empty table) = run once (mount only)
//   - {a, b, ...} = run when any dep changes
func (e *Engine) luaUseEffect(L *lua.State) int {
	comp := e.currentComp
	if comp == nil {
		L.PushString("useEffect: no current component")
		L.Error()
		return 0
	}

	// Arg 1: callback function
	L.CheckType(1, lua.TypeFunction)

	// Arg 2: deps table (optional)
	deps, hasDeps := luaTableToDeps(L, 2)

	slot, isNew := getHookSlot(L, comp, hookEffect)

	if isNew {
		// First render: store callback ref, mark pending
		L.PushValue(1)
		ref := L.Ref(lua.RegistryIndex)
		slot.effect = &effectSlot{
			deps:        deps,
			callbackRef: ref,
			pending:     true,
		}
	} else {
		eff := slot.effect
		// Update callback ref (it may be a new closure each render)
		L.Unref(lua.RegistryIndex, eff.callbackRef)
		L.PushValue(1)
		eff.callbackRef = L.Ref(lua.RegistryIndex)

		if !hasDeps {
			// No deps → run every render
			eff.pending = true
		} else {
			// Check if deps changed
			eff.pending = !depsEqual(eff.deps, deps)
		}
		if eff.pending {
			eff.deps = deps
		}
	}

	return 0
}

// --- useRef ---

// luaUseRef implements lumina.useRef(initialValue?)
// Returns the same table {current = value} across renders.
func (e *Engine) luaUseRef(L *lua.State) int {
	comp := e.currentComp
	if comp == nil {
		L.PushString("useRef: no current component")
		L.Error()
		return 0
	}

	slot, isNew := getHookSlot(L, comp, hookRef)

	if isNew {
		// First render: create {current = initialValue}
		L.NewTable()
		tbl := L.AbsIndex(-1)
		if L.GetTop() >= 2 && !L.IsNoneOrNil(1) {
			// arg 1 is the initial value (before hookIdx was incremented, but
			// the Lua args haven't changed — arg 1 is still at stack position 1)
			L.PushValue(1)
		} else {
			L.PushNil()
		}
		L.SetField(tbl, "current")
		// Store table in registry so we return the SAME table each render
		ref := L.Ref(lua.RegistryIndex)
		slot.ref = &refSlot{tableRef: ref}
	}

	// Push the same table every render
	L.RawGetI(lua.RegistryIndex, int64(slot.ref.tableRef))
	return 1
}

// --- useMemo ---

// luaUseMemo implements lumina.useMemo(factory, deps)
// factory: function that returns a value
// deps: table of dependency values
// Returns cached value if deps haven't changed.
func (e *Engine) luaUseMemo(L *lua.State) int {
	comp := e.currentComp
	if comp == nil {
		L.PushString("useMemo: no current component")
		L.Error()
		return 0
	}

	// Arg 1: factory function
	L.CheckType(1, lua.TypeFunction)

	// Arg 2: deps table
	deps, hasDeps := luaTableToDeps(L, 2)

	slot, isNew := getHookSlot(L, comp, hookMemo)

	if isNew {
		// First render: compute value
		L.PushValue(1) // push factory
		if status := L.PCall(0, 1, 0); status != lua.OK {
			L.Error() // propagate error
			return 0
		}
		// Store result in registry
		ref := L.Ref(lua.RegistryIndex)
		slot.memo = &memoSlot{deps: deps, ref: ref}
		// Push result for return
		L.RawGetI(lua.RegistryIndex, int64(ref))
	} else {
		m := slot.memo
		stale := !hasDeps || !depsEqual(m.deps, deps)
		if stale {
			// Recompute
			L.Unref(lua.RegistryIndex, m.ref)
			L.PushValue(1) // push factory
			if status := L.PCall(0, 1, 0); status != lua.OK {
				L.Error()
				return 0
			}
			m.ref = L.Ref(lua.RegistryIndex)
			m.deps = deps
		}
		// Push cached/new result
		L.RawGetI(lua.RegistryIndex, int64(m.ref))
	}

	return 1
}

// --- useCallback ---

// luaUseCallback implements lumina.useCallback(fn, deps)
// Sugar for useMemo that caches the function itself (not calling it).
func (e *Engine) luaUseCallback(L *lua.State) int {
	comp := e.currentComp
	if comp == nil {
		L.PushString("useCallback: no current component")
		L.Error()
		return 0
	}

	// Arg 1: callback function
	L.CheckType(1, lua.TypeFunction)

	// Arg 2: deps table
	deps, hasDeps := luaTableToDeps(L, 2)

	slot, isNew := getHookSlot(L, comp, hookMemo) // shares memoSlot

	if isNew {
		// First render: store the function directly (don't call it)
		L.PushValue(1)
		ref := L.Ref(lua.RegistryIndex)
		slot.memo = &memoSlot{deps: deps, ref: ref}
		L.RawGetI(lua.RegistryIndex, int64(ref))
	} else {
		m := slot.memo
		stale := !hasDeps || !depsEqual(m.deps, deps)
		if stale {
			L.Unref(lua.RegistryIndex, m.ref)
			L.PushValue(1)
			m.ref = L.Ref(lua.RegistryIndex)
			m.deps = deps
		}
		L.RawGetI(lua.RegistryIndex, int64(m.ref))
	}

	return 1
}

// --- Effect lifecycle ---

// firePendingEffects runs useEffect callbacks that are pending after a render cycle.
// Called after paint in RenderDirty/RenderAll.
func (e *Engine) firePendingEffects() {
	L := e.L
	for _, comp := range e.components {
		for _, slot := range comp.hookSlots {
			if slot.kind != hookEffect || slot.effect == nil || !slot.effect.pending {
				continue
			}
			eff := slot.effect
			eff.pending = false

			// Run cleanup from previous execution
			if eff.cleanupRef != 0 {
				L.RawGetI(lua.RegistryIndex, int64(eff.cleanupRef))
				if L.IsFunction(-1) {
					L.PCall(0, 0, 0)
				} else {
					L.Pop(1)
				}
				L.Unref(lua.RegistryIndex, eff.cleanupRef)
				eff.cleanupRef = 0
			}

			// Call effect callback
			L.RawGetI(lua.RegistryIndex, int64(eff.callbackRef))
			if L.IsFunction(-1) {
				if status := L.PCall(0, 1, 0); status == lua.OK {
					// If callback returned a function, store as cleanup
					if L.IsFunction(-1) {
						eff.cleanupRef = L.Ref(lua.RegistryIndex)
					} else {
						L.Pop(1)
					}
				} else {
					L.Pop(1) // pop error
				}
			} else {
				L.Pop(1)
			}
		}
	}
}

// cleanupComponentHooks frees all hook-related Lua refs for a component.
// Called when a component is unmounted.
func (e *Engine) cleanupComponentHooks(comp *Component) {
	L := e.L
	for _, slot := range comp.hookSlots {
		switch slot.kind {
		case hookEffect:
			if slot.effect == nil {
				continue
			}
			eff := slot.effect
			// Run cleanup
			if eff.cleanupRef != 0 {
				L.RawGetI(lua.RegistryIndex, int64(eff.cleanupRef))
				if L.IsFunction(-1) {
					L.PCall(0, 0, 0)
				} else {
					L.Pop(1)
				}
				L.Unref(lua.RegistryIndex, eff.cleanupRef)
			}
			// Free callback ref
			if eff.callbackRef != 0 {
				L.Unref(lua.RegistryIndex, eff.callbackRef)
			}
		case hookMemo:
			if slot.memo != nil && slot.memo.ref != 0 {
				L.Unref(lua.RegistryIndex, slot.memo.ref)
			}
		case hookRef:
			if slot.ref != nil && slot.ref.tableRef != 0 {
				L.Unref(lua.RegistryIndex, slot.ref.tableRef)
			}
		}
	}
	comp.hookSlots = nil
}
