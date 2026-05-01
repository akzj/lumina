-- lua/lux/badge.lua
-- Badge component for Lux
-- Usage: local Badge = require("lux.badge")

local Badge = lumina.defineComponent("Badge", function(props)
    local variant = props.variant or "default"
    local t = lumina.getTheme and lumina.getTheme() or {}
    local fg, bg
    if variant == "success" then
        fg = t.success or "#4ADE80"; bg = t.surface0 or "#141C2C"
    elseif variant == "warning" then
        fg = t.warning or "#F5C842"; bg = t.surface0 or "#141C2C"
    elseif variant == "error" then
        fg = t.error or "#F87171"; bg = t.surface0 or "#141C2C"
    else
        fg = t.primary or "#F5C842"; bg = t.surface0 or "#141C2C"
    end

    return lumina.createElement("text", {
        foreground = fg,
        background = bg,
        bold = true,
    }, " " .. (props.label or "") .. " ")
end)

return Badge
