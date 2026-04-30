-- tree_test.lua — Tests for Tree component

test.describe("Tree component", function()
    local app

    test.beforeEach(function()
        app = test.createApp(60, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Rendering
    test.it("renders top-level items", function()
        app:loadString([[
            local Tree = require("lux.tree")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Tree {
                            key = "t",
                            items = {
                                { id = "a", label = "Alpha" },
                                { id = "b", label = "Beta" },
                                { id = "c", label = "Gamma" },
                            },
                            width = 50,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Alpha"), true)
        test.assert.eq(app:screenContains("Beta"), true)
        test.assert.eq(app:screenContains("Gamma"), true)
    end)

    test.it("expanded nodes show children", function()
        app:loadString([[
            local Tree = require("lux.tree")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Tree {
                            key = "t",
                            items = {
                                { id = "folder", label = "Folder", children = {
                                    { id = "file1", label = "file1.txt" },
                                    { id = "file2", label = "file2.txt" },
                                }},
                            },
                            expandedIds = { "folder" },
                            width = 50,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Folder"), true)
        test.assert.eq(app:screenContains("file1.txt"), true)
        test.assert.eq(app:screenContains("file2.txt"), true)
    end)

    test.it("collapsed nodes hide children", function()
        app:loadString([[
            local Tree = require("lux.tree")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Tree {
                            key = "t",
                            items = {
                                { id = "folder", label = "Folder", children = {
                                    { id = "file1", label = "hidden_file.txt" },
                                }},
                            },
                            expandedIds = {},
                            width = 50,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Folder"), true)
        test.assert.eq(app:screenContains("hidden_file"), false)
    end)

    test.it("keyboard j/k navigates between nodes", function()
        app:loadString([[
            lumina.store.set("selectedId", "a")
            local Tree = require("lux.tree")
            lumina.app {
                id = "test",
                render = function()
                    local sel = lumina.useStore("selectedId") or "a"
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Tree {
                            key = "t",
                            items = {
                                { id = "a", label = "NodeA" },
                                { id = "b", label = "NodeB" },
                                { id = "c", label = "NodeC" },
                            },
                            selectedId = sel,
                            autoFocus = true,
                            width = 50,
                            onSelect = function(id)
                                lumina.store.set("selectedId", id)
                            end,
                        },
                        lumina.createElement("text", {id = "sel"}, "sel:" .. sel)
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("sel:a"), true)
        app:keyPress("j")
        test.assert.eq(app:screenContains("sel:b"), true)
        app:keyPress("j")
        test.assert.eq(app:screenContains("sel:c"), true)
        app:keyPress("k")
        test.assert.eq(app:screenContains("sel:b"), true)
    end)

    test.it("l expands collapsed node", function()
        app:loadString([[
            lumina.store.set("selectedId", "folder")
            lumina.store.set("expanded", {})
            local Tree = require("lux.tree")
            lumina.app {
                id = "test",
                render = function()
                    local sel = lumina.useStore("selectedId") or "folder"
                    local expanded = lumina.useStore("expanded") or {}
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Tree {
                            key = "t",
                            items = {
                                { id = "folder", label = "MyFolder", children = {
                                    { id = "child", label = "child.txt" },
                                }},
                            },
                            expandedIds = expanded,
                            selectedId = sel,
                            autoFocus = true,
                            width = 50,
                            onSelect = function(id)
                                lumina.store.set("selectedId", id)
                            end,
                            onToggle = function(id, exp, newIds)
                                lumina.store.set("expanded", newIds)
                            end,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("MyFolder"), true)
        test.assert.eq(app:screenContains("child.txt"), false)
        app:keyPress("l")
        test.assert.eq(app:screenContains("child.txt"), true)
    end)

    test.it("h collapses expanded node", function()
        app:loadString([[
            lumina.store.set("selectedId", "folder")
            lumina.store.set("expanded", {"folder"})
            local Tree = require("lux.tree")
            lumina.app {
                id = "test",
                render = function()
                    local sel = lumina.useStore("selectedId") or "folder"
                    local expanded = lumina.useStore("expanded") or {}
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Tree {
                            key = "t",
                            items = {
                                { id = "folder", label = "MyFolder", children = {
                                    { id = "child", label = "child.txt" },
                                }},
                            },
                            expandedIds = expanded,
                            selectedId = sel,
                            autoFocus = true,
                            width = 50,
                            onSelect = function(id)
                                lumina.store.set("selectedId", id)
                            end,
                            onToggle = function(id, exp, newIds)
                                lumina.store.set("expanded", newIds)
                            end,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("child.txt"), true)
        app:keyPress("h")
        test.assert.eq(app:screenContains("child.txt"), false)
    end)

    test.it("Enter toggles expand on branch nodes", function()
        app:loadString([[
            lumina.store.set("selectedId", "folder")
            lumina.store.set("expanded", {})
            local Tree = require("lux.tree")
            lumina.app {
                id = "test",
                render = function()
                    local sel = lumina.useStore("selectedId") or "folder"
                    local expanded = lumina.useStore("expanded") or {}
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Tree {
                            key = "t",
                            items = {
                                { id = "folder", label = "MyFolder", children = {
                                    { id = "child", label = "child.txt" },
                                }},
                            },
                            expandedIds = expanded,
                            selectedId = sel,
                            autoFocus = true,
                            width = 50,
                            onSelect = function(id)
                                lumina.store.set("selectedId", id)
                            end,
                            onToggle = function(id, exp, newIds)
                                lumina.store.set("expanded", newIds)
                            end,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("child.txt"), false)
        app:keyPress("Enter")
        test.assert.eq(app:screenContains("child.txt"), true)
        app:keyPress("Enter")
        test.assert.eq(app:screenContains("child.txt"), false)
    end)

    test.it("Enter calls onActivate on leaf nodes", function()
        app:loadString([[
            lumina.store.set("selectedId", "leaf")
            lumina.store.set("activated", "")
            local Tree = require("lux.tree")
            lumina.app {
                id = "test",
                render = function()
                    local sel = lumina.useStore("selectedId") or "leaf"
                    local activated = lumina.useStore("activated") or ""
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Tree {
                            key = "t",
                            items = {
                                { id = "leaf", label = "MyLeaf" },
                            },
                            selectedId = sel,
                            autoFocus = true,
                            width = 50,
                            onSelect = function(id)
                                lumina.store.set("selectedId", id)
                            end,
                            onActivate = function(id)
                                lumina.store.set("activated", id)
                            end,
                        },
                        lumina.createElement("text", {id = "out"}, "act:" .. activated)
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("act:"), true)
        app:keyPress("Enter")
        test.assert.eq(app:screenContains("act:leaf"), true)
    end)

    test.it("disabled nodes are skipped during navigation", function()
        app:loadString([[
            lumina.store.set("selectedId", "a")
            local Tree = require("lux.tree")
            lumina.app {
                id = "test",
                render = function()
                    local sel = lumina.useStore("selectedId") or "a"
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Tree {
                            key = "t",
                            items = {
                                { id = "a", label = "First" },
                                { id = "b", label = "Disabled", disabled = true },
                                { id = "c", label = "Third" },
                            },
                            selectedId = sel,
                            autoFocus = true,
                            width = 50,
                            onSelect = function(id)
                                lumina.store.set("selectedId", id)
                            end,
                        },
                        lumina.createElement("text", {id = "sel"}, "sel:" .. sel)
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("sel:a"), true)
        app:keyPress("j")
        -- Should skip disabled "b" and go to "c"
        test.assert.eq(app:screenContains("sel:c"), true)
    end)

    test.it("deep nesting works (3+ levels)", function()
        app:loadString([[
            local Tree = require("lux.tree")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Tree {
                            key = "t",
                            items = {
                                { id = "l1", label = "Level1", children = {
                                    { id = "l2", label = "Level2", children = {
                                        { id = "l3", label = "Level3", children = {
                                            { id = "l4", label = "DeepLeaf" },
                                        }},
                                    }},
                                }},
                            },
                            expandedIds = {"l1", "l2", "l3"},
                            width = 50,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Level1"), true)
        test.assert.eq(app:screenContains("Level2"), true)
        test.assert.eq(app:screenContains("Level3"), true)
        test.assert.eq(app:screenContains("DeepLeaf"), true)
    end)

    -- Demo file test
    test.it("demo loads and renders", function()
        app:destroy()
        app = test.createApp(60, 24)
        app:loadFile("../examples/tree_demo.lua")
        test.assert.eq(app:screenContains("Tree Demo"), true)
        test.assert.eq(app:screenContains("src"), true)
    end)

    test.it("demo expands src by default and shows children", function()
        app:destroy()
        app = test.createApp(60, 24)
        app:loadFile("../examples/tree_demo.lua")
        test.assert.eq(app:screenContains("main.go"), true)
        test.assert.eq(app:screenContains("app.go"), true)
    end)

    -- Accessible via lux umbrella
    test.it("accessible via require('lux').Tree", function()
        app:loadString([[
            local lux = require("lux")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        lux.Tree {
                            key = "t",
                            items = {{ id = "x", label = "LuxTree" }},
                            width = 50,
                        }
                    )
                end,
            }
        ]])
        test.assert.eq(app:screenContains("LuxTree"), true)
    end)
end)
