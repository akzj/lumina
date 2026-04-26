package lumina

import (
	"embed"

	"github.com/akzj/go-lua/pkg/lua"
)

//go:embed components/devtools/panel.lua
var devtoolsPanelLua embed.FS

// RegisterDevToolsPanel loads the embedded Lua DevTools panel script.
// After this, the global function _devtools_render() is available.
func RegisterDevToolsPanel(L *lua.State) {
	src, err := devtoolsPanelLua.ReadFile("components/devtools/panel.lua")
	if err != nil {
		return
	}
	if status := L.Load(string(src), "@devtools/panel.lua", "t"); status != lua.OK {
		L.Pop(1)
		return
	}
	if status := L.PCall(0, 0, 0); status != lua.OK {
		L.Pop(1)
	}
}

// CallDevToolsRender calls the Lua _devtools_render() function.
// This rebuilds the DevTools overlay using the framework's rendering pipeline.
// Must be called on the main goroutine (Lua State is not thread-safe).
func CallDevToolsRender(L *lua.State) {
	L.GetGlobal("_devtools_render")
	if !L.IsFunction(-1) {
		L.Pop(1)
		return
	}
	if status := L.PCall(0, 0, 0); status != lua.OK {
		msg, _ := L.ToString(-1)
		L.Pop(1)
		_ = msg // silently ignore errors in devtools render
	}
}
