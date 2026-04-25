-- shadcn/breadcrumb — Breadcrumb navigation
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    separator = "#45475A",
}

local Breadcrumb = lumina.defineComponent({
    name = "ShadcnBreadcrumb",

    init = function(props)
        return {
            items = props.items or {},
            separator = props.separator or "/",
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local children = {}
        local sep = self.separator or "/"

        for i, item in ipairs(self.items) do
            local label = type(item) == "table" and (item.label or item.href or ("Item " .. i)) or tostring(item)
            local isLink = type(item) == "table" and item.href ~= nil

            -- Text content
            children[#children + 1] = {
                type = "text",
                content = label,
                style = {
                    foreground = isLink and c.primary or c.fg,
                    dim = not isLink,
                },
            }

            -- Separator (not after last item)
            if i < #self.items then
                children[#children + 1] = {
                    type = "text",
                    content = " " .. sep .. " ",
                    style = { foreground = c.separator },
                }
            end
        end

        return {
            type = "hbox",
            id = self.id,
            style = { align = "center" },
            children = children,
        }
    end,
})

return Breadcrumb