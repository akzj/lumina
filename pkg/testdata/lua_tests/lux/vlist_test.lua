-- vlist_test.lua — Tests for the VList virtual scrolling component

test.describe("VList", function()
	local app

	test.beforeEach(function()
		app = test.createApp(80, 24)
	end)

	test.afterEach(function()
		app:destroy()
	end)

	-- Helper: load a VList app with given config
	local function loadVList(app, config)
		local totalCount = config.totalCount or 100
		local viewH = config.height or 10
		local overscan = config.overscan or 3
		local estimateHeight = config.estimateHeight or 1
		local itemPrefix = config.itemPrefix or "Item"

		app:loadString(string.format([[
			local VList = lumina.defineComponent("VList", function(props)
				local totalCount = props.totalCount or 0
				local renderItem = props.renderItem
				local overscan = props.overscan or 10
				local estimateHeight = props.estimateHeight or 1
				local viewH = props.height or 20

				local scrollY, setScrollY = lumina.useState("scrollY", 0)

				local startIdx = math.max(0, math.floor(scrollY / estimateHeight) - overscan)
				local endIdx = math.min(totalCount, math.ceil((scrollY + viewH) / estimateHeight) + overscan)

				local children = {}
				if startIdx > 0 then
					children[#children + 1] = lumina.createElement("box", {
						key = "vlist_top",
						style = { height = startIdx * estimateHeight },
					})
				end
				for i = startIdx, endIdx - 1 do
					local elem = renderItem(i)
					if elem then
						children[#children + 1] = elem
					end
				end
				if endIdx < totalCount then
					children[#children + 1] = lumina.createElement("box", {
						key = "vlist_bottom",
						style = { height = (totalCount - endIdx) * estimateHeight },
					})
				end

				return lumina.createElement("vbox", {
					id = props.id,
					style = { height = viewH, overflow = "scroll" },
					onScroll = function(e)
						setScrollY(e.scrollY)
					end,
				}, table.unpack(children))
			end)

			lumina.createComponent({
				id = "test", name = "Test",
				render = function()
					return lumina.createElement(VList, {
						id = "vlist",
						totalCount = %d,
						height = %d,
						overscan = %d,
						estimateHeight = %d,
						renderItem = function(index)
							return lumina.createElement("text", {
								key = "item-" .. index,
								style = { height = %d },
							}, "%s " .. index)
						end,
					})
				end,
			})
		]], totalCount, viewH, overscan, estimateHeight, estimateHeight, itemPrefix))
	end

	-- Helper: find a node in the VNode tree by key (DFS)
	local function findNodeByKey(node, key)
		if node == nil then return nil end
		if node.key == key then return node end
		if node.children then
			for _, child in ipairs(node.children) do
				local found = findNodeByKey(child, key)
				if found then return found end
			end
		end
		return nil
	end

	-- Helper: find the scroll container vbox inside the component
	local function findScrollContainer(app)
		local tree = app:vnodeTree()
		-- The tree is: component(id=vlist) → vbox(id=vlist, scrollHeight=...)
		-- find() returns the component placeholder; we need the inner vbox
		local vboxes = app:findAll("vbox")
		for _, vb in ipairs(vboxes) do
			if vb.scrollHeight then
				return vb
			end
		end
		return nil
	end

	test.it("renders only visible items + overscan (not all items)", function()
		loadVList(app, { totalCount = 100, height = 10, overscan = 3, estimateHeight = 1 })

		-- Items far outside viewport should NOT be on screen
		test.assert.eq(app:screenContains("Item 50"), false)
		test.assert.eq(app:screenContains("Item 99"), false)

		-- Items within the 10-row viewport (rows 0-9) should be visible
		test.assert.eq(app:screenContains("Item 0"), true)
		test.assert.eq(app:screenContains("Item 5"), true)
		test.assert.eq(app:screenContains("Item 9"), true)

		-- Items beyond viewport but within overscan are RENDERED but not visible on screen
		-- So Item 12 won't be on screen (it's at row 12, viewport ends at row 9)

		-- Verify not all 100 items are in the node tree
		local texts = app:findAll("text")
		-- We expect at most viewH + 2*overscan = 10 + 6 = 16 text items
		test.assert.eq(#texts <= 20, true)
	end)

	test.it("scrolling down changes visible items", function()
		loadVList(app, { totalCount = 100, height = 10, overscan = 3, estimateHeight = 1 })

		-- Initially: items 0-9 visible in viewport
		test.assert.eq(app:screenContains("Item 0"), true)
		test.assert.eq(app:screenContains("Item 50"), false)

		-- Scroll down: 5 ticks * 3 lines/tick = 15 lines → scrollY=15
		for i = 1, 5 do
			app:scroll(5, 2, 1)
		end

		-- After scrolling 15 lines, viewport shows rows 15-24
		-- Item 0 at row 0 is scrolled off screen
		test.assert.eq(app:screenContains("Item 0"), false)
		-- Item 15 at row 15 should be at top of viewport
		test.assert.eq(app:screenContains("Item 15"), true)
		-- Item 20 at row 20 should be visible
		test.assert.eq(app:screenContains("Item 20"), true)
	end)

	test.it("scrolling up brings back earlier items", function()
		loadVList(app, { totalCount = 100, height = 10, overscan = 3, estimateHeight = 1 })

		-- Scroll down first
		for i = 1, 5 do
			app:scroll(5, 2, 1)
		end

		-- Item 0 should be gone
		test.assert.eq(app:screenContains("Item 0"), false)

		-- Scroll back up: 5 ticks * -1 delta * 3 lines/tick = -15 → scrollY=0
		for i = 1, 5 do
			app:scroll(5, 2, -1)
		end

		-- Item 0 should be back
		test.assert.eq(app:screenContains("Item 0"), true)
	end)

	test.it("spacer nodes create correct total scroll height", function()
		-- 50 items, each 2 rows high, viewport=10 → total content = 100 rows
		loadVList(app, { totalCount = 50, height = 10, overscan = 3, estimateHeight = 2 })

		-- Find the inner scroll container (vbox with scrollHeight)
		local scrollBox = findScrollContainer(app)
		test.assert.notNil(scrollBox)
		test.assert.eq(scrollBox.scrollHeight, 100)

		-- Bottom spacer should exist (key=vlist_bottom)
		local tree = app:vnodeTree()
		local bottomSpacer = findNodeByKey(tree, "vlist_bottom")
		test.assert.notNil(bottomSpacer)
		-- Bottom spacer height: (50 - 8) * 2 = 84 (8 items visible: ceil(10/2)+3=5+3=8)
		test.assert.eq(bottomSpacer.h, 84)
	end)

	test.it("clamps startIdx to 0", function()
		loadVList(app, { totalCount = 100, height = 10, overscan = 3, estimateHeight = 1 })

		-- At scrollY=0, startIdx should be 0 (clamped)
		-- Item 0 must be visible
		test.assert.eq(app:screenContains("Item 0"), true)
	end)

	test.it("handles small totalCount (fewer items than viewport)", function()
		loadVList(app, { totalCount = 5, height = 10, overscan = 3, estimateHeight = 1 })

		-- All 5 items should be visible
		test.assert.eq(app:screenContains("Item 0"), true)
		test.assert.eq(app:screenContains("Item 4"), true)

		-- No item 5 should exist
		test.assert.eq(app:screenContains("Item 5"), false)

		-- Bottom spacer should not exist (endIdx >= totalCount)
		local tree = app:vnodeTree()
		local bottomSpacer = findNodeByKey(tree, "vlist_bottom")
		test.assert.isNil(bottomSpacer)
	end)

	test.it("handles zero totalCount", function()
		loadVList(app, { totalCount = 0, height = 10, overscan = 3, estimateHeight = 1 })

		-- Should not crash, no items visible
		test.assert.eq(app:screenContains("Item 0"), false)
	end)

	test.it("items have correct keys for reconciler reuse", function()
		loadVList(app, { totalCount = 100, height = 10, overscan = 3, estimateHeight = 1 })

		local tree = app:vnodeTree()

		-- Items should have keys like "item-0", "item-1", etc.
		local item0 = findNodeByKey(tree, "item-0")
		test.assert.notNil(item0)
		test.assert.eq(item0.content, "Item 0")

		-- Spacer keys
		local bottomSpacer = findNodeByKey(tree, "vlist_bottom")
		test.assert.notNil(bottomSpacer)
	end)

	test.it("estimateHeight > 1 changes scroll granularity", function()
		loadVList(app, { totalCount = 50, height = 10, overscan = 3, estimateHeight = 2 })

		-- With estimateHeight=2, each item takes 2 rows
		-- viewport=10 rows → 5 items fit in viewport
		-- Items 0-4 visible on screen initially

		test.assert.eq(app:screenContains("Item 0"), true)
		test.assert.eq(app:screenContains("Item 4"), true)
		-- Item 5 starts at row 10, which is at the very bottom edge — may or may not be visible

		-- Scroll down 4 ticks (4 * 3 = 12 lines)
		for i = 1, 4 do
			app:scroll(5, 2, 1)
		end

		-- scrollY = 12, viewport shows rows 12-21
		-- Item 0 at rows 0-1: scrolled off
		test.assert.eq(app:screenContains("Item 0"), false)
		-- Item 6 at rows 12-13: at top of viewport
		test.assert.eq(app:screenContains("Item 6"), true)
		-- Item 10 at rows 20-21: at bottom of viewport
		test.assert.eq(app:screenContains("Item 10"), true)
	end)

	test.it("via require('lux.vlist') — module loads and renders", function()
		-- Use the real lux.vlist module (registered in lux_modules.go)
		app:loadString([[
			local VList = require("lux.vlist")

			lumina.createComponent({
				id = "test", name = "Test",
				render = function()
					return lumina.createElement(VList, {
						id = "vlist-req",
						totalCount = 100,
						height = 10,
						overscan = 3,
						estimateHeight = 1,
						renderItem = function(index)
							return lumina.createElement("text", {
								key = "req-item-" .. index,
								style = { height = 1 },
							}, "ReqItem " .. index)
						end,
					})
				end,
			})
		]])

		-- Verify the real module renders correctly
		test.assert.eq(app:screenContains("ReqItem 0"), true)
		test.assert.eq(app:screenContains("ReqItem 5"), true)
		test.assert.eq(app:screenContains("ReqItem 50"), false) -- outside viewport

		-- Verify scrollHeight is set (totalCount * estimateHeight = 100)
		local scrollBox = findScrollContainer(app)
		test.assert.notNil(scrollBox)
		test.assert.eq(scrollBox.scrollHeight, 100)

		-- Scroll down and verify
		for i = 1, 5 do
			app:scroll(5, 2, 1)
		end
		test.assert.eq(app:screenContains("ReqItem 0"), false)
		test.assert.eq(app:screenContains("ReqItem 15"), true)
	end)

	test.it("clip-on-idle: component has isClipped state and renders", function()
		-- The real VList has clip-on-idle built in. Verify it renders without crashing.
		app:loadString([[
			local VList = require("lux.vlist")

			lumina.createComponent({
				id = "test", name = "Test",
				render = function()
					return lumina.createElement(VList, {
						id = "vlist-clip",
						totalCount = 200,
						height = 10,
						overscan = 10,
						estimateHeight = 1,
						clipDelay = 500,
						renderItem = function(index)
							return lumina.createElement("text", {
								key = "clip-item-" .. index,
								style = { height = 1 },
							}, "ClipItem " .. index)
						end,
					})
				end,
			})
		]])

		-- Verify items render
		test.assert.eq(app:screenContains("ClipItem 0"), true)
		test.assert.eq(app:screenContains("ClipItem 9"), true)

		-- Scroll a few times — onScroll sets isClipped=false (active scrolling mode)
		-- This exercises the clip-on-idle timer setup/teardown
		for i = 1, 3 do
			app:scroll(5, 2, 1)
		end

		-- After scrolling, items should still be rendered (no crash)
		test.assert.eq(app:screenContains("ClipItem 9"), true)
		-- Earlier items may or may not be visible depending on scroll position
		-- But the component should not have crashed
		local tree = app:vnodeTree()
		test.assert.notNil(tree)
	end)
end)
