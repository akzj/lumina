-- toast_test.lua — Tests for the Toast widget (rendering + interaction)

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
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Toast, {
                            visible = true,
                            message = "Operation complete",
                            key = "toast1",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Operation complete"), true)
    end)

    test.it("hidden when not visible", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Toast, {
                            visible = false,
                            message = "Hidden toast",
                            key = "toast2",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Hidden toast"), false)
    end)

    test.it("renders variant icons", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Toast, {
                            visible = true,
                            message = "Saved!",
                            variant = "success",
                            key = "toast3",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Saved!"), true)
    end)

    test.it("renders error variant", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Toast, {
                            visible = true,
                            message = "Failed!",
                            variant = "error",
                            key = "toast4",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Failed!"), true)
    end)

    test.it("renders default message when none given", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Toast, {
                            visible = true,
                            key = "toast5",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Notification"), true)
    end)

    test.it("click fires onChange with dismiss", function()
        app:loadString([[
            lumina.store.set("action", "none")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local action = lumina.useStore("action")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Toast, {
                            visible = true,
                            message = "Click to dismiss",
                            key = "toast6",
                            onChange = function(val)
                                lumina.store.set("action", val)
                            end,
                        }),
                        lumina.createElement("text", {id = "out"},
                            "action:" .. action)
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("action:none"), true)

        -- Click the toast
        app:click(5, 0)
        test.assert.eq(app:screenContains("action:dismiss"), true)
    end)

    test.it("show/hide controlled by state", function()
        app:loadString([[
            lumina.store.set("show", false)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local show = lumina.useStore("show")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Toast, {
                            visible = show,
                            message = "Dynamic Toast",
                            key = "toast7",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Dynamic Toast"), false)

        -- Show toast via store
        app:loadString([[lumina.store.set("show", true)]])
        app:render()
        test.assert.eq(app:screenContains("Dynamic Toast"), true)

        -- Hide toast via store
        app:loadString([[lumina.store.set("show", false)]])
        app:render()
        test.assert.eq(app:screenContains("Dynamic Toast"), false)
    end)
end)
