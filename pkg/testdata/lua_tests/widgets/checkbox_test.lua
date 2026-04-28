-- checkbox_test.lua — Tests for the Checkbox widget (rendering + interaction)

test.describe("Checkbox widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders checked state with label", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.Checkbox {
                            label = "Option A",
                            checked = true,
                            key = "cb1",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("[x]"), true)
        test.assert.eq(app:screenContains("Option A"), true)
    end)

    test.it("renders unchecked state", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.Checkbox {
                            label = "Option B",
                            checked = false,
                            key = "cb2",
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("[ ]"), true)
        test.assert.eq(app:screenContains("Option B"), true)
    end)

    test.it("toggles on click via onChange", function()
        app:loadString([[
            lumina.store.set("checked", false)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local checked = lumina.useStore("checked")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Checkbox, {
                            label = "Toggle Me",
                            checked = checked,
                            key = "cb3",
                            onChange = function(val)
                                lumina.store.set("checked", val)
                            end,
                        }),
                        lumina.createElement("text", {id = "status"},
                            "checked:" .. tostring(checked))
                    )
                end,
            })
        ]])
        -- Initial: unchecked
        test.assert.eq(app:screenContains("[ ]"), true)
        test.assert.eq(app:screenContains("checked:false"), true)

        -- Click the checkbox area to toggle
        app:click(1, 0)
        test.assert.eq(app:screenContains("[x]"), true)
        test.assert.eq(app:screenContains("checked:true"), true)

        -- Click again to uncheck
        app:click(1, 0)
        test.assert.eq(app:screenContains("[ ]"), true)
        test.assert.eq(app:screenContains("checked:false"), true)
    end)

    test.it("disabled checkbox does not toggle on click", function()
        app:loadString([[
            lumina.store.set("checked", false)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local checked = lumina.useStore("checked")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lumina.createElement(lumina.Checkbox, {
                            label = "Disabled",
                            checked = checked,
                            disabled = true,
                            key = "cb4",
                            onChange = function(val)
                                lumina.store.set("checked", val)
                            end,
                        }),
                        lumina.createElement("text", {id = "status"},
                            "checked:" .. tostring(checked))
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("[ ]"), true)
        test.assert.eq(app:screenContains("checked:false"), true)

        -- Click should NOT toggle
        app:click(1, 0)
        test.assert.eq(app:screenContains("[ ]"), true)
        test.assert.eq(app:screenContains("checked:false"), true)
    end)
end)
