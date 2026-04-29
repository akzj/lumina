-- examples/data_grid_paginated.lua — DataGrid + Pagination combo
--
-- Demonstrates: DataGrid with sorting, pagination, default renderCell
-- Usage: lumina examples/data_grid_paginated.lua
-- Quit: q

local lux = require("lux")
local DataGrid = lux.DataGrid
local Pagination = lux.Pagination

-- Generate sample data (100 items)
local allRows = {}
for i = 1, 100 do
	allRows[i] = {
		id = i,
		name = "Item-" .. string.format("%03d", i),
		value = math.floor(i * 7.3),
		status = (i % 3 == 0) and "active" or ((i % 3 == 1) and "pending" or "done"),
	}
end

local PAGE_SIZE = 10

lumina.app {
	id = "paginated-grid",
	store = {
		page = 1,
		selectedIdx = 1,
		sortColumnId = "",
		sortDirection = "",
	},
	keys = {
		["q"] = function() lumina.quit() end,
		["ctrl+c"] = function() lumina.quit() end,
	},
	render = function()
		local t = lumina.getTheme()
		local page = lumina.useStore("page")
		local selectedIdx = lumina.useStore("selectedIdx")
		local sortColumnId = lumina.useStore("sortColumnId")
		local sortDirection = lumina.useStore("sortDirection")

		local sort = nil
		if sortColumnId ~= "" then
			sort = { columnId = sortColumnId, direction = sortDirection }
		end

		-- Sort data
		local sorted = {}
		for i, r in ipairs(allRows) do sorted[i] = r end
		if sort then
			table.sort(sorted, function(a, b)
				local va, vb = a[sort.columnId], b[sort.columnId]
				if va == nil then va = "" end
				if vb == nil then vb = "" end
				if type(va) == "number" and type(vb) == "number" then
					if sort.direction == "desc" then return va > vb end
					return va < vb
				end
				va, vb = tostring(va), tostring(vb)
				if sort.direction == "desc" then return va > vb end
				return va < vb
			end)
		end

		-- Paginate
		local pageCount = math.ceil(#sorted / PAGE_SIZE)
		local startIdx = (page - 1) * PAGE_SIZE + 1
		local endIdx = math.min(page * PAGE_SIZE, #sorted)
		local pageRows = {}
		for i = startIdx, endIdx do
			pageRows[#pageRows + 1] = sorted[i]
		end

		local columns = {
			{ id = "id",     header = "#",      width = 6,  key = "id",     sortable = true, align = "right" },
			{ id = "name",   header = "Name",   width = 16, key = "name",   sortable = true },
			{ id = "value",  header = "Value",  width = 10, key = "value",  sortable = true, align = "right" },
			{ id = "status", header = "Status", width = 10, key = "status", sortable = true },
		}

		return lumina.createElement("vbox", {
			id = "root",
			style = { width = 60, height = 24, background = t.base or "#1E1E2E" },
		},
			lumina.createElement("text", {
				bold = true,
				foreground = t.primary or "#89B4FA",
				style = { height = 1 },
			}, "  Paginated DataGrid (100 items, sort + page)"),
			lumina.createElement("text", {
				foreground = t.muted or "#6C7086",
				style = { height = 1 },
			}, "  j/k=nav  Enter=select  click header=sort  q=quit"),
			DataGrid {
				id = "grid",
				width = 56,
				height = 14,
				columns = columns,
				rows = pageRows,
				selectedIndex = selectedIdx,
				sort = sort,
				onChangeIndex = function(i)
					lumina.store.set("selectedIdx", i)
				end,
				onSortChange = function(meta)
					lumina.store.set("sortColumnId", meta.columnId)
					lumina.store.set("sortDirection", meta.direction)
					lumina.store.set("page", 1)
					lumina.store.set("selectedIdx", 1)
				end,
				onActivate = function(i, row)
					-- no-op for this demo
				end,
				autoFocus = true,
			},
			Pagination {
				id = "pager",
				pageCount = pageCount,
				currentPage = page,
				onPageChange = function(p)
					lumina.store.set("page", p)
					lumina.store.set("selectedIdx", 1)
				end,
				pageRangeDisplayed = 3,
				marginPagesDisplayed = 1,
			},
			lumina.createElement("text", {
				foreground = t.muted or "#6C7086",
				style = { height = 1 },
			}, "  Page " .. tostring(page) .. "/" .. tostring(pageCount) .. "  |  " .. tostring(#sorted) .. " total rows")
		)
	end,
}
