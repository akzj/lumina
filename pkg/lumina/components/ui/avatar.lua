-- shadcn/avatar — User avatar display
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    surface = "#313244",
}

local Avatar = lumina.defineComponent({
    name = "ShadcnAvatar",

    init = function(props)
        return {
            src = props.src or "",
            alt = props.alt or "",
            fallback = props.fallback or "",
            initials = props.initials or "",
            size = props.size or "default", -- sm, default, lg
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        -- Determine fallback content
        local content
        if self.fallback ~= "" then
            content = self.fallback
        elseif self.initials ~= "" then
            content = self.initials
        else
            content = "?"
        end

        -- Size determines box dimensions
        local size = self.size
        local display = "[" .. content .. "]"

        local style = {
            foreground = c.fg,
            background = c.surface,
            border = "rounded",
            borderColor = c.border,
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
            content = display,
            style = style,
        }
    end,
})

return Avatar