package bridge

import (
	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/v2/component"
)

// RegisterHooks registers Lua-callable hooks on the global "lumina" table:
//
//	lumina.useState(key, initialValue) → value, setter
//	lumina.useEffect(fn, deps)         → (side-effect registration)
//	lumina.useMemo(fn, deps)           → cached value
//	lumina.createElement(type, props, children...) → VNode table
func (b *Bridge) RegisterHooks() {
	L := b.L

	// Create or get the "lumina" global table.
	L.GetGlobal("lumina")
	if !L.IsTable(-1) {
		L.Pop(1)
		L.NewTable()
	}
	tblIdx := L.AbsIndex(-1)

	// lumina.useState(key, initialValue) → value, setter
	L.PushFunction(b.luaUseState)
	L.SetField(tblIdx, "useState")

	// lumina.useEffect(fn, deps)
	L.PushFunction(b.luaUseEffect)
	L.SetField(tblIdx, "useEffect")

	// lumina.useMemo(fn, deps)
	L.PushFunction(b.luaUseMemo)
	L.SetField(tblIdx, "useMemo")

	// lumina.createElement(type, props, children...)
	L.PushFunction(b.luaCreateElement)
	L.SetField(tblIdx, "createElement")

	L.SetGlobal("lumina")
}

// luaUseState implements lumina.useState(key, initialValue).
// Returns: currentValue, setterFunction
//
// The setter function marks the component dirty via the Manager so it
// gets re-rendered on the next frame.
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
	// Capture comp and key by closure.
	mgr := b.manager
	L.PushFunction(func(L *lua.State) int {
		newValue := L.ToAny(1)
		comp.SetState(key, newValue)
		// Also notify manager if available.
		if mgr != nil {
			mgr.SetState(comp.ID(), key, newValue)
		}
		return 0
	})

	return 2
}

// EffectHook tracks a useEffect call across renders for a component.
type EffectHook struct {
	Deps       []any // dependency values from last run
	CleanupRef int   // Lua registry ref for cleanup function (0 = none)
	Ran        bool  // whether the effect has ever run
}

// luaUseEffect implements lumina.useEffect(fn, deps).
// Calls fn when deps change (or on every render if deps is nil).
// If fn returns a function, that function is called as cleanup before
// the next effect run.
func (b *Bridge) luaUseEffect(L *lua.State) int {
	comp := b.currentComp
	if comp == nil {
		L.PushString("useEffect: no current component")
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

	// Get or create the effect hook for this call index.
	hook := getEffectHook(comp, b)

	shouldRun := false
	if !hook.Ran {
		shouldRun = true
	} else if !hasDeps {
		// No deps = run every render.
		shouldRun = true
	} else {
		shouldRun = !depsEqual(hook.Deps, newDeps)
	}

	if shouldRun {
		// Run cleanup from previous effect.
		if hook.CleanupRef != 0 {
			if err := L.CallRef(hook.CleanupRef, 0, 0); err != nil {
				// Cleanup failed — ignore but release ref.
			}
			L.Unref(lua.RegistryIndex, hook.CleanupRef)
			hook.CleanupRef = 0
		}

		// Call the effect function.
		L.PushValue(1) // push the effect fn
		if status := L.PCall(0, 1, 0); status != lua.OK {
			L.Pop(1) // pop error
			return 0
		}

		// If effect returned a function, store it as cleanup.
		if L.Type(-1) == lua.TypeFunction {
			hook.CleanupRef = L.Ref(lua.RegistryIndex)
		} else {
			L.Pop(1)
		}

		hook.Deps = newDeps
		hook.Ran = true
	}

	return 0
}

// MemoHook tracks a useMemo call across renders.
type MemoHook struct {
	Deps     []any // dependency values from last computation
	ValueRef int   // Lua registry ref for cached value (0 = none)
	HasValue bool
}

// luaUseMemo implements lumina.useMemo(fn, deps).
// Returns a cached value, only recomputing when deps change.
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

	hook := getMemoHook(comp, b)

	shouldRecompute := false
	if !hook.HasValue {
		shouldRecompute = true
	} else if !hasDeps {
		shouldRecompute = true
	} else {
		shouldRecompute = !depsEqual(hook.Deps, newDeps)
	}

	if shouldRecompute {
		// Release old cached value.
		if hook.ValueRef != 0 {
			L.Unref(lua.RegistryIndex, hook.ValueRef)
			hook.ValueRef = 0
		}

		// Call the factory function.
		L.PushValue(1)
		if status := L.PCall(0, 1, 0); status != lua.OK {
			L.Pop(1) // pop error
			L.PushNil()
			return 1
		}

		// Store result as registry ref.
		L.PushValue(-1) // duplicate for Ref
		hook.ValueRef = L.Ref(lua.RegistryIndex)
		hook.Deps = newDeps
		hook.HasValue = true
		// Result is already on stack.
		return 1
	}

	// Return cached value.
	L.RawGetI(lua.RegistryIndex, int64(hook.ValueRef))
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
			// Note: resultIdx and childrenIdx are on stack, so arg indices shifted.
			// We need to push the original argument value.
			L.PushValue(i)
			L.RawSetI(childrenIdx, ci)
			ci++
		}
		L.SetField(resultIdx, "children")
	}

	return 1
}

// --- Hook storage ---
// Hooks are stored per-component in HookStore() (dedicated map, not Props).

const (
	effectHooksKey   = "_bridge_effect_hooks"
	memoHooksKey     = "_bridge_memo_hooks"
	effectHookIdxKey = "_bridge_effect_idx"
	memoHookIdxKey   = "_bridge_memo_idx"
)

// ResetHookIndices resets hook call indices for a new render pass.
// Must be called before each component render.
func (b *Bridge) ResetHookIndices() {
	if b.currentComp == nil {
		return
	}
	store := b.currentComp.HookStore()
	store[effectHookIdxKey] = 0
	store[memoHookIdxKey] = 0
}

func getEffectHook(comp *component.Component, b *Bridge) *EffectHook {
	store := comp.HookStore()
	hooks, _ := store[effectHooksKey].([]*EffectHook)
	idxVal, _ := store[effectHookIdxKey].(int)
	idx := idxVal

	if idx < len(hooks) {
		store[effectHookIdxKey] = idx + 1
		return hooks[idx]
	}

	// Grow.
	h := &EffectHook{}
	hooks = append(hooks, h)
	store[effectHooksKey] = hooks
	store[effectHookIdxKey] = idx + 1
	return h
}

func getMemoHook(comp *component.Component, b *Bridge) *MemoHook {
	store := comp.HookStore()
	hooks, _ := store[memoHooksKey].([]*MemoHook)
	idxVal, _ := store[memoHookIdxKey].(int)
	idx := idxVal

	if idx < len(hooks) {
		store[memoHookIdxKey] = idx + 1
		return hooks[idx]
	}

	h := &MemoHook{}
	hooks = append(hooks, h)
	store[memoHooksKey] = hooks
	store[memoHookIdxKey] = idx + 1
	return h
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
