-- theme_test.lua — Tests for lumina.setTheme() / lumina.getTheme()

test.describe("Theme switching", function()
    local app

    test.beforeEach(function()
        app = test.createApp(50, 10)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("getTheme returns default mocha colors", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local t = lumina.getTheme()
                    return lumina.createElement("text", {
                        id = "root",
                        style = { width = 50, height = 1 },
                    }, t.base or "nil")
                end,
            })
        ]])
        test.assert.eq(app:screenContains("#1E1E2E"), true)
    end)

    test.it("setTheme switches to latte", function()
        app:loadString([[
            lumina.setTheme("latte")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local t = lumina.getTheme()
                    return lumina.createElement("text", {
                        id = "root",
                        style = { width = 50, height = 1 },
                    }, t.base or "nil")
                end,
            })
        ]])
        test.assert.eq(app:screenContains("#EFF1F5"), true)
    end)

    test.it("setTheme with custom table", function()
        app:loadString([[
            lumina.setTheme({ base = "#FF0000", text = "#00FF00", primary = "#0000FF" })
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local t = lumina.getTheme()
                    return lumina.createElement("text", {
                        id = "root",
                        style = { width = 50, height = 1 },
                    }, t.base or "nil")
                end,
            })
        ]])
        test.assert.eq(app:screenContains("#FF0000"), true)
    end)

    test.it("setTheme back to mocha after latte", function()
        app:loadString([[
            lumina.setTheme("latte")
            lumina.setTheme("mocha")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local t = lumina.getTheme()
                    return lumina.createElement("text", {
                        id = "root",
                        style = { width = 50, height = 1 },
                    }, t.base or "nil")
                end,
            })
        ]])
        test.assert.eq(app:screenContains("#1E1E2E"), true)
    end)

    test.it("setTheme nord", function()
        app:loadString([[
            lumina.setTheme("nord")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local t = lumina.getTheme()
                    return lumina.createElement("text", {
                        id = "root",
                        style = { width = 50, height = 1 },
                    }, t.base or "nil")
                end,
            })
        ]])
        test.assert.eq(app:screenContains("#2E3440"), true)
    end)

    test.it("setTheme dracula", function()
        app:loadString([[
            lumina.setTheme("dracula")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local t = lumina.getTheme()
                    return lumina.createElement("text", {
                        id = "root",
                        style = { width = 50, height = 1 },
                    }, t.base or "nil")
                end,
            })
        ]])
        test.assert.eq(app:screenContains("#282A36"), true)
    end)

    test.it("custom theme overrides specific keys", function()
        app:loadString([[
            lumina.setTheme({ base = "#AABBCC", primary = "#112233", text = "#FFFFFF" })
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local t = lumina.getTheme()
                    return lumina.createElement("vbox", {
                        id = "root",
                        style = { width = 50, height = 3 },
                    },
                        lumina.createElement("text", {
                            id = "line1",
                            style = { height = 1 },
                        }, t.primary or "nil"),
                        lumina.createElement("text", {
                            id = "line2",
                            style = { height = 1 },
                        }, t.base or "nil")
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("#112233"), true)
        test.assert.eq(app:screenContains("#AABBCC"), true)
    end)
end)
