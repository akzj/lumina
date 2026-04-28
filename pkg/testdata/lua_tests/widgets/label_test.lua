-- label_test.lua — Tests for the Label widget

test.describe("Label widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders with text prop", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Label, {
                            text = "Hello World",
                            key = "lbl1",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Hello World"), true)
    end)

    test.it("renders default text when no prop given", function()
        app:loadString([[
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement(lumina.Label, {
                            key = "lbl2",
                        })
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("Label"), true)
    end)
end)
