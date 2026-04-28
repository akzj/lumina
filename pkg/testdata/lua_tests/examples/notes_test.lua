-- notes_test.lua — Smoke + interaction tests for examples/notes.lua

test.describe("Notes example app", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadFile("../examples/notes.lua")
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders initial notes list with 3 notes", function()
        test.assert.eq(app:screenContains("Notes"), true)
        test.assert.eq(app:screenContains("Welcome to Lumina"), true)
        test.assert.eq(app:screenContains("Syntax Sugar"), true)
        test.assert.eq(app:screenContains("Lux Components"), true)
        test.assert.eq(app:screenContains("3"), true)
    end)

    test.it("j/k navigates notes", function()
        -- Initially first note selected
        test.assert.eq(app:screenContains("Welcome to Lumina"), true)

        app:keyPress("j")
        -- Selection moved to second note
        test.assert.eq(app:screenContains("Syntax Sugar"), true)

        app:keyPress("j")
        -- Selection moved to third note
        test.assert.eq(app:screenContains("Lux Components"), true)

        app:keyPress("k")
        -- Back to second note
        test.assert.eq(app:screenContains("Syntax Sugar"), true)
    end)

    test.it("Enter shows detail view", function()
        app:keyPress("Enter")
        -- Should show the first note's content in a Card
        test.assert.eq(app:screenContains("TUI framework"), true)
    end)

    test.it("Escape returns to list from detail view", function()
        app:keyPress("Enter")
        test.assert.eq(app:screenContains("TUI framework"), true)
        app:keyPress("Escape")
        -- Detail content should be gone, header still present
        test.assert.eq(app:screenContains("TUI framework"), false)
        test.assert.eq(app:screenContains("Notes"), true)
    end)

    test.it("n creates a new note", function()
        app:keyPress("n")
        test.assert.eq(app:screenContains("Note #4"), true)
        test.assert.eq(app:screenContains("4"), true)
    end)

    test.it("d deletes the selected note", function()
        app:keyPress("d")
        -- First note deleted, should no longer appear
        test.assert.eq(app:screenContains("Welcome to Lumina"), false)
        test.assert.eq(app:screenContains("Syntax Sugar"), true)
        test.assert.eq(app:screenContains("2"), true)
    end)

    test.it("shows footer keybindings", function()
        test.assert.eq(app:screenContains("Navigate"), true)
        test.assert.eq(app:screenContains("Quit"), true)
    end)
end)
