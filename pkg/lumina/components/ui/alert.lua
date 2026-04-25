-- shadcn/alert — Alert component with variants
local lumina = require("lumina")

local Alert = lumina.defineComponent({
    name = "ShadcnAlert",
    init = function(props)
        return {
            variant = props.variant or "default",
            title = props.title or "",
            description = props.description or "",
            icon = props.icon or "",
        }
    end,
    render = function(self)
        local styles = {
            default     = { fg = "#E2E8F0", border_fg = "#334155", icon = "ℹ" },
            destructive = { fg = "#FCA5A5", border_fg = "#DC2626", icon = "⚠" },
        }
        local s = styles[self.variant] or styles.default
        local icon = self.icon ~= "" and self.icon or s.icon

        local children = {}
        if icon ~= "" then
            children[#children + 1] = { type = "text", content = icon .. " ", style = { bold = true } }
        end
        if self.title ~= "" then
            children[#children + 1] = { type = "text", content = self.title, style = { bold = true, foreground = s.fg } }
        end
        if self.description ~= "" then
            children[#children + 1] = { type = "text", content = self.description, style = { foreground = s.fg } }
        end

        return {
            type = "vbox",
            style = {
                border = "rounded",
                foreground = s.fg,
                padding = 1,
            },
            children = children,
        }
    end
})

return Alert
