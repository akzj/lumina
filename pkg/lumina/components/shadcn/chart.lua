-- chart.lua — Chart components using SubPixelCanvas concepts
local lumina = require("lumina")

local BarChart = lumina.defineComponent({
    name = "ShadcnBarChart",
    render = function(self)
        local data = self.props.data or {}
        local height = self.props.height or 10
        local width = self.props.width or 40
        local barWidth = self.props.barWidth or 3
        local gap = self.props.gap or 1
        local showLabels = self.props.showLabels ~= false
        local showValues = self.props.showValues ~= false
        local color = self.props.color or "#89B4FA"

        -- Find max value
        local maxVal = 0
        for _, d in ipairs(data) do
            if (d.value or 0) > maxVal then maxVal = d.value or 0 end
        end
        if maxVal == 0 then maxVal = 1 end

        local children = {}

        -- Build bars
        local barRow = {}
        for _, d in ipairs(data) do
            local barHeight = math.floor((d.value or 0) / maxVal * (height - 2))
            if barHeight < 1 then barHeight = 1 end

            local barContent = ""
            for i = 1, barHeight do
                barContent = barContent .. "█"
                if i < barHeight then barContent = barContent .. "\n" end
            end

            local barChildren = {}
            if showValues then
                barChildren[#barChildren + 1] = {
                    type = "text",
                    content = tostring(d.value or 0),
                    style = { foreground = "#CDD6F4", dim = true },
                }
            end
            barChildren[#barChildren + 1] = {
                type = "text",
                content = string.rep("█\n", barHeight):sub(1, -2),
                style = {
                    foreground = d.color or color,
                    height = barHeight,
                    width = barWidth,
                },
            }
            if showLabels then
                barChildren[#barChildren + 1] = {
                    type = "text",
                    content = d.label or "",
                    style = { foreground = "#6C7086", width = barWidth },
                }
            end

            barRow[#barRow + 1] = {
                type = "vbox",
                style = { align = "flex-end", width = barWidth + gap },
                children = barChildren,
            }
        end

        children[#children + 1] = {
            type = "hbox",
            style = { align = "flex-end", height = height },
            children = barRow,
        }

        return {
            type = "vbox",
            style = {
                width = width,
                height = height,
                border = self.props.border or "",
                background = self.props.background or "",
            },
            children = children,
        }
    end,
})

local LineChart = lumina.defineComponent({
    name = "ShadcnLineChart",
    render = function(self)
        local data = self.props.data or {}
        local height = self.props.height or 10
        local width = self.props.width or 40
        local color = self.props.color or "#89B4FA"
        local showDots = self.props.showDots ~= false

        -- Find min/max
        local minVal, maxVal = math.huge, -math.huge
        for _, d in ipairs(data) do
            local v = d.value or 0
            if v < minVal then minVal = v end
            if v > maxVal then maxVal = v end
        end
        if minVal == maxVal then maxVal = minVal + 1 end

        -- Build text-based line chart
        local grid = {}
        for y = 1, height do
            grid[y] = {}
            for x = 1, width do
                grid[y][x] = " "
            end
        end

        -- Plot points
        local n = #data
        for i, d in ipairs(data) do
            local x = math.floor((i - 1) / math.max(n - 1, 1) * (width - 1)) + 1
            local y = height - math.floor((d.value - minVal) / (maxVal - minVal) * (height - 1))
            if y < 1 then y = 1 end
            if y > height then y = height end
            if showDots then
                grid[y][x] = "●"
            else
                grid[y][x] = "─"
            end
        end

        -- Render grid to text
        local lines = {}
        for y = 1, height do
            lines[#lines + 1] = {
                type = "text",
                content = table.concat(grid[y]),
                style = { foreground = color },
            }
        end

        return {
            type = "vbox",
            style = {
                width = width,
                height = height,
                border = self.props.border or "",
            },
            children = lines,
        }
    end,
})

local PieChart = lumina.defineComponent({
    name = "ShadcnPieChart",
    render = function(self)
        local data = self.props.data or {}
        local width = self.props.width or 30
        local colors = { "#89B4FA", "#F38BA8", "#A6E3A1", "#F9E2AF", "#CBA6F7", "#89DCEB" }

        -- Calculate total
        local total = 0
        for _, d in ipairs(data) do total = total + (d.value or 0) end
        if total == 0 then total = 1 end

        -- Build legend-style pie chart (terminal-friendly)
        local children = {}
        for i, d in ipairs(data) do
            local pct = math.floor((d.value or 0) / total * 100)
            local barLen = math.floor(pct / 100 * (width - 15))
            if barLen < 1 then barLen = 1 end
            local color = d.color or colors[((i - 1) % #colors) + 1]

            children[#children + 1] = {
                type = "hbox",
                children = {
                    { type = "text", content = string.rep("█", barLen),
                      style = { foreground = color, width = width - 15 } },
                    { type = "text", content = string.format(" %s (%d%%)", d.label or "", pct),
                      style = { foreground = "#CDD6F4" } },
                },
            }
        end

        return {
            type = "vbox",
            style = {
                width = width,
                border = self.props.border or "",
            },
            children = children,
        }
    end,
})

return {
    Bar = BarChart,
    Line = LineChart,
    Pie = PieChart,
}
