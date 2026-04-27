-- shadcn/accordion — Collapsible sections
local lumina = require("lumina")

local Accordion = lumina.defineComponent({
    name = "ShadcnAccordion",
    init = function(props)
        return {
            items = props.items or {},
            defaultOpen = props.defaultOpen or {},
        }
    end,
    render = function(self)
        local children = {}
        for i, item in ipairs(self.items) do
            -- Check if this item is in defaultOpen
            local isOpen = false
            for _, openIdx in ipairs(self.defaultOpen) do
                if openIdx == i then isOpen = true; break end
            end

            -- Header with chevron
            local chevron = isOpen and "▼" or "▶"
            children[#children + 1] = {
                type = "hbox",
                style = { padding = 0 },
                children = {
                    { type = "text", content = chevron .. " ", style = { foreground = "#64748B" } },
                    { type = "text", content = item.title or ("Item " .. i), style = { foreground = "#E2E8F0", bold = true } },
                },
            }

            -- Content (only if open)
            if isOpen and item.content then
                children[#children + 1] = {
                    type = "vbox",
                    style = { padding = 1 },
                    children = {
                        { type = "text", content = item.content, style = { foreground = "#94A3B8" } },
                    },
                }
            end

            -- Separator between items
            if i < #self.items then
                children[#children + 1] = {
                    type = "text",
                    content = "────────────────────────────────",
                    style = { foreground = "#1E293B" },
                }
            end
        end
        return {
            type = "vbox",
            children = children,
        }
    end
})

return Accordion
