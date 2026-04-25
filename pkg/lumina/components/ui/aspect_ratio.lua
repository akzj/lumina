-- aspect_ratio.lua — Maintain aspect ratio container
local lumina = require("lumina")

local AspectRatio = lumina.defineComponent({
    name = "ShadcnAspectRatio",
    render = function(self)
        local ratio = self.props.ratio or 16/9
        local width = self.props.width or 40
        local height = math.floor(width / ratio)
        return {
            type = "vbox",
            style = {
                width = width,
                height = height,
                border = self.props.border or "",
                background = self.props.background or "",
            },
            children = self.props.children or {},
        }
    end,
})

return AspectRatio
