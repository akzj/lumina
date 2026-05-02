-- pkg/testdata/lua_tests/lux/vlist_scrollview_test.lua — VList + ScrollView integration tests

test.describe("VList inside ScrollView", function()
	local app

	test.beforeEach(function()
		app = test.createApp(80, 24)
	end)

	test.afterEach(function()
		app:destroy()
	end)

	-- Helper: find the scroll container vbox inside a component node
	local function findScrollBox(node)
		if not node then return nil end
		if node.children then
			for _, child in ipairs(node.children) do
				if child.type == "vbox" and child.scrollHeight then
					return child
				end
				local found = findScrollBox(child)
				if found then return found end
			end
		end
		return nil
	end

	test.it("VList inside ScrollView renders without crash", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")
			local VList = require("lux.vlist")

			lumina.createComponent({
				id = "test",
				render = function()
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10
					},
						lumina.createElement(VList, {
							id = "vlist-inner",
							totalCount = 200,
							height = 10,
							overscan = 3,
							estimateHeight = 1,
							renderItem = function(index)
								return lumina.createElement("text", {
									key = "vl-item-" .. index,
									style = { height = 1 },
								}, "VLItem " .. index)
							end,
						})
					)
				end,
			})
		]])

		-- Verify VList renders inside ScrollView
		test.assert.eq(app:screenContains("VLItem 0"), true)
		test.assert.eq(app:screenContains("VLItem 5"), true)

		-- Verify both scroll containers exist
		local svNode = app:find("sv")
		test.assert.notNil(svNode)

		local vlistNode = app:find("vlist-inner")
		test.assert.notNil(vlistNode)

		-- No crash
		test.assert.eq(true, true)
	end)

	test.it("VList inside ScrollView: scrollNode scrolls outer ScrollView", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")
			local VList = require("lux.vlist")

			lumina.createComponent({
				id = "test",
				render = function()
					-- VList height=20, ScrollView height=10 → ScrollView has overflow
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10
					},
						lumina.createElement(VList, {
							id = "vlist-inner2",
							totalCount = 200,
							height = 20,
							overscan = 3,
							estimateHeight = 1,
							renderItem = function(index)
								return lumina.createElement("text", {
									key = "vl2-item-" .. index,
									style = { height = 1 },
								}, "VL2Item " .. index)
							end,
						})
					)
				end,
			})
		]])

		-- Use scrollNode API directly to scroll the outer ScrollView
		app:loadString([[
			lumina.scrollNode("sv", 5)
		]])

		-- Verify outer ScrollView scrolled
		local svNode = app:find("sv")
		test.assert.notNil(svNode)
		local function findScrollBoxById(node, id)
			if node.children then
				for _, child in ipairs(node.children) do
					if child.type == "vbox" and child.id == id then
						return child
					end
					local found = findScrollBoxById(child, id)
					if found then return found end
				end
			end
			return nil
		end
		local box = findScrollBoxById(svNode, "sv")
		test.assert.notNil(box)
		local sy = box.scrollY or 0
		test.assert.eq(sy, 5)
	end)

	test.it("VList inside ScrollView: both have independent scrollHeights", function()
		app:loadString([[
			local ScrollView = require("lux.scrollview")
			local VList = require("lux.vlist")

			lumina.createComponent({
				id = "test",
				render = function()
					-- VList height=20, ScrollView height=10 → ScrollView has overflow
					return lumina.createElement(ScrollView, {
						id = "sv",
						height = 10
					},
						lumina.createElement(VList, {
							id = "vlist-inner3",
							totalCount = 200,
							height = 20,
							overscan = 3,
							estimateHeight = 1,
							renderItem = function(index)
								return lumina.createElement("text", {
									key = "vl3-item-" .. index,
									style = { height = 1 },
								}, "VL3Item " .. index)
							end,
						})
					)
				end,
			})
		]])

		-- The outer ScrollView's scrollHeight = VList's vbox height (20)
		-- The inner VList's scrollHeight = 200 (totalCount * estimateHeight)
		local svNode = app:find("sv")
		test.assert.notNil(svNode)
		local function findScrollBoxById(node, id)
			if node.children then
				for _, child in ipairs(node.children) do
					if child.type == "vbox" and child.id == id then
						return child
					end
					local found = findScrollBoxById(child, id)
					if found then return found end
				end
			end
			return nil
		end
		local outerBox = findScrollBoxById(svNode, "sv")
		test.assert.notNil(outerBox)
		-- Outer scrollHeight: VList vbox is height=20, viewport=10 → overflow=10
		test.assert.eq(outerBox.scrollHeight, 20)

		-- Inner VList has its own scrollHeight=200
		local vlistNode = app:find("vlist-inner3")
		test.assert.notNil(vlistNode)
		local innerBox = findScrollBoxById(vlistNode, "vlist-inner3")
		test.assert.notNil(innerBox)
		test.assert.eq(innerBox.scrollHeight, 200)
	end)
end)
