-- ============================================================================
-- Lumina Example: Form Builder
-- ============================================================================
-- Demonstrates: Form, Field, Input, Select, Checkbox, Toggle, Button components
-- Run: lumina examples/form-builder/main.lua
-- ============================================================================
local lumina = require("lumina")
local shadcn = require("shadcn")

lumina.setTheme("catppuccin-mocha")

local c = {
    bg = "#1E1E2E", fg = "#CDD6F4", accent = "#89B4FA",
    success = "#A6E3A1", error = "#F38BA8", muted = "#6C7086",
    surface = "#313244", border = "#45475A", warning = "#F9E2AF",
}

local store = lumina.createStore({
    state = {
        form = {
            firstName = "",
            lastName = "",
            email = "",
            role = "",
            notifications = false,
            darkMode = true,
            bio = "",
        },
        errors = {},
        submitted = false,
    },
})

local roles = {
    { label = "Developer", value = "developer" },
    { label = "Designer", value = "designer" },
    { label = "Manager", value = "manager" },
    { label = "Product", value = "product" },
}

local function InputField(props)
    local value = props.value or ""
    local error = props.error or ""
    local label = props.label or ""
    local placeholder = props.placeholder or ""
    local hasError = error ~= ""

    local display = value ~= "" and value or placeholder
    local fg = hasError and c.error or (value ~= "" and c.fg or c.muted)

    return {
        type = "vbox",
        children = {
            { type = "text", content = label, style = { foreground = hasError and c.error or c.fg, bold = true } },
            {
                type = "hbox",
                style = {
                    border = "rounded",
                    borderColor = hasError and c.error or c.border,
                    background = c.surface,
                    paddingLeft = 1, paddingRight = 1,
                },
                children = {
                    {
                        type = "text",
                        content = display,
                        style = { foreground = fg },
                    },
                    { type = "spacer" },
                },
            },
            hasError and { type = "text", content = "  " .. error, style = { foreground = c.error } } or nil,
        },
    }
end

local function CheckboxField(props)
    local checked = props.checked or false
    local label = props.label or ""

    return {
        type = "hbox",
        style = { align = "center" },
        children = {
            {
                type = "text",
                content = checked and "[✓]" or "[ ]",
                style = { foreground = checked and c.success or c.muted, bold = checked },
            },
            { type = "text", content = " " .. label, style = { foreground = c.fg } },
        },
    }
end

local function ToggleField(props)
    local checked = props.checked or false
    local label = props.label or ""

    return {
        type = "hbox",
        style = { align = "center" },
        children = {
            {
                type = "text",
                content = checked and "  [●━━] ON " or "  [━━●] OFF",
                style = { foreground = checked and c.success or c.muted, bold = true },
            },
            { type = "text", content = " " .. label, style = { foreground = c.fg } },
        },
    }
end

local function SelectField(props)
    local value = props.value or ""
    local label = props.label or ""
    local options = props.options or {}
    local error = props.error or ""
    local hasError = error ~= ""

    local display = value
    if display == "" then
        display = "Select..."
    end

    return {
        type = "vbox",
        children = {
            { type = "text", content = label, style = { foreground = hasError and c.error or c.fg, bold = true } },
            {
                type = "hbox",
                style = {
                    border = "rounded",
                    borderColor = hasError and c.error or c.border,
                    background = c.surface,
                    paddingLeft = 1, paddingRight = 1,
                },
                children = {
                    {
                        type = "text",
                        content = display,
                        style = { foreground = value ~= "" and c.fg or c.muted },
                    },
                    { type = "spacer" },
                    { type = "text", content = "▾", style = { foreground = c.muted } },
                },
            },
        },
    }
end

local function SectionHeader(title)
    return {
        type = "vbox",
        children = {
            { type = "text", content = "" },
            { type = "text", content = "  ▎ " .. title, style = { foreground = c.accent, bold = true } },
            { type = "text", content = "  " .. string.rep("─", 50), style = { foreground = c.border } },
        },
    }
end

