-- data_grid_test.lua — Acceptance tests for examples/data_grid_demo.lua

test.describe("DataGrid example", function()
	local app

	test.beforeEach(function()
		app = test.createApp(80, 24)
		app:loadFile("../examples/data_grid_demo.lua")
	end)

	test.afterEach(function()
		app:destroy()
	end)

	test.it("loads and shows title and column headers", function()
		test.assert.eq(app:screenContains("DataGrid demo"), true)
		test.assert.eq(app:screenContains("Code"), true)
		test.assert.eq(app:screenContains("Name"), true)
		test.assert.eq(app:screenContains("Note"), true)
		test.assert.eq(app:screenContains("Alpha"), true)
	end)

	test.it("j moves selection to next row (footer shows name)", function()
		test.assert.eq(app:screenContains("Selected: Alpha"), true)
		app:keyPress("j")
		test.assert.eq(app:screenContains("Selected: Bravo"), true)
		app:keyPress("j")
		test.assert.eq(app:screenContains("Selected: Charlie"), true)
	end)

	test.it("k moves selection up", function()
		app:keyPress("j")
		app:keyPress("j")
		test.assert.eq(app:screenContains("Selected: Charlie"), true)
		app:keyPress("k")
		test.assert.eq(app:screenContains("Selected: Bravo"), true)
	end)

	test.it("Enter sets last activate line", function()
		app:keyPress("j")
		app:keyPress("Enter")
		test.assert.eq(app:screenContains("Last Enter: Bravo@2"), true)
	end)

	test.it("Home goes to first row", function()
		app:keyPress("j")
		app:keyPress("j")
		test.assert.eq(app:screenContains("Selected: Charlie"), true)
		app:keyPress("Home")
		test.assert.eq(app:screenContains("Selected: Alpha"), true)
	end)

	test.it("End goes to last row", function()
		app:keyPress("End")
		test.assert.eq(app:screenContains("Selected: Lima"), true)
	end)

	test.it("PageDown jumps by page size", function()
		app:keyPress("PageDown")
		-- bodyHeight = 14 - 1(header) - 1(sep) = 12, so jumps 12 rows
		-- from row 1 to row 13 but clamped to #rows=12
		test.assert.eq(app:screenContains("Selected: Lima"), true)
	end)

	test.it("PageUp jumps back by page size", function()
		-- Move down several rows first
		for i = 1, 8 do app:keyPress("j") end
		-- Now at row 9 (India)
		test.assert.eq(app:screenContains("Selected: India"), true)
		app:keyPress("PageUp")
		-- bodyHeight=12, so jumps back from 9 to max(1, 9-12)=1
		test.assert.eq(app:screenContains("Selected: Alpha"), true)
	end)
end)
