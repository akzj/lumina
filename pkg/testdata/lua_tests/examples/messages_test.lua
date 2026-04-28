-- messages_test.lua — Tests for examples/messages.lua (nested scroll)

test.describe("Messages example app", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadFile("../examples/messages.lua")
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Test 1: Initial render shows header, messages, and footer
    test.it("renders initial message list with header and footer", function()
        test.assert.eq(app:screenContains("Messages"), true)
        test.assert.eq(app:screenContains("Alice"), true)
        test.assert.eq(app:screenContains("Quit"), true)
    end)

    -- Test 2: j/k navigation moves selection between messages
    test.it("j/k navigates between messages", function()
        -- Initially first message (Alice) is selected
        test.assert.eq(app:screenContains("Alice"), true)

        app:keyPress("j")
        -- Second message (Bob) should now be selected
        test.assert.eq(app:screenContains("Bob"), true)

        app:keyPress("j")
        -- Third message (Charlie) should be selected
        test.assert.eq(app:screenContains("Charlie"), true)

        app:keyPress("k")
        -- Back to second message
        test.assert.eq(app:screenContains("Bob"), true)
    end)

    -- Test 3: Outer scroll keeps header/footer visible with many messages
    test.it("outer scroll does not cover header or footer", function()
        -- Navigate to the last message
        for i = 1, 6 do
            app:keyPress("j")
        end
        -- Header and footer should still be visible
        test.assert.eq(app:screenContains("Messages"), true)
        test.assert.eq(app:screenContains("Quit"), true)
        -- Last message (Grace) should be visible
        test.assert.eq(app:screenContains("Grace"), true)
    end)

    -- Test 4: Long message content is clipped by inner scroll
    test.it("long message content is clipped within message box", function()
        -- Alice's message is long — the beginning should show
        test.assert.eq(app:screenContains("Hey"), true)
        -- The very end of Alice's long message should be clipped
        -- (it talks about "dashboard widgets" which is far down)
        test.assert.eq(app:screenContains("dashboard"), false)
    end)

    -- Test 5: Navigate to bottom and back — header stays
    test.it("navigate to bottom and back preserves header", function()
        -- Go to last
        for i = 1, 6 do
            app:keyPress("j")
        end
        test.assert.eq(app:screenContains("Messages"), true)
        -- Go back to first
        for i = 1, 6 do
            app:keyPress("k")
        end
        test.assert.eq(app:screenContains("Messages"), true)
        test.assert.eq(app:screenContains("Alice"), true)
    end)

    -- Test 6: Badge shows correct message count
    test.it("badge shows message count", function()
        test.assert.eq(app:screenContains("7"), true)
    end)

    -- Test 7: Inner scroll reveals hidden content in long messages
    test.it("ArrowDown scrolls inner message content to reveal end", function()
        -- Initially "scrolling" (end of Alice's message) is NOT visible
        test.assert.eq(app:screenContains("scrolling"), false)
        -- Scroll inner content down several times
        app:keyPress("ArrowDown")
        app:keyPress("ArrowDown")
        app:keyPress("ArrowDown")
        -- Now the end of Alice's message should be visible
        test.assert.eq(app:screenContains("scrolling"), true)
    end)
end)
