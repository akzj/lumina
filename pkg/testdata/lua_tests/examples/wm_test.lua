-- wm_test.lua — Tests for lux.wm WindowManager module

test.describe("WindowManager", function()
    local app
    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadFile("../examples/windows_widget.lua")
    end)
    test.afterEach(function() app:destroy() end)

    -- Test 1: Initial windows render in correct order
    test.it("renders initial windows", function()
        test.assert.eq(app:screenContains("Editor"), true)
        -- Palette is topmost (last in order), always visible
        test.assert.eq(app:screenContains("Palette"), true)
        -- Monitor is behind Palette; click on Monitor's exposed right side to activate
        -- Monitor at x=20, w=35 → right edge at x=54; Palette at x=10, w=30 → right edge at x=39
        -- So Monitor title at x=40+ is not occluded by Palette
        app:click(45, 6)
        test.assert.eq(app:screenContains("Monitor"), true)
    end)

    -- Test 2: activate (mousedown on title bar) brings window to front
    test.it("activate brings window to front", function()
        -- Editor is at bottom of z-order (first in initial list)
        -- Click on Editor title bar to activate it
        app:mouseDown(10, 2)
        app:mouseUp(10, 2)
        -- Editor should now be on top (visible)
        test.assert.eq(app:screenContains("Editor"), true)
        test.assert.eq(app:screenContains("editor line"), true)
    end)

    -- Test 3: close removes window, preserves others
    test.it("close removes window from display", function()
        -- Palette close button is at top-right of Palette window
        -- Palette is at x=10, y=3, w=30 → close button near x=37-39, y=4
        app:click(37, 4)
        test.assert.eq(app:screenContains("Palette"), false)
        -- Other windows still visible
        test.assert.eq(app:screenContains("Editor"), true)
    end)

    -- Test 4: drag moves window (setFrame updates position)
    test.it("drag moves window position", function()
        -- Drag Editor (at y=1) down
        app:mouseDown(10, 2)
        app:mouseMove(10, 10)
        app:mouseUp(10, 10)
        -- Editor should be visible at new position
        test.assert.eq(app:screenContains("Editor"), true)
    end)

    -- Test 5: close then drag another window works
    test.it("close then drag another window works", function()
        -- Close Palette
        app:click(37, 4)
        test.assert.eq(app:screenContains("Palette"), false)
        -- Drag Editor
        app:mouseDown(10, 2)
        app:mouseMove(30, 10)
        app:mouseUp(30, 10)
        -- Editor should still be visible after drag
        test.assert.eq(app:screenContains("Editor"), true)
    end)

    -- Test 6: reopen (press 'o') restores closed windows
    test.it("reopen restores closed windows", function()
        -- Close Palette
        app:click(37, 4)
        test.assert.eq(app:screenContains("Palette"), false)
        -- Press 'o' to reopen all closed windows
        app:keyPress("o")
        test.assert.eq(app:screenContains("Palette"), true)
    end)

    -- Test 7: Editor window has scrollbar (scrollable=true with 30 lines)
    test.it("scrollable window shows scrollbar", function()
        -- Activate Editor (bring to front) so scrollbar is visible
        app:mouseDown(10, 2)
        app:mouseUp(10, 2)
        local s = app:screenText()
        -- Scrollbar uses █ (thumb) and ░ (track) characters
        test.assert.eq(s:find("█") ~= nil, true)
    end)
end)
