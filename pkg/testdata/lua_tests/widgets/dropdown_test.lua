-- dropdown_test.lua — Tests for the Dropdown widget (rendering + interaction)

test.describe("Dropdown widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders trigger label", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Dropdown, {
                            label = "Actions",
                            items = {
                                {label = "Edit"},
                                {label = "Delete"},
                            },
                            key = "dd1",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Actions"), true)
    end)

    test.it("renders default label when none given", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Dropdown, {
                            items = {{label = "Item"}},
                            key = "dd2",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Menu"), true)
    end)

    test.it("renders with divider items", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Dropdown, {
                            label = "Options",
                            items = {
                                {label = "Copy"},
                                {divider = true},
                                {label = "Paste"},
                            },
                            key = "dd3",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Options"), true)
    end)

    test.it("renders with icon items", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Dropdown, {
                            label = "File",
                            items = {
                                {label = "New", icon = "+"},
                                {label = "Open", icon = "O"},
                            },
                            key = "dd4",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("File"), true)
    end)
end)
