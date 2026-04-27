package bridge

import (
	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/hooks"
)

// RegisterHooks registers Lua-callable hooks on the global "lumina" table:
//
//	lumina.useState(key, initialValue) → value, setter
//	lumina.useEffect(fn, deps)         → (side-effect registration)
//	lumina.useMemo(fn, deps)           → cached value
//	lumina.useCallback(fn, deps)       → cached function
//	lumina.useRef(initialValue)        → { current = value }
//	lumina.useReducer(reducer, init)   → state, dispatch
//	lumina.useId()                     → stable unique ID
//	lumina.useLayoutEffect(fn, deps)   → (synchronous side-effect)
//	lumina.createElement(type, props, children...) → VNode table
//	lumina.useAnimation(config)        → { value, start, stop }
//	lumina.navigate(path)              → (router navigation)
//	lumina.back()                      → bool
//	lumina.useRoute()                  → { path, params }
func (b *Bridge) RegisterHooks() {
	L := b.L

	// Create or get the "lumina" global table.
	L.GetGlobal("lumina")
	if !L.IsTable(-1) {
		L.Pop(1)
		L.NewTable()
	}
	tblIdx := L.AbsIndex(-1)

	// Core hooks
	L.PushFunction(b.luaUseState)
	L.SetField(tblIdx, "useState")

	L.PushFunction(b.luaUseEffect)
	L.SetField(tblIdx, "useEffect")

	L.PushFunction(b.luaUseMemo)
	L.SetField(tblIdx, "useMemo")

	L.PushFunction(b.luaCreateElement)
	L.SetField(tblIdx, "createElement")

	// New hooks
	L.PushFunction(b.luaUseCallback)
	L.SetField(tblIdx, "useCallback")

	L.PushFunction(b.luaUseRef)
	L.SetField(tblIdx, "useRef")

	L.PushFunction(b.luaUseReducer)
	L.SetField(tblIdx, "useReducer")

	L.PushFunction(b.luaUseId)
	L.SetField(tblIdx, "useId")

	L.PushFunction(b.luaUseLayoutEffect)
	L.SetField(tblIdx, "useLayoutEffect")

	// Animation hooks
	L.PushFunction(b.luaUseAnimation)
	L.SetField(tblIdx, "useAnimation")

	// Router hooks
	L.PushFunction(b.luaNavigate)
	L.SetField(tblIdx, "navigate")

	L.PushFunction(b.luaBack)
	L.SetField(tblIdx, "back")

	L.PushFunction(b.luaUseRoute)
	L.SetField(tblIdx, "useRoute")

	L.SetGlobal("lumina")
}

// luaUseState implements lumina.useState(key, initialValue).
// Returns: currentValue, setterFunction
//
// This uses key-based state stored in the component's State map (not
// call-index based) for Lua ergonomics.
func (b *Bridge) luaUseState(L *lua.State) int {
	comp := b.currentComp
	if comp == nil {
		L.PushString("useState: no current component")
		L.Error()
		return 0
	}

	key := L.CheckString(1)
	var initial any
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		initial = L.ToAny(2)
	}

	// Initialize state on first call.
	if _, exists := comp.State()[key]; !exists {
		comp.SetState(key, initial)
	}

	// Push current value.
	L.PushAny(comp.State()[key])

	// Push setter function.
	mgr := b.manager
	L.PushFunction(func(L *lua.State) int {
		newValue := L.ToAny(1)
		comp.SetState(key, newValue)
		if mgr != nil {
			mgr.SetState(comp.ID(), key, newValue)
		}
		return 0
	})

	return 2
}

// luaUseEffect implements lumina.useEffect(fn, deps).
// Uses HookContext for proper call-index tracking and dependency comparison.
func (b *Bridge) luaUseEffect(L *lua.State) int {
	return b.luaUseEffectInternal(L, false)
}

// luaUseLayoutEffect implements lumina.useLayoutEffect(fn, deps).
// Like useEffect but runs synchronously after render.
func (b *Bridge) luaUseLayoutEffect(L *lua.State) int {
	return b.luaUseEffectInternal(L, true)
}

