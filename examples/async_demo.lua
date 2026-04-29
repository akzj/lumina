-- examples/async_demo.lua — Async capabilities demo
--
-- Demonstrates: lumina.spawn, lumina.sleep, lumina.exec, async.await
-- Usage: lumina examples/async_demo.lua
-- Quit: q

lumina.app {
    id = "async-demo",
    store = {
        status = "Ready",
        execResult = "",
        counter = 0,
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
        ["s"] = function()
            lumina.store.set("status", "Running sleep...")
            lumina.spawn(function()
                local async = require("async")
                for i = 1, 5 do
                    async.await(lumina.sleep(500))
                    lumina.store.set("counter", i)
                    lumina.store.set("status", "Sleep tick " .. tostring(i) .. "/5")
                end
                lumina.store.set("status", "Sleep done!")
            end)
        end,
        ["e"] = function()
            lumina.store.set("status", "Running exec...")
            lumina.spawn(function()
                local async = require("async")
                local f = lumina.exec("date")
                local r = async.await(f)
                lumina.store.set("execResult", r.output or "")
                lumina.store.set("status", "Exec done!")
            end)
        end,
    },
    render = function()
        local t = lumina.getTheme()
        local status = lumina.useStore("status")
        local counter = lumina.useStore("counter")
        local execResult = lumina.useStore("execResult")

        return lumina.createElement("vbox", {
            style = { width = 60, height = 16, background = t.base or "#1E1E2E" },
        },
            lumina.createElement("text", {
                bold = true, foreground = t.primary or "#89B4FA",
                style = { height = 1 },
            }, "  Async Demo"),
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                style = { height = 1 },
            }, "  s=sleep demo  e=exec demo  q=quit"),
            lumina.createElement("text", {
                foreground = t.text or "#CDD6F4",
                style = { height = 1 },
            }, ""),
            lumina.createElement("text", {
                foreground = t.warning or "#F9E2AF",
                style = { height = 1 },
            }, "  Status: " .. status),
            lumina.createElement("text", {
                foreground = t.text or "#CDD6F4",
                style = { height = 1 },
            }, "  Counter: " .. tostring(counter)),
            lumina.createElement("text", {
                foreground = t.text or "#CDD6F4",
                style = { height = 1 },
            }, "  Exec: " .. execResult)
        )
    end,
}
