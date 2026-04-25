-- shadcn/skeleton — Loading placeholder
local lumina = require("lumina")

local Skeleton = lumina.defineComponent({
    name = "ShadcnSkeleton",
    init = function(props)
        return {
            width = props.width or 20,
            height = props.height or 1,
        }
    end,
    render = function(self)
        local line = string.rep("░", self.width)
        local children = {}
        for i = 1, self.height do
            children[i] = { type = "text", content = line, style = { foreground = "#334155" } }
        end
        return {
            type = "vbox",
            children = children,
        }
    end
})

return Skeleton
