-- examples/pagination_demo.lua — Lux Pagination demo
--
-- Demonstrates:
--   • require("lux").Pagination with pageCount, currentPage, onPageChange
--   • Arrow keys (h/l) for prev/next page
--   • Paginated list of items (120 items, 10 per page)
--
-- Usage: lumina examples/pagination_demo.lua
-- Quit: q

local lux = require("lux")
local Pagination = lux.Pagination

-- Generate 120 items
local allItems = {}
for i = 1, 120 do
    allItems[i] = "Item " .. tostring(i)
end

local itemsPerPage = 10
local totalPages = math.ceil(#allItems / itemsPerPage)

lumina.app {
    id = "pagination-demo",
    store = {
        currentPage = 1,
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
    },
    render = function()
        local t = lumina.getTheme()
        local currentPage = lumina.useStore("currentPage")

        -- Compute visible items for current page
        local startIdx = (currentPage - 1) * itemsPerPage + 1
        local endIdx = math.min(startIdx + itemsPerPage - 1, #allItems)

        -- Build children array manually to avoid table.unpack mid-argument issues
        local children = {}

        children[#children + 1] = lumina.createElement("text", {
            bold = true,
            foreground = t.text or "#CDD6F4",
            style = { height = 1, background = t.surface0 or "#313244" },
        }, "  Pagination demo (Lux)")

        children[#children + 1] = lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            style = { height = 1 },
        }, "  h/l arrows  |  q quit")

        children[#children + 1] = lumina.createElement("text", {
            foreground = t.primary or "#89B4FA",
            bold = true,
            style = { height = 1 },
        }, "  Page " .. tostring(currentPage) .. " of " .. tostring(totalPages))

        children[#children + 1] = lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            dim = true,
            style = { height = 1 },
        }, string.rep("-", 40))

        -- Item rows
        for i = startIdx, endIdx do
            children[#children + 1] = lumina.createElement("text", {
                key = "item-" .. tostring(i),
                foreground = t.text or "#CDD6F4",
                background = t.base or "#1E1E2E",
                style = { height = 1 },
            }, "  " .. allItems[i])
        end

        -- Spacer
        children[#children + 1] = lumina.createElement("vbox", {
            style = { flex = 1, background = t.base or "#1E1E2E" },
        })

        -- Pagination bar
        children[#children + 1] = Pagination {
            id = "pager",
            pageCount = totalPages,
            currentPage = currentPage,
            onPageChange = function(page)
                lumina.store.set("currentPage", page)
            end,
            pageRangeDisplayed = 3,
            marginPagesDisplayed = 1,
            autoFocus = true,
        }

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 60, height = 20 },
        }, table.unpack(children))
    end,
}
