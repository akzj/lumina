// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"sync"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// luaAny is a type alias for Go values in hooks.
type luaAny = interface{}

// -----------------------------------------------------------------------
// Hook state types
// -----------------------------------------------------------------------

// EffectHook tracks a single useEffect call across renders.
type EffectHook struct {
	Deps        []any // dependency values from last run
	CleanupRef  int   // Lua registry ref for cleanup function (0 = none)
	CallbackRef int   // Lua registry ref for the effect callback
	Ran         bool  // whether the effect has ever been run
}

// MemoHook tracks a single useMemo/useCallback call across renders.
type MemoHook struct {
	Deps     []any // dependency values from last computation
	Value    any   // cached Go value (for primitives)
	ValueRef int   // Lua registry ref (for tables/functions)
	HasValue bool  // true after first computation
}

// -----------------------------------------------------------------------
// Context system
// -----------------------------------------------------------------------

// Context represents a React-style context for passing values down the tree.
type Context struct {
	ID           int64
	DefaultValue any
}

var (
	contextCounter  int64
	contextMu       sync.Mutex
	contextValues   = make(map[int64]any)
	contextValuesMu sync.RWMutex
)

// NewContext creates a new context with a default value.
func NewContext(defaultValue any) *Context {
	contextMu.Lock()
	contextCounter++
	id := contextCounter
	contextMu.Unlock()
	return &Context{ID: id, DefaultValue: defaultValue}
}

// SetContextValue sets the current value for a context.
func SetContextValue(ctx *Context, value any) {
	contextValuesMu.Lock()
	contextValues[ctx.ID] = value
	contextValuesMu.Unlock()
}

// GetContextValue returns the current value for a context.
func GetContextValue(ctx *Context) any {
	contextValuesMu.RLock()
	v, ok := contextValues[ctx.ID]
	contextValuesMu.RUnlock()
	if ok {
		return v
	}
	return ctx.DefaultValue
}

// ClearContextValues resets all context values (for testing).
func ClearContextValues() {
	contextValuesMu.Lock()
	contextValues = make(map[int64]any)
	contextValuesMu.Unlock()
}

// -----------------------------------------------------------------------
// Hook implementations
// -----------------------------------------------------------------------

// useState implements the React-style useState hook.
func useState(L *lua.State) int {
	comp := GetCurrentComponent()
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
	comp.mu.Lock()
	if _, exists := comp.State[key]; !exists {
		comp.State[key] = initial
	}
	value := comp.State[key]
	comp.mu.Unlock()
	L.PushAny(value)
	L.PushFunction(func(L *lua.State) int {
		newValue := L.ToAny(1)
		comp.SetState(key, newValue)
		return 0
	})
	return 2
}

// useEffect implements the React-style useEffect hook with dependency tracking.
func useEffect(L *lua.State) int {
	comp := GetCurrentComponent()
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
	var newDeps []any
	hasDeps := false
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		L.CheckType(2, lua.TypeTable)
		hasDeps = true
		newDeps = luaTableToSlice(L, 2)
	}
	comp.mu.Lock()
	hook := comp.nextEffectHookLocked()
	comp.mu.Unlock()

	shouldRun := false
	if !hook.Ran {
		shouldRun = true
	} else if !hasDeps {
		shouldRun = true
	} else {
		shouldRun = !depsEqual(hook.Deps, newDeps)
	}
	if shouldRun {
		if hook.CleanupRef != 0 {
			L.RawGetI(lua.RegistryIndex, int64(hook.CleanupRef))
			if L.Type(-1) == lua.TypeFunction {
				status := L.PCall(0, 0, 0)
				if status != lua.OK {
					L.Pop(1)
				}
			} else {
				L.Pop(1)
			}
			L.Unref(lua.RegistryIndex, hook.CleanupRef)
			hook.CleanupRef = 0
		}
		L.PushValue(1)
		status := L.PCall(0, 1, 0)
		if status != lua.OK {
			msg, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString("useEffect error: " + msg)
			L.Error()
			return 0
		}
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

func (c *Component) nextEffectHookLocked() *EffectHook {
	idx := c.effectHookIndex
	c.effectHookIndex++
	if idx < len(c.effectHooks) {
		return c.effectHooks[idx]
	}
	h := &EffectHook{}
	c.effectHooks = append(c.effectHooks, h)
	return h
}

func (c *Component) nextMemoHookLocked() *MemoHook {
	idx := c.memoHookIndex
	c.memoHookIndex++
	if idx < len(c.memoHooks) {
		return c.memoHooks[idx]
	}
	h := &MemoHook{}
	c.memoHooks = append(c.memoHooks, h)
	return h
}

// useMemo implements memoized computation with dependency tracking.
func useMemo(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("useMemo: no current component")
		L.Error()
		return 0
	}
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
	comp.mu.Lock()
	hook := comp.nextMemoHookLocked()
	comp.mu.Unlock()

	shouldCompute := !hook.HasValue || !hasDeps || !depsEqual(hook.Deps, newDeps)
	if shouldCompute {
		if hook.ValueRef != 0 {
			L.Unref(lua.RegistryIndex, hook.ValueRef)
			hook.ValueRef = 0
		}
		L.PushValue(1)
		status := L.PCall(0, 1, 0)
		if status != lua.OK {
			msg, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString("useMemo: " + msg)
			L.Error()
			return 0
		}
		if L.Type(-1) == lua.TypeTable || L.Type(-1) == lua.TypeFunction {
			L.PushValue(-1)
			hook.ValueRef = L.Ref(lua.RegistryIndex)
			hook.Value = nil
		} else {
			hook.Value = L.ToAny(-1)
		}
		hook.Deps = newDeps
		hook.HasValue = true
		return 1
	}
	if hook.ValueRef != 0 {
		L.RawGetI(lua.RegistryIndex, int64(hook.ValueRef))
	} else {
		L.PushAny(hook.Value)
	}
	return 1
}

