-- shadcn/pagination — Pagination controls
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
}

local Pagination = lumina.defineComponent({
    name = "ShadcnPagination",

    init = function(props)
        return {
            page = props.page or 1,
            total = props.total or 10,
            perPage = props.perPage or 10,
            siblings = props.siblings or 1,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local total = self.total or 10
        local perPage = self.perPage or 10
        local totalPages = math.max(1, math.ceil(total / perPage))
        local page = math.max(1, math.min(self.page, totalPages))

        local children = {}

        -- Previous button
        local prevDisabled = (page <= 1)
        children[#children + 1] = {
            type = "text",
            content = prevDisabled and "[◀]" or "[◀]",
            style = { foreground = prevDisabled and c.muted or c.primary },
        }

        children[#children + 1] = {
            type = "text",
            content = " ",
            style = {},
        }

        -- Page number
        children[#children + 1] = {
            type = "text",
            content = " " .. page .. " / " .. totalPages .. " ",
            style = { foreground = c.fg, bold = true },
        }

        children[#children + 1] = {
            type = "text",
            content = " ",
            style = {},
        }

        -- Next button
        local nextDisabled = (page >= totalPages)
        children[#children + 1] = {
            type = "text",
            content = nextDisabled and "[▶]" or "[▶]",
            style = { foreground = nextDisabled and c.muted or c.primary },
        }

        local containerStyle = { align = "center" }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do containerStyle[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do containerStyle[k] = v end
        end

        return {
            type = "hbox",
            id = self.id,
            style = containerStyle,
            children = children,
        }
    end,
})

return Pagination