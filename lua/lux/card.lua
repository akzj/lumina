-- lua/lux/card.lua
-- Card component for Lux
-- Usage: local Card = require("lux.card")

local Card = lumina.defineComponent("Card", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local bg = props.bg
    if (bg == nil or bg == "") and props.variant == "elevated" then
        bg = t.surface0 or "#141C2C"
    end
    if bg == nil then
        bg = ""
    end

    local border = props.border
    if border == nil or border == "" then
        border = "rounded"
    end
    local borderColor = props.borderColor
    if (borderColor == nil or borderColor == "") and props.variant == "elevated" then
        borderColor = t.surface2 or "#2A3A56"
    end

    local boxStyle = {
        border = border,
        padding = props.padding or 1,
        background = bg,
    }
    if borderColor and borderColor ~= "" then
        boxStyle.borderColor = borderColor
    end
    if type(props.style) == "table" then
        for k, v in pairs(props.style) do
            boxStyle[k] = v
        end
    end

    return lumina.createElement("box", {
        style = boxStyle,
    },
        props.title and lumina.createElement("text", {
            bold = true,
        }, props.title) or nil,
        table.unpack(props.children or {})
    )
end)

return Card
