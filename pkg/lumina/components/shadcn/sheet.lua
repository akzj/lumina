-- shadcn/sheet — Side panel (slide in from edge)
local lumina = require("lumina")

local Sheet = lumina.defineComponent({
    name = "ShadcnSheet",
    init = function(props)
        return {
            open = props.open or false,
            side = props.side or "right",
            title = props.title or "",
            width = props.width or 40,
        }
    end,
    render = function(self)
        if not self.open then
            return { type = "fragment", children = {} }
        end
        local children = {}
        if self.title ~= "" then
            children[#children + 1] = {
                type = "text", content = self.title,
                style = { bold = true, foreground = "#F8FAFC" },
            }
            children[#children + 1] = {
                type = "text",
                content = string.rep("─", self.width - 4),
                style = { foreground = "#334155" },
            }
        end
        if self.props and self.props.children then
            for _, child in ipairs(self.props.children) do
                children[#children + 1] = child
            end
        end
        return {
            type = "vbox",
            style = {
                border = self.side == "left" and "single" or "single",
                background = "#0F172A",
                foreground = "#E2E8F0",
                padding = 1,
                width = self.width,
            },
            children = children,
        }
    end
})

return Sheet
