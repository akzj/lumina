-- shadcn/separator — Horizontal/vertical divider
local lumina = require("lumina")

local c = {
    border = "#45475A",
}

local Separator = lumina.defineComponent({
    name = "ShadcnSeparator",

    init = function(props)
        return {
            orientation = props.orientation or "horizontal", -- horizontal, vertical
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local content = self.orientation == "horizontal" and "────────────────────────────────" or "│"

        local style = {
            foreground = c.border,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        return {
            type = "text",
            id = self.id,
            content = content,
            style = style,
        }
    end,
})

return Separator