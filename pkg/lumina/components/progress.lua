-- Progress bar component for Lumina
local lumina = require("lumina")

local Progress = lumina.defineComponent({
    name = "Progress",
    init = function(props)
        return {
            value = props.value or 0,
            width = props.width or 30,
            showPercent = props.showPercent ~= false,
            style = props.style or {},
        }
    end,
    render = function(instance)
        local value = instance.value or 0
        if value < 0 then value = 0 end
        if value > 1 then value = 1 end
        local barWidth = instance.width or 30
        local fg = instance.style and instance.style.foreground or "#00FF00"
        local bg = instance.style and instance.style.background
        local filled = math.floor(value * barWidth + 0.5)
        local empty = barWidth - filled
        local bar = string.rep("█", filled) .. string.rep("░", empty)
        local content = "[" .. bar .. "]"
        if instance.showPercent then
            content = content .. " " .. math.floor(value * 100 + 0.5) .. "%"
        end
        return { type = "text", content = content, foreground = fg, background = bg }
    end
})

return Progress
