-- lua/lux/window.lua — Draggable, Resizable Window Component
-- Usage: local Window = require("lux.window")
--
-- Props:
--   id (required)      - window id (used for WM)
--   title (required)   - window title
--   x, y, w, h         - position and size
--   isActive            - whether this window is active
--   onActivate()        - called when window is clicked (for WM.activate)
--   onMove(x, y)        - called during drag
--   onResize(w, h)      - called during resize
--   children            - window content (rendered directly, no wrapper)

local Window = lumina.defineComponent("LuxWindow", function(props)
	local id = props.id
	local title = props.title or id
	local x = props.x or 0
	local y = props.y or 0
	local w = props.w or 30
	local h = props.h or 10
	local isActive = props.isActive or false

	local t = lumina.getTheme()
	local borderColor = isActive and t.primary or t.surface1
	local titleBg = isActive and t.primary or t.surface1
	local titleFg = isActive and t.base or t.text
	local bg = isActive and t.surface0 or t.base

	local dragRef = lumina.useRef({active = false, startX = 0, startY = 0, origX = 0, origY = 0})
	local resizeRef = lumina.useRef({active = false, startX = 0, startY = 0, origW = 0, origH = 0})

	-- Build title bar text matching original format
	local titleText = " " .. title .. string.rep(" ", math.max(0, w - #title - 4))

	local children = {
		-- Title bar
		lumina.createElement("text", {
			bold = true,
			foreground = titleFg,
			background = titleBg,
		}, titleText),
	}

	-- Add user children directly
	if props.children then
		for _, child in ipairs(props.children) do
			children[#children + 1] = child
		end
	end

	return lumina.createElement("vbox", {
		key = id,
		style = {
			position = "absolute",
			left = x, top = y,
			width = w, height = h,
			border = "rounded",
			background = bg,
		},
		onClick = function()
			if props.onActivate then props.onActivate() end
		end,
		onMouseDown = function(mx, my)
			if mx >= x + w - 2 and my >= y + h - 2 then
				resizeRef.current = {
					active = true,
					startX = mx, startY = my,
					origW = w, origH = h,
				}
				return
			end
			if my == y then
				dragRef.current = {
					active = true,
					startX = mx, startY = my,
					origX = x, origY = y,
				}
			end
		end,
		onMouseMove = function(mx, my)
			if dragRef.current.active then
				local dx = mx - dragRef.current.startX
				local dy = my - dragRef.current.startY
				local newX = math.max(0, dragRef.current.origX + dx)
				local newY = math.max(0, dragRef.current.origY + dy)
				if props.onMove then props.onMove(newX, newY) end
			elseif resizeRef.current.active then
				local dx = mx - resizeRef.current.startX
				local dy = my - resizeRef.current.startY
				local newW = math.max(10, resizeRef.current.origW + dx)
				local newH = math.max(3, resizeRef.current.origH + dy)
				if props.onResize then props.onResize(newW, newH) end
			end
		end,
		onMouseUp = function()
			dragRef.current.active = false
			resizeRef.current.active = false
		end,
	}, table.unpack(children))
end)

return Window
