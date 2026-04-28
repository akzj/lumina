package v2

// This file registers Lua modules so that require("lumina"), require("lux"),
// require("lux.card"), etc. work from any Lua script.
//
// Layering (see README + docs/DESIGN-widgets.md):
//   - Radix-style primitives: Go widgets in pkg/widget/, exposed as lumina.Button,
//     lumina.Checkbox, lumina.Select, … (registered in NewApp, not here).
//   - Lux: Lua-only presentation templates below, embedded from lua/lux/*.lua for
//     require() without shipping those files beside the binary.

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
M.Dialog = require("lux.dialog")
M.Layout = require("lux.layout")
M.CommandPalette = require("lux.command_palette")
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

const luxSlotLua = `
local function Slot(name)
    return setmetatable({}, {
        __call = function(self, arg)
            local children = {}
            local props = {}

            if type(arg) == "table" then
                for k, v in pairs(arg) do
                    if type(k) == "number" then
                        children[k] = v
                    else
                        props[k] = v
                    end
                end
            elseif type(arg) == "string" then
                children[1] = arg
            end

            return {
                type = "_slot",
                _slotName = name,
                children = children,
                props = props,
            }
        end,
    })
end
return Slot
`

const luxDialogLua = `
local Slot = require("lux.slot")

local DialogTitle = Slot("title")
local DialogContent = Slot("content")
local DialogActions = Slot("actions")

local Dialog = lumina.defineComponent("LuxDialog", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local open = props.open
    if not open then
        return lumina.createElement("box", { style = { width = 0, height = 0 } })
    end

    local title = props.title or "Dialog"
    local width = props.width or 40
    local children = props.children or {}

    local titleSlot = nil
    local contentSlot = nil
    local actionsSlot = nil
    local otherChildren = {}

    for _, child in ipairs(children) do
        if type(child) == "table" and child._slotName then
            if child._slotName == "title" then
                titleSlot = child
            elseif child._slotName == "content" then
                contentSlot = child
            elseif child._slotName == "actions" then
                actionsSlot = child
            end
        else
            otherChildren[#otherChildren + 1] = child
        end
    end

    local dialogChildren = {}

    if titleSlot and titleSlot.children and #titleSlot.children > 0 then
        local titleText = titleSlot.children[1]
        if type(titleText) == "string" then
            dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
                foreground = t.primary or "#89B4FA",
                bold = true,
            }, titleText)
        else
            dialogChildren[#dialogChildren + 1] = titleText
        end
    else
        dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
            foreground = t.primary or "#89B4FA",
            bold = true,
        }, title)
    end

    local divWidth = width - 4
    if divWidth < 1 then divWidth = 1 end
    dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
        foreground = t.muted or "#6C7086",
        dim = true,
    }, string.rep("-", divWidth))

    if contentSlot and contentSlot.children and #contentSlot.children > 0 then
        for _, child in ipairs(contentSlot.children) do
            if type(child) == "string" then
                dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
                    foreground = t.text or "#CDD6F4",
                }, child)
            else
                dialogChildren[#dialogChildren + 1] = child
            end
        end
    elseif props.message then
        dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
            foreground = t.text or "#CDD6F4",
        }, props.message)
    end

    for _, child in ipairs(otherChildren) do
        dialogChildren[#dialogChildren + 1] = child
    end

    if actionsSlot and actionsSlot.children and #actionsSlot.children > 0 then
        dialogChildren[#dialogChildren + 1] = lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            dim = true,
        }, string.rep("-", divWidth))
        dialogChildren[#dialogChildren + 1] = lumina.createElement("hbox", {
            style = { gap = 1 },
        }, table.unpack(actionsSlot.children))
    end

    return lumina.createElement("vbox", {
        style = {
            border = "rounded",
            padding = 1,
            width = width,
            background = t.surface0 or "#313244",
        },
    }, table.unpack(dialogChildren))
end)

Dialog.Title = DialogTitle
Dialog.Content = DialogContent
Dialog.Actions = DialogActions

return Dialog
`

