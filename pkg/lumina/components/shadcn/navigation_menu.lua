-- navigation_menu.lua — Top navigation with dropdowns
local lumina = require("lumina")

local NavigationMenu = lumina.defineComponent({
    name = "ShadcnNavigationMenu",
    render = function(self)
        local items = self.props.items or {}
        local activeItem = self.props.active or ""
        local onSelect = self.props.onSelect

        local children = {}
        for _, item in ipairs(items) do
            local isActive = item.id == activeItem or item.label == activeItem
            children[#children + 1] = {
                type = "text",
                content = item.label or item.id or "",
                style = {
                    foreground = isActive and "#89B4FA" or "#CDD6F4",
                    bold = isActive,
                    underline = isActive,
                    padding = 1,
                },
            }
            -- Add separator between items
            if _ < #items then
                children[#children + 1] = {
                    type = "text",
                    content = " │ ",
                    style = { foreground = "#45475A" },
                }
            end
        end

        return {
            type = "hbox",
            style = {
                border = self.props.border or "single",
                background = self.props.background or "#1E1E2E",
                padding = 0,
                height = self.props.height or 1,
            },
            children = children,
        }
    end,
})

return NavigationMenu
