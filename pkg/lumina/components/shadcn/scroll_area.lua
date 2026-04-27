-- shadcn/scroll_area — Custom scrollable area
local lumina = require("lumina")

local ScrollArea = lumina.defineComponent({
    name = "ShadcnScrollArea",
    init = function(props)
        return {
            height = props.height or 10,
            width = props.width or 40,
            scrollOffset = props.scrollOffset or 0,
            showScrollbar = props.showScrollbar ~= false,
        }
    end,
    render = function(self)
        local children = {}
        if self.props and self.props.children then
            for _, child in ipairs(self.props.children) do
                children[#children + 1] = child
            end
        end
        -- Scrollbar indicator
        if self.showScrollbar then
            children[#children + 1] = {
                type = "text",
                content = "▲ scroll ▼",
                style = { foreground = "#475569" },
            }
        end
        return {
            type = "vbox",
            style = {
                height = self.height,
                width = self.width,
                overflow = "scroll",
            },
            children = children,
        }
    end
})

return ScrollArea
