-- lua/lux/scrollview.lua — ScrollView: scrollable container with keyboard + scrollbar support
-- Usage: local ScrollView = require("lux.scrollview")
--
-- Props:
--   height (default 20)       - viewport height
--   id (optional)             - container ID for lumina.scrollNode
--   onScroll(e) (optional)    - scroll event handler
--   onKeyDown(key) (optional) - additional key handler (called after built-in keys)
--   children                  - content to render inside the scroll container
--   ...all other props passed through to the vbox

local ScrollView = lumina.defineComponent("LuxScrollView", function(props)
	local scrollY, setScrollY = lumina.useState("scrollY", 0)
	local containerID = props.id or "scrollview"
	local viewH = props.height or 20

	local function onScroll(e)
		setScrollY(e.scrollY)
		if props.onScroll then
			props.onScroll(e)
		end
	end

	local function onKeyDown(e)
		local key = e.key
		local step = 3 -- same as engine autoScroll step
		if key == "Up" or key == "k" then
			lumina.scrollNode(containerID, -step)
		elseif key == "Down" or key == "j" then
			lumina.scrollNode(containerID, step)
		elseif key == "PageUp" then
			lumina.scrollNode(containerID, -viewH)
		elseif key == "PageDown" then
			lumina.scrollNode(containerID, viewH)
		elseif key == "Home" then
			lumina.scrollNode(containerID, -999999) -- clamped to 0
		elseif key == "End" then
			lumina.scrollNode(containerID, 999999) -- clamped to maxScroll
		elseif props.onKeyDown then
			props.onKeyDown(e)
		end
	end

	-- Pass through all props except the ones we handle specially
	local passProps = {}
	for k, v in pairs(props) do
		if k ~= "height" and k ~= "onKeyDown" and k ~= "onScroll" and k ~= "id" and k ~= "children" then
			passProps[k] = v
		end
	end
	passProps.id = containerID
	passProps.style = passProps.style or {}
	-- Only set fixed height if explicitly provided via props.height.
	-- When no height prop is given and style has flex, let flex determine height.
	if props.height then
		passProps.style.height = props.height
	elseif not passProps.style.flex then
		passProps.style.height = 20  -- fallback only when no flex and no explicit height
	end
	passProps.style.overflow = "scroll"
	-- Apply scrollbar colors: explicit props > theme > engine defaults
	if props.scrollbarThumbColor then
		passProps.style.scrollbarThumbColor = props.scrollbarThumbColor
	elseif lumina.getTheme then
		local t = lumina.getTheme()
		if t and t.primary then
			passProps.style.scrollbarThumbColor = t.primary
		end
	end
	if props.scrollbarTrackColor then
		passProps.style.scrollbarTrackColor = props.scrollbarTrackColor
	elseif lumina.getTheme then
		local t = lumina.getTheme()
		if t and t.surface0 then
			passProps.style.scrollbarTrackColor = t.surface0
		end
	end
	passProps.onScroll = onScroll
	passProps.onKeyDown = onKeyDown
	passProps.focusable = true

	return lumina.createElement("vbox", passProps, table.unpack(props.children or {}))
end)

return ScrollView
