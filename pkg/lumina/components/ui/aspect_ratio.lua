-- shadcn/aspect_ratio — Aspect ratio container
local lumina = require("lumina")

local AspectRatio = lumina.defineComponent({
    name = "ShadcnAspectRatio",

    init = function(props)
        return {
            ratio = props.ratio or 16 / 9,
            id = props.id,
            className = props.className,
            style = props.style,
            children = props.children or {},
        }
    end,

    render = function(self)
        -- In terminal, aspect ratio is informational — pass through children
        local contentChildren = self.children
        local children = {}
        if type(contentChildren) == "table" then
            if contentChildren.type then
                children[1] = contentChildren
            else
                for _, child in ipairs(contentChildren) do
                    children[#children + 1] = child
                end
            end
        end

        return {
            type = "vbox",
            id = self.id,
            style = self.style or {},
            children = children,
        }
    end,
})

return AspectRatio