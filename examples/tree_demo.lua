-- examples/tree_demo.lua — Lux Tree component demo
-- Usage: lumina examples/tree_demo.lua
-- Quit: q

local lux = require("lux")
local Tree = lux.Tree

local fileTree = {
    { id = "src", label = "src", children = {
        { id = "src/main.go", label = "main.go" },
        { id = "src/app.go", label = "app.go" },
        { id = "src/render", label = "render", children = {
            { id = "src/render/engine.go", label = "engine.go" },
            { id = "src/render/layout.go", label = "layout.go" },
            { id = "src/render/paint.go", label = "paint.go" },
        }},
        { id = "src/widget", label = "widget", children = {
            { id = "src/widget/button.go", label = "button.go" },
            { id = "src/widget/input.go", label = "input.go" },
        }},
    }},
    { id = "lua", label = "lua", children = {
        { id = "lua/init.lua", label = "init.lua" },
        { id = "lua/lux", label = "lux", children = {
            { id = "lua/lux/toast.lua", label = "toast.lua" },
            { id = "lua/lux/tree.lua", label = "tree.lua" },
        }},
    }},
    { id = "docs", label = "docs", disabled = true, children = {
        { id = "docs/README.md", label = "README.md" },
    }},
    { id = "go.mod", label = "go.mod" },
    { id = "go.sum", label = "go.sum" },
}

lumina.app {
    id = "tree-demo",
    store = {
        expandedIds = { "src" },
        selectedId = "src",
        activated = "",
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
    },
    render = function()
        local t = lumina.getTheme()
        local expandedIds = lumina.useStore("expandedIds") or {}
        local selectedId = lumina.useStore("selectedId") or ""
        local activated = lumina.useStore("activated") or ""

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 60, height = 24, background = t.base or "#1E1E2E" },
        },
            lumina.createElement("text", {
                bold = true, foreground = t.primary or "#89B4FA",
                style = { height = 1 },
            }, "  Tree Demo"),
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                style = { height = 1 },
            }, "  j/k=nav  l=expand  h=collapse  Enter=toggle/activate  q=quit"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
            Tree {
                key = "file-tree",
                items = fileTree,
                expandedIds = expandedIds,
                selectedId = selectedId,
                indent = 2,
                width = 50,
                autoFocus = true,
                onToggle = function(id, expanded, newIds)
                    lumina.store.set("expandedIds", newIds)
                end,
                onSelect = function(id)
                    lumina.store.set("selectedId", id)
                end,
                onActivate = function(id)
                    lumina.store.set("activated", id)
                end,
            },
            lumina.createElement("text", { style = { height = 1 } }, ""),
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                style = { height = 1 },
            }, activated ~= "" and ("  Activated: " .. activated) or "")
        )
    end,
}
