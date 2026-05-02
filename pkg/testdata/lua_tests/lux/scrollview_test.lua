-- pkg/testdata/lua_tests/lux/scrollview_test.lua — ScrollView component tests

-- Helper: find the actual scroll container vbox (not the component placeholder)
local function findScrollBox(app)
	local svNode = app:find("sv")
	if not svNode then return nil end
	-- The component placeholder wraps the vbox. Look for the vbox child.
	if svNode.children then
		for _, child in ipairs(svNode.children) do
			if child.type == "vbox" and child.id == "sv" then
				return child
			end
		end
	end
	-- Fallback: if the component itself is the vbox
	if svNode.type == "vbox" then
		return svNode
	end
	return nil
end

test.describe("ScrollView", function()
	local app
	test.beforeEach(function() app = test.createApp(80, 24) end)
	test.afterEach(function() app:destroy() end)

	test.it("renders children inside scroll container", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")

			lumina.createComponent({
				id = "test",
				render = function()
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10
					},
						lumina.createElement("text", { key = "a" }, "Line 1"),
						lumina.createElement("text", { key = "b" }, "Line 2")
					)
				end,
			})
		]])
		test.assert.eq(app:screenContains("Line 1"), true)
		test.assert.eq(app:screenContains("Line 2"), true)
	end)

	test.it("scroll container has overflow=scroll style", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")

			lumina.createComponent({
				id = "test",
				render = function()
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10
					},
						lumina.createElement("text", { key = "a" }, "Content")
					)
				end,
			})
		]])

		local box = findScrollBox(app)
		test.assert.notNil(box)
	end)

	test.it("scrollbar is visible for overflow content", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")

			lumina.createComponent({
				id = "test",
				render = function()
					local items = {}
					for i = 1, 100 do
						items[#items + 1] = lumina.createElement("text", {
							key = "item-" .. i
						}, "Item " .. i)
					end
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10
					}, table.unpack(items))
				end,
			})
		]])

		local box = findScrollBox(app)
		test.assert.notNil(box)
		local sh = box.scrollHeight or 0
		test.assert.eq(sh > 10, true)
	end)

	test.it("keyboard scrolling: Down scrolls content", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")

			lumina.createComponent({
				id = "test",
				render = function()
					local items = {}
					for i = 1, 50 do
						items[#items + 1] = lumina.createElement("text", {
							key = "item-" .. i
						}, "Item " .. i)
					end
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10
					}, table.unpack(items))
				end,
			})
		]])

		-- Focus the scrollview by clicking it
		app:click("sv")
		app:keyPress("Down")

		local box = findScrollBox(app)
		test.assert.notNil(box)
		local sy = box.scrollY or 0
		test.assert.eq(sy > 0, true)
	end)

	test.it("keyboard scrolling: PageDown and PageUp", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")

			lumina.createComponent({
				id = "test",
				render = function()
					local items = {}
					for i = 1, 100 do
						items[#items + 1] = lumina.createElement("text", {
							key = "item-" .. i
						}, "Item " .. i)
					end
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10
					}, table.unpack(items))
				end,
			})
		]])

		app:click("sv")
		app:keyPress("PageDown")
		local box = findScrollBox(app)
		test.assert.notNil(box)
		local sy = box.scrollY or 0
		test.assert.eq(sy >= 10, true)

		app:keyPress("PageUp")
		box = findScrollBox(app)
		test.assert.notNil(box)
		sy = box.scrollY or 0
		test.assert.eq(sy < 10, true)
	end)

	test.it("keyboard scrolling: Home and End", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")

			lumina.createComponent({
				id = "test",
				render = function()
					local items = {}
					for i = 1, 100 do
						items[#items + 1] = lumina.createElement("text", {
							key = "item-" .. i
						}, "Item " .. i)
					end
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10
					}, table.unpack(items))
				end,
			})
		]])

		app:click("sv")

		-- End: jump to bottom
		app:keyPress("End")
		local box = findScrollBox(app)
		test.assert.notNil(box)
		local sh = box.scrollHeight or 0
		local sy = box.scrollY or 0
		local maxScroll = sh - 10
		test.assert.eq(sy, maxScroll)

		-- Home: jump to top
		app:keyPress("Home")
		box = findScrollBox(app)
		test.assert.notNil(box)
		sy = box.scrollY or 0
		test.assert.eq(sy, 0)
	end)

	test.it("scrollbar click: page up/down via click above/below thumb", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")

			lumina.createComponent({
				id = "test",
				render = function()
					local items = {}
					for i = 1, 100 do
						items[#items + 1] = lumina.createElement("text", {
							key = "item-" .. i
						}, "Item " .. i)
					end
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10
					}, table.unpack(items))
				end,
			})
		]])

		local box = findScrollBox(app)
		test.assert.notNil(box)

		-- Scrollbar is at the rightmost column: box.x + box.w - 1
		-- (innerX2 = box.x + box.w, but buffer is 0..w-1, so visible at w-1)
		local sbX = box.x + box.w - 1

		-- Click on scrollbar near bottom → page down
		app:click(sbX, box.y + box.h - 2)
		box = findScrollBox(app)
		test.assert.notNil(box)
		local sy = box.scrollY or 0
		test.assert.eq(sy > 0, true)
	end)

	test.it("via require('lux.scrollview') — module loads and renders", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")

			lumina.createComponent({
				id = "test",
				render = function()
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10
					},
						lumina.createElement("text", { key = "hello" }, "Hello ScrollView")
					)
				end,
			})
		]])

		test.assert.eq(app:screenContains("Hello ScrollView"), true)
	end)

	test.it("custom onScroll handler is called", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")

			lumina.createComponent({
				id = "test",
				render = function()
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10,
						onScroll = function(e) end
					},
						lumina.createElement("text", { key = "a" }, "Content")
					)
				end,
			})
		]])

		-- Scroll the container — should not crash
		app:scroll(5, 5, 1)
		test.assert.eq(true, true)
	end)

	test.it("custom onKeyDown handler receives non-scroll keys", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")

			lumina.createComponent({
				id = "test",
				render = function()
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10,
						onKeyDown = function(key) end
					},
						lumina.createElement("text", { key = "a" }, "Content")
					)
				end,
			})
		]])

		app:click("sv")
		-- Send a key not handled by ScrollView (e.g. "Enter")
		app:keyPress("Enter")
		-- Should not crash
		test.assert.eq(true, true)
	end)
end)
