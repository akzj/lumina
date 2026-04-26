package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)


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
