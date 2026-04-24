-- shadcn/context_menu — Right-click context menu
local lumina = require("lumina")

local ContextMenu = lumina.defineComponent({
    name = "ShadcnContextMenu",
    init = function(props)
        return {
            open = props.open or false,
            items = props.items or {},
            x = props.x or 0,
            y = props.y or 0,
        }
    end,
    render = function(self)
        if not self.open then
            return { type = "fragment", children = {} }
        end
        local menuItems = {}
        for _, item in ipairs(self.items) do
            if item.separator then
                menuItems[#menuItems + 1] = {
                    type = "text",
                    content = "──────────────────",
                    style = { foreground = "#334155" },
                }
            else
                menuItems[#menuItems + 1] = {
                    type = "text",
                    content = "  " .. (item.label or item),
                    style = {
                        foreground = item.disabled and "#475569" or "#E2E8F0",
                    },
                }
            end
        end
        return {
            type = "vbox",
            style = {
                border = "rounded",
                background = "#1E293B",
                foreground = "#E2E8F0",
                padding = 1,
            },
            children = menuItems,
        }
    end
})

return ContextMenu
