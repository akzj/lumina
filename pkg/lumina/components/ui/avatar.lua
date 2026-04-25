-- shadcn/avatar — Avatar with fallback text
local lumina = require("lumina")

local Avatar = lumina.defineComponent({
    name = "ShadcnAvatar",
    init = function(props)
        return {
            src = props.src or "",
            alt = props.alt or "",
            fallback = props.fallback or "?",
            size = props.size or "default",
        }
    end,
    render = function(self)
        local sizeMap = {
            sm      = { w = 3, h = 1 },
            default = { w = 5, h = 3 },
            lg      = { w = 7, h = 3 },
        }
        local sz = sizeMap[self.size] or sizeMap.default
        -- Terminal can't show images, so always show fallback
        local fb = string.sub(self.fallback, 1, sz.w)
        return {
            type = "hbox",
            style = {
                width = sz.w,
                height = sz.h,
                background = "#1E293B",
                foreground = "#E2E8F0",
                border = "rounded",
                justify = "center",
                align = "center",
            },
            children = {
                { type = "text", content = fb, style = { bold = true } },
            },
        }
    end
})

return Avatar
