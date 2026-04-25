-- shadcn/drawer — Bottom drawer panel
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
}

local Drawer = lumina.defineComponent({
    name = "ShadcnDrawer",

    init = function(props)
        return {
            open = props.open or false,
            title = props.title or "",
            height = props.height or 10,
            id = props.id,
            className = props.className,
            style = props.style,
            children = props.children or {},
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
            height = self.height,
        }

        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        local children = {}

        -- Handle/grab bar
        children[#children + 1] = {
            type = "hbox",
            style = { justify = "center", paddingTop = 0 },
            children = {
                { type = "text", content = "━━━━", style = { foreground = c.border } },
            },
        }

        if self.title ~= "" then
            children[#children + 1] = {
                type = "text",
                content = self.title,
                style = { bold = true, foreground = c.primary },
            }
        end

        -- Content children
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

        return {
            type = "vbox",
            id = self.id,
            style = style,
            children = children,
        }
    end,
})

return Drawer
