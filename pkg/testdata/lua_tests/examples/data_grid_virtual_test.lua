test.describe("DataGrid virtual scrolling", function()
    local app

    test.beforeEach(function()
        app = test.createApp(60, 20)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders only visible rows when virtualScroll enabled", function()
        app:loadString([[
            local lux = require("lux")
            local DataGrid = lux.DataGrid
            local rows = {}
            for i = 1, 100 do
                rows[i] = { name = "item-" .. string.format("%03d", i), val = i }
            end
            lumina.app {
                id = "vtest",
                store = { idx = 1 },
                render = function()
                    return DataGrid {
                        id = "grid",
                        width = 50,
                        height = 12,
                        columns = {
                            { id = "name", header = "Name", width = 20, key = "name" },
                            { id = "val", header = "Val", width = 10, key = "val" },
                        },
                        rows = rows,
                        selectedIndex = lumina.useStore("idx"),
                        onChangeIndex = function(i) lumina.store.set("idx", i) end,
                        virtualScroll = true,
                        autoFocus = true,
                    }
                end,
            }
        ]])
        -- Should see item-001 (first row) but NOT item-100 (last row)
        test.assert.eq(app:screenContains("item-001"), true)
        test.assert.eq(app:screenContains("item-100"), false)
    end)

    test.it("scrolling reveals later rows", function()
        app:loadString([[
            local lux = require("lux")
            local DataGrid = lux.DataGrid
            local rows = {}
            for i = 1, 100 do
                rows[i] = { name = "item-" .. string.format("%03d", i), val = i }
            end
            lumina.app {
                id = "vtest2",
                store = { idx = 50 },
                render = function()
                    return DataGrid {
                        id = "grid",
                        width = 50,
                        height = 12,
                        columns = {
                            { id = "name", header = "Name", width = 20, key = "name" },
                            { id = "val", header = "Val", width = 10, key = "val" },
                        },
                        rows = rows,
                        selectedIndex = lumina.useStore("idx"),
                        onChangeIndex = function(i) lumina.store.set("idx", i) end,
                        virtualScroll = true,
                        autoFocus = true,
                    }
                end,
            }
        ]])
        -- Should see item-050 (selected), NOT item-001
        test.assert.eq(app:screenContains("item-050"), true)
        test.assert.eq(app:screenContains("item-001"), false)
    end)

    test.it("navigation with j/k works in virtual mode", function()
        app:loadString([[
            local lux = require("lux")
            local DataGrid = lux.DataGrid
            local rows = {}
            for i = 1, 100 do
                rows[i] = { name = "item-" .. string.format("%03d", i), val = i }
            end
            lumina.app {
                id = "vtest3",
                store = { idx = 1 },
                render = function()
                    return DataGrid {
                        id = "grid",
                        width = 50,
                        height = 12,
                        columns = {
                            { id = "name", header = "Name", width = 20, key = "name" },
                            { id = "val", header = "Val", width = 10, key = "val" },
                        },
                        rows = rows,
                        selectedIndex = lumina.useStore("idx"),
                        onChangeIndex = function(i) lumina.store.set("idx", i) end,
                        virtualScroll = true,
                        autoFocus = true,
                    }
                end,
            }
        ]])
        -- Navigate down many times
        for i = 1, 20 do
            app:keyPress("j")
        end
        -- Should now show item-021 area
        test.assert.eq(app:screenContains("item-021"), true)
    end)

    test.it("PageDown works in virtual mode", function()
        app:loadString([[
            local lux = require("lux")
            local DataGrid = lux.DataGrid
            local rows = {}
            for i = 1, 100 do
                rows[i] = { name = "item-" .. string.format("%03d", i), val = i }
            end
            lumina.app {
                id = "vtest4",
                store = { idx = 1 },
                render = function()
                    return DataGrid {
                        id = "grid",
                        width = 50,
                        height = 12,
                        columns = {
                            { id = "name", header = "Name", width = 20, key = "name" },
                            { id = "val", header = "Val", width = 10, key = "val" },
                        },
                        rows = rows,
                        selectedIndex = lumina.useStore("idx"),
                        onChangeIndex = function(i) lumina.store.set("idx", i) end,
                        virtualScroll = true,
                        autoFocus = true,
                    }
                end,
            }
        ]])
        app:keyPress("PageDown")
        -- Should have jumped ahead
        test.assert.eq(app:screenContains("item-001"), false)
    end)

    test.it("End key jumps to last row in virtual mode", function()
        app:loadString([[
            local lux = require("lux")
            local DataGrid = lux.DataGrid
            local rows = {}
            for i = 1, 100 do
                rows[i] = { name = "item-" .. string.format("%03d", i), val = i }
            end
            lumina.app {
                id = "vtest5",
                store = { idx = 1 },
                render = function()
                    return DataGrid {
                        id = "grid",
                        width = 50,
                        height = 12,
                        columns = {
                            { id = "name", header = "Name", width = 20, key = "name" },
                            { id = "val", header = "Val", width = 10, key = "val" },
                        },
                        rows = rows,
                        selectedIndex = lumina.useStore("idx"),
                        onChangeIndex = function(i) lumina.store.set("idx", i) end,
                        virtualScroll = true,
                        autoFocus = true,
                    }
                end,
            }
        ]])
        app:keyPress("End")
        test.assert.eq(app:screenContains("item-100"), true)
    end)

    test.it("multi-select works with virtual scroll", function()
        app:loadString([[
            local lux = require("lux")
            local DataGrid = lux.DataGrid
            local rows = {}
            for i = 1, 50 do
                rows[i] = { name = "item-" .. string.format("%03d", i), val = i }
            end
            lumina.app {
                id = "vtest6",
                store = { idx = 1, ids = {} },
                render = function()
                    return DataGrid {
                        id = "grid",
                        width = 50,
                        height = 12,
                        columns = {
                            { id = "name", header = "Name", width = 20, key = "name" },
                            { id = "val", header = "Val", width = 10, key = "val" },
                        },
                        rows = rows,
                        selectedIndex = lumina.useStore("idx"),
                        onChangeIndex = function(i) lumina.store.set("idx", i) end,
                        selectionMode = "multi",
                        selectedIds = lumina.useStore("ids"),
                        onSelectionChange = function(ids) lumina.store.set("ids", ids) end,
                        getRowId = function(row, i) return row.name end,
                        virtualScroll = true,
                        autoFocus = true,
                    }
                end,
            }
        ]])
        app:keyPress(" ")  -- select item-001
        test.assert.eq(app:screenContains("\xe2\x97\x8f"), true)
    end)

    test.it("non-virtual mode still works (backward compat)", function()
        app:loadString([[
            local lux = require("lux")
            local DataGrid = lux.DataGrid
            local rows = {}
            for i = 1, 20 do
                rows[i] = { name = "row-" .. tostring(i), val = i }
            end
            lumina.app {
                id = "vtest7",
                store = { idx = 1 },
                render = function()
                    return DataGrid {
                        id = "grid",
                        width = 50,
                        height = 12,
                        columns = {
                            { id = "name", header = "Name", width = 20, key = "name" },
                        },
                        rows = rows,
                        selectedIndex = lumina.useStore("idx"),
                        onChangeIndex = function(i) lumina.store.set("idx", i) end,
                        autoFocus = true,
                        -- virtualScroll NOT set (default false)
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("row-1"), true)
    end)

    test.it("1000 rows with virtual scroll does not crash", function()
        app:loadString([[
            local lux = require("lux")
            local DataGrid = lux.DataGrid
            local rows = {}
            for i = 1, 1000 do
                rows[i] = { name = "r" .. tostring(i), val = i }
            end
            lumina.app {
                id = "vtest8",
                store = { idx = 500 },
                render = function()
                    return DataGrid {
                        id = "grid",
                        width = 50,
                        height = 12,
                        columns = {
                            { id = "name", header = "Name", width = 20, key = "name" },
                            { id = "val", header = "Val", width = 10, key = "val" },
                        },
                        rows = rows,
                        selectedIndex = lumina.useStore("idx"),
                        onChangeIndex = function(i) lumina.store.set("idx", i) end,
                        virtualScroll = true,
                        autoFocus = true,
                    }
                end,
            }
        ]])
        -- Should show row around 500
        test.assert.eq(app:screenContains("r500"), true)
        test.assert.eq(app:screenContains("r1 "), false)
    end)
end)
