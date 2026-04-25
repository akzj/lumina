-- shadcn/carousel — Carousel/slideshow
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
}

local Carousel = lumina.defineComponent({
    name = "ShadcnCarousel",

    init = function(props)
        return {
            items = props.items or {},
            activeIndex = props.activeIndex or 0,
            orientation = props.orientation or "horizontal",
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local idx = math.max(0, math.min(self.activeIndex, #self.items - 1))
        local activeItem = self.items[idx + 1]

        local style = {
            border = "rounded",
            borderColor = c.border,
            background = c.bg,
            foreground = c.fg,
            padding = 1,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        local children = {}

        -- Slide content
        if activeItem then
            local content = type(activeItem) == "string" and activeItem or (activeItem.content or activeItem.label or "")
            children[#children + 1] = {
                type = "text",
                content = tostring(content),
                style = { foreground = c.fg },
            }
        end

        -- Indicators
        if #self.items > 1 then
            local dots = ""
            for i = 1, #self.items do
                dots = dots .. (i == idx + 1 and "◉" or "○") .. " "
            end
            children[#children + 1] = {
                type = "hbox",
                style = { justify = "center", paddingTop = 1 },
                children = {
                    { type = "text", content = dots, style = { foreground = c.primary } },
                },
            }
        end

        return {
            type = "vbox",
            id = self.id,
            style = style,
            children = children,
        }
    end,
})

return Carousel