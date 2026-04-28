-- hook_ordering_test.lua — Tests for hook ordering consistency

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

