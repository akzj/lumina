-- Lumina v2 Built-in Component Library
-- Reusable UI components built on top of lumina.createElement.
--
-- NOTE: require() for local files is not yet supported in v2.
-- For now, inline these functions in your scripts, or use dofile().
-- This file serves as the canonical reference implementation.
--
-- Components:
--   M.ProgressBar(props)  — horizontal progress bar with percentage
--   M.Table(props)        — data table with headers and rows
--   M.Tabs(props)         — tab bar with switchable content panels
--   M.Modal(props)        — overlay dialog box with border
--   M.Select(props)       — list selector with highlighted selection
--
-- Theme: Catppuccin Mocha

local M = {}

-- ═══════════════════════════════════════════════════════════════════
-- ProgressBar
-- ═══════════════════════════════════════════════════════════════════
-- props:
--   value    : number 0-100 (default 0)
--   width    : bar character width (default 20)
--   color    : bar fill color (default "#A6E3A1")
--   bgColor  : bar empty color (default "#313244")
--   label    : optional label string (shown before bar)
--
-- Returns: hbox with [label] [████░░░░] [67%]

function M.ProgressBar(props)
    local value = math.max(0, math.min(100, props.value or 0))
    local width = props.width or 20
    local color = props.color or "#A6E3A1"
    local bgColor = props.bgColor or "#313244"
    local label = props.label or ""

    local filled = math.floor(value / 100 * width)
    local empty = width - filled
    local bar = string.rep("█", filled) .. string.rep("░", empty)
    local pct = string.format("%3d%%", value)

    -- Color the percentage based on value thresholds
    local pctColor = "#A6E3A1"  -- green
    if value > 80 then
        pctColor = "#F38BA8"    -- red
    elseif value > 60 then
        pctColor = "#F9E2AF"    -- yellow
    end

    local children = {}
    if label ~= "" then
        children[#children + 1] = lumina.createElement("text", {
            foreground = "#CDD6F4",
        }, label)
    end
    children[#children + 1] = lumina.createElement("text", {
        foreground = color,
    }, bar)
    children[#children + 1] = lumina.createElement("text", {
        foreground = pctColor,
    }, pct)

    return lumina.createElement("hbox", {
        style = {gap = 1},
    }, table.unpack(children))
end

-- ═══════════════════════════════════════════════════════════════════
-- Table
-- ═══════════════════════════════════════════════════════════════════
-- props:
--   headers     : array of column header strings
--   rows        : array of row arrays (each row = array of cell values)
--   colWidths   : optional array of column widths (auto-calculated if nil)
--   selectedRow : 1-based index of highlighted row (-1 or nil = none)
--   style       : optional style table for the outer vbox
--
-- Returns: vbox with header row, separator, and data rows

