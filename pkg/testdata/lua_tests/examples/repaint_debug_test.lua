test.describe("Repaint debug", function()
    local app
    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)
    test.afterEach(function() app:destroy() end)

    test.it("simple move without bringToFront", function()
        app:loadString([[
            lumina.store.set("w1", {x = 5, y = 1})
            lumina.store.set("w2", {x = 15, y = 10})
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local w1 = lumina.useStore("w1")
                    local w2 = lumina.useStore("w2")
                    return lumina.createElement("box", {
                        style = {width = 80, height = 24, background = "#000000"}},
                        lumina.createElement("vbox", {
                            id = "win-a",
                            style = {
                                position = "absolute",
                                left = w1.x, top = w1.y,
                                width = 20, height = 6,
                                background = "#333333",
                            },
                        },
                            lumina.createElement("text", {}, "WIN-A")
                        ),
                        lumina.createElement("vbox", {
                            id = "win-b",
                            style = {
                                position = "absolute",
                                left = w2.x, top = w2.y,
                                width = 20, height = 6,
                                background = "#333333",
                            },
                        },
                            lumina.createElement("text", {}, "WIN-B")
                        )
                    )
                end,
            })
        ]])

        -- Both should be visible
        test.assert.eq(app:screenContains("WIN-A"), true)
        test.assert.eq(app:screenContains("WIN-B"), true)
    end)

    test.it("multiple positioned boxes render correctly", function()
        app:loadString([[
            local windows = {
                {title = "WIN-A", x = 5, y = 1},
                {title = "WIN-B", x = 15, y = 10},
            }
            lumina.store.set("windows", windows)

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local wins = lumina.useStore("windows")
                    local children = {}
                    for i, win in ipairs(wins) do
                        children[#children + 1] = lumina.createElement("vbox", {
                            key = win.title,
                            style = {
                                position = "absolute",
                                left = win.x, top = win.y,
                                width = 20, height = 6,
                                background = "#333333",
                            },
                        },
                            lumina.createElement("text", {}, win.title)
                        )
                    end
                    return lumina.createElement("box", {
                        style = {width = 80, height = 24}},
                        table.unpack(children))
                end,
            })
        ]])

        -- Both should be visible
        test.assert.eq(app:screenContains("WIN-A"), true)
        test.assert.eq(app:screenContains("WIN-B"), true)
    end)
end)
