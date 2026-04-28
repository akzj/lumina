-- window_test.lua — Tests for the Window widget (rendering + drag/resize/close)

test.describe("Window widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders at correct position with title", function()
        app:loadString([[
            lumina.store.set("win", {x = 5, y = 2, w = 30, h = 10})
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local win = lumina.useStore("win")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Window, {
                            title = "My Window",
                            x = win.x, y = win.y,
                            width = win.w, height = win.h,
                            key = "w1",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("My Window"), true)
        -- Should have close button
        test.assert.eq(app:screenContains("✕"), true)
        -- Should have resize handle
        test.assert.eq(app:screenContains("◢"), true)
    end)

    test.it("drag title bar updates position via onChange", function()
        app:loadString([[
            lumina.store.set("win", {x = 10, y = 5, w = 30, h = 10})
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local win = lumina.useStore("win")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Window, {
                            title = "Draggable",
                            x = win.x, y = win.y,
                            width = win.w, height = win.h,
                            key = "w1",
                            onChange = function(e)
                                if type(e) == "table" and e.type == "move" then
                                    lumina.store.set("win", {x = e.x, y = e.y, w = win.w, h = win.h})
                                end
                            end,
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Draggable"), true)

        -- Mousedown on title bar (y=6 is title row: widget at y=5, border=row5, title=row6)
        app:mouseDown(15, 6)
        -- Drag to new position (move 5 right, 3 down)
        app:mouseMove(20, 9)
        -- Release
        app:mouseUp(20, 9)

        -- Window should still be visible after drag
        test.assert.eq(app:screenContains("Draggable"), true)
    end)

    test.it("resize handle updates dimensions via onChange", function()
        app:loadString([[
            lumina.store.set("win", {x = 5, y = 2, w = 30, h = 10})
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local win = lumina.useStore("win")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Window, {
                            title = "Resizable",
                            x = win.x, y = win.y,
                            width = win.w, height = win.h,
                            key = "w1",
                            onChange = function(e)
                                if type(e) == "table" and e.type == "resize" then
                                    lumina.store.set("win", {x = win.x, y = win.y, w = e.width, h = e.height})
                                end
                            end,
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Resizable"), true)

        -- Mousedown on resize handle (bottom-right corner of window)
        -- Window at x=5, y=2, w=30, h=10 → bottom-right is around (34, 11)
        app:mouseDown(34, 11)
        -- Drag to resize larger
        app:mouseMove(39, 14)
        -- Release
        app:mouseUp(39, 14)

        -- Window should still be visible
        test.assert.eq(app:screenContains("Resizable"), true)
    end)

    test.it("close button fires onChange with close", function()
        app:loadString([[
            lumina.store.set("closed", false)
            lumina.store.set("win", {x = 5, y = 2, w = 30, h = 10})
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local closed = lumina.useStore("closed")
                    local win = lumina.useStore("win")
                    if closed then
                        return lumina.createElement("vbox", {id = "root",
                            style = {width = 80, height = 24}},
                            lumina.createElement("text", {}, "Window Closed")
                        )
                    end
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Window, {
                            title = "Closable",
                            x = win.x, y = win.y,
                            width = win.w, height = win.h,
                            key = "w1",
                            onChange = function(e)
                                if e == "close" then
                                    lumina.store.set("closed", true)
                                end
                            end,
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Closable"), true)
        test.assert.eq(app:screenContains("Window Closed"), false)

        -- Click close button (top-right corner of window)
        -- Window at x=5, y=2, w=30 → close button near x=33, y=3 (title row)
        app:click(33, 3)

        test.assert.eq(app:screenContains("Window Closed"), true)
        test.assert.eq(app:screenContains("Closable"), false)
    end)
end)
