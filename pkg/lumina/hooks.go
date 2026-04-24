// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// luaAny is a type alias for Go values in hooks.
type luaAny = interface{}

// useState implements the React-style useState hook.
// Returns [value, setter] on the stack.
func useState(L *lua.State) int {
	// Get current component from registry
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("useState: no current component")
		L.Error()
		return 0
	}

	// Get key (required first argument)
	key := L.CheckString(1)

	// Get initial value (optional second argument)
	var initial any
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		initial = L.ToAny(2)
	}

	// Initialize state if not exists
	comp.mu.Lock()
	if _, exists := comp.State[key]; !exists {
		comp.State[key] = initial
	}
	value := comp.State[key]
	comp.mu.Unlock()

	// Push current value onto stack (return value 1)
	L.PushAny(value)

	// Create setter closure (return value 2)
	// The setter captures 'comp' and 'key' via closure
	L.PushFunction(func(L *lua.State) int {
		// Get new value
		newValue := L.ToAny(1)

		// Update component state
		comp.SetState(key, newValue)

		return 0 // setter doesn't return anything
	})

	return 2 // return [value, setter]
}

// useEffect implements the React-style useEffect hook.
// useEffect(effectFn, deps)
func useEffect(L *lua.State) int {
	// Get current component
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("useEffect: no current component")
		L.Error()
		return 0
	}

	// Arg 1: effect function (table or function)
	L.CheckAny(1)
	if L.Type(1) != lua.TypeFunction {
		L.PushString("useEffect: first argument must be a function")
		L.Error()
		return 0
	}

	// Arg 2: deps table (optional)
	if L.GetTop() >= 2 {
		L.CheckType(2, lua.TypeTable)
	}
	// For now, just call the effect function once

	// Call effect function
	L.PushValue(1) // duplicate effectFn
	status := L.PCall(0, 0, 0)
	if status != lua.OK {
		// Error occurred, push error message
		msg, _ := L.ToString(-1)
		L.Pop(1)
		L.PushString("useEffect error: " + msg)
		L.Error()
		return 0
	}

	return 0 // useEffect doesn't return values
}

// useMemo implements memoized computation.
// useMemo(computeFn, deps)
func useMemo(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("useMemo: no current component")
		L.Error()
		return 0
	}

	// Get compute function - check type manually
	if L.Type(1) != lua.TypeFunction {
		L.PushString("useMemo: first argument must be a function")
		L.Error()
		return 0
	}

	// Get deps table (optional)
	if L.GetTop() >= 2 {
		L.CheckType(2, lua.TypeTable)
	}

	// Simplified: just call the compute function every time
	// A full implementation would cache based on deps
	L.PushValue(1)
	status := L.PCall(0, 1, 0)
	if status != lua.OK {
		msg, _ := L.ToString(-1)
		L.Pop(1)
		L.PushString("useMemo: " + msg)
		L.Error()
		return 0
	}

	return 1
}

// useCallback wraps a function.
// useCallback(fn, deps)
func useCallback(L *lua.State) int {
	// Get function - check type manually
	if L.Type(1) != lua.TypeFunction {
		L.PushString("useCallback: first argument must be a function")
		L.Error()
		return 0
	}

	// Get deps (ignored for now)
	if L.GetTop() >= 2 {
		L.Pop(1) // pop deps
	}

	// Just return the function as-is
	L.PushValue(1)
	return 1
}

// useRef implements a ref container.
// useRef(initial)
func useRef(L *lua.State) int {
	comp := GetCurrentComponent()
	if comp == nil {
		L.PushString("useRef: no current component")
		L.Error()
		return 0
	}

	// Get initial value
	var initial any
	if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
		initial = L.ToAny(1)
	}

	L.NewTableFrom(map[string]any{
		"current": initial,
	})
	return 1
}

// useTheme returns the current theme values.
// Works both inside components (for reactivity) and outside (for one-time reads).
func useTheme(L *lua.State) int {
	// Get current theme from app context (theme is global, not per-component)
	theme := GetCurrentTheme()
	if theme == nil {
		theme = DefaultTheme()
	}

	L.NewTableFrom(map[string]any{
		"name":    theme.Name,
		"colors":  theme.Colors,
		"spacing": theme.Spacing,
		"borders": theme.Borders,
	})
	return 1
}

// RegisterHooks registers all hook functions in the lumina module.
func RegisterHooks(L *lua.State) {
	L.SetFuncs(map[string]lua.Function{
		"useState":    useState,
		"useEffect":   useEffect,
		"useMemo":     useMemo,
		"useCallback": useCallback,
		"useRef":      useRef,
		"useTheme":    useTheme,
	}, 0)
}
