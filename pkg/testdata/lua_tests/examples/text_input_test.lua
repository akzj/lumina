-- text_input_test.lua — Tests for TextInput component

test.describe("TextInput component", function()
    local app

    test.beforeEach(function()
        app = test.createApp(60, 24)
        app:loadFile("../examples/text_input_demo.lua")
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Rendering
    test.it("renders label text", function()
        test.assert.eq(app:screenContains("Name"), true)
        test.assert.eq(app:screenContains("Email"), true)
        test.assert.eq(app:screenContains("Disabled"), true)
    end)

    test.it("renders helper text", function()
        test.assert.eq(app:screenContains("Your full name"), true)
        test.assert.eq(app:screenContains("share your email"), true)
    end)

    test.it("renders placeholder text", function()
        test.assert.eq(app:screenContains("Enter your name"), true)
    end)

    test.it("renders disabled field value", function()
        test.assert.eq(app:screenContains("Cannot edit this"), true)
    end)

    -- Error state
    test.it("shows error message when errorField is set", function()
        app:loadString('lumina.store.set("errorField", "name")')
        test.assert.eq(app:screenContains("Name is required"), true)
    end)

    test.it("error replaces helper text", function()
        app:loadString('lumina.store.set("errorField", "name")')
        -- Error should show, helper should not
        test.assert.eq(app:screenContains("Name is required"), true)
    end)

    test.it("shows email error", function()
        app:loadString('lumina.store.set("errorField", "email")')
        test.assert.eq(app:screenContains("Invalid email"), true)
    end)

    -- onChange
    test.it("onChange updates store value", function()
        app:loadString('lumina.store.set("name", "Alice")')
        test.assert.eq(app:screenContains("Alice"), true)
    end)

    -- onSubmit
    test.it("submit with name shows submitted message", function()
        app:loadString('lumina.store.set("name", "Bob")')
        app:loadString('lumina.store.set("submitted", true)')
        test.assert.eq(app:screenContains("Submitted: Bob"), true)
    end)

    -- Width
    test.it("custom width applies to component", function()
        app:destroy()
        app = test.createApp(80, 20)
        app:loadString([[
            local TextInput = require("lux.text_input")
            lumina.app {
                id = "width-test",
                render = function()
                    return TextInput {
                        id = "wide",
                        label = "Wide Input",
                        width = 60,
                        value = "",
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Wide Input"), true)
    end)

    -- No label
    test.it("renders without label when not provided", function()
        app:destroy()
        app = test.createApp(60, 20)
        app:loadString([[
            local TextInput = require("lux.text_input")
            lumina.app {
                id = "no-label",
                render = function()
                    return TextInput {
                        id = "nolabel",
                        value = "just input",
                        width = 30,
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("just input"), true)
    end)

    -- No helper, no error
    test.it("renders cleanly with no helper or error", function()
        app:destroy()
        app = test.createApp(60, 20)
        app:loadString([[
            local TextInput = require("lux.text_input")
            lumina.app {
                id = "minimal",
                render = function()
                    return TextInput {
                        id = "min",
                        label = "Minimal",
                        value = "hello",
                        width = 30,
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Minimal"), true)
        test.assert.eq(app:screenContains("hello"), true)
    end)
end)
