-- error_handling_test.lua — Tests for Lua error surfacing in key handlers

test.describe("Key handler error propagation", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("key handler error is reported, not silently swallowed", function()
        app:loadString([[
            lumina.app {
                id = "test",
                store = { count = 0 },
                keys = {
                    ["x"] = function()
                        -- useStore outside render = error
                        local count = lumina.useStore("count")
                    end,
                },
                render = function()
                    local count = lumina.useStore("count")
                    return lumina.createElement("text", {id = "out"}, "count:" .. tostring(count))
                end,
            }
        ]])
        -- keyPress("x") should raise an error because useStore is called outside render
        local ok, err = pcall(function() app:keyPress("x") end)
        test.assert.eq(ok, false, "key handler error should propagate")
    end)

    test.it("valid key handlers still work after error handling fix", function()
        app:loadString([[
            lumina.app {
                id = "test",
                store = { count = 0 },
                keys = {
                    ["x"] = function()
                        local count = lumina.store.get("count")
                        lumina.store.set("count", count + 1)
                    end,
                },
                render = function()
                    local count = lumina.useStore("count")
                    return lumina.createElement("text", {id = "out"}, "count:" .. tostring(count))
                end,
            }
        ]])
        app:keyPress("x")
        local node = app:find("out")
        test.assert.notNil(node)
        test.assert.contains(node.content, "count:1")
    end)
end)
