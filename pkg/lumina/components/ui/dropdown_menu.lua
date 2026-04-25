-- shadcn/dropdown_menu — Dropdown menu with items
local lumina = require("lumina")

local DropdownMenu = lumina.defineComponent({
    name = "ShadcnDropdownMenu",
    init = function(props)
        return {
            open = props.open or false,
            items = props.items or {},
            trigger = props.trigger or "Menu",
            selectedIndex = 0,
        }
    end,
    render = function(self)
        local children = {}
        -- Trigger
        children[#children + 1] = {
            type = "text",
            content = self.trigger .. " ▾",
            style = { foreground = "#E2E8F0", bold = true },
        }
        -- Menu items (when open)
        if self.open then
            local menuItems = {}
            for i, item in ipairs(self.items) do
                if item.separator then
                    menuItems[#menuItems + 1] = {
                        type = "text",
                        content = "──────────────────",
                        style = { foreground = "#334155" },
                    }
                else
                    local isSelected = (i == self.selectedIndex)
                    local label = item.label or item
                    local shortcut = item.shortcut or ""
                    local content = "  " .. label
                    if shortcut ~= "" then
                        content = content .. string.rep(" ", 20 - #label) .. shortcut
                    end
                    menuItems[#menuItems + 1] = {
                        type = "text",
                        content = content,
                        style = {
                            foreground = item.disabled and "#475569" or (isSelected and "#E2E8F0" or "#94A3B8"),
                            background = isSelected and "#334155" or "",
                        },
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
        return {
            type = "vbox",
            children = children,
        }
    end
})

return DropdownMenu