const luxLayoutLua = `
local Layout = lumina.defineComponent("Layout", function(props)
    local children = props.children or {}

    -- Separate children by slot type (marker stored as _layoutSlot prop)
    local header, footer, sidebar, main
    local others = {}

    for _, child in ipairs(children) do
        if child and type(child) == "table" then
            local slot = child._layoutSlot
            if slot == "header" then
                header = child
            elseif slot == "footer" then
                footer = child
            elseif slot == "sidebar" then
                sidebar = child
            elseif slot == "main" then
                main = child
            else
                others[#others + 1] = child
            end
        else
            others[#others + 1] = child
        end
    end

    -- Extract children from a slot wrapper (vbox with marker props)
    -- The slot wrappers are vbox elements; we use their children directly
    local function slotChildren(slot)
        if not slot then return {} end
        return slot.children or {}
    end

    local function slotProp(slot, key, default)
        if not slot then return default end
        return slot[key] or default
    end

    -- Build vertical stack
    local vboxChildren = {}

    if header then
        vboxChildren[#vboxChildren + 1] = lumina.createElement("hbox", {
            style = {
                height = slotProp(header, "_height", 1),
                background = slotProp(header, "_bg", ""),
            },
        }, table.unpack(slotChildren(header)))
    end

    -- Middle section
    local mainChildren
    if main then
        mainChildren = slotChildren(main)
    else
        mainChildren = others
    end

    if sidebar then
        -- Horizontal: sidebar | main
        vboxChildren[#vboxChildren + 1] = lumina.createElement("hbox", {
            style = { flex = 1 },
        },
            lumina.createElement("vbox", {
                style = {
                    width = slotProp(sidebar, "_width", 20),
                    border = slotProp(sidebar, "_border", "none"),
                    background = slotProp(sidebar, "_bg", ""),
                },
            }, table.unpack(slotChildren(sidebar))),
            lumina.createElement("vbox", {
                style = { flex = 1 },
            }, table.unpack(mainChildren))
        )
    else
        vboxChildren[#vboxChildren + 1] = lumina.createElement("vbox", {
            style = { flex = 1 },
        }, table.unpack(mainChildren))
    end

    if footer then
        vboxChildren[#vboxChildren + 1] = lumina.createElement("hbox", {
            style = {
                height = slotProp(footer, "_height", 1),
                background = slotProp(footer, "_bg", ""),
            },
        }, table.unpack(slotChildren(footer)))
    end

    return lumina.createElement("vbox", {
        style = {
            width = props.width,
            height = props.height,
        },
    }, table.unpack(vboxChildren))
end)

-- Slot constructors
-- Return createElement-based descriptors with marker props so they survive
-- the Go component props pipeline (no raw Lua arrays that panic in propsEqual).

function Layout.Header(props)
    local children = {}
    for i, v in ipairs(props) do
        children[i] = v
    end
    return lumina.createElement("vbox", {
        _layoutSlot = "header",
        _height = props.height or 1,
        _bg = props.background or props.bg,
    }, table.unpack(children))
end

function Layout.Footer(props)
    local children = {}
    for i, v in ipairs(props) do
        children[i] = v
    end
    return lumina.createElement("vbox", {
        _layoutSlot = "footer",
        _height = props.height or 1,
        _bg = props.background or props.bg,
    }, table.unpack(children))
end

function Layout.Sidebar(props)
    local children = {}
    for i, v in ipairs(props) do
        children[i] = v
    end
    return lumina.createElement("vbox", {
        _layoutSlot = "sidebar",
        _width = props.width or 20,
        _border = props.border,
        _bg = props.background or props.bg,
    }, table.unpack(children))
end

function Layout.Main(props)
    local children = {}
    for i, v in ipairs(props) do
        children[i] = v
    end
    return lumina.createElement("vbox", {
        _layoutSlot = "main",
    }, table.unpack(children))
end

return Layout

`

