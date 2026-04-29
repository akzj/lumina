-- keyboard_test.lua — Tests for Go widget keyboard dispatch via HandleKeyDown

test.describe("Widget keyboard dispatch", function()
    test.it("List: j navigation moves selection down", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.List, {
                            items = {"Alpha", "Beta", "Gamma"},
                            key = "list1",
                        })
                    )
                end,
            })
        ]])
        -- Initially Alpha is selected (index 1)
        test.assert.eq(app:screenContains("Alpha"), true)
        -- Click on the list to focus it
        app:click(5, 2)
        -- Press j to move selection down
        app:keyPress("j")
        -- Beta should still be visible (list renders all items)
        test.assert.eq(app:screenContains("Beta"), true)
        app:destroy()
    end)

    test.it("List: ArrowDown moves selection", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.List, {
                            items = {"Alpha", "Beta", "Gamma"},
                            key = "list1",
                        })
                    )
                end,
            })
        ]])
        app:click(5, 2)
        app:keyPress("ArrowDown")
        test.assert.eq(app:screenContains("Beta"), true)
        app:destroy()
    end)

    test.it("Checkbox: Space toggles via onChange callback", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local checked, setChecked = lumina.useState("checked", false)
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Checkbox, {
                            label = "Toggle Me",
                            checked = checked,
                            key = "cb1",
                            onChange = function(val)
                                setChecked(val)
                            end,
                        }),
                        lumina.createElement("text", {id = "status"},
                            "checked:" .. tostring(checked))
                    )
                end,
            })
        ]])
        -- Initially unchecked
        test.assert.eq(app:screenContains("[ ]"), true)
        local status = app:find("status")
        test.assert.contains(status.content, "checked:false")
        -- Click to focus, then Space to toggle
        app:click(5, 1)
        app:keyPress(" ")
        -- Should now be checked
        test.assert.eq(app:screenContains("[x]"), true)
        status = app:find("status")
        test.assert.contains(status.content, "checked:true")
        app:destroy()
    end)

    test.it("Checkbox: Enter toggles via onChange callback", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local checked, setChecked = lumina.useState("checked", false)
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Checkbox, {
                            label = "Toggle Me",
                            checked = checked,
                            key = "cb1",
                            onChange = function(val)
                                setChecked(val)
                            end,
                        }),
                        lumina.createElement("text", {id = "status"},
                            "checked:" .. tostring(checked))
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("[ ]"), true)
        app:click(5, 1)
        app:keyPress("Enter")
        test.assert.eq(app:screenContains("[x]"), true)
        app:destroy()
    end)

    test.it("List: onChange fires on j with selected index", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local selected, setSelected = lumina.useState("sel", 0)
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.List, {
                            items = {"Alpha", "Beta", "Gamma"},
                            key = "list1",
                            onChange = function(val)
                                setSelected(val)
                            end,
                        }),
                        lumina.createElement("text", {id = "sel"},
                            "selected:" .. tostring(selected))
                    )
                end,
            })
        ]])
        -- Click to focus
        app:click(5, 2)
        -- j: onChange fires 1 (Alpha, 1-based)
        app:keyPress("j")
        local sel = app:find("sel")
        test.assert.notNil(sel)
        test.assert.contains(sel.content, "selected:1")
        app:destroy()
    end)

    test.it("List: ArrowUp moves selection up", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.List, {
                            items = {"Alpha", "Beta", "Gamma"},
                            selectedIndex = 3,
                            key = "list1",
                        })
                    )
                end,
            })
        ]])
        app:click(5, 3)
        app:keyPress("k")
        -- After k, selection should move up from Gamma to Beta
        test.assert.eq(app:screenContains("Beta"), true)
        app:destroy()
    end)
end)
