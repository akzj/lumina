-- engine/events_mouse_test.lua — Tests for mouse event handling

test.describe("Engine: Mouse Events", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("onClick fires on click", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local count, setCount = lumina.useState("count", 0)
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="btn",
                        style={width=10, height=1},
                        onClick=function() setCount(count + 1) end},
                        "Click:" .. tostring(count)
                    )
                )
            end})
        ]])
        test.assert.eq(app:screenContains("Click:0"), true)
        app:click(2, 0)
        test.assert.eq(app:screenContains("Click:1"), true)
    end)

    test.it("onMouseEnter fires on hover", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local hovered, setHovered = lumina.useState("hover", false)
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="target",
                        style={width=20, height=3},
                        onMouseEnter=function() setHovered(true) end,
                        onMouseLeave=function() setHovered(false) end},
                        hovered and "HOVERED" or "NORMAL"
                    )
                )
            end})
        ]])
        test.assert.eq(app:screenContains("NORMAL"), true)
        app:mouseMove(5, 1)
        test.assert.eq(app:screenContains("HOVERED"), true)
    end)

    test.it("click hits correct nested child", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local clicked, setClicked = lumina.useState("clicked", "none")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="a",
                        style={width=20, height=2},
                        onClick=function() setClicked("A") end}, "AAA"),
                    lumina.createElement("text", {id="b",
                        style={width=20, height=2},
                        onClick=function() setClicked("B") end}, "BBB"),
                    lumina.createElement("text", {id="result"}, "Clicked:" .. clicked)
                )
            end})
        ]])
        test.assert.eq(app:screenContains("Clicked:none"), true)
        -- Click on B (at y=2, which is the second child's row)
        app:click(5, 2)
        test.assert.eq(app:screenContains("Clicked:B"), true)
    end)

    test.it("click on child does not fire parent onClick when child handles it", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local who, setWho = lumina.useState("who", "none")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24},
                    onClick=function() setWho("parent") end},
                    lumina.createElement("text", {id="child",
                        style={width=20, height=3},
                        onClick=function() setWho("child") end}, "ChildArea"),
                    lumina.createElement("text", {id="result"}, "Who:" .. who)
                )
            end})
        ]])
        -- Click on child area
        app:click(5, 1)
        test.assert.eq(app:screenContains("Who:child"), true)
    end)
end)
