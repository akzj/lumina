-- lux_widgets_test.lua — Lux widgets (Button is native Lux + CSS; others wrap Go)

test.describe("Lux widget wrappers", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Button tests

    test.it("Button renders with label", function()
        app:loadString([[
            local Button = require("lux.button")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Button { label = "Save", key = "btn1" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Save"), true)
    end)

    test.it("Button click fires onClick", function()
        app:loadString([[
            lumina.store.set("clicked", "no")
            local Button = require("lux.button")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local clicked = lumina.useStore("clicked")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Button {
                            label = "Go",
                            key = "btn1",
                            onClick = function()
                                lumina.store.set("clicked", "yes")
                            end,
                        },
                        lumina.createElement("text", {id = "out"},
                            "clicked:" .. clicked)
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("clicked:no"), true)
        app:click(3, 1)
        test.assert.eq(app:screenContains("clicked:yes"), true)
    end)

    test.it("Button disabled does not fire onClick", function()
        app:loadString([[
            lumina.store.set("fired", "no")
            local Button = require("lux.button")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local fired = lumina.useStore("fired")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Button {
                            label = "Nope",
                            key = "btn1",
                            disabled = true,
                            onClick = function()
                                lumina.store.set("fired", "yes")
                            end,
                        },
                        lumina.createElement("text", {id = "out"},
                            "fired:" .. fired)
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("fired:no"), true)
        app:click(3, 1)
        test.assert.eq(app:screenContains("fired:no"), true)
    end)

    -- Checkbox tests

    test.it("Checkbox renders with label", function()
        app:loadString([[
            local Checkbox = require("lux.checkbox")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Checkbox { label = "Accept Terms", key = "cb1" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Accept Terms"), true)
    end)

    test.it("Checkbox renders checked state", function()
        app:loadString([[
            local Checkbox = require("lux.checkbox")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Checkbox { label = "Checked", checked = true, key = "cb1" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("[x]"), true)
        test.assert.eq(app:screenContains("Checked"), true)
    end)

    test.it("Checkbox onChange fires on click", function()
        app:loadString([[
            lumina.store.set("val", "false")
            local Checkbox = require("lux.checkbox")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local val = lumina.useStore("val")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Checkbox {
                            label = "Toggle",
                            key = "cb1",
                            onChange = function(checked)
                                lumina.store.set("val", tostring(checked))
                            end,
                        },
                        lumina.createElement("text", {id = "out"},
                            "val:" .. val)
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("val:false"), true)
        app:click(2, 1)
        test.assert.eq(app:screenContains("val:true"), true)
    end)

    -- Radio tests

    test.it("Radio renders with label", function()
        app:loadString([[
            local Radio = require("lux.radio")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Radio { label = "Option A", value = "a", key = "r1" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Option A"), true)
    end)

    test.it("Radio renders checked state", function()
        app:loadString([[
            local Radio = require("lux.radio")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Radio { label = "Selected", value = "x", checked = true, key = "r1" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("(●)"), true)
        test.assert.eq(app:screenContains("Selected"), true)
    end)

    -- Switch tests

    test.it("Switch renders with label", function()
        app:loadString([[
            local Switch = require("lux.switch")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Switch { label = "Dark Mode", key = "sw1" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Dark Mode"), true)
    end)

    test.it("Switch renders checked state", function()
        app:loadString([[
            local Switch = require("lux.switch")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Switch { label = "On", checked = true, key = "sw1" }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("On"), true)
    end)

    -- Dropdown tests

    test.it("Dropdown renders with label", function()
        app:loadString([[
            local Dropdown = require("lux.dropdown")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Dropdown {
                            label = "Actions",
                            key = "dd1",
                            items = {
                                {label = "Cut"},
                                {label = "Copy"},
                                {label = "Paste"},
                            },
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Actions"), true)
    end)

    -- Umbrella access via require("lux")

    test.it("all wrappers accessible via lux umbrella", function()
        app:loadString([[
            local lux = require("lux")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        lux.Button { label = "LuxBtn", key = "b1" },
                        lux.Checkbox { label = "LuxCB", key = "c1" },
                        lux.Radio { label = "LuxRad", value = "r", key = "r1" },
                        lux.Switch { label = "LuxSw", key = "s1" },
                        lux.Dropdown { label = "LuxDD", key = "d1", items = {{label = "Item"}} }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("LuxBtn"), true)
        test.assert.eq(app:screenContains("LuxCB"), true)
        test.assert.eq(app:screenContains("LuxRad"), true)
        test.assert.eq(app:screenContains("LuxSw"), true)
        test.assert.eq(app:screenContains("LuxDD"), true)
    end)

    test.it("Lux Button.Group renders fused labels", function()
        app:loadString([[
            local Button = require("lux.button")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Button.Group {
                            key = "grp",
                            severity = "primary",
                            items = {
                                { label = "Save", onClick = function() end },
                                { label = "Del", onClick = function() end },
                            },
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Save"), true)
        test.assert.eq(app:screenContains("Del"), true)
    end)

    test.it("Lux Button.Split renders label and chevron", function()
        app:loadString([[
            local Button = require("lux.button")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Button.Split {
                            key = "spl",
                            label = "Options",
                            severity = "secondary",
                            onClick = function() end,
                            onMenuClick = function() end,
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Options"), true)
        test.assert.eq(app:screenContains("▼"), true)
    end)
end)
