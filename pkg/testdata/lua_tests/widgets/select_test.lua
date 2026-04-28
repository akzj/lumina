-- select_test.lua — Tests for the Select widget (rendering + interaction)

test.describe("Select widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders with placeholder", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Select, {
                            placeholder = "Choose...",
                            options = {
                                {label = "Apple", value = "apple"},
                                {label = "Banana", value = "banana"},
                            },
                            key = "sel1",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Choose..."), true)
    end)

    test.it("renders default placeholder", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Select, {
                            options = {{label = "One", value = "1"}},
                            key = "sel2",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Select..."), true)
    end)

    test.it("renders selected value", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Select, {
                            value = "banana",
                            options = {
                                {label = "Apple", value = "apple"},
                                {label = "Banana", value = "banana"},
                            },
                            key = "sel3",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Banana"), true)
    end)

    test.it("renders disabled state", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Select, {
                            placeholder = "Disabled",
                            disabled = true,
                            options = {},
                            key = "sel4",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Disabled"), true)
    end)

    test.it("controlled value updates via store", function()
        app:loadString([[
            lumina.store.set("val", "apple")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local val = lumina.useStore("val")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Select, {
                            value = val,
                            options = {
                                {label = "Apple", value = "apple"},
                                {label = "Banana", value = "banana"},
                                {label = "Cherry", value = "cherry"},
                            },
                            key = "sel5",
                        }),
                        lumina.createElement("text", {id = "out"},
                            "val:" .. val)
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Apple"), true)
        test.assert.eq(app:screenContains("val:apple"), true)

        -- Change value via store
        app:loadString([[lumina.store.set("val", "cherry")]])
        app:render()
        test.assert.eq(app:screenContains("Cherry"), true)
        test.assert.eq(app:screenContains("val:cherry"), true)
    end)
end)
