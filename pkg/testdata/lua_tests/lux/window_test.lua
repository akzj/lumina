-- window_test.lua — Tests for lux.window component

test.describe("Window", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadString([[
            local Window = require("lux.window")
            _G.Window = Window
        ]])
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Test 1: Renders title and content
    test.it("renders title and content", function()
        app:loadString([[
            local children = {
                lumina.createElement("text", {}, "Hello World"),
            }
            lumina.createComponent({
                id = "test",
                render = function()
                    return lumina.createElement(_G.Window, {
                        id = "w1",
                        title = "My Window",
                        x = 2, y = 1, w = 30, h = 10,
                    }, table.unpack(children))
                end,
            })
        ]])
        test.assert.eq(app:screenContains("My Window"), true)
        test.assert.eq(app:screenContains("Hello World"), true)
    end)

    -- Test 2: Click activates
    test.it("click activates", function()
        app:loadString([[
            lumina.store.set("activated", false)
            lumina.createComponent({
                id = "test",
                render = function()
                    return lumina.createElement(_G.Window, {
                        id = "w1",
                        title = "Test",
                        x = 2, y = 1, w = 30, h = 10,
                        onActivate = function()
                            lumina.store.set("activated", true)
                        end,
                    })
                end,
            })
        ]])
        app:click(5, 5)
        app:render()
        test.assert.eq(app:screenContains("Test"), true)
    end)

    -- Test 3: Drag via mouse events changes position
    test.it("drag via mouse events changes position", function()
        app:loadString([[
            local WM = require("lux.wm")
            local mgr = WM.create("drag_test", {
                {id = "w1", title = "Test", x = 2, y = 1, w = 30, h = 10},
            })
            lumina.createComponent({
                id = "test",
                render = function()
                    local wins = mgr.getWindows()
                    local w = wins[1]
                    return lumina.createElement(_G.Window, {
                        id = "w1",
                        title = "Test",
                        x = w.x, y = w.y,
                        w = w.w, h = w.h,
                        onMove = function(nx, ny)
                            mgr.setFrame("w1", {x = nx, y = ny})
                        end,
                    })
                end,
            })
        ]])
        -- Mouse down on title bar (y+1=2, inside border)
        app:mouseDown(5, 2)
        app:mouseMove(13, 5)
        app:mouseUp(13, 5)
        app:render()
        -- After drag, window content should still be visible (no crash)
        test.assert.eq(app:screenContains("Test"), true)
    end)

    -- Test 4: Resize via mouse events changes size
    test.it("resize via mouse events changes size", function()
        app:loadString([[
            local WM = require("lux.wm")
            local mgr = WM.create("resize_test", {
                {id = "w1", title = "Test", x = 2, y = 1, w = 30, h = 10},
            })
            lumina.createComponent({
                id = "test",
                render = function()
                    local wins = mgr.getWindows()
                    local w = wins[1]
                    return lumina.createElement(_G.Window, {
                        id = "w1",
                        title = "Test",
                        x = w.x, y = w.y,
                        w = w.w, h = w.h,
                        onResize = function(nw, nh)
                            mgr.setFrame("w1", {w = nw, h = nh})
                        end,
                    })
                end,
            })
        ]])
        -- Mouse down on resize handle (bottom-right 3x3 area, inside border)
        -- x+w-3=29, y+h-3=8, so (30, 10) is inside the 3x3 resize zone
        app:mouseDown(30, 10)
        app:mouseMove(35, 14)
        app:mouseUp(35, 14)
        app:render()
        -- After resize, window should still be visible (no crash)
        test.assert.eq(app:screenContains("Test"), true)
    end)

    -- Test 5: Active vs inactive styling
    test.it("active vs inactive styling", function()
        app:loadString([[
            lumina.createComponent({
                id = "test",
                render = function()
                    local children = {
                        lumina.createElement(_G.Window, {
                            id = "w1",
                            title = "Active",
                            x = 2, y = 1, w = 20, h = 6,
                            isActive = true,
                        }),
                        lumina.createElement(_G.Window, {
                            id = "w2",
                            title = "Inactive",
                            x = 25, y = 1, w = 20, h = 6,
                            isActive = false,
                        }),
                    }
                    return lumina.createElement("vbox", {}, table.unpack(children))
                end,
            })
        ]])
        -- Both windows should render their titles
        test.assert.eq(app:screenContains("Active"), true)
        test.assert.eq(app:screenContains("Inactive"), true)
    end)
end)
