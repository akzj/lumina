-- shadcn/radio_group — Radio button group
local lumina = require("lumina")

local RadioGroup = lumina.defineComponent({
    name = "ShadcnRadioGroup",
    init = function(props)
        return {
            options = props.options or {},
            value = props.value or "",
            disabled = props.disabled or false,
            orientation = props.orientation or "vertical",
        }
    end,
    render = function(self)
        local children = {}
        for _, opt in ipairs(self.options) do
            local val = type(opt) == "table" and opt.value or opt
            local label = type(opt) == "table" and (opt.label or opt.value) or opt
            local selected = (val == self.value)
            local icon = selected and "(●)" or "( )"
            local fg = self.disabled and "#475569" or (selected and "#3B82F6" or "#94A3B8")

            children[#children + 1] = {
                type = "hbox",
                children = {
                    { type = "text", content = icon, style = { foreground = fg, bold = selected } },
                    { type = "text", content = " " .. label, style = { foreground = self.disabled and "#475569" or "#E2E8F0" } },
                },
            }
        end
        return {
            type = self.orientation == "horizontal" and "hbox" or "vbox",
            children = children,
        }
    end
})

return RadioGroup
