-- examples/text_input_demo.lua — Lux TextInput component demo
-- Usage: lumina examples/text_input_demo.lua
-- Quit: q

local TextInput = require("lux.text_input")

local _pendingInput = ""

lumina.app {
    id = "textinput-demo",
    store = {
        name = "",
        email = "",
        submitted = false,
        errorField = "",
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
    },
    render = function()
        local t = lumina.getTheme()
        local name = lumina.useStore("name")
        local email = lumina.useStore("email")
        local submitted = lumina.useStore("submitted")
        local errorField = lumina.useStore("errorField")

        local children = {
            lumina.createElement("text", {
                bold = true, foreground = t.primary or "#89B4FA",
                style = { height = 1 },
            }, "  TextInput Demo"),
            lumina.createElement("text", {
                foreground = t.muted or "#6C7086",
                style = { height = 1 },
            }, "  Tab to switch fields, Enter to submit"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
        }

        -- Name field
        children[#children + 1] = TextInput {
            key = "name-input",
            id = "name-field",
            label = "Name",
            placeholder = "Enter your name",
            value = name,
            width = 40,
            autoFocus = true,
            helperText = "Your full name",
            error = (errorField == "name") and "Name is required" or nil,
            onChange = function(val)
                lumina.store.set("name", val)
            end,
            onSubmit = function()
                if name == "" then
                    lumina.store.set("errorField", "name")
                else
                    lumina.store.set("submitted", true)
                    lumina.store.set("errorField", "")
                end
            end,
        }

        children[#children + 1] = lumina.createElement("text", { style = { height = 1 } }, "")

        -- Email field
        children[#children + 1] = TextInput {
            key = "email-input",
            id = "email-field",
            label = "Email",
            placeholder = "user@example.com",
            value = email,
            width = 40,
            helperText = "We won't share your email",
            error = (errorField == "email") and "Invalid email" or nil,
            onChange = function(val)
                lumina.store.set("email", val)
            end,
        }

        children[#children + 1] = lumina.createElement("text", { style = { height = 1 } }, "")

        -- Disabled field
        children[#children + 1] = TextInput {
            key = "disabled-input",
            id = "disabled-field",
            label = "Disabled",
            value = "Cannot edit this",
            width = 40,
            disabled = true,
        }

        -- Submitted message
        if submitted then
            children[#children + 1] = lumina.createElement("text", { style = { height = 1 } }, "")
            children[#children + 1] = lumina.createElement("text", {
                foreground = t.success or "#A6E3A1",
                style = { height = 1 },
            }, "  Submitted: " .. name)
        end

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 60, height = 24, background = t.base or "#1E1E2E" },
        }, table.unpack(children))
    end,
}
