-- lux/split_pane.lua — Resizable split pane layout
-- Renders children in horizontal or vertical split with single-line dividers.
--
-- Props:
--   direction: "horizontal" (default) or "vertical"
--   sizes: array of sizes (number > 0 = fixed width/height, 0 = flex)
--   minSizes: array of minimum sizes (optional, default 10)
--   maxSizes: array of maximum sizes (optional, default 200)
--   borderStyle: "single", "double", "rounded", or nil (no pane border)
--   borderColor: string (border/divider color)
--   divider: true (default) — show 1-col vertical divider between panes
--   background: string or array of background colors per pane
--   onResize: function(newSizes) — called when user drags a divider
--   children: pane contents (passed as createElement children)
--
-- Note: Use 0 (not nil) to indicate a flex pane.
--
-- Usage:
--   lumina.createElement(SplitPane, {
--       direction = "horizontal",
--       sizes = { 34, 0, 36 },
--       divider = true,
--       borderColor = "#38605c",
--       onResize = function(newSizes) end,
--   }, pane1, pane2, pane3)

local SplitPane = lumina.defineComponent("SplitPane", function(props)
    local direction = props.direction or "horizontal"
    local initialSizes = props.sizes or {}
    local minSizes = props.minSizes or {}
    local maxSizes = props.maxSizes or {}
    local borderStyle = props.borderStyle  -- nil = no border on panes
    local borderColor = props.borderColor
    local showDivider = (props.divider ~= false)  -- default true
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

    -- Build pane elements with optional dividers between them
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

        -- Optional border on panes
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

        -- Fill parent height
        paneStyle.height = "100%"

        -- Ensure child has a key to prevent defineComponent dedup bug
        -- (without key, sibling defineComponent elements in hbox may only render the last one)
        if type(child) == "table" and child.key == nil then
            child.key = "split-child-" .. i
        end

        elements[#elements + 1] = lumina.createElement("vbox", {
            key = "split-pane-" .. i,
            style = paneStyle,
        }, child)

        -- Divider between panes (not after the last one)
        if showDivider and i < #children then
            local divStyle = {}
            if direction == "horizontal" then
                divStyle.width = 1
            else
                divStyle.height = 1
            end
            if borderColor then
                divStyle.background = borderColor
            end

            -- The divider is a 1-col vbox that acts as drag handle
            local divIdx = i
            elements[#elements + 1] = lumina.createElement("vbox", {
                key = "split-div-" .. i,
                style = divStyle,
                onMouseDown = function(event)
                    local startPos = (direction == "horizontal") and event.x or event.y
                    local idx = divIdx
                    local origSize = currentSizes[divIdx] or 0
                    local sign = 1
                    -- If the left pane is flex (size=0), drag the right pane instead.
                    -- This enables resizing when the divider is between a flex and a fixed pane.
                    if origSize == 0 and divIdx < #currentSizes then
                        idx = divIdx + 1
                        origSize = currentSizes[idx] or 0
                        sign = -1
                    end
                    dragRef.current = {
                        active = true,
                        paneIndex = idx,
                        startPos = startPos,
                        origSize = origSize,
                        sign = sign,
                    }
                end,
                onMouseMove = function(event)
                    if not dragRef.current.active then return end
                    local pos = (direction == "horizontal") and event.x or event.y
                    local delta = pos - dragRef.current.startPos
                    local idx = dragRef.current.paneIndex
                    local origSize = dragRef.current.origSize
                    if origSize == 0 then return end

                    local newSize = origSize + delta * (dragRef.current.sign or 1)

                    local minS = minSizes[idx]
                    if not minS or minS == 0 then minS = 10 end
                    local maxS = maxSizes[idx]
                    if not maxS or maxS == 0 then maxS = 200 end
                    newSize = math.max(minS, math.min(maxS, newSize))

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
        style = { flex = 1, width = "100%" },
    }, table.unpack(elements))
end)

return SplitPane
