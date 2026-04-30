-- lua/lux/progress.lua
-- Progress bar component for Lux
-- Usage: local Progress = require("lux.progress")

local Progress = lumina.defineComponent("Progress", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local value = math.max(0, math.min(100, props.value or 0))
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