// useCallback wraps a function with memoization.
func useCallback(L *lua.State) int {
	if L.Type(1) != lua.TypeFunction {
		L.PushString("useCallback: first argument must be a function")
		L.Error()
		return 0
	}
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushValue(1)
		return 1
	}
	var newDeps []any
	hasDeps := false
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		L.CheckType(2, lua.TypeTable)
		hasDeps = true
		newDeps = luaTableToSlice(L, 2)
	}
	comp.mu.Lock()
	hook := comp.nextMemoHookLocked()
	comp.mu.Unlock()

	shouldCache := !hook.HasValue || !hasDeps || !depsEqual(hook.Deps, newDeps)
	if shouldCache {
		if hook.ValueRef != 0 {
			L.Unref(lua.RegistryIndex, hook.ValueRef)
		}
		L.PushValue(1)
		hook.ValueRef = L.Ref(lua.RegistryIndex)
		hook.Deps = newDeps
		hook.HasValue = true
	}
	L.RawGetI(lua.RegistryIndex, int64(hook.ValueRef))
	return 1
}

// useRef implements a ref container.
func useRef(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("useRef: no current component")
		L.Error()
		return 0
	}
	var initial any
	if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
		initial = L.ToAny(1)
	}
	L.NewTableFrom(map[string]any{"current": initial})
	return 1
}

// useTheme returns the current theme values.
func useTheme(L *lua.State) int {
	theme := GetCurrentTheme()
	if theme == nil {
		theme = DefaultTheme()
	}
	L.NewTableFrom(map[string]any{
		"name": theme.Name, "colors": theme.Colors,
		"spacing": theme.Spacing, "borders": theme.Borders,
	})
	return 1
}

// useReducer implements a reducer-based state hook.
func useReducer(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("useReducer: no current component")
		L.Error()
		return 0
	}
	if L.Type(1) != lua.TypeFunction {
		L.PushString("useReducer: first argument must be a function (reducer)")
		L.Error()
		return 0
	}
	L.CheckAny(2)

	comp.mu.Lock()
	hookIdx := comp.generalHookIndex
	comp.generalHookIndex++
	key := fmt.Sprintf("__reducer_%d", hookIdx)
	if _, exists := comp.State[key]; !exists {
		comp.State[key] = L.ToAny(2)
	}
	currentState := comp.State[key]
	comp.mu.Unlock()

	reducerKey := key + "_ref"
	comp.mu.RLock()
	_, hasRef := comp.State[reducerKey]
	comp.mu.RUnlock()
	if !hasRef {
		L.PushValue(1)
		refID := L.Ref(lua.RegistryIndex)
		comp.mu.Lock()
		comp.State[reducerKey] = refID
		comp.mu.Unlock()
	}

	L.PushAny(currentState)
	L.PushFunction(func(L *lua.State) int {
		action := L.ToAny(1)
		comp.mu.RLock()
		refIDRaw, ok := comp.State[reducerKey]
		comp.mu.RUnlock()
		if !ok {
			return 0
		}
		refID, ok := refIDRaw.(int)
		if !ok {
			return 0
		}
		L.RawGetI(lua.RegistryIndex, int64(refID))
		if L.Type(-1) != lua.TypeFunction {
			L.Pop(1)
			return 0
		}
		comp.mu.RLock()
		curState := comp.State[key]
		comp.mu.RUnlock()
		L.PushAny(curState)
		L.PushAny(action)
		status := L.PCall(2, 1, 0)
		if status != lua.OK {
			L.Pop(1)
			return 0
		}
		newState := L.ToAny(-1)
		L.Pop(1)
		comp.SetState(key, newState)
		return 0
	})
	return 2
}

