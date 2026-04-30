-- examples/toast_demo.lua — Lux Toast component demo
-- Usage: lumina examples/toast_demo.lua
-- Quit: q

local lux = require("lux")
local Toast = lux.Toast

lumina.app {
    id = "toast-demo",
    store = {
        toasts = {},
        nextId = 1,
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
        ["1"] = function()
            local toasts = lumina.store.get("toasts") or {}
            local id = lumina.store.get("nextId") or 1
            toasts[#toasts + 1] = { id = id, message = "Info notification", variant = "info" }
            lumina.store.set("toasts", toasts)
            lumina.store.set("nextId", id + 1)
        end,
        ["2"] = function()
            local toasts = lumina.store.get("toasts") or {}
            local id = lumina.store.get("nextId") or 1
            toasts[#toasts + 1] = { id = id, message = "Success!", variant = "success" }
            lumina.store.set("toasts", toasts)
            lumina.store.set("nextId", id + 1)
        end,
        ["3"] = function()
            local toasts = lumina.store.get("toasts") or {}
            local id = lumina.store.get("nextId") or 1
            toasts[#toasts + 1] = { id = id, message = "Warning issued", variant = "warning" }
            lumina.store.set("toasts", toasts)
            lumina.store.set("nextId", id + 1)
        end,
        ["4"] = function()
            local toasts = lumina.store.get("toasts") or {}
            local id = lumina.store.get("nextId") or 1
            toasts[#toasts + 1] = { id = id, message = "Error occurred", variant = "error" }
            lumina.store.set("toasts", toasts)
            lumina.store.set("nextId", id + 1)
        end,
        ["c"] = function()
            lumina.store.set("toasts", {})
        end,
    },
    render = function()
        local t = lumina.getTheme()
        local toasts = lumina.useStore("toasts") or {}

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 60, height = 24, background = t.base or "#1E1E2E" },
        },
            lumina.createElement("text", {
                bold = true, foreground = t.primary or "#89B4FA",
                style = { height = 1 },
            }, "  Toast Demo"),
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                style = { height = 1 },
            }, "  1=info 2=success 3=warning 4=error c=clear q=quit"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            Toast {
                key = "toasts",
                items = toasts,
                width = 50,
                maxVisible = 5,
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
