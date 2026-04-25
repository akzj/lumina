-- shadcn/input_group — Input with prefix/suffix addons
local lumina = require("lumina")

local InputGroup = lumina.defineComponent({
    name = "ShadcnInputGroup",
    init = function(props)
        return {
            prefix = props.prefix or "",
            suffix = props.suffix or "",
            value = props.value or "",
            placeholder = props.placeholder or "",
            disabled = props.disabled or false,
        }
    end,
    render = function(self)
        local children = {}
        local fg = self.disabled and "#475569" or "#E2E8F0"

        if self.prefix ~= "" then
            children[#children + 1] = {
                type = "text",
                content = self.prefix .. " │ ",
                style = { foreground = "#64748B" },
            }
        end

        local display = self.value ~= "" and self.value or self.placeholder
        local inputFg = self.value == "" and "#64748B" or fg
        children[#children + 1] = {
            type = "text",
            content = display,
            style = { foreground = inputFg },
        }

        if self.suffix ~= "" then
            children[#children + 1] = {
                type = "text",
                content = " │ " .. self.suffix,
                style = { foreground = "#64748B" },
            }
        end

        return {
            type = "hbox",
            style = {
                border = "rounded",
                foreground = fg,
                padding = 1,
                height = 3,
            },
            children = children,
        }
    end
})

return InputGroup
