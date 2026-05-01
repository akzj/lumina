-- form_widgets_demo.lua — Demo of Lux widget wrappers in a form layout
-- Run: go run ./cmd/lumina examples/form_widgets_demo.lua

local lux = require("lux")

lumina.createComponent({
    id = "form-demo", name = "FormWidgetsDemo",
    render = function()
        local theme = lumina.getTheme and lumina.getTheme() or {}
        local darkMode, setDarkMode = lumina.useState("darkMode", false)
        local agree, setAgree = lumina.useState("agree", false)
        local notify, setNotify = lumina.useState("notify", true)
        local plan, setPlan = lumina.useState("plan", "free")
        local submitted, setSubmitted = lumina.useState("submitted", false)

        return lumina.createElement("vbox", {
            id = "root",
            style = {
                padding = 1,
                gap = 1,
            },
        },
            -- Title
            lumina.createElement("text", {
                style = { bold = true, foreground = theme.primary or "#89B4FA" },
            }, "╭─ Form Widgets Demo ─╮"),

            -- Dark mode toggle (Switch)
            lux.Switch {
                key = "dark-mode",
                label = "Dark Mode",
                checked = darkMode,
                onChange = function(val) setDarkMode(val) end,
            },

            -- Notifications toggle (Switch)
            lux.Switch {
                key = "notifications",
                label = "Enable Notifications",
                checked = notify,
                onChange = function(val) setNotify(val) end,
            },

            -- Terms agreement (Checkbox)
            lux.Checkbox {
                key = "agree",
                label = "I agree to the Terms of Service",
                checked = agree,
                onChange = function(val) setAgree(val) end,
            },

            -- Plan selection (Radio buttons)
            lumina.createElement("text", {
                style = { foreground = theme.text or "#CDD6F4", bold = true },
            }, "Select Plan:"),

            lux.Radio {
                key = "plan-free",
                label = "Free",
                value = "free",
                checked = (plan == "free"),
                onChange = function() setPlan("free") end,
            },
            lux.Radio {
                key = "plan-pro",
                label = "Pro ($9/mo)",
                value = "pro",
                checked = (plan == "pro"),
                onChange = function() setPlan("pro") end,
            },
            lux.Radio {
                key = "plan-enterprise",
                label = "Enterprise ($49/mo)",
                value = "enterprise",
                checked = (plan == "enterprise"),
                onChange = function() setPlan("enterprise") end,
            },

            -- Actions dropdown
            lumina.Dropdown {
                key = "actions",
                label = "More Actions",
                items = {
                    { label = "Export as CSV" },
                    { label = "Print" },
                    { label = "Share" },
                },
                onChange = function(idx)
                    -- handle action selection
                end,
            },

            -- Submit button
            lux.Button {
                key = "submit",
                label = submitted and "✓ Submitted" or "Submit",
                variant = submitted and "secondary" or "primary",
                disabled = not agree,
                onClick = function()
                    if agree then setSubmitted(true) end
                end,
            },

            -- Status line
            lumina.createElement("text", {
                style = { foreground = theme.muted or "#6C7086" },
            }, string.format(
                "Dark: %s | Notify: %s | Plan: %s | Agreed: %s",
                tostring(darkMode), tostring(notify), plan, tostring(agree)
            ))
        )
    end,
})
