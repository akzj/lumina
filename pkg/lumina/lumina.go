// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"os"
	"sync"
	"time"

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
		"onCapture":           registerCaptureEvent,
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
		// Overlay API
		// Overlay API
		"showOverlay":      luaShowOverlay,
		"hideOverlay":      luaHideOverlay,
		"isOverlayVisible": luaIsOverlayVisible,
		"toggleOverlay":    luaToggleOverlay,
		// Hot Reload API
		"enableHotReload":  luaEnableHotReload,
		"disableHotReload": luaDisableHotReload,
		// Focus Scope API
		"pushFocusScope":   luaPushFocusScope,
		"popFocusScope":    luaPopFocusScope,
		// Router API
		"createRouter":     luaCreateRouter,
		"navigate":         luaNavigate,
		"back":             luaBack,
		"useRoute":         useRoute,
		"getCurrentPath":   luaGetCurrentPath,
		// Scroll behavior API
		"setScrollBehavior": luaSetScrollBehavior,
		// SubPixel Canvas API
		"createCanvas":      luaCreateCanvas,
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
		"useTransition":          useTransition,
		"useDeferredValue":       useDeferredValue,
		"useSyncExternalStore":   useSyncExternalStore,
		"useDebugValue":          useDebugValue,
		"createContext":          createContext,
		"useContext":             useContext,
		"setContextValue":        setContextValueLua,
		// Animation hooks
		"useAnimation":           useAnimation,
		"startAnimation":         luaStartAnimation,
		"stopAnimation":          luaStopAnimation,
		// State management
		"createStore":            luaCreateStore,
		"useStore":               luaUseStore,
		// Theme system (extends existing setTheme/defineTheme in style_api.go)
		"useTheme":               luaUseTheme,
		"defineStyles":           luaDefineStyles,
		"getThemeColor":          luaGetThemeColor,
		// Web runtime
		"serve":                  luaServe,
		"serveBackground":        luaServeBackground,
		// i18n
		"useTranslation":         luaUseTranslation,
		// Data fetching
		"fetch":                  luaFetch,
		"useFetch":               luaUseFetch,
		"useQuery":               luaUseQuery,
		"invalidateQuery":        luaInvalidateQuery,
		"invalidateAllQueries":   luaInvalidateAllQueries,
		// Accessibility
		"announce":               luaAnnounce,
		// Testing utilities
		"createTestRenderer":     luaCreateTestRenderer,
		// Grid + Virtual scrolling
		"createVirtualList":      luaCreateVirtualList,
		// Form validation
		"useForm":                luaUseForm,
	}, 0)

	// Register lumina.animation sub-table with preset factories
	registerAnimationPresets(L)

	// Register lumina.i18n sub-table
	registerI18nModule(L)

	// Register lumina.devtools sub-table
	registerDevToolsModule(L)

	// Register lumina.Suspense as a component factory table (not a function)
	registerSuspenseFactory(L)

	// Register lumina.Profiler as a component factory
	registerProfilerFactory(L)

	// Register lumina.StrictMode as a component factory
	registerStrictModeFactory(L)

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

	// Register shadcn component preloads
	RegisterShadcn(L)
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

// -----------------------------------------------------------------------
// Profiler factory
// -----------------------------------------------------------------------

// registerProfilerFactory creates the Profiler component factory table
// and sets it as lumina.Profiler on the module table at stack top.
func registerProfilerFactory(L *lua.State) {
	L.NewTable()
	L.PushString("Profiler")
	L.SetField(-2, "name")
	L.PushBoolean(true)
	L.SetField(-2, "isComponent")
	L.PushBoolean(true)
	L.SetField(-2, "_isProfiler")

	// render function: returns a profiler VNode
	L.PushFunction(func(L *lua.State) int {
		// self is arg 1 — has .id, .onRender, .children
		L.NewTable()
		L.PushString("profiler")
		L.SetField(-2, "type")

		// Copy id
		L.GetField(1, "id")
		L.SetField(-2, "id")

		// Copy onRender callback
		L.GetField(1, "onRender")
		L.SetField(-2, "onRender")

		// Copy children
		L.GetField(1, "children")
		L.SetField(-2, "children")

		return 1
	})
	L.SetField(-2, "render")

	L.SetField(-2, "Profiler")
}

