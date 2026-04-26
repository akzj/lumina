-- ============================================================================
-- Lumina Example: Data Table
-- ============================================================================
-- Demonstrates: Table component, sorting, filtering, pagination
-- Run: lumina examples/data-table/main.lua
-- ============================================================================
local lumina = require("lumina")
local shadcn = require("shadcn")

lumina.setTheme("catppuccin-mocha")

local c = {
    bg = "#1E1E2E", fg = "#CDD6F4", accent = "#89B4FA",
    success = "#A6E3A1", error = "#F38BA8", muted = "#6C7086",
    surface = "#313244", border = "#45475A", warning = "#F9E2AF",
}

-- Sample data
local allData = {
    { id = 1, name = "Alice Chen", email = "alice@example.com", role = "Developer", status = "Active", score = 92 },
    { id = 2, name = "Bob Wang", email = "bob@example.com", role = "Designer", status = "Active", score = 78 },
    { id = 3, name = "Carol Li", email = "carol@example.com", role = "Manager", status = "Inactive", score = 85 },
    { id = 4, name = "David Kim", email = "david@example.com", role = "Developer", status = "Active", score = 91 },
    { id = 5, name = "Eva Müller", email = "eva@example.com", role = "Product", status = "Active", score = 88 },
    { id = 6, name = "Frank Lee", email = "frank@example.com", role = "Designer", status = "Inactive", score = 72 },
    { id = 7, name = "Grace Patel", email = "grace@example.com", role = "Developer", status = "Active", score = 95 },
    { id = 8, name = "Henry Wong", email = "henry@example.com", role = "Manager", status = "Active", score = 81 },
    { id = 9, name = "Ivy Santos", email = "ivy@example.com", role = "Product", status = "Active", score = 79 },
    { id = 10, name = "Jack Brown", email = "jack@example.com", role = "Developer", status = "Inactive", score = 67 },
    { id = 11, name = "Kate Liu", email = "kate@example.com", role = "Designer", status = "Active", score = 83 },
    { id = 12, name = "Leo Zhang", email = "leo@example.com", role = "Developer", status = "Active", score = 89 },
}

local store = lumina.createStore({
    state = {
        sortBy = "name",
        sortDir = "asc",
        filter = "",
        page = 1,
        perPage = 5,
    },
})

local function getFilteredData(state)
    local data = allData
    local filter = state.filter and state.filter:lower() or ""

    -- Filter
    if filter ~= "" then
        local filtered = {}
        for _, row in ipairs(data) do
            if row.name:lower():find(filter, 1, true) or
               row.email:lower():find(filter, 1, true) or
               row.role:lower():find(filter, 1, true) or
               row.status:lower():find(filter, 1, true) then
                table.insert(filtered, row)
            end
        end
        data = filtered
    end

    -- Sort
    local sortBy = state.sortBy or "name"
    local sortDir = state.sortDir or "asc"
    local function cmpStrings(va, vb)
        return string.lower(va) < string.lower(vb)
    end
    table.sort(data, function(a, b)
        local va, vb = a[sortBy], b[sortBy]
        if va == vb then return false end
        if type(va) == "string" then
            if sortDir == "asc" then
                return cmpStrings(va, vb)
            else
                return cmpStrings(vb, va)
            end
        else
            if sortDir == "asc" then
                return tonumber(va) < tonumber(vb)
            else
                return tonumber(vb) < tonumber(va)
            end
        end
    end)

    return data
end

