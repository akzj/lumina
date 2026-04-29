-- examples/data_grid_demo.lua — Lux DataGrid (fixed header, scrollable body)
--
-- Demonstrates:
--   • require("lux").DataGrid with columns, rows, renderCell(row, rowIndex, column, ctx)
--   • j/k / arrows + Enter; autoFocus; opaque cell backgrounds
--
-- Usage: lumina examples/data_grid_demo.lua
-- Quit: q

local lux = require("lux")
local DataGrid = lux.DataGrid

local columns = {
	{ id = "code", header = "Code", width = 6, key = "code" },
	{ id = "name", header = "Name", width = 18, key = "name" },
	{ id = "note", header = "Note", width = 44, key = "note" },
}

local rows = {
	{ code = "A1", name = "Alpha", note = "First row" },
	{ code = "B2", name = "Bravo", note = "Second row" },
	{ code = "C3", name = "Charlie", note = "Third row" },
	{ code = "D4", name = "Delta", note = "Fourth" },
	{ code = "E5", name = "Echo", note = "Fifth" },
	{ code = "F6", name = "Foxtrot", note = "Sixth" },
	{ code = "G7", name = "Golf", note = "Seventh" },
	{ code = "H8", name = "Hotel", note = "Eighth" },
	{ code = "I9", name = "India", note = "Ninth" },
	{ code = "J0", name = "Juliet", note = "Tenth" },
	{ code = "K1", name = "Kilo", note = "Eleventh" },
	{ code = "L2", name = "Lima", note = "Twelfth" },
}

lumina.app {
	id = "data-grid-demo",
	store = {
		selectedIdx = 1,
		lastActivate = "",
	},
	keys = {
		["q"] = function() lumina.quit() end,
		["ctrl+c"] = function() lumina.quit() end,
	},
	render = function()
		local t = lumina.getTheme()
		local selectedIdx = lumina.useStore("selectedIdx")
		local lastActivate = lumina.useStore("lastActivate")

		local function cellText(str, ctx, fg)
			return lumina.createElement("text", {
				foreground = fg or (t.text or "#CDD6F4"),
				style = { height = 1, background = ctx.selected and (t.surface1 or "#45475A") or (t.base or "#1E1E2E") },
			}, str)
		end

		local function renderCell(row, rowIndex, col, ctx)
			local v = row[col.key] or ""
			if type(v) ~= "string" then
				v = tostring(v)
			end
			local fg = ctx.selected and (t.primary or "#89B4FA") or (t.text or "#CDD6F4")
			return cellText(" " .. v, ctx, fg)
		end

		return lumina.createElement("vbox", {
			id = "root",
			style = { width = 80, height = 24 },
		},
			lumina.createElement("text", {
				bold = true,
				foreground = t.text or "#CDD6F4",
				style = { height = 1, background = t.surface0 or "#313244" },
			}, "  DataGrid demo (Lux)"),
			lumina.createElement("text", {
				foreground = t.muted or "#6C7086",
				style = { height = 1 },
			}, "  j/k arrows Enter  |  q quit"),
			DataGrid {
				id = "demo-grid",
				width = 76,
				height = 14,
				columns = columns,
				rows = rows,
				selectedIndex = selectedIdx,
				onChangeIndex = function(i)
					lumina.store.set("selectedIdx", i)
				end,
				onActivate = function(i, row)
					lumina.store.set("lastActivate", row.name .. "@" .. tostring(i))
				end,
				renderCell = renderCell,
				autoFocus = true,
			},
			lumina.createElement("text", {
				foreground = t.muted or "#6C7086",
				style = { height = 1 },
			}, "  Selected: " .. tostring(rows[selectedIdx] and rows[selectedIdx].name or "")),
			lumina.createElement("text", {
				foreground = t.warning or "#F9E2AF",
				style = { height = 1 },
			}, "  " .. (lastActivate ~= "" and ("Last Enter: " .. lastActivate) or ""))
		)
	end,
}
