-- shadcn/spinner — Loading spinner indicator
local lumina = require("lumina")

local c = {
    primary = "#89B4FA",
}

local Spinner = lumina.defineComponent({
    name = "ShadcnSpinner",

    init = function(props)
        return {
            size = props.size or "default", -- sm, default, lg
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local frames = { "⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏" }

        -- Use a fixed frame for static display (animation requires JS)
        local content = "◐"

        local style = {
            foreground = c.primary,
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

return Spinner