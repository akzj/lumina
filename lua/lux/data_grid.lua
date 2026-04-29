-- lua/lux/data_grid.lua — Lux DataGrid: fixed header, scrollable body, rich cells.
-- See lua/lux/data_grid.md for design and P0 scope.
-- Usage: local DataGrid = require("lux.data_grid")

local DataGrid = lumina.defineComponent("DataGrid", function(props)
	local rows = props.rows or {}
	local columns = props.columns or {}
	local renderCell = props.renderCell
	local t = lumina.getTheme and lumina.getTheme() or {}

	local gridW = props.width
	local ncols = #columns
	local function columnWidth(col)
		if col.width and col.width > 0 then
			return col.width
		end
		if gridW and gridW > 0 and ncols > 0 then
			return math.max(3, math.floor(gridW / ncols))
		end
		return 12
	end

	local totalHeight = props.height or 12
	local rowHeight = props.rowHeight or 1
	local headerH = 1
	local sepH = 1
	local bodyHeight = math.max(1, totalHeight - headerH - sepH)

	local selectedIdx = props.selectedIndex or 1
	if #rows == 0 then
		selectedIdx = 1
	elseif selectedIdx > #rows then
		selectedIdx = #rows
	elseif selectedIdx < 1 then
		selectedIdx = 1
	end

	local contentLines = #rows * rowHeight
	local maxScroll = math.max(0, contentLines - bodyHeight)
	local ideal = 0
	if #rows > 0 then
		ideal = (selectedIdx - 1) * rowHeight - math.floor(bodyHeight / 2)
	end
	local scrollY = math.max(0, math.min(maxScroll, ideal))

	local function onKeyDown(e)
		if #rows == 0 then
			return
		end
		if e.key == "ArrowUp" or e.key == "Up" or e.key == "k" then
			local n = math.max(1, selectedIdx - 1)
			if props.onChangeIndex then
				props.onChangeIndex(n)
			end
		elseif e.key == "ArrowDown" or e.key == "Down" or e.key == "j" then
			local n = math.min(#rows, selectedIdx + 1)
			if props.onChangeIndex then
				props.onChangeIndex(n)
			end
		elseif e.key == "Enter" then
			if props.onActivate and rows[selectedIdx] then
				props.onActivate(selectedIdx, rows[selectedIdx])
			end
		end
	end

	local sepW = gridW
	if (not sepW or sepW < 4) and ncols > 0 then
		local s = 0
		for _, col in ipairs(columns) do
			s = s + columnWidth(col)
		end
		sepW = math.max(4, s)
	end
	sepW = sepW or 40

	local headerCells = {}
	if ncols == 0 then
		headerCells[1] = lumina.createElement("text", {
			foreground = t.error or "#F38BA8",
			style = { height = 1 },
			background = t.surface0 or "#313244",
		}, "  DataGrid: columns required")
	else
		if props.renderHeaderCell then
			for _, col in ipairs(columns) do
				headerCells[#headerCells + 1] = props.renderHeaderCell(col, {})
			end
		else
			for _, col in ipairs(columns) do
				local w = columnWidth(col)
				local label = col.header or col.id or "?"
				headerCells[#headerCells + 1] = lumina.createElement("text", {
					bold = true,
					foreground = t.text or "#CDD6F4",
					background = t.surface0 or "#313244",
					style = { width = w, height = 1 },
				}, " " .. label)
			end
		end
	end

	local headerRow = lumina.createElement("hbox", {
		style = { height = headerH },
	}, table.unpack(headerCells))

	local sep = lumina.createElement("text", {
		foreground = t.surface1 or "#45475A",
		dim = true,
		style = { height = sepH, background = t.base or "#1E1E2E" },
	}, string.rep("─", sepW))

	local bodyChildren = {}
	if ncols == 0 then
		bodyChildren[1] = lumina.createElement("text", {
			foreground = t.error or "#F38BA8",
			style = { height = 1 },
			background = t.base or "#1E1E2E",
		}, "  DataGrid: columns required")
	elseif #rows == 0 then
		local emptyText = props.empty
		if type(emptyText) ~= "string" then
			emptyText = "No rows"
		end
		bodyChildren[1] = lumina.createElement("text", {
			foreground = t.muted or "#6C7086",
			style = { height = 1 },
			background = t.base or "#1E1E2E",
		}, "  " .. emptyText)
	elseif type(renderCell) ~= "function" then
		bodyChildren[1] = lumina.createElement("text", {
			foreground = t.error or "#F38BA8",
			style = { height = 1 },
			background = t.base or "#1E1E2E",
		}, "  DataGrid: renderCell function required")
	else
		for i, row in ipairs(rows) do
			local ctx = { selected = (i == selectedIdx) }
			local cells = {}
			for _, col in ipairs(columns) do
				local cw = columnWidth(col)
				local cell = renderCell(row, i, col, ctx)
				cells[#cells + 1] = lumina.createElement("vbox", {
					key = "c-" .. tostring(i) .. "-" .. tostring(col.id or col.key or "?"),
					style = { width = cw, height = rowHeight },
				}, cell)
			end
			bodyChildren[#bodyChildren + 1] = lumina.createElement("hbox", {
				key = "r-" .. tostring(i),
				style = { height = rowHeight },
			}, table.unpack(cells))
		end
	end

	local bodyScroll = lumina.createElement("vbox", {
		style = {
			flex = 1,
			minHeight = 1,
			overflow = "scroll",
		},
		scrollY = scrollY,
	}, table.unpack(bodyChildren))

	local rootStyle = {
		height = totalHeight,
	}
	if gridW then
		rootStyle.width = gridW
	end

	return lumina.createElement("vbox", {
		id = props.id,
		key = props.key,
		style = rootStyle,
		focusable = true,
		autoFocus = props.autoFocus == true,
		onKeyDown = onKeyDown,
	}, headerRow, sep, bodyScroll)
end)

return DataGrid
