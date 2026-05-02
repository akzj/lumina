-- wm_test.lua — Tests for the lux.wm WindowManager module

test.describe("WindowManager", function()
	local app

	test.beforeEach(function()
		app = test.createApp(80, 24)
	end)

	test.afterEach(function()
		app:destroy()
	end)

	test.it("create initializes with windows in order", function()
		app:loadString([[
			local WM = require("lux.wm")
			local mgr = WM.create("test_wm", {
				{id = "a", title = "A", x = 0, y = 0, w = 10, h = 5},
				{id = "b", title = "B", x = 5, y = 3, w = 10, h = 5},
			})

			lumina.createComponent({
				id = "wm-test",
				render = function()
					local wins = mgr.getWindows()
					local activeId = mgr.getActiveId()
					local lines = {}
					for _, w in ipairs(wins) do
						local marker = (w.id == activeId) and "*" or " "
						lines[#lines + 1] = marker .. w.id
					end
					return lumina.createElement("text", {}, "WM:" .. table.concat(lines, ","))
				end,
			})
		]])

		-- b is last = topmost = active
		test.assert.eq(app:screenContains("WM: a,*b"), true)
	end)

	test.it("register adds window at top", function()
		app:loadString([[
			local WM = require("lux.wm")
			local mgr = WM.create("test_wm2", {
				{id = "a", title = "A", x = 0, y = 0, w = 10, h = 5},
			})

			mgr.register("c", {title = "C", x = 10, y = 10, w = 15, h = 8})

			lumina.createComponent({
				id = "wm-test2",
				render = function()
					local wins = mgr.getWindows()
					local activeId = mgr.getActiveId()
					local lines = {}
					for _, w in ipairs(wins) do
						local marker = (w.id == activeId) and "*" or " "
						lines[#lines + 1] = marker .. w.id
					end
					return lumina.createElement("text", {}, "WM2:" .. table.concat(lines, ","))
				end,
			})
		]])

		-- c is last = topmost = active
		test.assert.eq(app:screenContains("WM2: a,*c"), true)
	end)

	test.it("close removes from order, preserves frame", function()
		app:loadString([[
			local WM = require("lux.wm")
			local mgr = WM.create("test_wm3", {
				{id = "a", title = "A", x = 0, y = 0, w = 10, h = 5},
				{id = "b", title = "B", x = 5, y = 3, w = 10, h = 5},
			})

			mgr.close("b")

			lumina.createComponent({
				id = "wm-test3",
				render = function()
					local wins = mgr.getWindows()
					local activeId = mgr.getActiveId()
					local lines = {}
					for _, w in ipairs(wins) do
						local marker = (w.id == activeId) and "*" or " "
						lines[#lines + 1] = marker .. w.id
					end
					return lumina.createElement("text", {}, "WM3:" .. table.concat(lines, ","))
				end,
			})
		]])

		-- only a remains, active is now a
		test.assert.eq(app:screenContains("WM3:*a"), true)
	end)

	test.it("reopen restores closed window at top", function()
		app:loadString([[
			local WM = require("lux.wm")
			local mgr = WM.create("test_wm4", {
				{id = "a", title = "A", x = 0, y = 0, w = 10, h = 5},
				{id = "b", title = "B", x = 5, y = 3, w = 10, h = 5},
			})

			mgr.close("b")
			mgr.reopen("b")

			lumina.createComponent({
				id = "wm-test4",
				render = function()
					local wins = mgr.getWindows()
					local activeId = mgr.getActiveId()
					local lines = {}
					for _, w in ipairs(wins) do
						local marker = (w.id == activeId) and "*" or " "
						lines[#lines + 1] = marker .. w.id
					end
					return lumina.createElement("text", {}, "WM4:" .. table.concat(lines, ","))
				end,
			})
		]])

		-- b is back at top, active
		test.assert.eq(app:screenContains("WM4: a,*b"), true)
	end)

	test.it("activate brings window to front", function()
		app:loadString([[
			local WM = require("lux.wm")
			local mgr = WM.create("test_wm5", {
				{id = "a", title = "A", x = 0, y = 0, w = 10, h = 5},
				{id = "b", title = "B", x = 5, y = 3, w = 10, h = 5},
				{id = "c", title = "C", x = 10, y = 6, w = 10, h = 5},
			})

			-- Activate a (currently first, should move to last)
			mgr.activate("a")

			lumina.createComponent({
				id = "wm-test5",
				render = function()
					local wins = mgr.getWindows()
					local activeId = mgr.getActiveId()
					local lines = {}
					for _, w in ipairs(wins) do
						local marker = (w.id == activeId) and "*" or " "
						lines[#lines + 1] = marker .. w.id
					end
					return lumina.createElement("text", {}, "WM5:" .. table.concat(lines, ","))
				end,
			})
		]])

		-- order: b, c, a (a moved to top), active is a
		test.assert.eq(app:screenContains("WM5: b, c,*a"), true)
	end)

	test.it("setFrame updates position without changing z-order", function()
		app:loadString([[
			local WM = require("lux.wm")
			local mgr = WM.create("test_wm6", {
				{id = "a", title = "A", x = 0, y = 0, w = 10, h = 5},
				{id = "b", title = "B", x = 5, y = 3, w = 10, h = 5},
			})

			-- Move window a without changing z-order
			mgr.setFrame("a", {x = 15, y = 10})

			lumina.createComponent({
				id = "wm-test6",
				render = function()
					local wins = mgr.getWindows()
					local activeId = mgr.getActiveId()
					local lines = {}
					for _, w in ipairs(wins) do
						local marker = (w.id == activeId) and "*" or " "
						lines[#lines + 1] = marker .. w.id .. ":" .. w.x .. "," .. w.y
					end
					return lumina.createElement("text", {}, "WM6:" .. table.concat(lines, "|"))
				end,
			})
		]])

		-- order unchanged (a, b), a's position updated, b unchanged
		test.assert.eq(app:screenContains("WM6: a:15,10|*b:5,3"), true)
	end)

	test.it("create is idempotent: second call does not overwrite", function()
		app:loadString([[
			local WM = require("lux.wm")

			local mgr1 = WM.create("test_wm7", {
				{id = "a", title = "A", x = 0, y = 0, w = 10, h = 5},
			})

			-- Modify state
			mgr1.setFrame("a", {x = 99, y = 99})

			-- Second create with same key — should NOT overwrite
			local mgr2 = WM.create("test_wm7", {
				{id = "z", title = "Z", x = 50, y = 50, w = 10, h = 5},
			})

			lumina.createComponent({
				id = "wm-test7",
				render = function()
					local wins = mgr2.getWindows()
					local activeId = mgr2.getActiveId()
					local lines = {}
					for _, w in ipairs(wins) do
						local marker = (w.id == activeId) and "*" or " "
						lines[#lines + 1] = marker .. w.id .. ":" .. w.x .. "," .. w.y
					end
					return lumina.createElement("text", {}, "WM7:" .. table.concat(lines, "|"))
				end,
			})
		]])

		-- still has "a" (not "z"), with modified position
		test.assert.eq(app:screenContains("WM7:*a:99,99"), true)
	end)

	test.it("close updates activeId when closing active window", function()
		app:loadString([[
			local WM = require("lux.wm")
			local mgr = WM.create("test_wm8", {
				{id = "a", title = "A", x = 0, y = 0, w = 10, h = 5},
				{id = "b", title = "B", x = 5, y = 3, w = 10, h = 5},
			})

			-- b is active (last). Close b.
			mgr.close("b")

			lumina.createComponent({
				id = "wm-test8",
				render = function()
					local wins = mgr.getWindows()
					local activeId = mgr.getActiveId()
					local lines = {}
					for _, w in ipairs(wins) do
						local marker = (w.id == activeId) and "*" or " "
						lines[#lines + 1] = marker .. w.id
					end
					return lumina.createElement("text", {}, "WM8:" .. table.concat(lines, ","))
				end,
			})
		]])

		-- a is now active
		test.assert.eq(app:screenContains("WM8:*a"), true)
	end)

	test.it("register prevents duplicate ids in order", function()
		app:loadString([[
			local WM = require("lux.wm")
			local mgr = WM.create("test_wm9", {
				{id = "a", title = "A", x = 0, y = 0, w = 10, h = 5},
				{id = "b", title = "B", x = 5, y = 3, w = 10, h = 5},
			})

			-- Re-register "a" — should move to top, no duplicate
			mgr.register("a", {title = "A Updated", x = 20, y = 20, w = 15, h = 8})

			lumina.createComponent({
				id = "wm-test9",
				render = function()
					local wins = mgr.getWindows()
					local activeId = mgr.getActiveId()
					local lines = {}
					for _, w in ipairs(wins) do
						local marker = (w.id == activeId) and "*" or " "
						lines[#lines + 1] = marker .. w.id .. ":" .. w.x .. "," .. w.y
					end
					return lumina.createElement("text", {}, "WM9:" .. table.concat(lines, "|"))
				end,
			})
		]])

		-- 2 windows, no duplicate, a is last with updated frame
		test.assert.eq(app:screenContains("WM9: b:5,3|*a:20,20"), true)
	end)
end)
