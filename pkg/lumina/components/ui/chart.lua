-- shadcn/chart — Chart components (Bar, Line, Pie)
local lumina = require("lumina")

local c = {
    fg = "#CDD6F4",
    muted = "#6C7086",
    primary = "#89B4FA",
    secondary = "#CBA6F7",
    tertiary = "#A6E3A1",
    quaternary = "#F9E2AF",
    quinary = "#F38BA8",
}

-- Bar chart
local Bar = lumina.defineComponent({
    name = "ShadcnChartBar",

    init = function(props)
        return {
            data = props.data or {},
            width = props.width or 30,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local data = self.data
        local maxVal = 0
        for _, v in ipairs(data) do
            if type(v) == "table" then v = v.value end
            if v > maxVal then maxVal = v end
        end
        if maxVal == 0 then maxVal = 1 end

        local colors = { c.primary, c.secondary, c.tertiary, c.quaternary, c.quinary }
        local lines = {}

        for i, entry in ipairs(data) do
            local label = type(entry) == "table" and (entry.label or tostring(i)) or tostring(i)
            local val = type(entry) == "table" and (entry.value or 0) or entry
            local barLen = math.floor((val / maxVal) * self.width)
            local bar = string.rep("█", barLen)
            local color = colors[((i - 1) % #colors) + 1]

            lines[#lines + 1] = {
                type = "hbox",
                children = {
                    { type = "text", content = label, style = { foreground = c.muted, width = 10 } },
                    { type = "text", content = bar, style = { foreground = color } },
                },
            }
        end

        return {
            type = "vbox",
            id = self.id,
            children = lines,
        }
    end,
})

-- Line chart (ASCII)
local Line = lumina.defineComponent({
    name = "ShadcnChartLine",

    init = function(props)
        return {
            data = props.data or {},
            width = props.width or 30,
            height = props.height or 10,
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local data = self.data
        local minVal, maxVal = math.huge, -math.huge
        for _, v in ipairs(data) do
            if type(v) == "table" then v = v.value end
            if v < minVal then minVal = v end
            if v > maxVal then maxVal = v end
        end
        if maxVal == minVal then maxVal = minVal + 1 end

        local range = maxVal - minVal
        local points = {}
        for i, v in ipairs(data) do
            if type(v) == "table" then v = v.value end
            local x = math.floor(((i - 1) / (#data - 1)) * (self.width - 1))
            local y = self.height - 1 - math.floor(((v - minVal) / range) * (self.height - 1))
            points[x] = y
        end

        -- Build grid
        local grid = {}
        for y = 0, self.height - 1 do
            local row = {}
            for x = 0, self.width - 1 do
                row[x + 1] = " "
            end
            grid[y + 1] = row
        end

        -- Draw line
        local prevX, prevY = -1, -1
        for x = 0, self.width - 1 do
            if points[x] ~= nil then
                local y = points[x]
                if prevX >= 0 then
                    -- Draw connecting line
                    local step = (y - prevY > 0) and 1 or -1
                    for py = prevY, y, step do
                        if py >= 0 and py < self.height then
                            grid[py + 1][x + 1] = "─"
                        end
                    end
                end
                grid[y + 1][x + 1] = "●"
                prevX, prevY = x, y
            end
        end

        local lines = {}
        for y = 0, self.height - 1 do
            lines[#lines + 1] = {
                type = "text",
                content = table.concat(grid[y + 1], ""),
                style = { foreground = c.primary },
            }
        end

        return {
            type = "vbox",
            id = self.id,
            children = lines,
        }
    end,
})

-- Pie chart (ASCII)
local Pie = lumina.defineComponent({
    name = "ShadcnChartPie",

    init = function(props)
        return {
            data = props.data or {},
            id = props.id,
            className = props.className,
            style = props.style,
        }
    end,

    render = function(self)
        local data = self.data
        local total = 0
        for _, v in ipairs(data) do
            local n = type(v) == "table" and (v.value or 0) or v
            total = total + n
        end

        local colors = { c.primary, c.secondary, c.tertiary, c.quaternary, c.quinary }
        local items = {}
        local slice = 0

        for i, entry in ipairs(data) do
            local label = type(entry) == "table" and (entry.label or tostring(i)) or tostring(i)
            local val = type(entry) == "table" and (entry.value or 0) or entry
            local pct = total > 0 and math.floor((val / total) * 100) or 0
            local color = colors[((i - 1) % #colors) + 1]

            items[#items + 1] = {
                type = "hbox",
                children = {
                    { type = "text", content = "◐", style = { foreground = color } },
                    { type = "text", content = " " .. label, style = { foreground = c.fg } },
                    { type = "text", content = " " .. pct .. "%", style = { foreground = c.muted } },
                },
            }
        end

        return {
            type = "vbox",
            id = self.id,
            children = items,
        }
    end,
})

return {
    Bar = Bar,
    Line = Line,
    Pie = Pie,
}