// luaUseEffectInternal is shared by useEffect and useLayoutEffect.
func (b *Bridge) luaUseEffectInternal(L *lua.State, isLayout bool) int {
	comp := b.currentComp
	if comp == nil {
		if isLayout {
			L.PushString("useLayoutEffect: no current component")
		} else {
			L.PushString("useEffect: no current component")
		}
		L.Error()
		return 0
	}

	L.CheckAny(1)
	if L.Type(1) != lua.TypeFunction {
		L.PushString("useEffect: first argument must be a function")
		L.Error()
		return 0
	}

	// Read deps array if provided.
	var newDeps []any
	hasDeps := false
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		L.CheckType(2, lua.TypeTable)
		hasDeps = true
		newDeps = luaTableToSlice(L, 2)
	}

	// Use HookContext for dependency tracking.
	hc := b.GetHookContext(comp)
	var eff *hooks.Effect
	if isLayout {
		eff = hc.UseLayoutEffect(newDeps, hasDeps)
	} else {
		eff = hc.UseEffect(newDeps, hasDeps)
	}

	if eff.IsPending() {
		// Run cleanup from previous effect.
		eff.RunCleanup()

		// Call the effect function.
		L.PushValue(1) // push the effect fn
		if status := L.PCall(0, 1, 0); status != lua.OK {
			L.Pop(1) // pop error
			eff.ClearPending()
			return 0
		}

		// If effect returned a function, store it as cleanup.
		if L.Type(-1) == lua.TypeFunction {
			cleanupRef := L.Ref(lua.RegistryIndex)
			luaState := b.L
			eff.SetCleanup(func() {
				if err := luaState.CallRef(cleanupRef, 0, 0); err != nil {
					// Cleanup failed — ignore but release ref.
				}
				luaState.Unref(lua.RegistryIndex, cleanupRef)
			})
		} else {
			L.Pop(1)
		}

		eff.ClearPending()
	}

	return 0
}

// luaUseMemo implements lumina.useMemo(fn, deps).
// Uses HookContext for proper call-index tracking and caching.
func (b *Bridge) luaUseMemo(L *lua.State) int {
	comp := b.currentComp
	if comp == nil {
		L.PushString("useMemo: no current component")
		L.Error()
		return 0
	}

	L.CheckAny(1)
	if L.Type(1) != lua.TypeFunction {
		L.PushString("useMemo: first argument must be a function")
		L.Error()
		return 0
	}

	var newDeps []any
	hasDeps := false
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		L.CheckType(2, lua.TypeTable)
		hasDeps = true
		newDeps = luaTableToSlice(L, 2)
	}

	hc := b.GetHookContext(comp)
	memo := hc.UseMemo(newDeps, hasDeps)

	if memo.IsStale() {
		// Release old cached value if it was a Lua ref.
		if old, ok := memo.Value().(int); ok && old > 0 {
			L.Unref(lua.RegistryIndex, old)
		}

		// Call the factory function.
		L.PushValue(1)
		if status := L.PCall(0, 1, 0); status != lua.OK {
			L.Pop(1) // pop error
			L.PushNil()
			return 1
		}

		// Store result as registry ref so it persists.
		L.PushValue(-1) // duplicate for Ref
		ref := L.Ref(lua.RegistryIndex)
		memo.Set(ref)
		// Result is already on stack.
		return 1
	}

	// Return cached value from registry.
	ref, ok := memo.Value().(int)
	if ok && ref > 0 {
		L.RawGetI(lua.RegistryIndex, int64(ref))
	} else {
		L.PushNil()
	}
	return 1
}

// luaUseCallback implements lumina.useCallback(fn, deps).
// Caches a Lua function reference, only updating when deps change.
func (b *Bridge) luaUseCallback(L *lua.State) int {
	comp := b.currentComp
	if comp == nil {
		L.PushString("useCallback: no current component")
		L.Error()
		return 0
	}

	L.CheckAny(1)
	if L.Type(1) != lua.TypeFunction {
		L.PushString("useCallback: first argument must be a function")
		L.Error()
		return 0
	}

	var deps []any
	hasDeps := false
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		L.CheckType(2, lua.TypeTable)
		hasDeps = true
		deps = luaTableToSlice(L, 2)
	}

	hc := b.GetHookContext(comp)
	memo := hc.UseCallback(deps, hasDeps)

	if memo.IsStale() {
		// Release old cached ref.
		if old, ok := memo.Value().(int); ok && old > 0 {
			L.Unref(lua.RegistryIndex, old)
		}
		// Cache the callback function as a registry ref.
		L.PushValue(1)
		ref := L.Ref(lua.RegistryIndex)
		memo.Set(ref)
	}

	// Push cached function from registry.
	ref, ok := memo.Value().(int)
	if ok && ref > 0 {
		L.RawGetI(lua.RegistryIndex, int64(ref))
	} else {
		L.PushNil()
	}
	return 1
}

