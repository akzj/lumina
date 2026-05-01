-- engine/layer_test.lua — Tests for layer management and z-order

test.describe("Engine: Layers", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("absolute positioned element renders at specified coordinates", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("box", {id="root",
                    style={width=80, height=24, background="#000000"}},
                    lumina.createElement("text", {id="normal"}, "Normal"),
                    lumina.createElement("vbox", {id="abs",
                        style={position="absolute", left=20, top=10, width=15, height=3}},
                        lumina.createElement("text", {}, "Floating")
                    )
                )
            end})
        ]])
        local abs = app:find("abs")
        test.assert.notNil(abs)
        test.assert.eq(abs.x, 20)
        test.assert.eq(abs.y, 10)
        test.assert.eq(app:screenContains("Floating"), true)
    end)

    test.it("later sibling renders on top of earlier sibling (z-order)", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("box", {id="root",
                    style={width=80, height=24}},
                    -- First: background box at 0,0
                    lumina.createElement("vbox", {id="back",
                        style={position="absolute", left=0, top=0, width=20, height=3,
                               background="#111111"}},
                        lumina.createElement("text", {}, "BACK")
                    ),
                    -- Second: overlapping box at same position (renders on top)
                    lumina.createElement("vbox", {id="front",
                        style={position="absolute", left=0, top=0, width=20, height=3,
                               background="#222222"}},
                        lumina.createElement("text", {}, "FRONT")
                    )
                )
            end})
        ]])
        -- FRONT should be visible (it's on top)
        test.assert.eq(app:screenContains("FRONT"), true)
    end)

    test.it("absolute element does not affect flow layout of siblings", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="a"}, "FlowA"),
                    lumina.createElement("vbox", {
                        style={position="absolute", left=50, top=0, width=10, height=5}},
                        lumina.createElement("text", {}, "ABS")
                    ),
                    lumina.createElement("text", {id="b"}, "FlowB")
                )
            end})
        ]])
        local a = app:find("a")
        local b = app:find("b")
        test.assert.notNil(a)
        test.assert.notNil(b)
        -- FlowB should be directly below FlowA (absolute element doesn't take space)
        test.assert.eq(b.y, a.y + a.h)
    end)
end)
