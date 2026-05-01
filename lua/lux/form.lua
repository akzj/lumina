-- lua/lux/form.lua — Lux Form: composable form with validation
-- Usage: local Form = require("lux.form")
--
-- Props:
--   fields: array of field definitions
--     { id, type, label, placeholder?, required?, validate?, defaultValue? }
--     type: "text" | "checkbox" | "select"
--   values: table { [fieldId] = value }
--   errors: table { [fieldId] = errorMessage }
--   onFieldChange: function(fieldId, newValue)
--   onSubmit: function(values)
--   onReset: function()
--   submitLabel: string (default "Submit")
--   resetLabel: string (default "Reset")
--   width: number (default 40)
--   disabled: boolean

local Form = lumina.defineComponent("LuxForm", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local fields = props.fields or {}
    local values = props.values or {}
    local errors = props.errors or {}
    local onFieldChange = props.onFieldChange
    local onSubmit = props.onSubmit
    local onReset = props.onReset
    local width = props.width or 40
    local disabled = props.disabled

    local children = {}

    -- Render each field
    for _, field in ipairs(fields) do
        local fieldId = field.id
        local value = values[fieldId]
        local err = errors[fieldId]
        local fieldType = field.type or "text"

        if fieldType == "text" then
            -- Label
            if field.label then
                local labelText = field.label
                if field.required then labelText = labelText .. " *" end
                children[#children + 1] = lumina.createElement("text", {
                    key = "lbl-" .. fieldId,
                    foreground = t.text or "#E8EDF7",
                    bold = true,
                    style = { height = 1 },
                }, labelText)
            end
            -- Input
            children[#children + 1] = lumina.createElement("input", {
                key = "inp-" .. fieldId,
                id = "form-" .. fieldId,
                value = value or field.defaultValue or "",
                placeholder = field.placeholder or "",
                foreground = err and (t.error or "#F87171") or (t.text or "#E8EDF7"),
                background = t.surface0 or "#141C2C",
                focusable = not disabled,
                style = { height = 1, width = width },
                onChange = onFieldChange and function(text)
                    onFieldChange(fieldId, text)
                end or nil,
            })
            -- Error
            if err then
                children[#children + 1] = lumina.createElement("text", {
                    key = "err-" .. fieldId,
                    foreground = t.error or "#F87171",
                    style = { height = 1 },
                }, err)
            end
            -- Spacer
            children[#children + 1] = lumina.createElement("text", {
                key = "sp-" .. fieldId,
                style = { height = 1 },
            }, "")

        elseif fieldType == "checkbox" then
            local checked = value == true
            local indicator = checked and "[x]" or "[ ]"
            local label = field.label or fieldId
            children[#children + 1] = lumina.createElement("text", {
                key = "cb-" .. fieldId,
                foreground = t.text or "#E8EDF7",
                style = { height = 1 },
                onClick = (not disabled and onFieldChange) and function()
                    onFieldChange(fieldId, not checked)
                end or nil,
            }, indicator .. " " .. label)
            children[#children + 1] = lumina.createElement("text", {
                key = "sp-" .. fieldId,
                style = { height = 1 },
            }, "")
        end
    end

    -- Submit/Reset buttons row
    local buttons = {}
    buttons[#buttons + 1] = lumina.createElement("text", {
        key = "submit-btn",
        foreground = disabled and (t.muted or "#8B9BB4") or (t.primary or "#F5C842"),
        bold = true,
        style = { height = 1 },
        onClick = (not disabled and onSubmit) and function()
            onSubmit(values)
        end or nil,
    }, " [" .. (props.submitLabel or "Submit") .. "] ")

    if onReset then
        buttons[#buttons + 1] = lumina.createElement("text", {
            key = "reset-btn",
            foreground = t.muted or "#8B9BB4",
            style = { height = 1 },
            onClick = (not disabled) and function()
                onReset()
            end or nil,
        }, " [" .. (props.resetLabel or "Reset") .. "] ")
    end

    children[#children + 1] = lumina.createElement("hbox", {
        key = "buttons",
        style = { height = 1 },
    }, table.unpack(buttons))

    return lumina.createElement("vbox", {
        id = props.id,
        key = props.key,
        style = { width = width },
    }, table.unpack(children))
end)

return Form
