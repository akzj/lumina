-- shadcn/kbd — Keyboard key indicator
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    border = "#45475A",
    bg = "#181825",
}

local Kbd = lumina.defineComponent({
    name = "ShadcnKbd",

    init = function(props)
        return {
            children = props.children or props.key or "",
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local content = self.children
        if type(content) ~= "string" then
            content = tostring(content)
        end

        local style = {
            foreground = c.fg,
            background = c.bg,
            border = "single",
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

        return {
            type = "text",
            id = self.id,
            content = " " .. content .. " ",
            style = style,
        }
    end,
})

return Kbd