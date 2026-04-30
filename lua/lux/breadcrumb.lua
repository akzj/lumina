-- lua/lux/breadcrumb.lua — Lux Breadcrumb: navigation trail.
-- Usage: local Breadcrumb = require("lux.breadcrumb")

local Breadcrumb = lumina.defineComponent("Breadcrumb", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local items = props.items or {}
    local separator = props.separator or " › "
    local onNavigate = props.onNavigate
    local barBg = props.barBackground or props.background or t.base or "#1E1E2E"

    local children = {}
    for i, item in ipairs(items) do
        local isLast = (i == #items)
        local fg
        if isLast then
            fg = t.text or "#CDD6F4"
        else
            fg = t.primary or "#89B4FA"
        end

        -- Include item.id in the key so changing the active page (same index, new
        -- label) does not reuse a stale text node when the reconciler keys match.
        children[#children + 1] = lumina.createElement("text", {
            key = "bc-" .. tostring(item.id or i),
            foreground = fg,
            background = barBg,
            bold = isLast,
            underline = not isLast,
            style = { height = 1 },
            onClick = (not isLast and onNavigate) and function()
                onNavigate(item.id, i)
            end or nil,
        }, item.label or item.id)

        -- Separator (not after last)
        if not isLast then
            children[#children + 1] = lumina.createElement("text", {
                key = "sep-" .. tostring(i),
                foreground = t.muted or "#6C7086",
                background = barBg,
                style = { height = 1 },
            }, separator)
        end
    end

    local rootStyle = { height = 1 }
    if props.width then rootStyle.width = props.width end
    local extra = props.style
    if type(extra) == "table" then
        for k, v in pairs(extra) do
            rootStyle[k] = v
        end
    end

    return lumina.createElement("hbox", {
        id = props.id,
        key = props.key,
        style = rootStyle,
    }, table.unpack(children))
end)

return Breadcrumb
