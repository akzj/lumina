-- component_toggle_test.lua — Tests for renderInOrder handling newly-created child components

test.describe("Component toggle rendering", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Test 1: basic defineComponent with lumina.app + useStore (no toggle)
    test.it("defineComponent with useStore renders", function()
        app:loadString([[
            local CompA = lumina.defineComponent("CompA", function()
                return lumina.createElement("text", {}, "COMP_A_CONTENT")
            end)
            lumina.app {
                id = "test1",
                store = { show = "a" },
                render = function()
                    local show = lumina.useStore("show")
                    return lumina.createElement("vbox", {},
                        CompA { key = "ca" },
                        lumina.createElement("text", {}, "show:" .. tostring(show))
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("COMP_A_CONTENT"), true)
        test.assert.eq(app:screenContains("show:a"), true)
    end)

    -- Test 2: toggle via store.set (the actual bug scenario)
    test.it("store.set triggers re-render with new content", function()
        app:loadString([[
            lumina.app {
                id = "test2",
                store = { show = "a" },
                keys = {
                    ["x"] = function()
                        local s = lumina.store.get("show")
                        lumina.store.set("show", s == "a" and "b" or "a")
                    end,
                },
                render = function()
                    local show = lumina.useStore("show")
                    if show == "a" then
                        return lumina.createElement("text", {}, "SHOWING_A")
                    else
                        return lumina.createElement("text", {}, "SHOWING_B")
                    end
                end,
            }
        ]])
        test.assert.eq(app:screenContains("SHOWING_A"), true)
        app:keyPress("x")
        test.assert.eq(app:screenContains("SHOWING_B"), true)
        app:keyPress("x")
        test.assert.eq(app:screenContains("SHOWING_A"), true)
    end)

    -- Test 3: toggle with defineComponent children
    test.it("toggle between defineComponents via store", function()
        app:loadString([[
            local CompA = lumina.defineComponent("CompA", function()
                return lumina.createElement("text", {}, "COMP_A_CONTENT")
            end)
            local CompB = lumina.defineComponent("CompB", function()
                return lumina.createElement("text", {}, "COMP_B_CONTENT")
            end)
            lumina.app {
                id = "test3",
                store = { show = "a" },
                keys = {
                    ["x"] = function()
                        local s = lumina.store.get("show")
                        lumina.store.set("show", s == "a" and "b" or "a")
                    end,
                },
                render = function()
                    local show = lumina.useStore("show")
                    if show == "a" then
                        return lumina.createElement("vbox", {},
                            CompA { key = "ca" }
                        )
                    else
                        return lumina.createElement("vbox", {},
                            CompB { key = "cb" }
                        )
                    end
                end,
            }
        ]])
        test.assert.eq(app:screenContains("COMP_A_CONTENT"), true)
        app:keyPress("x")
        test.assert.eq(app:screenContains("COMP_B_CONTENT"), true)
        app:keyPress("x")
        test.assert.eq(app:screenContains("COMP_A_CONTENT"), true)
    end)
end)
