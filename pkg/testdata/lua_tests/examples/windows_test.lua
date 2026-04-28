-- windows_test.lua — Tests for examples/windows.lua (multi-window manager)

test.describe("Window manager example", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadFile("../examples/windows.lua")
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Test 1: All three window titles are visible
    test.it("renders three overlapping windows", function()
        test.assert.eq(app:screenContains("Editor"), true)
        test.assert.eq(app:screenContains("Monitor"), true)
        test.assert.eq(app:screenContains("Palette"), true)
    end)

    -- Test 2: Clicking win3's button increments its counter
    test.it("top window button click works", function()
        -- Win3 (Palette) is on top initially, button at approx (35, 12)
        app:click(35, 12)
        test.assert.eq(app:screenContains("Click Me (1)"), true)
    end)

    -- Test 3: Clicking at win1's button position does NOT trigger it
    -- because win2/win3 overlap that area
    test.it("occluded button click does NOT trigger", function()
        -- Win1 button is at approx y=10, x=10
        -- But win2 (x=15..44, y=5..16) and win3 (x=28..57, y=3..14) overlap
        -- Click at (10, 10) — this IS in win1's button area
        -- but only if win1 is behind win2/win3 at that position
        -- Actually (10, 10) is x=10 which is inside win1 (x=2..31) but
        -- x=10 < 15 so it's NOT inside win2. Check win3: x=10 < 28, not inside win3.
        -- So (10, 10) is only inside win1 — it WILL fire win1's button.
        -- We need a position where win1's button is behind another window.
        -- Win1 button at y=10, x range 3..31. Win2 at x=15..44, y=5..16.
        -- So at x=20, y=10 — win1 button AND win2 body overlap.
        -- Since win2 is on top of win1, clicking (20, 10) should hit win2, not win1.
        app:click(20, 10)
        -- Win1's button should NOT have been clicked (still 0)
        test.assert.eq(app:screenContains("Click Me (1)"), false)
    end)

    -- Test 4: Clicking on win1's visible area brings it to front
    test.it("click brings window to front", function()
        -- Win1 visible area: x=3..14 (before win2 starts at x=15), y=2..12
        -- Click at (8, 8) — inside win1 only
        app:click(8, 8)
        -- Win1 should now be active
        test.assert.eq(app:screenContains("Active: 📝"), true)
        -- Win1's full content should be visible now (it's on top)
        test.assert.eq(app:screenContains("editor window"), true)
    end)

    -- Test 5: After bringing win1 to front, its button works
    test.it("after bring-to-front button works", function()
        -- Bring win1 to front via keyboard
        app:keyPress("1")
        test.assert.eq(app:screenContains("Active: 📝"), true)
        -- Win1 button at y=10, x~10. Now win1 is on top so button is clickable.
        app:click(10, 10)
        test.assert.eq(app:screenContains("Click Me (1)"), true)
    end)

    -- Test 6: Keyboard movement moves the active window
    test.it("keyboard movement moves active window", function()
        -- Select win1
        app:keyPress("1")
        local before = app:screenText()
        -- Move right
        app:keyPress("ArrowRight")
        local after = app:screenText()
        -- The screen should change (win1 moved right by 2)
        test.assert.neq(before, after)
        -- Win1's content should still be visible
        test.assert.eq(app:screenContains("Editor"), true)
    end)
end)
