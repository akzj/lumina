-- examples/api_client_demo.lua — Real-world API client pattern
--
-- Demonstrates: async data loading, loading states, DataGrid + Pagination,
-- error handling with Alert, refresh capability
--
-- Usage: lumina examples/api_client_demo.lua
-- Quit: q | Refresh: r

local lux = require("lux")
local DataGrid = lux.DataGrid
local Pagination = lux.Pagination
local Alert = lux.Alert
local Spinner = lux.Spinner

-- Simulated "API" data (in a real app, this would come from lumina.fetch)
local function generateData()
	local data = {}
	local statuses = { "running", "stopped", "pending", "error" }
	for i = 1, 50 do
		data[i] = {
			id = i,
			name = "service-" .. string.format("%02d", i),
			status = statuses[(i % 4) + 1],
			cpu = math.floor(i * 1.7),
			mem = math.floor(i * 128),
		}
	end
	return data
end

local PAGE_SIZE = 8

lumina.app {
	id = "api-client",
	store = {
		loading = true,
		error = nil,
		data = {},
		page = 1,
		selectedIdx = 1,
		sortColumnId = "",
		sortDirection = "",
	},
	keys = {
		["q"] = function() lumina.quit() end,
		["ctrl+c"] = function() lumina.quit() end,
		["r"] = function()
			-- Refresh: reload data
			lumina.store.set("loading", true)
			lumina.store.set("error", nil)
			lumina.spawn(function()
				local async = require("async")
				async.await(lumina.sleep(300))  -- Simulate network delay
				local data = generateData()
				lumina.store.set("data", data)
				lumina.store.set("loading", false)
				lumina.store.set("page", 1)
				lumina.store.set("selectedIdx", 1)
			end)
		end,
	},
	render = function()
		local t = lumina.getTheme()
		local loading = lumina.useStore("loading")
		local err = lumina.useStore("error")
		local data = lumina.useStore("data")
		local page = lumina.useStore("page")
		local selectedIdx = lumina.useStore("selectedIdx")
		local sortColumnId = lumina.useStore("sortColumnId")
		local sortDirection = lumina.useStore("sortDirection")

		-- Sort data
		local sorted = {}
		if type(data) == "table" then
			for i, r in ipairs(data) do sorted[i] = r end
		end
		if sortColumnId ~= "" then
			table.sort(sorted, function(a, b)
				local va, vb = a[sortColumnId], b[sortColumnId]
				if va == nil then va = "" end
				if vb == nil then vb = "" end
				if type(va) == "number" and type(vb) == "number" then
					if sortDirection == "desc" then return va > vb end
					return va < vb
				end
				va, vb = tostring(va), tostring(vb)
				if sortDirection == "desc" then return va > vb end
				return va < vb
			end)
		end

		-- Paginate
		local totalRows = #sorted
		local pageCount = math.max(1, math.ceil(totalRows / PAGE_SIZE))
		local startIdx = (page - 1) * PAGE_SIZE + 1
		local endIdx = math.min(page * PAGE_SIZE, totalRows)
		local pageRows = {}
		for i = startIdx, endIdx do
			pageRows[#pageRows + 1] = sorted[i]
		end

		local columns = {
			{ id = "id", header = "#", width = 5, key = "id", sortable = true, align = "right" },
			{ id = "name", header = "Service", width = 16, key = "name", sortable = true },
			{ id = "status", header = "Status", width = 10, key = "status", sortable = true },
			{ id = "cpu", header = "CPU%", width = 8, key = "cpu", sortable = true, align = "right" },
			{ id = "mem", header = "Mem MB", width = 9, key = "mem", sortable = true, align = "right" },
		}

		-- Build sort prop
		local sort = nil
		if sortColumnId ~= "" then
			sort = { columnId = sortColumnId, direction = sortDirection }
		end

		-- Build UI
		local children = {
			lumina.createElement("text", {
				bold = true, foreground = t.primary or "#89B4FA",
				style = { height = 1 },
			}, "  Service Monitor"),
			lumina.createElement("text", {
				foreground = t.muted or "#6C7086",
				style = { height = 1 },
			}, "  r=refresh  j/k=nav  click header=sort  q=quit"),
			lumina.createElement("text", { style = { height = 1 } }, ""),
		}

		if err then
			children[#children + 1] = Alert {
				key = "err-alert",
				variant = "error",
				title = "Error",
				message = err,
				width = 56,
				dismissible = true,
				onDismiss = function() lumina.store.set("error", nil) end,
			}
			children[#children + 1] = lumina.createElement("text", { style = { height = 1 } }, "")
		end

		if loading then
			children[#children + 1] = lumina.createElement("hbox", { style = { height = 1 } },
				Spinner { key = "sp" },
				lumina.createElement("text", {
					foreground = t.muted or "#6C7086",
					style = { height = 1 },
				}, "  Loading services...")
			)
		else
			children[#children + 1] = DataGrid {
				id = "services",
				key = "grid",
				width = 56,
				height = 12,
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
				autoFocus = true,
			}
			children[#children + 1] = lumina.createElement("text", { style = { height = 1 } }, "")
			children[#children + 1] = Pagination {
				key = "pager",
				pageCount = pageCount,
				currentPage = page,
				onPageChange = function(p)
					lumina.store.set("page", p)
					lumina.store.set("selectedIdx", 1)
				end,
				pageRangeDisplayed = 3,
				marginPagesDisplayed = 1,
			}
			children[#children + 1] = lumina.createElement("text", {
				foreground = t.muted or "#6C7086",
				style = { height = 1 },
			}, "  " .. tostring(totalRows) .. " services | Page " .. tostring(page) .. "/" .. tostring(pageCount))
		end

		return lumina.createElement("vbox", {
			id = "root",
			style = { width = 60, height = 24, background = t.base or "#1E1E2E" },
		}, table.unpack(children))
	end,
}

-- Initial data load (simulates fetching from an API endpoint)
lumina.spawn(function()
	local async = require("async")
	async.await(lumina.sleep(300))  -- Simulate initial network delay
	local data = generateData()
	lumina.store.set("data", data)
	lumina.store.set("loading", false)
end)
