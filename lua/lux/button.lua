-- lua/lux/button.lua
-- Lux Button wrapper for the Go Button widget
-- Usage: local Button = require("lux.button")

local Button = lumina.defineComponent("LuxButton", function(props)
    return lumina.createElement(lumina.Button, {
        id = props.id,
        key = props.key,
        label = props.label or "Button",
        variant = props.variant or "primary",
        disabled = props.disabled,
        onClick = props.onClick,
        style = props.style,
    })
end)

return Button
