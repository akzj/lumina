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

	-- Sort tests: use loadString to set the global store
	test.it("sorting by name desc shows Item-100 first", function()
		test.assert.eq(app:screenContains("Item-001"), true)
		app:loadString([[
			lumina.store.set("sortColumnId", "name")
			lumina.store.set("sortDirection", "desc")
			lumina.store.set("page", 1)
			lumina.store.set("selectedIdx", 1)
		]])
		test.assert.eq(app:screenContains("Item-100"), true)
	end)

	test.it("sorting by name asc shows Item-001 first", function()
		app:loadString([[
			lumina.store.set("sortColumnId", "name")
			lumina.store.set("sortDirection", "asc")
			lumina.store.set("page", 1)
			lumina.store.set("selectedIdx", 1)
		]])
		test.assert.eq(app:screenContains("Item-001"), true)
	end)

	test.it("sort indicator shows asc arrow after sorting", function()
		app:loadString([[
			lumina.store.set("sortColumnId", "name")
			lumina.store.set("sortDirection", "asc")
		]])
		test.assert.eq(app:screenContains("\226\150\178"), true) -- ▲
	end)

	test.it("sort indicator shows desc arrow after sorting", function()
		app:loadString([[
			lumina.store.set("sortColumnId", "value")
			lumina.store.set("sortDirection", "desc")
		]])
		test.assert.eq(app:screenContains("\226\150\188"), true) -- ▼
	end)

	-- Pagination tests
	test.it("page 2 shows different items than page 1", function()
		test.assert.eq(app:screenContains("Item-001"), true)
		app:loadString('lumina.store.set("page", 2) lumina.store.set("selectedIdx", 1)')
		test.assert.eq(app:screenContains("Item-011"), true)
		test.assert.eq(app:screenContains("Item-001"), false)
	end)

	test.it("last page shows remaining items", function()
		app:loadString('lumina.store.set("page", 10) lumina.store.set("selectedIdx", 1)')
		test.assert.eq(app:screenContains("Item-100"), true)
		test.assert.eq(app:screenContains("Page 10/"), true)
	end)

	test.it("sort + pagination combo: sort desc then page 2", function()
		app:loadString([[
			lumina.store.set("sortColumnId", "name")
			lumina.store.set("sortDirection", "desc")
			lumina.store.set("page", 1)
			lumina.store.set("selectedIdx", 1)
		]])
		test.assert.eq(app:screenContains("Item-100"), true)

		app:loadString('lumina.store.set("page", 2) lumina.store.set("selectedIdx", 1)')
		test.assert.eq(app:screenContains("Item-090"), true)
		test.assert.eq(app:screenContains("Item-100"), false)
	end)
end)
