// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"os"
	"sync"

	"github.com/akzj/go-lua/pkg/lua"
)

// ModuleName is the name used to require this module from Lua.
const ModuleName = "lumina"

// ComponentRegistry stores registered component factories.
var (
	componentRegistry = make(map[string]*lua.State) // name -> state (for future use)
	registryMu        sync.RWMutex
)

// luaLoader is the module loader function that creates the lumina module table.
func luaLoader(L *lua.State) int {
	// Create module table
	L.NewTable()

	// Register module functions using SetFuncs
	L.SetFuncs(map[string]lua.Function{
		"version": func(L *lua.State) int {
			L.PushString("0.3.0")
			return 1
		},
		"echo": func(L *lua.State) int {
			L.PushValue(1)
			return 1
		},
		"info": func(L *lua.State) int {
			L.NewTableFrom(map[string]any{
				"version":     "0.3.0",
				"description": "Lumina Terminal UI Framework",
				"year":        int64(2024),
			})
			return 1
		},
		
		"defineComponent": defineComponent,
		"createComponent": createComponent,
		"createElement":        createElement,
		"createErrorBoundary": createErrorBoundary,
		"memo":                luaMemo,
		"createPortal":        luaCreatePortal,
		"forwardRef":          luaForwardRef,
		"lazy":                luaLazy,
		"render":          renderComponent,
		"createState":     createState,
		// Style & Theme API
		"defineStyle":        defineStyle,
		"defineGlobalStyles": defineGlobalStyles,
		// UI Components
		"Select":    SelectComponent,
		"Checkbox":  CheckboxComponent,
		"Menu":      MenuComponent,
		"TextField": TextFieldComponent,
		"getStyle":           getStyle,
		"defineTheme":        defineTheme,
		"setTheme":           setTheme,
		// Event API
		"on":                  registerEvent,
		"off":                 unregisterEvent,
		"emit":                emitEvent,
		"registerShortcut":    registerShortcut,
		"setFocus":            setFocus,
		"getFocused":          getFocused,
		"focusNext":           focusNext,
		"focusPrev":           focusPrev,
		"registerFocusable":   registerFocusable,
		"unregisterFocusable": unregisterFocusable,
		"isFocusable":         isFocusable,
		"getFocusableIDs":     getFocusableIDs,
		"emitKeyEvent":        emitKeyEvent,
		// Output mode API
		"setOutputMode":           setOutputMode,
		"getOutputMode":           getOutputMode,
		"getMCPFrame":             getMCPFrame,
		"createComponentRequest":  createComponentRequest,
		"createEventNotification": createEventNotification,
		// MCP DevTools - Inspect
		"inspect":          inspectDispatch,
		"inspectTree":      inspectTree,
		"inspectComponent": inspectComponent,
		"inspectStyles":    inspectStyles,
		"inspectFrames":    inspectFrames,
		"getState":         getState,
		"getAllComponents": getAllComponents,
		// MCP DevTools - Simulate
		"simulate":       simulate,
		"simulateClick":  simulateClick,
		"simulateKey":    simulateKey,
		"simulateChange": simulateChange,
		// MCP DevTools - Console
		"consoleLog":       consoleLog,
		"consoleGet":       consoleGet,
		"consoleGetErrors": consoleGetErrors,
		"consoleClear":     consoleClear,
		"consoleSize":      consoleSize,
		// MCP DevTools - Diff
		"diff":       diff,
		"diffFrames": diffFrames,
		// MCP DevTools - Patch & Eval
		"patch": patchComponent,
		"eval":  eval,
		// MCP DevTools - Profile
		"profile":      profile,
		"profileReset": profileReset,
		"profileSize":  profileSize,
		// Async API
		"useAsync": useAsync,
		"delay":    luminaDelay,
		// Viewport / Scroll API
		"scrollTo":       luaScrollTo,
		"scrollToBottom": luaScrollToBottom,
		"scrollToTop":    luaScrollToTop,
		"scrollBy":       luaScrollBy,
		"getScrollInfo":  luaGetScrollInfo,
		// Text Input API
		"setInputValue":    luaSetInputValue,
		"getInputValue":    luaGetInputValue,
		"registerInput":    luaRegisterInput,
		"focusInput":       luaFocusInput,
	}, 0)

	// Register hooks as sub-table
	L.PushString("hooks")
	L.NewTable()
	RegisterHooks(L)
	L.SetField(-3, "hooks")
	L.Pop(1)

	// Also register common hooks directly on lumina for convenience
	L.SetFuncs(map[string]lua.Function{
		"useState":             useState,
		"useEffect":            useEffect,
		"useLayoutEffect":      useLayoutEffect,
		"useMemo":              useMemo,
		"useCallback":          useCallback,
		"useRef":               useRef,
		"useReducer":           useReducer,
		"useId":                useId,
		"useImperativeHandle":  useImperativeHandle,
		"useTransition":        useTransition,
		"useDeferredValue":     useDeferredValue,
		"createContext":        createContext,
		"useContext":           useContext,
		"setContextValue":      setContextValueLua,
	}, 0)

	// Register lumina.Suspense as a component factory table (not a function)
	registerSuspenseFactory(L)

	// Register console as sub-table
	L.PushString("console")
	L.NewTable()
	L.SetFuncs(map[string]lua.Function{
		"log":   func(L *lua.State) int { L.PushString("log"); return consoleLog(L) },
		"warn":  func(L *lua.State) int { L.PushString("warn"); return consoleLog(L) },
		"error": func(L *lua.State) int { L.PushString("error"); return consoleLog(L) },
		"get":   consoleGet,
		"clear": consoleClear,
		"size":  consoleSize,
	}, 0)
	L.SetField(-3, "console")

	// Register debug as sub-table: lumina.debug.*
	L.NewTable()
	RegisterDebugAPI(L)
	L.SetField(-3, "debug")

	L.Pop(1)

	return 1
}

