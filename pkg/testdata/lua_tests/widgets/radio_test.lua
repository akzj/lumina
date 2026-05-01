-- radio_test.lua — Tests for the Radio widget (lux version, rendering + interaction)

test.describe("Radio widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders checked and unchecked radios", function()
        app:loadString([[
            local Radio = require("lux.radio")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(Radio, {
                            label = "Red", value = "red", checked = true, key = "r1",
                        }),
                        lumina.createElement(Radio, {
                            label = "Blue", value = "blue", checked = false, key = "r2",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Red"), true)
        test.assert.eq(app:screenContains("Blue"), true)
        test.assert.eq(app:screenContains("( )"), true)
    end)

    test.it("renders radio group", function()
        app:loadString([[
            local Radio = require("lux.radio")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(Radio, {
                            label = "Small", value = "s", checked = false, key = "r3",
                        }),
                        lumina.createElement(Radio, {
                            label = "Medium", value = "m", checked = true, key = "r4",
                        }),
                        lumina.createElement(Radio, {
                            label = "Large", value = "l", checked = false, key = "r5",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Small"), true)
        test.assert.eq(app:screenContains("Medium"), true)
        test.assert.eq(app:screenContains("Large"), true)
    end)

    test.it("renders disabled radio", function()
        app:loadString([[
            local Radio = require("lux.radio")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(Radio, {
                            label = "Disabled", value = "d", disabled = true, key = "r6",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Disabled"), true)
    end)

    test.it("controlled selection via store", function()
        app:loadString([[
            local Radio = require("lux.radio")
            lumina.store.set("selected", "a")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local selected = lumina.useStore("selected")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(Radio, {
                            label = "Option A", value = "a",
                            checked = (selected == "a"),
                            key = "ra",
                        }),
                        lumina.createElement(Radio, {
                            label = "Option B", value = "b",
                            checked = (selected == "b"),
                            key = "rb",
                        }),
                        lumina.createElement("text", {id = "out"},
                            "selected:" .. selected)
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("selected:a"), true)

        -- Change selection via store
        app:loadString([[lumina.store.set("selected", "b")]])
        app:render()
        test.assert.eq(app:screenContains("selected:b"), true)
    end)
end)
