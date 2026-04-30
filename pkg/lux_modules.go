package v2

// This file registers Lua modules so that require("lumina"), require("lux"),
// require("lux.card"), etc. work from any Lua script.
//
// Layering (see README + docs/DESIGN-widgets.md):
//   - Radix-style primitives: Go widgets in pkg/widget/, exposed as lumina.Button,
//     lumina.Checkbox, lumina.Select, … (registered in NewApp, not here).
//   - Lux: Lua-only presentation templates under lua/lux/*.lua, embedded at build
//     time (see lua/lux/embed.go) so require() works without shipping files beside
//     the binary.

import (
	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/lua/lux"
	"github.com/akzj/lumina/lua/theme"
)

// registerLuxModules registers all lux modules into package.preload
// so that require("lux"), require("lux.card"), etc. work.
func registerLuxModules(L *lua.State) {
	// Make require("lumina") work by returning the global lumina table
	preloadLuaSource(L, "lumina", `return lumina`)

	preloadLuaSource(L, "theme", mustReadThemeLua("init.lua"))

	preloadLuaSource(L, "lux.slot", mustReadLuxLua("slot.lua"))
	preloadLuaSource(L, "lux.card", mustReadLuxLua("card.lua"))
	preloadLuaSource(L, "lux.badge", mustReadLuxLua("badge.lua"))
	preloadLuaSource(L, "lux.divider", mustReadLuxLua("divider.lua"))
	preloadLuaSource(L, "lux.progress", mustReadLuxLua("progress.lua"))
	preloadLuaSource(L, "lux.spinner", mustReadLuxLua("spinner.lua"))
	preloadLuaSource(L, "lux.dialog", mustReadLuxLua("dialog.lua"))
	preloadLuaSource(L, "lux.layout", mustReadLuxLua("layout.lua"))
	preloadLuaSource(L, "lux.command_palette", mustReadLuxLua("command_palette.lua"))
	preloadLuaSource(L, "lux.list", mustReadLuxLua("list.lua"))
	preloadLuaSource(L, "lux.data_grid", mustReadLuxLua("data_grid.lua"))
	preloadLuaSource(L, "lux.wm", mustReadLuxLua("wm.lua"))
	preloadLuaSource(L, "lux.pagination", mustReadLuxLua("pagination.lua"))
	preloadLuaSource(L, "lux.tabs", mustReadLuxLua("tabs.lua"))
	preloadLuaSource(L, "lux.alert", mustReadLuxLua("alert.lua"))
	preloadLuaSource(L, "lux.accordion", mustReadLuxLua("accordion.lua"))
	preloadLuaSource(L, "lux.breadcrumb", mustReadLuxLua("breadcrumb.lua"))
	preloadLuaSource(L, "lux.text_input", mustReadLuxLua("text_input.lua"))
	preloadLuaSource(L, "lux.button", mustReadLuxLua("button.lua"))
	preloadLuaSource(L, "lux.checkbox", mustReadLuxLua("checkbox.lua"))
	preloadLuaSource(L, "lux.radio", mustReadLuxLua("radio.lua"))
	preloadLuaSource(L, "lux.switch", mustReadLuxLua("switch.lua"))
	preloadLuaSource(L, "lux.dropdown", mustReadLuxLua("dropdown.lua"))
	preloadLuaSource(L, "lux.toast", mustReadLuxLua("toast.lua"))
	preloadLuaSource(L, "lux.tree", mustReadLuxLua("tree.lua"))
	preloadLuaSource(L, "lux.form", mustReadLuxLua("form.lua"))
	preloadLuaSource(L, "lux.atlantis", mustReadLuxLua("atlantis.lua"))

	// Register the lux umbrella module (requires individual modules)
	preloadLuaSource(L, "lux", mustReadLuxLua("init.lua"))
}

func mustReadLuxLua(filename string) string {
	b, err := lux.Sources.ReadFile(filename)
	if err != nil {
		panic("lux Lua embed " + filename + ": " + err.Error())
	}
	return string(b)
}

func mustReadThemeLua(filename string) string {
	b, err := theme.Sources.ReadFile(filename)
	if err != nil {
		panic("theme Lua embed " + filename + ": " + err.Error())
	}
	return string(b)
}

// preloadLuaSource registers a Lua source string as a preloaded module.
// When require(name) is called, the source is compiled and executed,
// and its return value becomes the module value.
func preloadLuaSource(L *lua.State, name, source string) {
	L.GetGlobal("package")
	if L.IsNil(-1) {
		L.Pop(1)
		return
	}
	L.GetField(-1, "preload")
	if L.IsNil(-1) {
		L.Pop(2)
		return
	}

	// Capture source in closure
	src := source
	modName := name

	L.PushFunction(func(L *lua.State) int {
		// Load the chunk (pushes compiled function on success)
		status := L.Load(src, "="+modName, "t")
		if status != lua.OK {
			msg, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString("error loading module '" + modName + "': " + msg)
			L.Error()
			return 0
		}
		// Execute the chunk: 0 args, 1 result
		L.Call(0, 1)
		return 1
	})
	L.SetField(-2, name) // package.preload[name] = opener
	L.Pop(2)              // pop preload + package
}
