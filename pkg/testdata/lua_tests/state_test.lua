test.describe("state API", function()
    test.it("getState reads component state", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "counter", name = "Counter",
                render = function(props)
                    local count, setCount = lumina.useState("count", 42)
                    return lumina.createElement("text", {
                        id = "out",
                        onClick = function() setCount(count + 1) end,
                    }, "count:" .. tostring(count))
                end,
            })
        ]])
        local val = app:getState("counter", "count")
        test.assert.eq(val, 42)
        app:destroy()
    end)

    test.it("setState updates and re-renders", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "counter", name = "Counter",
                render = function(props)
                    local count, setCount = lumina.useState("count", 0)
                    return lumina.createElement("text", {id = "out"}, "count:" .. tostring(count))
                end,
            })
        ]])
        app:setState("counter", "count", 99)
        test.assert.contains(app:screenText(), "count:99")
        app:destroy()
    end)

    test.it("snapshot returns full state", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "myapp", name = "MyApp",
                render = function(props)
                    return lumina.createElement("text", {id = "out"}, "snapshot test")
                end,
            })
        ]])
        local snap = app:snapshot()
        test.assert.notNil(snap)
        test.assert.notNil(snap.screen)
        test.assert.contains(snap.screen, "snapshot test")
        test.assert.notNil(snap.components)
        test.assert.gt(#snap.components, 0)
        app:destroy()
    end)

    test.it("focusedID returns nil when no focus", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    return lumina.createElement("text", {}, "no focus")
                end,
            })
        ]])
        test.assert.isNil(app:focusedID())
        app:destroy()
    end)
end)
