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

    test.it("drag title bar moves window", function()
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

        -- Mousedown on title bar (border at y=5, title row at y=6)
        app:mouseDown(20, 6)
        -- Drag 5 right, 3 down
        app:mouseMove(25, 9)
        -- Release
        app:mouseUp(25, 9)

        -- Window should still be visible after drag
        test.assert.eq(app:screenContains("Draggable"), true)
    end)

    test.it("resize handle resizes window", function()
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

        -- Mousedown on resize handle (bottom-right corner)
        -- Window at x=5, y=2, w=30, h=10 → bottom-right around (34, 11)
        app:mouseDown(34, 11)
        -- Drag to resize larger
        app:mouseMove(39, 14)
        -- Release
        app:mouseUp(39, 14)

        -- Window should still be visible after resize
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

    test.it("two windows at different y positions render correctly", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("box", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Window, {
                            title = "TopWin",
                            x = 2, y = 1, width = 30, height = 6,
                            key = "w1",
                        }),
                        lumina.createElement(lumina.Window, {
                            title = "BotWin",
                            x = 2, y = 10, width = 30, height = 6,
                            key = "w2",
                        })
                    )
                end,
            })
        ]])
        -- Both windows should be visible at their correct positions
        test.assert.eq(app:screenContains("TopWin"), true)
        test.assert.eq(app:screenContains("BotWin"), true)
        -- Verify BotWin is at y=10 (not double-offset from vbox stacking)
        local screen = app:screenText()
        local lines = {}
        for line in screen:gmatch("[^\n]+") do
            lines[#lines + 1] = line
        end
        -- BotWin title should be on line 12 (y=10 border + 1 for title row + 1 for 1-based)
        -- Actually: border at row 10 (0-based), title at row 11 (0-based) = line 12 (1-based)
        local titleLine = lines[12] or ""
        test.assert.eq(titleLine:find("BotWin") ~= nil, true)
    end)

    test.it("drag clears old position (no ghost artifacts)", function()
        app:loadString([[
            lumina.store.set("win", {x = 5, y = 2, w = 30, h = 8})
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local win = lumina.useStore("win")
                    return lumina.createElement("box", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Window, {
                            title = "Ghost",
                            x = win.x, y = win.y,
                            width = win.w, height = win.h,
                            key = "w1",
                            onChange = function(e)
                                if type(e) == "table" and e.type == "move" then
                                    lumina.store.set("win", {x = e.x, y = e.y, w = 30, h = 8})
                                end
                            end,
                        })
                    )
                end,
            })
        ]])
        -- Window starts at y=2, title at y=3
        test.assert.eq(app:screenContains("Ghost"), true)
        local screen = app:screenText()
        local lines = {}
        for line in screen:gmatch("[^\n]+") do
            lines[#lines + 1] = line
        end
        -- Title "Ghost" should be on line 4 (y=2 border, y=3 title, 1-based = line 4)
        test.assert.eq((lines[4] or ""):find("Ghost") ~= nil, true)

        -- Drag window down by 5 rows: mouseDown on title bar (y=3), move to y=8
        app:mouseDown(15, 3)
        app:mouseMove(15, 8)
        app:mouseUp(15, 8)

        -- After drag, window should be at new position
        test.assert.eq(app:screenContains("Ghost"), true)
        local screen2 = app:screenText()
        local lines2 = {}
        for line in screen2:gmatch("[^\n]+") do
            lines2[#lines2 + 1] = line
        end
        -- Old position (line 4) should be cleared (no "Ghost" text)
        local oldLine = lines2[4] or ""
        test.assert.eq(oldLine:find("Ghost") == nil, true)
    end)
end)
