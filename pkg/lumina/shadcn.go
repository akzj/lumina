// Package lumina — shadcn component loader.
// Embeds the shadcn Lua component files and registers them as package preloads
// so that require("shadcn.button") etc. work from Lua.
package lumina

import (
	"embed"

	"github.com/akzj/go-lua/pkg/lua"
)

//go:embed components/shadcn/*.lua
var shadcnFS embed.FS

// shadcnComponents maps require name → file path inside the embed FS.
var shadcnComponents = map[string]string{
	"shadcn.button":       "components/shadcn/button.lua",
	"shadcn.badge":        "components/shadcn/badge.lua",
	"shadcn.card":         "components/shadcn/card.lua",
	"shadcn.alert":        "components/shadcn/alert.lua",
	"shadcn.label":        "components/shadcn/label.lua",
	"shadcn.separator":    "components/shadcn/separator.lua",
	"shadcn.skeleton":     "components/shadcn/skeleton.lua",
	"shadcn.spinner":      "components/shadcn/spinner.lua",
	"shadcn.avatar":       "components/shadcn/avatar.lua",
	"shadcn.breadcrumb":   "components/shadcn/breadcrumb.lua",
	"shadcn.kbd":          "components/shadcn/kbd.lua",
	"shadcn.input":        "components/shadcn/input.lua",
	"shadcn.switch":       "components/shadcn/switch.lua",
	"shadcn.progress":     "components/shadcn/progress.lua",
	"shadcn.accordion":    "components/shadcn/accordion.lua",
	"shadcn.tabs":         "components/shadcn/tabs.lua",
	"shadcn.table":        "components/shadcn/table.lua",
	"shadcn.pagination":   "components/shadcn/pagination.lua",
	"shadcn.toggle":       "components/shadcn/toggle.lua",
	"shadcn.toggle_group": "components/shadcn/toggle_group.lua",
	"shadcn":              "components/shadcn/init.lua",
}

// RegisterShadcn registers all shadcn components as Lua package preloads.
// After calling this, require("shadcn.button") etc. will work.
func RegisterShadcn(L *lua.State) {
	L.GetGlobal("package")
	L.GetField(-1, "preload")

	for modName, filePath := range shadcnComponents {
		src, err := shadcnFS.ReadFile(filePath)
		if err != nil {
			continue // skip missing files
		}
		registerLuaPreload(L, modName, string(src))
	}

	L.Pop(2) // pop preload + package
}

// registerLuaPreload registers a Lua source string as a package preload.
func registerLuaPreload(L *lua.State, name, source string) {
	// Capture source in closure
	src := source
	modName := name
	L.PushFunction(func(L *lua.State) int {
		if status := L.Load(src, "@"+modName, "t"); status != lua.OK {
			msg, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString("error loading " + modName + ": " + msg)
			L.Error()
			return 0
		}
		// Execute the loaded chunk
		if status := L.PCall(0, 1, 0); status != 0 {
			msg, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString("error executing " + modName + ": " + msg)
			L.Error()
			return 0
		}
		return 1
	})
	L.SetField(-2, name)
}
