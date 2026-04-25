-- color_picker.lua — Color selection component
local lumina = require("lumina")

local ColorPicker = lumina.defineComponent({
    name = "ShadcnColorPicker",
    render = function(self)
        local value = self.props.value or "#89B4FA"
        local onSelect = self.props.onSelect
        local presets = self.props.presets or {
            "#F38BA8", "#FAB387", "#F9E2AF", "#A6E3A1",
            "#89DCEB", "#89B4FA", "#CBA6F7", "#F5C2E7",
            "#EBA0AC", "#F2CDCD", "#B4BEFE", "#94E2D5",
            "#74C7EC", "#BAC2DE", "#A6ADC8", "#6C7086",
        }

        local rows = {}
        local currentRow = {}
        local colsPerRow = self.props.columns or 8

        for i, color in ipairs(presets) do
            local isSelected = color:lower() == value:lower()
            currentRow[#currentRow + 1] = {
                type = "text",
                content = isSelected and "◉" or "●",
                style = {
                    foreground = color,
                    bold = isSelected,
                    width = 2,
                },
            }
            if #currentRow >= colsPerRow then
                rows[#rows + 1] = { type = "hbox", children = currentRow }
                currentRow = {}
            end
        end
        if #currentRow > 0 then
            rows[#rows + 1] = { type = "hbox", children = currentRow }
        end

        -- Show current selection
        local children = {
            {
                type = "hbox",
                children = {
                    { type = "text", content = "██", style = { foreground = value } },
                    { type = "text", content = " " .. value, style = { foreground = "#CDD6F4" } },
                },
            },
            { type = "text", content = string.rep("─", 16), style = { foreground = "#45475A" } },
        }
        for _, row in ipairs(rows) do
            children[#children + 1] = row
        end

        return {
            type = "vbox",
            style = {
                border = self.props.border or "rounded",
                padding = 1,
                width = self.props.width or 20,
                background = self.props.background or "#1E1E2E",
            },
            children = children,
        }
    end,
})

return ColorPicker
