-- engine/layout_scroll_test.lua — Tests for overflow scroll behavior

test.describe("Engine: Scroll", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("overflow scroll clips content to container height", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local lines = {}
                for i = 1, 20 do
                    lines[i] = lumina.createElement("text", {key="l"..i}, "Line " .. i)
                end
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="scroll",
                        style={width=40, height=5, overflow="scroll"}},
                        table.unpack(lines)
                    )
                )
            end})
        ]])
        -- Line 1 should be visible (within 5 rows)
        test.assert.eq(app:screenContains("Line 1"), true)
        -- Line 10 should NOT be visible (scrolled off)
        test.assert.eq(app:screenContains("Line 10"), false)
    end)

    test.it("scroll event changes visible content", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local lines = {}
                for i = 1, 30 do
                    lines[i] = lumina.createElement("text", {key="l"..i}, "Row " .. string.format("%02d", i))
                end
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="scroll",
                        style={width=40, height=5, overflow="scroll"}},
                        table.unpack(lines)
                    )
                )
            end})
        ]])
        -- Initially Row 01 visible, Row 20 not
        test.assert.eq(app:screenContains("Row 01"), true)
        test.assert.eq(app:screenContains("Row 20"), false)
        -- Scroll down (delta=5 ticks * 3 lines/tick = 15 lines)
        for i = 1, 5 do
            app:scroll(5, 2, 1)
        end
        -- After scrolling 15 lines, Row 01 should be gone, later rows visible
        test.assert.eq(app:screenContains("Row 01"), false)
    end)

    test.it("scroll container has fixed size regardless of content", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local lines = {}
                for i = 1, 50 do
                    lines[i] = lumina.createElement("text", {key="l"..i}, "Item " .. i)
                end
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="scroll",
                        style={width=40, height=8, overflow="scroll"}},
                        table.unpack(lines)
                    )
                )
            end})
        ]])
        local scroll = app:find("scroll")
        test.assert.notNil(scroll)
        test.assert.eq(scroll.h, 8)  -- stays at 8, not expanded by 50 children
    end)
end)
