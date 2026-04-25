-- shadcn/input_group — Input with label and helper text
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    border = "#45475A",
}

local InputGroup = lumina.defineComponent({
    name = "ShadcnInputGroup",

    init = function(props)
        return {
            label = props.label or "",
            helper = props.helper or "",
            children = props.children or {},
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local children = {}

        if self.label ~= "" then
            children[#children + 1] = {
                type = "text",
                content = self.label,
                style = { foreground = c.fg },
            }
        end

        local contentChildren = self.children
        if type(contentChildren) == "table" then
            if contentChildren.type then
                children[#children + 1] = contentChildren
            else
                for _, child in ipairs(contentChildren) do
                    children[#children + 1] = child
                end
            end
        end

        if self.helper ~= "" then
            children[#children + 1] = {
                type = "text",
                content = self.helper,
                style = { foreground = c.muted },
            }
        end

        return {
            type = "vbox",
            id = self.id,
            style = self.style or {},
            children = children,
        }
    end,
})

return InputGroup