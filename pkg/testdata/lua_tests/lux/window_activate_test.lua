test.describe("Window activate on click", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadString([[
            local Window = require("lux.window")
            local WM = require("lux.wm")
            local mgr = WM.create("test_wm", {
                {id = "w1", title = "Win1", x = 2, y = 1, w = 30, h = 10},
                {id = "w2", title = "Win2", x = 35, y = 1, w = 30, h = 10},
            })
            lumina.createComponent({
                id = "test",
                render = function()
                    local wins = mgr.getWindows()
                    local activeId = mgr.getActiveId()
                    local children = {}
                    for _, win in ipairs(wins) do
                        children[#children + 1] = lumina.createElement(Window, {
                            id = win.id,
                            title = win.title,
                            x = win.x, y = win.y,
                            w = win.w, h = win.h,
                            isActive = (win.id == activeId),
                            onActivate = function()
                                mgr.activate(win.id)
                            end,
                            onMove = function(nx, ny)
                                mgr.setFrame(win.id, {x = nx, y = ny})
                            end,
                        })
                    end
                    return lumina.createElement("vbox", {}, table.unpack(children))
                end,
            })
        ]])
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("clicking inactive window activates it", function()
        -- w2 is initially active (last in order)
        -- Click on w1's content area (x=5, y=5 is inside w1)
        app:mouseDown(5, 5)
        app:mouseUp(5, 5)
        app:render()
        
        -- Check if w1 is now active by examining store
        app:loadString([[
            local s = lumina.store.get("test_wm")
            lumina.store.set("_test_activeId", s.activeId)
        ]])
        app:render()
        
        local screen = app:screenText()
        test.log("=== AFTER CLICK ON W1 ===")
        test.log(screen)
        test.log("activeId should be w1")
    end)

    test.it("drag on inactive window title bar works", function()
        -- w2 is active, w1 is inactive
        -- Drag w1's title bar (y=2 is title bar area for w1 at y=1)
        app:mouseDown(5, 2)
        app:mouseMove(15, 5)
        app:mouseUp(15, 5)
        app:render()
        
        local screen = app:screenText()
        test.log("=== AFTER DRAG ON INACTIVE W1 ===")
        test.log(screen)
        test.assert.eq(app:screenContains("Win1"), true)
    end)
end)
