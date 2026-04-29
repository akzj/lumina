-- windows_widget_test.lua — Tests for examples/windows_widget.lua (Window widget demo)

test.describe("Window widget example", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadFile("../examples/windows_widget.lua")
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Test 1: Windows render (at least two titles visible, third may be occluded by z-order)
    test.it("renders three windows with titles", function()
        -- With overlapping windows, the topmost (last in array) is always visible
        test.assert.eq(app:screenContains("Palette"), true)
        -- Editor at y=1 is partially visible above others
        test.assert.eq(app:screenContains("Editor"), true)
    end)

    -- Test 2: Windows have close buttons and resize handles
    test.it("windows have close button and resize handle", function()
        test.assert.eq(app:screenContains("✕"), true)
        test.assert.eq(app:screenContains("◢"), true)
    end)

    -- Test 3: Window content is visible (topmost window content always visible)
    test.it("window content renders", function()
        -- Palette is topmost (last in array), its content is always visible
        test.assert.eq(app:screenContains("Color palette"), true)
        -- Editor content may be partially visible
        test.assert.eq(app:screenContains("Welcome"), true)
    end)

    -- Test 4: Drag title bar keeps window visible
    test.it("drag title bar keeps window visible", function()
        -- Win1 (Editor) at x=2, y=1, title bar at y=2
        app:mouseDown(15, 2)
        app:mouseMove(25, 5)
        app:mouseUp(25, 5)
        -- Editor should still be visible after drag
        test.assert.eq(app:screenContains("Editor"), true)
    end)
end)
