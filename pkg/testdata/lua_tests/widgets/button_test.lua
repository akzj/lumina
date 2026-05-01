-- button_test.lua — Tests for the Lux Button component (rendering + interaction)

test.describe("Button widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders with label", function()
        app:loadString([[
            local Button = require("lux.button")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Button { label = "Click Me", key = "btn1" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Click Me"), true)
    end)

    test.it("fires onClick on click", function()
        app:loadString([[
            local Button = require("lux.button")
            lumina.store.set("count", 0)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local count = lumina.useStore("count")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Button {
                            label = "Increment",
                            key = "btn2",
                            onClick = function()
                                lumina.store.set("count", count + 1)
                            end,
                        },
                        lumina.createElement("text", {id = "out"},
                            "count:" .. tostring(count))
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("count:0"), true)

        -- Click the button area
        app:click(3, 1)
        test.assert.eq(app:screenContains("count:1"), true)

        -- Click again
        app:click(3, 1)
        test.assert.eq(app:screenContains("count:2"), true)
    end)

    test.it("disabled button does not fire onClick", function()
        app:loadString([[
            local Button = require("lux.button")
            lumina.store.set("clicked", "no")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local clicked = lumina.useStore("clicked")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Button {
                            label = "Disabled",
                            disabled = true,
                            key = "btn3",
                            onClick = function()
                                lumina.store.set("clicked", "yes")
                            end,
                        },
                        lumina.createElement("text", {id = "out"},
                            "clicked:" .. clicked)
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("clicked:no"), true)

        -- Click should NOT fire
        app:click(3, 1)
        test.assert.eq(app:screenContains("clicked:no"), true)
    end)

    test.it("supports variant prop", function()
        app:loadString([[
            local Button = require("lux.button")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Button { label = "Primary", severity = "primary", key = "bp" },
                        Button { label = "Ghost", appearance = "text", key = "bg" },
                        Button { label = "Secondary", severity = "secondary", key = "bs" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Primary"), true)
        test.assert.eq(app:screenContains("Ghost"), true)
        test.assert.eq(app:screenContains("Secondary"), true)
    end)
end)
