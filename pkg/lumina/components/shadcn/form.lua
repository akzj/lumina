-- shadcn/form — Form container with validation
local lumina = require("lumina")

local Form = lumina.defineComponent({
    name = "ShadcnForm",
    init = function(props)
        return {
            fields = props.fields or {},
            errors = props.errors or {},
            onSubmit = props.onSubmit,
        }
    end,
    render = function(self)
        local children = {}

        -- Render child fields
        if self.props and self.props.children then
            for _, child in ipairs(self.props.children) do
                children[#children + 1] = child
            end
        end

        -- Show submit button area
        children[#children + 1] = {
            type = "text",
            content = "",
            style = { foreground = "#334155" },
        }

        return {
            type = "vbox",
            style = { gap = 1 },
            children = children,
        }
    end
})

return Form
