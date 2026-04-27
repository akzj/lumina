-- shadcn/pagination — Page navigation
local lumina = require("lumina")

local Pagination = lumina.defineComponent({
    name = "ShadcnPagination",
    init = function(props)
        return {
            currentPage = props.currentPage or 1,
            totalPages = props.totalPages or 1,
            maxVisible = props.maxVisible or 5,
        }
    end,
    render = function(self)
        local children = {}
        local cur = self.currentPage
        local total = self.totalPages

        -- Previous button
        children[#children + 1] = {
            type = "text",
            content = cur > 1 and "← " or "  ",
            style = { foreground = cur > 1 and "#3B82F6" or "#475569" },
        }

        -- Page numbers
        local startPage = math.max(1, cur - math.floor(self.maxVisible / 2))
        local endPage = math.min(total, startPage + self.maxVisible - 1)
        startPage = math.max(1, endPage - self.maxVisible + 1)

        if startPage > 1 then
            children[#children + 1] = { type = "text", content = "1 … ", style = { foreground = "#64748B" } }
        end

        for p = startPage, endPage do
            children[#children + 1] = {
                type = "text",
                content = " " .. tostring(p) .. " ",
                style = {
                    foreground = p == cur and "#E2E8F0" or "#64748B",
                    background = p == cur and "#2563EB" or "",
                    bold = p == cur,
                },
            }
        end

        if endPage < total then
            children[#children + 1] = { type = "text", content = " … " .. tostring(total), style = { foreground = "#64748B" } }
        end

        -- Next button
        children[#children + 1] = {
            type = "text",
            content = cur < total and " →" or "  ",
            style = { foreground = cur < total and "#3B82F6" or "#475569" },
        }

        return {
            type = "hbox",
            style = { justify = "center" },
            children = children,
        }
    end
})

return Pagination
