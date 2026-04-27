-- shadcn/textarea — Multi-line text input
local lumina = require("lumina")

local Textarea = lumina.defineComponent({
    name = "ShadcnTextarea",
    init = function(props)
        return {
            value = props.value or "",
            placeholder = props.placeholder or "",
            disabled = props.disabled or false,
            rows = props.rows or 4,
        }
    end,
    render = function(self)
        local display = self.value ~= "" and self.value or self.placeholder
        local fg = self.value == "" and "#64748B" or "#E2E8F0"
        if self.disabled then fg = "#475569" end

        return {
            type = "vbox",
            style = {
                border = "rounded",
                foreground = fg,
                padding = 1,
                height = self.rows + 2,
            },
            children = {
                { type = "text", content = display },
            },
        }
    end
})

return Textarea
