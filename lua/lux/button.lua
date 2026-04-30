-- lua/lux/button.lua — Lux Button (Prime-style severities + appearances + group + split).
-- Pure Lua / CSS style fields per docs/css-properties.md (border, borderColor, padding, flex).
-- Does not wrap lumina.Button (Go); supports richer visuals and merged props.style.
--
-- Usage:
--   local Button = require("lux.button")
--   Button { label = "OK", severity = "success", appearance = "solid", onClick = fn }
--   Button { label = "Cancel", appearance = "outlined", severity = "danger" }
--   Button { appearance = "text", severity = "info", label = "Link" }
--   Button { icon = "▶", label = "Run", iconPosition = "left" }
--   Button.Group { items = { { label = "Save", onClick = a }, { label = "Del", onClick = b } }, severity = "primary" }
--   Button.Split { label = "Action", onClick = main, onMenuClick = menu, severity = "primary" }
--
-- Backward compatible with previous Lux wrapper → Go mapping:
--   variant = "primary" | "secondary" | "outline" | "ghost" still work.

--- Non-empty theme string; Lua treats "" as truthy, so `t.surface0 or fallback` can yield "" and break paint.
local function themeStr(v, fallback)
    if type(v) ~= "string" or v == "" then
        return fallback
    end
    return v
end

--- #RRGGBB only: multiply RGB by mult (<1 darkens, >1 lightens). Invalid input returns hex unchanged.
local function adjustHexBrightness(hex, mult)
    if type(hex) ~= "string" or #hex ~= 7 or string.sub(hex, 1, 1) ~= "#" then
        return hex
    end
    local function byte(i)
        return tonumber(string.sub(hex, i, i + 1), 16)
    end
    local r, g, b = byte(2), byte(4), byte(6)
    if not r or not g or not b then
        return hex
    end
    r = math.max(0, math.min(255, math.floor(r * mult + 0.5)))
    g = math.max(0, math.min(255, math.floor(g * mult + 0.5)))
    b = math.max(0, math.min(255, math.floor(b * mult + 0.5)))
    return string.format("#%02X%02X%02X", r, g, b)
end

local function theme()
    if lumina.getTheme then
        return lumina.getTheme()
    end
    return {
        primary = "#89B4FA",
        primaryDark = "#1E1E2E",
        text = "#CDD6F4",
        muted = "#6C7086",
        surface0 = "#313244",
        surface1 = "#45475A",
        surface2 = "#585B70",
        success = "#A6E3A1",
        warning = "#F9E2AF",
        error = "#F38BA8",
        hover = "#B4BEFE",
        pressed = "#74C7EC",
    }
end

--- Solid / outlined / text colors per severity (bg, fg, edge for border & outlined fg).
-- All theme lookups use themeStr: empty string is truthy in Lua and would break paint (see danger → t.error).
local function palette(sev, t)
    local err = themeStr(t.error, themeStr(t.danger, "#F87171"))
    local p = {
        primary = {
            themeStr(t.primary, "#F5C842"),
            themeStr(t.primaryDark, "#0B1220"),
            themeStr(t.primary, "#F5C842"),
        },
        -- Use surface2 (not surface1): on dark shells surface1 blends into surface0/base.
        secondary = {
            themeStr(t.surface2, "#334B6B"),
            themeStr(t.text, "#E8EDF7"),
            themeStr(t.muted, "#8B9BB4"),
        },
        success = { themeStr(t.success, "#4ADE80"), "#0B1220", themeStr(t.success, "#4ADE80") },
        danger = { err, "#0B1220", err },
        warning = { themeStr(t.warning, "#F9E2AF"), "#0B1220", themeStr(t.warning, "#F9E2AF") },
        info = { "#8B5CF6", "#E8EDF7", "#A78BFA" },
        help = { "#EC4899", "#E8EDF7", "#F472B6" },
    }
    return table.unpack(p[sev] or p.primary)
end

-- Solid/raised: darken fill on hover/press. Secondary is already dark slate — lighten slightly on hover so it does not sink into the page.
local function hoverPressedBg(sev, t, baseBg, hovered, pressed)
    if pressed then
        if sev == "primary" then
            return themeStr(t.pressed, themeStr(t.primary, baseBg))
        end
        if sev == "secondary" then
            return adjustHexBrightness(baseBg, 0.9)
        end
        return adjustHexBrightness(baseBg, 0.72)
    end
    if hovered then
        if sev == "primary" then
            return themeStr(t.hover, baseBg)
        end
        if sev == "secondary" then
            return adjustHexBrightness(baseBg, 1.07)
        end
        return adjustHexBrightness(baseBg, 0.86)
    end
    return baseBg
