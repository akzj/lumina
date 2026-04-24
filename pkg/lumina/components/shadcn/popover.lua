-- shadcn/popover — Popup content anchored to trigger
local lumina = require("lumina")

local Popover = lumina.defineComponent({
    name = "ShadcnPopover",
    init = function(props)
        return {
            open = props.open or false,
            trigger = props.trigger or "Open",
            align = props.align or "start",
            width = props.width or 30,
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
        -- Content (when open)
        if self.open then
            local popoverChildren = {}
            if self.props and self.props.children then
                for _, child in ipairs(self.props.children) do
                    popoverChildren[#popoverChildren + 1] = child
                end
            end
            children[#children + 1] = {
                type = "vbox",
                style = {
                    border = "rounded",
                    background = "#1E293B",
                    foreground = "#E2E8F0",
                    padding = 1,
                    width = self.width,
                },
                children = popoverChildren,
            }
        end
        return {
            type = "vbox",
            children = children,
        }
    end
})

return Popover
