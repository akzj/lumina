-- engine/component_test.lua — Tests for component system

test.describe("Engine: Components", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("defineComponent creates reusable component", function()
        app:loadString([[
            local Greeting = lumina.defineComponent("Greeting", function(props)
                return lumina.createElement("text", {id="greet"}, "Hello " .. (props.name or "World"))
            end)
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    Greeting {name="Engine", key="g1"}
                )
            end})
        ]])
        test.assert.eq(app:screenContains("Hello Engine"), true)
    end)

    test.it("props pass from parent to child component", function()
        app:loadString([[
            local Display = lumina.defineComponent("Display", function(props)
                return lumina.createElement("text", {id="out"},
                    (props.label or "") .. "=" .. tostring(props.value or 0))
            end)
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    Display {label="Score", value=99, key="d1"}
                )
            end})
        ]])
        test.assert.eq(app:screenContains("Score=99"), true)
    end)

    test.it("children slot via props.children", function()
        app:loadString([[
            local Panel = lumina.defineComponent("Panel", function(props)
                local children = props.children or {}
                return lumina.createElement("vbox", {id="panel",
                    style={width=40, height=10, border="single"}},
                    lumina.createElement("text", {}, "Panel:" .. tostring(#children) .. " children"),
                    table.unpack(children)
                )
            end)
            lumina.createComponent({id="test", name="Test", render=function()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    Panel {key="p1",
                        lumina.createElement("text", {key="c1"}, "ChildA"),
                        lumina.createElement("text", {key="c2"}, "ChildB"),
                    }
                )
            end})
        ]])
        test.assert.eq(app:screenContains("Panel:2 children"), true)
        test.assert.eq(app:screenContains("ChildA"), true)
        test.assert.eq(app:screenContains("ChildB"), true)
    end)

    test.it("component re-renders when props change", function()
        app:loadString([[
            local Counter = lumina.defineComponent("Counter", function(props)
                return lumina.createElement("text", {id="cval"}, "N:" .. tostring(props.n or 0))
            end)
            lumina.createComponent({id="test", name="Test", render=function()
                local n, setN = lumina.useState("n", 0)
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    Counter {n=n, key="ctr"},
                    lumina.createElement("text", {id="inc",
                        style={width=10, height=1},
                        onClick=function() setN(n + 1) end}, "Inc")
                )
            end})
        ]])
        test.assert.eq(app:screenContains("N:0"), true)
        app:click(2, 1)
        test.assert.eq(app:screenContains("N:1"), true)
    end)
end)
