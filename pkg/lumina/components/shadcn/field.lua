-- shadcn/field — Form field wrapper (label + input + error)
local lumina = require("lumina")

local Field = lumina.defineComponent({
    name = "ShadcnField",
    init = function(props)
        return {
            label = props.label or "",
            error = props.error or "",
            description = props.description or "",
            required = props.required or false,
        }
    end,
    render = function(self)
        local children = {}

        -- Label
        if self.label ~= "" then
            local labelText = self.label
            if self.required then labelText = labelText .. " *" end
            children[#children + 1] = {
                type = "text",
                content = labelText,
                style = { foreground = "#E2E8F0", bold = true },
            }
        end

        -- Description
        if self.description ~= "" then
            children[#children + 1] = {
                type = "text",
                content = self.description,
                style = { foreground = "#94A3B8" },
            }
        end

        -- Slot for input (children from props)
        if self.props and self.props.children then
            for _, child in ipairs(self.props.children) do
                children[#children + 1] = child
            end
        end

        -- Error message
        if self.error ~= "" then
            children[#children + 1] = {
                type = "text",
                content = "⚠ " .. self.error,
                style = { foreground = "#EF4444" },
            }
        end

        return {
            type = "vbox",
            style = { gap = 1 },
            children = children,
        }
    end
})

return Field
