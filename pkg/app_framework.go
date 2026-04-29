package v2

import (
	"fmt"
	"os"

	"github.com/akzj/go-lua/pkg/lua"
)

// --- Types for framework bindings ---

// storeBinding tracks a component's subscription to a store key.
type storeBinding struct {
	compID   string
	stateKey string
}

// routeBinding tracks a component subscribed to route changes.
type routeBinding struct {
	compID   string
	stateKey string
}

// globalKeyBinding holds a registered global keybinding.
type globalKeyBinding struct {
	key string
	ref int // Lua registry ref to handler function
}

// registerFrameworkAPIs registers lumina.store, lumina.router, lumina.useStore,
// lumina.useRoute, and lumina.app on the Lua global table.
// Called from NewApp after registerAppLuaAPIs.
func (a *App) registerFrameworkAPIs() {
	L := a.luaState
	L.GetGlobal("lumina")
	if !L.IsTable(-1) {
		L.Pop(1)
		return
	}
	tblIdx := L.AbsIndex(-1)

	a.registerStoreLuaAPI(L, tblIdx)
	a.registerRouterLuaAPI(L, tblIdx)

	// lumina.useStore(key) → value
	L.PushFunction(a.luaUseStore)
	L.SetField(tblIdx, "useStore")

	// lumina.useRoute() → {path, params}
	L.PushFunction(a.luaUseRoute)
	L.SetField(tblIdx, "useRoute")

	// lumina.app(config)
	a.registerAppLuaEntry(L, tblIdx)

	L.SetGlobal("lumina")
}

// --- Store API ---

func (a *App) registerStoreLuaAPI(L *lua.State, tblIdx int) {
	L.NewTable()
	storeIdx := L.AbsIndex(-1)

	// lumina.store.get(key) → value
	L.PushFunction(func(L *lua.State) int {
		key := L.CheckString(1)
		val, ok := a.store.Get(key)
		if !ok {
			L.PushNil()
			return 1
		}
		L.PushAny(val)
		return 1
	})
	L.SetField(storeIdx, "get")

	// lumina.store.set(key, value)
	L.PushFunction(func(L *lua.State) int {
		key := L.CheckString(1)
		value := L.ToAny(2)
		a.store.Set(key, value)
		a.notifyStoreSubscribers(key)
		return 0
	})
	L.SetField(storeIdx, "set")

	// lumina.store.delete(key)
	L.PushFunction(func(L *lua.State) int {
		key := L.CheckString(1)
		a.store.Delete(key)
		a.notifyStoreSubscribers(key)
		return 0
	})
	L.SetField(storeIdx, "delete")

	// lumina.store.getAll() → table
	L.PushFunction(func(L *lua.State) int {
		all := a.store.GetAll()
		L.PushAny(all)
		return 1
	})
	L.SetField(storeIdx, "getAll")

	// lumina.store.batch(updates_table)
	L.PushFunction(func(L *lua.State) int {
		L.CheckType(1, lua.TypeTable)
		updates := make(map[string]any)
		L.PushNil()
		for L.Next(1) {
			key, _ := L.ToString(-2)
			val := L.ToAny(-1)
			updates[key] = val
			L.Pop(1)
		}
		a.store.Batch(updates)
		for key := range updates {
			a.notifyStoreSubscribers(key)
		}
		return 0
	})
	L.SetField(storeIdx, "batch")

	L.SetField(tblIdx, "store")
}

// --- Store bindings ---

func (a *App) registerStoreBinding(storeKey, compID, stateKey string) {
	// Check if already registered
	for _, b := range a.storeBindings[storeKey] {
		if b.compID == compID && b.stateKey == stateKey {
			return
		}
	}
	if a.storeBindings == nil {
		a.storeBindings = make(map[string][]storeBinding)
	}
	a.storeBindings[storeKey] = append(a.storeBindings[storeKey], storeBinding{
		compID:   compID,
		stateKey: stateKey,
	})
}

