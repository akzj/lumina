package v2

// This file registers Lua modules so that require("lumina"), require("lux"),
// require("lux.card"), etc. work from any Lua script.

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// Embedded Lua source for lux components.
// These are the canonical sources from lua/lux/*.lua, stored as Go constants
// so they can be preloaded into the Lua VM without filesystem access.

const luxInitLua = `
local M = {}
M.Card = require("lux.card")
M.Badge = require("lux.badge")
M.Divider = require("lux.divider")
M.Progress = require("lux.progress")
M.Spinner = require("lux.spinner")
return M
`

const luxCardLua = `
local Card = lumina.defineComponent("Card", function(props)
    return lumina.createElement("box", {
        style = {
            border = props.border or "rounded",
            padding = props.padding or 1,
            background = props.bg or "",
        },
    },
        props.title and lumina.createElement("text", {
            bold = true,
        }, props.title) or nil,
        table.unpack(props.children or {})
    )
end)
return Card
`

const luxBadgeLua = `
local Badge = lumina.defineComponent("Badge", function(props)
    local variant = props.variant or "default"
    local t = lumina.getTheme and lumina.getTheme() or {}
    local fg, bg
    if variant == "success" then
        fg = t.success or "#A6E3A1"; bg = t.surface0 or "#313244"
    elseif variant == "warning" then
        fg = t.warning or "#F9E2AF"; bg = t.surface0 or "#313244"
    elseif variant == "error" then
        fg = t.error or "#F38BA8"; bg = t.surface0 or "#313244"
    else
        fg = t.primary or "#89B4FA"; bg = t.surface0 or "#313244"
    end
    return lumina.createElement("text", {
        foreground = fg,
        background = bg,
        bold = true,
    }, " " .. (props.label or "") .. " ")
end)
return Badge
`

const luxDividerLua = `
local Divider = lumina.defineComponent("Divider", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local char = props.char or "─"
    local width = props.width or 40
    return lumina.createElement("text", {
        foreground = props.fg or t.surface1 or "#45475A",
        dim = true,
    }, string.rep(char, width))
end)
return Divider
`

const luxProgressLua = `
local Progress = lumina.defineComponent("Progress", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local value = props.value or 0
    local width = props.width or 20
    local filled = math.floor(width * value / 100)
    local empty = width - filled
    local bar = string.rep("█", filled) .. string.rep("░", empty)
    local label = string.format(" %d%%", value)
    return lumina.createElement("hbox", {},
        lumina.createElement("text", {
            foreground = t.primary or "#89B4FA",
        }, bar),
        lumina.createElement("text", {
            foreground = t.text or "#CDD6F4",
        }, label)
    )
end)
return Progress
`

const luxSpinnerLua = `
local Spinner = lumina.defineComponent("Spinner", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local frames = {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
    local frame, setFrame = lumina.useState("frame", 1)
    local label = props.label or "Loading..."
    lumina.useEffect("spin", function()
        local timer = lumina.setInterval(function()
            setFrame(function(f) return (f % #frames) + 1 end)
        end, 80)
        return function() lumina.clearInterval(timer) end
    end, {})
    return lumina.createElement("hbox", {},
        lumina.createElement("text", {
            foreground = t.primary or "#89B4FA",
        }, frames[frame] .. " "),
        lumina.createElement("text", {
            foreground = t.text or "#CDD6F4",
        }, label)
    )
end)
return Spinner
`

const luxThemeLua = `
local M = {}
local _fallback = {
    base     = "#1E1E2E",
    surface0 = "#313244",
    surface1 = "#45475A",
    surface2 = "#585B70",
    text     = "#CDD6F4",
    muted    = "#6C7086",
    primary  = "#89B4FA",
    hover    = "#B4BEFE",
    pressed  = "#74C7EC",
    success  = "#A6E3A1",
    warning  = "#F9E2AF",
    error    = "#F38BA8",
}
function M.current()
    if lumina and lumina.getTheme then
        return lumina.getTheme()
    end
    return _fallback
end
return M
`

// registerLuxModules registers all lux modules into package.preload
// so that require("lux"), require("lux.card"), etc. work.
func registerLuxModules(L *lua.State) {
	// Make require("lumina") work by returning the global lumina table
	preloadLuaSource(L, "lumina", `return lumina`)

	// Register theme module
	preloadLuaSource(L, "theme", luxThemeLua)

	// Register individual lux components
	preloadLuaSource(L, "lux.card", luxCardLua)
	preloadLuaSource(L, "lux.badge", luxBadgeLua)
	preloadLuaSource(L, "lux.divider", luxDividerLua)
	preloadLuaSource(L, "lux.progress", luxProgressLua)
	preloadLuaSource(L, "lux.spinner", luxSpinnerLua)

	// Register the lux umbrella module (requires individual modules)
	preloadLuaSource(L, "lux", luxInitLua)
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
	L.Pop(2)             // pop preload + package
}
