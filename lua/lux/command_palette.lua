-- lua/lux/command_palette.lua
-- CommandPalette: searchable command list overlay.
-- Usage:
--   local CommandPalette = require("lux.command_palette")
--   CommandPalette {
--       commands = {
--           { title = "Save File", action = function() ... end },
--           { title = "Open File", action = function() ... end },
--           { title = "Quit",      action = function() lumina.quit() end },
--       },
--       onClose = function() ... end,
--   }

-- Returns true if s is a single UTF-8 character (1-4 bytes).
local function isSingleChar(s)
    local n = #s
    if n == 0 or n > 4 then return false end
    local b = s:byte(1)
    if b < 0x80 then return n == 1
    elseif b < 0xC0 then return false -- continuation byte
    elseif b < 0xE0 then return n == 2
    elseif b < 0xF0 then return n == 3
    else return n == 4
    end
end

-- Remove the last UTF-8 character from a string.
local function utf8RemoveLast(s)
    local n = #s
    if n == 0 then return s end
    -- Walk backwards past continuation bytes (10xxxxxx = 0x80..0xBF)
    local i = n
    while i > 1 and s:byte(i) >= 0x80 and s:byte(i) < 0xC0 do
        i = i - 1
    end
    return s:sub(1, i - 1)
end

local CommandPalette = lumina.defineComponent("CommandPalette", function(props)
    local query, setQuery = lumina.useState("query", "")
    local selectedIdx, setSelectedIdx = lumina.useState("selectedIdx", 1)

    local commands = props.commands or {}
    local t = lumina.getTheme and lumina.getTheme() or {}

    -- Filter commands by query (substring match, case-insensitive)
    local filtered = {}
    for _, cmd in ipairs(commands) do
        if query == "" or cmd.title:lower():find(query:lower(), 1, true) then
            filtered[#filtered + 1] = cmd
        end
    end

    -- Clamp selected index
    if selectedIdx > #filtered then
        selectedIdx = #filtered
    end
    if selectedIdx < 1 then
        selectedIdx = 1
    end

    -- Build list items
    local items = {}
    for i, cmd in ipairs(filtered) do
        local isSelected = (i == selectedIdx)
        local prefix = isSelected and "> " or "  "
        items[#items + 1] = lumina.createElement("text", {
            id = "cmd-" .. i,
            foreground = isSelected and (t.primary or "#89B4FA") or (t.text or "#CDD6F4"),
            bold = isSelected,
            style = {
                height = 1,
                background = isSelected and (t.surface1 or "#45475A") or "",
            },
        }, prefix .. cmd.title)
    end

    -- Keyboard handler
    local function handleKey(e)
        if e.key == "Escape" then
            if props.onClose then props.onClose() end
        elseif e.key == "Enter" then
            if filtered[selectedIdx] and filtered[selectedIdx].action then
                filtered[selectedIdx].action()
            end
            if props.onClose then props.onClose() end
        elseif e.key == "ArrowUp" or e.key == "k" then
            setSelectedIdx(math.max(1, selectedIdx - 1))
        elseif e.key == "ArrowDown" or e.key == "j" then
            setSelectedIdx(math.min(#filtered, selectedIdx + 1))
        elseif e.key == "Backspace" then
            if #query > 0 then
                setQuery(utf8RemoveLast(query))
                setSelectedIdx(1)
            end
        elseif isSingleChar(e.key) then
            -- Single character input (ASCII or multi-byte UTF-8)
            setQuery(query .. e.key)
            setSelectedIdx(1)
        end
    end

    local paletteWidth = props.width or 50
    local maxHeight = props.maxHeight or 15
    local visibleItems = math.min(#filtered, maxHeight - 3)
    local paletteHeight = visibleItems + 3

    -- Divider width
    local divWidth = paletteWidth - 4
    if divWidth < 1 then divWidth = 1 end

    return lumina.createElement("vbox", {
        id = "command-palette",
        style = {
            border = "rounded",
            width = paletteWidth,
            height = paletteHeight,
            background = t.surface0 or "#313244",
        },
        onKeyDown = handleKey,
        focusable = true,
    },
        -- Search input display
        lumina.createElement("text", {
            id = "cp-input",
            foreground = t.text or "#CDD6F4",
            style = { height = 1 },
        }, "> " .. query .. "▏"),

        -- Divider
        lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            dim = true,
            style = { height = 1 },
        }, string.rep("─", divWidth)),

        -- Command list
        table.unpack(items)
    )
end)

return CommandPalette
