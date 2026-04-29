-- lua/lux/tabs.lua — Lux Tabs: tabbed navigation with keyboard support.
-- Usage: local Tabs = require("lux.tabs")

local Tabs = lumina.defineComponent("Tabs", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local tabs = props.tabs or {}
    local activeTab = props.activeTab
    local onTabChange = props.onTabChange
    local renderContent = props.renderContent

    -- Find active index
    local activeIdx = 1
    for i, tab in ipairs(tabs) do
        if tab.id == activeTab then
            activeIdx = i
            break
        end
    end

    -- Find next/prev non-disabled tab
    local function findNextTab(from, direction)
        local n = #tabs
        if n == 0 then return nil end
        local i = from
        for _ = 1, n do
            i = i + direction
            if i < 1 then i = n end
            if i > n then i = 1 end
            if not tabs[i].disabled then
                return tabs[i].id
            end
        end
        return nil
    end

    local function onKeyDown(e)
        if #tabs == 0 then return end
        local newId = nil
        if e.key == "ArrowLeft" or e.key == "Left" or e.key == "h" then
            newId = findNextTab(activeIdx, -1)
        elseif e.key == "ArrowRight" or e.key == "Right" or e.key == "l" then
            newId = findNextTab(activeIdx, 1)
        elseif e.key == "Home" then
            -- First non-disabled tab
            for _, tab in ipairs(tabs) do
                if not tab.disabled then newId = tab.id; break end
            end
        elseif e.key == "End" then
            -- Last non-disabled tab
            for i = #tabs, 1, -1 do
                if not tabs[i].disabled then newId = tabs[i].id; break end
            end
        end
        if newId and newId ~= activeTab and onTabChange then
            onTabChange(newId)
        end
    end

    -- Build tab bar
    local tabBarChildren = {}
    for _, tab in ipairs(tabs) do
        local isActive = (tab.id == activeTab)
        local isDisabled = tab.disabled == true
        local fg, bg
        if isDisabled then
            fg = t.muted or "#6C7086"
            bg = t.base or "#1E1E2E"
        elseif isActive then
            fg = t.primary or "#89B4FA"
            bg = t.surface1 or "#45475A"
        else
            fg = t.text or "#CDD6F4"
            bg = t.surface0 or "#313244"
        end

        tabBarChildren[#tabBarChildren + 1] = lumina.createElement("text", {
            key = "tab-" .. tab.id,
            foreground = fg,
            background = bg,
            bold = isActive,
            dim = isDisabled,
            onClick = (not isDisabled) and function()
                if onTabChange and tab.id ~= activeTab then
                    onTabChange(tab.id)
                end
            end or nil,
            style = { height = 1 },
        }, " " .. (tab.label or tab.id) .. " ")
    end

    local tabBar = lumina.createElement("hbox", {
        style = { height = 1 },
    }, table.unpack(tabBarChildren))

    -- Underline/separator
    local sepW = props.width or 40
    local sep = lumina.createElement("text", {
        foreground = t.surface1 or "#45475A",
        dim = true,
        style = { height = 1 },
    }, string.rep("─", sepW))

    -- Content area
    local content = lumina.createElement("text", {
        style = { height = 1 },
    }, "")
    if type(renderContent) == "function" and activeTab then
        content = renderContent(activeTab)
    end

    local contentBox = lumina.createElement("vbox", {
        style = { flex = 1, minHeight = 1 },
    }, content)

    -- Root
    local rootStyle = { height = props.height or 10 }
    if props.width then rootStyle.width = props.width end

    return lumina.createElement("vbox", {
        id = props.id,
        key = props.key,
        style = rootStyle,
        focusable = true,
        autoFocus = props.autoFocus == true,
        onKeyDown = onKeyDown,
    }, tabBar, sep, contentBox)
end)

return Tabs
