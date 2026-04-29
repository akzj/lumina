test.describe("Repaint debug", function()
    local app
    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadFile("../examples/windows_widget.lua")
    end)
    test.afterEach(function() app:destroy() end)

    test.it("simple move without bringToFront", function()
        -- Modify example to NOT call bringToFront during drag
        -- Actually, let's just test with a simpler setup
        app:loadString([[
            lumina.store.set("w1", {x = 5, y = 1, w = 30, h = 8})
            lumina.store.set("w2", {x = 15, y = 5, w = 30, h = 8})
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
        
        test.log("INITIAL (2 windows):")
        test.log(app:screenText())
        
        -- Drag WIN-A down (away from WIN-B overlap)
        app:mouseDown(15, 2)  -- WIN-A title bar
        app:mouseMove(15, 10) -- move down
        app:mouseUp(15, 10)
        test.log("After WIN-A moved down:")
        test.log(app:screenText())
        
        -- Both should be visible
        test.assert.eq(app:screenContains("WIN-A"), true)
        test.assert.eq(app:screenContains("WIN-B"), true)
    end)
end)
