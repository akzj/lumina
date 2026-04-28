-- syntax_sugar_test.lua — Tests for factory __call syntax, lux modules, and defineComponent

test.describe("Factory __call syntax (Go widgets)", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("Checkbox __call syntax with id", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.Checkbox { id = "cb", label = "Accept", checked = true, key = "cb1" }
                    )
                end,
            })
        ]])
        local node = app:find("cb")
        test.assert.notNil(node)
        test.assert.eq(app:screenContains("[x]"), true)
        test.assert.eq(app:screenContains("Accept"), true)
    end)

    test.it("Button __call syntax with id", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.Button { id = "btn", label = "Click", key = "btn1" }
                    )
                end,
            })
        ]])
        local node = app:find("btn")
        test.assert.notNil(node)
        test.assert.eq(app:screenContains("Click"), true)
    end)

    test.it("Switch __call syntax with id", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.Switch { id = "sw", label = "Toggle", checked = true, key = "sw1" }
                    )
                end,
            })
        ]])
        local node = app:find("sw")
        test.assert.notNil(node)
        test.assert.eq(app:screenContains("Toggle"), true)
    end)

    test.it("Radio __call syntax with id", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.Radio { id = "rd", label = "Option", value = "opt1", checked = true, key = "rd1" }
                    )
                end,
            })
        ]])
        local node = app:find("rd")
        test.assert.notNil(node)
        test.assert.eq(app:screenContains("Option"), true)
    end)

    test.it("Label __call syntax with id", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.Label { id = "lbl", text = "Hello Label", key = "lbl1" }
                    )
                end,
            })
        ]])
        local node = app:find("lbl")
        test.assert.notNil(node)
        test.assert.eq(app:screenContains("Hello Label"), true)
    end)

    test.it("Select __call syntax with id", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.Select {
                            id = "sel",
                            placeholder = "Pick...",
                            options = {
                                {label = "A", value = "a"},
                                {label = "B", value = "b"},
                            },
                            key = "sel1",
                        }
                    )
                end,
            })
        ]])
        local node = app:find("sel")
        test.assert.notNil(node)
        test.assert.eq(app:screenContains("Pick..."), true)
    end)
end)

test.describe("Mixed table pattern (props + children)", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("multiple widgets in vbox with id", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.Checkbox { id = "cb1", label = "A", checked = true, key = "cba" },
                        lumina.Checkbox { id = "cb2", label = "B", checked = false, key = "cbb" }
                    )
                end,
            })
        ]])
        test.assert.notNil(app:find("cb1"))
        test.assert.notNil(app:find("cb2"))
        test.assert.eq(app:screenContains("A"), true)
        test.assert.eq(app:screenContains("B"), true)
        test.assert.eq(app:screenContains("[x]"), true)
        test.assert.eq(app:screenContains("[ ]"), true)
    end)

    test.it("createElement still works for Go widgets", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Checkbox, {
                            id = "old-cb",
                            label = "Old Syntax",
                            checked = false,
                            key = "old1",
                        })
                    )
                end,
            })
        ]])
        test.assert.notNil(app:find("old-cb"))
        test.assert.eq(app:screenContains("[ ]"), true)
        test.assert.eq(app:screenContains("Old Syntax"), true)
    end)
end)