local App = lumina.defineComponent({
    name = "FormBuilder",
    render = function(self)
        local state = lumina.useStore(store)
        local form = state.form or {}
        local errors = state.errors or {}
        local submitted = state.submitted or false

        local content

        if submitted then
            content = {
                { type = "text", content = "" },
                {
                    type = "vbox",
                    style = { border = "rounded", borderColor = c.success, padding = 1 },
                    children = {
                        { type = "text", content = "  ✓ Form submitted successfully!", style = { foreground = c.success, bold = true } },
                        { type = "text", content = "" },
                        { type = "text", content = "  Name: " .. (form.firstName or "") .. " " .. (form.lastName or ""), style = { foreground = c.fg } },
                        { type = "text", content = "  Email: " .. (form.email or ""), style = { foreground = c.fg } },
                        { type = "text", content = "  Role: " .. (form.role or ""), style = { foreground = c.fg } },
                        { type = "text", content = "  Notifications: " .. ((form.notifications and "Yes") or "No"), style = { foreground = c.fg } },
                        { type = "text", content = "" },
                        { type = "text", content = "  [r] Reset form", style = { foreground = c.muted } },
                    },
                },
            }
        else
            content = {
                SectionHeader("Personal Information"),
                InputField({
                    label = "First Name",
                    placeholder = "John",
                    value = form.firstName or "",
                    error = errors.firstName or "",
                }),
                { type = "text", content = "" },
                InputField({
                    label = "Last Name",
                    placeholder = "Doe",
                    value = form.lastName or "",
                    error = errors.lastName or "",
                }),
                { type = "text", content = "" },
                InputField({
                    label = "Email",
                    placeholder = "john@example.com",
                    value = form.email or "",
                    error = errors.email or "",
                }),
                { type = "text", content = "" },
                SectionHeader("Role"),
                SelectField({
                    label = "Job Role",
                    options = roles,
                    value = form.role or "",
                    error = errors.role or "",
                }),
                { type = "text", content = "" },
                SectionHeader("Preferences"),
                CheckboxField({
                    label = "Enable email notifications",
                    checked = form.notifications or false,
                }),
                { type = "text", content = "" },
                ToggleField({
                    label = "Dark mode",
                    checked = form.darkMode or false,
                }),
                { type = "text", content = "" },
                SectionHeader("Actions"),
                {
                    type = "hbox",
                    children = {
                        {
                            type = "text",
                            content = " [Submit] ",
                            style = { foreground = "#1E1E2E", background = c.accent, bold = true },
                        },
                        { type = "text", content = "  " },
                        {
                            type = "text",
                            content = " [Clear] ",
                            style = { foreground = c.fg, border = "rounded", borderColor = c.border },
                        },
                    },
                },
            }
        end

        return {
            type = "vbox",
            style = { background = c.bg },
            children = {
                { type = "text", content = " 📋 Form Builder ", style = { foreground = c.accent, bold = true, background = "#181825" } },
                { type = "text", content = " Build and validate forms with shadcn components ", style = { foreground = c.muted } },
                { type = "text", content = "" },
                { type = "vbox", style = { paddingLeft = 1 }, children = content },
                { type = "text", content = "" },
                { type = "text", content = " [t] toggle  [q] quit", style = { foreground = c.muted, dim = true } },
            },
        }
    end,
})

-- Validation
local function validateForm(form)
    local errors = {}
    if not form.firstName or form.firstName == "" then
        errors.firstName = "First name is required"
    end
    if not form.lastName or form.lastName == "" then
        errors.lastName = "Last name is required"
    end
    if not form.email or form.email == "" then
        errors.email = "Email is required"
    elseif not form.email:match("[^@]+@[^@]+") then
        errors.email = "Invalid email format"
    end
    if not form.role or form.role == "" then
        errors.role = "Please select a role"
    end
    return errors
end

lumina.onKey("s", function()
    local state = store.getState()
    local errors = validateForm(state.form)
    if next(errors) then
        store.dispatch("setState", { errors = errors })
    else
        store.dispatch("setState", { errors = {}, submitted = true })
    end
end)

lumina.onKey("c", function()
    store.dispatch("setState", {
        form = {
            firstName = "", lastName = "", email = "",
            role = "", notifications = false, darkMode = true, bio = "",
        },
        errors = {},
        submitted = false,
    })
end)

lumina.onKey("t", function()
    local state = store.getState()
    store.dispatch("setState", {
        form = {
            firstName = "Jane", lastName = "Smith",
            email = "jane@example.com", role = "developer",
            notifications = true, darkMode = true, bio = "",
        },
        errors = {},
        submitted = false,
    })
end)

lumina.onKey("q", function() lumina.quit() end)

lumina.mount(App)
lumina.run()
