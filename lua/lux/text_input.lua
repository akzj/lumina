-- lua/lux/text_input.lua — Lux TextInput: themed single-line input with label/error/helper.
-- Usage: local TextInput = require("lux.text_input")
--
-- Props:
--   value: string (controlled mode)
--   initialValue: string (uncontrolled initial)
--   placeholder: string
--   onChange: function(text)
--   onSubmit: function(text)
--   label: string
--   error: string
--   helperText: string
--   width: number (default 30)
--   fill: boolean — if true, flex instead of fixed width
--   disabled: boolean
--   autoFocus: boolean
--   id: string
--   inputId: string
--   rootStyle: table

local Textarea = require("lux.textarea")

local TextInput = lumina.defineComponent("LuxTextInput", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local children = {}

    -- Optional label
    if props.label and props.label ~= "" then
        children[#children + 1] = lumina.createElement("text", {
            key = "label",
            foreground = t.text or "#E8EDF7",
            style = { height = 1 },
            bold = true,
        }, props.label)
    end

    -- Input (single-line textarea)
    local inputFg = t.text or "#E8EDF7"
    local inputBg = t.surface0 or "#141C2C"
    if props.disabled then
        inputFg = t.muted or "#8B9BB4"
    end

    local inputStyle = { height = 1 }
    if props.fill then
        inputStyle.flex = 1
    else
        inputStyle.width = props.width or 30
    end

    children[#children + 1] = lumina.createElement(Textarea, {
        key = "input",
        id = props.inputId or (props.id and (props.id .. "-input")),
        value = props.value,
        initialValue = props.initialValue or "",
        placeholder = props.placeholder or "",
        foreground = inputFg,
        background = inputBg,
        focusable = not props.disabled,
        disabled = props.disabled,
        autoFocus = props.autoFocus,
        maxHeight = 1,  -- single line
        style = inputStyle,
        onChange = props.onChange,
        onSubmit = props.onSubmit,
    })

    -- Helper text or error message
    if props.error and type(props.error) == "string" then
        children[#children + 1] = lumina.createElement("text", {
            key = "error",
            foreground = t.error or "#F87171",
            style = { height = 1 },
        }, props.error)
    elseif props.helperText and props.helperText ~= "" then
        children[#children + 1] = lumina.createElement("text", {
            key = "helper",
            foreground = t.muted or "#8B9BB4",
            style = { height = 1 },
        }, props.helperText)
    end

    local rootHeight = 1
    if props.label and props.label ~= "" then rootHeight = rootHeight + 1 end
    if (props.error and type(props.error) == "string") or (props.helperText and props.helperText ~= "") then
        rootHeight = rootHeight + 1
    end

    local rootStyle = props.rootStyle or { height = rootHeight }
    if not props.fill then
        rootStyle.width = rootStyle.width or props.width or 30
    end

    return lumina.createElement("vbox", {
        id = props.id,
        key = props.key,
        style = rootStyle,
    }, table.unpack(children))
end)

return TextInput
