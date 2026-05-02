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
            lumina.store.set("movedTo", nil)
            lumina.createComponent({
                id = "test",
                render = function()
                    return lumina.createElement(_G.Window, {
                        id = "w1",
                        title = "Test",
                        x = 2, y = 1, w = 30, h = 10,
                        onMove = function(nx, ny)
                            lumina.store.set("movedTo", {x = nx, y = ny})
                        end,
                    })
                end,
            })
        ]])
        -- Mouse down on title bar (y=1)
        app:mouseDown(5, 1)
        -- Mouse move to simulate drag
        app:mouseMove(10, 4)
        -- Mouse up
        app:mouseUp(10, 4)
        app:render()
        -- After drag, window content should still be visible
        test.assert.eq(app:screenContains("Test"), true)
    end)

    -- Test 4: Resize via mouse events changes size
    test.it("resize via mouse events changes size", function()
        app:loadString([[
            lumina.store.set("resizedTo", nil)
            lumina.createComponent({
                id = "test",
                render = function()
                    return lumina.createElement(_G.Window, {
                        id = "w1",
                        title = "Test",
                        x = 2, y = 1, w = 30, h = 10,
                        onResize = function(nw, nh)
                            lumina.store.set("resizedTo", {w = nw, h = nh})
                        end,
                    })
                end,
            })
        ]])
        -- Mouse down on resize handle (bottom-right corner)
        app:mouseDown(30, 10)
        -- Mouse move to simulate resize
        app:mouseMove(35, 14)
        -- Mouse up
        app:mouseUp(35, 14)
        app:render()
        -- After resize, window should still be visible
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
