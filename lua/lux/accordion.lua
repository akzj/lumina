-- lua/lux/accordion.lua — Lux Accordion: collapsible panels.
-- Usage: local Accordion = require("lux.accordion")

local Accordion = lumina.defineComponent("Accordion", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local items = props.items or {}
    local mode = props.mode or "single"
    local openItems = props.openItems or {}
    local onToggle = props.onToggle

    -- Build set of open item ids
    local openSet = {}
    for _, id in ipairs(openItems) do
        openSet[id] = true
    end

    -- Selected (focused) item index for keyboard nav
    local selectedIdx = props.selectedIndex or 1
    if selectedIdx < 1 then selectedIdx = 1 end
    if selectedIdx > #items then selectedIdx = #items end
    if #items == 0 then selectedIdx = 0 end

    local function isOpen(id)
        return openSet[id] == true
    end

    local function toggle(id)
        if not onToggle then return end
        local wasOpen = isOpen(id)
        local newOpen
        if mode == "multi" then
            newOpen = {}
            for _, oid in ipairs(openItems) do
                if oid ~= id then
                    newOpen[#newOpen + 1] = oid
                end
            end
            if not wasOpen then
                newOpen[#newOpen + 1] = id
            end
        else -- single
            if wasOpen then
                newOpen = {}
            else
                newOpen = { id }
            end
        end
        onToggle(id, not wasOpen, newOpen)
    end

    local function onKeyDown(e)
        if #items == 0 then return end
        if e.key == "ArrowUp" or e.key == "Up" or e.key == "k" then
            -- Move to previous non-disabled item
            local n = selectedIdx
            for i = 1, #items do
                n = n - 1
                if n < 1 then n = #items end
                if not items[n].disabled then break end
            end
            if props.onSelectedChange then props.onSelectedChange(n) end
        elseif e.key == "ArrowDown" or e.key == "Down" or e.key == "j" then
            local n = selectedIdx
            for i = 1, #items do
                n = n + 1
                if n > #items then n = 1 end
                if not items[n].disabled then break end
            end
            if props.onSelectedChange then props.onSelectedChange(n) end
        elseif e.key == "Enter" or e.key == " " then
            local item = items[selectedIdx]
            if item and not item.disabled then
                toggle(item.id)
            end
        end
    end

    -- Build children
    local children = {}
    for i, item in ipairs(items) do
        local opened = isOpen(item.id)
        local isSelected = (i == selectedIdx)
        local isDisabled = item.disabled == true

        -- Header
        local arrow = opened and "▾ " or "▸ "
        local headerFg
        if isDisabled then
            headerFg = t.muted or "#6C7086"
        elseif isSelected then
            headerFg = t.primary or "#89B4FA"
        else
            headerFg = t.text or "#CDD6F4"
        end
        local headerBg = isSelected and (t.surface0 or "#313244") or (t.base or "#1E1E2E")

        children[#children + 1] = lumina.createElement("text", {
            key = "hdr-" .. item.id,
            foreground = headerFg,
            background = headerBg,
            bold = isSelected,
            dim = isDisabled,
            style = { height = 1 },
            onClick = (not isDisabled) and function()
                toggle(item.id)
            end or nil,
        }, " " .. arrow .. (item.title or item.id))

        -- Content (if open)
        if opened and not isDisabled then
            local content
            if type(item.render) == "function" then
                content = item.render()
            elseif type(item.content) == "string" then
                content = lumina.createElement("text", {
                    foreground = t.text or "#CDD6F4",
                    background = t.base or "#1E1E2E",
                    style = { height = 1 },
                }, "   " .. item.content)
            else
                content = lumina.createElement("text", { style = { height = 1 } }, "")
            end
            children[#children + 1] = content
        end
    end

    local rootStyle = {}
    if props.width then rootStyle.width = props.width end
    if props.height then rootStyle.height = props.height end

    return lumina.createElement("vbox", {
        id = props.id,
        key = props.key,
        style = rootStyle,
        focusable = true,
        autoFocus = props.autoFocus == true,
        onKeyDown = onKeyDown,
    }, table.unpack(children))
end)

return Accordion
