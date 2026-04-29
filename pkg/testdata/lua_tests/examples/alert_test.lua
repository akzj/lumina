test.describe("Alert component", function()
    local app

    test.beforeEach(function()
        app = test.createApp(60, 24)
        app:loadFile("../examples/alert_demo.lua")
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Rendering variants
    test.it("shows info alert with icon", function()
        test.assert.eq(app:screenContains("Info"), true)
        test.assert.eq(app:screenContains("informational"), true)
    end)

    test.it("shows success alert", function()
        test.assert.eq(app:screenContains("Done"), true)
        test.assert.eq(app:screenContains("completed successfully"), true)
    end)

    test.it("shows warning alert with default title", function()
        test.assert.eq(app:screenContains("Warning"), true)
        test.assert.eq(app:screenContains("Disk space"), true)
    end)

    test.it("shows error alert", function()
        test.assert.eq(app:screenContains("Error"), true)
        test.assert.eq(app:screenContains("Connection failed"), true)
    end)

    test.it("shows dismissible alert with close button", function()
        test.assert.eq(app:screenContains("Dismissible"), true)
        test.assert.eq(app:screenContains("dismiss this"), true)
    end)

    -- Dismiss interaction
    test.it("dismiss hides the alert", function()
        test.assert.eq(app:screenContains("Dismissible"), true)
        app:loadString('lumina.store.set("showDismissible", false)')
        test.assert.eq(app:screenContains("Dismissible"), false)
    end)

    test.it("r key restores dismissed alert", function()
        app:loadString('lumina.store.set("showDismissible", false)')
        test.assert.eq(app:screenContains("Dismissible"), false)
        app:keyPress("r")
        test.assert.eq(app:screenContains("Dismissible"), true)
    end)

    -- Edge cases
    test.it("alert with no message shows only title", function()
        app:destroy()
        app = test.createApp(60, 10)
        app:loadString([[
            local Alert = require("lux.alert")
            lumina.app {
                id = "no-msg",
                render = function()
                    return Alert { variant = "info", title = "Title Only", width = 40 }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Title Only"), true)
    end)

    test.it("alert with no title uses variant as default title", function()
        app:destroy()
        app = test.createApp(60, 10)
        app:loadString([[
            local Alert = require("lux.alert")
            lumina.app {
                id = "no-title",
                render = function()
                    return Alert { variant = "error", message = "Something broke", width = 40 }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Error"), true)
        test.assert.eq(app:screenContains("Something broke"), true)
    end)

    test.it("alert with empty message", function()
        app:destroy()
        app = test.createApp(60, 10)
        app:loadString([[
            local Alert = require("lux.alert")
            lumina.app {
                id = "empty-msg",
                render = function()
                    return Alert { variant = "success", title = "OK", message = "", width = 40 }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("OK"), true)
    end)
end)
