-- shadcn/toggle — Toggle button (pressed/unpressed)
local lumina = require("lumina")

local Toggle = lumina.defineComponent({
    name = "ShadcnToggle",
    init = function(props)
        return {
            pressed = props.pressed or false,
            variant = props.variant or "default",
            size = props.size or "default",
            label = props.label or "",
            disabled = props.disabled or false,
        }
    end,
    render = function(self)
        local fg, bg
        if self.disabled then
            fg = "#475569"
            bg = ""
        elseif self.pressed then
            if self.variant == "outline" then
                fg = "#E2E8F0"
                bg = "#1E293B"
            else
                fg = "#E2E8F0"
                bg = "#334155"
            end
        else
            fg = "#94A3B8"
            bg = ""
        end

        local sizeMap = {
            default = { padding = 1 },
            sm = { padding = 0 },
            lg = { padding = 2 },
        }
        local sz = sizeMap[self.size] or sizeMap.default

        return {
            type = "hbox",
            style = {
                foreground = fg,
                background = bg,
                border = self.variant == "outline" and "rounded" or "",
                padding = sz.padding,
                justify = "center",
                align = "center",
            },
            children = {
                { type = "text", content = self.label },
            },
        }
    end
})

return Toggle
