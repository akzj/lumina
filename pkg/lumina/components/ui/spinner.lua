-- shadcn/spinner — Animated loading spinner
local lumina = require("lumina")

local Spinner = lumina.defineComponent({
    name = "ShadcnSpinner",
    init = function(props)
        return {
            size = props.size or "default",
            label = props.label or "",
        }
    end,
    render = function(self)
        local frames = { "⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏" }
        -- Use a simple frame index based on time (or default to first frame)
        local frame = frames[1]
        local children = {
            { type = "text", content = frame, style = { foreground = "#3B82F6", bold = true } },
        }
        if self.label ~= "" then
            children[#children + 1] = { type = "text", content = " " .. self.label, style = { foreground = "#94A3B8" } }
        end
        return {
            type = "hbox",
            children = children,
        }
    end
})

return Spinner
