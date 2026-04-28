-- menu_test.lua — Tests for the Menu widget (rendering + interaction)

test.describe("Menu widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders menu items", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Menu, {
                            items = {
                                {label = "Home"},
                                {label = "Settings"},
                                {label = "Help"},
                            },
                            key = "menu1",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Home"), true)
        test.assert.eq(app:screenContains("Settings"), true)
        test.assert.eq(app:screenContains("Help"), true)
    end)

    test.it("renders with icons", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Menu, {
                            items = {
                                {label = "Dashboard", icon = "D"},
                                {label = "Profile", icon = "P"},
                            },
                            key = "menu2",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Dashboard"), true)
        test.assert.eq(app:screenContains("Profile"), true)
    end)

    test.it("renders divider items", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Menu, {
                            items = {
                                {label = "Item1"},
                                {divider = true},
                                {label = "Item2"},
                            },
                            key = "menu3",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Item1"), true)
        test.assert.eq(app:screenContains("Item2"), true)
        test.assert.eq(app:screenContains("─"), true)
    end)

    test.it("renders disabled items", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Menu, {
                            items = {
                                {label = "Active"},
                                {label = "Disabled", disabled = true},
                            },
                            key = "menu4",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Active"), true)
        test.assert.eq(app:screenContains("Disabled"), true)
    end)

    test.it("renders compact mode without border", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Menu, {
                            items = {
                                {label = "Compact Item"},
                            },
                            compact = true,
                            key = "menu5",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Compact Item"), true)
    end)

    test.it("controlled selected index", function()
        app:loadString([[
            lumina.store.set("sel", 1)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local sel = lumina.useStore("sel")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Menu, {
                            items = {
                                {label = "First"},
                                {label = "Second"},
                                {label = "Third"},
                            },
                            selected = sel,
                            key = "menu6",
                        }),
                        lumina.createElement("text", {id = "out"},
                            "sel:" .. tostring(sel))
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("sel:1"), true)

        -- Change selection via store
        app:loadString([[lumina.store.set("sel", 2)]])
        app:render()
        test.assert.eq(app:screenContains("sel:2"), true)
    end)
end)
