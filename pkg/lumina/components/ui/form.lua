-- shadcn/form — Form container with validation
local lumina = require("lumina")

local c = {
    border = "#45475A",
    bg = "#181825",
}

local Form = lumina.defineComponent({
    name = "ShadcnForm",

    init = function(props)
        return {
            children = props.children or {},
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local style = {
            border = "rounded",
            borderColor = c.border,
            background = c.bg,
            padding = 1,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        local contentChildren = self.children
        local children = {}
        if type(contentChildren) == "table" then
            if contentChildren.type then
                children[1] = contentChildren
            else
                for _, child in ipairs(contentChildren) do
                    children[#children + 1] = child
                end
            end
        end

        return {
            type = "vbox",
            id = self.id,
            style = style,
            children = children,
        }
    end,
})

return Form