// createContext creates a new context.
func createContext(L *lua.State) int {
	var defaultValue any
	if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
		defaultValue = L.ToAny(1)
	}
	ctx := NewContext(defaultValue)
	L.NewUserdata(0, 0)
	L.SetUserdataValue(-1, ctx)
	return 1
}

// useContext reads the current value of a context.
func useContext(L *lua.State) int {
	val := L.UserdataValue(1)
	ctx, ok := val.(*Context)
	if !ok {
		L.PushString("useContext: argument must be a context created by createContext")
		L.Error()
		return 0
	}

	// Walk up the component tree to find a provider
	comp := GetCurrentComponent()
	if comp != nil {
		if v, found := resolveContextFromTree(comp, ctx.ID); found {
			L.PushAny(v)
			return 1
		}
	}

	// Fall back to global context value
	value := GetContextValue(ctx)
	L.PushAny(value)
	return 1
}

// resolveContextFromTree walks up the component tree to find a context value.
func resolveContextFromTree(comp *Component, contextID int64) (any, bool) {
	for c := comp; c != nil; c = c.Parent {
		c.mu.RLock()
		if c.ContextValues != nil {
			if val, ok := c.ContextValues[contextID]; ok {
				c.mu.RUnlock()
				return val, true
			}
		}
		c.mu.RUnlock()
	}
	return nil, false
}

// setContextValueLua sets a context's current value.
// If called within a component render, sets on the current component (tree-scoped).
// Otherwise sets globally.
func setContextValueLua(L *lua.State) int {
	val := L.UserdataValue(1)
	ctx, ok := val.(*Context)
	if !ok {
		L.PushString("setContextValue: first argument must be a context")
		L.Error()
		return 0
	}
	var value any
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		value = L.ToAny(2)
	}

	// If we're inside a component render, set on the component (tree-scoped)
	comp := GetCurrentComponent()
	if comp != nil {
		comp.mu.Lock()
		if comp.ContextValues == nil {
			comp.ContextValues = make(map[int64]any)
		}
		comp.ContextValues[ctx.ID] = value
		comp.mu.Unlock()
	}

	// Also set globally for backward compatibility
	SetContextValue(ctx, value)
	return 0
}

// RegisterHooks registers all hook functions in the lumina module.
func RegisterHooks(L *lua.State) {
	L.SetFuncs(map[string]lua.Function{
		"useState": useState, "useEffect": useEffect,
		"useMemo": useMemo, "useCallback": useCallback,
		"useRef": useRef, "useTheme": useTheme,
		"useReducer": useReducer, "createContext": createContext,
		"useContext": useContext, "setContextValue": setContextValueLua,
	}, 0)
}

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

func luaTableToSlice(L *lua.State, idx int) []any {
	var result []any
	L.PushNil()
	for L.Next(idx) {
		result = append(result, L.ToAny(-1))
		L.Pop(1)
	}
	return result
}

func depsEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !shallowEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

// RunEffectCleanups runs all effect cleanup functions for a component.
func RunEffectCleanups(L *lua.State, comp *Component) {
	comp.mu.Lock()
	hooks := comp.effectHooks
	comp.mu.Unlock()
	for _, hook := range hooks {
		if hook.CleanupRef != 0 {
			L.RawGetI(lua.RegistryIndex, int64(hook.CleanupRef))
			if L.Type(-1) == lua.TypeFunction {
				status := L.PCall(0, 0, 0)
				if status != lua.OK {
					L.Pop(1)
				}
			} else {
				L.Pop(1)
			}
			L.Unref(lua.RegistryIndex, hook.CleanupRef)
			hook.CleanupRef = 0
		}
	}
}

