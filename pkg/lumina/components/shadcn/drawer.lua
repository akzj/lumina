-- shadcn/drawer — Bottom drawer panel
local lumina = require("lumina")

local Drawer = lumina.defineComponent({
    name = "ShadcnDrawer",
    init = function(props)
        return {
            open = props.open or false,
            title = props.title or "",
            height = props.height or 10,
        }
    end,
    render = function(self)
        if not self.open then
            return { type = "fragment", children = {} }
        end
        local children = {}
        -- Handle/grab bar
        children[#children + 1] = {
            type = "hbox",
            style = { justify = "center" },
            children = {
                { type = "text", content = "━━━━", style = { foreground = "#475569" } },
            },
        }
        if self.title ~= "" then
            children[#children + 1] = {
                type = "text", content = self.title,
                style = { bold = true, foreground = "#F8FAFC" },
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
                border = "rounded",
                background = "#0F172A",
                foreground = "#E2E8F0",
                padding = 1,
                height = self.height,
            },
            children = children,
        }
    end
})

return Drawer
