-- notes_test.lua — Smoke test for examples/notes.lua

test.describe("Notes example app", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders initial notes list", function()
        app:loadFile("../examples/notes.lua")
        test.assert.eq(app:screenContains("Notes"), true)
        test.assert.eq(app:screenContains("Welcome to Lumina"), true)
        test.assert.eq(app:screenContains("Syntax Sugar"), true)
        test.assert.eq(app:screenContains("Lux Components"), true)
    end)

    test.it("shows note count badge", function()
        app:loadFile("../examples/notes.lua")
        -- 3 initial notes
        test.assert.eq(app:screenContains("3"), true)
    end)

    test.it("shows footer keybindings", function()
        app:loadFile("../examples/notes.lua")
        test.assert.eq(app:screenContains("Navigate"), true)
        test.assert.eq(app:screenContains("Quit"), true)
    end)
end)
