-- lua/lux/data_grid.lua — Lux DataGrid: fixed header, scrollable body, rich cells.
-- See lua/lux/data_grid.md for design and P0/P1 scope.
-- Usage: local DataGrid = require("lux.data_grid")

-- UTF-8 helpers: count characters and truncate safely.
-- utf8Len returns the number of UTF-8 characters (not bytes) in a string.
local function utf8Len(s)
	local count = 0
	local i = 1
	local n = #s
	while i <= n do
		local b = s:byte(i)
		if b < 0x80 then
			i = i + 1
		elseif b < 0xC0 then
			-- Continuation byte (shouldn't start a char) — skip
			i = i + 1
		elseif b < 0xE0 then
			i = i + 2
		elseif b < 0xF0 then
			i = i + 3
		else
			i = i + 4
		end
		count = count + 1
	end
	return count
end

-- utf8Sub returns the substring from character position i to j (1-based, inclusive).
-- Like string.sub but operates on UTF-8 characters, not bytes.
local function utf8Sub(s, i, j)
	local n = #s
	local charIdx = 0
	local byteStart = nil
	local byteEnd = nil
	local pos = 1
	while pos <= n do
		charIdx = charIdx + 1
		if charIdx == i then byteStart = pos end
		local b = s:byte(pos)
		local charLen = 1
		if b >= 0xF0 then charLen = 4
		elseif b >= 0xE0 then charLen = 3
		elseif b >= 0xC0 then charLen = 2
		end
		if charIdx == j then
			byteEnd = pos + charLen - 1
			break
		end
		pos = pos + charLen
	end
	if not byteStart then return "" end
	if not byteEnd then byteEnd = n end
	return s:sub(byteStart, byteEnd)
end

-- Pad/truncate text to a given width (in characters) with alignment.
local function padText(str, width, align)
	local len = utf8Len(str)
	if len >= width then return utf8Sub(str, 1, width) end
	local pad = width - len
	if align == "right" then
		return string.rep(" ", pad) .. str
	elseif align == "center" then
		local left = math.floor(pad / 2)
		return string.rep(" ", left) .. str .. string.rep(" ", pad - left)
	else -- "left" or default
		return str .. string.rep(" ", pad)
	end
end

local DataGrid = lumina.defineComponent("DataGrid", function(props)
	local rows = props.rows or {}
	local columns = props.columns or {}
	local renderCell = props.renderCell
	local sort = props.sort
	local t = lumina.getTheme and lumina.getTheme() or {}

	-- Multi-select props
	local selectionMode = props.selectionMode or "single"
	local selectedIds = props.selectedIds or {}
	local onSelectionChange = props.onSelectionChange
	local getRowId = props.getRowId or function(row, index) return tostring(index) end

	-- Editable cell props
	local editable = props.editable == true
	local editingCell = props.editingCell -- { rowIndex = N, columnId = "col" } or nil
	local onEditStart = props.onEditStart
	local onEditEnd = props.onEditEnd
	local onCellChange = props.onCellChange
	local onEditCancel = props.onEditCancel
	local editableColumns = props.editableColumns -- nil = all columns editable

	-- Controlled edit value: parent manages via store
	local editValueProp = props.editValue
	local onEditValueChange = props.onEditValueChange

	-- Closure variable to capture latest edit value from onChange (for onSubmit)
	local lastEditValue = editValueProp

	-- Focus the edit input after it's rendered
	-- Use a comparable string key as dep (Lua tables are not comparable in Go's depsEqual)
	local editingKey = editingCell
		and (tostring(editingCell.rowIndex) .. "-" .. tostring(editingCell.columnId))
		or nil
	lumina.useEffect(function()
		if editingCell then
			local editId = "edit-" .. tostring(editingCell.rowIndex) .. "-" .. tostring(editingCell.columnId)
			lumina.focusById(editId)
		end
	end, {editingKey})

	-- Build selectedIds lookup set for O(1) membership test
	local selectedIdSet = {}
	for _, id in ipairs(selectedIds) do
		selectedIdSet[id] = true
	end

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

	-- Helper: find first editable column ID
	local function firstEditableColumnId()
		for _, col in ipairs(columns) do
			if not editableColumns or editableColumns[col.id] then
				return col.id
			end
		end
		return nil
	end

	local function onKeyDown(e)
		if #rows == 0 then
			return
		end

		-- Handle Escape: cancel edit mode
		if e.key == "Escape" then
			if editingCell and onEditCancel then
				onEditCancel(editingCell.rowIndex, editingCell.columnId)
			end
			return
		end

		-- When in edit mode, don't process navigation keys
		if editingCell then
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
		elseif e.key == "Enter" or e.key == "F2" then
			if editable then
				-- Enter edit mode on first editable column
				local colId = firstEditableColumnId()
				if colId and onEditStart then
					onEditStart(selectedIdx, colId)
				end
			elseif props.onActivate and rows[selectedIdx] then
				props.onActivate(selectedIdx, rows[selectedIdx])
			end
		elseif e.key == "Home" then
			if props.onChangeIndex then props.onChangeIndex(1) end
		elseif e.key == "End" then
			if props.onChangeIndex then props.onChangeIndex(#rows) end
		elseif e.key == "PageUp" then
			local pageSize = math.max(1, bodyHeight)
			local n = math.max(1, selectedIdx - pageSize)
			if props.onChangeIndex then props.onChangeIndex(n) end
		elseif e.key == "PageDown" then
			local pageSize = math.max(1, bodyHeight)
			local n = math.min(#rows, selectedIdx + pageSize)
			if props.onChangeIndex then props.onChangeIndex(n) end
		elseif e.key == " " then
			-- Space toggles selection of current row in multi mode
			if selectionMode == "multi" and onSelectionChange then
				local row = rows[selectedIdx]
				if row then
					local id = getRowId(row, selectedIdx)
					local newIds = {}
					local found = false
					for _, sid in ipairs(selectedIds) do
						if sid == id then
							found = true
						else
							newIds[#newIds + 1] = sid
						end
					end
					if not found then
						newIds[#newIds + 1] = id
					end
					onSelectionChange(newIds)
				end
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
			foreground = t.error or "#F87171",
			style = { height = 1 },
			background = t.surface0 or "#141C2C",
		}, "  DataGrid: columns required")
	else
		if props.renderHeaderCell then
			for _, col in ipairs(columns) do
				headerCells[#headerCells + 1] = props.renderHeaderCell(col, { sort = sort })
			end
		else
			for _, col in ipairs(columns) do
				local w = columnWidth(col)
				local label = col.header or col.id or "?"

				-- Sort indicator
				local sortIndicator = ""
				if col.sortable and sort and sort.columnId == col.id then
					if sort.direction == "asc" then
						sortIndicator = " \226\150\178"
					else
						sortIndicator = " \226\150\188"
					end
				elseif col.sortable then
					sortIndicator = " \226\135\133"
				end

				local headerFg = (sort and sort.columnId == col.id)
					and (t.primary or "#F5C842")
					or (t.text or "#E8EDF7")

				headerCells[#headerCells + 1] = lumina.createElement("text", {
					bold = true,
					foreground = headerFg,
					background = t.surface0 or "#141C2C",
					style = { width = w, height = 1 },
					onClick = col.sortable and function()
						if props.onSortChange then
							local newDir = "asc"
							if sort and sort.columnId == col.id and sort.direction == "asc" then
								newDir = "desc"
							end
							props.onSortChange({ columnId = col.id, direction = newDir })
						end
					end or nil,
				}, " " .. label .. sortIndicator)
			end
		end
	end

	local headerRow = lumina.createElement("hbox", {
		style = { height = headerH },
	}, table.unpack(headerCells))

	local sep = lumina.createElement("text", {
		foreground = t.surface1 or "#1B2639",
		dim = true,
		style = { height = sepH, background = t.base or "#0B1220" },
	}, string.rep("\226\148\128", sepW))

	-- Default renderCell when none provided
	if type(renderCell) ~= "function" then
		renderCell = function(row, rowIndex, col, ctx)
			local v
			if type(col.accessor) == "function" then
				v = col.accessor(row)
			elseif col.key then
				v = row[col.key]
			end
			if v == nil then v = "" end
			if type(v) ~= "string" then v = tostring(v) end

			local fg = ctx.selected and (t.primary or "#F5C842") or (t.text or "#E8EDF7")
			local bg = ctx.selected and (t.surface1 or "#1B2639") or (t.base or "#0B1220")
			if ctx.multiSelected then
				bg = t.surface0 or "#141C2C"
			end
			local w = columnWidth(col)
			-- Selection indicator prefix in multi mode
			local prefix = " "
			if selectionMode == "multi" then
				if ctx.multiSelected then
					prefix = "\226\151\143 "  -- ● (U+25CF)
				else
					prefix = "\226\151\139 "  -- ○ (U+25CB)
				end
			end
			local text = prefix .. v
			text = padText(text, w, col.align)
			return lumina.createElement("text", {
				foreground = fg,
				background = bg,
				style = { height = 1 },
			}, text)
		end
	end

	local bodyChildren = {}
	if ncols == 0 then
		bodyChildren[1] = lumina.createElement("text", {
			foreground = t.error or "#F87171",
			style = { height = 1 },
			background = t.base or "#0B1220",
		}, "  DataGrid: columns required")
	elseif #rows == 0 then
		local emptyText = props.empty
		if type(emptyText) ~= "string" then
			emptyText = "No rows"
		end
		bodyChildren[1] = lumina.createElement("text", {
			foreground = t.muted or "#8B9BB4",
			style = { height = 1 },
			background = t.base or "#0B1220",
		}, "  " .. emptyText)
	else
		-- Determine row range to render
		local renderStart = 1
		local renderEnd = #rows

		if props.virtualScroll then
			local buffer = props.virtualBuffer or 3
			local visibleCount = math.ceil(bodyHeight / rowHeight)
			local windowStart = math.floor(scrollY / rowHeight) + 1
			renderStart = math.max(1, windowStart - buffer)
			renderEnd = math.min(#rows, windowStart + visibleCount + buffer)

			-- Top spacer (represents rows above rendered window)
			local topHeight = (renderStart - 1) * rowHeight
			if topHeight > 0 then
				bodyChildren[#bodyChildren + 1] = lumina.createElement("vbox", {
					key = "vs-top",
					style = { height = topHeight },
				})
			end
		end

		for i = renderStart, renderEnd do
			local row = rows[i]
			local rowId = getRowId(row, i)
			local ctx = {
				selected = (i == selectedIdx),
				multiSelected = (selectedIdSet[rowId] == true),
			}
			local cells = {}
			for _, col in ipairs(columns) do
				local cw = columnWidth(col)
				local colId = col.id or col.key or "?"
				local isEditing = editingCell
					and editingCell.rowIndex == i
					and editingCell.columnId == colId
				local cell
				if isEditing then
					-- Get current cell value
					local currentValue = ""
					if type(col.accessor) == "function" then
						local v = col.accessor(row)
						if v ~= nil then currentValue = tostring(v) end
					elseif col.key then
						local v = row[col.key]
						if v ~= nil then currentValue = tostring(v) end
					end
					-- Controlled input: value comes from editValueProp (store-backed)
					-- If editValueProp is nil, use currentValue as initial
					local inputValue = editValueProp
					if inputValue == nil then inputValue = currentValue end
					lastEditValue = inputValue
					-- Render native input for editing
					cell = lumina.createElement("input", {
						id = "edit-" .. tostring(i) .. "-" .. colId,
						value = inputValue,
						focusable = true,
						foreground = t.text or "#E8EDF7",
						background = t.surface1 or "#1B2639",
						style = { height = 1, width = cw },
						onChange = function(text)
							lastEditValue = text
							if onEditValueChange then
								onEditValueChange(text)
							end
						end,
						onSubmit = function()
							-- onChange fires RIGHT BEFORE onSubmit in same event cycle
							-- lastEditValue has the latest content
							if onCellChange then
								onCellChange(i, colId, lastEditValue)
							end
							if onEditEnd then
								onEditEnd(i, colId, lastEditValue, false)
							end
						end,
					})
				else
					cell = renderCell(row, i, col, ctx)
				end
				cells[#cells + 1] = lumina.createElement("vbox", {
					key = "c-" .. tostring(i) .. "-" .. tostring(colId),
					style = { width = cw, height = rowHeight },
				}, cell)
			end
			bodyChildren[#bodyChildren + 1] = lumina.createElement("hbox", {
				key = "r-" .. tostring(i),
				style = { height = rowHeight },
			}, table.unpack(cells))
		end

		if props.virtualScroll then
			-- Bottom spacer (represents rows below rendered window)
			local bottomHeight = (#rows - renderEnd) * rowHeight
			if bottomHeight > 0 then
				bodyChildren[#bodyChildren + 1] = lumina.createElement("vbox", {
					key = "vs-bottom",
					style = { height = bottomHeight },
				})
			end
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
