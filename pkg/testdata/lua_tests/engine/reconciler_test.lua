-- engine/reconciler_test.lua — Tests for reconciliation (diffing/patching)

test.describe("Engine: Reconciler", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("key-based matching preserves state across re-render", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local count, setCount = lumina.useState("count", 0)
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="counter", key="ctr"},
                        "Count:" .. tostring(count)),
                    lumina.createElement("text", {id="btn",
                        style={width=10, height=1},
                        onClick=function() setCount(count + 1) end}, "Inc")
                )
            end})
        ]])
        test.assert.eq(app:screenContains("Count:0"), true)
        app:click(2, 1)
        test.assert.eq(app:screenContains("Count:1"), true)
        app:click(2, 1)
        test.assert.eq(app:screenContains("Count:2"), true)
    end)

    test.it("adding new child renders correctly", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local show, setShow = lumina.useState("show", false)
                local children = {
                    lumina.createElement("text", {key="a"}, "ItemA"),
                }
                if show then
                    children[#children + 1] = lumina.createElement("text", {key="b"}, "ItemB")
                end
                children[#children + 1] = lumina.createElement("text", {id="toggle",
                    style={width=10, height=1},
                    onClick=function() setShow(true) end}, "Add")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    table.unpack(children))
            end})
        ]])
        test.assert.eq(app:screenContains("ItemA"), true)
        test.assert.eq(app:screenContains("ItemB"), false)
        app:click(2, 1)  -- click "Add"
        test.assert.eq(app:screenContains("ItemB"), true)
    end)

    test.it("removing child updates display", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local show, setShow = lumina.useState("show", true)
                local children = {}
                if show then
                    children[#children + 1] = lumina.createElement("text", {key="removable"}, "REMOVABLE")
                end
                children[#children + 1] = lumina.createElement("text", {id="toggle",
                    style={width=10, height=1},
                    onClick=function() setShow(false) end}, "Remove")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    table.unpack(children))
            end})
        ]])
        test.assert.eq(app:screenContains("REMOVABLE"), true)
        app:click(2, 1)  -- click "Remove"
        test.assert.eq(app:screenContains("REMOVABLE"), false)
    end)

    test.it("useState survives re-render", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local name, setName = lumina.useState("name", "Alice")
                local count, setCount = lumina.useState("count", 0)
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="info"},
                        name .. ":" .. tostring(count)),
                    lumina.createElement("text", {id="inc",
                        style={width=10, height=1},
                        onClick=function() setCount(count + 1) end}, "Inc")
                )
            end})
        ]])
        test.assert.eq(app:screenContains("Alice:0"), true)
        app:click(2, 1)
        test.assert.eq(app:screenContains("Alice:1"), true)
        -- Name state preserved across re-render triggered by count change
        app:click(2, 1)
        test.assert.eq(app:screenContains("Alice:2"), true)
    end)
end)
