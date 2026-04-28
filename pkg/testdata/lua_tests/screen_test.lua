test.describe("screen API", function()
    test.it("screenText returns rendered content", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    return lumina.createElement("text", {}, "Hello Screen")
                end,
            })
        ]])
        local text = app:screenText()
        test.assert.contains(text, "Hello Screen")
        app:destroy()
    end)

    test.it("screenContains returns boolean", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    return lumina.createElement("text", {}, "Visible Text")
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Visible Text"), true)
        test.assert.eq(app:screenContains("Not Here"), false)
        app:destroy()
    end)

    test.it("cellAt returns cell info", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    return lumina.createElement("text", {
                        style = { foreground = "#FF0000" },
                    }, "Red")
                end,
            })
        ]])
        local cell = app:cellAt(0, 0)
        test.assert.notNil(cell)
        test.assert.eq(cell.char, "R")
        test.assert.eq(cell.fg, "#FF0000")
        app:destroy()
    end)

    test.it("cellAt out of bounds returns nil", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    return lumina.createElement("text", {}, "X")
                end,
            })
        ]])
        test.assert.isNil(app:cellAt(-1, 0))
        test.assert.isNil(app:cellAt(0, -1))
        test.assert.isNil(app:cellAt(80, 0))
        test.assert.isNil(app:cellAt(0, 24))
        app:destroy()
    end)
end)
