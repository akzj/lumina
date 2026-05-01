-- dialog_test.lua — Tests for the Lux Dialog component (rendering + interaction)

test.describe("Dialog widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders when open", function()
        app:loadString([[
            local Dialog = require("lux.dialog")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Dialog {
                            open = true,
                            title = "Confirm",
                            message = "Are you sure?",
                            width = 40,
                            key = "dlg1",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Confirm"), true)
        test.assert.eq(app:screenContains("Are you sure?"), true)
    end)

    test.it("hidden when not open", function()
        app:loadString([[
            local Dialog = require("lux.dialog")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Dialog {
                            open = false,
                            title = "Hidden",
                            message = "Should not appear",
                            key = "dlg2",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Hidden"), false)
        test.assert.eq(app:screenContains("Should not appear"), false)
    end)

    test.it("renders with default title", function()
        app:loadString([[
            local Dialog = require("lux.dialog")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Dialog {
                            open = true,
                            message = "Content here",
                            width = 40,
                            key = "dlg3",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Dialog"), true)
        test.assert.eq(app:screenContains("Content here"), true)
    end)

    test.it("renders with children replacing message", function()
        app:loadString([[
            local Dialog = require("lux.dialog")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Dialog {
                            open = true,
                            title = "Custom",
                            width = 40,
                            key = "dlg4",
                            Dialog.Content { "Custom Content" },
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Custom Content"), true)
    end)

    test.it("open/close controlled by state", function()
        app:loadString([[
            local Dialog = require("lux.dialog")
            lumina.store.set("open", true)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local open = lumina.useStore("open")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Dialog {
                            open = open,
                            title = "Toggle Dialog",
                            message = "Visible when open",
                            width = 40,
                            key = "dlg5",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Toggle Dialog"), true)

        -- Close via store
        app:loadString([[lumina.store.set("open", false)]])
        app:render()
        test.assert.eq(app:screenContains("Toggle Dialog"), false)

        -- Re-open via store
        app:loadString([[lumina.store.set("open", true)]])
        app:render()
        test.assert.eq(app:screenContains("Toggle Dialog"), true)
    end)
end)