// -----------------------------------------------------------------------
// StrictMode factory
// -----------------------------------------------------------------------

// registerStrictModeFactory creates the StrictMode component factory table
// and sets it as lumina.StrictMode on the module table at stack top.
func registerStrictModeFactory(L *lua.State) {
	L.NewTable()
	L.PushString("StrictMode")
	L.SetField(-2, "name")
	L.PushBoolean(true)
	L.SetField(-2, "isComponent")
	L.PushBoolean(true)
	L.SetField(-2, "_isStrictMode")

	// render function: returns a strictmode VNode
	L.PushFunction(func(L *lua.State) int {
		// self is arg 1 — has .children
		L.NewTable()
		L.PushString("strictmode")
		L.SetField(-2, "type")

		// Copy children
		L.GetField(1, "children")
		L.SetField(-2, "children")

		return 1
	})
	L.SetField(-2, "render")

	L.SetField(-2, "StrictMode")
}

// -----------------------------------------------------------------------
// Overlay Lua API
// -----------------------------------------------------------------------

// luaShowOverlay shows an overlay.
// lumina.showOverlay({ id="...", content={...}, x=N, y=N, width=N, height=N, zIndex=N, modal=bool })
func luaShowOverlay(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("showOverlay: argument must be a table")
		L.Error()
		return 0
	}

	// Extract fields
	L.GetField(1, "id")
	id, _ := L.ToString(-1)
	L.Pop(1)
	if id == "" {
		L.PushString("showOverlay: 'id' is required")
		L.Error()
		return 0
	}

	L.GetField(1, "x")
	x, _ := L.ToInteger(-1)
	L.Pop(1)

	L.GetField(1, "y")
	y, _ := L.ToInteger(-1)
	L.Pop(1)

	L.GetField(1, "width")
	w, _ := L.ToInteger(-1)
	L.Pop(1)
	if w <= 0 {
		w = 20 // default width
	}

	L.GetField(1, "height")
	h, _ := L.ToInteger(-1)
	L.Pop(1)
	if h <= 0 {
		h = 10 // default height
	}

	L.GetField(1, "zIndex")
	zIndex, _ := L.ToInteger(-1)
	L.Pop(1)

	L.GetField(1, "modal")
	modal := L.ToBoolean(-1)
	L.Pop(1)

	// Get content VNode
	L.GetField(1, "content")
	var vnode *VNode
	if L.Type(-1) == lua.TypeTable {
		vnode = LuaVNodeToVNode(L, -1)
	}
	L.Pop(1)

	overlay := &Overlay{
		ID:      id,
		VNode:   vnode,
		X:       int(x),
		Y:       int(y),
		W:       int(w),
		H:       int(h),
		ZIndex:  int(zIndex),
		Visible: true,
		Modal:   modal,
	}

	globalOverlayManager.Show(overlay)
	return 0
}

// luaHideOverlay hides an overlay by ID.
// lumina.hideOverlay("my-dialog")
func luaHideOverlay(L *lua.State) int {
	id := L.CheckString(1)
	globalOverlayManager.Hide(id)
	return 0
}

// luaIsOverlayVisible checks if an overlay is visible.
// lumina.isOverlayVisible("my-dialog") → bool
func luaIsOverlayVisible(L *lua.State) int {
	id := L.CheckString(1)
	L.PushBoolean(globalOverlayManager.IsVisible(id))
	return 1
}

// luaToggleOverlay toggles an overlay's visibility.
// lumina.toggleOverlay("my-dialog") → bool (new state)
func luaToggleOverlay(L *lua.State) int {
	id := L.CheckString(1)
	newState := globalOverlayManager.Toggle(id)
	L.PushBoolean(newState)
	return 1
}

// -----------------------------------------------------------------------
// Animation Lua API
// -----------------------------------------------------------------------

