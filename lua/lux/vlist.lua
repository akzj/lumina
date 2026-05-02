-- lua/lux/vlist.lua — VList: Virtual Scrolling List Component
-- Strategy: clip-on-idle (ScrollU-inspired). O(visible) memory when idle.
-- Variable height: measures actual item heights via ref prop, caches them.
-- Usage: local VList = require("lux.vlist")
--
-- Props:
--   totalCount (required)     - total number of items
--   renderItem(index) (required) - returns element for item at index (0-based)
--   overscan (default 10)     - items to keep beyond viewport
--   estimateHeight (default 1) - fallback height per item in rows
--   height (default 20)       - viewport height
--   clipDelay (default 500)   - ms idle before clipping

local VList = lumina.defineComponent("LuxVList", function(props)
	local totalCount = props.totalCount or 0
	local renderItem = props.renderItem
	local overscan = props.overscan or 10
	local estimateHeight = props.estimateHeight or 1
	local viewH = props.height or 20
	local clipDelay = props.clipDelay or 500

	-- Track scrollY via state (NOT ref -- avoids frame lag)
	local scrollY, setScrollY = lumina.useState("scrollY", 0)
	-- Track whether we're in "clipped" mode (idle) vs "accumulating" mode (active scroll)
	local isClipped, setClipped = lumina.useState("isClipped", false)
	-- Timer ref for clip-on-idle debounce
	local clipTimerRef = lumina.useRef()

	-- Variable height support
	local heightCache, setHeightCache = lumina.useState("heightCache", {})
	local itemRefs = lumina.useRef({}) -- itemRefs.current[i] = {current = nil}

	-- Get height for item at idx (cached actual or fallback estimate)
	local function getHeight(idx)
		return heightCache[idx] or estimateHeight
	end

	-- Find which item is at scrollY (linear scan accumulating heights)
	local function findIndexAt(targetY)
		local acc = 0
		for i = 0, totalCount - 1 do
			local h = getHeight(i)
			if acc + h > targetY then
				return i
			end
			acc = acc + h
		end
		return totalCount
	end

	-- Accumulated height up to (not including) idx
	local function heightUpTo(idx)
		local acc = 0
		for i = 0, idx - 1 do
			acc = acc + getHeight(i)
		end
		return acc
	end

	-- Calculate visible range using actual heights
	local effectiveOverscan = overscan
	if isClipped then
		effectiveOverscan = math.max(2, math.floor(overscan / 2))
	end

	local startIdx = math.max(0, findIndexAt(scrollY) - effectiveOverscan)
	local endIdx = math.min(totalCount, findIndexAt(scrollY + viewH) + effectiveOverscan)

	-- Build children
	local children = {}

	-- Top spacer: actual accumulated height
	local topH = heightUpTo(startIdx)
	if topH > 0 then
		children[#children + 1] = lumina.createElement("box", {
			key = "vlist_top",
			style = { height = topH },
		})
	end

	-- Visible items with ref wrappers for measurement
	for i = startIdx, endIdx - 1 do
		if not itemRefs.current[i] then
			itemRefs.current[i] = { current = nil }
		end
		local elem = renderItem(i)
		if elem then
			children[#children + 1] = lumina.createElement("box", {
				key = "vlist-item-" .. i,
				ref = itemRefs.current[i],
			}, elem)
		end
	end

	-- Bottom spacer
	local bottomH = heightUpTo(totalCount) - heightUpTo(endIdx)
	if bottomH > 0 then
		children[#children + 1] = lumina.createElement("box", {
			key = "vlist_bottom",
			style = { height = bottomH },
		})
	end

	-- Effect: measure heights after layout, update cache if changed
	lumina.useEffect(function()
		local changed = false
		local newCache = {}
		for k, v in pairs(heightCache) do
			newCache[k] = v
		end
		for i = startIdx, endIdx - 1 do
			local ref = itemRefs.current[i]
			if ref and ref.current and ref.current.h then
				local h = ref.current.h
				if newCache[i] ~= h then
					newCache[i] = h
					changed = true
				end
			end
		end
		if changed then
			setHeightCache(newCache)
		end
	end)

	-- onScroll handler: update scrollY, manage clip-on-idle timer
	local function onScroll(e)
		setClipped(false) -- active scrolling: exit clipped mode
		setScrollY(e.scrollY)

		-- Clip-on-idle: clear old timer, set new one
		if clipTimerRef.current then
			lumina.clearTimeout(clipTimerRef.current)
			clipTimerRef.current = nil
		end
		clipTimerRef.current = lumina.setTimeout(function()
			clipTimerRef.current = nil
			setClipped(true) -- idle: enter clipped mode, trigger re-render
		end, clipDelay)
	end

	return lumina.createElement("vbox", {
		id = props.id,
		style = {
			height = viewH,
			overflow = "scroll",
		},
		onScroll = onScroll,
	}, table.unpack(children))
end)

return VList
