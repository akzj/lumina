-- Verify windows.lua drag/resize by dumping screen text

test.describe("Windows Drag/Resize Verification", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadFile("../examples/windows.lua")
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("initial screen shows three windows", function()
        local screen = app:screenText()
        test.log("=== INITIAL SCREEN ===")
        test.log(screen)
        test.assert.eq(app:screenContains("Editor"), true)
        test.assert.eq(app:screenContains("Monitor"), true)
        test.assert.eq(app:screenContains("Palette"), true)
    end)

    test.it("drag win1: click title bar + move + release", function()
        -- win1 is at x=2, y=1. Title bar is at y+1=2 (inside border)
        -- Click on title bar at (5, 2)
        test.log("=== DRAG TEST: clicking title bar at (5,2) ===")
        app:mouseDown(5, 2)
        
        -- Move mouse to simulate drag (dx=8, dy=3)
        test.log("=== Moving to (13, 5) ===")
        app:mouseMove(13, 5)
        
        -- Release
        app:mouseUp(13, 5)
        app:render()
        
        local screen = app:screenText()
        test.log("=== AFTER DRAG ===")
        test.log(screen)
        
        -- Window should still be visible
        test.assert.eq(app:screenContains("Editor"), true)
    end)

    test.it("resize win1: click bottom-right + move + release", function()
        -- win1 at x=2, y=1, w=30, h=12
        -- Resize handle at bottom-right: (x+w-2, y+h-2) = (30, 10)
        test.log("=== RESIZE TEST: clicking resize at (30,10) ===")
        app:mouseDown(30, 10)
        
        -- Move to expand
        test.log("=== Moving to (38, 15) ===")
        app:mouseMove(38, 15)
        
        -- Release
        app:mouseUp(38, 15)
        app:render()
        
        local screen = app:screenText()
        test.log("=== AFTER RESIZE ===")
        test.log(screen)
        
        test.assert.eq(app:screenContains("Editor"), true)
    end)
end)
