-- shadcn/navigation_menu — Top navigation bar
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
    surface = "#313244",
}

local NavigationMenu = lumina.defineComponent({
    name = "ShadcnNavigationMenu",

    init = function(props)
        return {
            items = props.items or {},
            activeItem = props.activeItem or 1,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local children = {}

        for i, item in ipairs(self.items) do
            local label = type(item) == "table" and (item.label or item.title) or tostring(item)
            local isActive = (i == self.activeItem)

            local style = {
                foreground = isActive and c.primary or c.muted,
                background = isActive and c.surface or "",
                bold = isActive,
                paddingLeft = 1,
                paddingRight = 1,
            }

            if isActive then
                style.borderBottom = "single"
                style.borderColor = c.primary
            end

            children[#children + 1] = {
                type = "hbox",
                id = self.id and (self.id .. "-item-" .. i) or nil,
                style = style,
                children = {
                    { type = "text", content = " " .. label .. " " },
                },
            }

            if i < #self.items then
                children[#children + 1] = {
                    type = "text",
                    content = "│",
                    style = { foreground = c.border },
                }
            end
        end

        local containerStyle = {
            background = c.bg,
            borderBottom = "single",
            borderColor = c.border,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do containerStyle[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do containerStyle[k] = v end
        end

        return {
            type = "hbox",
            id = self.id,
            style = containerStyle,
            children = children,
        }
    end,
})

return NavigationMenu