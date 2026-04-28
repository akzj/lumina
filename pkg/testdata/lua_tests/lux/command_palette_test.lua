test.describe("CommandPalette", function()
    test.it("renders with commands", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            local CommandPalette = require("lux.command_palette")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {},
                        CommandPalette {
                            commands = {
                                { title = "Save File", action = function() end },
                                { title = "Open File", action = function() end },
                                { title = "Quit", action = function() end },
                            },
                        }
                    )
                end,
            })
        ]])
        test.assert.notNil(app:find("command-palette"))
        test.assert.notNil(app:find("cp-input"))
        app:destroy()
    end)

    test.it("shows all commands initially", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            local CommandPalette = require("lux.command_palette")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {},
                        CommandPalette {
                            commands = {
                                { title = "Alpha", action = function() end },
                                { title = "Beta", action = function() end },
                                { title = "Gamma", action = function() end },
                            },
                        }
                    )
                end,
            })
        ]])
        local text = app:screenText()
        test.assert.contains(text, "Alpha")
        test.assert.contains(text, "Beta")
        test.assert.contains(text, "Gamma")
        app:destroy()
    end)

    test.it("first item selected by default", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            local CommandPalette = require("lux.command_palette")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {},
                        CommandPalette {
                            commands = {
                                { title = "First", action = function() end },
                                { title = "Second", action = function() end },
                            },
                        }
                    )
                end,
            })
        ]])
        local cmd1 = app:find("cmd-1")
        test.assert.notNil(cmd1)
        test.assert.contains(cmd1.content, "> First")
        app:destroy()
    end)

    test.it("via lux umbrella module", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            local lux = require("lux")
            local CP = lux.CommandPalette

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {},
                        CP {
                            commands = {
                                { title = "Test Cmd", action = function() end },
                            },
                        }
                    )
                end,
            })
        ]])
        local text = app:screenText()
        test.assert.contains(text, "Test Cmd")
        app:destroy()
    end)
end)
