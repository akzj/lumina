-- data_grid_accessor_test.lua — Tests for DataGrid with col.accessor function,
-- default renderCell, column alignment, and empty sort prop.

test.describe("DataGrid accessor and defaults", function()
	local app

	test.beforeEach(function()
		app = test.createApp(60, 12)
		app:loadString([[
			local lux = require("lux")
			local DataGrid = lux.DataGrid

			lumina.app {
				id = "accessor-test",
				store = { selectedIdx = 1 },
				render = function()
					local selectedIdx = lumina.useStore("selectedIdx")
					return DataGrid {
						id = "grid",
						width = 50,
						height = 8,
						columns = {
							{ id = "label", header = "Label", width = 20, accessor = function(row)
								return row.first .. " " .. row.last
							end },
							{ id = "age", header = "Age", width = 10, key = "age", align = "right" },
							{ id = "mid", header = "Mid", width = 10, key = "tag", align = "center" },
						},
						rows = {
							{ first = "Alice", last = "Smith", age = 30, tag = "A" },
							{ first = "Bob", last = "Jones", age = 25, tag = "BB" },
							{ first = "Carol", last = "Lee", age = 40, tag = "CCC" },
						},
						selectedIndex = selectedIdx,
						onChangeIndex = function(i)
							lumina.store.set("selectedIdx", i)
						end,
						autoFocus = true,
					}
				end,
			}
		]])
	end)

	test.afterEach(function()
		app:destroy()
	end)

	test.it("accessor function renders combined value", function()
		test.assert.eq(app:screenContains("Alice Smith"), true)
		test.assert.eq(app:screenContains("Bob Jones"), true)
		test.assert.eq(app:screenContains("Carol Lee"), true)
	end)

	test.it("right-aligned column works with default renderCell", function()
		-- Age column is right-aligned, width=10
		-- Values 30, 25, 40 should be present
		test.assert.eq(app:screenContains("30"), true)
		test.assert.eq(app:screenContains("25"), true)
		test.assert.eq(app:screenContains("40"), true)
	end)

	test.it("center-aligned column works with default renderCell", function()
		test.assert.eq(app:screenContains("A"), true)
		test.assert.eq(app:screenContains("BB"), true)
		test.assert.eq(app:screenContains("CCC"), true)
	end)

	test.it("navigation works with accessor columns", function()
		test.assert.eq(app:screenContains("Alice Smith"), true)
		app:keyPress("j")
		app:keyPress("j")
		-- Selected row 3 (Carol Lee), press Enter — no crash = success
		app:keyPress("Enter")
		test.assert.eq(app:screenContains("Carol Lee"), true)
	end)

	test.it("empty sort prop does not crash", function()
		-- No sort prop set, should render fine with headers
		test.assert.eq(app:screenContains("Label"), true)
		test.assert.eq(app:screenContains("Age"), true)
		test.assert.eq(app:screenContains("Mid"), true)
	end)
end)