// defineComponent creates a component factory from a config table.
func defineComponent(L *lua.State) int {
	// defineComponent(config) -> Component

	if L.Type(-1) != lua.TypeTable {
		typeName := L.TypeName(L.Type(-1))
		L.PushString("defineComponent: expected table, got " + typeName)
		L.Error()
		return 0
	}

	L.NewTable() // stack: [config, componentTable]

	// Get component name from config (now at -2)
	L.PushValue(-2)
	L.GetField(-1, "name")
	name, _ := L.ToString(-1)
	L.Remove(-2)

	if name == "" {
		L.Pop(3)
		L.PushString("defineComponent: 'name' is required")
		L.Error()
		return 0
	}

	L.PushString(name)
	L.SetField(-3, "name")
	L.Pop(1)

	// Verify render function exists
	L.PushValue(-2)
	L.GetField(-1, "render")
	L.Remove(-2)
	if L.Type(-1) != lua.TypeFunction {
		L.Pop(3)
		L.PushString("defineComponent: 'render' function is required")
		L.Error()
		return 0
	}
	L.SetField(-3, "render")
	L.Pop(1)

	// Set other fields on componentTable (at -2)
	L.PushString("component:" + name)
	L.SetField(-2, "id")
	L.PushBoolean(true)
	L.SetField(-2, "isComponent")

	// Pop config from stack, leaving componentTable as return value
	L.Remove(-2)

	registryMu.Lock()
	componentRegistry[name] = L
	registryMu.Unlock()

	return 1
}

// createComponent instantiates a component with props.
// createComponent(factory, props) -> componentInstance
func createComponent(L *lua.State) int {
	// Get factory table
	factoryIdx := 1

	// Get props table (optional, defaults to empty)
	props := make(map[string]any)
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		if m, ok := L.ToMap(2); ok {
			props = m
		}
	}

	// Create component instance
	comp, err := NewComponent(L, factoryIdx, props)
	if err != nil {
		L.PushString(fmt.Sprintf("createComponent: %v", err))
		L.Error()
		return 0
	}

	// Call init function if present
	if comp.PushInitFn(L) {
		// Push props table
		L.PushAny(props)

		// Call init(props) -> instanceState
		SetCurrentComponent(comp)
		status := L.PCall(1, lua.MultiRet, 0)
		if status == lua.OK && L.GetTop() > 0 {
			// Init returned a table - merge into component state
			if initState, ok := L.ToMap(-1); ok {
				for k, v := range initState {
					comp.State[k] = v
				}
			}
			L.Pop(1) // pop return value
		} else if L.GetTop() > 0 {
			L.Pop(1) // pop error message
		}
		SetCurrentComponent(nil)
	}

	// Create component instance table for Lua
	L.NewTableFrom(map[string]any{
		"_id":         comp.ID,
		"type":        comp.Type,
		"_isInstance": true,
	})

	// Push component userdata as hidden field
	L.NewUserdata(0, 0)
	L.SetUserdataValue(-1, comp)
	L.SetField(-2, "_comp")
	L.Pop(1) // pop userdata from stack

	return 1
}

