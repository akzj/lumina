-- badge_test.lua — Tests for the lux Badge component

test.describe("lux.Badge", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders badge with label", function()
        app:loadString([[
            local Badge = require("lux.badge")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Badge { label = "New" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("New"), true)
    end)

    test.it("renders default variant", function()
        app:loadString([[
            local Badge = require("lux.badge")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Badge { label = "Default", variant = "default" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Default"), true)
    end)

    test.it("renders success variant", function()
        app:loadString([[
            local Badge = require("lux.badge")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Badge { label = "OK", variant = "success" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("OK"), true)
    end)

    test.it("renders warning variant", function()
        app:loadString([[
            local Badge = require("lux.badge")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Badge { label = "Warn", variant = "warning" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Warn"), true)
    end)

    test.it("renders error variant", function()
        app:loadString([[
            local Badge = require("lux.badge")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Badge { label = "Err", variant = "error" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Err"), true)
    end)

    test.it("via lux umbrella module", function()
        app:loadString([[
            local lux = require("lux")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lux.Badge { label = "Umbrella" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Umbrella"), true)
    end)
end)
