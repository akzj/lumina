test.describe("Accordion component", function()
    local app

    test.beforeEach(function()
        app = test.createApp(60, 24)
        app:loadFile("../examples/accordion_demo.lua")
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Rendering
    test.it("renders all item titles", function()
        test.assert.eq(app:screenContains("What is Lumina"), true)
        test.assert.eq(app:screenContains("How to install"), true)
        test.assert.eq(app:screenContains("Disabled Section"), true)
        test.assert.eq(app:screenContains("License"), true)
    end)

    test.it("first item is open by default", function()
        test.assert.eq(app:screenContains("TUI framework"), true)
    end)

    test.it("closed items do not show content", function()
        test.assert.eq(app:screenContains("go install"), false)
        test.assert.eq(app:screenContains("MIT License"), false)
    end)

    -- Keyboard navigation
    test.it("j moves to next item", function()
        app:keyPress("j")
        -- Now on item 2, press Enter to open
        app:keyPress("Enter")
        test.assert.eq(app:screenContains("go install"), true)
    end)

    test.it("k moves to previous item", function()
        app:keyPress("j")  -- item 2
        app:keyPress("k")  -- back to item 1
        app:keyPress("Enter")  -- toggle item 1 (close it)
        test.assert.eq(app:screenContains("TUI framework"), false)
    end)

    test.it("Enter toggles open/close", function()
        -- Item 1 is open
        test.assert.eq(app:screenContains("TUI framework"), true)
        app:keyPress("Enter")  -- close it
        test.assert.eq(app:screenContains("TUI framework"), false)
        app:keyPress("Enter")  -- reopen
        test.assert.eq(app:screenContains("TUI framework"), true)
    end)

    test.it("Space also toggles", function()
        app:keyPress(" ")  -- close item 1
        test.assert.eq(app:screenContains("TUI framework"), false)
    end)

    test.it("single mode: opening one closes others", function()
        -- Item 1 is open
        test.assert.eq(app:screenContains("TUI framework"), true)
        app:keyPress("j")      -- move to item 2
        app:keyPress("Enter")  -- open item 2
        test.assert.eq(app:screenContains("go install"), true)
        -- Item 1 should be closed in single mode
        test.assert.eq(app:screenContains("TUI framework"), false)
    end)

    test.it("skips disabled item when navigating", function()
        app:keyPress("j")  -- item 2
        app:keyPress("j")  -- should skip disabled, go to item 4
        app:keyPress("Enter")
        test.assert.eq(app:screenContains("MIT License"), true)
    end)

    test.it("wraps around navigation", function()
        app:keyPress("k")  -- wrap to last non-disabled (item 4)
        app:keyPress("Enter")
        test.assert.eq(app:screenContains("MIT License"), true)
    end)

    -- Multi mode
    test.it("multi mode allows multiple open", function()
        app:keyPress("m")  -- switch to multi mode
        test.assert.eq(app:screenContains("multi"), true)
        -- Open item 1
        app:keyPress("Enter")
        test.assert.eq(app:screenContains("TUI framework"), true)
        -- Open item 2
        app:keyPress("j")
        app:keyPress("Enter")
        test.assert.eq(app:screenContains("go install"), true)
        -- Both should be visible
        test.assert.eq(app:screenContains("TUI framework"), true)
    end)

    -- Edge cases
    test.it("no crash with empty items", function()
        app:destroy()
        app = test.createApp(60, 10)
        app:loadString([[
            local Accordion = require("lux.accordion")
            lumina.app {
                id = "empty",
                render = function()
                    return Accordion { id = "a", items = {}, width = 40 }
                end,
            }
        ]])
        -- Should not crash - just renders empty
        test.assert.notNil(app:screenText())
    end)

    test.it("disabled item cannot be toggled", function()
        -- Force selectedIdx to 3 via store
        app:loadString('lumina.store.set("selectedIdx", 3)')
        app:keyPress("Enter")
        -- Disabled content should NOT appear
        test.assert.eq(app:screenContains("Cannot open"), false)
    end)
end)
