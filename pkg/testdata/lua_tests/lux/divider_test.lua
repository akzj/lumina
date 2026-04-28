-- divider_test.lua — Tests for the lux Divider component

test.describe("lux.Divider", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders horizontal line with default char", function()
        app:loadString([[
            local Divider = require("lux.divider")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Divider {}
                    )
                end,
            })
        ]])
        -- Default char is "─", default width is 40
        test.assert.eq(app:screenContains("─"), true)
    end)

    test.it("renders with custom char", function()
        app:loadString([[
            local Divider = require("lux.divider")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Divider { char = "=", width = 20 }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("===================="), true)
    end)

    test.it("renders with custom width", function()
        app:loadString([[
            local Divider = require("lux.divider")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Divider { char = "-", width = 10 }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("----------"), true)
    end)

    test.it("via lux umbrella module", function()
        app:loadString([[
            local lux = require("lux")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lux.Divider { char = "*", width = 5 }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("*****"), true)
    end)
end)
