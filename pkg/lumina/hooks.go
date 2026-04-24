// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"sync"

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
	idx := c.hookIndex
	c.hookIndex++
	if idx < len(c.effectHooks) {
		return c.effectHooks[idx]
	}
	h := &EffectHook{}
	c.effectHooks = append(c.effectHooks, h)
	return h
}

func (c *Component) nextMemoHookLocked() *MemoHook {
	idx := c.hookIndex
	c.hookIndex++
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
	hookIdx := comp.hookIndex
	comp.hookIndex++
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
	value := GetContextValue(ctx)
	L.PushAny(value)
	return 1
}

// setContextValueLua sets a context's current value.
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
