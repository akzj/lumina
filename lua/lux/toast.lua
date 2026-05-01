-- lua/lux/toast.lua — Lux Toast: notification stack.
-- Usage: local Toast = require("lux.toast")
--
-- Props:
--   items: array of { id, message, variant?, duration? }
--   onDismiss: function(id) — called when a toast should be removed
--   maxVisible: number (default 5)
--   width: number (default 40)

local Toast = lumina.defineComponent("LuxToast", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local items = props.items or {}
    local onDismiss = props.onDismiss
    local maxVisible = props.maxVisible or 5

    -- Only show last N toasts
    local visible = {}
    local start = math.max(1, #items - maxVisible + 1)
    for i = start, #items do
        visible[#visible + 1] = items[i]
    end

    local children = {}
    for i, item in ipairs(visible) do
        local variant = item.variant or "info"
        local icon, fg
        if variant == "error" then
            icon = "x "; fg = t.error or "#F87171"
        elseif variant == "warning" then
            icon = "! "; fg = t.warning or "#F5C842"
        elseif variant == "success" then
            icon = "v "; fg = t.success or "#4ADE80"
        else
            icon = "i "; fg = t.primary or "#F5C842"
        end

        local bg = t.surface0 or "#141C2C"
        local width = props.width or 40

        children[#children + 1] = lumina.createElement("hbox", {
            key = "toast-" .. tostring(item.id),
            style = { height = 1, width = width, background = bg },
        },
            lumina.createElement("text", {
                foreground = fg,
                background = bg,
                bold = true,
                style = { width = 3, height = 1 },
            }, " " .. icon),
            lumina.createElement("text", {
                foreground = t.text or "#E8EDF7",
                background = bg,
                style = { flex = 1, height = 1 },
            }, item.message or ""),
            lumina.createElement("text", {
                foreground = t.muted or "#8B9BB4",
                background = bg,
                style = { width = 3, height = 1 },
                onClick = onDismiss and function()
                    onDismiss(item.id)
                end or nil,
            }, " x")
        )
        -- Spacer between toasts
        if i < #visible then
            children[#children + 1] = lumina.createElement("text", {
                key = "gap-" .. tostring(item.id),
                style = { height = 1 },
            }, "")
        end
    end

    if #children == 0 then
        return lumina.createElement("vbox", {
            id = props.id,
            key = props.key,
            style = { height = 0, width = 0 },
        })
    end

    local totalHeight = #visible + (#visible - 1) -- toasts + gaps
    return lumina.createElement("vbox", {
        id = props.id,
        key = props.key,
        style = { height = totalHeight, width = props.width or 40 },
    }, table.unpack(children))
end)

return Toast
