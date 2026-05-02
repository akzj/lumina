test.describe("WM Store Reactivity Verification", function()
    local app

    test.it("setFrame triggers re-render (pointer equality bug check)", function()
        app = test.createApp(80, 24)
        
        -- Create WM + component in one loadString
        app:loadString([[
            local WM = require("lux.wm")
            _G.mgr = WM.create("test_wm", {
                {id = "a", title = "A", x = 2, y = 1, w = 30, h = 12},
            })
            
            lumina.createComponent({
                id = "test-comp",
                render = function()
                    local wins = _G.mgr.getWindows()
                    local w = wins[1]
                    return lumina.createElement("text", {},
                        "POS:" .. w.x .. "," .. w.y)
                end,
            })
        ]])
        
        local before = app:screenText()
        test.log("=== BEFORE setFrame ===")
        test.log(before)
        
        -- Now call setFrame via loadString (simulating what happens during drag)
        app:loadString([[
            _G.mgr.setFrame("a", {x = 15, y = 8})
        ]])
        
        app:render()
        
        local after = app:screenText()
        test.log("=== AFTER setFrame ===")
        test.log(after)
        
        -- Verify position changed
        test.assert.eq(app:screenContains("POS:15,8"), true)
    end)
end)
