-- lua/lux/pagination.lua — Lux Pagination: react-paginate style for TUI.
-- Usage: local Pagination = require("lux.pagination")

-- Returns array of items: each is a number (page) or "break"
local function buildPageItems(currentPage, pageCount, pageRange, marginPages)
    local items = {}
    local set = {}

    -- Left margin
    for i = 1, math.min(marginPages, pageCount) do
        if not set[i] then set[i] = true; items[#items + 1] = i end
    end

    -- Center range around current
    local half = math.floor(pageRange / 2)
    local rangeStart = math.max(1, currentPage - half)
    local rangeEnd = math.min(pageCount, rangeStart + pageRange - 1)
    -- Adjust if clamped at end
    rangeStart = math.max(1, rangeEnd - pageRange + 1)

    for i = rangeStart, rangeEnd do
        if not set[i] then set[i] = true; items[#items + 1] = i end
    end

    -- Right margin
    for i = math.max(1, pageCount - marginPages + 1), pageCount do
        if not set[i] then set[i] = true; items[#items + 1] = i end
    end

    -- Sort
    table.sort(items)

    -- Insert breaks where gaps > 1
    local result = {}
    for idx, page in ipairs(items) do
        if idx > 1 and page - items[idx - 1] > 1 then
            result[#result + 1] = "break"
        end
        result[#result + 1] = page
    end

    return result
end

local Pagination = lumina.defineComponent("LuxPagination", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}

    local pageCount = props.pageCount or 1
    local currentPage = props.currentPage or 1
    local onPageChange = props.onPageChange
    local pageRange = props.pageRangeDisplayed or 3
    local marginPages = props.marginPagesDisplayed or 1
    local prevLabel = props.previousLabel or "\226\128\185"
    local nextLabel = props.nextLabel or "\226\128\186"
    local breakLabel = props.breakLabel or "\226\128\166"

    -- Clamp
    if currentPage < 1 then currentPage = 1 end
    if currentPage > pageCount then currentPage = pageCount end

    local pageItems = buildPageItems(currentPage, pageCount, pageRange, marginPages)

    local function goTo(page)
        if page < 1 or page > pageCount or page == currentPage then return end
        if onPageChange then onPageChange(page) end
    end

    -- Keyboard handler
    local function onKeyDown(e)
        if e.key == "ArrowLeft" or e.key == "Left" or e.key == "h" then
            goTo(currentPage - 1)
        elseif e.key == "ArrowRight" or e.key == "Right" or e.key == "l" then
            goTo(currentPage + 1)
        elseif e.key == "Home" then
            goTo(1)
        elseif e.key == "End" then
            goTo(pageCount)
        end
    end

    -- Build children
    local children = {}

    -- Previous button
    local prevDisabled = (currentPage <= 1)
    children[#children + 1] = lumina.createElement("text", {
        key = "prev",
        foreground = prevDisabled and (t.muted or "#8B9BB4") or (t.text or "#E8EDF7"),
        background = t.base or "#0B1220",
        dim = prevDisabled,
        onClick = not prevDisabled and function() goTo(currentPage - 1) end or nil,
        style = { height = 1 },
    }, " " .. prevLabel .. " ")

    -- Page items
    for idx, item in ipairs(pageItems) do
        if item == "break" then
            children[#children + 1] = lumina.createElement("text", {
                key = "brk-" .. tostring(idx),
                foreground = t.muted or "#8B9BB4",
                background = t.base or "#0B1220",
                dim = true,
                style = { height = 1 },
            }, " " .. breakLabel .. " ")
        else
            local isCurrent = (item == currentPage)
            children[#children + 1] = lumina.createElement("text", {
                key = "p-" .. tostring(item),
                foreground = isCurrent and (t.primary or "#F5C842") or (t.text or "#E8EDF7"),
                background = isCurrent and (t.surface1 or "#1B2639") or (t.base or "#0B1220"),
                bold = isCurrent,
                onClick = not isCurrent and function() goTo(item) end or nil,
                style = { height = 1 },
            }, " " .. tostring(item) .. " ")
        end
    end

    -- Next button
    local nextDisabled = (currentPage >= pageCount)
    children[#children + 1] = lumina.createElement("text", {
        key = "next",
        foreground = nextDisabled and (t.muted or "#8B9BB4") or (t.text or "#E8EDF7"),
        background = t.base or "#0B1220",
        dim = nextDisabled,
        onClick = not nextDisabled and function() goTo(currentPage + 1) end or nil,
        style = { height = 1 },
    }, " " .. nextLabel .. " ")

    local rootStyle = { height = 1 }
    if props.width then rootStyle.width = props.width end

    return lumina.createElement("hbox", {
        id = props.id,
        key = props.key,
        role = "navigation",
        style = rootStyle,
        focusable = true,
        autoFocus = props.autoFocus == true,
        onKeyDown = onKeyDown,
    }, table.unpack(children))
end)

return Pagination
