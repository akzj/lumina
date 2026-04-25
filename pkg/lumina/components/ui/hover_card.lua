-- shadcn/hover_card — Card that appears on hover
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
}

local HoverCard = lumina.defineComponent({
    name = "ShadcnHoverCard",

    init = function(props)
        return {
            open = props.open or false,
            id = props.id,
            className = props.className,
            style = props.style,
            children = props.content or props.children or {},
        }
    end,

    render = function(self)
        if not self.open then
            return { type = "empty" }
        end

        local style = {
            border = "rounded",
            borderColor = c.border,
            background = c.bg,
            foreground = c.fg,
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
                for i, child in ipairs(contentChildren) do
                    children[i] = child
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

return HoverCard
