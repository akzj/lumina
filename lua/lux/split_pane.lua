-- lux/split_pane.lua — Resizable split pane layout
-- Renders children in a horizontal or vertical split with draggable borders.
--
-- Props:
--   direction: "horizontal" (default) or "vertical"
--   sizes: array of sizes (number > 0 = fixed width/height, 0 = flex)
--   minSizes: array of minimum sizes (optional, default 10)
--   maxSizes: array of maximum sizes (optional, default 200)
--   borderStyle: "single" (default), "double", "rounded", or nil (no border)
--   borderColor: string (border color for panes)
--   background: string or array of background colors per pane
--   onResize: function(newSizes) — called when user drags a divider
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

    -- Sizes: if onResize is provided, use props.sizes directly (controlled mode).
    -- Otherwise, use internal state (uncontrolled mode).
    local internalSizes, setInternalSizes = lumina.useState("splitSizes", initialSizes)
    local currentSizes, setCurrentSizes
    if onResize then
        -- Controlled: parent manages sizes via onResize callback
        currentSizes = initialSizes
        setCurrentSizes = function() end  -- no-op; parent updates via onResize
    else
        -- Uncontrolled: SplitPane manages its own sizes
        currentSizes = internalSizes
        setCurrentSizes = setInternalSizes
    end

    -- Drag state (mutable ref, persists across renders)
    local dragRef = lumina.useRef({ active = false, dividerIndex = 0, startPos = 0, origSize = 0 })

    -- Build the element list: pane, divider, pane, divider, pane, ...
    local elements = {}
    for i, child in ipairs(children) do
        -- Pane style (0 = flex, >0 = fixed size)
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

        -- Divider between panes (not after the last one)
        if i < #children then
            local divStyle = {}
            if direction == "horizontal" then
                divStyle.width = 1
                divStyle.height = "100%"
            else
                divStyle.height = 1
                divStyle.width = "100%"
            end
            divStyle.background = borderColor

            local divIdx = i  -- capture for closure
            elements[#elements + 1] = lumina.createElement("vbox", {
                key = "split-div-" .. i,
                style = divStyle,
                onMouseDown = function(event)
                    local startPos
                    if direction == "horizontal" then
                        startPos = event.x
                    else
                        startPos = event.y
                    end
                    dragRef.current = {
                        active = true,
                        dividerIndex = divIdx,
                        startPos = startPos,
                        origSize = currentSizes[divIdx] or 0,
                    }
                end,
                onMouseMove = function(event)
                    if not dragRef.current.active then return end
                    local pos
                    if direction == "horizontal" then
                        pos = event.x
                    else
                        pos = event.y
                    end
                    local delta = pos - dragRef.current.startPos
                    local idx = dragRef.current.dividerIndex
                    local origSize = dragRef.current.origSize
                    if origSize == 0 then return end  -- can't resize flex pane

                    local newSize = origSize + delta

                    -- Clamp to min/max
                    local minS = minSizes[idx]
                    if not minS or minS == 0 then minS = 10 end
                    local maxS = maxSizes[idx]
                    if not maxS or maxS == 0 then maxS = 200 end
                    newSize = math.max(minS, math.min(maxS, newSize))

                    -- Build new sizes array
                    local newSizes = {}
                    for j = 1, #currentSizes do
                        newSizes[j] = currentSizes[j]
                    end
                    newSizes[idx] = newSize
                    setCurrentSizes(newSizes)
                    if onResize then onResize(newSizes) end
                end,
                onMouseUp = function(event)
                    dragRef.current.active = false
                end,
            })
        end
    end

    -- Outer container
    local containerType = (direction == "horizontal") and "hbox" or "vbox"
    return lumina.createElement(containerType, {
        style = { flex = 1 },
    }, table.unpack(elements))
end)

return SplitPane
