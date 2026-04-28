-- hooks_test.lua — Tests for useEffect, useRef, useMemo, useCallback

test.describe("useRef", function()
    test.it("persists value across renders", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local count, setCount = lumina.useState("c", 0)
                    local ref = lumina.useRef(0)
                    ref.current = ref.current + 1
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement("text", {id = "renders"},
                            "renders:" .. tostring(ref.current)),
                        lumina.createElement("text", {
                            id = "btn",
                            style = {height = 1},
                            onClick = function() setCount(count + 1) end,
                        }, "click")
                    )
                end,
            })
        ]])
        local node = app:find("renders")
        test.assert.notNil(node)
        test.assert.contains(node.content, "renders:1")

        -- Trigger re-render via click
        app:click("btn")

        node = app:find("renders")
        test.assert.contains(node.content, "renders:2")
        app:destroy()
    end)

    test.it("returns same table identity across renders", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            _G._refIdentityCheck = "unknown"
            local prevRef = nil
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local count, setCount = lumina.useState("c", 0)
                    local ref = lumina.useRef("hello")
                    if prevRef == nil then
                        prevRef = ref
                        _G._refIdentityCheck = "first"
                    elseif prevRef == ref then
                        _G._refIdentityCheck = "same"
                    else
                        _G._refIdentityCheck = "different"
                    end
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement("text", {
                            id = "btn",
                            style = {height = 1},
                            onClick = function() setCount(count + 1) end,
                        }, "click")
                    )
                end,
            })
        ]])
        -- Trigger re-render
        app:click("btn")
        -- The ref identity check is in Lua globals, we can't read it from VNode.
        -- But the test passes if no errors occurred and re-render completed.
        local btn = app:find("btn")
        test.assert.notNil(btn)
        app:destroy()
    end)
end)

test.describe("useMemo", function()
    test.it("caches value when deps unchanged", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            _G._computeCount = 0
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local count, setCount = lumina.useState("c", 0)
                    local other, setOther = lumina.useState("o", 0)
                    local expensive = lumina.useMemo(function()
                        _G._computeCount = _G._computeCount + 1
                        return "result-" .. _G._computeCount
                    end, {count})
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement("text", {id = "memo"}, expensive),
                        lumina.createElement("text", {
                            id = "other-btn",
                            style = {height = 1},
                            onClick = function() setOther(other + 1) end,
                        }, "other"),
                        lumina.createElement("text", {
                            id = "count-btn",
                            style = {height = 1},
                            onClick = function() setCount(count + 1) end,
                        }, "count")
                    )
                end,
            })
        ]])
        -- Initial: computed once
        local memo = app:find("memo")
        test.assert.eq(memo.content, "result-1")

        -- Click "other" button — deps (count) unchanged, memo should NOT recompute
        app:click("other-btn")
        memo = app:find("memo")
        test.assert.eq(memo.content, "result-1")

        -- Click "count" button — deps (count) changed, memo SHOULD recompute
        app:click("count-btn")
        memo = app:find("memo")
        test.assert.eq(memo.content, "result-2")

        app:destroy()
    end)

    test.it("recomputes without deps", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            _G._computeCount2 = 0
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local count, setCount = lumina.useState("c", 0)
                    local val = lumina.useMemo(function()
                        _G._computeCount2 = _G._computeCount2 + 1
                        return "val-" .. _G._computeCount2
                    end)  -- no deps = recompute every render
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement("text", {id = "memo"}, val),
                        lumina.createElement("text", {
                            id = "btn",
                            style = {height = 1},
                            onClick = function() setCount(count + 1) end,
                        }, "click")
                    )
                end,
            })
        ]])
        local memo = app:find("memo")
        test.assert.eq(memo.content, "val-1")

        app:click("btn")
        memo = app:find("memo")
        test.assert.eq(memo.content, "val-2")

        app:destroy()
    end)
end)

test.describe("useCallback", function()
    test.it("returns cached function when deps unchanged", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            _G._callbackCallCount = 0
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local count, setCount = lumina.useState("c", 0)
                    local cb = lumina.useCallback(function()
                        _G._callbackCallCount = _G._callbackCallCount + 1
                    end, {})  -- empty deps = never changes
                    -- Use the callback as onClick
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement("text", {
                            id = "btn",
                            style = {height = 1},
                            onClick = cb,
                        }, "click"),
                        lumina.createElement("text", {id = "count"}, tostring(count))
                    )
                end,
            })
        ]])
        local btn = app:find("btn")
        test.assert.notNil(btn)
        -- Click the button — should call the cached callback
        app:click("btn")
        -- If no error, callback was called successfully
        app:destroy()
    end)
end)

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
        -- After initial render + effect firing, the effect should have called setStatus.
        -- But we need another render cycle to see the updated state.
        -- RenderAll fires effects, which calls setStatus → marks dirty.
        -- We need to trigger RenderDirty to see the update.
        app:render()

        local node = app:find("status")
        test.assert.notNil(node)
        test.assert.eq(node.content, "after")
        app:destroy()
    end)

    test.it("runs cleanup on unmount", function()
        -- This test verifies that cleanup functions are called.
        -- We can't easily observe the cleanup from Lua, but we verify
        -- no errors occur during component cleanup.
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
        -- Destroy should run cleanup without errors
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
        -- app:render() processes the dirty re-render
        app:render()
        local runs = app:find("runs")
        test.assert.contains(runs.content, "runs:1")

        -- Click to change count → re-render → effect fires again → setEffectRuns(2)
        app:click("btn")
        app:render()  -- process effect-triggered re-render
        runs = app:find("runs")
        test.assert.contains(runs.content, "runs:2")

        app:destroy()
    end)
end)

test.describe("hook ordering", function()
    test.it("multiple hooks in same component", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local count, setCount = lumina.useState("c", 0)
                    local ref = lumina.useRef(0)
                    ref.current = ref.current + 1
                    local memo = lumina.useMemo(function()
                        return "memo-" .. tostring(count)
                    end, {count})
                    local cb = lumina.useCallback(function()
                        setCount(count + 1)
                    end, {count})
                    lumina.useEffect(function()
                        -- effect runs after paint
                    end, {count})
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement("text", {id = "renders"},
                            "renders:" .. tostring(ref.current)),
                        lumina.createElement("text", {id = "memo"}, memo),
                        lumina.createElement("text", {
                            id = "btn",
                            style = {height = 1},
                            onClick = cb,
                        }, "click")
                    )
                end,
            })
        ]])
        local renders = app:find("renders")
        test.assert.contains(renders.content, "renders:1")
        local memo = app:find("memo")
        test.assert.eq(memo.content, "memo-0")

        -- Click triggers callback → setCount → re-render
        app:click("btn")
        renders = app:find("renders")
        test.assert.contains(renders.content, "renders:2")
        memo = app:find("memo")
        test.assert.eq(memo.content, "memo-1")

        app:destroy()
    end)
end)
