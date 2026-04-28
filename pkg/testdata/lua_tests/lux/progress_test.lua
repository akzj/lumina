-- progress_test.lua — Tests for the lux Progress component

test.describe("lux.Progress", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders progress bar at 0%", function()
        app:loadString([[
            local Progress = require("lux.progress")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Progress { value = 0 }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("0%"), true)
    end)

    test.it("renders progress bar at 50%", function()
        app:loadString([[
            local Progress = require("lux.progress")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Progress { value = 50 }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("50%"), true)
    end)

    test.it("renders progress bar at 100%", function()
        app:loadString([[
            local Progress = require("lux.progress")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Progress { value = 100 }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("100%"), true)
    end)

    test.it("renders with custom width", function()
        app:loadString([[
            local Progress = require("lux.progress")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Progress { value = 75, width = 10 }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("75%"), true)
    end)

    test.it("updates when value changes via store", function()
        app:loadString([[
            local Progress = require("lux.progress")
            lumina.store.set("pct", 20)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local pct = lumina.useStore("pct")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Progress { value = pct }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("20%"), true)

        -- Update store value and re-render
        app:loadString([[lumina.store.set("pct", 80)]])
        app:render()
        test.assert.eq(app:screenContains("80%"), true)
    end)
end)
