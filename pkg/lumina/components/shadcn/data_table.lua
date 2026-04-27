-- data_table.lua — Enhanced table with sorting, filtering, pagination
local lumina = require("lumina")

local DataTable = lumina.defineComponent({
    name = "ShadcnDataTable",
    render = function(self)
        local columns = self.props.columns or {}
        local data = self.props.data or {}
        local pageSize = self.props.pageSize or 10
        local currentPage = self.props.page or 1
        local sortColumn = self.props.sortColumn or ""
        local sortDirection = self.props.sortDirection or "asc"
        local filter = self.props.filter or ""
        local onSort = self.props.onSort
        local onPageChange = self.props.onPageChange

        -- Filter data
        local filtered = data
        if filter ~= "" then
            filtered = {}
            for _, row in ipairs(data) do
                for _, col in ipairs(columns) do
                    local val = tostring(row[col.key] or "")
                    if val:lower():find(filter:lower(), 1, true) then
                        filtered[#filtered + 1] = row
                        break
                    end
                end
            end
        end

        -- Sort data
        if sortColumn ~= "" then
            table.sort(filtered, function(a, b)
                local va = a[sortColumn] or ""
                local vb = b[sortColumn] or ""
                if type(va) == "number" and type(vb) == "number" then
                    return sortDirection == "asc" and va < vb or va > vb
                end
                va, vb = tostring(va), tostring(vb)
                return sortDirection == "asc" and va < vb or va > vb
            end)
        end

        -- Paginate
        local totalPages = math.ceil(#filtered / pageSize)
        if totalPages < 1 then totalPages = 1 end
        local startIdx = (currentPage - 1) * pageSize + 1
        local endIdx = math.min(startIdx + pageSize - 1, #filtered)

        -- Build header row
        local headerCells = {}
        for _, col in ipairs(columns) do
            local arrow = ""
            if col.key == sortColumn then
                arrow = sortDirection == "asc" and " ↑" or " ↓"
            end
            headerCells[#headerCells + 1] = {
                type = "text",
                content = (col.label or col.key) .. arrow,
                style = {
                    foreground = "#CDD6F4",
                    bold = true,
                    width = col.width or 15,
                },
            }
        end

        local children = {
            { type = "hbox", style = { background = "#313244" }, children = headerCells },
            { type = "text", content = string.rep("─", 60), style = { foreground = "#45475A" } },
        }

        -- Build data rows
        for i = startIdx, endIdx do
            local row = filtered[i]
            if row then
                local cells = {}
                for _, col in ipairs(columns) do
                    cells[#cells + 1] = {
                        type = "text",
                        content = tostring(row[col.key] or ""),
                        style = {
                            foreground = "#CDD6F4",
                            width = col.width or 15,
                        },
                    }
                end
                children[#children + 1] = {
                    type = "hbox",
                    style = { background = (i % 2 == 0) and "#1E1E2E" or "" },
                    children = cells,
                }
            end
        end

        -- Footer with pagination info
        children[#children + 1] = { type = "text", content = string.rep("─", 60), style = { foreground = "#45475A" } }
        children[#children + 1] = {
            type = "hbox",
            children = {
                { type = "text",
                  content = string.format("Showing %d-%d of %d | Page %d/%d",
                    startIdx, endIdx, #filtered, currentPage, totalPages),
                  style = { foreground = "#6C7086" } },
            },
        }

        return {
            type = "vbox",
            style = {
                border = self.props.border or "rounded",
                width = self.props.width,
                background = self.props.background or "#1E1E2E",
            },
            children = children,
        }
    end,
})

return DataTable
