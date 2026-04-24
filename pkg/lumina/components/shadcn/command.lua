-- shadcn/command — Command palette (like VS Code Ctrl+K)
local lumina = require("lumina")

local Command = lumina.defineComponent({
    name = "ShadcnCommand",
    init = function(props)
        return {
            open = props.open or false,
            search = props.search or "",
            placeholder = props.placeholder or "Type a command...",
            items = props.items or {},
            groups = props.groups or {},
            selectedIndex = 0,
            width = props.width or 50,
        }
    end,
    render = function(self)
        if not self.open then
            return { type = "fragment", children = {} }
        end

        local children = {}

        -- Search input
        local searchDisplay = self.search ~= "" and self.search or self.placeholder
        local searchFg = self.search == "" and "#64748B" or "#E2E8F0"
        children[#children + 1] = {
            type = "hbox",
            style = { padding = 1 },
            children = {
                { type = "text", content = "🔍 " .. searchDisplay, style = { foreground = searchFg } },
            },
        }

        -- Separator
        children[#children + 1] = {
            type = "text",
            content = string.rep("─", self.width - 4),
            style = { foreground = "#334155" },
        }

        -- Groups and items
        if #self.groups > 0 then
            for _, group in ipairs(self.groups) do
                children[#children + 1] = {
                    type = "text",
                    content = "  " .. (group.heading or ""),
                    style = { foreground = "#64748B", bold = true },
                }
                for _, item in ipairs(group.items or {}) do
                    children[#children + 1] = {
                        type = "text",
                        content = "    " .. (item.label or item),
                        style = { foreground = "#E2E8F0" },
                    }
                end
            end
        else
            for i, item in ipairs(self.items) do
                local isSelected = (i == self.selectedIndex)
                local label = type(item) == "table" and (item.label or "") or item
                local shortcut = type(item) == "table" and (item.shortcut or "") or ""
                local content = "  " .. label
                if shortcut ~= "" then
                    local pad = self.width - 8 - #label - #shortcut
                    if pad < 1 then pad = 1 end
                    content = content .. string.rep(" ", pad) .. shortcut
                end
                children[#children + 1] = {
                    type = "text",
                    content = content,
                    style = {
                        foreground = isSelected and "#E2E8F0" or "#94A3B8",
                        background = isSelected and "#334155" or "",
                    },
                }
            end
        end

        -- Empty state
        if #self.items == 0 and #self.groups == 0 then
            children[#children + 1] = {
                type = "text",
                content = "  No results found.",
                style = { foreground = "#64748B" },
            }
        end

        return {
            type = "vbox",
            style = {
                border = "rounded",
                background = "#1E1E2E",
                foreground = "#CDD6F4",
                padding = 1,
                width = self.width,
            },
            children = children,
        }
    end
})

return Command
