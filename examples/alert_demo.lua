-- examples/alert_demo.lua — Lux Alert component demo
-- Usage: lumina examples/alert_demo.lua
-- Quit: q

local lux = require("lux")
local Alert = lux.Alert

lumina.app {
    id = "alert-demo",
    store = {
        showDismissible = true,
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
        ["r"] = function() lumina.store.set("showDismissible", true) end,
    },
    render = function()
        local t = lumina.getTheme()
        local showDismissible = lumina.useStore("showDismissible")

        local children = {
            lumina.createElement("text", {
                bold = true, foreground = t.primary or "#89B4FA",
                style = { height = 1 },
            }, "  Alert Demo"),
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                style = { height = 1 },
            }, "  r=restore dismissed  q=quit"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            Alert { key = "alert-info", variant = "info", title = "Info", message = "This is an informational message.", width = 50 },
            lumina.createElement("text", { style = { height = 1 } }, ""),
            Alert { key = "alert-success", variant = "success", title = "Done", message = "Operation completed successfully.", width = 50 },
            lumina.createElement("text", { style = { height = 1 } }, ""),
            Alert { key = "alert-warning", variant = "warning", message = "Disk space running low.", width = 50 },
            lumina.createElement("text", { style = { height = 1 } }, ""),
            Alert { key = "alert-error", variant = "error", title = "Error", message = "Connection failed.", width = 50 },
        }

        -- Dismissible alert
        if showDismissible then
            children[#children + 1] = lumina.createElement("text", { style = { height = 1 } }, "")
            children[#children + 1] = Alert {
                key = "alert-dismiss",
                variant = "info",
                title = "Dismissible",
                message = "Click the X to dismiss this alert.",
                dismissible = true,
                onDismiss = function()
                    lumina.store.set("showDismissible", false)
                end,
                width = 50,
            }
        end

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 60, height = 24, background = t.base or "#1E1E2E" },
        }, table.unpack(children))
    end,
}
