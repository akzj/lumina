-- lua/lux/checkbox.lua
-- Pure Lua Checkbox component for Lux (controlled pattern)
-- Usage: local Checkbox = require("lux.checkbox")

local Checkbox = lumina.defineComponent("LuxCheckbox", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local checked = props.checked or false
    local disabled = props.disabled == true
    local label = props.label or ""
    local hovered, setHovered = lumina.useState("hover", false)

    -- Visual
    local indicator = checked and "[x]" or "[ ]"
    local fg = disabled and (t.muted or "#6C7086") or (t.text or "#CDD6F4")
    local indicatorFg = disabled and (t.muted or "#6C7086")
        or (hovered and (t.hover or "#B4BEFE") or (t.primary or "#89B4FA"))

    local children = {
        lumina.createElement("text", {
            key = "indicator",
            foreground = indicatorFg,
        }, indicator),
    }
    if label ~= "" then
        children[#children + 1] = lumina.createElement("text", {
            key = "label",
            foreground = fg,
            dim = disabled,
        }, " " .. label)
    end

    return lumina.createElement("hbox", {
        id = props.id,
        key = props.key,
        focusable = not disabled,
        disabled = disabled,
        style = props.style,
        onClick = not disabled and function()
            if props.onChange then
                props.onChange(not checked)
            end
        end or nil,
        onKeyDown = not disabled and function(e)
            local k = type(e) == "table" and (e.key or e.Key) or tostring(e or "")
            if k == " " or k == "Enter" then
                if props.onChange then
                    props.onChange(not checked)
                end
            end
        end or nil,
        onMouseEnter = not disabled and function() setHovered(true) end or nil,
        onMouseLeave = not disabled and function() setHovered(false) end or nil,
    }, table.unpack(children))
end)

return Checkbox
