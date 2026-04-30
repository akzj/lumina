-- lua/lux/layout.lua
-- Layout: standard TUI app structure with Header, Sidebar, Main, Footer slots.
--
-- Main scroll: `Layout.Main { scroll = true }` sets overflow=scroll on the main
-- slot. Scroll containers lay out flow children at natural height: containers with
-- children get an intrinsic-height pass (see pkg/render/layout.go scroll branch),
-- so a single inner vbox of cards is valid.
--
-- Usage:
--   local Layout = require("lux.layout")
--   Layout {
--       Layout.Header { height = 1, Text { "Title" } },
--       Layout.Sidebar { width = 20, NavMenu {} },
--       Layout.Main { Content {} },
--       Layout.Footer { height = 1, Text { "Ready" } },
--   }

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
        local mainStyle = { flex = 1 }
        local mainOverflow = slotProp(main, "_overflow", "")
        if mainOverflow ~= "" then
            mainStyle.overflow = mainOverflow
        end
        local sbStyle = {
            width = slotProp(sidebar, "_width", 20),
            border = slotProp(sidebar, "_border", "none"),
            background = slotProp(sidebar, "_bg", ""),
        }
        local sbBorderColor = slotProp(sidebar, "_borderColor", "")
        if sbBorderColor ~= "" then
            sbStyle.borderColor = sbBorderColor
        end
        vboxChildren[#vboxChildren + 1] = lumina.createElement("hbox", {
            style = { flex = 1 },
        },
            lumina.createElement("vbox", {
                style = sbStyle,
            }, table.unpack(slotChildren(sidebar))),
            lumina.createElement("vbox", {
                style = mainStyle,
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
        _borderColor = props.borderColor,
        _bg = props.background or props.bg,
    }, table.unpack(children))
end

function Layout.Main(props)
    local children = {}
    for i, v in ipairs(props) do
        children[i] = v
    end
    local overflow = props.overflow
    if overflow == nil and props.scroll then
        overflow = "scroll"
    end
    return lumina.createElement("vbox", {
        _layoutSlot = "main",
        _overflow = overflow or "",
    }, table.unpack(children))
end

return Layout
