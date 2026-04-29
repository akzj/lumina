-- examples/breadcrumb_demo.lua — Lux Breadcrumb component demo
-- Usage: lumina examples/breadcrumb_demo.lua
-- Quit: q

local lux = require("lux")
local Breadcrumb = lux.Breadcrumb

local pages = {
    { id = "home", label = "Home", content = "Welcome to the home page." },
    { id = "products", label = "Products", content = "Browse our products." },
    { id = "electronics", label = "Electronics", content = "Electronic devices." },
    { id = "phones", label = "Phones", content = "Mobile phones catalog." },
}

lumina.app {
    id = "breadcrumb-demo",
    store = {
        depth = 4,  -- how deep in the breadcrumb trail
        lastNav = "",
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
        ["1"] = function() lumina.store.set("depth", 1) end,
        ["2"] = function() lumina.store.set("depth", 2) end,
        ["3"] = function() lumina.store.set("depth", 3) end,
        ["4"] = function() lumina.store.set("depth", 4) end,
    },
    render = function()
        local t = lumina.getTheme()
        local depth = lumina.useStore("depth")
        local lastNav = lumina.useStore("lastNav")

        -- Build breadcrumb items up to current depth
        local bcItems = {}
        for i = 1, math.min(depth, #pages) do
            bcItems[i] = pages[i]
        end

        local currentPage = pages[math.min(depth, #pages)]

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 60, height = 16, background = t.base or "#1E1E2E" },
        },
            lumina.createElement("text", {
                bold = true, foreground = t.primary or "#89B4FA",
                style = { height = 1 },
            }, "  Breadcrumb Demo"),
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                style = { height = 1 },
            }, "  1-4=set depth  q=quit"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            lumina.createElement("hbox", { style = { height = 1 } },
                lumina.createElement("text", {
                    foreground = t.muted or "#6C7086",
                    style = { height = 1 },
                }, "  "),
                Breadcrumb {
                    id = "nav",
                    items = bcItems,
                    onNavigate = function(id, index)
                        lumina.store.set("depth", index)
                        lumina.store.set("lastNav", id)
                    end,
                }
            ),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            lumina.createElement("text", {
                foreground = t.text or "#CDD6F4",
                style = { height = 1 },
            }, "  Page: " .. (currentPage and currentPage.content or "")),
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                style = { height = 1 },
            }, "  Last nav: " .. lastNav)
        )
    end,
}
