-- shadcn/command — Command palette (like VS Code Ctrl+K)
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    bg = "#181825",
    surface = "#313244",
}

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
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        if not self.open then
            return { type = "empty" }
        end

        local style = {
            border = "rounded",
            borderColor = c.border,
            background = c.bg,
            foreground = c.fg,
            padding = 1,
            width = self.width,
        }
        if self.className and type(self.className) == "table" then
            for k, v in pairs(self.className) do style[k] = v end
        end
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do style[k] = v end
        end

        local children = {}

        -- Search input
        local searchDisplay = self.search ~= "" and self.search or self.placeholder
        local searchFg = self.search == "" and c.muted or c.fg
        children[#children + 1] = {
            type = "hbox",
            style = { paddingTop = 1 },
            children = {
                { type = "text", content = "🔍 " .. searchDisplay, style = { foreground = searchFg } },
            },
        }

        -- Separator
        children[#children + 1] = {
            type = "text",
            content = string.rep("─", self.width - 4),
            style = { foreground = c.border },
        }

        -- Groups and items
        if #self.groups > 0 then
            for _, group in ipairs(self.groups) do
                children[#children + 1] = {
                    type = "text",
                    content = "  " .. (group.heading or ""),
                    style = { foreground = c.muted, bold = true },
                }
                for _, item in ipairs(group.items or {}) do
                    children[#children + 1] = {
                        type = "text",
                        content = "    " .. (item.label or item),
                        style = { foreground = c.fg },
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
                    local pad = math.max(1, self.width - 8 - #label - #shortcut)
                    content = content .. string.rep(" ", pad) .. shortcut
                end
                children[#children + 1] = {
                    type = "text",
                    content = content,
                    style = {
                        foreground = isSelected and c.fg or c.muted,
                        background = isSelected and c.surface or "",
                    },
                }
            end
        end

        -- Empty state
        if #self.items == 0 and #self.groups == 0 then
            children[#children + 1] = {
                type = "text",
                content = "  No results found.",
                style = { foreground = c.muted },
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

return Command