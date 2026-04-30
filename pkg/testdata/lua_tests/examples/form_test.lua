-- form_test.lua — Tests for Form component

test.describe("Form component", function()
    local app

    test.beforeEach(function()
        app = test.createApp(60, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Rendering
    test.it("renders field labels", function()
        app:loadString([[
            local Form = require("lux.form")
            lumina.app {
                id = "test",
                render = function()
                    return Form {
                        key = "f",
                        fields = {
                            { id = "name", type = "text", label = "Name" },
                            { id = "email", type = "text", label = "Email" },
                        },
                        values = { name = "", email = "" },
                        errors = {},
                        width = 40,
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Name"), true)
        test.assert.eq(app:screenContains("Email"), true)
    end)

    test.it("renders required indicator", function()
        app:loadString([[
            local Form = require("lux.form")
            lumina.app {
                id = "test",
                render = function()
                    return Form {
                        key = "f",
                        fields = {
                            { id = "name", type = "text", label = "Name", required = true },
                            { id = "opt", type = "text", label = "Optional" },
                        },
                        values = {},
                        errors = {},
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Name *"), true)
        test.assert.eq(app:screenContains("Optional"), true)
    end)

    test.it("renders input placeholders", function()
        app:loadString([[
            local Form = require("lux.form")
            lumina.app {
                id = "test",
                render = function()
                    return Form {
                        key = "f",
                        fields = {
                            { id = "name", type = "text", label = "Name", placeholder = "Your name" },
                        },
                        values = {},
                        errors = {},
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Your name"), true)
    end)

    test.it("shows error messages", function()
        app:loadString([[
            local Form = require("lux.form")
            lumina.app {
                id = "test",
                render = function()
                    return Form {
                        key = "f",
                        fields = {
                            { id = "name", type = "text", label = "Name" },
                            { id = "email", type = "text", label = "Email" },
                        },
                        values = { name = "", email = "" },
                        errors = { name = "Name is required", email = "Invalid email" },
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Name is required"), true)
        test.assert.eq(app:screenContains("Invalid email"), true)
    end)

    test.it("renders checkbox fields", function()
        app:loadString([[
            local Form = require("lux.form")
            lumina.app {
                id = "test",
                render = function()
                    return Form {
                        key = "f",
                        fields = {
                            { id = "terms", type = "checkbox", label = "I accept the terms" },
                        },
                        values = { terms = false },
                        errors = {},
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("[ ] I accept the terms"), true)
    end)

    test.it("checkbox shows checked state", function()
        app:loadString([[
            local Form = require("lux.form")
            lumina.app {
                id = "test",
                render = function()
                    return Form {
                        key = "f",
                        fields = {
                            { id = "terms", type = "checkbox", label = "I accept the terms" },
                        },
                        values = { terms = true },
                        errors = {},
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("[x] I accept the terms"), true)
    end)

    test.it("renders submit button", function()
        app:loadString([[
            local Form = require("lux.form")
            lumina.app {
                id = "test",
                render = function()
                    return Form {
                        key = "f",
                        fields = {},
                        values = {},
                        errors = {},
                        onSubmit = function() end,
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("[Submit]"), true)
    end)

    test.it("renders reset button when onReset provided", function()
        app:loadString([[
            local Form = require("lux.form")
            lumina.app {
                id = "test",
                render = function()
                    return Form {
                        key = "f",
                        fields = {},
                        values = {},
                        errors = {},
                        onSubmit = function() end,
                        onReset = function() end,
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("[Reset]"), true)
    end)

    test.it("custom submit label", function()
        app:loadString([[
            local Form = require("lux.form")
            lumina.app {
                id = "test",
                render = function()
                    return Form {
                        key = "f",
                        fields = {},
                        values = {},
                        errors = {},
                        submitLabel = "Save",
                        onSubmit = function() end,
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("[Save]"), true)
    end)

    test.it("renders field values in inputs", function()
        app:loadString([[
            local Form = require("lux.form")
            lumina.app {
                id = "test",
                render = function()
                    return Form {
                        key = "f",
                        fields = {
                            { id = "name", type = "text", label = "Name" },
                        },
                        values = { name = "Alice" },
                        errors = {},
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Alice"), true)
    end)

    test.it("accessible via require lux", function()
        app:loadString([[
            local lux = require("lux")
            lumina.app {
                id = "test",
                render = function()
                    return lux.Form {
                        key = "f",
                        fields = {
                            { id = "x", type = "text", label = "Field X" },
                        },
                        values = {},
                        errors = {},
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Field X"), true)
    end)

    test.it("disabled form shows submit button text", function()
        app:loadString([[
            local Form = require("lux.form")
            lumina.app {
                id = "test",
                render = function()
                    return Form {
                        key = "f",
                        disabled = true,
                        fields = {
                            { id = "name", type = "text", label = "Name" },
                        },
                        values = { name = "Bob" },
                        errors = {},
                        onSubmit = function() end,
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("Name"), true)
        test.assert.eq(app:screenContains("[Submit]"), true)
    end)
end)
