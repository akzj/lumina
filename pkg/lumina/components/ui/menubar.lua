-- shadcn/menubar — Horizontal menu bar with dropdown submenus
local lumina = require("lumina")

local Menubar = lumina.defineComponent({
    name = "ShadcnMenubar",
    init = function(props)
        return {
            menus = props.menus or {},
            activeMenu = props.activeMenu or 0,
        }
    end,
    render = function(self)
        local menuButtons = {}
        for i, menu in ipairs(self.menus) do
            local isActive = (i == self.activeMenu)
            menuButtons[#menuButtons + 1] = {
                type = "text",
                content = " " .. (menu.label or "Menu") .. " ",
                style = {
                    foreground = isActive and "#E2E8F0" or "#94A3B8",
                    background = isActive and "#334155" or "",
                    bold = isActive,
                },
            }
        end

        local children = {
            {
                type = "hbox",
                style = { background = "#0F172A", border = "rounded" },
                children = menuButtons,
            },
        }

        -- Show active menu dropdown
        if self.activeMenu > 0 then
            local activeMenu = self.menus[self.activeMenu]
            if activeMenu and activeMenu.items then
                local menuItems = {}
                for _, item in ipairs(activeMenu.items) do
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
                            style = { foreground = "#E2E8F0" },
                        }
                    end
                end
                children[#children + 1] = {
                    type = "vbox",
                    style = {
                        border = "rounded",
                        background = "#1E293B",
                        padding = 1,
                    },
                    children = menuItems,
                }
            end
        end

        return {
            type = "vbox",
            children = children,
        }
    end
})

return Menubar
