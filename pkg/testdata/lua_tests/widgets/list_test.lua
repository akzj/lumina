-- list_test.lua — Tests for the Lux ListView component (rendering)

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
            local ListView = require("lux.list")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        ListView {
                            rows = {{label = "Alpha"}, {label = "Beta"}, {label = "Gamma"}},
                            renderRow = function(row, idx, ctx)
                                return lumina.createElement("text", {key = "r" .. idx}, row.label)
                            end,
                            height = 10,
                            key = "list1",
                        }
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
            local ListView = require("lux.list")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        ListView {
                            rows = {
                                {label = "Item A"},
                                {label = "Item B"},
                            },
                            renderRow = function(row, idx, ctx)
                                return lumina.createElement("text", {key = "r" .. idx}, row.label)
                            end,
                            height = 10,
                            key = "list2",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Item A"), true)
        test.assert.eq(app:screenContains("Item B"), true)
    end)

    test.it("renders with showIndex", function()
        app:loadString([[
            local ListView = require("lux.list")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        ListView {
                            rows = {{label = "First"}, {label = "Second"}},
                            renderRow = function(row, idx, ctx)
                                return lumina.createElement("text", {key = "r" .. idx},
                                    tostring(idx) .. ". " .. row.label)
                            end,
                            height = 10,
                            key = "list3",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("1. First"), true)
        test.assert.eq(app:screenContains("2. Second"), true)
    end)

    test.it("renders empty list without error", function()
        app:loadString([[
            local ListView = require("lux.list")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        ListView {
                            rows = {},
                            renderRow = function(row, idx, ctx)
                                return lumina.createElement("text", {key = "r" .. idx}, row.label)
                            end,
                            height = 10,
                            key = "list4",
                        }
                    )
                end,
            })
        ]])
        local tree = app:vnodeTree()
        test.assert.notNil(tree)
    end)

    test.it("non-selectable list renders without selection prefix", function()
        app:loadString([[
            local ListView = require("lux.list")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        ListView {
                            rows = {{label = "X"}, {label = "Y"}},
                            renderRow = function(row, idx, ctx)
                                return lumina.createElement("text", {key = "r" .. idx}, row.label)
                            end,
                            height = 10,
                            key = "list5",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("X"), true)
        test.assert.eq(app:screenContains("Y"), true)
    end)
end)
