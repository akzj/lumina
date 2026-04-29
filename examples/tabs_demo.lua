-- examples/tabs_demo.lua — Lux Tabs component demo
-- Usage: lumina examples/tabs_demo.lua
-- Quit: q

local lux = require("lux")
local Tabs = lux.Tabs

lumina.app {
    id = "tabs-demo",
    store = { activeTab = "general" },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
    },
    render = function()
        local t = lumina.getTheme()
        local activeTab = lumina.useStore("activeTab")

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 60, height = 20, background = t.base or "#1E1E2E" },
        },
            lumina.createElement("text", {
                bold = true,
                foreground = t.primary or "#89B4FA",
                style = { height = 1 },
            }, "  Tabs Demo"),
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                style = { height = 1 },
            }, "  h/l arrows=switch tabs  q=quit"),
            Tabs {
                id = "demo-tabs",
                width = 56,
                height = 12,
                activeTab = activeTab,
                onTabChange = function(id)
                    lumina.store.set("activeTab", id)
                end,
                tabs = {
                    { id = "general", label = "General" },
                    { id = "settings", label = "Settings" },
                    { id = "network", label = "Network" },
                    { id = "disabled", label = "Disabled", disabled = true },
                    { id = "about", label = "About" },
                },
                renderContent = function(tabId)
                    if tabId == "general" then
                        return lumina.createElement("vbox", {},
                            lumina.createElement("text", { foreground = t.text or "#CDD6F4", style = { height = 1 } }, "  General settings go here."),
                            lumina.createElement("text", { foreground = t.muted or "#6C7086", style = { height = 1 } }, "  User preferences, language, etc.")
                        )
                    elseif tabId == "settings" then
                        return lumina.createElement("text", { foreground = t.text or "#CDD6F4", style = { height = 1 } }, "  Advanced settings panel")
                    elseif tabId == "network" then
                        return lumina.createElement("text", { foreground = t.warning or "#F9E2AF", style = { height = 1 } }, "  Network configuration")
                    elseif tabId == "about" then
                        return lumina.createElement("text", { foreground = t.text or "#CDD6F4", style = { height = 1 } }, "  Lumina v2 — TUI Framework")
                    end
                    return lumina.createElement("text", {}, "")
                end,
                autoFocus = true,
            }
        )
    end,
}