// renderComponent renders a component instance.
// renderComponent(componentInstance, props) -> vdom
func renderComponent(L *lua.State) int {
	// Get component instance (passed as argument)
	instanceIdx := 1

	// Get optional props override
	var props map[string]any
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		if m, ok := L.ToMap(2); ok {
			props = m
		}
	}

	// Get component from userdata (if instance has one)
	var comp *Component
	L.GetField(instanceIdx, "_comp")
	if L.Type(-1) == lua.TypeUserdata {
		comp = UserdataToComponent(L, -1)
	}
	L.Pop(1)

	// If no component found, check if this is a factory (create new instance)
	if comp == nil {
		// Check if this is a factory by looking for isComponent=true
		L.GetField(instanceIdx, "isComponent")
		isFactory := L.ToBoolean(-1)
		L.Pop(1)

		if isFactory {
			// This is a factory, create a new component instance
			var err error
			comp, err = NewComponent(L, instanceIdx, props)
			if err != nil {
				L.PushString(fmt.Sprintf("renderComponent: %v", err))
				L.Error()
				return 0
			}
			// Call init function to populate initial state
			if comp.PushInitFn(L) {
				// Push props table
				L.PushAny(props)
				// Call init(props) -> instanceState
				SetCurrentComponent(comp)
				status := L.PCall(1, lua.MultiRet, 0)
				if status == lua.OK && L.GetTop() > 0 {
					if initState, ok := L.ToMap(-1); ok {
						for k, v := range initState {
							comp.State[k] = v
						}
					}
					L.Pop(1)
				} else if L.GetTop() > 0 {
					L.Pop(1)
				}
				SetCurrentComponent(nil)
			}
		} else {
			// This is neither a factory nor an instance - error
			L.PushString("renderComponent: expected component factory or instance, got unknown type")
			L.Error()
			return 0
		}
	}

	// Set as current component for hooks
	SetCurrentComponent(comp)

	// Call render function
	if !comp.PushRenderFn(L) {
		L.PushString("renderComponent: no render function")
		L.Error()
		return 0
	}

	// Create instance table for render
	fields := map[string]any{
		"_instance": comp.ID,
		"_props":    props,
	}
	for k, v := range comp.State {
		fields[k] = v
	}
	L.NewTableFrom(fields)

	// Call render(instance)
	status := L.PCall(1, 1, 0)
	SetCurrentComponent(nil)
	if status != lua.OK {
		msg, _ := L.ToString(-1)
		L.Pop(1)
		L.PushString(fmt.Sprintf("render: %v", msg))
		L.Error()
		return 0
	}

	// vdom is now on stack at -1
	// Convert Lua table to VNode tree.
	newVNode := LuaVNodeToVNode(L, -1)

	// Diff against previous render.
	var frame *Frame
	if comp != nil && comp.LastVNode != nil {
		patches := DiffVNode(comp.LastVNode, newVNode)
		_ = patches // diff available for future incremental updates
		frame = VNodeToFrame(newVNode, 80, 24)
	} else {
		frame = VNodeToFrame(newVNode, 80, 24)
	}
	if comp != nil {
		comp.LastVNode = newVNode
	}

	// Write to terminal
	adapter := GetOutputAdapter()
	if adapter == nil {
		adapter = &NopAdapter{}
	}
	if err := adapter.Write(frame); err != nil {
		// Error writing, but we still return the vdom
	}

	return 1
}

// createState creates a component state container.
func createState(L *lua.State) int {
	fields := map[string]any{
		"state":     "type",
		"stateType": "LuminaState",
	}
	if !L.IsNoneOrNil(1) {
		fields["value"] = L.ToAny(1)
	}
	L.NewTableFrom(fields)
	return 1
}

