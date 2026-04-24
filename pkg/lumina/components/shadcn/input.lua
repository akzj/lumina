-- shadcn/input — Styled text input wrapper
local lumina = require("lumina")

local Input = lumina.defineComponent({
    name = "ShadcnInput",
    init = function(props)
        return {
            placeholder = props.placeholder or "",
            value = props.value or "",
            disabled = props.disabled or false,
            inputType = props.type or "text",
        }
    end,
    render = function(self)
        local display = self.value
        if display == "" then
            display = self.placeholder
        end
        local fg = self.value == "" and "#64748B" or "#E2E8F0"
        if self.disabled then
            fg = "#475569"
        end
        return {
            type = "hbox",
            style = {
                border = "rounded",
                foreground = fg,
                padding = 1,
                height = 3,
            },
            children = {
                { type = "text", content = display },
            },
        }
    end
})

return Input