function M.Table(props)
    local headers = props.headers or {}
    local rows = props.rows or {}
    local selectedRow = props.selectedRow or -1
    local colWidths = props.colWidths

    -- Auto-calculate column widths if not provided
    if not colWidths then
        colWidths = {}
        for i, h in ipairs(headers) do
            colWidths[i] = #tostring(h) + 2
        end
        for _, row in ipairs(rows) do
            for i, cell in ipairs(row) do
                local w = #tostring(cell) + 2
                if w > (colWidths[i] or 0) then
                    colWidths[i] = w
                end
            end
        end
    end

    -- Calculate total width for separator
    local totalWidth = 0
    for _, w in ipairs(colWidths) do
        totalWidth = totalWidth + w
    end

    local children = {}

    -- Header row
    local headerCells = {}
    for i, h in ipairs(headers) do
        local text = tostring(h)
        local cw = colWidths[i] or #text
        local padded = text .. string.rep(" ", math.max(0, cw - #text))
        headerCells[#headerCells + 1] = lumina.createElement("text", {
            foreground = "#89B4FA",
            bold = true,
        }, padded)
    end
    children[#children + 1] = lumina.createElement("hbox", {},
        table.unpack(headerCells))

    -- Separator
    children[#children + 1] = lumina.createElement("text", {
        foreground = "#585B70",
    }, string.rep("─", totalWidth))

    -- Data rows
    for ri, row in ipairs(rows) do
        local rowCells = {}
        local isSelected = (ri == selectedRow)
        for i, cell in ipairs(row) do
            local text = tostring(cell)
            local cw = colWidths[i] or #text
            local padded = text .. string.rep(" ", math.max(0, cw - #text))
            local fg = isSelected and "#1E1E2E" or "#CDD6F4"
            local cellProps = {foreground = fg}
            if isSelected then
                cellProps.background = "#89B4FA"
            end
            rowCells[#rowCells + 1] = lumina.createElement("text",
                cellProps, padded)
        end
        children[#children + 1] = lumina.createElement("hbox", {},
            table.unpack(rowCells))
    end

    local outerProps = {}
    if props.style then
        outerProps.style = props.style
    end
    return lumina.createElement("vbox", outerProps, table.unpack(children))
end

-- ═══════════════════════════════════════════════════════════════════
-- Tabs
-- ═══════════════════════════════════════════════════════════════════
-- props:
--   tabs         : array of {label = "...", content = VNode}
--   activeTab    : 1-based index of active tab (default 1)
--   onTabChange  : optional callback(newIndex)  (for future click support)
--   style        : optional style table for the outer vbox
--   separatorLen : separator line width (default 40)
--
-- Returns: vbox with tab bar + separator + active tab content

function M.Tabs(props)
    local tabs = props.tabs or {}
    local activeTab = props.activeTab or 1
    local separatorLen = props.separatorLen or 40

    -- Tab buttons
    local tabButtons = {}
    for i, tab in ipairs(tabs) do
        local isActive = (i == activeTab)
        tabButtons[#tabButtons + 1] = lumina.createElement("text", {
            foreground = isActive and "#1E1E2E" or "#CDD6F4",
            background = isActive and "#89B4FA" or "#313244",
            bold = isActive,
        }, " " .. tab.label .. " ")
    end

    local children = {}
    children[#children + 1] = lumina.createElement("hbox", {
        style = {gap = 1},
    }, table.unpack(tabButtons))
    children[#children + 1] = lumina.createElement("text", {
        foreground = "#585B70",
    }, string.rep("─", separatorLen))

    -- Active tab content
    if tabs[activeTab] and tabs[activeTab].content then
        children[#children + 1] = tabs[activeTab].content
    end

    local outerProps = {}
    if props.style then
        outerProps.style = props.style
    end
    return lumina.createElement("vbox", outerProps, table.unpack(children))
end

-- ═══════════════════════════════════════════════════════════════════
-- Modal
-- ═══════════════════════════════════════════════════════════════════
-- props:
--   visible  : boolean — when false, returns an invisible empty node
--   title    : dialog title string (default "Dialog")
--   width    : inner content width for separator (default 40)
--   children : VNode content to display inside the modal
--
-- Returns: box with rounded border, title, separator, content, hint

function M.Modal(props)
    if not props.visible then
        return lumina.createElement("text", {}, "")
    end

    local w = props.width or 40
    local title = props.title or "Dialog"

    return lumina.createElement("box", {
        style = {
            border = "rounded",
            background = "#1E1E2E",
            padding = 1,
        },
    },
        lumina.createElement("text", {
            foreground = "#89B4FA",
            bold = true,
        }, title),
        lumina.createElement("text", {
            foreground = "#585B70",
        }, string.rep("─", math.max(0, w - 4))),
        props.children or lumina.createElement("text", {}, ""),
        lumina.createElement("text", {}, ""),
        lumina.createElement("text", {
            foreground = "#6C7086",
        }, "[Esc] Close")
    )
end

-- ═══════════════════════════════════════════════════════════════════
-- Select
-- ═══════════════════════════════════════════════════════════════════
-- props:
--   options  : array of option strings
--   selected : 1-based index of selected option (default 1)
--   onSelect : optional callback(newIndex)  (for future click support)
--   label    : optional label string shown above options
--   style    : optional style table for the outer vbox
--
-- Returns: vbox with optional label + option list with highlighted selection

function M.Select(props)
    local options = props.options or {}
    local selected = props.selected or 1
    local label = props.label or ""

    local children = {}
    if label ~= "" then
        children[#children + 1] = lumina.createElement("text", {
            foreground = "#89B4FA",
            bold = true,
        }, label)
    end

    for i, opt in ipairs(options) do
        local isSelected = (i == selected)
        local prefix = isSelected and "▸ " or "  "
        local fg = isSelected and "#A6E3A1" or "#CDD6F4"
        local cellProps = {foreground = fg, bold = isSelected}
        if isSelected then
            cellProps.background = "#313244"
        end
        children[#children + 1] = lumina.createElement("text",
            cellProps, prefix .. opt)
    end

    local outerProps = {}
    if props.style then
        outerProps.style = props.style
    end
    return lumina.createElement("vbox", outerProps, table.unpack(children))
end

return M
