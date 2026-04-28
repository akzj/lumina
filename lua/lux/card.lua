local lumina = require("lumina")

local Card = lumina.defineComponent("Card", function(props)
    return lumina.createElement("box", {
        style = {
            border = props.border or "rounded",
            padding = props.padding or 1,
            background = props.bg or "",
        },
    },
        props.title and lumina.createElement("text", {
            bold = true,
        }, props.title) or nil,
        table.unpack(props.children or {})
    )
end)

return Card
