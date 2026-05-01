-- switch_test.lua — Tests for the Switch widget (lux version, rendering + interaction)

test.describe("Switch widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders on/off states", function()
        app:loadString([[
            local Switch = require("lux.switch")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(Switch, {
                            label = "Dark Mode", checked = false, key = "sw1",
                        }),
                        lumina.createElement(Switch, {
                            label = "Notifications", checked = true, key = "sw2",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Dark Mode"), true)
        test.assert.eq(app:screenContains("Notifications"), true)
    end)

    test.it("toggles on click via onChange", function()
        app:loadString([[
            local Switch = require("lux.switch")
            lumina.store.set("on", false)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local on = lumina.useStore("on")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(Switch, {
                            label = "Toggle",
                            checked = on,
                            key = "sw3",
                            onChange = function(val)
                                lumina.store.set("on", val)
                            end,
                        }),
                        lumina.createElement("text", {id = "status"},
                            "on:" .. tostring(on))
                    )
                end,
            })
        ]])
        -- Initial: off
        test.assert.eq(app:screenContains("on:false"), true)

        -- Click to turn on
        app:click(2, 0)
        test.assert.eq(app:screenContains("on:true"), true)

        -- Click again to turn off
        app:click(2, 0)
        test.assert.eq(app:screenContains("on:false"), true)
    end)

    test.it("disabled switch does not toggle", function()
        app:loadString([[
            local Switch = require("lux.switch")
            lumina.store.set("on", false)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local on = lumina.useStore("on")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(Switch, {
                            label = "Locked",
                            checked = on,
                            disabled = true,
                            key = "sw4",
                            onChange = function(val)
                                lumina.store.set("on", val)
                            end,
                        }),
                        lumina.createElement("text", {id = "status"},
                            "on:" .. tostring(on))
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("on:false"), true)

        -- Click should NOT toggle
        app:click(2, 0)
        test.assert.eq(app:screenContains("on:false"), true)
    end)
end)
