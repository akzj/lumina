// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// ModuleName is the name used to require this module from Lua.
const ModuleName = "lumina"

// luaLoader is the module loader function that creates the lumina module table.
func luaLoader(L *lua.State) int {
	// Create module table
	L.NewTable()

	// Register module functions using SetFuncs
	L.SetFuncs(map[string]lua.Function{
		"version": func(L *lua.State) int {
			L.PushString("0.1.0")
			return 1
		},
		"echo": func(L *lua.State) int {
			// Echo back the argument
			L.PushValue(1) // copy argument to return
			return 1
		},
		"info": func(L *lua.State) int {
			// Return module info as a table
			L.NewTable()
			L.PushString("0.1.0")
			L.SetField(-2, "version")
			L.PushString("Lumina Terminal UI Framework")
			L.SetField(-2, "description")
			L.PushInteger(2024)
			L.SetField(-2, "year")
			return 1
		},
	}, 0)

	return 1 // module count (1 table returned)
}

// Open registers the lumina module into the global namespace AND into package.preload.
// This allows both global.lumina and require("lumina") to work.
func Open(L *lua.State) {
	// Set as global "lumina" by calling luaLoader directly
	luaLoader(L)
	L.SetGlobal(ModuleName)

	// Register in package.preload["lumina"] for require() support
	// Get package table
	L.GetGlobal("package")
	// Get package.preload
	L.GetField(-1, "preload")
	// Push the loader function so luaLoader result goes to preload[key]
	L.PushFunction(luaLoader)
	// Set package.preload["lumina"] = luaLoader
	L.SetField(-2, ModuleName)
	// Pop package.preload and package tables
	L.Pop(2)
}