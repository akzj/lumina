test.describe("Toast debug", function()
    local app

    test.beforeEach(function()
        app = test.createApp(60, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("debug render", function()
        app:loadString([[
            local Toast = require("lux.toast")
            lumina.app {
                id = "test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 60, height = 24}},
                        Toast {
                            key = "t",
                            items = {
                                { id = 1, message = "Hello World", variant = "info" },
                            },
                            width = 50,
                        }
                    )
                end,
            }
        ]])
        test.log("SCREEN: " .. app:screenText())
        test.assert.eq(app:screenContains("Hello World"), true)
    end)
end)
