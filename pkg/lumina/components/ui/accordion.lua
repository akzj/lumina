-- shadcn/accordion — Collapsible sections
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    border = "#45475A",
    surface = "#313244",
}

local Accordion = lumina.defineComponent({
    name = "ShadcnAccordion",

    init = function(props)
        return {
            items = props.items or {},
            type = props.type or "single", -- single, multiple
            defaultValue = props.defaultValue,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local children = {}

        for i, item in ipairs(self.items) do
            local title = item.title or item.label or ("Item " .. i)
            local content = item.content or item.children or ""

            -- Check if open (simplified: compare against defaultValue)
            local isOpen = false
            if type(self.defaultValue) == "number" then
                isOpen = (i == self.defaultValue)
            elseif type(self.defaultValue) == "string" then
                isOpen = (item.value or title) == self.defaultValue
            elseif type(self.defaultValue) == "table" then
                for _, v in ipairs(self.defaultValue) do
                    if v == i or v == (item.value or title) then
                        isOpen = true
                        break
                    end
                end
            end

            local chevron = isOpen and "▼" or "▶"

            -- Header row
            local headerStyle = {
                paddingLeft = 0,
                paddingRight = 0,
                paddingTop = 0,
                paddingBottom = 0,
            }
            local header = {
                type = "hbox",
                id = self.id and (self.id .. "-item-" .. i .. "-header") or nil,
                style = headerStyle,
                children = {
                    {
                        type = "text",
                        content = " " .. chevron .. " ",
                        style = { foreground = c.muted, bold = false },
                    },
                    {
                        type = "text",
                        content = title,
                        style = { foreground = c.fg, bold = true },
                    },
                },
            }

            -- Content
            local contentChildren
            if isOpen then
                if type(content) == "string" then
                    contentChildren = {
                        {
                            type = "vbox",
                            style = { paddingTop = 1, paddingLeft = 1 },
                            children = {
                                { type = "text", content = content, style = { foreground = c.muted } },
                            },
                        },
                    }
                else
                    contentChildren = content
                end
            else
                contentChildren = {}
            end

            -- Separator
            if i < #self.items then
                children[#children + 1] = {
                    type = "text",
                    content = "────────────────────────────────",
                    style = { foreground = c.surface },
                }
            end

            children[#children + 1] = header
            for _, c in ipairs(contentChildren) do
                children[#children + 1] = c
            end
        end

        return {
            type = "vbox",
            id = self.id,
            style = self.style or {},
            children = children,
        }
    end,
})

return Accordion
