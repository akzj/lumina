-- shadcn/badge — Status/count badge
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    destructive = "#F38BA8",
    success = "#A6E3A1",
    warning = "#F9E2AF",
    border = "#45475A",
}

local Badge = lumina.defineComponent({
    name = "ShadcnBadge",

    init = function(props)
        return {
            variant = props.variant or "default", -- default, secondary, destructive, outline
            children = props.children or props.label or "",
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local variant = self.variant

        local styles = {
            default = { fg = c.fg, bg = "#313244" },
            secondary = { fg = c.fg, bg = "#45475A" },
            destructive = { fg = c.destructive, bg = "" },
            outline = { fg = c.fg, bg = "" },
            success = { fg = c.success, bg = "" },
            warning = { fg = c.warning, bg = "" },
        }
        local s = styles[variant] or styles.default

        local style = {
            foreground = s.fg,
            background = s.bg,
            border = "rounded",
            borderColor = c.border,
            paddingLeft = 1,
            paddingRight = 1,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        local content = self.children
        if type(content) ~= "string" then
            content = tostring(content)
        end

        return {
            type = "text",
            id = self.id,
            content = " " .. content .. " ",
            style = style,
        }
    end,
})

return Badge