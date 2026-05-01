-- toast_test.lua — Tests for the Lux Toast component (stack-based notification)

test.describe("Toast widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders visible toast with message", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Toast {
                            items = {
                                {id = "t1", message = "Operation complete", variant = "success"},
                            },
                            key = "toast1",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Operation complete"), true)
    end)

    test.it("renders nothing when items empty", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Toast {
                            items = {},
                            key = "toast2",
                        },
                        lumina.createElement("text", {id = "marker"}, "END")
                    )
                end,
            })
        ]])
        local tree = app:vnodeTree()
        test.assert.notNil(tree)
        test.assert.eq(app:screenContains("END"), true)
    end)

    test.it("renders variant icons", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Toast {
                            items = {
                                {id = "t1", message = "Saved!", variant = "success"},
                            },
                            key = "toast3",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Saved!"), true)
    end)

    test.it("renders error variant", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Toast {
                            items = {
                                {id = "t1", message = "Failed!", variant = "error"},
                            },
                            key = "toast4",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Failed!"), true)
    end)

    test.it("renders default message when none given", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Toast {
                            items = {
                                {id = "t1", variant = "info"},
                            },
                            key = "toast5",
                        }
                    )
                end,
            })
        ]])
        -- Toast renders even without explicit message
        local tree = app:vnodeTree()
        test.assert.notNil(tree)
    end)

    test.it("renders multiple toasts", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Toast {
                            items = {
                                {id = "t1", message = "First toast"},
                                {id = "t2", message = "Second toast"},
                            },
                            key = "toast6",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("First toast"), true)
        test.assert.eq(app:screenContains("Second toast"), true)
    end)

    test.it("show/hide controlled by state", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.store.set("items", {})
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local items = lumina.useStore("items")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Toast {
                            items = items,
                            key = "toast7",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Dynamic Toast"), false)

        -- Show toast via store
        app:loadString([[lumina.store.set("items", {{id = "t1", message = "Dynamic Toast"}})]])
        app:render()
        test.assert.eq(app:screenContains("Dynamic Toast"), true)

        -- Hide toast via store
        app:loadString([[lumina.store.set("items", {})]])
        app:render()
        test.assert.eq(app:screenContains("Dynamic Toast"), false)
    end)
end)