// -----------------------------------------------------------------------
// useLayoutEffect — synchronous effect that runs after render, before paint
// -----------------------------------------------------------------------

// LayoutEffectHook tracks a single useLayoutEffect call across renders.
type LayoutEffectHook struct {
	Deps        []any // dependency values from last run
	CleanupRef  int   // Lua registry ref for cleanup function (0 = none)
	CallbackRef int   // Lua registry ref for the effect callback
	Ran         bool  // whether the effect has ever been run
}

// useLayoutEffect registers a layout effect (runs synchronously after render).
// useLayoutEffect(callback, deps?)
func useLayoutEffect(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("useLayoutEffect: no current component")
		L.Error()
		return 0
	}
	L.CheckAny(1)
	if L.Type(1) != lua.TypeFunction {
		L.PushString("useLayoutEffect: first argument must be a function")
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

	comp.mu.Lock()
	hook := comp.nextLayoutEffectHookLocked()
	comp.mu.Unlock()

	shouldRun := false
	if !hook.Ran {
		shouldRun = true
	} else if !hasDeps {
		shouldRun = true
	} else {
		shouldRun = !depsEqual(hook.Deps, newDeps)
	}

	if shouldRun {
		// Run cleanup from previous invocation
		if hook.CleanupRef != 0 {
			L.RawGetI(lua.RegistryIndex, int64(hook.CleanupRef))
			if L.Type(-1) == lua.TypeFunction {
				status := L.PCall(0, 0, 0)
				if status != lua.OK {
					L.Pop(1)
				}
			} else {
				L.Pop(1)
			}
			L.Unref(lua.RegistryIndex, hook.CleanupRef)
			hook.CleanupRef = 0
		}

		// Run the effect callback
		L.PushValue(1)
		status := L.PCall(0, 1, 0)
		if status != lua.OK {
			msg, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString("useLayoutEffect error: " + msg)
			L.Error()
			return 0
		}

		// Store cleanup function if returned
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

func (c *Component) nextLayoutEffectHookLocked() *LayoutEffectHook {
	idx := c.layoutEffectHookIndex
	c.layoutEffectHookIndex++
	if idx < len(c.layoutEffectHooks) {
		return c.layoutEffectHooks[idx]
	}
	hook := &LayoutEffectHook{}
	c.layoutEffectHooks = append(c.layoutEffectHooks, hook)
	return hook
}

// RunLayoutEffectCleanups runs all layout effect cleanups for a component.
func RunLayoutEffectCleanups(L *lua.State, comp *Component) {
	comp.mu.Lock()
	hooks := comp.layoutEffectHooks
	comp.mu.Unlock()
	for _, hook := range hooks {
		if hook.CleanupRef != 0 {
			L.RawGetI(lua.RegistryIndex, int64(hook.CleanupRef))
			if L.Type(-1) == lua.TypeFunction {
				status := L.PCall(0, 0, 0)
				if status != lua.OK {
					L.Pop(1)
				}
			} else {
				L.Pop(1)
			}
			L.Unref(lua.RegistryIndex, hook.CleanupRef)
			hook.CleanupRef = 0
		}
	}
}

// -----------------------------------------------------------------------
// useId — generates a unique, stable ID per component + hook position
// -----------------------------------------------------------------------

func useId(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("")
		return 1
	}

	comp.mu.Lock()
	idx := comp.generalHookIndex
	comp.generalHookIndex++
	comp.mu.Unlock()

	// Generate stable ID based on component ID + hook index
	id := fmt.Sprintf(":r%s:%d:", comp.ID, idx)
	L.PushString(id)
	return 1
}

// -----------------------------------------------------------------------
// useImperativeHandle — customizes the instance value exposed via ref
// -----------------------------------------------------------------------

// useImperativeHandle(ref, createHandle, deps?)
// Sets ref.current to the result of createHandle() when deps change.
func useImperativeHandle(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("useImperativeHandle: no current component")
		L.Error()
		return 0
	}

	// arg1: ref (table with .current field)
	// arg2: createHandle (function)
	// arg3: deps (optional array)
	if L.Type(1) != lua.TypeTable {
		L.PushString("useImperativeHandle: first argument must be a ref table")
		L.Error()
		return 0
	}
	if L.Type(2) != lua.TypeFunction {
		L.PushString("useImperativeHandle: second argument must be a function")
		L.Error()
		return 0
	}

	var newDeps []any
	hasDeps := false
	if L.GetTop() >= 3 && !L.IsNoneOrNil(3) {
		L.CheckType(3, lua.TypeTable)
		hasDeps = true
		newDeps = luaTableToSlice(L, 3)
	}

	comp.mu.Lock()
	hook := comp.nextMemoHookLocked() // Reuse MemoHook for deps tracking
	comp.mu.Unlock()

	shouldRun := false
	if !hook.HasValue {
		shouldRun = true
	} else if !hasDeps {
		shouldRun = true
	} else {
		shouldRun = !depsEqual(hook.Deps, newDeps)
	}

	if shouldRun {
		// Call createHandle()
		L.PushValue(2) // push createHandle function
		status := L.PCall(0, 1, 0)
		if status != lua.OK {
			msg, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString("useImperativeHandle error: " + msg)
			L.Error()
			return 0
		}

		// Set ref.current = result
		L.SetField(1, "current") // ref.current = handle

		hook.Deps = newDeps
		hook.HasValue = true
	}

	return 0
}

