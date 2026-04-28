-- use_memo_test.lua — Tests for useMemo

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

