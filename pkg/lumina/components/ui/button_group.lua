-- shadcn/button_group — Group of buttons
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    border = "#45475A",
}

local ButtonGroup = lumina.defineComponent({
    name = "ShadcnButtonGroup",

    init = function(props)
        return {
            children = props.children or {},
            orientation = props.orientation or "horizontal", -- horizontal, vertical
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local style = {
            border = "single",
            borderColor = c.border,
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
            type = self.orientation == "vertical" and "vbox" or "hbox",
            id = self.id,
            style = style,
            children = children,
        }
    end,
})

return ButtonGroup