// -----------------------------------------------------------------------
// useTransition — marks state updates as non-urgent
// -----------------------------------------------------------------------

// useTransition() → isPending, startTransition
// startTransition(fn) runs fn immediately but marks the component as "transitioning".
// isPending is true while the transition callback is pending.
func useTransition(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushBoolean(false) // isPending
		L.PushFunction(func(L *lua.State) int { return 0 })
		return 2
	}

	comp.mu.Lock()
	idx := comp.generalHookIndex
	comp.generalHookIndex++

	// Use a MemoHook to store isPending state
	var hook *MemoHook
	if idx < len(comp.memoHooks) {
		hook = comp.memoHooks[idx]
	} else {
		hook = &MemoHook{}
		comp.memoHooks = append(comp.memoHooks, hook)
	}
	comp.mu.Unlock()

	// isPending
	isPending := false
	if v, ok := hook.Value.(bool); ok {
		isPending = v
	}
	L.PushBoolean(isPending)

	// startTransition(fn)
	L.PushFunction(func(L *lua.State) int {
		if L.Type(1) != lua.TypeFunction {
			return 0
		}

		// Mark as pending
		comp.mu.Lock()
		hook.Value = true
		hook.HasValue = true
		comp.mu.Unlock()

		// Run the callback immediately (synchronous for simplicity)
		L.PushValue(1)
		status := L.PCall(0, 0, 0)
		if status != lua.OK {
			L.Pop(1) // pop error
		}

		// Mark as no longer pending
		comp.mu.Lock()
		hook.Value = false
		comp.mu.Unlock()

		return 0
	})

	return 2
}

// -----------------------------------------------------------------------
// useDeferredValue — returns a deferred version of a value
// -----------------------------------------------------------------------

// useDeferredValue(value) → deferredValue
// On first render, returns the value immediately.
// On subsequent renders, returns the previous value (lags by one render).
// The deferred value catches up on the next render cycle.
func useDeferredValue(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushValue(1) // return value as-is
		return 1
	}

	comp.mu.Lock()
	idx := comp.generalHookIndex
	comp.generalHookIndex++

	var hook *MemoHook
	if idx < len(comp.memoHooks) {
		hook = comp.memoHooks[idx]
	} else {
		hook = &MemoHook{}
		comp.memoHooks = append(comp.memoHooks, hook)
	}
	comp.mu.Unlock()

	// Get current value from Lua
	var currentValue any
	switch L.Type(1) {
	case lua.TypeString:
		currentValue, _ = L.ToString(1)
	case lua.TypeNumber:
		if v, ok := L.ToInteger(1); ok {
			currentValue = v
		} else if v, ok := L.ToNumber(1); ok {
			currentValue = v
		}
	case lua.TypeBoolean:
		currentValue = L.ToBoolean(1)
	default:
		// For non-primitive types, just pass through
		L.PushValue(1)
		return 1
	}

	if !hook.HasValue {
		// First render — return current value, store it
		hook.Value = currentValue
		hook.HasValue = true
		L.PushValue(1)
		return 1
	}

	// Subsequent renders — return previous value, then update stored value
	previousValue := hook.Value
	hook.Value = currentValue

	// Push the previous (deferred) value
	L.PushAny(previousValue)
	return 1
}

