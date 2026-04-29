-- api_client_test.lua — Acceptance tests for examples/api_client_demo.lua

test.describe("API client demo", function()
	local app

	test.beforeEach(function()
		app = test.createApp(60, 24)
		app:loadFile("../examples/api_client_demo.lua")
	end)

	test.afterEach(function()
		app:destroy()
	end)

	test.it("shows loading state initially", function()
		test.assert.eq(app:screenContains("Loading"), true)
		test.assert.eq(app:screenContains("Service Monitor"), true)
	end)

	test.it("shows data after async load completes", function()
		local ok = app:waitAsync(2000)
		app:render()
		test.assert.eq(ok, true)
		test.assert.eq(app:screenContains("service-"), true)
		test.assert.eq(app:screenContains("Service Monitor"), true)
	end)

	test.it("shows pagination info after load", function()
		app:waitAsync(2000)
		app:render()
		test.assert.eq(app:screenContains("50 services"), true)
	end)

	test.it("shows column headers after load", function()
		app:waitAsync(2000)
		app:render()
		test.assert.eq(app:screenContains("Service"), true)
		test.assert.eq(app:screenContains("Status"), true)
		test.assert.eq(app:screenContains("CPU"), true)
		test.assert.eq(app:screenContains("Mem"), true)
	end)

	test.it("keyboard navigation works after load", function()
		app:waitAsync(2000)
		app:render()
		app:keyPress("j")
		app:keyPress("j")
		-- Grid navigates without crash
		test.assert.eq(app:screenContains("service-"), true)
	end)

	test.it("refresh triggers loading state then reloads", function()
		app:waitAsync(2000)
		app:render()
		-- Data loaded
		test.assert.eq(app:screenContains("service-"), true)
		-- Press r to refresh
		app:keyPress("r")
		test.assert.eq(app:screenContains("Loading"), true)
		-- Wait for reload
		app:waitAsync(2000)
		app:render()
		test.assert.eq(app:screenContains("service-"), true)
	end)

	test.it("error state shows alert", function()
		app:waitAsync(2000)
		app:render()
		-- Simulate error
		app:loadString('lumina.store.set("error", "Connection timeout")')
		test.assert.eq(app:screenContains("Connection timeout"), true)
	end)

	test.it("error can be cleared", function()
		app:waitAsync(2000)
		app:render()
		app:loadString('lumina.store.set("error", "Connection timeout")')
		test.assert.eq(app:screenContains("Connection timeout"), true)
		-- Clear error
		app:loadString('lumina.store.set("error", nil)')
		test.assert.eq(app:screenContains("Connection timeout"), false)
	end)

	test.it("sort by name desc shows service-50 first", function()
		app:waitAsync(2000)
		app:render()
		app:loadString([[
			lumina.store.set("sortColumnId", "name")
			lumina.store.set("sortDirection", "desc")
			lumina.store.set("page", 1)
			lumina.store.set("selectedIdx", 1)
		]])
		test.assert.eq(app:screenContains("service-50"), true)
	end)

	test.it("page 2 shows different services", function()
		app:waitAsync(2000)
		app:render()
		test.assert.eq(app:screenContains("service-01"), true)
		app:loadString('lumina.store.set("page", 2) lumina.store.set("selectedIdx", 1)')
		test.assert.eq(app:screenContains("service-09"), true)
		test.assert.eq(app:screenContains("service-01"), false)
	end)

	test.it("sort + pagination combo", function()
		app:waitAsync(2000)
		app:render()
		-- Sort by name desc
		app:loadString([[
			lumina.store.set("sortColumnId", "name")
			lumina.store.set("sortDirection", "desc")
			lumina.store.set("page", 1)
			lumina.store.set("selectedIdx", 1)
		]])
		test.assert.eq(app:screenContains("service-50"), true)
		-- Go to page 2
		app:loadString('lumina.store.set("page", 2) lumina.store.set("selectedIdx", 1)')
		test.assert.eq(app:screenContains("service-42"), true)
		test.assert.eq(app:screenContains("service-50"), false)
	end)

	test.it("loading state hides grid", function()
		app:waitAsync(2000)
		app:render()
		test.assert.eq(app:screenContains("service-"), true)
		-- Set loading
		app:loadString('lumina.store.set("loading", true)')
		test.assert.eq(app:screenContains("Loading"), true)
	end)
end)
