-- toast_test.lua — Tests for Toast notification component

test.describe("Toast component", function()
    local app

    test.beforeEach(function()
        app = test.createApp(60, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Rendering
    test.it("renders toast items with messages", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Toast {
                            key = "t",
                            items = {
                                { id = 1, message = "Hello World", variant = "info" },
                                { id = 2, message = "Success!", variant = "success" },
                            },
                            width = 50,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Hello World"), true)
        test.assert.eq(app:screenContains("Success!"), true)
    end)

    test.it("shows variant icons", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Toast {
                            key = "t",
                            items = {
                                { id = 1, message = "info msg", variant = "info" },
                                { id = 2, message = "err msg", variant = "error" },
                                { id = 3, message = "warn msg", variant = "warning" },
                                { id = 4, message = "ok msg", variant = "success" },
                            },
                            width = 50,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("info msg"), true)
        test.assert.eq(app:screenContains("err msg"), true)
        test.assert.eq(app:screenContains("warn msg"), true)
        test.assert.eq(app:screenContains("ok msg"), true)
    end)

    test.it("maxVisible limits shown toasts", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Toast {
                            key = "t",
                            items = {
                                { id = 1, message = "First" },
                                { id = 2, message = "Second" },
                                { id = 3, message = "Third" },
                                { id = 4, message = "Fourth" },
                            },
                            maxVisible = 2,
                            width = 50,
                        }
                    )
                end,
            }
        ]])
        -- Only last 2 should be visible
        test.assert.eq(app:screenContains("First"), false)
        test.assert.eq(app:screenContains("Second"), false)
        test.assert.eq(app:screenContains("Third"), true)
        test.assert.eq(app:screenContains("Fourth"), true)
    end)

    test.it("empty items renders nothing", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Toast {
                            key = "t",
                            items = {},
                            width = 50,
                        },
                        lumina.createElement("text", {}, "below-toast")
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("below-toast"), true)
    end)

    test.it("dismiss removes toast via store", function()
        app:loadString([[
            lumina.store.set("toasts", {
                { id = 1, message = "Keep me" },
                { id = 2, message = "Remove me" },
            })
            local Toast = require("lux.toast")
            lumina.app {
                id = "test",
                render = function()
                    local toasts = lumina.useStore("toasts") or {}
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Toast {
                            key = "t",
                            items = toasts,
                            width = 50,
                            onDismiss = function(id)
                                local current = lumina.store.get("toasts") or {}
                                local filtered = {}
                                for _, item in ipairs(current) do
                                    if item.id ~= id then
                                        filtered[#filtered + 1] = item
                                    end
                                end
                                lumina.store.set("toasts", filtered)
                            end,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Keep me"), true)
        test.assert.eq(app:screenContains("Remove me"), true)
        -- Dismiss second toast via store manipulation
        app:loadString([[
            lumina.store.set("toasts", {{ id = 1, message = "Keep me" }})
        ]])
        test.assert.eq(app:screenContains("Keep me"), true)
        test.assert.eq(app:screenContains("Remove me"), false)
    end)

    test.it("multiple toasts stack vertically", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Toast {
                            key = "t",
                            items = {
                                { id = 1, message = "Toast A", variant = "info" },
                                { id = 2, message = "Toast B", variant = "success" },
                                { id = 3, message = "Toast C", variant = "error" },
                            },
                            width = 50,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Toast A"), true)
        test.assert.eq(app:screenContains("Toast B"), true)
        test.assert.eq(app:screenContains("Toast C"), true)
    end)

    -- Demo file test
    test.it("demo loads and renders", function()
        app:destroy()
        app = test.createApp(60, 24)
        app:loadFile("../examples/toast_demo.lua")
        test.assert.eq(app:screenContains("Toast Demo"), true)
    end)

    test.it("demo adds toasts via keys", function()
        app:destroy()
        app = test.createApp(60, 24)
        app:loadFile("../examples/toast_demo.lua")
        app:keyPress("1")
        test.assert.eq(app:screenContains("Info notification"), true)
        app:keyPress("2")
        test.assert.eq(app:screenContains("Success!"), true)
    end)

    -- Accessible via lux umbrella
    test.it("accessible via require('lux').Toast", function()
        app:loadString([[
            local lux = require("lux")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        lux.Toast {
                            key = "t",
                            items = {{ id = 1, message = "Lux Toast" }},
                            width = 50,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Lux Toast"), true)
    end)
end)