// luaStartAnimation starts a named animation imperatively.
// lumina.startAnimation({ id="fade", from=0, to=1, duration=300, easing="easeInOut", loop=false })
func luaStartAnimation(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("startAnimation: argument must be a table")
		L.Error()
		return 0
	}

	L.GetField(1, "id")
	id, _ := L.ToString(-1)
	L.Pop(1)
	if id == "" {
		L.PushString("startAnimation: 'id' is required")
		L.Error()
		return 0
	}

	from := 0.0
	to := 1.0
	duration := int64(300)
	easingName := "linear"
	loop := false

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

	anim := &AnimationState{
		ID:        id,
		StartTime: timeNowMs(),
		Duration:  duration,
		From:      from,
		To:        to,
		Current:   from,
		Easing:    easingByName(easingName),
		Loop:      loop,
	}
	globalAnimationManager.Start(anim)
	return 0
}

// luaStopAnimation stops an animation by ID.
// lumina.stopAnimation("fade")
func luaStopAnimation(L *lua.State) int {
	id := L.CheckString(1)
	globalAnimationManager.Stop(id)
	return 0
}

// registerAnimationPresets creates the lumina.animation sub-table with preset factories.
// Each preset returns a config table suitable for useAnimation.
func registerAnimationPresets(L *lua.State) {
	// lumina is at top of stack (-1) during luaLoader
	L.PushString("animation")
	L.NewTable()

	// lumina.animation.fadeIn(duration) → { from=0, to=1, duration=N, easing="easeInOut" }
	L.SetFuncs(map[string]lua.Function{
		"fadeIn": func(L *lua.State) int {
			dur := int64(300)
			if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
				d, _ := L.ToNumber(1)
				dur = int64(d)
			}
			L.NewTable()
			L.PushNumber(0)
			L.SetField(-2, "from")
			L.PushNumber(1)
			L.SetField(-2, "to")
			L.PushNumber(float64(dur))
			L.SetField(-2, "duration")
			L.PushString("easeInOut")
			L.SetField(-2, "easing")
			return 1
		},
		"fadeOut": func(L *lua.State) int {
			dur := int64(300)
			if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
				d, _ := L.ToNumber(1)
				dur = int64(d)
			}
			L.NewTable()
			L.PushNumber(1)
			L.SetField(-2, "from")
			L.PushNumber(0)
			L.SetField(-2, "to")
			L.PushNumber(float64(dur))
			L.SetField(-2, "duration")
			L.PushString("easeInOut")
			L.SetField(-2, "easing")
			return 1
		},
		"pulse": func(L *lua.State) int {
			dur := int64(1000)
			if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
				d, _ := L.ToNumber(1)
				dur = int64(d)
			}
			L.NewTable()
			L.PushNumber(0)
			L.SetField(-2, "from")
			L.PushNumber(1)
			L.SetField(-2, "to")
			L.PushNumber(float64(dur))
			L.SetField(-2, "duration")
			L.PushString("easeInOut")
			L.SetField(-2, "easing")
			L.PushBoolean(true)
			L.SetField(-2, "loop")
			return 1
		},
		"spin": func(L *lua.State) int {
			dur := int64(1000)
			if L.GetTop() >= 1 && !L.IsNoneOrNil(1) {
				d, _ := L.ToNumber(1)
				dur = int64(d)
			}
			L.NewTable()
			L.PushNumber(0)
			L.SetField(-2, "from")
			L.PushNumber(360)
			L.SetField(-2, "to")
			L.PushNumber(float64(dur))
			L.SetField(-2, "duration")
			L.PushString("linear")
			L.SetField(-2, "easing")
			L.PushBoolean(true)
			L.SetField(-2, "loop")
			return 1
		},
	}, 0)

	L.SetTable(-3) // lumina.animation = table
}

// -----------------------------------------------------------------------
// Hot Reload Lua API
// -----------------------------------------------------------------------

// luaEnableHotReload enables hot reload with optional config.
// lumina.enableHotReload({ paths = {"app.lua"}, interval = 500 })
func luaEnableHotReload(L *lua.State) int {
	globalHotReloader.Enable(true)

	if L.Type(1) == lua.TypeTable {
		L.GetField(1, "interval")
		if !L.IsNoneOrNil(-1) {
			ms, _ := L.ToNumber(-1)
			if ms > 0 {
				globalHotReloader.config.Interval = time.Duration(ms) * time.Millisecond
			}
		}
		L.Pop(1)

		L.GetField(1, "paths")
		if L.Type(-1) == lua.TypeTable {
			var paths []string
			n := int(L.RawLen(-1))
			for i := 1; i <= n; i++ {
				L.RawGetI(-1, int64(i))
				if s, ok := L.ToString(-1); ok {
					paths = append(paths, s)
				}
				L.Pop(1)
			}
			globalHotReloader.config.WatchPaths = paths
		}
		L.Pop(1)
	}

	return 0
}

