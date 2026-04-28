-- async_test.lua — Tests for lumina.spawn, sleep, readFile, exec

test.describe("lumina.spawn + async.await", function()
    test.it("spawn runs synchronous function and updates state", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local status, setStatus = lumina.useState("s", "not-run")
                    lumina.useEffect(function()
                        lumina.spawn(function()
                            -- No await → runs synchronously within Spawn
                            setStatus("ran")
                        end)
                    end, {})
                    return lumina.createElement("text", {id = "out"}, status)
                end,
            })
        ]])
        -- RenderAll fires effect → spawn runs immediately → setStatus("ran") → marks dirty
        -- app:render() processes the dirty re-render
        app:render()
        local node = app:find("out")
        test.assert.notNil(node)
        test.assert.eq(node.content, "ran")
        app:destroy()
    end)

    test.it("spawn with sleep + await resumes after waitAsync", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local status, setStatus = lumina.useState("s", "loading")
                    lumina.useEffect(function()
                        lumina.spawn(function()
                            local async = require("async")
                            local future = lumina.sleep(1)  -- 1ms
                            async.await(future)
                            setStatus("done")
                        end)
                    end, {})
                    return lumina.createElement("text", {id = "status"}, status)
                end,
            })
        ]])
        -- After initial render + effect, spawn starts but yields on sleep
        app:render()
        local node = app:find("status")
        test.assert.eq(node.content, "loading")

        -- Wait for async to complete (sleep 1ms), then tick scheduler + render
        app:waitAsync(500)
        app:render()
        node = app:find("status")
        test.assert.eq(node.content, "done")
        app:destroy()
    end)

    test.it("spawn with setState triggers re-render", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local count, setCount = lumina.useState("c", 0)
                    lumina.useEffect(function()
                        lumina.spawn(function()
                            local async = require("async")
                            async.await(lumina.sleep(1))
                            setCount(42)
                        end)
                    end, {})
                    return lumina.createElement("text", {id = "count"},
                        "count:" .. tostring(count))
                end,
            })
        ]])
        app:render()  -- fire effect
        app:waitAsync(500)  -- wait for coroutine to complete
        app:render()  -- render state update
        local node = app:find("count")
        test.assert.eq(node.content, "count:42")
        app:destroy()
    end)
end)

test.describe("lumina.readFile", function()
    test.it("reads file content via async", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local content, setContent = lumina.useState("c", "loading")
                    lumina.useEffect(function()
                        lumina.spawn(function()
                            local async = require("async")
                            local result, readErr = async.await(lumina.readFile("/etc/hosts"))
                            if readErr then
                                setContent("error:" .. readErr)
                            else
                                setContent(result)
                            end
                        end)
                    end, {})
                    return lumina.createElement("text", {id = "content"}, content)
                end,
            })
        ]])
        app:render()  -- fire effect
        app:waitAsync(1000)  -- wait for file read
        app:render()  -- render result

        local node = app:find("content")
        test.assert.notNil(node)
        -- /etc/hosts should exist on all systems
        test.assert.neq(node.content, "loading")
        test.assert.neq(node.content, "")
        app:destroy()
    end)
end)

test.describe("lumina.exec", function()
    test.it("executes shell command via async", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local result, setResult = lumina.useState("r", "pending")
                    lumina.useEffect(function()
                        lumina.spawn(function()
                            local async = require("async")
                            local val = async.await(lumina.exec("echo hello"))
                            if val and val.output then
                                setResult(val.output)
                            else
                                setResult("failed")
                            end
                        end)
                    end, {})
                    return lumina.createElement("text", {id = "result"}, result)
                end,
            })
        ]])
        app:render()  -- fire effect
        app:waitAsync(2000)  -- wait for exec
        app:render()  -- render result

        local node = app:find("result")
        test.assert.notNil(node)
        test.assert.contains(node.content, "hello")
        app:destroy()
    end)
end)

test.describe("lumina.cancel", function()
    test.it("destroy cleans up pending coroutines", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function(props)
                    local status, setStatus = lumina.useState("s", "running")
                    lumina.useEffect(function()
                        lumina.spawn(function()
                            local async = require("async")
                            async.await(lumina.sleep(10000))  -- 10 seconds (won't complete)
                            setStatus("completed")  -- should never reach here
                        end)
                    end, {})
                    return lumina.createElement("text", {id = "status"}, status)
                end,
            })
        ]])
        app:render()  -- fire effect, spawn starts and yields
        local node = app:find("status")
        test.assert.eq(node.content, "running")
        -- Destroy should clean up without errors
        app:destroy()
    end)
end)
