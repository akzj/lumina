-- card_test.lua — Tests for the lux Card component

test.describe("lux.Card", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders card with title", function()
        app:loadString([[
            local Card = require("lux.card")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Card { title = "My Card" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("My Card"), true)
    end)

    test.it("renders card with children", function()
        app:loadString([[
            local Card = require("lux.card")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Card {
                            title = "Info",
                            lumina.createElement("text", {}, "Card body content")
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Info"), true)
        test.assert.eq(app:screenContains("Card body content"), true)
    end)

    test.it("renders card without title", function()
        app:loadString([[
            local Card = require("lux.card")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Card {
                            lumina.createElement("text", {}, "Just content")
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Just content"), true)
    end)

    test.it("via lux umbrella module", function()
        app:loadString([[
            local lux = require("lux")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lux.Card { title = "Umbrella Card" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Umbrella Card"), true)
    end)
end)