// luaUseRef implements lumina.useRef(initialValue).
// Returns a table { current = value } that persists across renders.
// The Lua table is recreated each call, but the underlying Ref is stable.
func (b *Bridge) luaUseRef(L *lua.State) int {
	comp := b.currentComp
	if comp == nil {
		L.PushString("useRef: no current component")
		L.Error()
		return 0
	}

	var initial any
	if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
		initial = L.ToAny(1)
	}

	hc := b.GetHookContext(comp)
	ref := hc.UseRef(initial)

	// Return a table with { current = value }.
	L.NewTable()
	tblIdx := L.AbsIndex(-1)
	L.PushAny(ref.Current)
	L.SetField(tblIdx, "current")
	return 1
}

// luaUseReducer implements lumina.useReducer(reducer, initialState).
// Returns: currentState, dispatchFunction
func (b *Bridge) luaUseReducer(L *lua.State) int {
	comp := b.currentComp
	if comp == nil {
		L.PushString("useReducer: no current component")
		L.Error()
		return 0
	}

	L.CheckAny(1)
	if L.Type(1) != lua.TypeFunction {
		L.PushString("useReducer: first argument must be a function")
		L.Error()
		return 0
	}

	L.CheckAny(2)
	initialState := L.ToAny(2)

	// Store the reducer function as a registry ref.
	L.PushValue(1)
	reducerRef := L.Ref(lua.RegistryIndex)

	luaState := b.L
	reducer := hooks.ReducerFunc(func(state any, action any) any {
		luaState.RawGetI(lua.RegistryIndex, int64(reducerRef))
		luaState.PushAny(state)
		luaState.PushAny(action)
		if status := luaState.PCall(2, 1, 0); status != lua.OK {
			luaState.Pop(1)
			return state
		}
		result := luaState.ToAny(-1)
		luaState.Pop(1)
		return result
	})

	hc := b.GetHookContext(comp)
	state, dispatch := hc.UseReducer(reducer, initialState)

	// Push current state.
	L.PushAny(state)

	// Push dispatch function.
	L.PushFunction(func(L *lua.State) int {
		action := L.ToAny(1)
		dispatch(action)
		return 0
	})

	return 2
}

// luaUseId implements lumina.useId().
// Returns a stable unique ID string for this hook call position.
func (b *Bridge) luaUseId(L *lua.State) int {
	comp := b.currentComp
	if comp == nil {
		L.PushString("useId: no current component")
		L.Error()
		return 0
	}

	hc := b.GetHookContext(comp)
	id := hc.UseId()
	L.PushString(id)
	return 1
}

// luaCreateElement implements lumina.createElement(type, props, children...).
// Returns a Lua table representing a VNode:
//
//	{ type = "box", id = props.id, style = props.style, children = {...}, ... }
func (b *Bridge) luaCreateElement(L *lua.State) int {
	nodeType := L.CheckString(1)

	// Create result table.
	L.NewTable()
	resultIdx := L.AbsIndex(-1)

	// Set type.
	L.PushString(nodeType)
	L.SetField(resultIdx, "type")

	// Copy props into result table.
	if L.GetTop() >= 2 && L.IsTable(2) {
		L.ForEach(2, func(L *lua.State) bool {
			if L.Type(-2) == lua.TypeString {
				key, _ := L.ToString(-2)
				L.PushValue(-1) // push value copy
				L.SetField(resultIdx, key)
			}
			return true
		})
	}

	// Collect vararg children (args 3+) into a children array.
	nArgs := L.GetTop()
	if nArgs > 2 {
		L.CreateTable(nArgs-2, 0)
		childrenIdx := L.AbsIndex(-1)
		ci := int64(1)
		for i := 3; i <= nArgs; i++ {
			L.PushValue(i)
			L.RawSetI(childrenIdx, ci)
			ci++
		}
		L.SetField(resultIdx, "children")
	}

	return 1
}

// --- Hook lifecycle (legacy compat) ---

// ResetHookIndices resets hook call indices for a new render pass.
// Must be called before each component render.
// This is a compatibility shim — prefer BeginComponentRender.
func (b *Bridge) ResetHookIndices() {
	if b.currentComp == nil {
		return
	}
	hc := b.GetHookContext(b.currentComp)
	hc.BeginRender()
}

// --- Utility ---

// luaTableToSlice reads a Lua array table into a Go []any slice.
func luaTableToSlice(L *lua.State, idx int) []any {
	absIdx := L.AbsIndex(idx)
	n := int(L.RawLen(absIdx))
	result := make([]any, 0, n)
	for i := 1; i <= n; i++ {
		L.RawGetI(absIdx, int64(i))
		result = append(result, L.ToAny(-1))
		L.Pop(1)
	}
	return result
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