const luxCommandPaletteLua = `
local CommandPalette = lumina.defineComponent("CommandPalette", function(props)
    local query, setQuery = lumina.useState("query", "")
    local selectedIdx, setSelectedIdx = lumina.useState("selectedIdx", 1)

    local commands = props.commands or {}
    local t = lumina.getTheme and lumina.getTheme() or {}

    -- Filter commands by query (substring match, case-insensitive)
    local filtered = {}
    for _, cmd in ipairs(commands) do
        if query == "" or cmd.title:lower():find(query:lower(), 1, true) then
            filtered[#filtered + 1] = cmd
        end
    end

    -- Clamp selected index
    if selectedIdx > #filtered then
        selectedIdx = #filtered
    end
    if selectedIdx < 1 then
        selectedIdx = 1
    end

    -- Build list items
    local items = {}
    for i, cmd in ipairs(filtered) do
        local isSelected = (i == selectedIdx)
        local prefix = isSelected and "> " or "  "
        items[#items + 1] = lumina.createElement("text", {
            id = "cmd-" .. i,
            foreground = isSelected and (t.primary or "#89B4FA") or (t.text or "#CDD6F4"),
            bold = isSelected,
            style = {
                height = 1,
                background = isSelected and (t.surface1 or "#45475A") or "",
            },
        }, prefix .. cmd.title)
    end

    -- Keyboard handler
    local function handleKey(e)
        if e.key == "Escape" then
            if props.onClose then props.onClose() end
        elseif e.key == "Enter" then
            if filtered[selectedIdx] and filtered[selectedIdx].action then
                filtered[selectedIdx].action()
            end
            if props.onClose then props.onClose() end
        elseif e.key == "ArrowUp" or e.key == "k" then
            setSelectedIdx(math.max(1, selectedIdx - 1))
        elseif e.key == "ArrowDown" or e.key == "j" then
            setSelectedIdx(math.min(#filtered, selectedIdx + 1))
        elseif e.key == "Backspace" then
            if #query > 0 then
                setQuery(query:sub(1, -2))
                setSelectedIdx(1)
            end
        elseif #e.key == 1 then
            -- Single character input
            setQuery(query .. e.key)
            setSelectedIdx(1)
        end
    end

    local paletteWidth = props.width or 50
    local maxHeight = props.maxHeight or 15
    local visibleItems = math.min(#filtered, maxHeight - 3)
    local paletteHeight = visibleItems + 3

    -- Divider width
    local divWidth = paletteWidth - 4
    if divWidth < 1 then divWidth = 1 end

    return lumina.createElement("vbox", {
        id = "command-palette",
        style = {
            border = "rounded",
            width = paletteWidth,
            height = paletteHeight,
            background = t.surface0 or "#313244",
        },
        onKeyDown = handleKey,
        focusable = true,
    },
        -- Search input display
        lumina.createElement("text", {
            id = "cp-input",
            foreground = t.text or "#CDD6F4",
            style = { height = 1 },
        }, "> " .. query .. "▏"),

        -- Divider
        lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            dim = true,
            style = { height = 1 },
        }, string.rep("─", divWidth)),

        -- Command list
        table.unpack(items)
    )
end)

return CommandPalette

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
	preloadLuaSource(L, "lux.slot", luxSlotLua)
	preloadLuaSource(L, "lux.card", luxCardLua)
	preloadLuaSource(L, "lux.badge", luxBadgeLua)
	preloadLuaSource(L, "lux.divider", luxDividerLua)
	preloadLuaSource(L, "lux.progress", luxProgressLua)
	preloadLuaSource(L, "lux.spinner", luxSpinnerLua)
	preloadLuaSource(L, "lux.dialog", luxDialogLua)
	preloadLuaSource(L, "lux.layout", luxLayoutLua)
	preloadLuaSource(L, "lux.command_palette", luxCommandPaletteLua)

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
