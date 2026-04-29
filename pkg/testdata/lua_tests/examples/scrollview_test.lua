test.describe("ScrollView", function()
    local app
    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadFile("../examples/scrollview.lua")
    end)
    test.afterEach(function() app:destroy() end)

    test.it("renders scrollview with content", function()
        local s = app:screenText()
        test.assert.eq(s:find("ScrollView Demo") ~= nil, true)
        -- First few lines should be visible
        test.assert.eq(s:find("Line 1") ~= nil, true)
        test.assert.eq(s:find("Line 2") ~= nil, true)
        -- Line 50 should NOT be visible (scrolled off)
        test.assert.eq(s:find("Line 50") == nil, true)
    end)

    test.it("scrolls down with keyboard", function()
        -- Click inside ScrollView to focus it
        app:click(10, 5)
        -- Press Down multiple times to scroll
        for i = 1, 10 do
            app:keyPress("ArrowDown")
        end
        local s = app:screenText()
        -- After scrolling 10 lines down, Line 1 should no longer be visible
        -- and later lines should appear
        test.assert.eq(s:find("Line 11") ~= nil or s:find("Line 12") ~= nil, true)
    end)

    test.it("scrolls to bottom with End", function()
        -- Click inside ScrollView to focus it
        app:click(10, 5)
        app:keyPress("End")
        local s = app:screenText()
        -- Line 50 should now be visible
        test.assert.eq(s:find("Line 50") ~= nil, true)
        -- Line 1 should NOT be visible
        test.assert.eq(s:find("Line 1 ") == nil, true)
    end)

    test.it("scrolls to top with Home after End", function()
        -- Click inside ScrollView to focus it
        app:click(10, 5)
        app:keyPress("End")
        app:keyPress("Home")
        local s = app:screenText()
        -- Line 1 should be visible again
        test.assert.eq(s:find("Line 1") ~= nil, true)
    end)

    test.it("displays scrollbar", function()
        local s = app:screenText()
        -- Scrollbar uses █ (thumb) and ░ (track) characters
        test.assert.eq(s:find("█") ~= nil, true)
        test.assert.eq(s:find("░") ~= nil, true)
    end)
end)
