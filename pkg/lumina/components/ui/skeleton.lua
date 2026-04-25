-- shadcn/skeleton — Loading placeholder
local lumina = require("lumina")

local c = {
    muted = "#313244",
    border = "#45475A",
}

local Skeleton = lumina.defineComponent({
    name = "ShadcnSkeleton",

    init = function(props)
        return {
            width = props.width or 10,
            height = props.height or 1,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local style = {
            foreground = c.muted,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        local content = string.rep("█", self.width)

        return {
            type = "text",
            id = self.id,
            content = content,
            style = style,
        }
    end,
})

return Skeleton