-- shadcn/collapsible — Simple collapsible section
local lumina = require("lumina")

local Collapsible = lumina.defineComponent({
    name = "ShadcnCollapsible",
    init = function(props)
        return {
            open = props.open or false,
            title = props.title or "",
        }
    end,
    render = function(self)
        local chevron = self.open and "▼" or "▶"
        local children = {
            {
                type = "hbox",
                children = {
                    { type = "text", content = chevron .. " ", style = { foreground = "#64748B" } },
                    { type = "text", content = self.title, style = { foreground = "#E2E8F0", bold = true } },
                },
            },
        }
        if self.open and self.props and self.props.children then
            children[#children + 1] = {
                type = "vbox",
                style = { padding = 1 },
                children = self.props.children,
            }
        end
        return {
            type = "vbox",
            children = children,
        }
    end
})

return Collapsible
