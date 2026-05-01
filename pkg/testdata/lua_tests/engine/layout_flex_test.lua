-- engine/layout_flex_test.lua — Tests for flex layout primitives

test.describe("Engine: Flex Layout", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("vbox arranges children vertically", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=20, height=10}},
                    lumina.createElement("text", {id="a"}, "AAA"),
                    lumina.createElement("text", {id="b"}, "BBB")
                )
            end})
        ]])
        local a = app:find("a")
        local b = app:find("b")
        test.assert.eq(a.y, 0)
        test.assert.eq(b.y, 1)  -- b is below a
        test.assert.eq(a.x, b.x)  -- same x (vertical stack)
    end)

    test.it("hbox arranges children horizontally", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("hbox", {id="root", style={width=20, height=5}},
                    lumina.createElement("text", {id="a"}, "AAA"),
                    lumina.createElement("text", {id="b"}, "BBB")
                )
            end})
        ]])
        local a = app:find("a")
        local b = app:find("b")
        test.assert.eq(a.y, b.y)  -- same row
        test.assert.gt(b.x, a.x)  -- b is to the right of a
    end)

    test.it("flex-grow distributes remaining space", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=20, height=10}},
                    lumina.createElement("text", {id="fixed", style={height=2}}, "Fixed"),
                    lumina.createElement("vbox", {id="grow", style={flexGrow=1}})
                )
            end})
        ]])
        local fixed = app:find("fixed")
        local grow = app:find("grow")
        test.assert.eq(fixed.h, 2)
        test.assert.eq(grow.h, 8)  -- remaining space (10 - 2)
    end)

    test.it("width/height sets fixed dimensions", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="box", style={width=15, height=7}})
                )
            end})
        ]])
        local box = app:find("box")
        test.assert.eq(box.w, 15)
        test.assert.eq(box.h, 7)
    end)

    test.it("padding adds internal space", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="padded", style={width=20, height=10, padding=2}},
                        lumina.createElement("text", {id="child"}, "X")
                    )
                )
            end})
        ]])
        local padded = app:find("padded")
        local child = app:find("child")
        -- Child should be offset by padding from parent
        test.assert.eq(child.x - padded.x, 2)
        test.assert.eq(child.y - padded.y, 2)
    end)

    test.it("gap adds space between children", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=20, height=10, gap=2}},
                    lumina.createElement("text", {id="a", style={height=1}}, "A"),
                    lumina.createElement("text", {id="b", style={height=1}}, "B")
                )
            end})
        ]])
        local a = app:find("a")
        local b = app:find("b")
        test.assert.eq(b.y - a.y - a.h, 2)  -- gap between a and b
    end)
end)
