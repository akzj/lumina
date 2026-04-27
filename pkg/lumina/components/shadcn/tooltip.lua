-- shadcn/tooltip — Tooltip (shows on focus in terminal)
local lumina = require("lumina")

local Tooltip = lumina.defineComponent({
    name = "ShadcnTooltip",
    init = function(props)
        return {
            content = props.content or "",
            visible = props.visible or false,
            side = props.side or "top",
        }
    end,
    render = function(self)
        local children = {}
        -- Trigger (children from props)
        if self.props and self.props.children then
            for _, child in ipairs(self.props.children) do
                children[#children + 1] = child
            end
        end
        -- Tooltip popup
        if self.visible and self.content ~= "" then
            children[#children + 1] = {
                type = "text",
                content = " " .. self.content .. " ",
                style = {
                    background = "#1E293B",
                    foreground = "#E2E8F0",
                    bold = true,
                },
            }
        end
        return {
            type = "vbox",
            children = children,
        }
    end
})

return Tooltip
