-- data_grid_edit_test.lua — Tests for DataGrid editable cells feature

test.describe("DataGrid editable cells", function()
	local app

	test.afterEach(function()
		if app then app:destroy() end
	end)

	test.it("Enter key enters edit mode", function()
		app = test.createApp(60, 20)
		app:loadString([[
			local lux = require("lux")
			local DataGrid = lux.DataGrid
			lumina.app {
				id = "edit-test",
				store = {
					idx = 1,
					editCell = nil,
					editValue = nil,
				},
				render = function()
					local idx = lumina.useStore("idx")
					local editCell = lumina.useStore("editCell")
					local editValue = lumina.useStore("editValue")
					return DataGrid {
						id = "grid",
						width = 50,
						height = 12,
						columns = {
							{ id = "name", header = "Name", width = 20, key = "name" },
							{ id = "age", header = "Age", width = 10, key = "age" },
						},
						rows = {
							{ name = "Alice", age = "30" },
							{ name = "Bob", age = "25" },
						},
						selectedIndex = idx,
						onChangeIndex = function(i) lumina.store.set("idx", i) end,
						editable = true,
						editingCell = editCell,
						editValue = editValue,
						onEditStart = function(row, col)
							lumina.store.set("editCell", { rowIndex = row, columnId = col })
							lumina.store.set("editValue", nil)
						end,
						onEditValueChange = function(text)
							lumina.store.set("editValue", text)
						end,
						onCellChange = function(row, col, val)
							lumina.store.set("editCell", nil)
							lumina.store.set("editValue", nil)
						end,
						onEditCancel = function(row, col)
							lumina.store.set("editCell", nil)
							lumina.store.set("editValue", nil)
						end,
						autoFocus = true,
					}
				end,
			}
		]])
		-- Enter should trigger edit mode on first editable column
		app:keyPress("Enter")
		local node = app:find("edit-1-name")
		test.assert.notNil(node)
	end)

	test.it("F2 key enters edit mode", function()
		app = test.createApp(60, 20)
		app:loadString([[
			local lux = require("lux")
			local DataGrid = lux.DataGrid
			lumina.app {
				id = "f2-test",
				store = { idx = 1, editCell = nil, editValue = nil },
				render = function()
					local idx = lumina.useStore("idx")
					local editCell = lumina.useStore("editCell")
					local editValue = lumina.useStore("editValue")
					return DataGrid {
						id = "grid",
						width = 50, height = 12,
						columns = {
							{ id = "name", header = "Name", width = 20, key = "name" },
						},
						rows = { { name = "Alice" } },
						selectedIndex = idx,
						editable = true,
						editingCell = editCell,
						editValue = editValue,
						onEditStart = function(row, col)
							lumina.store.set("editCell", { rowIndex = row, columnId = col })
							lumina.store.set("editValue", nil)
						end,
						onEditValueChange = function(text)
							lumina.store.set("editValue", text)
						end,
						onEditCancel = function(row, col)
							lumina.store.set("editCell", nil)
							lumina.store.set("editValue", nil)
						end,
						autoFocus = true,
					}
				end,
			}
		]])
		app:keyPress("F2")
		local node = app:find("edit-1-name")
		test.assert.notNil(node)
	end)

	test.it("Escape cancels edit mode", function()
		app = test.createApp(60, 20)
		app:loadString([[
			local lux = require("lux")
			local DataGrid = lux.DataGrid
			lumina.app {
				id = "esc-test",
				store = { idx = 1, editCell = nil, editValue = nil },
				render = function()
					local idx = lumina.useStore("idx")
					local editCell = lumina.useStore("editCell")
					local editValue = lumina.useStore("editValue")
					return DataGrid {
						id = "grid",
						width = 50, height = 12,
						columns = {
							{ id = "name", header = "Name", width = 20, key = "name" },
						},
						rows = { { name = "Alice" } },
						selectedIndex = idx,
						editable = true,
						editingCell = editCell,
						editValue = editValue,
						onEditStart = function(row, col)
							lumina.store.set("editCell", { rowIndex = row, columnId = col })
							lumina.store.set("editValue", nil)
						end,
						onEditValueChange = function(text)
							lumina.store.set("editValue", text)
						end,
						onEditCancel = function(row, col)
							lumina.store.set("editCell", nil)
							lumina.store.set("editValue", nil)
						end,
						autoFocus = true,
					}
				end,
			}
		]])
		app:keyPress("Enter")  -- enter edit
		local node = app:find("edit-1-name")
		test.assert.notNil(node)
		app:keyPress("Escape")  -- cancel
		node = app:find("edit-1-name")
		test.assert.isNil(node)
	end)

	test.it("shows current cell value in edit input", function()
		app = test.createApp(60, 20)
		app:loadString([[
			local lux = require("lux")
			local DataGrid = lux.DataGrid
			lumina.app {
				id = "val-test",
				store = { idx = 1, editCell = nil, editValue = nil },
				render = function()
					local idx = lumina.useStore("idx")
					local editCell = lumina.useStore("editCell")
					local editValue = lumina.useStore("editValue")
					return DataGrid {
						id = "grid",
						width = 50, height = 12,
						columns = {
							{ id = "name", header = "Name", width = 20, key = "name" },
						},
						rows = { { name = "Alice" } },
						selectedIndex = idx,
						editable = true,
						editingCell = editCell,
						editValue = editValue,
						onEditStart = function(row, col)
							lumina.store.set("editCell", { rowIndex = row, columnId = col })
							lumina.store.set("editValue", nil)
						end,
						onEditValueChange = function(text)
							lumina.store.set("editValue", text)
						end,
						onEditCancel = function(row, col)
							lumina.store.set("editCell", nil)
							lumina.store.set("editValue", nil)
						end,
						autoFocus = true,
					}
				end,
			}
		]])
		app:keyPress("Enter")
		-- The input should display "Alice"
		test.assert.eq(app:screenContains("Alice"), true)
	end)

	test.it("editable false by default (backward compat)", function()
		app = test.createApp(60, 20)
		app:loadString([[
			local lux = require("lux")
			local DataGrid = lux.DataGrid
			lumina.app {
				id = "compat-test",
				store = { idx = 1, activated = "" },
				render = function()
					local idx = lumina.useStore("idx")
					local activated = lumina.useStore("activated")
					return DataGrid {
						id = "grid",
						width = 50, height = 10,
						columns = { { id = "x", header = "X", width = 20, key = "x" } },
						rows = { { x = "hello" } },
						selectedIndex = idx,
						onActivate = function(i, row)
							lumina.store.set("activated", row.x)
						end,
						autoFocus = true,
					}
				end,
			}
		]])
		app:keyPress("Enter")
		-- Should not enter edit mode (editable defaults to false)
		local node = app:find("edit-1-x")
		test.assert.isNil(node)
	end)

	test.it("editableColumns restricts which columns are editable", function()
		app = test.createApp(60, 20)
		app:loadString([[
			local lux = require("lux")
			local DataGrid = lux.DataGrid
			lumina.app {
				id = "restrict-test",
				store = { idx = 1, editCell = nil, editValue = nil },
				render = function()
					local idx = lumina.useStore("idx")
					local editCell = lumina.useStore("editCell")
					local editValue = lumina.useStore("editValue")
					return DataGrid {
						id = "grid",
						width = 50, height = 12,
						columns = {
							{ id = "id", header = "ID", width = 10, key = "id" },
							{ id = "name", header = "Name", width = 20, key = "name" },
						},
						rows = { { id = "1", name = "Alice" } },
						selectedIndex = idx,
						editable = true,
						editingCell = editCell,
						editValue = editValue,
						editableColumns = { name = true },
						onEditStart = function(row, col)
							lumina.store.set("editCell", { rowIndex = row, columnId = col })
							lumina.store.set("editValue", nil)
						end,
						onEditValueChange = function(text)
							lumina.store.set("editValue", text)
						end,
						onEditCancel = function(row, col)
							lumina.store.set("editCell", nil)
							lumina.store.set("editValue", nil)
						end,
						autoFocus = true,
					}
				end,
			}
		]])
		app:keyPress("Enter")
		-- Should start editing "name" column (skipping "id")
		local node = app:find("edit-1-name")
		test.assert.notNil(node)
		-- "id" column should NOT have an input
		local idNode = app:find("edit-1-id")
		test.assert.isNil(idNode)
	end)

	test.it("onCellChange fires on Enter with correct value", function()
		app = test.createApp(60, 20)
		app:loadString([[
			local lux = require("lux")
			local DataGrid = lux.DataGrid
			lumina.app {
				id = "change-test",
				store = { idx = 1, editCell = nil, editValue = nil, lastChange = "", dbg = "" },
				render = function()
					local idx = lumina.useStore("idx")
					local editCell = lumina.useStore("editCell")
					local editValue = lumina.useStore("editValue")
					local lastChange = lumina.useStore("lastChange")
					local dbg = lumina.useStore("dbg")
					return lumina.createElement("vbox", {},
						DataGrid {
							id = "grid",
							width = 50, height = 10,
							columns = {
								{ id = "name", header = "Name", width = 20, key = "name" },
							},
							rows = { { name = "Alice" } },
							selectedIndex = idx,
							editable = true,
							editingCell = editCell,
							editValue = editValue,
							onEditStart = function(row, col)
								lumina.store.set("editCell", { rowIndex = row, columnId = col })
								lumina.store.set("editValue", nil)
								lumina.store.set("dbg", "start")
							end,
							onEditValueChange = function(text)
								lumina.store.set("editValue", text)
								lumina.store.set("dbg", "vc:" .. text)
							end,
							onCellChange = function(row, col, val)
								lumina.store.set("lastChange", tostring(val))
								lumina.store.set("editCell", nil)
								lumina.store.set("editValue", nil)
								lumina.store.set("dbg", "cc:" .. tostring(val))
							end,
							onEditCancel = function(row, col)
								lumina.store.set("editCell", nil)
								lumina.store.set("editValue", nil)
							end,
							autoFocus = true,
						},
						lumina.createElement("text", { id = "status" }, "changed:" .. lastChange),
						lumina.createElement("text", { id = "dbg" }, "dbg:" .. dbg)
					)
				end,
			}
		]])
		app:keyPress("Enter")  -- enter edit mode
		-- Type a character (input is focused via useEffect + focusById)
		-- Cursor starts at position 0, so "!" is inserted before "Alice"
		app:keyPress("!")
		-- Verify onEditValueChange was called
		test.assert.eq(app:screenContains("dbg:vc:!Alice"), true)
		-- Confirm with Enter
		app:keyPress("Enter")
		-- onCellChange should receive "!Alice"
		test.assert.eq(app:screenContains("changed:!Alice"), true)
	end)
end)
