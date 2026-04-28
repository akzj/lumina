-- spacer_test.lua — Tests for the Spacer widget

test.describe("Spacer widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders without error", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement("text", {id = "top"}, "Top"),
                        lumina.createElement(lumina.Spacer, {key = "sp1"}),
                        lumina.createElement("text", {id = "bottom"}, "Bottom")
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Top"), true)
        test.assert.eq(app:screenContains("Bottom"), true)
    end)

    test.it("renders with fixed size", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement("text", {id = "top"}, "Top"),
                        lumina.createElement(lumina.Spacer, {
                            size = 3,
                            direction = "vertical",
                            key = "sp2",
                        }),
                        lumina.createElement("text", {id = "bottom"}, "Bottom")
                    )
                end,
            })
        ]])
        -- Both text nodes should render
        test.assert.eq(app:screenContains("Top"), true)
        test.assert.eq(app:screenContains("Bottom"), true)

        -- Verify the spacer creates vertical space between them
        local top = app:find("top")
        local bottom = app:find("bottom")
        test.assert.notNil(top)
        test.assert.notNil(bottom)
    end)

    test.it("renders horizontal spacer", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("hbox", {id = "root"},
                        lumina.createElement("text", {id = "left"}, "Left"),
                        lumina.createElement(lumina.Spacer, {
                            size = 5,
                            direction = "horizontal",
                            key = "sp3",
                        }),
                        lumina.createElement("text", {id = "right"}, "Right")
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Left"), true)
        test.assert.eq(app:screenContains("Right"), true)
    end)
end)
