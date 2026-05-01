-- engine/hooks_test.lua — Tests for component hooks

test.describe("Engine: Hooks", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("useState returns initial value", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local val, _ = lumina.useState("myval", 42)
                return lumina.createElement("text", {id="out"}, "Val:" .. tostring(val))
            end})
        ]])
        test.assert.eq(app:screenContains("Val:42"), true)
    end)

    test.it("useState setter triggers re-render with new value", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local val, setVal = lumina.useState("myval", 0)
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="display"}, "V:" .. tostring(val)),
                    lumina.createElement("text", {id="btn",
                        style={width=10, height=1},
                        onClick=function() setVal(val + 10) end}, "Add10")
                )
            end})
        ]])
        test.assert.eq(app:screenContains("V:0"), true)
        app:click(2, 1)
        test.assert.eq(app:screenContains("V:10"), true)
    end)

    test.it("useEffect runs on mount and triggers re-render", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local mounted, setMounted = lumina.useState("mounted", "no")
                lumina.useEffect(function()
                    setMounted("yes")
                end, {})
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="out"}, "Mounted:" .. mounted)
                )
            end})
        ]])
        -- Effect may fire after render; force a re-render cycle
        app:render()
        -- After effect + re-render, should show "yes"
        test.assert.eq(app:screenContains("Mounted:yes"), true)
    end)

    test.it("useEffect with deps only runs when deps change", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local count, setCount = lumina.useState("count", 0)
                local effectRuns, setEffectRuns = lumina.useState("runs", 0)
                lumina.useEffect(function()
                    setEffectRuns(effectRuns + 1)
                end, {count})
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="info"},
                        "C:" .. tostring(count) .. " R:" .. tostring(effectRuns)),
                    lumina.createElement("text", {id="inc",
                        style={width=10, height=1},
                        onClick=function() setCount(count + 1) end}, "Inc")
                )
            end})
        ]])
        -- Initial: count=0, effect ran once on mount
        test.assert.eq(app:screenContains("C:0"), true)
        -- Click to increment count -> effect should run again
        app:click(2, 1)
        test.assert.eq(app:screenContains("C:1"), true)
    end)

    test.it("multiple useState hooks maintain independent state", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local a, setA = lumina.useState("a", "X")
                local b, setB = lumina.useState("b", "Y")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="display"}, a .. ":" .. b),
                    lumina.createElement("text", {id="btnA",
                        style={width=10, height=1},
                        onClick=function() setA("Z") end}, "SetA"),
                    lumina.createElement("text", {id="btnB",
                        style={width=10, height=1},
                        onClick=function() setB("W") end}, "SetB")
                )
            end})
        ]])
        test.assert.eq(app:screenContains("X:Y"), true)
        app:click(2, 1)  -- SetA
        test.assert.eq(app:screenContains("Z:Y"), true)  -- only A changed
        app:click(2, 2)  -- SetB
        test.assert.eq(app:screenContains("Z:W"), true)  -- only B changed
    end)
end)
