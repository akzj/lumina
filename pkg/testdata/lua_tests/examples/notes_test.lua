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
        -- Detail view shows note content
        test.assert.eq(app:screenContains("TUI framework"), true)
        app:keyPress("Escape")
        -- After escape, list view is restored: multiple note titles visible
        test.assert.eq(app:screenContains("Notes"), true)
        test.assert.eq(app:screenContains("Syntax Sugar"), true)
        test.assert.eq(app:screenContains("Lux Components"), true)
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

    test.it("many notes do not overflow over header/footer", function()
        -- Add 10 notes (total 13 with 3 initial = 26 lines of NotePreview)
        for i = 1, 10 do
            app:keyPress("n")
        end
        -- Navigate to the last note (selectedIdx is already 13 after adding)
        -- Header and footer should still be visible despite overflow
        test.assert.eq(app:screenContains("Notes"), true)
        test.assert.eq(app:screenContains("Navigate"), true)
        -- Navigate up to first note
        for i = 1, 12 do
            app:keyPress("k")
        end
        -- Header and footer should still be visible
        test.assert.eq(app:screenContains("Notes"), true)
        test.assert.eq(app:screenContains("Navigate"), true)
        -- First note should be visible after scrolling up
        test.assert.eq(app:screenContains("Welcome to Lumina"), true)
    end)
end)
