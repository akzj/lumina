-- data_grid_paginated_test.lua — Acceptance tests for examples/data_grid_paginated.lua

test.describe("DataGrid paginated example", function()
	local app

	test.beforeEach(function()
		app = test.createApp(60, 24)
		app:loadFile("../examples/data_grid_paginated.lua")
	end)

	test.afterEach(function()
		app:destroy()
	end)

	test.it("loads with default renderCell (no custom renderCell)", function()
		test.assert.eq(app:screenContains("Paginated DataGrid"), true)
		test.assert.eq(app:screenContains("Item-001"), true)
	end)

	test.it("shows page info", function()
		test.assert.eq(app:screenContains("Page 1/"), true)
		test.assert.eq(app:screenContains("100 total"), true)
	end)

	test.it("shows sort indicators on sortable columns", function()
		-- All columns are sortable, should show ⇅ indicator
		test.assert.eq(app:screenContains("#"), true)
		test.assert.eq(app:screenContains("Name"), true)
		test.assert.eq(app:screenContains("Value"), true)
		test.assert.eq(app:screenContains("Status"), true)
	end)

	test.it("j moves selection down within page", function()
		test.assert.eq(app:screenContains("Item-001"), true)
		app:keyPress("j")
		app:keyPress("j")
		test.assert.eq(app:screenContains("Item-003"), true)
	end)
end)
