-- lua/lux/alert.lua — Lux Alert: themed notification banner.
-- Usage: local Alert = require("lux.alert")

local Alert = lumina.defineComponent("Alert", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local variant = props.variant or "info"

    -- Variant colors and icons
    local icon, fg, bg
    if variant == "error" then
        icon = "✗ "
        fg = t.error or "#F38BA8"
        bg = t.surface0 or "#313244"
    elseif variant == "warning" then
        icon = "⚠ "
        fg = t.warning or "#F9E2AF"
        bg = t.surface0 or "#313244"
    elseif variant == "success" then
        icon = "✓ "
        fg = t.success or "#A6E3A1"
        bg = t.surface0 or "#313244"
    else -- info
        icon = "ℹ "
        fg = t.primary or "#89B4FA"
        bg = t.surface0 or "#313244"
    end

    local children = {}

    -- Title + icon row
    local titleRow = {}
    local titleText = icon .. (props.title or variant:sub(1,1):upper() .. variant:sub(2))
    titleRow[#titleRow + 1] = lumina.createElement("text", {
        foreground = fg,
        background = bg,
        bold = true,
        style = { flex = 1, height = 1 },
    }, " " .. titleText)

    -- Dismiss button
    if props.dismissible then
        titleRow[#titleRow + 1] = lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            background = bg,
            style = { width = 4, height = 1 },
            onClick = function()
                if props.onDismiss then props.onDismiss() end
            end,
        }, " ✕ ")
    end

    children[#children + 1] = lumina.createElement("hbox", {
        style = { height = 1 },
    }, table.unpack(titleRow))

    -- Message
    if props.message and props.message ~= "" then
        children[#children + 1] = lumina.createElement("text", {
            foreground = t.text or "#CDD6F4",
            background = bg,
            style = { height = 1 },
        }, " " .. props.message)
    end

    local rootStyle = { background = bg }
    if props.width then rootStyle.width = props.width end
    -- Height: title + message (if any)
    local h = 1
    if props.message and props.message ~= "" then h = 2 end
    rootStyle.height = props.height or h

    return lumina.createElement("vbox", {
        id = props.id,
        key = props.key,
        style = rootStyle,
    }, table.unpack(children))
end)

return Alert
