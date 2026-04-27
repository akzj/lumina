-- shadcn/progress — Progress bar
local lumina = require("lumina")

local Progress = lumina.defineComponent({
    name = "ShadcnProgress",
    init = function(props)
        return {
            value = props.value or 0,
            max = props.max or 100,
            width = props.width or 30,
            showLabel = props.showLabel or false,
        }
    end,
    render = function(self)
        local pct = math.min(1, math.max(0, self.value / self.max))
        local filled = math.floor(pct * self.width)
        local empty = self.width - filled
        local bar = string.rep("█", filled) .. string.rep("░", empty)

        local children = {
            { type = "text", content = bar, style = { foreground = "#3B82F6" } },
        }
        if self.showLabel then
            children[#children + 1] = {
                type = "text",
                content = " " .. math.floor(pct * 100) .. "%",
                style = { foreground = "#94A3B8" },
            }
        end
        return {
            type = "hbox",
            children = children,
        }
    end
})

return Progress