// luaDisableHotReload disables hot reload.
// lumina.disableHotReload()
func luaDisableHotReload(L *lua.State) int {
	globalHotReloader.Enable(false)
	return 0
}

// -----------------------------------------------------------------------
// Focus Scope Lua API
// -----------------------------------------------------------------------

// luaPushFocusScope pushes a new focus scope.
// lumina.pushFocusScope({ focusableIDs = {"input1", "btn-ok", "btn-cancel"} })
func luaPushFocusScope(L *lua.State) int {
	scope := &FocusScope{}

	if L.Type(1) == lua.TypeTable {
		L.GetField(1, "id")
		if !L.IsNoneOrNil(-1) {
			scope.ID, _ = L.ToString(-1)
		}
		L.Pop(1)

		L.GetField(1, "focusableIDs")
		if L.Type(-1) == lua.TypeTable {
			n := int(L.RawLen(-1))
			for i := 1; i <= n; i++ {
				L.RawGetI(-1, int64(i))
				if s, ok := L.ToString(-1); ok {
					scope.FocusableIDs = append(scope.FocusableIDs, s)
				}
				L.Pop(1)
			}
		}
		L.Pop(1)
	}

	PushFocusScope(scope)
	return 0
}

// luaPopFocusScope pops the top focus scope.
// lumina.popFocusScope()
func luaPopFocusScope(L *lua.State) int {
	PopFocusScope()
	return 0
}

// -----------------------------------------------------------------------
// Router Lua API
// -----------------------------------------------------------------------

// luaCreateRouter creates a router with route definitions.
// lumina.createRouter({ routes = { {path="/"}, {path="/users/:id"} } })
func luaCreateRouter(L *lua.State) int {
	// Reset global router
	globalRouter = NewRouter()

	if L.Type(1) == lua.TypeTable {
		L.GetField(1, "routes")
		if L.Type(-1) == lua.TypeTable {
			n := int(L.RawLen(-1))
			for i := 1; i <= n; i++ {
				L.RawGetI(-1, int64(i))
				if L.Type(-1) == lua.TypeTable {
					L.GetField(-1, "path")
					if path, ok := L.ToString(-1); ok {
						globalRouter.AddRoute(path)
					}
					L.Pop(1)
				}
				L.Pop(1)
			}
		}
		L.Pop(1)

		// Optional initial path
		L.GetField(1, "initialPath")
		if !L.IsNoneOrNil(-1) {
			if path, ok := L.ToString(-1); ok {
				globalRouter.Navigate(path)
			}
		}
		L.Pop(1)
	}

	// Return the router as a lightweight handle (table with route count)
	L.NewTable()
	L.PushNumber(float64(globalRouter.RouteCount()))
	L.SetField(-2, "routeCount")
	return 1
}

// luaNavigate navigates to a new path.
// lumina.navigate("/users/123")
func luaNavigate(L *lua.State) int {
	path := L.CheckString(1)
	globalRouter.Navigate(path)

	// Mark all components dirty to re-render with new route
	globalRegistry.mu.RLock()
	for _, comp := range globalRegistry.components {
		comp.Dirty.Store(true)
	}
	globalRegistry.mu.RUnlock()

	return 0
}

// luaBack navigates back in history.
// lumina.back() → bool
func luaBack(L *lua.State) int {
	ok := globalRouter.Back()
	L.PushBoolean(ok)

	if ok {
		// Mark all components dirty
		globalRegistry.mu.RLock()
		for _, comp := range globalRegistry.components {
			comp.Dirty.Store(true)
		}
		globalRegistry.mu.RUnlock()
	}

	return 1
}

// luaGetCurrentPath returns the current route path.
// lumina.getCurrentPath() → string
func luaGetCurrentPath(L *lua.State) int {
	L.PushString(globalRouter.GetCurrentPath())
	return 1
}

// -----------------------------------------------------------------------
// Scroll Behavior Lua API
// -----------------------------------------------------------------------