// -----------------------------------------------------------------------
// useSyncExternalStore — subscribe to external stores
// -----------------------------------------------------------------------

// ExternalStoreHook stores subscription state for useSyncExternalStore.
type ExternalStoreHook struct {
	LastSnapshot any // last snapshot value returned to the component
	Subscribed   bool
}

// useSyncExternalStore subscribes to an external store.
// Args: subscribe(callback) → unsubscribe, getSnapshot() → value
func useSyncExternalStore(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("useSyncExternalStore: no current component")
		L.Error()
		return 0
	}

	comp.mu.Lock()
	idx := comp.generalHookIndex
	comp.generalHookIndex++

	// Grow externalStoreHooks if needed
	for idx >= len(comp.externalStoreHooks) {
		comp.externalStoreHooks = append(comp.externalStoreHooks, &ExternalStoreHook{})
	}
	hook := comp.externalStoreHooks[idx]
	comp.mu.Unlock()

	// arg1 = subscribe function (not called here — subscription is app-level)
	// arg2 = getSnapshot function

	// Call getSnapshot() to get current value
	L.PushValue(2) // getSnapshot
	if status := L.PCall(0, 1, 0); status != 0 {
		msg, _ := L.ToString(-1)
		L.Pop(1)
		L.PushString(fmt.Sprintf("useSyncExternalStore: getSnapshot error: %s", msg))
		L.Error()
		return 0
	}

	// Get the snapshot value
	snapshot := luaToAny(L, -1)
	L.Pop(1)

	// On first render, call subscribe with a no-op callback
	// (real subscription management would be done at the app level)
	if !hook.Subscribed {
		hook.Subscribed = true
		hook.LastSnapshot = snapshot

		// Call subscribe(callback) — callback is a no-op for now
		// In a real app, the callback would mark the component dirty
		L.PushValue(1) // subscribe function
		L.PushFunction(func(L *lua.State) int {
			// Notification callback — mark component dirty
			comp.Dirty.Store(true)
			return 0
		})
		if status := L.PCall(1, 1, 0); status != 0 {
			// subscribe may fail — that's OK, just ignore
			L.Pop(1)
		} else {
			// subscribe returns an unsubscribe function — store as cleanup ref
			// For now, just pop it (cleanup would be handled on unmount)
			L.Pop(1)
		}
	} else {
		hook.LastSnapshot = snapshot
	}

	// Return the current snapshot
	L.PushAny(snapshot)
	return 1
}

// luaToAny converts a Lua value at the given stack index to a Go value.
func luaToAny(L *lua.State, idx int) any {
	switch L.Type(idx) {
	case lua.TypeNil:
		return nil
	case lua.TypeBoolean:
		return L.ToBoolean(idx)
	case lua.TypeNumber:
		if n, ok := L.ToInteger(idx); ok {
			return n
		}
		n, _ := L.ToNumber(idx)
		return n
	case lua.TypeString:
		s, _ := L.ToString(idx)
		return s
	case lua.TypeTable:
		// Convert table to map
		result := make(map[string]any)
		L.PushNil()
		for L.Next(idx - 1) {
			key, ok := L.ToString(-2)
			if ok {
				result[key] = luaToAny(L, -1)
			}
			L.Pop(1) // pop value, keep key
		}
		return result
	default:
		return nil
	}
}

// -----------------------------------------------------------------------
// useDebugValue — debug labels for hooks
// -----------------------------------------------------------------------

// useDebugValue stores a debug label on the current component for inspection.
// Args: value (any), optional formatFn
func useDebugValue(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		// Silent no-op if no component (like React)
		return 0
	}

	comp.mu.Lock()
	defer comp.mu.Unlock()

	var label string
	if L.GetTop() >= 2 && L.IsFunction(2) {
		// formatFn provided: call formatFn(value) → formatted string
		L.PushValue(2) // formatFn
		L.PushValue(1) // value
		if status := L.PCall(1, 1, 0); status == 0 {
			label, _ = L.ToString(-1)
			L.Pop(1)
		} else {
			L.Pop(1) // pop error message
		}
	} else {
		label, _ = L.ToString(1)
	}

	comp.debugValues = append(comp.debugValues, label)
	return 0
}

