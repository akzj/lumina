-- lua/lux/badge.lua
-- Badge component for Lux
-- Usage: local Badge = require("lux.badge")

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