// scrollBehavior controls whether scrolling is "instant" or "smooth".
var scrollBehavior = "instant" // default for backward compat

// GetScrollBehavior returns the current scroll behavior.
func GetScrollBehavior() string { return scrollBehavior }

// luaSetScrollBehavior sets the scroll behavior.
// lumina.setScrollBehavior("smooth") or lumina.setScrollBehavior("instant")
func luaSetScrollBehavior(L *lua.State) int {
	behavior := L.CheckString(1)
	if behavior == "smooth" || behavior == "instant" {
		scrollBehavior = behavior
	}
	return 0
}

// -----------------------------------------------------------------------
// SubPixel Canvas Lua API
// -----------------------------------------------------------------------

// luaCreateCanvas creates a SubPixelCanvas.
// lumina.createCanvas(cellW, cellH) → canvas userdata with methods
func luaCreateCanvas(L *lua.State) int {
	cellW := int(L.CheckInteger(1))
	cellH := int(L.CheckInteger(2))
	if cellW <= 0 {
		cellW = 1
	}
	if cellH <= 0 {
		cellH = 1
	}

	canvas := NewSubPixelCanvas(cellW, cellH)

	// Return as a table with methods
	L.NewTable()

	// Store canvas pointer as light userdata
	L.PushAny(canvas)
	L.SetField(-2, "_canvas")

	L.PushNumber(float64(canvas.CellW))
	L.SetField(-2, "width")
	L.PushNumber(float64(canvas.CellH))
	L.SetField(-2, "height")
	L.PushNumber(float64(canvas.PixW))
	L.SetField(-2, "pixelWidth")
	L.PushNumber(float64(canvas.PixH))
	L.SetField(-2, "pixelHeight")

	// setPixel(x, y, color)
	L.PushFunction(func(L *lua.State) int {
		x := int(L.CheckInteger(1))
		y := int(L.CheckInteger(2))
		hex := L.CheckString(3)
		canvas.SetPixel(x, y, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "setPixel")

	// drawLine(x1, y1, x2, y2, color)
	L.PushFunction(func(L *lua.State) int {
		x1 := int(L.CheckInteger(1))
		y1 := int(L.CheckInteger(2))
		x2 := int(L.CheckInteger(3))
		y2 := int(L.CheckInteger(4))
		hex := L.CheckString(5)
		canvas.DrawLine(x1, y1, x2, y2, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "drawLine")

	// drawRect(x, y, w, h, color)
	L.PushFunction(func(L *lua.State) int {
		x := int(L.CheckInteger(1))
		y := int(L.CheckInteger(2))
		w := int(L.CheckInteger(3))
		h := int(L.CheckInteger(4))
		hex := L.CheckString(5)
		canvas.DrawRect(x, y, w, h, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "drawRect")

	// fillRect(x, y, w, h, color)
	L.PushFunction(func(L *lua.State) int {
		x := int(L.CheckInteger(1))
		y := int(L.CheckInteger(2))
		w := int(L.CheckInteger(3))
		h := int(L.CheckInteger(4))
		hex := L.CheckString(5)
		canvas.FillRect(x, y, w, h, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "fillRect")

	// drawCircle(cx, cy, r, color)
	L.PushFunction(func(L *lua.State) int {
		cx := int(L.CheckInteger(1))
		cy := int(L.CheckInteger(2))
		r := int(L.CheckInteger(3))
		hex := L.CheckString(4)
		canvas.DrawCircle(cx, cy, r, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "drawCircle")

	// drawRoundedRect(x, y, w, h, radius, color)
	L.PushFunction(func(L *lua.State) int {
		x := int(L.CheckInteger(1))
		y := int(L.CheckInteger(2))
		w := int(L.CheckInteger(3))
		h := int(L.CheckInteger(4))
		radius := int(L.CheckInteger(5))
		hex := L.CheckString(6)
		canvas.DrawRoundedRect(x, y, w, h, radius, ColorFromHex(hex))
		return 0
	})
	L.SetField(-2, "drawRoundedRect")

	// clear()
	L.PushFunction(func(L *lua.State) int {
		canvas.Clear()
		return 0
	})
	L.SetField(-2, "clear")

	return 1
}
