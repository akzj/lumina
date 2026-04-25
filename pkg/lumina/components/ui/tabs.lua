-- shadcn/tabs — Tab navigation with keyboard support
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    primary_fg = "#1E1E2E",
    border = "#45475A",
    surface = "#313244",
    bg = "#1E1E2E",
}

local Tabs = lumina.defineComponent({
    name = "ShadcnTabs",

    init = function(props)
        return {
            tabs = props.tabs or {},
            value = props.value or (props.defaultValue or 1),
            variant = props.variant or "default",
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local activeIdx = self.value
        local tabs = self.tabs

        -- Tab buttons
        local tabButtons = {}
        for i, tab in ipairs(tabs) do
            local isActive = (i == activeIdx)
            local label = tab.label or tab.title or ("Tab " .. i)

            local btnStyle = {
                foreground = isActive and c.primary or c.muted,
                background = isActive and c.surface or "",
                bold = isActive,
                paddingLeft = 1,
                paddingRight = 1,
            }
            if isActive then
                btnStyle.borderBottom = "single"
                btnStyle.borderColor = c.primary
            end

            local btn = {
                type = "hbox",
                id = self.id and (self.id .. "-tab-" .. i) or nil,
                style = btnStyle,
                children = {
                    { type = "text", content = " " .. label .. " " },
                },
            }

            -- Add click handler
            btn.onClick = function()
                -- Tab change is handled by parent via onValueChange
            end

            tabButtons[#tabButtons + 1] = btn

            -- Separator between tabs
            if i < #tabs then
                tabButtons[#tabButtons + 1] = {
                    type = "text",
                    content = "│",
                    style = { foreground = c.border },
                }
            end
        end

        -- Active tab content
        local content = ""
        local activeTab = tabs[activeIdx]
        if activeTab then
            if type(activeTab.content) == "string" then
                content = activeTab.content
            elseif activeTab.content and activeTab.content.children then
                -- Pass through VNode content
            end
        end

        local contentBox
        if type(activeTab) == "table" and activeTab.content then
            if activeTab.content.type then
                contentBox = activeTab.content
            else
                contentBox = {
                    type = "vbox",
                    children = {
                        { type = "text", content = tostring(activeTab.content) },
                    },
                }
            end
        else
            contentBox = {
                type = "vbox",
                children = {
                    { type = "text", content = content },
                },
            }
        end

        local containerStyle = {}
        if self.style and type(self.style) == "table" then
            for k, v in pairs(self.style) do containerStyle[k] = v end
        end

        return {
            type = "vbox",
            id = self.id,
            style = containerStyle,
            children = {
                {
                    type = "hbox",
                    id = self.id and (self.id .. "-tablist") or nil,
                    style = { background = c.bg, borderBottom = "single", borderColor = c.border },
                    children = tabButtons,
                },
                {
                    type = "vbox",
                    id = self.id and (self.id .. "-content") or nil,
                    style = { paddingTop = 1 },
                    children = contentBox and { contentBox } or {},
                },
            },
        }
    end,
})

return Tabs
