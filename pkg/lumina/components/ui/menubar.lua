-- shadcn/menubar — Menu bar (File, Edit, etc.)
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
    surface = "#313244",
}

local MenuBar = lumina.defineComponent({
    name = "ShadcnMenuBar",

    init = function(props)
        return {
            menus = props.menus or {},
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local children = {}

        for i, menu in ipairs(self.menus) do
            local label = type(menu) == "table" and (menu.label or menu.title) or tostring(menu)

            children[#children + 1] = {
                type = "hbox",
                id = self.id and (self.id .. "-menu-" .. i) or nil,
                style = {
                    foreground = c.fg,
                    paddingLeft = 1,
                    paddingRight = 1,
                },
                children = {
                    { type = "text", content = " " .. label .. " " },
                    { type = "text", content = "▾", style = { foreground = c.muted } },
                },
            }

            if i < #self.menus then
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

return MenuBar