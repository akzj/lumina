-- shadcn/tooltip — Hover tooltip
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    border = "#45475A",
    bg = "#181825",
}

local Tooltip = lumina.defineComponent({
    name = "ShadcnTooltip",

    init = function(props)
        return {
            content = props.content or props.label or "",
            side = props.side or "top", -- top, bottom, left, right
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        if self.content == "" then
            return { type = "empty" }
        end

        local style = {
            border = "rounded",
            borderColor = c.border,
            background = c.bg,
            foreground = c.fg,
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
            content = " " .. self.content .. " ",
            style = style,
        }
    end,
})

return Tooltip