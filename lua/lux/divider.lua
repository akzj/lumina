local lumina = require("lumina")

local Divider = lumina.defineComponent("Divider", function(props)
    local char = props.char or "─"
    local width = props.width or 40
    return lumina.createElement("text", {
        foreground = props.fg or "#45475A",
        dim = true,
    }, string.rep(char, width))
end)

return Divider
