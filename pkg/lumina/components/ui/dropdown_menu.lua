-- shadcn/dropdown_menu — Dropdown menu with items
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
    surface = "#313244",
}

local DropdownMenu = lumina.defineComponent({
    name = "ShadcnDropdownMenu",

    init = function(props)
        return {
            open = props.open or false,
            items = props.items or {},
            trigger = props.trigger or "Menu",
            selectedIndex = 0,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local children = {}

        -- Trigger
        children[#children + 1] = {
            type = "text",
            content = self.trigger .. " ▾",
            style = { foreground = c.fg, bold = true },
        }

        -- Menu items (when open)
        if self.open then
            local menuItems = {}
            for i, item in ipairs(self.items) do
                if item.separator then
                    menuItems[#menuItems + 1] = {
                        type = "text",
                        content = "──────────────────",
                        style = { foreground = c.border },
                    }
                else
                    local isSelected = (i == self.selectedIndex)
                    local label = type(item) == "table" and (item.label or tostring(item)) or tostring(item)
                    local shortcut = type(item) == "table" and (item.shortcut or "") or ""
                    local disabled = type(item) == "table" and (item.disabled or false) or false

                    local content = "  " .. label
                    if shortcut ~= "" then
                        content = content .. string.rep(" ", math.max(0, 20 - #label)) .. shortcut
                    end

                    local fg = disabled and c.muted or (isSelected and c.fg or c.muted)
                    local bg = isSelected and c.surface or ""

                    menuItems[#menuItems + 1] = {
                        type = "text",
                        content = content,
                        style = { foreground = fg, background = bg },
                    }
                end
            end

            children[#children + 1] = {
                type = "vbox",
                style = {
                    border = "rounded",
                    borderColor = c.border,
                    background = c.bg,
                    foreground = c.fg,
                    padding = 1,
                },
                children = menuItems,
            }
        end

        return {
            type = "vbox",
            id = self.id,
            style = self.style or {},
            children = children,
        }
    end,
})

return DropdownMenu