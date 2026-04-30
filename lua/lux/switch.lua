-- lua/lux/switch.lua
-- Lux Switch wrapper for the Go Switch widget
-- Usage: local Switch = require("lux.switch")

local Switch = lumina.defineComponent("LuxSwitch", function(props)
    return lumina.createElement(lumina.Switch, {
        id = props.id,
        key = props.key,
        label = props.label,
        checked = props.checked,
        disabled = props.disabled,
        onChange = props.onChange,
        style = props.style,
    })
end)

return Switch
