-- engine/paint_test.lua — Tests for painting capabilities

test.describe("Engine: Paint", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("text content renders on screen", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="msg"}, "Hello Engine")
                )
            end})
        ]])
        test.assert.eq(app:screenContains("Hello Engine"), true)
    end)

    test.it("foreground color applied to text", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="colored",
                        style={foreground="#FF0000"}}, "Red Text")
                )
            end})
        ]])
        local node = app:find("colored")
        test.assert.notNil(node)
        test.assert.eq(app:screenContains("Red Text"), true)
    end)

    test.it("background color applied to container", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="bg",
                        style={width=10, height=3, background="#00FF00"}},
                        lumina.createElement("text", {}, "Inside")
                    )
                )
            end})
        ]])
        test.assert.eq(app:screenContains("Inside"), true)
    end)

    test.it("border renders around container", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="bordered",
                        style={width=20, height=5, border="single"}},
                        lumina.createElement("text", {}, "Bordered")
                    )
                )
            end})
        ]])
        test.assert.eq(app:screenContains("Bordered"), true)
        -- Border characters should be on screen (single border uses ─ │ ┌ ┐ └ ┘)
        local s = app:screenText()
        test.assert.eq(s:find("─") ~= nil, true)
    end)

    test.it("content outside parent bounds is clipped", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="small",
                        style={width=5, height=1}},
                        lumina.createElement("text", {}, "This is a very long text that should be clipped")
                    )
                )
            end})
        ]])
        -- The full text should NOT appear on screen (clipped to 5 chars width)
        test.assert.eq(app:screenContains("This is a very long text that should be clipped"), false)
    end)
end)
