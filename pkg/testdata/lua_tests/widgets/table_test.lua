-- table_test.lua — Tests for the Table widget (rendering + interaction)

test.describe("Table widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders columns and rows", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Table, {
                            columns = {
                                {header = "Name", key = "name", width = 15},
                                {header = "Age", key = "age", width = 5},
                            },
                            rows = {
                                {name = "Alice", age = 30},
                                {name = "Bob", age = 25},
                            },
                            key = "tbl1",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Name"), true)
        test.assert.eq(app:screenContains("Age"), true)
        test.assert.eq(app:screenContains("Alice"), true)
        test.assert.eq(app:screenContains("Bob"), true)
    end)

    test.it("renders separator line", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Table, {
                            columns = {{header = "Col", key = "col", width = 10}},
                            rows = {{col = "val"}},
                            key = "tbl2",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("─"), true)
    end)

    test.it("renders empty table with header only", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Table, {
                            columns = {{header = "X", key = "x", width = 10}},
                            rows = {},
                            key = "tbl3",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("X"), true)
    end)

    test.it("renders with selectable prop", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Table, {
                            columns = {{header = "Item", key = "item", width = 15}},
                            rows = {
                                {item = "Row1"},
                                {item = "Row2"},
                            },
                            selectable = true,
                            key = "tbl4",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Row1"), true)
        test.assert.eq(app:screenContains("Row2"), true)
    end)

    test.it("renders striped rows", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Table, {
                            columns = {{header = "Data", key = "d", width = 10}},
                            rows = {
                                {d = "A"},
                                {d = "B"},
                                {d = "C"},
                            },
                            striped = true,
                            key = "tbl5",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("A"), true)
        test.assert.eq(app:screenContains("B"), true)
        test.assert.eq(app:screenContains("C"), true)
    end)
end)
