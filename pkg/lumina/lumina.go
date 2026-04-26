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
		"mount":           luaMount,
		"run":             luaRun,
		"quit":            luaQuit,
		"onKey":           luaOnKey,
		"createState":     createState,
		"getSize":         luaGetSize,
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
		"isFocused":           luaIsFocused,
		"isHovered":           luaIsHovered,
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
		// Window Manager API
		"createWindow":     luaCreateWindow,
		"closeWindow":      luaCloseWindow,
		"focusWindow":      luaFocusWindow,
		"moveWindow":       luaMoveWindow,
		"resizeWindow":     luaResizeWindow,
		"minimizeWindow":   luaMinimizeWindow,
		"maximizeWindow":   luaMaximizeWindow,
		"restoreWindow":    luaRestoreWindow,
		"tileWindows":      luaTileWindows,
		"getFocusedWindow": luaGetFocusedWindow,
		"getWindow":        luaGetWindow,
		"listWindows":      luaListWindows,
	}, 0)

	// Register hooks as sub-table on the lumina module
	L.NewTable()
	RegisterHooks(L)
	L.SetField(-2, "hooks")

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
		// Drag & Drop
		"useDrag":                luaUseDrag,
		"useDrop":                luaUseDrop,
		// Plugin system
		"registerPlugin":         luaRegisterPlugin,
		"usePlugin":              luaUsePlugin,
		"getPlugins":             luaGetPlugins,
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
	L.NewTable()
	L.SetFuncs(map[string]lua.Function{
		"log":   makeConsoleLog("log"),
		"warn":  makeConsoleLog("warn"),
		"error": makeConsoleLog("error"),
		"get":   consoleGet,
		"clear": consoleClear,
		"size":  consoleSize,
	}, 0)
	L.SetField(-2, "console")

	// Register debug as sub-table: lumina.debug.*
	L.NewTable()
	RegisterDebugAPI(L)
	L.SetField(-2, "debug")

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
			comp.IsRoot = true // mounted via lumina.mount() → renderComponent
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
	// NOTE: Do NOT call SetCurrentComponent(nil) here — LuaVNodeToVNode below
	// may create child components that need GetCurrentComponent() to return
	// this component so they are properly parented via AddChild.
	// SetCurrentComponent(nil) is called after VNode processing below.
	if status != lua.OK {
		SetCurrentComponent(nil)
		msg, _ := L.ToString(-1)
		L.Pop(1)
		L.PushString(fmt.Sprintf("render: %v", msg))
		L.Error()
		return 0
	}

	// vdom is now on stack at -1
	// Convert Lua table to VNode tree.
	// NOTE: LuaVNodeToVNode may call luaComponentToVNode which creates
	// child components. GetCurrentComponent() must still return this
	// component so children are properly parented via AddChild.
	newVNode := LuaVNodeToVNode(L, -1)
	SetCurrentComponent(nil) // safe to clear now — child components are created

	// Get actual terminal size from app (fallback to 80x24)
	renderW, renderH := 80, 24
	if appInst := GetApp(L); appInst != nil {
		if aw := appInst.getWidth(); aw > 0 {
			renderW = aw
		}
		if ah := appInst.getHeight(); ah > 0 {
			renderH = ah
		}
	}

	// Diff against previous render.
	var frame *Frame
	if comp != nil && comp.LastVNode != nil {
		patches := DiffVNode(comp.LastVNode, newVNode)
		_ = patches // diff available for future incremental updates
		frame = VNodeToFrame(newVNode, renderW, renderH)
	} else {
		frame = VNodeToFrame(newVNode, renderW, renderH)
	}
	// Do NOT set comp.LastVNode here — let Go-side renderComponent be
	// the single source of truth for LastVNode. This ensures
	// app.renderAllDirty() does a full re-render on first tick.

	// Bridge VNode event handlers to EventBus
	if app := GetApp(L); app != nil {
		app.bridgeVNodeEvents(newVNode)
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

	// Register shadcn component preloads (backward compat)
	RegisterShadcn(L)

	// Register lumina/ui component preloads (new naming)
	RegisterUI(L)

	// Register DevTools panel (Lua implementation)
	RegisterDevToolsPanel(L)
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
