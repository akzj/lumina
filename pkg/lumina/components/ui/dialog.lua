-- shadcn/dialog — Modal dialog with overlay backdrop
local lumina = require("lumina")

local Dialog = lumina.defineComponent({
    name = "ShadcnDialog",
    init = function(props)
        return {
            open = props.open or false,
            title = props.title or "",
            description = props.description or "",
            width = props.width or 50,
            height = props.height or 20,
        }
    end,
    render = function(self)
        if not self.open then
            return { type = "fragment", children = {} }
        end
        local children = {}
        -- Title
        if self.title ~= "" then
            children[#children + 1] = {
                type = "text", content = self.title,
                style = { bold = true, foreground = "#F8FAFC" },
            }
        end
        -- Description
        if self.description ~= "" then
            children[#children + 1] = {
                type = "text", content = self.description,
                style = { foreground = "#94A3B8" },
            }
        end
        -- Separator
        children[#children + 1] = {
            type = "text",
            content = string.rep("─", self.width - 4),
            style = { foreground = "#334155" },
        }
        -- Slot for content
        if self.props and self.props.children then
            for _, child in ipairs(self.props.children) do
                children[#children + 1] = child
            end
        end
        return {
            type = "vbox",
            style = {
                border = "rounded",
                background = "#1E1E2E",
                foreground = "#CDD6F4",
                padding = 1,
                width = self.width,
                height = self.height,
            },
            children = children,
        }
    end
})

return Dialog
