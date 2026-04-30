-- lua/lux/radio.lua
-- Lux Radio wrapper for the Go Radio widget
-- Usage: local Radio = require("lux.radio")

local Radio = lumina.defineComponent("LuxRadio", function(props)
    return lumina.createElement(lumina.Radio, {
        id = props.id,
        key = props.key,
        label = props.label,
        value = props.value,
        checked = props.checked,
        disabled = props.disabled,
        onChange = props.onChange,
        style = props.style,
    })
end)

return Radio
