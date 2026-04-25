-- shadcn/scroll_area — Scrollable content area
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
}

local ScrollArea = lumina.defineComponent({
    name = "ShadcnScrollArea",

    init = function(props)
        return {
            height = props.height or 10,
            id = props.id,
            className = props.className,
            style = props.style,
            children = props.children or {},
        }
    end,

    render = function(self)
        local style = {
            border = "rounded",
            borderColor = c.border,
            background = c.bg,
            foreground = c.fg,
            padding = 1,
            height = self.height,
            overflow = "scroll", -- terminal: just renders content
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

return ScrollArea