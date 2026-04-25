-- shadcn/context_menu — Right-click context menu
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
    surface = "#313244",
}

local ContextMenu = lumina.defineComponent({
    name = "ShadcnContextMenu",

    init = function(props)
        return {
            open = props.open or false,
            items = props.items or {},
            x = props.x or 0,
            y = props.y or 0,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        if not self.open then
            return { type = "empty" }
        end

        local menuItems = {}
        for _, item in ipairs(self.items) do
            if item.separator then
                menuItems[#menuItems + 1] = {
                    type = "text",
                    content = "──────────────────",
                    style = { foreground = c.border },
                }
            else
                local label = type(item) == "table" and (item.label or tostring(item)) or tostring(item)
                local disabled = type(item) == "table" and (item.disabled or false) or false

                menuItems[#menuItems + 1] = {
                    type = "text",
                    content = "  " .. label,
                    style = { foreground = disabled and c.muted or c.fg },
                }
            end
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

        return {
            type = "vbox",
            id = self.id,
            style = style,
            children = menuItems,
        }
    end,
})

return ContextMenu