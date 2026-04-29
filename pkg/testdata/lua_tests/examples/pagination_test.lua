-- pagination_test.lua — Acceptance tests for examples/pagination_demo.lua

test.describe("Pagination example", function()
	local app

	test.beforeEach(function()
		app = test.createApp(60, 20)
		app:loadFile("../examples/pagination_demo.lua")
	end)

	test.afterEach(function()
		app:destroy()
	end)

	test.it("loads and shows title and page 1 items", function()
		test.assert.eq(app:screenContains("Pagination demo"), true)
		test.assert.eq(app:screenContains("Page 1 of 12"), true)
		test.assert.eq(app:screenContains("Item 1"), true)
		test.assert.eq(app:screenContains("Item 10"), true)
	end)

	test.it("l moves to next page", function()
		test.assert.eq(app:screenContains("Page 1 of 12"), true)
		app:keyPress("l")
		test.assert.eq(app:screenContains("Page 2 of 12"), true)
		test.assert.eq(app:screenContains("Item 11"), true)
	end)

	test.it("h at page 1 stays at page 1 (no underflow)", function()
		test.assert.eq(app:screenContains("Page 1 of 12"), true)
		app:keyPress("h")
		test.assert.eq(app:screenContains("Page 1 of 12"), true)
	end)

	test.it("navigates to last page", function()
		for i = 1, 11 do
			app:keyPress("l")
		end
		test.assert.eq(app:screenContains("Page 12 of 12"), true)
		test.assert.eq(app:screenContains("Item 111"), true)
		test.assert.eq(app:screenContains("Item 120"), true)
	end)

	test.it("l at last page stays at last page (no overflow)", function()
		for i = 1, 11 do
			app:keyPress("l")
		end
		test.assert.eq(app:screenContains("Page 12 of 12"), true)
		app:keyPress("l")
		test.assert.eq(app:screenContains("Page 12 of 12"), true)
	end)

	test.it("h moves back a page", function()
		app:keyPress("l")
		app:keyPress("l")
		test.assert.eq(app:screenContains("Page 3 of 12"), true)
		app:keyPress("h")
		test.assert.eq(app:screenContains("Page 2 of 12"), true)
	end)
end)
