-- shadcn/separator — Horizontal/vertical separator
local lumina = require("lumina")

local Separator = lumina.defineComponent({
    name = "ShadcnSeparator",
    init = function(props)
        return {
            orientation = props.orientation or "horizontal",
        }
    end,
    render = function(self)
        if self.orientation == "vertical" then
            return {
                type = "text",
                content = "│",
                style = { foreground = "#334155" },
            }
        end
        return {
            type = "text",
            content = "────────────────────────────────────────",
            style = { foreground = "#334155" },
        }
    end
})

return Separator