end

--- Map legacy variant → severity + appearance.
local function normalizeLegacy(props)
    local sev = props.severity
    local app = props.appearance
    if props.variant and not app then
        local v = props.variant
        if v == "outline" then
            app = "outlined"
            sev = sev or "primary"
        elseif v == "ghost" then
            app = "text"
            sev = sev or "primary"
        elseif v == "secondary" then
            sev = "secondary"
            app = app or "solid"
        elseif v == "primary" then
            sev = "primary"
            app = app or "solid"
        elseif v == "danger" or v == "success" or v == "warning" or v == "info" or v == "help" then
            sev = v
            app = app or "solid"
        else
            sev = sev or "primary"
            app = app or "solid"
        end
    end
    return sev or "primary", app or "solid"
end

local LuxButton = lumina.defineComponent("LuxButton", function(props)
    local t = theme()
    local severity, appearance = normalizeLegacy(props)
    local disabled = props.disabled == true
    local hovered, setHovered = lumina.useState("btnHover", false)
    local pressed, setPressed = lumina.useState("btnPressed", false)

    local bg0, fg0, edge0 = palette(severity, t)
    local label = props.label or ""
    local icon = props.icon
    local iconPos = props.iconPosition or "left"
    local linkLike = props.link == true or appearance == "link"
    -- props.link alone must become text appearance (else we stay on solid and look like a filled pill).
    if appearance == "link" or props.link == true then
        appearance = "text"
    end

    local padX = props.paddingX
    if padX == nil then
        padX = props.size == "sm" and 1 or (props.size == "lg" and 3 or 2)
    end
    local padY = props.paddingY
    if padY == nil then
        padY = 0
    end

    local bg, fg, border, borderColor, dim, underline, height

    if disabled then
        bg = themeStr(t.surface0, "#313244")
        fg = themeStr(t.muted, "#6C7086")
        border = "rounded"
        borderColor = themeStr(t.surface2, "#585B70")
        dim = true
        underline = false
        height = (appearance == "text" or appearance == "link") and 1 or 3
    elseif appearance == "text" or appearance == "link" then
        bg = ""
        fg = edge0
        border = "none"
        borderColor = ""
        dim = false
        underline = linkLike or props.underline == true
        height = 1
    elseif appearance == "outlined" then
        bg = hovered and themeStr(t.surface0, "#313244") or ""
        fg = edge0
        border = props.shape == "square" and "single" or "rounded"
        borderColor = edge0
        dim = false
        underline = false
        height = 3
    elseif appearance == "raised" then
        bg = hoverPressedBg(severity, t, bg0, hovered, pressed)
        fg = fg0
        border = "single"
        borderColor = themeStr(t.surface2, "#585B70")
        dim = false
        underline = false
        height = 3
    else
        -- solid (default)
        bg = hoverPressedBg(severity, t, bg0, hovered, pressed)
        fg = fg0
        border = (props.shape == "square") and "single" or "rounded"
        borderColor = edge0
        dim = false
        underline = false
        height = 3
    end

    local function labelNode()
        return lumina.createElement("text", {
            bold = props.bold ~= false and not disabled,
            dim = dim,
            underline = underline,
            foreground = fg,
            background = (appearance == "text" or appearance == "link") and bg or "",
            style = { height = 1 },
        }, label)
    end

    local function iconNode()
        if not icon or icon == "" then
            return nil
        end
        return lumina.createElement("text", {
            foreground = fg,
            background = (appearance == "text") and bg or "",
            style = { height = 1 },
        }, icon)
    end

    local inner = {}
    if icon and icon ~= "" then
        if iconPos == "right" then
            if label ~= "" then
                inner[#inner + 1] = labelNode()
            end
            inner[#inner + 1] = iconNode()
        else
            inner[#inner + 1] = iconNode()
            if label ~= "" then
                inner[#inner + 1] = labelNode()
            end
        end
    else
        inner[#inner + 1] = labelNode()
    end

    local rootStyle = {
        height = height,
        minWidth = props.minWidth,
        width = props.width,
        border = border,
        background = bg,
        paddingLeft = padX,
        paddingRight = padX,
        paddingTop = padY,
        paddingBottom = padY,
        justify = "center",
        align = "center",
        gap = (icon and icon ~= "" and label ~= "") and 1 or 0,
        flexShrink = props.flexShrink,
        flex = props.flex,
        alignSelf = props.alignSelf,
    }
    if props.iconOnly or (icon and icon ~= "" and label == "") then
        rootStyle.minWidth = rootStyle.minWidth or 3
        rootStyle.width = rootStyle.width or 3
    end
    if borderColor and borderColor ~= "" then
        rootStyle.borderColor = borderColor
        rootStyle.foreground = borderColor
    end
    if type(props.style) == "table" then
        for k, v in pairs(props.style) do
            rootStyle[k] = v
        end
    end

    return lumina.createElement("hbox", {
        id = props.id,
        key = props.key,
        focusable = not disabled,
        disabled = disabled,
        style = rootStyle,
        onClick = (not disabled and props.onClick) and function()
            props.onClick()
        end or nil,
        onMouseEnter = not disabled and function()
            setHovered(true)
        end or nil,
        onMouseLeave = not disabled and function()
            setHovered(false)
            setPressed(false)
        end or nil,
        onMouseDown = not disabled and function()
            setPressed(true)
        end or nil,
        onMouseUp = not disabled and function()
            setPressed(false)
        end or nil,
        onKeyDown = (not disabled and props.onClick) and function(e)
            local k = type(e) == "table" and (e.key or e.Key) or tostring(e or "")
            if k == " " or k == "Enter" then
                props.onClick()
            end
        end or nil,
    }, table.unpack(inner))
end)

--- Fused row: shared chrome, vertical dividers, gap = 0.
local LuxButtonGroup = lumina.defineComponent("LuxButtonGroup", function(props)
    local t = theme()
    local sev = props.severity or "primary"
    local disabled = props.disabled == true
    local bg0, fg0, edge0 = palette(sev, t)
    local bg = disabled and themeStr(t.surface0, "#313244") or bg0
    local fg = disabled and themeStr(t.muted, "#6C7086") or fg0
    local items = props.items or {}

    local cells = {}
    for i, it in ipairs(items) do
        if i > 1 then
            cells[#cells + 1] = lumina.createElement("text", {
                key = "gsep-" .. tostring(i),
                foreground = fg,
                background = bg,
                dim = true,
                style = { height = 3, width = 1 },
            }, "│")
        end
        cells[#cells + 1] = lumina.createElement("text", {
            key = "gbtn-" .. tostring(i),
            foreground = fg,
            background = bg,
            bold = true,
            dim = disabled,
            style = { flex = 1, height = 3, paddingLeft = 1, paddingRight = 1, align = "center" },
            onClick = (not disabled and it.onClick) and function()
                it.onClick()
            end or nil,
        }, it.label or "")
    end

    local st = {
        border = "rounded",
        borderColor = edge0,
        background = bg,
        gap = 0,
        width = props.width,
        flex = props.flex,
    }
    if type(props.style) == "table" then
        for k, v in pairs(props.style) do
            st[k] = v
        end
    end

    return lumina.createElement("hbox", {
        id = props.id,
        key = props.key,
        style = st,
    }, table.unpack(cells))
end)

--- Main action + menu chevron (two click targets).
local LuxSplitButton = lumina.defineComponent("LuxSplitButton", function(props)
    local t = theme()
    local sev = props.severity or "primary"
    local disabled = props.disabled == true
    local bg0, fg0, edge0 = palette(sev, t)
    local bg = disabled and themeStr(t.surface0, "#313244") or bg0
    local fg = disabled and themeStr(t.muted, "#6C7086") or fg0

    local st = {
        border = "rounded",
        borderColor = edge0,
        background = bg,
        gap = 0,
        height = 3,
        width = props.width,
    }
    if type(props.style) == "table" then
        for k, v in pairs(props.style) do
            st[k] = v
        end
    end

    return lumina.createElement("hbox", {
        id = props.id,
        key = props.key,
        style = st,
    },
        lumina.createElement("text", {
            foreground = fg,
            background = bg,
            bold = true,
            dim = disabled,
            style = { flex = 1, height = 3, paddingLeft = 2, paddingRight = 1, justify = "center" },
            onClick = (not disabled and props.onClick) and function()
                props.onClick()
            end or nil,
        }, props.label or "Action"),
        lumina.createElement("text", {
            foreground = fg,
            background = bg,
            dim = disabled,
            style = { height = 3, width = 3, justify = "center" },
            onClick = (not disabled and props.onMenuClick) and function()
                props.onMenuClick()
            end or nil,
        }, "▼")
    )
end)

LuxButton.Group = LuxButtonGroup
LuxButton.Split = LuxSplitButton

return LuxButton
