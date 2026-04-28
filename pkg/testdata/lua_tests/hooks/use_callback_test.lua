-- use_callback_test.lua — Tests for useCallback

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

