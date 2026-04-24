-- shadcn/checkbox — Checkbox with checked/unchecked/indeterminate
local lumina = require("lumina")

local Checkbox = lumina.defineComponent({
    name = "ShadcnCheckbox",
    init = function(props)
        return {
            checked = props.checked or false,
            indeterminate = props.indeterminate or false,
            disabled = props.disabled or false,
            label = props.label or "",
        }
    end,
    render = function(self)
        local icon
        if self.indeterminate then
            icon = "[-]"
        elseif self.checked then
            icon = "[✓]"
        else
            icon = "[ ]"
        end

        local fg = self.disabled and "#475569" or (self.checked and "#3B82F6" or "#94A3B8")
        local children = {
            { type = "text", content = icon, style = { foreground = fg, bold = self.checked } },
        }
        if self.label ~= "" then
            children[#children + 1] = {
                type = "text",
                content = " " .. self.label,
                style = { foreground = self.disabled and "#475569" or "#E2E8F0" },
            }
        end
        return {
            type = "hbox",
            children = children,
        }
    end
})

return Checkbox
