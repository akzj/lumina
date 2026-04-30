-- lua/lux/card.lua
-- Card component for Lux
-- Usage: local Card = require("lux.card")

local Card = lumina.defineComponent("Card", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local bg = props.bg
    if (bg == nil or bg == "") and props.variant == "elevated" then
        bg = t.surface0 or "#313244"
    end
    if bg == nil then
        bg = ""
    end

    return lumina.createElement("box", {
        style = {
            border = props.border or "rounded",
            padding = props.padding or 1,
            background = bg,
        },
    },
        props.title and lumina.createElement("text", {
            bold = true,
        }, props.title) or nil,
        table.unpack(props.children or {})
    )
end)

return Card
