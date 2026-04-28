package v2

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// registerAppLuaAPIs registers app-level functions for the V2 engine.
// The engine already provides createComponent/createElement/useState/defineComponent.
// This adds: quit, setInterval, setTimeout, clearInterval, clearTimeout.
func (a *App) registerAppLuaAPIs() {
	L := a.luaState

	// Get the existing "lumina" global table (already created by engine.RegisterLuaAPI).
	L.GetGlobal("lumina")
	if !L.IsTable(-1) {
		L.Pop(1)
		return
	}
	tblIdx := L.AbsIndex(-1)

	// lumina.quit()
	L.PushFunction(a.luaQuit)
	L.SetField(tblIdx, "quit")

	// lumina.setInterval(fn, ms)
	L.PushFunction(a.luaSetInterval)
	L.SetField(tblIdx, "setInterval")

	// lumina.setTimeout(fn, ms)
	L.PushFunction(a.luaSetTimeout)
	L.SetField(tblIdx, "setTimeout")

	// lumina.clearInterval(id)
	L.PushFunction(a.luaClearTimer)
	L.SetField(tblIdx, "clearInterval")

	// lumina.clearTimeout(id)
	L.PushFunction(a.luaClearTimer)
	L.SetField(tblIdx, "clearTimeout")

	// lumina.reload(moduleName) → result table or nil, error string
	L.PushFunction(a.luaReload)
	L.SetField(tblIdx, "reload")

	L.SetGlobal("lumina")
}

// luaReload implements lumina.reload(moduleName) → result table or nil, error.
// Performs a module-level hot reload that preserves state.
func (a *App) luaReload(L *lua.State) int {
	name := L.CheckString(1)
	result, err := L.ReloadModule(name)
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}
	// Push result table
	L.NewTable()
	tbl := L.AbsIndex(-1)
	L.PushInteger(int64(result.Replaced))
	L.SetField(tbl, "replaced")
	L.PushInteger(int64(result.Skipped))
	L.SetField(tbl, "skipped")
	L.PushInteger(int64(result.Added))
	L.SetField(tbl, "added")
	L.PushInteger(int64(result.Removed))
	L.SetField(tbl, "removed")

	// Mark all components dirty so they re-render with new code.
	a.engine.MarkAllComponentsDirty()

	return 1
}

// luaQuit implements lumina.quit().
func (a *App) luaQuit(_ *lua.State) int {
	a.Stop()
	return 0
}
