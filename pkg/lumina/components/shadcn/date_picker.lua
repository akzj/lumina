-- date_picker.lua — Date input with calendar popup
local lumina = require("lumina")

local DatePicker = lumina.defineComponent({
    name = "ShadcnDatePicker",
    render = function(self)
        local value = self.props.value -- {year, month, day} or nil
        local onSelect = self.props.onSelect
        local placeholder = self.props.placeholder or "Pick a date"
        local open = self.props.open or false
        local onToggle = self.props.onToggle

        local displayText = placeholder
        if value then
            displayText = string.format("%04d-%02d-%02d", value.year, value.month, value.day)
        end

        local children = {
            {
                type = "hbox",
                style = {
                    border = "rounded",
                    padding = 0,
                    background = "#1E1E2E",
                    foreground = value and "#CDD6F4" or "#6C7086",
                    width = self.props.width or 20,
                },
                children = {
                    { type = "text", content = "📅 " .. displayText },
                },
            },
        }

        if open then
            children[#children + 1] = lumina.createElement("ShadcnCalendar", {
                year = value and value.year or 2025,
                month = value and value.month or 1,
                selected = value,
                onSelect = onSelect,
            })
        end

        return {
            type = "vbox",
            children = children,
        }
    end,
})

return DatePicker
