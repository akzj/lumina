-- shadcn/alert — Alert component with variants
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    destructive = "#F38BA8",
    destructive_fg = "#1E1E2E",
    warning = "#F9E2AF",
    warning_fg = "#1E1E2E",
    success = "#A6E3A1",
    success_fg = "#1E1E2E",
    info = "#89B4FA",
    info_fg = "#1E1E2E",
    border = "#45475A",
}

local Alert = lumina.defineComponent({
    name = "ShadcnAlert",

    init = function(props)
        return {
            variant = props.variant or "default", -- default, destructive, warning, success
            title = props.title or "",
            description = props.description or "",
            icon = props.icon,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local variant = self.variant

        local styles = {
            default = { icon = "ℹ", iconFg = c.info, fg = c.fg, borderFg = c.border },
            destructive = { icon = "✕", iconFg = c.destructive, fg = c.destructive, borderFg = c.destructive },
            warning = { icon = "!", iconFg = c.warning, fg = c.warning, borderFg = c.warning },
            success = { icon = "✓", iconFg = c.success, fg = c.success, borderFg = c.success },
        }
        local s = styles[variant] or styles.default
        local icon = self.icon or s.icon

        local children = {}
        if icon ~= "" then
            children[#children + 1] = {
                type = "text",
                content = " " .. icon .. " ",
                style = { foreground = s.iconFg, bold = true },
            }
        end
        if self.title ~= "" then
            children[#children + 1] = {
                type = "text",
                content = " " .. self.title,
                style = { foreground = s.fg, bold = true },
            }
        end
        if self.description ~= "" then
            local prefix = (self.title ~= "" or icon ~= "") and " " or ""
            children[#children + 1] = {
                type = "text",
                content = prefix .. self.description,
                style = { foreground = s.fg },
            }
        end

        local style = {
            border = "rounded",
            borderColor = s.borderFg,
            foreground = s.fg,
            paddingLeft = 1,
            paddingRight = 1,
            paddingTop = 0,
            paddingBottom = 0,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        return {
            type = "hbox",
            id = self.id,
            style = style,
            children = children,
        }
    end,
})

return Alert
