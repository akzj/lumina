-- lua/lux/radio.lua
-- Pure Lua Radio button component for Lux (controlled pattern)
-- Usage: local Radio = require("lux.radio")
-- Props: label, value, checked, disabled, onChange, id, key, style
-- onChange fires with the "value" prop (string) when clicked or Space/Enter pressed.

local Radio = lumina.defineComponent("LuxRadio", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local checked = props.checked or false
    local disabled = props.disabled == true
    local label = props.label or ""
    local value = props.value or ""
    local hovered, setHovered = lumina.useState("hover", false)

    -- Visual
    local indicator = checked and "(●)" or "( )"
    local fg = disabled and (t.muted or "#8B9BB4") or (t.text or "#E8EDF7")
    local indicatorFg = disabled and (t.muted or "#8B9BB4")
        or (hovered and (t.hover or "#FFD35A") or (t.primary or "#F5C842"))

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
                props.onChange(value)
            end
        end or nil,
        onKeyDown = not disabled and function(e)
            local k = type(e) == "table" and (e.key or e.Key) or tostring(e or "")
            if k == " " or k == "Enter" then
                if props.onChange then
                    props.onChange(value)
                end
            end
        end or nil,
        onMouseEnter = not disabled and function() setHovered(true) end or nil,
        onMouseLeave = not disabled and function() setHovered(false) end or nil,
    }, table.unpack(children))
end)

return Radio
