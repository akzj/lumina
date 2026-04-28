local lumina = require("lumina")

local Progress = lumina.defineComponent("Progress", function(props)
    local value = props.value or 0  -- 0-100
    local width = props.width or 20
    local filled = math.floor(width * value / 100)
    local empty = width - filled

    local bar = string.rep("█", filled) .. string.rep("░", empty)
    local label = string.format(" %d%%", value)

    return lumina.createElement("hbox", {},
        lumina.createElement("text", {
            foreground = "#89B4FA",
        }, bar),
        lumina.createElement("text", {
            foreground = "#CDD6F4",
        }, label)
    )
end)

return Progress
