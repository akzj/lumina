// Package lumina — lumina/ui component loader.
// Embeds the lumina/ui Lua component files and registers them as package preloads
// so that require("lumina.ui.button") etc. work from Lua.
// Also registers backward-compatible require("shadcn.xxx") aliases.
package lumina

import (
	"embed"

	"github.com/akzj/go-lua/pkg/lua"
)

//go:embed components/ui/*.lua
var uiFS embed.FS

// uiComponents maps require name → file path inside the embed FS.
var uiComponents = map[string]string{
	"lumina.ui.button":          "components/ui/button.lua",
	"lumina.ui.badge":           "components/ui/badge.lua",
	"lumina.ui.card":            "components/ui/card.lua",
	"lumina.ui.alert":           "components/ui/alert.lua",
	"lumina.ui.label":           "components/ui/label.lua",
	"lumina.ui.separator":       "components/ui/separator.lua",
	"lumina.ui.skeleton":        "components/ui/skeleton.lua",
	"lumina.ui.spinner":         "components/ui/spinner.lua",
	"lumina.ui.avatar":          "components/ui/avatar.lua",
	"lumina.ui.breadcrumb":      "components/ui/breadcrumb.lua",
	"lumina.ui.kbd":             "components/ui/kbd.lua",
	"lumina.ui.input":           "components/ui/input.lua",
	"lumina.ui.switch":          "components/ui/switch.lua",
	"lumina.ui.progress":        "components/ui/progress.lua",
	"lumina.ui.accordion":       "components/ui/accordion.lua",
	"lumina.ui.tabs":            "components/ui/tabs.lua",
	"lumina.ui.table":           "components/ui/table.lua",
	"lumina.ui.pagination":      "components/ui/pagination.lua",
	"lumina.ui.toggle":          "components/ui/toggle.lua",
	"lumina.ui.toggle_group":    "components/ui/toggle_group.lua",
	"lumina.ui.select":          "components/ui/select.lua",
	"lumina.ui.checkbox":        "components/ui/checkbox.lua",
	"lumina.ui.radio_group":     "components/ui/radio_group.lua",
	"lumina.ui.slider":          "components/ui/slider.lua",
	"lumina.ui.textarea":        "components/ui/textarea.lua",
	"lumina.ui.field":           "components/ui/field.lua",
	"lumina.ui.input_group":     "components/ui/input_group.lua",
	"lumina.ui.input_otp":       "components/ui/input_otp.lua",
	"lumina.ui.combobox":        "components/ui/combobox.lua",
	"lumina.ui.native_select":   "components/ui/native_select.lua",
	"lumina.ui.form":            "components/ui/form.lua",
	"lumina.ui.dialog":          "components/ui/dialog.lua",
	"lumina.ui.alert_dialog":    "components/ui/alert_dialog.lua",
	"lumina.ui.sheet":           "components/ui/sheet.lua",
	"lumina.ui.drawer":          "components/ui/drawer.lua",
	"lumina.ui.dropdown_menu":   "components/ui/dropdown_menu.lua",
	"lumina.ui.context_menu":    "components/ui/context_menu.lua",
	"lumina.ui.popover":         "components/ui/popover.lua",
	"lumina.ui.tooltip":         "components/ui/tooltip.lua",
	"lumina.ui.hover_card":      "components/ui/hover_card.lua",
	"lumina.ui.command":         "components/ui/command.lua",
	"lumina.ui.menubar":         "components/ui/menubar.lua",
	"lumina.ui.scroll_area":     "components/ui/scroll_area.lua",
	"lumina.ui.collapsible":     "components/ui/collapsible.lua",
	"lumina.ui.carousel":        "components/ui/carousel.lua",
	"lumina.ui.sonner":          "components/ui/sonner.lua",
	"lumina.ui.aspect_ratio":    "components/ui/aspect_ratio.lua",
	"lumina.ui.button_group":    "components/ui/button_group.lua",
	"lumina.ui.calendar":        "components/ui/calendar.lua",
	"lumina.ui.date_picker":     "components/ui/date_picker.lua",
	"lumina.ui.navigation_menu": "components/ui/navigation_menu.lua",
	"lumina.ui.resizable":       "components/ui/resizable.lua",
	"lumina.ui.sidebar":         "components/ui/sidebar.lua",
	"lumina.ui.chart":           "components/ui/chart.lua",
	"lumina.ui.data_table":      "components/ui/data_table.lua",
	"lumina.ui.color_picker":    "components/ui/color_picker.lua",
	"lumina.ui":                 "components/ui/init.lua",
}

// RegisterUI registers all lumina/ui components as Lua package preloads.
// After calling this, require("lumina.ui.button") etc. will work.
// Also registers backward-compatible require("shadcn.xxx") aliases.
func RegisterUI(L *lua.State) {
	L.GetGlobal("package")
	L.GetField(-1, "preload")

	for modName, filePath := range uiComponents {
		src, err := uiFS.ReadFile(filePath)
		if err != nil {
			continue // skip missing files
		}
		registerLuaPreloadUI(L, modName, string(src))
	}

	L.Pop(2) // pop preload + package
}

// registerLuaPreloadUI registers a Lua source string as a package preload.
func registerLuaPreloadUI(L *lua.State, name, source string) {
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
