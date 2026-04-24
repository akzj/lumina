-- sidebar.lua — Application sidebar layout
local lumina = require("lumina")

local Sidebar = lumina.defineComponent({
    name = "ShadcnSidebar",
    render = function(self)
        local collapsed = self.props.collapsed or false
        local width = collapsed and 3 or (self.props.width or 20)
        local side = self.props.side or "left"
        local items = self.props.items or {}

        local children = {}
        for _, item in ipairs(items) do
            local icon = item.icon or ""
            local label = item.label or ""
            local active = item.active or false

            if collapsed then
                children[#children + 1] = {
                    type = "text",
                    content = icon,
                    style = {
                        foreground = active and "#89B4FA" or "#CDD6F4",
                        bold = active,
                        padding = 0,
                    },
                }
            else
                children[#children + 1] = {
                    type = "hbox",
                    style = {
                        background = active and "#313244" or "",
                        foreground = active and "#89B4FA" or "#CDD6F4",
                        padding = 0,
                    },
                    children = {
                        { type = "text", content = icon .. " " .. label,
                          style = { bold = active } },
                    },
                }
            end
        end

        -- Add custom children after items
        if self.props.children then
            for _, child in ipairs(self.props.children) do
                children[#children + 1] = child
            end
        end

        return {
            type = "vbox",
            style = {
                width = width,
                border = self.props.border or "single",
                background = self.props.background or "#181825",
                height = self.props.height or "100%",
            },
            children = children,
        }
    end,
})

local SidebarLayout = lumina.defineComponent({
    name = "ShadcnSidebarLayout",
    render = function(self)
        local side = self.props.side or "left"
        local sidebarChildren = {}
        local contentChildren = {}

        -- Separate sidebar from content
        for _, child in ipairs(self.props.children or {}) do
            if type(child) == "table" and child._isSidebar then
                sidebarChildren[#sidebarChildren + 1] = child
            else
                contentChildren[#contentChildren + 1] = child
            end
        end

        local content = {
            type = "vbox",
            style = { flex = 1 },
            children = contentChildren,
        }

        if side == "left" then
            return {
                type = "hbox",
                style = { width = "100%", height = "100%" },
                children = {
                    sidebarChildren[1] or { type = "text", content = "" },
                    content,
                },
            }
        else
            return {
                type = "hbox",
                style = { width = "100%", height = "100%" },
                children = {
                    content,
                    sidebarChildren[1] or { type = "text", content = "" },
                },
            }
        end
    end,
})

return {
    Sidebar = Sidebar,
    Layout = SidebarLayout,
}
