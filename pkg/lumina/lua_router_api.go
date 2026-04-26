package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)


// -----------------------------------------------------------------------
// Router Lua API
// -----------------------------------------------------------------------

// luaCreateRouter creates a router with route definitions.
// lumina.createRouter({ routes = { {path="/"}, {path="/users/:id"} } })
// NOTE: Only called from Lua main thread — globalRouter assignment is safe.
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


// luaNavigate navigates to a new path.
// lumina.navigate("/users/123")
func luaNavigate(L *lua.State) int {
	path := L.CheckString(1)
	globalRouter.Navigate(path)

	// Mark all components dirty to re-render with new route
	for _, comp := range globalRegistry.components {
		comp.Dirty.Store(true)
	}

	return 0
}

// luaBack navigates back in history.
// lumina.back() → bool


// luaBack navigates back in history.
// lumina.back() → bool
func luaBack(L *lua.State) int {
	ok := globalRouter.Back()
	L.PushBoolean(ok)

	if ok {
		// Mark all components dirty
		for _, comp := range globalRegistry.components {
			comp.Dirty.Store(true)
		}
	}

	return 1
}

// luaGetCurrentPath returns the current route path.
// lumina.getCurrentPath() → string


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
// Only accessed from the Lua main thread (via luaSetScrollBehavior and GetScrollBehavior).
var scrollBehavior = "instant" // default for backward compat

// GetScrollBehavior returns the current scroll behavior.


// GetScrollBehavior returns the current scroll behavior.
func GetScrollBehavior() string { return scrollBehavior }

// luaSetScrollBehavior sets the scroll behavior.
// lumina.setScrollBehavior("smooth") or lumina.setScrollBehavior("instant")


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