test.describe("require lux modules", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("require lux umbrella module", function()
        app:loadString([[
            local lux = require("lux")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lux.Card { id = "card", title = "Hello", key = "card1" }
                    )
                end,
            })
        ]])
        app:render()
        test.assert.notNil(app:find("card"))
        test.assert.eq(app:screenContains("Hello"), true)
    end)

    test.it("require individual lux.card module", function()
        app:loadString([[
            local Card = require("lux.card")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Card { id = "c", title = "Test Card", key = "c1" }
                    )
                end,
            })
        ]])
        app:render()
        test.assert.notNil(app:find("c"))
        test.assert.eq(app:screenContains("Test Card"), true)
    end)

    test.it("require individual lux.badge module", function()
        app:loadString([[
            local Badge = require("lux.badge")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Badge { id = "b", label = "New", variant = "success", key = "b1" }
                    )
                end,
            })
        ]])
        app:render()
        test.assert.notNil(app:find("b"))
        test.assert.eq(app:screenContains("New"), true)
    end)

    test.it("require individual lux.divider module", function()
        app:loadString([[
            local Divider = require("lux.divider")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Divider { id = "div", char = "=", width = 10, key = "div1" }
                    )
                end,
            })
        ]])
        app:render()
        test.assert.notNil(app:find("div"))
        test.assert.eq(app:screenContains("=========="), true)
    end)

    test.it("require individual lux.progress module", function()
        app:loadString([[
            local Progress = require("lux.progress")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Progress { id = "pg", value = 50, key = "pg1" }
                    )
                end,
            })
        ]])
        app:render()
        test.assert.notNil(app:find("pg"))
        test.assert.eq(app:screenContains("50%"), true)
    end)
end)

test.describe("defineComponent factory callable", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("defineComponent returns callable factory", function()
        app:loadString([[
            local MyComp = lumina.defineComponent("MyComp", function(props)
                return lumina.createElement("text", {id = "out"}, "custom:" .. (props.name or ""))
            end)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        MyComp { name = "hello", key = "mc1" }
                    )
                end,
            })
        ]])
        app:render()
        local node = app:find("out")
        test.assert.notNil(node)
        test.assert.contains(node.content, "custom:hello")
    end)

    test.it("defineComponent with children", function()
        app:loadString([[
            local Panel = lumina.defineComponent("Panel", function(props)
                local children = props.children or {}
                local count = 0
                for _ in ipairs(children) do count = count + 1 end
                return lumina.createElement("text", {id = "out"}, "PANEL:" .. tostring(count))
            end)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Panel {
                            key = "panel1",
                            lumina.createElement("text", {}, "child1"),
                            lumina.createElement("text", {}, "child2"),
                        }
                    )
                end,
            })
        ]])
        app:render()
        local node = app:find("out")
        test.assert.notNil(node)
        test.assert.contains(node.content, "PANEL:2")
    end)
end)

test.describe("Nested lux composition", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("nested lux components", function()
        app:loadString([[
            local Card = require("lux.card")
            local Badge = require("lux.badge")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Card { id = "card", title = "Status", key = "card1",
                            Badge { label = "Active", variant = "success", key = "badge1" },
                        }
                    )
                end,
            })
        ]])
        app:render()
        test.assert.notNil(app:find("card"))
        test.assert.eq(app:screenContains("Status"), true)
    end)
end)

test.describe("theme module", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("theme module returns colors", function()
        app:loadString([[
            local theme = require("theme")
            local colors = theme.current()
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("text", {id = "out"}, "base:" .. (colors.base or "none"))
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.notNil(node)
        test.assert.contains(node.content, "base:#")
    end)

    test.it("theme has all expected keys", function()
        app:loadString([[
            local theme = require("theme")
            local colors = theme.current()
            local keys = {"base", "surface0", "surface1", "text", "primary", "success", "warning", "error"}
            local result = ""
            for _, k in ipairs(keys) do
                if colors[k] then
                    result = result .. k .. ":ok "
                else
                    result = result .. k .. ":missing "
                end
            end
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("text", {id = "out"}, result)
                end,
            })
        ]])
        local node = app:find("out")
        test.assert.notNil(node)
        test.assert.contains(node.content, "base:ok")
        test.assert.contains(node.content, "primary:ok")
        test.assert.contains(node.content, "success:ok")
    end)
end)

test.describe("defineComponent as direct root return (graftWalk root-level)", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders when root render returns defineComponent directly", function()
        app:loadString([[
            local MyChild = lumina.defineComponent("DirectChild", function(props)
                return lumina.createElement("vbox", {id = "child-root"},
                    lumina.createElement("text", {id = "child-text"}, "hello from direct root")
                )
            end)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return MyChild { key = "dr" }
                end,
            })
        ]])
        test.assert.eq(app:screenContains("hello from direct root"), true)
    end)
end)
