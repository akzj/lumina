test.describe("Repaint debug", function()
    local app
    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)
    test.afterEach(function() app:destroy() end)

    test.it("simple move without bringToFront", function()
        app:loadString([[
            lumina.store.set("w1", {x = 5, y = 1, w = 20, h = 6})
            lumina.store.set("w2", {x = 15, y = 5, w = 20, h = 6})
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local w1 = lumina.useStore("w1")
                    local w2 = lumina.useStore("w2")
                    return lumina.createElement("box", {
                        style = {width = 80, height = 24, background = "#000000"}},
                        lumina.createElement(lumina.Window, {
                            title = "WIN-A", x = w1.x, y = w1.y,
                            width = w1.w, height = w1.h, key = "w1",
                            onChange = function(e)
                                if type(e) == "table" and e.type == "move" then
                                    lumina.store.set("w1", {x = e.x, y = e.y, w = w1.w, h = w1.h})
                                end
                            end,
                        }),
                        lumina.createElement(lumina.Window, {
                            title = "WIN-B", x = w2.x, y = w2.y,
                            width = w2.w, height = w2.h, key = "w2",
                            onChange = function(e)
                                if type(e) == "table" and e.type == "move" then
                                    lumina.store.set("w2", {x = e.x, y = e.y, w = w2.w, h = w2.h})
                                end
                            end,
                        })
                    )
                end,
            })
        ]])

        -- Drag WIN-A down (away from WIN-B overlap)
        app:mouseDown(15, 2)  -- WIN-A title bar
        app:mouseMove(15, 10) -- move down
        app:mouseUp(15, 10)

        -- Both should be visible
        test.assert.eq(app:screenContains("WIN-A"), true)
        test.assert.eq(app:screenContains("WIN-B"), true)
    end)

    test.it("drag with bringToFront preserves all windows", function()
        app:loadString([[
            local windows = {
                {title = "WIN-A", x = 5, y = 1, w = 20, h = 6, open = true},
                {title = "WIN-B", x = 15, y = 10, w = 20, h = 6, open = true},
            }
            lumina.store.set("windows", windows)

            local function bringToFront(idx)
                local wins = lumina.store.get("windows")
                local w = table.remove(wins, idx)
                wins[#wins + 1] = w
                lumina.store.set("windows", wins)
            end

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local wins = lumina.useStore("windows")
                    local children = {}
                    for i, win in ipairs(wins) do
                        if win.open then
                            local winIdx = i
                            children[#children + 1] = lumina.createElement(lumina.Window, {
                                title = win.title, x = win.x, y = win.y,
                                width = win.w, height = win.h, key = win.title,
                                onChange = function(e)
                                    if type(e) == "table" and e.type == "move" then
                                        bringToFront(winIdx)
                                        local ws = lumina.store.get("windows")
                                        ws[#ws].x = e.x
                                        ws[#ws].y = e.y
                                        lumina.store.set("windows", ws)
                                    end
                                end,
                            })
                        end
                    end
                    return lumina.createElement("box", {
                        style = {width = 80, height = 24}},
                        table.unpack(children))
                end,
            })
        ]])

        -- Drag WIN-A right by 3
        app:mouseDown(10, 2)
        app:mouseMove(13, 2)
        app:mouseUp(13, 2)

        -- Both should be visible (non-overlapping titles)
        test.assert.eq(app:screenContains("WIN-A"), true)
        test.assert.eq(app:screenContains("WIN-B"), true)
    end)
end)
