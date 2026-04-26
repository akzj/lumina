// Package lumina — shadcn: alias for lumina.ui component loader.
// Embeds the shadcn Lua component files and registers them as package preloads
// so that require("shadcn.button") etc. work from Lua.
package lumina

import (
	"embed"

	"github.com/akzj/go-lua/pkg/lua"
)

//go:embed components/ui/*.lua
var shadcnFS embed.FS

// shadcnComponents maps require name → file path inside the embed FS.
var shadcnComponents = map[string]string{
	"shadcn.button":       "components/ui/button.lua",
	"shadcn.badge":        "components/ui/badge.lua",
	"shadcn.card":         "components/ui/card.lua",
	"shadcn.alert":        "components/ui/alert.lua",
	"shadcn.label":        "components/ui/label.lua",
	"shadcn.separator":    "components/ui/separator.lua",
	"shadcn.skeleton":     "components/ui/skeleton.lua",
	"shadcn.spinner":      "components/ui/spinner.lua",
	"shadcn.avatar":       "components/ui/avatar.lua",
	"shadcn.breadcrumb":   "components/ui/breadcrumb.lua",
	"shadcn.kbd":          "components/ui/kbd.lua",
	"shadcn.input":        "components/ui/input.lua",
	"shadcn.switch":       "components/ui/switch.lua",
	"shadcn.progress":     "components/ui/progress.lua",
	"shadcn.accordion":    "components/ui/accordion.lua",
	"shadcn.tabs":         "components/ui/tabs.lua",
	"shadcn.table":        "components/ui/table.lua",
	"shadcn.pagination":   "components/ui/pagination.lua",
	"shadcn.toggle":       "components/ui/toggle.lua",
	"shadcn.toggle_group": "components/ui/toggle_group.lua",
	// Phase 22: Form components
	"shadcn.select":        "components/ui/select.lua",
	"shadcn.checkbox":      "components/ui/checkbox.lua",
	"shadcn.radio_group":   "components/ui/radio_group.lua",
	"shadcn.slider":        "components/ui/slider.lua",
	"shadcn.textarea":      "components/ui/textarea.lua",
	"shadcn.field":         "components/ui/field.lua",
	"shadcn.input_group":   "components/ui/input_group.lua",
	"shadcn.input_otp":     "components/ui/input_otp.lua",
	"shadcn.combobox":      "components/ui/combobox.lua",
	"shadcn.native_select": "components/ui/native_select.lua",
	"shadcn.form":          "components/ui/form.lua",
	// Phase 23: Overlay components
	"shadcn.dialog":         "components/ui/dialog.lua",
	"shadcn.alert_dialog":   "components/ui/alert_dialog.lua",
	"shadcn.sheet":          "components/ui/sheet.lua",
	"shadcn.drawer":         "components/ui/drawer.lua",
	"shadcn.dropdown_menu":  "components/ui/dropdown_menu.lua",
	"shadcn.context_menu":   "components/ui/context_menu.lua",
	"shadcn.popover":        "components/ui/popover.lua",
	"shadcn.tooltip":        "components/ui/tooltip.lua",
	"shadcn.hover_card":     "components/ui/hover_card.lua",
	// Phase 23: Complex components
	"shadcn.command":        "components/ui/command.lua",
	"shadcn.menubar":        "components/ui/menubar.lua",
	"shadcn.scroll_area":    "components/ui/scroll_area.lua",
	"shadcn.collapsible":    "components/ui/collapsible.lua",
	"shadcn.carousel":       "components/ui/carousel.lua",
	"shadcn.sonner":         "components/ui/sonner.lua",
	// Phase 38: Additional components
	"shadcn.aspect_ratio":     "components/ui/aspect_ratio.lua",
	"shadcn.button_group":     "components/ui/button_group.lua",
	"shadcn.calendar":         "components/ui/calendar.lua",
	"shadcn.date_picker":      "components/ui/date_picker.lua",
	"shadcn.navigation_menu":  "components/ui/navigation_menu.lua",
	"shadcn.resizable":        "components/ui/resizable.lua",
	"shadcn.sidebar":          "components/ui/sidebar.lua",
	"shadcn.chart":            "components/ui/chart.lua",
	"shadcn.data_table":       "components/ui/data_table.lua",
	"shadcn.color_picker":     "components/ui/color_picker.lua",
	"shadcn":              "components/ui/init.lua",
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
