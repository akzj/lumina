-- examples/web_demo.lua
-- Run with: lumina --web :8080 examples/web_demo.lua
-- Then open http://localhost:8080 in browser

lumina.app {
    id = "web-demo",
    store = { count = 0 },
    keys = {
        ["q"] = function() lumina.quit() end,
        [" "] = function()
            local c = lumina.store.get("count")
            lumina.store.set("count", c + 1)
        end,
    },
    render = function()
        local t = lumina.getTheme()
        local count = lumina.useStore("count")
        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 40, height = 10, background = t.base },
        },
            lumina.createElement("text", {
                foreground = t.primary,
                bold = true,
                style = { height = 1 },
            }, "  Lumina Web Demo"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            lumina.createElement("text", {
                foreground = t.text,
                style = { height = 1 },
            }, "  Count: " .. tostring(count)),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            lumina.createElement("text", {
                foreground = t.muted,
                style = { height = 1 },
            }, "  [Space] increment  [q] quit")
        )
    end,
}
