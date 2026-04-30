-- lua/lux/dropdown.lua
-- Lux Dropdown wrapper for the Go Dropdown widget
-- Usage: local Dropdown = require("lux.dropdown")

local Dropdown = lumina.defineComponent("LuxDropdown", function(props)
    return lumina.createElement(lumina.Dropdown, {
        id = props.id,
        key = props.key,
        label = props.label,
        items = props.items,
        selectedIndex = props.selectedIndex,
        onChange = props.onChange,
        disabled = props.disabled,
        width = props.width,
        style = props.style,
    })
end)

return Dropdown
