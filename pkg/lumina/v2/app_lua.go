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

	L.SetGlobal("lumina")
}

// luaQuit implements lumina.quit().
func (a *App) luaQuit(_ *lua.State) int {
	a.Stop()
	return 0
}