// Open registers the lumina module into the global namespace AND into package.preload.
func Open(L *lua.State) {
	// Only set default output adapter if none is configured
	if GetOutputAdapter() == nil {
		SetOutputAdapter(NewANSIAdapter(os.Stdout))
	}

	luaLoader(L)
	L.SetGlobal(ModuleName)

	L.GetGlobal("package")
	L.GetField(-1, "preload")
	L.PushFunction(luaLoader)
	L.SetField(-2, ModuleName)
	L.Pop(2)
}

func init() {
	// RegisterGlobal allows require("lumina") to work from any new State
	// without needing Open(L) called first. The opener pushes the module
	// table onto the stack (what require() expects).
	lua.RegisterGlobal(ModuleName, func(L *lua.State) {
		if GetOutputAdapter() == nil {
			SetOutputAdapter(NewANSIAdapter(os.Stdout))
		}
		luaLoader(L)
		// luaLoader pushes the module table — leave it on stack for require()
		// Also set as global for convenience
		L.PushValue(-1) // dup
		L.SetGlobal(ModuleName)
	})
}

// IsComponent checks if the value at idx is a Lumina component.
func IsComponent(L *lua.State, idx int) bool {
	L.GetField(idx, "isComponent")
	if L.IsNone(-1) {
		L.Pop(1)
		return false
	}
	result := L.ToBoolean(-1)
	L.Pop(1)
	return result
}

// GetComponentName returns the name of a component.
func GetComponentName(L *lua.State, idx int) string {
	L.GetField(idx, "name")
	if L.IsNoneOrNil(-1) {
		L.Pop(1)
		return ""
	}
	name, _ := L.ToString(-1)
	L.Pop(1)
	return name
}

// createElement creates a VNode-like table representing a component to be rendered.
// createElement(factory, props) → {type="component", _factory=factory, _props=props}
func createElement(L *lua.State) int {
	// Arg 1: factory table (component definition)
	if L.Type(1) != lua.TypeTable {
		L.PushString("createElement: first argument must be a component factory table")
		L.Error()
		return 0
	}

	L.NewTable()

	// type = "component"
	L.PushString("component")
	L.SetField(-2, "type")

	// _factory = factory (arg 1)
	L.PushValue(1)
	L.SetField(-2, "_factory")

	// _props = props (arg 2, or empty table)
	if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
		L.PushValue(2)
	} else {
		L.NewTable()
	}
	L.SetField(-2, "_props")

	return 1
}

// createErrorBoundary creates an error boundary component factory.
// createErrorBoundary({ fallback = function(err) ... end }) → factory table
func createErrorBoundary(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("createErrorBoundary: expected config table")
		L.Error()
		return 0
	}

	// Create a component factory table
	L.NewTable()

	// name = "ErrorBoundary"
	L.PushString("ErrorBoundary")
	L.SetField(-2, "name")

	// isComponent = true
	L.PushBoolean(true)
	L.SetField(-2, "isComponent")

	// _isErrorBoundary = true
	L.PushBoolean(true)
	L.SetField(-2, "_isErrorBoundary")

	// render = passthrough (renders children)
	L.PushFunction(func(L *lua.State) int {
		// self is arg 1
		L.GetField(1, "children")
		if L.Type(-1) != lua.TypeTable {
			// No children — return empty box
			L.Pop(1)
			L.NewTableFrom(map[string]any{"type": "box"})
			return 1
		}
		// Wrap children in a box
		L.NewTable()
		L.PushString("box")
		L.SetField(-2, "type")
		L.PushValue(-2) // push children
		L.SetField(-2, "children")
		L.Remove(-2) // remove original children
		return 1
	})
	L.SetField(-2, "render")

	// Copy fallback function from config
	L.GetField(1, "fallback")
	if L.Type(-1) == lua.TypeFunction {
		L.SetField(-2, "_fallback")
	} else {
		L.Pop(1)
	}

	return 1
}

