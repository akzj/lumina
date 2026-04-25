-- shadcn/tabs — Tab navigation
local lumina = require("lumina")

local Tabs = lumina.defineComponent({
    name = "ShadcnTabs",
    init = function(props)
        return {
            tabs = props.tabs or {},
            activeTab = props.defaultValue or props.activeTab or 1,
            variant = props.variant or "default",
        }
    end,
    render = function(self)
        -- Tab list
        local tabButtons = {}
        for i, tab in ipairs(self.tabs) do
            local isActive = (i == self.activeTab)
            tabButtons[#tabButtons + 1] = {
                type = "text",
                content = " " .. (tab.label or tab.title or ("Tab " .. i)) .. " ",
                style = {
                    foreground = isActive and "#E2E8F0" or "#64748B",
                    background = isActive and "#1E293B" or "",
                    bold = isActive,
                    underline = isActive,
                },
            }
        end

        -- Active tab content
        local contentChildren = {}
        local activeTab = self.tabs[self.activeTab]
        if activeTab and activeTab.content then
            contentChildren[#contentChildren + 1] = {
                type = "text",
                content = activeTab.content,
                style = { foreground = "#E2E8F0" },
            }
        end

        return {
            type = "vbox",
            children = {
                {
                    type = "hbox",
                    style = { background = "#0F172A" },
                    children = tabButtons,
                },
                {
                    type = "text",
                    content = "────────────────────────────────",
                    style = { foreground = "#334155" },
                },
                {
                    type = "vbox",
                    style = { padding = 1 },
                    children = contentChildren,
                },
            },
        }
    end
})

return Tabs
