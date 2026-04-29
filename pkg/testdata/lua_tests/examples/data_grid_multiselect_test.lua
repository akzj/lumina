-- data_grid_multiselect_test.lua — Tests for DataGrid multi-select feature

test.describe("DataGrid multi-select", function()
	local app

	test.beforeEach(function()
		app = test.createApp(60, 20)
		app:loadString([[
			local lux = require("lux")
			local DataGrid = lux.DataGrid

			lumina.app {
				id = "ms-test",
				store = {
					selectedIdx = 1,
					selectedIds = {},
					selectionMode = "multi",
				},
				render = function()
					local selectedIdx = lumina.useStore("selectedIdx")
					local selectedIds = lumina.useStore("selectedIds")
					local selectionMode = lumina.useStore("selectionMode")

					return DataGrid {
						id = "grid",
						width = 50,
						height = 12,
						columns = {
							{ id = "name", header = "Name", width = 20, key = "name" },
							{ id = "val", header = "Value", width = 10, key = "val" },
						},
						rows = {
							{ name = "Alpha", val = 10 },
							{ name = "Bravo", val = 20 },
							{ name = "Charlie", val = 30 },
							{ name = "Delta", val = 40 },
							{ name = "Echo", val = 50 },
						},
						selectedIndex = selectedIdx,
						onChangeIndex = function(i) lumina.store.set("selectedIdx", i) end,
						selectionMode = selectionMode,
						selectedIds = selectedIds,
						onSelectionChange = function(ids) lumina.store.set("selectedIds", ids) end,
						getRowId = function(row, i) return row.name end,
						autoFocus = true,
					}
				end,
			}
		]])
	end)

	test.afterEach(function()
		app:destroy()
	end)

	-- Basic multi-select
	test.it("Space selects current row", function()
		app:keyPress(" ")
		test.assert.eq(app:screenContains("\226\151\143"), true)
	end)

	test.it("Space toggles selection off", function()
		app:keyPress(" ")  -- select Alpha
		test.assert.eq(app:screenContains("\226\151\143"), true)
		app:keyPress(" ")  -- deselect Alpha
		-- No filled indicator for any row
		test.assert.eq(app:screenContains("\226\151\143"), false)
	end)

	test.it("multiple rows can be selected", function()
		app:keyPress(" ")  -- select Alpha
		app:keyPress("j")  -- move to Bravo
		app:keyPress(" ")  -- select Bravo
		-- Both should show selected indicator (2 columns × 2 rows = 4 indicators)
		local screen = app:screenText()
		local count = 0
		for _ in screen:gmatch("\226\151\143") do count = count + 1 end
		test.assert.eq(count, 4)
	end)

	test.it("j/k still moves cursor without changing selection", function()
		app:keyPress(" ")  -- select Alpha
		app:keyPress("j")  -- move to Bravo (not selected yet)
		app:keyPress("j")  -- move to Charlie
		-- Alpha still selected
		test.assert.eq(app:screenContains("\226\151\143"), true)
		-- Only 1 row selected × 2 columns = 2 filled indicators
		local screen = app:screenText()
		local count = 0
		for _ in screen:gmatch("\226\151\143") do count = count + 1 end
		test.assert.eq(count, 2)
	end)

	test.it("single mode ignores Space for multi-select", function()
		app:loadString('lumina.store.set("selectionMode", "single")')
		app:keyPress(" ")
		-- No multi-select indicator
		test.assert.eq(app:screenContains("\226\151\143"), false)
	end)

	test.it("multi mode shows circle indicators for unselected rows", function()
		-- In multi mode, unselected rows show ○
		test.assert.eq(app:screenContains("\226\151\139"), true)
	end)

	-- Edge cases
	test.it("empty rows with multi-select does not crash", function()
		app:destroy()
		app = test.createApp(60, 12)
		app:loadString([[
			local lux = require("lux")
			local DataGrid = lux.DataGrid
			lumina.app {
				id = "empty-ms",
				render = function()
					return DataGrid {
						id = "g",
						width = 40, height = 8,
						columns = { { id = "x", header = "X", width = 10, key = "x" } },
						rows = {},
						selectionMode = "multi",
						selectedIds = {},
						selectedIndex = 1,
						autoFocus = true,
					}
				end,
			}
		]])
		app:keyPress(" ")
		-- Should not crash, shows empty message
		test.assert.eq(app:screenContains("No rows"), true)
	end)

	test.it("getRowId provides stable identity across re-renders", function()
		-- Select Alpha by Space
		app:keyPress(" ")
		test.assert.eq(app:screenContains("\226\151\143"), true)
		-- Move cursor away — Alpha should still show selected
		app:keyPress("j")
		app:keyPress("j")
		-- Alpha is still in selectedIds (1 row × 2 columns = 2 indicators)
		local screen = app:screenText()
		local count = 0
		for _ in screen:gmatch("\226\151\143") do count = count + 1 end
		test.assert.eq(count, 2)
	end)

	test.it("deselecting middle item preserves other selections", function()
		-- Select Alpha, Bravo, Charlie
		app:keyPress(" ")  -- Alpha
		app:keyPress("j")
		app:keyPress(" ")  -- Bravo
		app:keyPress("j")
		app:keyPress(" ")  -- Charlie
		-- 3 rows × 2 columns = 6 indicators
		local screen = app:screenText()
		local count = 0
		for _ in screen:gmatch("\226\151\143") do count = count + 1 end
		test.assert.eq(count, 6)

		-- Deselect Bravo (move back up)
		app:keyPress("k")
		app:keyPress(" ")  -- toggle Bravo off
		-- 2 rows × 2 columns = 4 indicators
		screen = app:screenText()
		count = 0
		for _ in screen:gmatch("\226\151\143") do count = count + 1 end
		test.assert.eq(count, 4)
	end)

	test.it("ctx.multiSelected passed to custom renderCell", function()
		app:destroy()
		app = test.createApp(60, 12)
		app:loadString([[
			local lux = require("lux")
			local DataGrid = lux.DataGrid
			lumina.app {
				id = "ctx-test",
				store = { idx = 1, ids = { "row1" } },
				render = function()
					return DataGrid {
						id = "g",
						width = 40, height = 8,
						columns = { { id = "x", header = "X", width = 20, key = "x" } },
						rows = { { x = "hello" }, { x = "world" } },
						selectedIndex = lumina.useStore("idx"),
						selectionMode = "multi",
						selectedIds = lumina.useStore("ids"),
						getRowId = function(row, i) return "row" .. tostring(i) end,
						renderCell = function(row, rowIndex, col, ctx)
							local mark = ctx.multiSelected and "[SEL]" or "[---]"
							return lumina.createElement("text", {
								style = { height = 1 },
							}, mark .. " " .. tostring(row[col.key]))
						end,
						autoFocus = true,
					}
				end,
			}
		]])
		-- Row 1 is in selectedIds as "row1", so should show [SEL]
		test.assert.eq(app:screenContains("[SEL] hello"), true)
		test.assert.eq(app:screenContains("[---] world"), true)
	end)

	test.it("selectionMode none shows no indicators and ignores Space", function()
		app:loadString('lumina.store.set("selectionMode", "none")')
		app:keyPress(" ")
		-- No multi-select indicators
		test.assert.eq(app:screenContains("\226\151\143"), false)
		test.assert.eq(app:screenContains("\226\151\139"), false)
	end)
end)