func (a *App) notifyStoreSubscribers(key string) {
	if a.storeBindings == nil {
		return
	}
	val, _ := a.store.Get(key)
	for _, b := range a.storeBindings[key] {
		a.engine.SetState(b.compID, b.stateKey, val)
	}
}

// --- useStore hook ---

// luaUseStore implements lumina.useStore(key) → value
// Reads the current value from the global store and subscribes the
// current component to changes on that key.
func (a *App) luaUseStore(L *lua.State) int {
	key := L.CheckString(1)

	comp := a.engine.CurrentComponent()
	if comp == nil {
		L.PushString("useStore: must be called inside a component render function")
		L.Error()
		return 0
	}

	stateKey := "__store__" + key

	// Register this component as a subscriber (idempotent)
	a.registerStoreBinding(key, comp.ID, stateKey)

	// Get current value from store
	val, _ := a.store.Get(key)

	// Store in component state (without marking dirty — we're inside render)
	comp.State[stateKey] = val

	// Return current value
	L.PushAny(val)
	return 1
}

// --- Router API ---

func (a *App) registerRouterLuaAPI(L *lua.State, tblIdx int) {
	L.NewTable()
	routerIdx := L.AbsIndex(-1)

	// lumina.router.navigate(path)
	L.PushFunction(func(L *lua.State) int {
		path := L.CheckString(1)
		a.routerMgr.Navigate(path)
		a.notifyRouteSubscribers()
		return 0
	})
	L.SetField(routerIdx, "navigate")

	// lumina.router.back() → bool
	L.PushFunction(func(L *lua.State) int {
		ok := a.routerMgr.Back()
		if ok {
			a.notifyRouteSubscribers()
		}
		L.PushBoolean(ok)
		return 1
	})
	L.SetField(routerIdx, "back")

	// lumina.router.path() → string
	L.PushFunction(func(L *lua.State) int {
		L.PushString(a.routerMgr.CurrentPath())
		return 1
	})
	L.SetField(routerIdx, "path")

	// lumina.router.params() → table
	L.PushFunction(func(L *lua.State) int {
		params := a.routerMgr.Params()
		L.NewTable()
		tbl := L.AbsIndex(-1)
		for k, v := range params {
			L.PushString(v)
			L.SetField(tbl, k)
		}
		return 1
	})
	L.SetField(routerIdx, "params")

	// lumina.router.addRoute(pattern)
	L.PushFunction(func(L *lua.State) int {
		pattern := L.CheckString(1)
		a.routerMgr.AddRoute(pattern)
		return 0
	})
	L.SetField(routerIdx, "addRoute")

	L.SetField(tblIdx, "router")
}

// --- Route bindings ---

func (a *App) registerRouteBinding(compID, stateKey string) {
	for _, b := range a.routeBindings {
		if b.compID == compID {
			return
		}
	}
	a.routeBindings = append(a.routeBindings, routeBinding{compID, stateKey})
}

func (a *App) notifyRouteSubscribers() {
	path := a.routerMgr.CurrentPath()
	for _, b := range a.routeBindings {
		a.engine.SetState(b.compID, b.stateKey, path)
	}
}

// --- useRoute hook ---

// luaUseRoute implements lumina.useRoute() → {path, params}
// Returns the current route info and subscribes the component to route changes.
func (a *App) luaUseRoute(L *lua.State) int {
	comp := a.engine.CurrentComponent()
	if comp == nil {
		L.PushString("useRoute: must be called inside a component render function")
		L.Error()
		return 0
	}

	stateKey := "__route__"

	// Register this component for route change notifications (idempotent)
	a.registerRouteBinding(comp.ID, stateKey)

	// Store in component state (without marking dirty — we're inside render)
	comp.State[stateKey] = a.routerMgr.CurrentPath()

	// Return {path = "...", params = {...}}
	L.NewTable()
	tbl := L.AbsIndex(-1)
	L.PushString(a.routerMgr.CurrentPath())
	L.SetField(tbl, "path")

	params := a.routerMgr.Params()
	L.NewTable()
	paramsTbl := L.AbsIndex(-1)
	for k, v := range params {
		L.PushString(v)
		L.SetField(paramsTbl, k)
	}
	L.SetField(tbl, "params")

	return 1
}

