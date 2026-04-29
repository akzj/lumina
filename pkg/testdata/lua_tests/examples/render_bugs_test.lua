test.describe("Render bugs", function()
    local app
    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadFile("../examples/windows_widget.lua")
    end)
    test.afterEach(function() app:destroy() end)

    test.it("initial render - check content", function()
        test.log("=== INITIAL RENDER ===")
        test.log(app:screenText())
    end)

    test.it("resize window to minimum", function()
        -- Resize Editor by dragging bottom-right corner (◢) upward-left
        -- Editor is at x=2, y=1, w=35, h=12
        -- Bottom-right corner ◢ is at x=36, y=12
        app:mouseDown(36, 12)
        app:mouseMove(12, 5)
        app:mouseUp(12, 5)
        test.log("=== AFTER RESIZE TO SMALL ===")
        test.log(app:screenText())
    end)
end)
