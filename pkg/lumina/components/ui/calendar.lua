-- calendar.lua — Date picker calendar
local lumina = require("lumina")

local Calendar = lumina.defineComponent({
    name = "ShadcnCalendar",
    render = function(self)
        local year = self.props.year or 2025
        local month = self.props.month or 1
        local selected = self.props.selected -- {year, month, day}
        local onSelect = self.props.onSelect
        local weekStart = self.props.weekStart or 0 -- 0=Sunday

        local monthNames = {
            "January", "February", "March", "April", "May", "June",
            "July", "August", "September", "October", "November", "December"
        }
        local dayHeaders = { "Su", "Mo", "Tu", "We", "Th", "Fr", "Sa" }

        -- Days in month
        local daysInMonth = { 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31 }
        if year % 4 == 0 and (year % 100 ~= 0 or year % 400 == 0) then
            daysInMonth[2] = 29
        end

        -- First day of month (0=Sun, 1=Mon, ...)
        -- Zeller's formula simplified
        local y, m = year, month
        if m < 3 then y = y - 1; m = m + 12 end
        local firstDay = (1 + math.floor(13*(m+1)/5) + y + math.floor(y/4) - math.floor(y/100) + math.floor(y/400)) % 7

        -- Build header
        local headerChildren = {}
        for _, d in ipairs(dayHeaders) do
            headerChildren[#headerChildren + 1] = {
                type = "text", content = d,
                style = { foreground = "#6C7086", width = 3 },
            }
        end

        -- Build day grid
        local rows = {}
        local currentRow = {}
        -- Pad initial empty days
        for i = 1, firstDay do
            currentRow[#currentRow + 1] = { type = "text", content = "  ", style = { width = 3 } }
        end

        for day = 1, daysInMonth[self.props.month or 1] do
            local isSelected = selected and selected.year == year and selected.month == (self.props.month or 1) and selected.day == day
            local fg = isSelected and "#1E1E2E" or "#CDD6F4"
            local bg = isSelected and "#89B4FA" or ""
            local content = string.format("%2d", day)

            currentRow[#currentRow + 1] = {
                type = "text",
                content = content,
                style = { foreground = fg, background = bg, width = 3, bold = isSelected },
            }

            if #currentRow == 7 then
                rows[#rows + 1] = {
                    type = "hbox",
                    children = currentRow,
                }
                currentRow = {}
            end
        end
        if #currentRow > 0 then
            rows[#rows + 1] = { type = "hbox", children = currentRow }
        end

        -- Assemble
        local children = {
            -- Month/year header
            {
                type = "hbox",
                style = { justify = "center", padding = 0 },
                children = {
                    { type = "text", content = monthNames[self.props.month or 1] .. " " .. tostring(year),
                      style = { bold = true, foreground = "#CDD6F4" } },
                },
            },
            -- Day headers
            { type = "hbox", children = headerChildren },
            -- Separator
            { type = "text", content = string.rep("─", 21), style = { foreground = "#45475A" } },
        }
        for _, row in ipairs(rows) do
            children[#children + 1] = row
        end

        return {
            type = "vbox",
            style = {
                border = self.props.border or "rounded",
                padding = 1,
                width = self.props.width or 24,
                background = self.props.background or "#1E1E2E",
            },
            children = children,
        }
    end,
})

return Calendar
