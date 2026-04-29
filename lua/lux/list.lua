-- lua/lux/list.lua — Lux ListView: scrollable list with rich rows (Lua-only).
-- See lua/lux/list.md for design, props, and constraints.
-- Usage: local ListView = require("lux.list")

local ListView = lumina.defineComponent("ListView", function(props)
	local rows = props.rows or {}
	local renderRow = props.renderRow
	local t = lumina.getTheme and lumina.getTheme() or {}

	local height = props.height or 10
	local rowHeight = props.rowHeight or 1
	local selectedIdx = props.selectedIndex or 1
	if #rows == 0 then
		selectedIdx = 1
	elseif selectedIdx > #rows then
		selectedIdx = #rows
	elseif selectedIdx < 1 then
		selectedIdx = 1
	end

	local contentLines = #rows * rowHeight
	local maxScroll = math.max(0, contentLines - height)
	local ideal = 0
	if #rows > 0 then
		ideal = (selectedIdx - 1) * rowHeight - math.floor(height / 2)
	end
	local scrollY = math.max(0, math.min(maxScroll, ideal))

	local items = {}
	if #rows == 0 then
		local emptyText = props.empty
		if type(emptyText) ~= "string" then
			emptyText = "No items"
		end
		items[1] = lumina.createElement("text", {
			foreground = t.muted or "#6C7086",
			style = { height = 1 },
		}, "  " .. emptyText)
	else
		if type(renderRow) ~= "function" then
			items[1] = lumina.createElement("text", {
				foreground = t.error or "#F38BA8",
				style = { height = 1 },
			}, "  ListView: renderRow function required")
		else
			for i, row in ipairs(rows) do
				local ctx = { selected = (i == selectedIdx) }
				items[#items + 1] = renderRow(row, i, ctx)
			end
		end
	end

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

	local style = {
		height = height,
		overflow = "scroll",
	}
	if props.width then
		style.width = props.width
	end

	return lumina.createElement("vbox", {
		id = props.id,
		key = props.key,
		style = style,
		scrollY = scrollY,
		focusable = true,
		onKeyDown = onKeyDown,
	}, table.unpack(items))
end)

return ListView
