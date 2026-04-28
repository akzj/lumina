local lumina = require("lumina")

local Badge = lumina.defineComponent("Badge", function(props)
    local variant = props.variant or "default"
    local fg, bg
    if variant == "success" then
        fg = "#A6E3A1"; bg = "#313244"
    elseif variant == "warning" then
        fg = "#F9E2AF"; bg = "#313244"
    elseif variant == "error" then
        fg = "#F38BA8"; bg = "#313244"
    else
        fg = "#89B4FA"; bg = "#313244"
    end

    return lumina.createElement("text", {
        foreground = fg,
        background = bg,
        bold = true,
    }, " " .. (props.label or "") .. " ")
end)

return Badge
