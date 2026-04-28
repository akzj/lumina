-- dialog_test.lua — Tests for the Dialog widget (rendering + interaction)

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
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Dialog, {
                            open = true,
                            title = "Confirm",
                            message = "Are you sure?",
                            width = 40,
                            key = "dlg1",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Confirm"), true)
        test.assert.eq(app:screenContains("Are you sure?"), true)
    end)

    test.it("hidden when not open", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Dialog, {
                            open = false,
                            title = "Hidden",
                            message = "Should not appear",
                            key = "dlg2",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Hidden"), false)
        test.assert.eq(app:screenContains("Should not appear"), false)
    end)

    test.it("renders with default title", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Dialog, {
                            open = true,
                            message = "Content here",
                            width = 40,
                            key = "dlg3",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Dialog"), true)
        test.assert.eq(app:screenContains("Content here"), true)
    end)

    test.it("renders with children replacing message", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Dialog, {
                            open = true,
                            title = "Custom",
                            message = "Should NOT appear",
                            width = 40,
                            key = "dlg4",
                        },
                            lumina.createElement("text", {}, "Custom Content")
                        )
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Custom Content"), true)
        test.assert.eq(app:screenContains("Should NOT appear"), false)
    end)

    test.it("open/close controlled by state", function()
        app:loadString([[
            lumina.store.set("open", true)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local open = lumina.useStore("open")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Dialog, {
                            open = open,
                            title = "Toggle Dialog",
                            message = "Visible when open",
                            width = 40,
                            key = "dlg5",
                        })
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
