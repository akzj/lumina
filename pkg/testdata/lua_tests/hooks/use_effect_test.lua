-- use_effect_test.lua — Tests for useEffect

test.describe("useEffect", function()
    test.it("runs effect on mount", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local status, setStatus = lumina.useState("s", "before")
                    lumina.useEffect(function()
                        setStatus("after")
                    end, {})  -- empty deps = run once on mount
                    return lumina.createElement("text", {id = "status"}, status)
                end,
            })
        ]])
        -- RenderAll fires effects, which calls setStatus → marks dirty.
        -- We need to trigger RenderDirty to see the update.
        app:render()

        local node = app:find("status")
        test.assert.notNil(node)
        test.assert.eq(node.content, "after")
        app:destroy()
    end)

    test.it("runs cleanup on unmount", function()
        -- Verifies cleanup functions are accepted and do not crash on destroy.
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    lumina.useEffect(function()
                        return function()
                            -- cleanup function
                        end
                    end, {})
                    return lumina.createElement("text", {id = "out"}, "hello")
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.notNil(node)
        app:destroy()
    end)

    test.it("re-runs when deps change", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local count, setCount = lumina.useState("c", 0)
                    local effectRuns, setEffectRuns = lumina.useState("er", 0)
                    lumina.useEffect(function()
                        setEffectRuns(effectRuns + 1)
                    end, {count})
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement("text", {id = "runs"},
                            "runs:" .. tostring(effectRuns)),
                        lumina.createElement("text", {
                            id = "btn",
                            style = {height = 1},
                            onClick = function() setCount(count + 1) end,
                        }, "click")
                    )
                end,
            })
        ]])
        -- After mount: effect fires → setEffectRuns(1) → marks dirty
        app:render()
        local runs = app:find("runs")
        test.assert.contains(runs.content, "runs:1")

        -- Click to change count → effect fires again → setEffectRuns(2)
        app:click("btn")
        app:render()
        runs = app:find("runs")
        test.assert.contains(runs.content, "runs:2")

        app:destroy()
    end)
end)