local function getPaginatedData(data, state)
    local page = state.page or 1
    local perPage = state.perPage or 5
    local start = (page - 1) * perPage + 1
    local result = {}
    for i = start, math.min(start + perPage - 1, #data) do
        table.insert(result, data[i])
    end
    return result
end

local function StatusBadge(status)
    if status == "Active" then
        return {
            type = "text",
            content = "● Active",
            style = { foreground = c.success },
        }
    else
        return {
            type = "text",
            content = "○ Inactive",
            style = { foreground = c.muted },
        }
    end
end

local function ScoreBar(score)
    local width = 10
    local filled = math.floor(score / 10)
    local bar = string.rep("█", filled) .. string.rep("░", width - filled)
    local color = score >= 90 and c.success or score >= 70 and c.accent or c.warning
    return {
        type = "text",
        content = bar .. " " .. score .. "%",
        style = { foreground = color },
    }
end

local App = lumina.defineComponent({
    name = "DataTable",
    render = function(self)
        local state = lumina.useStore(store)
        local data = getFilteredData(state)
        local pageData = getPaginatedData(data, state)
        local totalPages = math.max(1, math.ceil(#data / (state.perPage or 5)))
        local page = math.max(1, math.min(state.page or 1, totalPages))

        local colWidths = { 12, 22, 12, 10, 12 }

        -- Header
        local function SortHeader(label, key)
            local isActive = state.sortBy == key
            local dir = isActive and (state.sortDir == "asc" and "↑" or "↓") or ""
            return {
                type = "text",
                content = string.format("%-" .. (colWidths[1] or 12) .. "s", " " .. label .. dir),
                style = { foreground = isActive and c.accent or c.muted, bold = isActive },
            }
        end

        -- Table row
        local function TableRow(row, bg)
            return {
                type = "hbox",
                style = { background = bg and c.surface or "" },
                children = {
                    { type = "text", content = string.format("%-" .. colWidths[1] .. "s", " " .. row.name), style = { foreground = c.fg } },
                    { type = "text", content = string.format("%-" .. colWidths[2] .. "s", " " .. row.email), style = { foreground = c.muted } },
                    { type = "text", content = string.format("%-" .. colWidths[3] .. "s", " " .. row.role), style = { foreground = c.accent } },
                    StatusBadge(row.status),
                    ScoreBar(row.score),
                },
            }
        end

        local sep = "  " .. string.rep("─", colWidths[1] + colWidths[2] + colWidths[3] + colWidths[4] + colWidths[5] + 10)

        return {
            type = "vbox",
            style = { background = c.bg },
            children = {
                { type = "text", content = " 📊 Data Table ", style = { foreground = c.accent, bold = true, background = "#181825" } },
                { type = "text", content = " Sort: [↑/↓]  Filter: [f]  Page: [< / >]  Quit: [q] ", style = { foreground = c.muted } },
                { type = "text", content = "" },

                -- Stats bar
                {
                    type = "hbox",
                    children = {
                        { type = "text", content = " Showing " .. #pageData .. "/" .. #data .. " rows", style = { foreground = c.fg } },
                        { type = "spacer" },
                        { type = "text", content = " Page " .. page .. "/" .. totalPages, style = { foreground = c.muted } },
                    },
                },
                { type = "text", content = "" },

                -- Filter indicator
                state.filter and state.filter ~= "" and {
                    type = "text",
                    content = " Filter: \"" .. state.filter .. "\"",
                    style = { foreground = c.warning, bold = true },
                } or nil,

                -- Table
                {
                    type = "vbox",
                    style = { border = "rounded", borderColor = c.border, padding = 1 },
                    children = (function()
                        local children = {}

                        -- Header row
                        children[#children + 1] = {
                            type = "hbox",
                            style = { foreground = c.muted },
                            children = {
                                { type = "text", content = " Name        ", style = { bold = true } },
                                { type = "text", content = " Email                  ", style = { bold = true } },
                                { type = "text", content = " Role        ", style = { bold = true } },
                                { type = "text", content = " Status    ", style = { bold = true } },
                                { type = "text", content = " Score       ", style = { bold = true } },
                            },
                        }
                        children[#children + 1] = {
                            type = "text",
                            content = sep,
                            style = { foreground = c.border },
                        }

                        -- Data rows
                        for i, row in ipairs(pageData) do
                            children[#children + 1] = TableRow(row, i % 2 == 0)
                        end

                        -- Empty state
                        if #pageData == 0 then
                            children[#children + 1] = {
                                type = "text",
                                content = "  No results found",
                                style = { foreground = c.muted },
                            }
                        end

                        return children
                    end)(),
                },

                { type = "text", content = "" },

                -- Pagination controls
                {
                    type = "hbox",
                    children = {
                        { type = "text", content = page > 1 and "[◀ Prev]" or "[◀ Prev]", style = { foreground = page > 1 and c.accent or c.muted } },
                        { type = "text", content = "  " },
                        -- Page indicators
                        (function()
                            local indicators = {}
                            for i = 1, totalPages do
                                local content = " " .. i .. " "
                                if i == page then
                                    content = "[" .. i .. "]"
                                end
                                indicators[#indicators + 1] = {
                                    type = "text",
                                    content = content,
                                    style = { foreground = i == page and c.accent or c.muted, bold = i == page },
                                }
                            end
                            return { type = "hbox", children = indicators }
                        end)(),
                        { type = "text", content = "  " },
                        { type = "text", content = page < totalPages and "[Next ▶]" or "[Next ▶]", style = { foreground = page < totalPages and c.accent or c.muted } },
                    },
                },

                { type = "text", content = "" },
                { type = "text", content = " [↑] name  [↓] name  [f] filter  [c] clear filter  [q] quit", style = { foreground = c.muted, dim = true } },
            },
        }
    end,
})

lumina.onKey("j", function()
    local state = store.getState()
    if state.page < math.ceil(#getFilteredData(state) / (state.perPage or 5)) then
        store.dispatch("setState", { page = (state.page or 1) + 1 })
    end
end)

lumina.onKey("k", function()
    local state = store.getState()
    if (state.page or 1) > 1 then
        store.dispatch("setState", { page = (state.page or 1) - 1 })
    end
end)

lumina.onKey("n", function()
    local state = store.getState()
    store.dispatch("setState", { sortBy = "name", sortDir = state.sortBy == "name" and (state.sortDir == "asc" and "desc" or "asc") or "asc", page = 1 })
end)

lumina.onKey("s", function()
    local state = store.getState()
    store.dispatch("setState", { sortBy = "score", sortDir = state.sortBy == "score" and (state.sortDir == "desc" and "asc" or "desc") or "desc", page = 1 })
end)

lumina.onKey("f", function()
    store.dispatch("setState", { filter = "Developer", page = 1 })
end)

lumina.onKey("c", function()
    store.dispatch("setState", { filter = "", page = 1 })
end)

lumina.onKey("q", function() lumina.quit() end)

lumina.mount(App)
lumina.run()
