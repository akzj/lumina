-- engine/layout_constraints_test.lua — Tests for min/max/measure constraints

test.describe("Engine: Layout Constraints", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("minWidth enforces minimum width", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="box",
                        style={width=5, minWidth=15, height=3}})
                )
            end})
        ]])
        local box = app:find("box")
        test.assert.notNil(box)
        -- minWidth should override the smaller width
        test.assert.gt(box.w, 5)
    end)

    test.it("minHeight enforces minimum height", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="box",
                        style={width=10, height=2, minHeight=8}})
                )
            end})
        ]])
        local box = app:find("box")
        test.assert.notNil(box)
        test.assert.gt(box.h, 2)
    end)

    test.it("maxWidth caps width", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="box",
                        style={width=50, maxWidth=20, height=3}})
                )
            end})
        ]])
        local box = app:find("box")
        test.assert.notNil(box)
        test.assert.eq(box.w <= 20, true)
    end)

    test.it("maxHeight caps height", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="box",
                        style={width=10, height=50, maxHeight=10}})
                )
            end})
        ]])
        local box = app:find("box")
        test.assert.notNil(box)
        test.assert.eq(box.h <= 10, true)
    end)

    test.it("text natural width measured from content", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("hbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="short"}, "Hi"),
                    lumina.createElement("text", {id="long"}, "Hello World")
                )
            end})
        ]])
        local short = app:find("short")
        local long = app:find("long")
        test.assert.notNil(short)
        test.assert.notNil(long)
        test.assert.gt(long.w, short.w)
    end)
end)
