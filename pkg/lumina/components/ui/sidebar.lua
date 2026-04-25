-- shadcn/sidebar — Application sidebar layout
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
    surface = "#313244",
}

-- Sidebar: Main sidebar container
local Sidebar = lumina.defineComponent({
    name = "ShadcnSidebar",

    init = function(props)
        return {
            collapsed = props.collapsed or false,
            width = props.width or 20,
            side = props.side or "left",
            items = props.items or {},
            id = props.id,
            className = props.className,
            style = props.style,
            children = props.children or {},
        }
    end,

    render = function(self)
        local width = self.collapsed and 3 or self.width
        local items = self.items or {}
        local children = {}

        for _, item in ipairs(items) do
            local icon = item.icon or ""
            local label = item.label or ""
            local active = item.active or false

            if self.collapsed then
                children[#children + 1] = {
                    type = "text",
                    content = icon,
                    style = {
                        foreground = active and c.primary or c.fg,
                        bold = active,
                    },
                }
            else
                children[#children + 1] = {
                    type = "hbox",
                    style = {
                        background = active and c.surface or "",
                        foreground = active and c.primary or c.fg,
                        paddingLeft = 0,
                        paddingRight = 0,
                    },
                    children = {
                        { type = "text", content = icon .. " " .. label, style = { bold = active } },
                    },
                }
            end
        end

        local extraChildren = self.children
        if type(extraChildren) == "table" then
            if extraChildren.type then
                children[#children + 1] = extraChildren
            else
                for _, child in ipairs(extraChildren) do
                    children[#children + 1] = child
                end
            end
        end

        local style = {
            border = "right",
            borderColor = c.border,
            background = c.bg,
            foreground = c.fg,
            padding = 0,
            width = width,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        return {
            type = "vbox",
            id = self.id,
            style = style,
            children = children,
        }
    end,
})

-- Layout: Sidebar + content layout
local Layout = lumina.defineComponent({
    name = "ShadcnSidebarLayout",

    init = function(props)
        return {
            sidebar = props.sidebar,
            content = props.content or {},
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local sidebarNode = self.sidebar
        local contentChildren = self.content

        local children = {}
        if sidebarNode then
            children[#children + 1] = sidebarNode
        end

        local contentBox
        if type(contentChildren) == "table" then
            if contentChildren.type then
                contentBox = contentChildren
            else
                contentBox = {
                    type = "vbox",
                    children = contentChildren,
                }
            end
        else
            contentBox = {
                type = "vbox",
                children = {},
            }
        end

        children[#children + 1] = contentBox

        return {
            type = "hbox",
            id = self.id,
            style = self.style or {},
            children = children,
        }
    end,
})

return {
    Sidebar = Sidebar,
    Layout = Layout,
}