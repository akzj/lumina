-- shadcn/button — Terminal-adapted Button component
local lumina = require("lumina")

local Button = lumina.defineComponent({
    name = "ShadcnButton",
    init = function(props)
        return {
            variant = props.variant or "default",
            size = props.size or "default",
            disabled = props.disabled or false,
            label = props.label or "Button",
        }
    end,
    render = function(self)
        local styles = {
            default      = { bg = "#2563EB", fg = "#FFFFFF" },
            outline      = { bg = "",        fg = "#E2E8F0" },
            secondary    = { bg = "#1E293B", fg = "#E2E8F0" },
            ghost        = { bg = "",        fg = "#E2E8F0" },
            destructive  = { bg = "#DC2626", fg = "#FFFFFF" },
            link         = { bg = "",        fg = "#3B82F6" },
        }
        local s = styles[self.variant] or styles.default

        local sizeMap = {
            default = { paddingLeft = 1, paddingRight = 1, height = 3 },
            sm      = { padding = 0, height = 1 },
            lg      = { padding = 2, height = 7 },
            icon    = { padding = 0, width = 3, height = 3 },
            xs      = { padding = 0, height = 1 },
        }
        local sz = sizeMap[self.size] or sizeMap.default

        return {
            type = "hbox",
            style = {
                background = s.bg,
                foreground = self.disabled and "#666666" or s.fg,
                border = "rounded",
                padding = sz.padding,
                paddingLeft = sz.paddingLeft,
                paddingRight = sz.paddingRight,
                height = sz.height,
                width = sz.width,
                justify = "center",
                align = "center",
                dim = self.disabled,
            },
            children = self.props and self.props.children or {
                { type = "text", content = self.label }
            },
        }
    end
})

return Button
