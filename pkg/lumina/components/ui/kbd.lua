-- shadcn/kbd — Keyboard shortcut display
local lumina = require("lumina")

local Kbd = lumina.defineComponent({
    name = "ShadcnKbd",
    init = function(props)
        return {
            keys = props.keys or props.key or "",
        }
    end,
    render = function(self)
        return {
            type = "text",
            content = "⌨ " .. self.keys,
            style = {
                foreground = "#94A3B8",
                background = "#1E293B",
                bold = true,
            },
        }
    end
})

return Kbd
