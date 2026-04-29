-- examples/accordion_demo.lua — Lux Accordion component demo
-- Usage: lumina examples/accordion_demo.lua
-- Quit: q

local lux = require("lux")
local Accordion = lux.Accordion

lumina.app {
    id = "accordion-demo",
    store = {
        openItems = { "faq1" },
        selectedIdx = 1,
        mode = "single",
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
        ["m"] = function()
            local mode = lumina.store.get("mode")
            lumina.store.set("mode", mode == "single" and "multi" or "single")
            lumina.store.set("openItems", {})
        end,
    },
    render = function()
        local t = lumina.getTheme()
        local openItems = lumina.useStore("openItems")
        local selectedIdx = lumina.useStore("selectedIdx")
        local mode = lumina.useStore("mode")

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 60, height = 24, background = t.base or "#1E1E2E" },
        },
            lumina.createElement("text", {
                bold = true, foreground = t.primary or "#89B4FA",
                style = { height = 1 },
            }, "  Accordion Demo (" .. mode .. " mode)"),
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                style = { height = 1 },
            }, "  j/k=navigate  Enter/Space=toggle  m=switch mode  q=quit"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            Accordion {
                id = "demo-accordion",
                width = 56,
                mode = mode,
                openItems = openItems,
                selectedIndex = selectedIdx,
                onToggle = function(id, isOpen, newOpenItems)
                    lumina.store.set("openItems", newOpenItems)
                end,
                onSelectedChange = function(idx)
                    lumina.store.set("selectedIdx", idx)
                end,
                items = {
                    { id = "faq1", title = "What is Lumina?", content = "A TUI framework powered by Lua." },
                    { id = "faq2", title = "How to install?", content = "go install github.com/akzj/lumina" },
                    { id = "faq3", title = "Disabled Section", disabled = true, content = "Cannot open." },
                    { id = "faq4", title = "License", content = "MIT License" },
                },
                autoFocus = true,
            }
        )
    end,
}