// GetDebugValues returns the debug values for a component.
func (c *Component) GetDebugValues() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.debugValues))
	copy(result, c.debugValues)
	return result
}

// -----------------------------------------------------------------------
// Profiler — render performance measurement
// -----------------------------------------------------------------------

// ProfilerData holds timing data for a Profiler boundary.
type ProfilerData struct {
	ID             string
	Phase          string  // "mount" or "update"
	ActualDuration float64 // milliseconds
	BaseDuration   float64 // milliseconds (estimated without memoization)
	StartTime      float64 // milliseconds since epoch
	CommitTime     float64 // milliseconds since epoch
}

// -----------------------------------------------------------------------
// StrictMode — double-render detection
// -----------------------------------------------------------------------

// strictModeEnabled is a per-render flag that enables double-rendering.
var strictModeEnabled bool

// IsStrictMode returns whether strict mode is currently active.
func IsStrictMode() bool {
	return strictModeEnabled
}

// SetStrictMode enables or disables strict mode.
func SetStrictMode(enabled bool) {
	strictModeEnabled = enabled
}

// useAnimation implements the React-style animation hook.
// Usage: local anim = lumina.useAnimation({ from=0, to=1, duration=300, easing="easeInOut", loop=false })
// Returns: { value = <current>, done = <bool> }
func useAnimation(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		// No component context — return a static table
		L.NewTable()
		L.PushNumber(0)
		L.SetField(-2, "value")
		L.PushBoolean(true)
		L.SetField(-2, "done")
		return 1
	}

	idx := comp.generalHookIndex
	comp.generalHookIndex++

	// Parse config from Lua table argument
	from := 0.0
	to := 1.0
	duration := int64(300)
	easingName := "linear"
	loop := false

	if L.Type(1) == lua.TypeTable {
		L.GetField(1, "from")
		if !L.IsNoneOrNil(-1) {
			from, _ = L.ToNumber(-1)
		}
		L.Pop(1)

		L.GetField(1, "to")
		if !L.IsNoneOrNil(-1) {
			to, _ = L.ToNumber(-1)
		}
		L.Pop(1)

		L.GetField(1, "duration")
		if !L.IsNoneOrNil(-1) {
			d, _ := L.ToNumber(-1)
			duration = int64(d)
		}
		L.Pop(1)

		L.GetField(1, "easing")
		if !L.IsNoneOrNil(-1) {
			easingName, _ = L.ToString(-1)
		}
		L.Pop(1)

		L.GetField(1, "loop")
		if !L.IsNoneOrNil(-1) {
			loop = L.ToBoolean(-1)
		}
		L.Pop(1)
	}

	// On first render for this hook index, create the animation
	animID := fmt.Sprintf("%s:anim:%d", comp.ID, idx)

	if idx >= len(comp.animationHooks) {
		anim := &AnimationState{
			ID:        animID,
			StartTime: timeNowMs(),
			Duration:  duration,
			From:      from,
			To:        to,
			Current:   from,
			Easing:    easingByName(easingName),
			Loop:      loop,
			CompID:    comp.ID,
		}
		globalAnimationManager.Start(anim)
		comp.animationHooks = append(comp.animationHooks, animID)
	}

	// Get current animation state
	anim := globalAnimationManager.Get(animID)
	currentVal := from
	done := false
	if anim != nil {
		currentVal = anim.Current
		done = anim.Done
	}

	// Return { value = current, done = done }
	L.NewTable()
	L.PushNumber(currentVal)
	L.SetField(-2, "value")
	L.PushBoolean(done)
	L.SetField(-2, "done")
	return 1
}

// timeNowMs returns the current time in milliseconds.
// This is a variable so tests can override it.
var timeNowMs = func() int64 {
	return time.Now().UnixMilli()
}

// useRoute returns the current route path and parameters.
// Usage: local route = lumina.useRoute()
// Returns: { path = "/users/123", params = { id = "123" } }
func useRoute(L *lua.State) int {
	router := globalRouter

	L.NewTable()

	// path
	L.PushString(router.GetCurrentPath())
	L.SetField(-2, "path")

	// params as a table
	params := router.GetParams()
	L.NewTable()
	for k, v := range params {
		L.PushString(v)
		L.SetField(-2, k)
	}
	L.SetField(-2, "params")

	return 1
}
