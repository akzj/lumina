-- engine/store_test.lua — Tests for global store

test.describe("Engine: Store", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("useStore returns initial state", function()
        app:loadString([[
            lumina.store.set("counter", 0)
            lumina.createComponent({id="test", name="Test", render=function()
                local val = lumina.useStore("counter")
                return lumina.createElement("text", {id="out"}, "Counter:" .. tostring(val))
            end})
        ]])
        test.assert.eq(app:screenContains("Counter:0"), true)
    end)

    test.it("store.set updates state and triggers re-render", function()
        app:loadString([[
            lumina.store.set("count", 0)
            lumina.createComponent({id="test", name="Test", render=function()
                local val = lumina.useStore("count")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="display"}, "N:" .. tostring(val)),
                    lumina.createElement("text", {id="inc",
                        style={width=10, height=1},
                        onClick=function()
                            lumina.store.set("count", val + 1)
                        end}, "Inc")
                )
            end})
        ]])
        test.assert.eq(app:screenContains("N:0"), true)
        app:click(2, 1)
        test.assert.eq(app:screenContains("N:1"), true)
        app:click(2, 1)
        test.assert.eq(app:screenContains("N:2"), true)
    end)

    test.it("store supports table values", function()
        app:loadString([[
            lumina.store.set("user", {name="Alice", age=30})
            lumina.createComponent({id="test", name="Test", render=function()
                local user = lumina.useStore("user")
                return lumina.createElement("text", {id="out"},
                    (user.name or "?") .. ":" .. tostring(user.age or 0))
            end})
        ]])
        test.assert.eq(app:screenContains("Alice:30"), true)
    end)

    test.it("multiple components share store state", function()
        app:loadString([[
            lumina.store.set("shared", "initial")

            local Reader = lumina.defineComponent("Reader", function(props)
                local val = lumina.useStore("shared")
                return lumina.createElement("text", {id="reader"}, "Read:" .. tostring(val))
            end)

            lumina.createComponent({id="test", name="Test", render=function()
                local val = lumina.useStore("shared")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    Reader {key="r1"},
                    lumina.createElement("text", {id="writer",
                        style={width=10, height=1},
                        onClick=function()
                            lumina.store.set("shared", "updated")
                        end}, "Update")
                )
            end})
        ]])
        test.assert.eq(app:screenContains("Read:initial"), true)
        app:click(2, 1)  -- click "Update"
        test.assert.eq(app:screenContains("Read:updated"), true)
    end)
end)
