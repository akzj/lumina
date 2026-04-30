-- lua/lux/checkbox.lua
-- Lux Checkbox wrapper for the Go Checkbox widget
-- Usage: local Checkbox = require("lux.checkbox")

local Checkbox = lumina.defineComponent("LuxCheckbox", function(props)
    return lumina.createElement(lumina.Checkbox, {
        id = props.id,
        key = props.key,
        label = props.label,
        checked = props.checked,
        disabled = props.disabled,
        onChange = props.onChange,
        style = props.style,
    })
end)

return Checkbox
