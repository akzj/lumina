-- shadcn/badge — Badge component
local lumina = require("lumina")

local Badge = lumina.defineComponent({
    name = "ShadcnBadge",
    init = function(props)
        return {
            variant = props.variant or "default",
            label = props.label or "",
        }
    end,
    render = function(self)
        local styles = {
            default      = { bg = "#2563EB", fg = "#FFFFFF" },
            secondary    = { bg = "#1E293B", fg = "#E2E8F0" },
            destructive  = { bg = "#DC2626", fg = "#FFFFFF" },
            outline      = { bg = "",        fg = "#E2E8F0" },
            ghost        = { bg = "",        fg = "#94A3B8" },
        }
        local s = styles[self.variant] or styles.default
        return {
            type = "text",
            content = self.label,
            style = {
                background = s.bg,
                foreground = s.fg,
                bold = true,
            },
        }
    end
})

return Badge
