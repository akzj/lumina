-- lux/split_pane.lua — Resizable split pane layout
-- Renders children in a horizontal or vertical split with draggable borders.
-- No visible divider elements — pane borders touch directly (┐┌).
-- Drag-to-resize is triggered by clicking on the border between panes.
--
-- Props:
--   direction: "horizontal" (default) or "vertical"
--   sizes: array of sizes (number > 0 = fixed width/height, 0 = flex)
--   minSizes: array of minimum sizes (optional, default 10)
--   maxSizes: array of maximum sizes (optional, default 200)
--   borderStyle: "single" (default), "double", "rounded", or nil (no border)
--   borderColor: string (border color for panes)
--   background: string or array of background colors per pane
--   onResize: function(newSizes) — called when user drags a border
--   children: pane contents (passed as createElement children)
--
-- Note: Use 0 (not nil) to indicate a flex pane. Lua tables cannot reliably
-- store nil at interior positions.
--
-- Usage:
--   local SplitPane = require("lux.split_pane")
--   lumina.createElement(SplitPane, {
--       direction = "horizontal",
--       sizes = { 34, 0, 36 },   -- 0 means flex
--       minSizes = { 20, 0, 20 },
--       maxSizes = { 60, 0, 60 },
--       borderStyle = "single",
--       borderColor = "#38605c",
--       onResize = function(newSizes) end,
--   }, pane1, pane2, pane3)

local SplitPane = lumina.defineComponent("SplitPane", function(props)
    local direction = props.direction or "horizontal"
    local initialSizes = props.sizes or {}
    local minSizes = props.minSizes or {}
    local maxSizes = props.maxSizes or {}
    local borderStyle = props.borderStyle or "single"
    local borderColor = props.borderColor
    local backgrounds = props.background
    local onResize = props.onResize
    local children = props.children or {}

    -- Sizes: controlled vs uncontrolled mode
    local internalSizes, setInternalSizes = lumina.useState("splitSizes", initialSizes)
    local currentSizes, setCurrentSizes
    if onResize then
        currentSizes = initialSizes
        setCurrentSizes = function() end
    else
        currentSizes = internalSizes
        setCurrentSizes = setInternalSizes
    end

    -- Drag state (mutable ref)
    local dragRef = lumina.useRef({ active = false, paneIndex = 0, startPos = 0, origSize = 0 })

    -- Build pane elements (NO dividers — borders touch directly)
    local elements = {}
    for i, child in ipairs(children) do
        local size = currentSizes[i]
        local isFixed = (size and size > 0)
        local paneStyle = {}
        if direction == "horizontal" then
            if isFixed then
                paneStyle.width = size
            else
                paneStyle.flex = 1
            end
        else
            if isFixed then
                paneStyle.height = size
            else
                paneStyle.flex = 1
            end
        end

        if borderStyle then
            paneStyle.border = borderStyle
            paneStyle.borderColor = borderColor
        end
        paneStyle.overflow = "hidden"

        -- Per-pane background
        if type(backgrounds) == "table" then
            paneStyle.background = backgrounds[i]
        elseif type(backgrounds) == "string" then
            paneStyle.background = backgrounds
        end

        elements[#elements + 1] = lumina.createElement("vbox", {
            key = "split-pane-" .. i,
            style = paneStyle,
        }, child)
    end

    -- Mouse handlers on the outer container for border-drag detection
    local function handleMouseDown(event)
        local pos = (direction == "horizontal") and event.x or event.y

        -- Compute cumulative positions of pane borders.
        -- We need the container's start position. We infer it:
        -- The first pane starts at container.X. Each pane occupies its width.
        -- Border between pane i and pane i+1 is at cumX after pane i.
        -- We don't know container.X, but we can store it from the click position
        -- by checking which border is closest.

        -- Compute total known width to find container start
        -- Strategy: try each border position. If click is within ±1 of a border,
        -- that's a resize drag. We compute borders as cumulative sums from pos 0,
        -- then offset by (event.pos - expected_border_pos) to find container start.

        -- Simple approach: iterate panes, track cumulative width.
        -- The border between pane i and pane i+1 is at cumulative position after pane i.
        -- We don't know absolute start, so we try all borders and pick the closest.

        -- Actually, we CAN compute it: the outer hbox starts at some X.
        -- Total width of all panes = sum of fixed sizes + flex size.
        -- But we don't know flex size at this point...

        -- Pragmatic approach: store startPos and figure out which pane to resize
        -- based on relative position within the container.
        -- Since panes have known sizes, compute cumulative offsets.
        -- The container's X is unknown, but we can detect borders by checking
        -- if the click is on the last column of a fixed-width pane.

        -- Simplest: just store the click position. On mouseMove, compute delta.
        -- Determine which pane border was clicked by cumulative sum.
        -- We approximate containerX = pos - cumulative_to_click_point.

        local cumX = 0
        for i = 1, #currentSizes do
            local s = currentSizes[i]
            if s and s > 0 then
                cumX = cumX + s
            else
                -- Flex pane: we don't know its actual rendered width.
                -- Skip — we can only resize fixed-size panes.
                -- For the border AFTER a flex pane, we'd need its actual width.
                -- For now, only support resizing the first fixed pane (left divider).
                cumX = cumX + 0  -- placeholder; can't compute flex width
            end

            -- Check if this is a border between pane i and pane i+1
            if i < #currentSizes and s and s > 0 then
                -- Store this as a potential resize target
                -- We'll match on mouseMove delta
                dragRef.current = {
                    active = true,
                    paneIndex = i,
                    startPos = pos,
                    origSize = s,
                }
                -- Only handle the first fixed pane's right border for now
                -- (the click is somewhere — we'll refine on move)
                return
            end
        end

        -- If we get here, also check the LAST pane (right panel)
        -- The right panel's LEFT border is the resize handle
        local lastIdx = #currentSizes
        local lastSize = currentSizes[lastIdx]
        if lastSize and lastSize > 0 and lastIdx > 1 then
            dragRef.current = {
                active = true,
                paneIndex = lastIdx,
                startPos = pos,
                origSize = lastSize,
                isRight = true,  -- dragging left border of right pane
            }
        end
    end

    local function handleMouseMove(event)
        if not dragRef.current.active then return end
        local pos = (direction == "horizontal") and event.x or event.y
        local delta = pos - dragRef.current.startPos
        local idx = dragRef.current.paneIndex
        local origSize = dragRef.current.origSize
        if origSize == 0 then return end

        local newSize
        if dragRef.current.isRight then
            -- Right pane: dragging left = grow, dragging right = shrink
            newSize = origSize - delta
        else
            -- Left pane: dragging right = grow, dragging left = shrink
            newSize = origSize + delta
        end

        -- Clamp
        local minS = minSizes[idx]
        if not minS or minS == 0 then minS = 10 end
        local maxS = maxSizes[idx]
        if not maxS or maxS == 0 then maxS = 200 end
        newSize = math.max(minS, math.min(maxS, newSize))

        -- Build new sizes
        local newSizes = {}
        for j = 1, #currentSizes do
            newSizes[j] = currentSizes[j]
        end
        newSizes[idx] = newSize
        setCurrentSizes(newSizes)
        if onResize then onResize(newSizes) end
    end

    local function handleMouseUp(event)
        dragRef.current.active = false
    end

    -- Outer container
    local containerType = (direction == "horizontal") and "hbox" or "vbox"
    return lumina.createElement(containerType, {
        style = { flex = 1 },
        onMouseDown = handleMouseDown,
        onMouseMove = handleMouseMove,
        onMouseUp = handleMouseUp,
    }, table.unpack(elements))
end)

return SplitPane
