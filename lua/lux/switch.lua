-- lua/lux/switch.lua
-- Pure Lua Switch component for Lux (controlled pattern)
-- Usage: local Switch = require("lux.switch")

local Switch = lumina.defineComponent("LuxSwitch", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local checked = props.checked or false
    local disabled = props.disabled == true
    local label = props.label or ""
    local hovered, setHovered = lumina.useState("hover", false)

    -- Switch visual
    local indicator = checked and "(●)" or "( )"
    local trackFg = disabled and (t.muted or "#8B9BB4")
        or (checked and (t.primary or "#F5C842") or (t.surface2 or "#2A3A56"))
    local fg = disabled and (t.muted or "#8B9BB4") or (t.text or "#E8EDF7")

    local children = {
        lumina.createElement("text", {
            key = "track",
            foreground = trackFg,
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

return Switch
