-- lua/lux/vlist.lua — VList: Virtual Scrolling List Component
-- Strategy: clip-on-idle (ScrollU-inspired). O(visible) memory when idle.
-- Usage: local VList = require("lux.vlist")
--
-- Props:
--   totalCount (required)     - total number of items
--   renderItem(index) (required) - returns element for item at index (0-based)
--   overscan (default 10)     - items to keep beyond viewport
--   estimateHeight (default 1) - height per item in rows
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

	-- Calculate visible range.
	-- During active scroll: use full overscan (accumulate).
	-- After clip (idle): use reduced overscan (clipRange = overscan//2, min 2).
	local effectiveOverscan = overscan
	if isClipped then
		effectiveOverscan = math.max(2, math.floor(overscan / 2))
	end

	local startIdx = math.max(0, math.floor(scrollY / estimateHeight) - effectiveOverscan)
	local endIdx = math.min(totalCount, math.ceil((scrollY + viewH) / estimateHeight) + effectiveOverscan)

	-- Build children
	local children = {}

	-- Top spacer
	if startIdx > 0 then
		children[#children + 1] = lumina.createElement("box", {
			key = "vlist_top",
			style = { height = startIdx * estimateHeight },
		})
	end

	-- Visible items (reconciler reuses by key)
	for i = startIdx, endIdx - 1 do
		local elem = renderItem(i)
		if elem then
			children[#children + 1] = elem
		end
	end

	-- Bottom spacer
	if endIdx < totalCount then
		children[#children + 1] = lumina.createElement("box", {
			key = "vlist_bottom",
			style = { height = (totalCount - endIdx) * estimateHeight },
		})
	end

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
		key = props.key,
		id = props.id,
		style = {
			height = viewH,
			overflow = "scroll",
		},
		onScroll = onScroll,
	}, table.unpack(children))
end)

return VList
