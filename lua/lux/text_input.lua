-- lua/lux/text_input.lua — Lux TextInput: themed text input field with label/error/helper.
-- Usage: local TextInput = require("lux.text_input")

local TextInput = lumina.defineComponent("TextInput", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}

    local children = {}

    -- Optional label
    if props.label and props.label ~= "" then
        children[#children + 1] = lumina.createElement("text", {
            foreground = t.text or "#CDD6F4",
            style = { height = 1 },
            bold = true,
        }, props.label)
    end

    -- Input element (native)
    local inputFg = t.text or "#CDD6F4"
    local inputBg = t.surface0 or "#313244"
    if props.disabled then
        inputFg = t.muted or "#6C7086"
    end

    local fill = props.fill == true
    local inputStyle = { height = 1 }
    if not fill then
        inputStyle.width = props.width or 30
    end

    children[#children + 1] = lumina.createElement("input", {
        id = props.inputId or (props.id and (props.id .. "-input")),
        value = props.value or "",
        placeholder = props.placeholder or "",
        foreground = inputFg,
        background = inputBg,
        focusable = not props.disabled,
        autoFocus = props.autoFocus,
        style = inputStyle,
        onChange = props.onChange,
        onSubmit = props.onSubmit,
        onFocus = props.onFocus,
        onBlur = props.onBlur,
    })

    -- Helper text or error message
    if props.error and type(props.error) == "string" then
        children[#children + 1] = lumina.createElement("text", {
            foreground = t.error or "#F38BA8",
            style = { height = 1 },
        }, props.error)
    elseif props.helperText and props.helperText ~= "" then
        children[#children + 1] = lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            style = { height = 1 },
        }, props.helperText)
    end

    local rootHeight = 1
    if props.label and props.label ~= "" then rootHeight = rootHeight + 1 end
    if (props.error and type(props.error) == "string") or (props.helperText and props.helperText ~= "") then
        rootHeight = rootHeight + 1
    end

    local rootStyle = { height = rootHeight }
    if not fill then
        rootStyle.width = props.width or 30
    end

    return lumina.createElement("vbox", {
        id = props.id,
        key = props.key,
        style = rootStyle,
    }, table.unpack(children))
end)

return TextInput
