-- examples/theme_demo.lua — Theme switching demo
-- Usage: lumina examples/theme_demo.lua
-- Keys: t=toggle theme  q=quit

local lux = require("lux")
local Alert = lux.Alert
local Badge = lux.Badge

local themes = { "mocha", "latte", "nord", "dracula" }

lumina.app {
    id = "theme-demo",
    store = {
        themeIdx = 1,
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
        ["t"] = function()
            local idx = lumina.store.get("themeIdx")
            idx = (idx % #themes) + 1
            lumina.setTheme(themes[idx])
            lumina.store.set("themeIdx", idx)
        end,
    },
    render = function()
        local t = lumina.getTheme()
        local idx = lumina.useStore("themeIdx")
        local name = themes[idx]

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 60, height = 24, background = t.base },
        },
            lumina.createElement("text", {
                bold = true, foreground = t.primary,
                style = { height = 1 },
            }, "  Theme Demo — press [t] to cycle themes"),
            lumina.createElement("text", {
                foreground = t.muted,
                style = { height = 1 },
            }, "  Current theme: " .. name),
            lumina.createElement("text", { style = { height = 1 } }, ""),

            -- Color swatches
            lumina.createElement("text", {
                foreground = t.text,
                style = { height = 1 },
            }, "  Colors:"),
            lumina.createElement("text", {
                foreground = t.primary,
                style = { height = 1 },
            }, "  ■ primary: " .. (t.primary or "?")),
            lumina.createElement("text", {
                foreground = t.success,
                style = { height = 1 },
            }, "  ■ success: " .. (t.success or "?")),
            lumina.createElement("text", {
                foreground = t.warning,
                style = { height = 1 },
            }, "  ■ warning: " .. (t.warning or "?")),
            lumina.createElement("text", {
                foreground = t.error,
                style = { height = 1 },
            }, "  ■ error:   " .. (t.error or "?")),
            lumina.createElement("text", {
                foreground = t.muted,
                style = { height = 1 },
            }, "  ■ muted:   " .. (t.muted or "?")),
            lumina.createElement("text", { style = { height = 1 } }, ""),

            -- Lux components with theme colors
            Alert {
                key = "alert-info",
                variant = "info",
                title = "Themed Alert",
                message = "This alert uses the current theme colors.",
                width = 50,
            },
            lumina.createElement("text", { style = { height = 1 } }, ""),
            lumina.createElement("hbox", {
                style = { height = 1 },
            },
                lumina.createElement("text", { style = { width = 2 } }, "  "),
                Badge { label = name, variant = "default" },
                lumina.createElement("text", { style = { width = 2 } }, "  "),
                Badge { label = "Active", variant = "success" }
            ),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            lumina.createElement("text", {
                foreground = t.muted,
                style = { height = 1 },
            }, "  [t] cycle theme  [q] quit")
        )
    end,
}
