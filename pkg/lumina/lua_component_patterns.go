package lumina

import (
	"sync"

	"github.com/akzj/go-lua/pkg/lua"
)


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
// Mount / Run Lua API
// -----------------------------------------------------------------------

// mountedFactory stores the factory table ref for lumina.mount().
// lumina.run() renders it and enters the event loop.
var mountedFactoryRef int

// luaMount registers a component factory as the root component.
// lumina.mount(Factory) — stores the factory for lumina.run() to render.
// In headless/test mode, it also renders immediately.


// luaMount registers a component factory as the root component.
// lumina.mount(Factory) — stores the factory for lumina.run() to render.
// In headless/test mode, it also renders immediately.
func luaMount(L *lua.State) int {
	if L.Type(1) != lua.TypeTable {
		L.PushString("mount: expected component factory table")
		L.Error()
		return 0
	}

	// Store factory ref in Lua registry
	L.PushValue(1)
	mountedFactoryRef = L.Ref(lua.RegistryIndex)

	// Render immediately (same as lumina.render(Factory, {}))
	L.PushValue(1)
	L.NewTable() // empty props
	renderComponent(L)

	return 0
}

// luaRun starts the event loop. In headless/test mode (no App attached),
// it's a no-op — the script has already rendered via mount().


// luaRun starts the event loop. In headless/test mode (no App attached),
// it's a no-op — the script has already rendered via mount().
func luaRun(L *lua.State) int {
	// Check if there's an App attached to this State
	app := GetApp(L)
	if app == nil {
		// No app — headless/test mode. Just return.
		return 0
	}

	// If we have an app, mark it as running.
	// The actual event loop is managed by App.Run/RunInteractive/RunWithTermIO,
	// not by the Lua-side lumina.run() call.
	app.running = true
	return 0
}

// luaQuit stops the application.
// lumina.quit()


// luaQuit stops the application.
// lumina.quit()
func luaQuit(L *lua.State) int {
	app := GetApp(L)
	if app != nil {
		app.Stop()
	}
	return 0
}

// luaGetSize returns the terminal width and height.
// Usage: local w, h = lumina.getSize()


// luaGetSize returns the terminal width and height.
// Usage: local w, h = lumina.getSize()
func luaGetSize(L *lua.State) int {
	app := GetApp(L)
	w, h := 80, 24 // defaults
	if app != nil {
		if aw := app.getWidth(); aw > 0 {
			w = aw
		}
		if ah := app.getHeight(); ah > 0 {
			h = ah
		}
	}
	L.PushInteger(int64(w))
	L.PushInteger(int64(h))
	return 2
}

// keyBindings stores key → Lua function ref mappings for lumina.onKey().
var (
	keyBindings   = make(map[string]int) // key → Lua registry ref
	keyBindingsMu sync.Mutex
)

// luaOnKey registers a key binding: lumina.onKey("q", function() ... end)