// luaMemo wraps a component factory with memoization.
// memo(factory) → memoized factory table
func luaMemo(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("memo: expected component factory table")
		L.Error()
		return 0
	}

	// Create a new factory table that wraps the original
	L.NewTable()

	// Copy name from original
	L.GetField(1, "name")
	if name, ok := L.ToString(-1); ok {
		L.Pop(1)
		L.PushString(name)
	}
	L.SetField(-2, "name")

	// isComponent = true
	L.PushBoolean(true)
	L.SetField(-2, "isComponent")

	// _memoized = true
	L.PushBoolean(true)
	L.SetField(-2, "_memoized")

	// Copy render from original
	L.GetField(1, "render")
	L.SetField(-2, "render")

	// Copy init from original (if exists)
	L.GetField(1, "init")
	if L.Type(-1) == lua.TypeFunction {
		L.SetField(-2, "init")
	} else {
		L.Pop(1)
	}

	// Copy cleanup from original (if exists)
	L.GetField(1, "cleanup")
	if L.Type(-1) == lua.TypeFunction {
		L.SetField(-2, "cleanup")
	} else {
		L.Pop(1)
	}

	return 1
}

// luaCreatePortal creates a portal VNode that renders at a target container.
// createPortal(vnodeTable, targetID) → portal VNode table
func luaCreatePortal(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("createPortal: first argument must be a VNode table")
		L.Error()
		return 0
	}
	targetID := ""
	if L.GetTop() >= 2 {
		targetID = L.CheckString(2)
	}

	L.NewTable()
	L.PushString("portal")
	L.SetField(-2, "type")

	L.PushValue(1) // content VNode
	L.SetField(-2, "_content")

	L.PushString(targetID)
	L.SetField(-2, "_target")

	return 1
}

// luaForwardRef wraps a render function to receive a ref.
// forwardRef(function(props, ref) ... end) → factory table
func luaForwardRef(L *lua.State) int {
	if L.Type(1) != lua.TypeFunction {
		L.PushString("forwardRef: expected render function")
		L.Error()
		return 0
	}

	// Store the render function ref
	L.PushValue(1)
	renderRef := L.Ref(lua.RegistryIndex)

	L.NewTable()
	L.PushString("ForwardRef")
	L.SetField(-2, "name")
	L.PushBoolean(true)
	L.SetField(-2, "isComponent")
	L.PushBoolean(true)
	L.SetField(-2, "_forwardRef")

	// render function: calls the stored function with (props, ref)
	L.PushFunction(func(L *lua.State) int {
		// self is arg 1
		L.RawGetI(lua.RegistryIndex, int64(renderRef))
		if L.Type(-1) != lua.TypeFunction {
			L.Pop(1)
			L.NewTableFrom(map[string]any{"type": "box"})
			return 1
		}

		// Push props (self minus internal fields)
		L.PushValue(1) // push self as props

		// Get ref from self.ref or generate one
		L.GetField(1, "ref")
		if L.IsNoneOrNil(-1) {
			L.Pop(1)
			L.PushString("") // empty ref
		}

		// Call render(props, ref)
		status := L.PCall(2, 1, 0)
		if status != lua.OK {
			msg, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString("forwardRef render error: " + msg)
			L.Error()
			return 0
		}
		return 1
	})
	L.SetField(-2, "render")

	return 1
}

// -----------------------------------------------------------------------
// Suspense + lazy
// -----------------------------------------------------------------------

// registerSuspenseFactory creates the Suspense component factory table
// and sets it as lumina.Suspense on the module table at stack top.
func registerSuspenseFactory(L *lua.State) {
	// lumina table is at -1 on the stack
	L.NewTable() // create the factory table

	L.PushString("Suspense")
	L.SetField(-2, "name")
	L.PushBoolean(true)
	L.SetField(-2, "isComponent")
	L.PushBoolean(true)
	L.SetField(-2, "_isSuspense")

	// render function: check children for pending lazy components
	L.PushFunction(func(L *lua.State) int {
		// self is arg 1 — has .children and .fallback props
		L.GetField(1, "children")
		hasChildren := L.Type(-1) == lua.TypeTable
		L.Pop(1)

		L.GetField(1, "fallback")
		hasFallback := L.Type(-1) == lua.TypeTable
		L.Pop(1)

		// Check if any child is a pending lazy component
		hasPending := false
		if hasChildren {
			L.GetField(1, "children")
			n := int(L.RawLen(-1))
			for i := 1; i <= n; i++ {
				L.RawGetI(-1, int64(i))
				if L.Type(-1) == lua.TypeTable {
					L.GetField(-1, "_lazy_status")
					if status, ok := L.ToString(-1); ok && status == "pending" {
						hasPending = true
					}
					L.Pop(1) // pop _lazy_status
				}
				L.Pop(1) // pop child
			}
			L.Pop(1) // pop children
		}

		if hasPending && hasFallback {
			// Return fallback
			L.GetField(1, "fallback")
			return 1
		}

		// Return box wrapping children
		L.NewTable()
		L.PushString("box")
		L.SetField(-2, "type")
		if hasChildren {
			L.GetField(1, "children")
			L.SetField(-2, "children")
		}
		return 1
	})
	L.SetField(-2, "render")

	// Set as lumina.Suspense
	L.SetField(-2, "Suspense")
}

