-- shadcn/breadcrumb — Breadcrumb navigation
local lumina = require("lumina")

local Breadcrumb = lumina.defineComponent({
    name = "ShadcnBreadcrumb",
    init = function(props)
        return {
            items = props.items or {},
            separator = props.separator or "/",
        }
    end,
    render = function(self)
        local children = {}
        for i, item in ipairs(self.items) do
            if i > 1 then
                children[#children + 1] = {
                    type = "text",
                    content = " " .. self.separator .. " ",
                    style = { foreground = "#475569" },
                }
            end
            local isLast = (i == #self.items)
            children[#children + 1] = {
                type = "text",
                content = item.label or item,
                style = {
                    foreground = isLast and "#E2E8F0" or "#3B82F6",
                    bold = isLast,
                    underline = not isLast,
                },
            }
        end
        return {
            type = "hbox",
            children = children,
        }
    end
})

return Breadcrumb
