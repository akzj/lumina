-- shadcn/carousel — Content carousel with navigation
local lumina = require("lumina")

local Carousel = lumina.defineComponent({
    name = "ShadcnCarousel",
    init = function(props)
        return {
            items = props.items or {},
            currentIndex = props.currentIndex or 1,
            showIndicators = props.showIndicators ~= false,
        }
    end,
    render = function(self)
        local children = {}
        local total = #self.items
        local idx = self.currentIndex

        -- Navigation + content
        local navChildren = {}
        navChildren[#navChildren + 1] = {
            type = "text",
            content = idx > 1 and "◄ " or "  ",
            style = { foreground = idx > 1 and "#3B82F6" or "#475569" },
        }

        -- Current item
        local currentItem = self.items[idx]
        if currentItem then
            if type(currentItem) == "table" then
                navChildren[#navChildren + 1] = currentItem
            else
                navChildren[#navChildren + 1] = {
                    type = "text",
                    content = tostring(currentItem),
                    style = { foreground = "#E2E8F0" },
                }
            end
        end

        navChildren[#navChildren + 1] = {
            type = "text",
            content = idx < total and " ►" or "  ",
            style = { foreground = idx < total and "#3B82F6" or "#475569" },
        }

        children[#children + 1] = {
            type = "hbox",
            style = { justify = "center", align = "center" },
            children = navChildren,
        }

        -- Dot indicators
        if self.showIndicators and total > 1 then
            local dots = ""
            for i = 1, total do
                dots = dots .. (i == idx and "●" or "○") .. " "
            end
            children[#children + 1] = {
                type = "hbox",
                style = { justify = "center" },
                children = {
                    { type = "text", content = dots, style = { foreground = "#64748B" } },
                },
            }
        end

        return {
            type = "vbox",
            children = children,
        }
    end
})

return Carousel
