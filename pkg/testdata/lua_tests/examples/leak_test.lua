test.describe("Text leak test", function()
    local app
    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadFile("../examples/windows_widget.lua")
    end)
    test.afterEach(function() app:destroy() end)

    test.it("resize Editor very small - text leaks?", function()
        -- Editor at x=2,y=1,w=35,h=12
        -- ◢ is at bottom-right: x=2+35-2=35, y=1+12-2=11
        -- Drag ◢ to make window very small
        app:mouseDown(35, 11)
        app:mouseMove(12, 5)
        app:mouseUp(12, 5)
        test.log("=== After resize Editor to minimum ===")
        test.log(app:screenText())
    end)

    test.it("resize Monitor very small", function()
        -- Monitor at x=20,y=5,w=35,h=12
        -- ◢ at x=20+35-2=53, y=5+12-2=15
        app:mouseDown(53, 15)
        app:mouseMove(30, 9)
        app:mouseUp(30, 9)
        test.log("=== After resize Monitor to minimum ===")
        test.log(app:screenText())
    end)

    test.it("resize Palette very small", function()
        -- Palette at x=10,y=3,w=30,h=10
        -- ◢ at x=10+30-2=38, y=3+10-2=11
        app:mouseDown(38, 11)
        app:mouseMove(20, 7)
        app:mouseUp(20, 7)
        test.log("=== After resize Palette to minimum ===")
        test.log(app:screenText())
    end)
end)
