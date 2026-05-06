-- pkg/testdata/lua_tests/lux/scrollbar_hidden_test.lua — scrollbar="none" tests

test.describe("ScrollbarHidden", function()
	local app
	test.beforeEach(function() app = test.createApp(80, 24) end)
	test.afterEach(function() app:destroy() end)

	test.it("scrollbar='none' hides scrollbar characters", function()
		app:loadString([[
			lumina.createComponent({
				id = "test",
				render = function()
					local items = {}
					for i = 1, 50 do
						items[#items + 1] = lumina.createElement("text", {
							key = "item-" .. i
						}, "Line " .. i)
					end
					return lumina.createElement("vbox", {
						id = "scroll-box",
						style = {
							height = 10,
							overflow = "scroll",
							scrollbar = "none",
							border = "single",
						},
					}, table.unpack(items))
				end,
			})
		]])

		local text = app:screenText()
		-- No scrollbar characters should be present
		test.assert.eq(text:find("\226\150\136") == nil, true) -- █ (U+2588)
		test.assert.eq(text:find("\226\150\145") == nil, true) -- ░ (U+2591)
		-- But content should be visible
		test.assert.eq(text:find("Line 1") ~= nil, true)
	end)

	test.it("default scrollbar shows scrollbar characters", function()
		app:loadString([[
			lumina.createComponent({
				id = "test",
				render = function()
					local items = {}
					for i = 1, 50 do
						items[#items + 1] = lumina.createElement("text", {
							key = "item-" .. i
						}, "Line " .. i)
					end
					return lumina.createElement("vbox", {
						id = "scroll-box",
						style = {
							height = 10,
							overflow = "scroll",
							border = "single",
						},
					}, table.unpack(items))
				end,
			})
		]])

		local text = app:screenText()
		-- Scrollbar characters should be present
		test.assert.eq(text:find("\226\150\136") ~= nil, true) -- █ (U+2588)
	end)

	test.it("scrollbar='none' still allows scrolling", function()
		app:loadString([[
			lumina.createComponent({
				id = "test",
				render = function()
					local items = {}
					for i = 1, 50 do
						items[#items + 1] = lumina.createElement("text", {
							key = "item-" .. i
						}, "Line " .. i)
					end
					return lumina.createElement("vbox", {
						id = "scroll-box",
						style = {
							height = 10,
							width = 40,
							overflow = "scroll",
							scrollbar = "none",
							border = "single",
						},
					}, table.unpack(items))
				end,
			})
		]])

		-- Verify initial content
		test.assert.eq(app:screenContains("Line 1"), true)

		-- Scroll down many times (x=5, y=5, delta=1 means scroll down)
		for i = 1, 20 do
			app:scroll(5, 5, 1)
		end
		-- After heavy scrolling, Line 1 should be gone and later lines visible
		test.assert.eq(app:screenContains("Line 50"), true)
	end)
end)
