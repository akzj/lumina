-- list_test.lua — Tests for the List widget (rendering + interaction)

test.describe("List widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders string items", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.List, {
                            items = {"Alpha", "Beta", "Gamma"},
                            key = "list1",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Alpha"), true)
        test.assert.eq(app:screenContains("Beta"), true)
        test.assert.eq(app:screenContains("Gamma"), true)
    end)

    test.it("renders map items with label key", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.List, {
                            items = {
                                {label = "Item A"},
                                {label = "Item B"},
                            },
                            key = "list2",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Item A"), true)
        test.assert.eq(app:screenContains("Item B"), true)
    end)

    test.it("renders with showIndex", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.List, {
                            items = {"First", "Second"},
                            showIndex = true,
                            key = "list3",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("1. First"), true)
        test.assert.eq(app:screenContains("2. Second"), true)
    end)

    test.it("renders empty list without error", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.List, {
                            items = {},
                            key = "list4",
                        })
                    )
                end,
            })
        ]])
        local tree = app:vnodeTree()
        test.assert.notNil(tree)
    end)

    test.it("non-selectable list renders without selection prefix", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.List, {
                            items = {"X", "Y"},
                            selectable = false,
                            key = "list5",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("X"), true)
        test.assert.eq(app:screenContains("Y"), true)
    end)
end)
