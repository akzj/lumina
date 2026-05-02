-- examples/vlist_demo.lua — VList Demo: 50,000 virtual items
-- Run: lumina examples/vlist_demo.lua

local VList = require("lux.vlist")

local ITEM_COUNT = 50000

lumina.app {
	id = "vlist-demo",
	name = "VList Demo",
	render = function()
		return lumina.createElement(VList, {
			id = "vlist",
			totalCount = ITEM_COUNT,
			height = 20,
			overscan = 15,
			estimateHeight = 1,
			clipDelay = 500,
			renderItem = function(index)
				local bg = (index % 2 == 0) and "#1a1a2e" or "#16213e"
				return lumina.createElement("hbox", {
					key = "msg-" .. index,
					style = {
						height = 1,
						background = bg,
					},
				},
					lumina.createElement("text", {
						style = {
							foreground = "#7f8c8d",
							width = 6,
						},
					}, string.format("%5d", index)),
					lumina.createElement("text", {
						style = {
							foreground = "#ecf0f1",
						},
					}, "  Message number " .. (index + 1)),
				)
			end,
		})
	end,
}
