package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// registerDevToolsModule registers the lumina.devtools subtable.
func registerDevToolsModule(L *lua.State) {
	L.NewTable()

	L.PushFunction(luaDevToolsEnable)
	L.SetField(-2, "enable")

	L.PushFunction(luaDevToolsDisable)
	L.SetField(-2, "disable")

	L.PushFunction(luaDevToolsToggle)
	L.SetField(-2, "toggle")

	L.PushFunction(luaDevToolsIsEnabled)
	L.SetField(-2, "isEnabled")

	L.PushFunction(luaDevToolsIsVisible)
	L.SetField(-2, "isVisible")

	L.PushFunction(luaDevToolsGetTree)
	L.SetField(-2, "getTree")

	L.PushFunction(luaDevToolsGetInspector)
	L.SetField(-2, "getInspector")

	L.PushFunction(luaDevToolsSelect)
	L.SetField(-2, "select")

	L.PushFunction(luaDevToolsSummary)
	L.SetField(-2, "summary")

	L.SetField(-2, "devtools")
}

func luaDevToolsEnable(L *lua.State) int {
	globalDevTools.Enable()
	return 0
}

func luaDevToolsDisable(L *lua.State) int {
	globalDevTools.Disable()
	return 0
}

func luaDevToolsToggle(L *lua.State) int {
	globalDevTools.Toggle()
	return 0
}

func luaDevToolsIsEnabled(L *lua.State) int {
	L.PushBoolean(globalDevTools.IsEnabled())
	return 1
}

func luaDevToolsIsVisible(L *lua.State) int {
	L.PushBoolean(globalDevTools.IsVisible())
	return 1
}

func luaDevToolsGetTree(L *lua.State) int {
	tree := globalDevTools.RenderTree()
	L.PushString(tree)
	return 1
}

func luaDevToolsGetInspector(L *lua.State) int {
	inspector := globalDevTools.RenderInspector()
	L.PushString(inspector)
	return 1
}

func luaDevToolsSelect(L *lua.State) int {
	id := L.CheckString(1)
	globalDevTools.SetSelected(id)
	return 0
}

func luaDevToolsSummary(L *lua.State) int {
	summary := globalDevTools.Summary()
	L.PushAny(summary)
	return 1
}
