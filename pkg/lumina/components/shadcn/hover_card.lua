-- shadcn/hover_card — Rich hover card (larger tooltip)
local lumina = require("lumina")

local HoverCard = lumina.defineComponent({
    name = "ShadcnHoverCard",
    init = function(props)
        return {
            open = props.open or false,
            trigger = props.trigger or "",
            width = props.width or 40,
        }
    end,
    render = function(self)
        local children = {}
        -- Trigger
        children[#children + 1] = {
            type = "text",
            content = self.trigger,
            style = { foreground = "#3B82F6", underline = true },
        }
        -- Card content
        if self.open then
            local cardChildren = {}
            if self.props and self.props.children then
                for _, child in ipairs(self.props.children) do
                    cardChildren[#cardChildren + 1] = child
                end
            end
            children[#children + 1] = {
                type = "vbox",
                style = {
                    border = "rounded",
                    background = "#0F172A",
                    foreground = "#E2E8F0",
                    padding = 1,
                    width = self.width,
                },
                children = cardChildren,
            }
        end
        return {
            type = "vbox",
            children = children,
        }
    end
})

return HoverCard
