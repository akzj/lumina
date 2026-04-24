-- shadcn/label — Label component
local lumina = require("lumina")

local Label = lumina.defineComponent({
    name = "ShadcnLabel",
    init = function(props)
        return {
            text = props.text or props.label or "",
            htmlFor = props.htmlFor or "",
        }
    end,
    render = function(self)
        return {
            type = "text",
            content = self.text,
            style = {
                foreground = "#E2E8F0",
            },
        }
    end
})

return Label
