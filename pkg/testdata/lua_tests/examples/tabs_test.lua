-- tabs_test.lua — Comprehensive tests for Tabs component

test.describe("Tabs component", function()
    local app

    test.beforeEach(function()
        app = test.createApp(60, 20)
        app:loadFile("../examples/tabs_demo.lua")
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Rendering
    test.it("renders tab labels", function()
        test.assert.eq(app:screenContains("General"), true)
        test.assert.eq(app:screenContains("Settings"), true)
        test.assert.eq(app:screenContains("Network"), true)
        test.assert.eq(app:screenContains("Disabled"), true)
        test.assert.eq(app:screenContains("About"), true)
    end)

    test.it("shows active tab content", function()
        test.assert.eq(app:screenContains("General settings go here"), true)
    end)

    -- Keyboard navigation
    test.it("l moves to next tab", function()
        app:keyPress("l")
        test.assert.eq(app:screenContains("Advanced settings panel"), true)
    end)

    test.it("h moves to previous tab", function()
        app:keyPress("l")  -- go to settings
        app:keyPress("h")  -- back to general
        test.assert.eq(app:screenContains("General settings go here"), true)
    end)

    test.it("skips disabled tab when navigating right", function()
        -- general → settings → network → (skip disabled) → about
        app:keyPress("l")  -- settings
        app:keyPress("l")  -- network
        app:keyPress("l")  -- should skip disabled, go to about
        test.assert.eq(app:screenContains("Lumina v2"), true)
    end)

    test.it("skips disabled tab when navigating left", function()
        -- Go to about first
        app:keyPress("l")
        app:keyPress("l")
        app:keyPress("l")  -- about
        app:keyPress("h")  -- should skip disabled, go to network
        test.assert.eq(app:screenContains("Network configuration"), true)
    end)

    test.it("wraps around from last to first", function()
        -- Go to about (last non-disabled)
        app:keyPress("l")
        app:keyPress("l")
        app:keyPress("l")  -- about
        app:keyPress("l")  -- should wrap to general
        test.assert.eq(app:screenContains("General settings go here"), true)
    end)

    test.it("wraps around from first to last", function()
        app:keyPress("h")  -- should wrap to about (last non-disabled)
        test.assert.eq(app:screenContains("Lumina v2"), true)
    end)

    test.it("Home goes to first non-disabled tab", function()
        app:keyPress("l")
        app:keyPress("l")  -- network
        app:keyPress("Home")
        test.assert.eq(app:screenContains("General settings go here"), true)
    end)

    test.it("End goes to last non-disabled tab", function()
        app:keyPress("End")
        test.assert.eq(app:screenContains("Lumina v2"), true)
    end)

    -- Store-driven tab switching (simulates click)
    test.it("store-driven tab switch shows correct content", function()
        app:loadString('lumina.store.set("activeTab", "network")')
        test.assert.eq(app:screenContains("Network configuration"), true)
    end)

    test.it("switching to settings via store works", function()
        app:loadString('lumina.store.set("activeTab", "settings")')
        test.assert.eq(app:screenContains("Advanced settings panel"), true)
    end)

    -- Edge cases
    test.it("no crash with empty tabs array", function()
        app:destroy()
        app = test.createApp(60, 20)
        app:loadString([[
            local lux = require("lux")
            local Tabs = lux.Tabs
            lumina.app {
                id = "empty-tabs",
                render = function()
                    return Tabs {
                        id = "t",
                        width = 40, height = 6,
                        tabs = {},
                        activeTab = nil,
                    }
                end,
            }
        ]])
        -- Should not crash — just renders empty
        test.assert.eq(type(app:screenText()), "string")
    end)

    test.it("no crash with single tab", function()
        app:destroy()
        app = test.createApp(60, 20)
        app:loadString([[
            local lux = require("lux")
            local Tabs = lux.Tabs
            lumina.app {
                id = "single-tab",
                render = function()
                    return Tabs {
                        id = "t",
                        width = 40, height = 6,
                        tabs = { { id = "only", label = "Only Tab" } },
                        activeTab = "only",
                        renderContent = function() return lumina.createElement("text", {}, "solo") end,
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Only Tab"), true)
        test.assert.eq(app:screenContains("solo"), true)
    end)

    test.it("all tabs disabled — keyboard does nothing", function()
        app:destroy()
        app = test.createApp(60, 20)
        app:loadString([[
            local lux = require("lux")
            local Tabs = lux.Tabs
            lumina.app {
                id = "all-disabled",
                store = { tab = "a" },
                render = function()
                    return Tabs {
                        id = "t",
                        width = 40, height = 6,
                        tabs = {
                            { id = "a", label = "A", disabled = true },
                            { id = "b", label = "B", disabled = true },
                        },
                        activeTab = lumina.useStore("tab"),
                        onTabChange = function(id) lumina.store.set("tab", id) end,
                        renderContent = function(id) return lumina.createElement("text", {}, "content-" .. id) end,
                        autoFocus = true,
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("content-a"), true)
        app:keyPress("l")
        -- Should still show content-a — no change since all disabled
        test.assert.eq(app:screenContains("content-a"), true)
    end)

    test.it("rapid tab switching works correctly", function()
        -- Rapidly press l multiple times
        app:keyPress("l")
        app:keyPress("l")
        app:keyPress("l")
        app:keyPress("l")  -- wrap back to general
        test.assert.eq(app:screenContains("General settings go here"), true)
    end)

    test.it("mixed h and l navigation", function()
        app:keyPress("l")  -- settings
        app:keyPress("l")  -- network
        app:keyPress("h")  -- settings
        app:keyPress("h")  -- general
        app:keyPress("h")  -- wrap to about
        test.assert.eq(app:screenContains("Lumina v2"), true)
    end)
end)
