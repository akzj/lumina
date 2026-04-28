test.describe("lumina.store", function()
    test.it("set and get", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.store.set("count", 42)
            lumina.store.set("name", "Alice")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local count = lumina.store.get("count")
                    local name = lumina.store.get("name")
                    return lumina.createElement("text", {id = "out"},
                        "count:" .. tostring(count) .. " name:" .. tostring(name))
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "count:42")
        test.assert.contains(node.content, "name:Alice")
        app:destroy()
    end)

    test.it("getAll returns all state", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.store.set("a", 1)
            lumina.store.set("b", 2)

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local all = lumina.store.getAll()
                    return lumina.createElement("text", {id = "out"},
                        "a:" .. tostring(all.a) .. " b:" .. tostring(all.b))
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "a:1")
        test.assert.contains(node.content, "b:2")
        app:destroy()
    end)

    test.it("delete removes key", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.store.set("x", 99)
            lumina.store.delete("x")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local x = lumina.store.get("x")
                    return lumina.createElement("text", {id = "out"},
                        "x:" .. tostring(x))
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "x:nil")
        app:destroy()
    end)

    test.it("batch sets multiple keys", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.store.batch({x = 10, y = 20})

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local x = lumina.store.get("x")
                    local y = lumina.store.get("y")
                    return lumina.createElement("text", {id = "out"},
                        "x:" .. tostring(x) .. " y:" .. tostring(y))
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "x:10")
        test.assert.contains(node.content, "y:20")
        app:destroy()
    end)
end)

test.describe("lumina.useStore", function()
    test.it("reads store value in component", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.store.set("user", "Bob")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local user = lumina.useStore("user")
                    return lumina.createElement("text", {id = "out"},
                        "user:" .. tostring(user))
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "user:Bob")
        app:destroy()
    end)

    test.it("re-renders when store changes", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.store.set("count", 0)

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local count = lumina.useStore("count")
                    return lumina.createElement("text", {id = "out"},
                        "count:" .. tostring(count))
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "count:0")

        -- Update store
        app:loadString([[lumina.store.set("count", 5)]])
        app:render()
        node = app:find("out")
        test.assert.contains(node.content, "count:5")
        app:destroy()
    end)
end)

test.describe("lumina.router", function()
    test.it("navigate and path", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.router.addRoute("/")
            lumina.router.addRoute("/about")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local path = lumina.router.path()
                    return lumina.createElement("text", {id = "out"},
                        "path:" .. path)
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "path:/")

        app:loadString([[lumina.router.navigate("/about")]])
        app:render()
        node = app:find("out")
        test.assert.contains(node.content, "path:/about")
        app:destroy()
    end)

    test.it("back returns to previous", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.router.addRoute("/")
            lumina.router.addRoute("/page2")
            lumina.router.navigate("/page2")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("text", {id = "out"},
                        "path:" .. lumina.router.path())
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "path:/page2")

        app:loadString([[lumina.router.back()]])
        app:render()
        node = app:find("out")
        test.assert.contains(node.content, "path:/")
        app:destroy()
    end)

    test.it("params extracts route parameters", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.router.addRoute("/users/:id")
            lumina.router.navigate("/users/42")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local params = lumina.router.params()
                    return lumina.createElement("text", {id = "out"},
                        "id:" .. tostring(params.id))
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "id:42")
        app:destroy()
    end)
end)

test.describe("lumina.useRoute", function()
    test.it("returns current route info", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.router.addRoute("/")
            lumina.router.addRoute("/todos")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local route = lumina.useRoute()
                    return lumina.createElement("text", {id = "out"},
                        "route:" .. route.path)
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "route:/")
        app:destroy()
    end)

    test.it("re-renders on navigate", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.router.addRoute("/")
            lumina.router.addRoute("/settings")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local route = lumina.useRoute()
                    return lumina.createElement("text", {id = "out"},
                        "at:" .. route.path)
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "at:/")

        app:loadString([[lumina.router.navigate("/settings")]])
        app:render()
        node = app:find("out")
        test.assert.contains(node.content, "at:/settings")
        app:destroy()
    end)
end)

test.describe("lumina.app", function()
    test.it("creates app with render function", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.app {
                id = "myapp",
                render = function()
                    return lumina.createElement("text", {id = "out"}, "Hello App")
                end,
            }
        ]])
        local node = app:find("out")
        test.assert.notNil(node)
        test.assert.eq(node.content, "Hello App")
        app:destroy()
    end)

    test.it("initializes store from config", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.app {
                id = "myapp",
                store = {
                    theme = "dark",
                    count = 10,
                },
                render = function()
                    local theme = lumina.useStore("theme")
                    local count = lumina.useStore("count")
                    return lumina.createElement("text", {id = "out"},
                        theme .. ":" .. tostring(count))
                end,
            }
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "dark:10")
        app:destroy()
    end)

    test.it("registers routes from config", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.app {
                id = "myapp",
                routes = {
                    ["/"] = true,
                    ["/about"] = true,
                    ["/users/:id"] = true,
                },
                render = function()
                    local route = lumina.useRoute()
                    return lumina.createElement("text", {id = "out"},
                        "path:" .. route.path)
                end,
            }
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "path:/")

        app:loadString([[lumina.router.navigate("/about")]])
        app:render()
        node = app:find("out")
        test.assert.contains(node.content, "path:/about")
        app:destroy()
    end)

    test.it("handles global keybindings", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.app {
                id = "myapp",
                store = { pressed = "none" },
                keys = {
                    ["ctrl+x"] = function()
                        lumina.store.set("pressed", "ctrl+x")
                    end,
                },
                render = function()
                    local pressed = lumina.useStore("pressed")
                    return lumina.createElement("text", {id = "out"},
                        "pressed:" .. tostring(pressed))
                end,
            }
        ]])
        local node = app:find("out")
        test.assert.contains(node.content, "pressed:none")

        app:keyPress("ctrl+x")
        node = app:find("out")
        test.assert.contains(node.content, "pressed:ctrl+x")
        app:destroy()
    end)
end)
