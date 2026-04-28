-- lua/lux/divider.lua
-- Divider component for Lux
-- Usage: local Divider = require("lux.divider")

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