// --- lumina.app entry point ---

func (a *App) registerAppLuaEntry(L *lua.State, tblIdx int) {
	L.PushFunction(func(L *lua.State) int {
		L.CheckType(1, lua.TypeTable)
		cfgIdx := L.AbsIndex(1)

		// Read id (required)
		id := L.GetFieldString(cfgIdx, "id")
		if id == "" {
			L.PushString("lumina.app: 'id' is required")
			L.Error()
			return 0
		}

		// Read name (optional, defaults to id)
		name := L.GetFieldString(cfgIdx, "name")
		if name == "" {
			name = id
		}

		// Initialize store from config
		L.GetField(cfgIdx, "store")
		if L.IsTable(-1) {
			storeIdx2 := L.AbsIndex(-1)
			L.PushNil()
			for L.Next(storeIdx2) {
				key, _ := L.ToString(-2)
				val := L.ToAny(-1)
				a.store.Set(key, val)
				L.Pop(1)
			}
		}
		L.Pop(1)

		// Register routes from config
		L.GetField(cfgIdx, "routes")
		if L.IsTable(-1) {
			routesIdx := L.AbsIndex(-1)
			L.PushNil()
			for L.Next(routesIdx) {
				pattern, _ := L.ToString(-2)
				if pattern != "" {
					a.routerMgr.AddRoute(pattern)
				}
				L.Pop(1)
			}
		}
		L.Pop(1)

		// Store global keybindings
		L.GetField(cfgIdx, "keys")
		if L.IsTable(-1) {
			keysIdx := L.AbsIndex(-1)
			L.PushNil()
			for L.Next(keysIdx) {
				keyName, _ := L.ToString(-2)
				if keyName != "" && L.IsFunction(-1) {
					L.PushValue(-1) // push function again for Ref
					ref := L.Ref(lua.RegistryIndex)
					a.registerGlobalKey(keyName, ref)
				}
				L.Pop(1)
			}
		}
		L.Pop(1)

		// Get render function (required)
		L.GetField(cfgIdx, "render")
		if !L.IsFunction(-1) {
			L.Pop(1)
			L.PushString("lumina.app: 'render' function is required")
			L.Error()
			return 0
		}
		renderRef := L.Ref(lua.RegistryIndex)

		// Create the root component
		a.engine.CreateRootComponent(id, name, int64(renderRef))

		return 0
	})
	L.SetField(tblIdx, "app")
}

// --- Global keybindings ---

func (a *App) registerGlobalKey(key string, ref int) {
	a.globalKeys = append(a.globalKeys, globalKeyBinding{key: key, ref: ref})
}

// handleGlobalKeys checks if a key event matches a global keybinding.
// Returns true if handled.
func (a *App) handleGlobalKeys(key string) bool {
	for _, binding := range a.globalKeys {
		if binding.key == key {
			L := a.luaState
			L.RawGetI(lua.RegistryIndex, int64(binding.ref))
			if L.IsFunction(-1) {
				if status := L.PCall(0, 0, 0); status != lua.OK {
					errMsg, _ := L.ToString(-1)
					L.Pop(1)
					a.setLastError(fmt.Sprintf("key handler [%s]: %s", key, errMsg))
				}
			} else {
				L.Pop(1)
			}
			return true
		}
	}
	return false
}

// setLastError stores an error message and logs it to stderr.
func (a *App) setLastError(msg string) {
	a.lastError = msg
	fmt.Fprintf(os.Stderr, "lumina: %s\n", msg)
}
