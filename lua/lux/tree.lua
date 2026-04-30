-- lua/lux/tree.lua — Lux Tree: hierarchical tree view.
-- Usage: local Tree = require("lux.tree")
--
-- Props:
--   items: array of { id, label, children?, icon?, disabled? }
--   expandedIds: table (array of expanded node IDs)
--   onToggle: function(id, expanded, newExpandedIds)
--   selectedId: string (currently selected node)
--   onSelect: function(id)
--   onActivate: function(id) — Enter key on leaf
--   indent: number (default 2)
--   autoFocus: boolean

local Tree = lumina.defineComponent("Tree", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local items = props.items or {}
    local expandedIds = props.expandedIds or {}
    local onToggle = props.onToggle
    local selectedId = props.selectedId
    local onSelect = props.onSelect
    local onActivate = props.onActivate
    local indent = props.indent or 2

    -- Build expanded set
    local expandedSet = {}
    for _, id in ipairs(expandedIds) do
        expandedSet[id] = true
    end

    -- Flatten tree into visible list
    local flatList = {}
    local function flatten(nodeList, depth)
        for _, node in ipairs(nodeList) do
            local hasChildren = node.children and #node.children > 0
            local isExpanded = expandedSet[node.id] == true
            flatList[#flatList + 1] = {
                id = node.id,
                label = node.label or node.id,
                icon = node.icon,
                depth = depth,
                hasChildren = hasChildren,
                isExpanded = isExpanded,
                disabled = node.disabled,
            }
            if hasChildren and isExpanded then
                flatten(node.children, depth + 1)
            end
        end
    end
    flatten(items, 0)

    -- Find selected index in flat list
    local selectedIdx = 0
    for i, item in ipairs(flatList) do
        if item.id == selectedId then
            selectedIdx = i
            break
        end
    end

    local function onKeyDown(e)
        if #flatList == 0 then return end
        if e.key == "ArrowUp" or e.key == "Up" or e.key == "k" then
            local n = selectedIdx - 1
            if n < 1 then n = 1 end
            -- Skip disabled
            while n >= 1 and flatList[n] and flatList[n].disabled do n = n - 1 end
            if n >= 1 and flatList[n] and onSelect then
                onSelect(flatList[n].id)
            end
        elseif e.key == "ArrowDown" or e.key == "Down" or e.key == "j" then
            local n = selectedIdx + 1
            if n > #flatList then n = #flatList end
            while n <= #flatList and flatList[n] and flatList[n].disabled do n = n + 1 end
            if n <= #flatList and flatList[n] and onSelect then
                onSelect(flatList[n].id)
            end
        elseif e.key == "ArrowRight" or e.key == "Right" or e.key == "l" then
            -- Expand current node (if it has children and is collapsed)
            local item = flatList[selectedIdx]
            if item and item.hasChildren and not item.isExpanded and onToggle then
                local newExpanded = {}
                for _, id in ipairs(expandedIds) do newExpanded[#newExpanded + 1] = id end
                newExpanded[#newExpanded + 1] = item.id
                onToggle(item.id, true, newExpanded)
            end
        elseif e.key == "ArrowLeft" or e.key == "Left" or e.key == "h" then
            -- Collapse current node
            local item = flatList[selectedIdx]
            if item and item.hasChildren and item.isExpanded and onToggle then
                local newExpanded = {}
                for _, id in ipairs(expandedIds) do
                    if id ~= item.id then newExpanded[#newExpanded + 1] = id end
                end
                onToggle(item.id, false, newExpanded)
            end
        elseif e.key == "Enter" then
            local item = flatList[selectedIdx]
            if item then
                if item.hasChildren and onToggle then
                    -- Toggle expand
                    local wasExpanded = item.isExpanded
                    local newExpanded = {}
                    for _, id in ipairs(expandedIds) do
                        if id ~= item.id then newExpanded[#newExpanded + 1] = id end
                    end
                    if not wasExpanded then
                        newExpanded[#newExpanded + 1] = item.id
                    end
                    onToggle(item.id, not wasExpanded, newExpanded)
                elseif onActivate then
                    onActivate(item.id)
                end
            end
        end
    end

    -- Render flat list
    local children = {}
    for _, item in ipairs(flatList) do
        local isSelected = (item.id == selectedId)
        local prefix = string.rep(" ", item.depth * indent)

        -- Arrow/icon
        local arrow = "  "
        if item.hasChildren then
            arrow = item.isExpanded and "v " or "> "
        end

        local icon = ""
        if item.icon then
            icon = item.icon .. " "
        elseif item.hasChildren then
            icon = item.isExpanded and "[-] " or "[+] "
        end

        local fg
        if item.disabled then
            fg = t.muted or "#6C7086"
        elseif isSelected then
            fg = t.primary or "#89B4FA"
        else
            fg = t.text or "#CDD6F4"
        end
        local bg = isSelected and (t.surface0 or "#313244") or (t.base or "#1E1E2E")

        children[#children + 1] = lumina.createElement("text", {
            key = "tn-" .. item.id,
            foreground = fg,
            background = bg,
            bold = isSelected,
            dim = item.disabled,
            style = { height = 1 },
            onClick = (not item.disabled and onSelect) and function()
                onSelect(item.id)
            end or nil,
        }, prefix .. arrow .. icon .. item.label)
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

return Tree