// luaLazy creates a lazy-loading component wrapper.
// lazy(loaderFn) → factory table
// loaderFn is called on first render; it should return a component factory.
func luaLazy(L *lua.State) int {
	if L.Type(1) != lua.TypeFunction {
		L.PushString("lazy: expected loader function")
		L.Error()
		return 0
	}

	// Store the loader function ref
	L.PushValue(1)
	loaderRef := L.Ref(lua.RegistryIndex)

	// Track loading state
	type lazyState struct {
		status     string // "pending", "resolved", "rejected"
		resolvedRef int   // Lua registry ref to resolved factory
		err        string
	}
	state := &lazyState{status: "pending"}

	// Create a factory table
	L.NewTable()
	L.PushString("Lazy")
	L.SetField(-2, "name")
	L.PushBoolean(true)
	L.SetField(-2, "isComponent")

	// render function
	L.PushFunction(func(L *lua.State) int {
		switch state.status {
		case "pending":
			// Call the loader function
			L.RawGetI(lua.RegistryIndex, int64(loaderRef))
			status := L.PCall(0, 1, 0)
			if status != lua.OK {
				msg, _ := L.ToString(-1)
				L.Pop(1)
				state.status = "rejected"
				state.err = msg
				// Return error text
				L.NewTableFrom(map[string]any{
					"type":    "text",
					"content": "Lazy load error: " + msg,
				})
				return 1
			}

			// Check if result is a component factory
			if L.Type(-1) == lua.TypeTable {
				state.resolvedRef = L.Ref(lua.RegistryIndex)
				state.status = "resolved"
				// Fall through to render the resolved component
			} else {
				L.Pop(1)
				state.status = "rejected"
				state.err = "lazy loader did not return a component factory"
				L.NewTableFrom(map[string]any{
					"type":    "text",
					"content": "Lazy load error: " + state.err,
				})
				return 1
			}
			// Render the resolved component
			L.RawGetI(lua.RegistryIndex, int64(state.resolvedRef))
			L.GetField(-1, "render")
			if L.Type(-1) != lua.TypeFunction {
				L.Pop(2)
				L.NewTableFrom(map[string]any{"type": "box"})
				return 1
			}
			L.PushValue(1) // push self as arg
			renderStatus := L.PCall(1, 1, 0)
			L.Remove(-2) // remove factory table
			if renderStatus != lua.OK {
				msg, _ := L.ToString(-1)
				L.Pop(1)
				L.NewTableFrom(map[string]any{
					"type":    "text",
					"content": "Lazy render error: " + msg,
				})
			}
			return 1

		case "resolved":
			// Render the resolved component
			L.RawGetI(lua.RegistryIndex, int64(state.resolvedRef))
			L.GetField(-1, "render")
			if L.Type(-1) != lua.TypeFunction {
				L.Pop(2)
				L.NewTableFrom(map[string]any{"type": "box"})
				return 1
			}
			L.PushValue(1) // push self as arg
			renderStatus := L.PCall(1, 1, 0)
			L.Remove(-2) // remove factory table
			if renderStatus != lua.OK {
				msg, _ := L.ToString(-1)
				L.Pop(1)
				L.NewTableFrom(map[string]any{
					"type":    "text",
					"content": "Lazy render error: " + msg,
				})
			}
			return 1

		case "rejected":
			L.NewTableFrom(map[string]any{
				"type":    "text",
				"content": "Lazy load error: " + state.err,
			})
			return 1
		}

		L.NewTableFrom(map[string]any{"type": "box"})
		return 1
	})
	L.SetField(-2, "render")

	return 1
}
