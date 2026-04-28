-- Lumina v2 Example: Component Library Showcase
-- Demonstrates: ProgressBar, Table, Tabs, Modal, Select
--
-- Usage: lumina-v2 examples/v2/components_showcase.lua
-- Quit:  q or Ctrl+Q
--
-- Keyboard:
--   1/2/3       - Switch tabs
--   j/k         - Navigate select list / table rows
--   m           - Toggle modal
--   Esc         - Close modal
--   q / Ctrl+Q  - Quit
--
-- NOTE: Component functions are inlined because v2 does not yet support
-- require() for local Lua files. See examples/v2/lib/components.lua for
-- the canonical library source.

-- ═══════════════════════════════════════════════════════════════════
-- Theme (Catppuccin Mocha)
-- ═══════════════════════════════════════════════════════════════════
local theme = {
    bg       = "#1E1E2E",
    surface  = "#313244",
    overlay  = "#45475A",
    fg       = "#CDD6F4",
    accent   = "#89B4FA",
    green    = "#A6E3A1",
    yellow   = "#F9E2AF",
    red      = "#F38BA8",
    peach    = "#FAB387",
    muted    = "#6C7086",
    border   = "#585B70",
    headerBg = "#181825",
}

-- ═══════════════════════════════════════════════════════════════════
-- Component Library (inlined)
-- ═══════════════════════════════════════════════════════════════════

local function ProgressBar(props)
    local value = math.max(0, math.min(100, props.value or 0))
    local width = props.width or 20
    local color = props.color or theme.green
    local label = props.label or ""

    local filled = math.floor(value / 100 * width)
    local empty = width - filled
    local bar = string.rep("█", filled) .. string.rep("░", empty)
    local pct = string.format("%3d%%", value)

    local pctColor = theme.green
    if value > 80 then pctColor = theme.red
    elseif value > 60 then pctColor = theme.yellow
    end

    local children = {}
    if label ~= "" then
        children[#children + 1] = lumina.createElement("text", {
            foreground = theme.fg,
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

local function DataTable(props)
    local headers = props.headers or {}
    local rows = props.rows or {}
    local selectedRow = props.selectedRow or -1
    local colWidths = props.colWidths

    if not colWidths then
        colWidths = {}
        for i, h in ipairs(headers) do
            colWidths[i] = #tostring(h) + 2
        end
        for _, row in ipairs(rows) do
            for i, cell in ipairs(row) do
                local w = #tostring(cell) + 2
                if w > (colWidths[i] or 0) then colWidths[i] = w end
            end
        end
    end

    local totalWidth = 0
    for _, w in ipairs(colWidths) do totalWidth = totalWidth + w end

    local children = {}

    local headerCells = {}
    for i, h in ipairs(headers) do
        local text = tostring(h)
        local cw = colWidths[i] or #text
        local padded = text .. string.rep(" ", math.max(0, cw - #text))
        headerCells[#headerCells + 1] = lumina.createElement("text", {
            foreground = theme.accent, bold = true,
        }, padded)
    end
    children[#children + 1] = lumina.createElement("hbox", {},
        table.unpack(headerCells))

    children[#children + 1] = lumina.createElement("text", {
        foreground = theme.border,
    }, string.rep("─", totalWidth))

    for ri, row in ipairs(rows) do
        local rowCells = {}
        local isSelected = (ri == selectedRow)
        for i, cell in ipairs(row) do
            local text = tostring(cell)
            local cw = colWidths[i] or #text
            local padded = text .. string.rep(" ", math.max(0, cw - #text))
            local fg = isSelected and theme.bg or theme.fg
            local cellProps = {foreground = fg}
            if isSelected then cellProps.background = theme.accent end
            rowCells[#rowCells + 1] = lumina.createElement("text",
                cellProps, padded)
        end
        children[#children + 1] = lumina.createElement("hbox", {},
            table.unpack(rowCells))
    end

    local outerProps = {}
    if props.style then outerProps.style = props.style end
    return lumina.createElement("vbox", outerProps, table.unpack(children))
end

local function Tabs(props)
    local tabs = props.tabs or {}
    local activeTab = props.activeTab or 1
    local separatorLen = props.separatorLen or 40

    local tabButtons = {}
    for i, tab in ipairs(tabs) do
        local isActive = (i == activeTab)
        tabButtons[#tabButtons + 1] = lumina.createElement("text", {
            foreground = isActive and theme.bg or theme.fg,
            background = isActive and theme.accent or theme.surface,
            bold = isActive,
        }, " " .. tab.label .. " ")
    end

    local children = {}
    children[#children + 1] = lumina.createElement("hbox", {
        style = {gap = 1},
    }, table.unpack(tabButtons))
    children[#children + 1] = lumina.createElement("text", {
        foreground = theme.border,
    }, string.rep("─", separatorLen))

    if tabs[activeTab] and tabs[activeTab].content then
        children[#children + 1] = tabs[activeTab].content
    end

    local outerProps = {}
    if props.style then outerProps.style = props.style end
    return lumina.createElement("vbox", outerProps, table.unpack(children))
end

local function Modal(props)
    if not props.visible then
        return lumina.createElement("text", {}, "")
    end

    local w = props.width or 40
    local title = props.title or "Dialog"

    return lumina.createElement("box", {
        style = {
            position = "absolute",
            top = 5,
            left = 20,
            width = w,
            height = 12,
            border = "rounded",
            background = theme.bg,
            padding = 1,
            zIndex = 10,
        },
    },
        lumina.createElement("text", {
            foreground = theme.accent, bold = true,
        }, title),
        lumina.createElement("text", {
            foreground = theme.border,
        }, string.rep("─", math.max(0, w - 4))),
        props.children or lumina.createElement("text", {}, ""),
        lumina.createElement("text", {}, ""),
        lumina.createElement("text", {
            foreground = theme.muted,
        }, "[Esc] Close")
    )
end

local function Select(props)
    local options = props.options or {}
    local selected = props.selected or 1
    local label = props.label or ""

    local children = {}
    if label ~= "" then
        children[#children + 1] = lumina.createElement("text", {
            foreground = theme.accent, bold = true,
        }, label)
    end

    for i, opt in ipairs(options) do
        local isSelected = (i == selected)
        local prefix = isSelected and "▸ " or "  "
        local fg = isSelected and theme.green or theme.fg
        local cellProps = {foreground = fg, bold = isSelected}
        if isSelected then cellProps.background = theme.surface end
        children[#children + 1] = lumina.createElement("text",
            cellProps, prefix .. opt)
    end

    local outerProps = {}
    if props.style then outerProps.style = props.style end
    return lumina.createElement("vbox", outerProps, table.unpack(children))
end

-- ═══════════════════════════════════════════════════════════════════
-- Sample Data
-- ═══════════════════════════════════════════════════════════════════

local tableData = {
    headers = {"Name", "Role", "Status"},
    rows = {
        {"Alice",   "Admin",     "Active"},
        {"Bob",     "Developer", "Active"},
        {"Charlie", "Designer",  "Away"},
        {"Diana",   "DevOps",    "Active"},
        {"Eve",     "QA",        "Offline"},
    },
}

local selectOptions = {"Dark Mode", "Light Mode", "System Default", "High Contrast", "Solarized"}

-- ═══════════════════════════════════════════════════════════════════
-- Main Component
-- ═══════════════════════════════════════════════════════════════════

lumina.createComponent({
    id = "showcase",
    name = "ComponentShowcase",
    x = 0, y = 0,
    w = 80, h = 24,
    zIndex = 0,

    render = function(state, props)
        local activeTab, setActiveTab = lumina.useState("activeTab", 1)
        local selectedOption, setSelectedOption = lumina.useState("selectedOption", 1)
        local selectedRow, setSelectedRow = lumina.useState("selectedRow", 1)
        local showModal, setShowModal = lumina.useState("showModal", false)

        local function handleKey(e)
            -- Modal takes priority
            if showModal then
                if e.key == "Escape" then
                    setShowModal(false)
                end
                return
            end

            if e.key == "q" then
                lumina.quit()
            elseif e.key == "1" then
                setActiveTab(1)
            elseif e.key == "2" then
                setActiveTab(2)
            elseif e.key == "3" then
                setActiveTab(3)
            elseif e.key == "m" then
                setShowModal(true)
            elseif e.key == "j" or e.key == "ArrowDown" then
                if activeTab == 2 then
                    local maxRow = #tableData.rows
                    if selectedRow < maxRow then
                        setSelectedRow(selectedRow + 1)
                    end
                elseif activeTab == 3 then
                    local maxOpt = #selectOptions
                    if selectedOption < maxOpt then
                        setSelectedOption(selectedOption + 1)
                    end
                end
            elseif e.key == "k" or e.key == "ArrowUp" then
                if activeTab == 2 then
                    if selectedRow > 1 then
                        setSelectedRow(selectedRow - 1)
                    end
                elseif activeTab == 3 then
                    if selectedOption > 1 then
                        setSelectedOption(selectedOption - 1)
                    end
                end
            end
        end

        -- ── Tab 1: Progress Bars ──
        local progressContent = lumina.createElement("vbox", {
            style = {gap = 1},
        },
            ProgressBar({label = "CPU  ", value = 67, width = 24}),
            ProgressBar({label = "RAM  ", value = 45, width = 24}),
            ProgressBar({label = "Disk ", value = 89, width = 24}),
            ProgressBar({label = "Net  ", value = 23, width = 24})
        )

        -- ── Tab 2: Data Table ──
        local tableContent = DataTable({
            headers = tableData.headers,
            rows = tableData.rows,
            selectedRow = selectedRow,
        })

        -- ── Tab 3: Select List ──
        local selectContent = Select({
            label = "Theme:",
            options = selectOptions,
            selected = selectedOption,
        })

        -- ── Header ──
        local header = lumina.createElement("hbox", {
            id = "header",
            style = {background = theme.headerBg, height = 1},
        },
            lumina.createElement("text", {
                foreground = theme.accent, bold = true,
                style = {flex = 1},
            }, " Component Library Showcase"),
            lumina.createElement("text", {
                foreground = theme.muted,
            }, "q: quit  m: modal ")
        )

        -- ── Footer ──
        local footer = lumina.createElement("text", {
            id = "footer",
            foreground = theme.muted,
            style = {background = theme.headerBg, height = 1},
        }, " [1/2/3] Tabs  [j/k] Navigate  [m] Modal  [q] Quit")

        -- ── Main content with tabs ──
        local mainContent = Tabs({
            activeTab = activeTab,
            separatorLen = 76,
            tabs = {
                {label = "Progress", content = progressContent},
                {label = "Table",    content = tableContent},
                {label = "Select",   content = selectContent},
            },
            style = {flex = 1, background = theme.bg, padding = 1},
        })

        -- ── Modal overlay ──
        local modalContent = Modal({
            visible = showModal,
            title = "Example Modal",
            width = 40,
            children = lumina.createElement("vbox", {},
                lumina.createElement("text", {foreground = theme.fg},
                    "This is a modal dialog."),
                lumina.createElement("text", {foreground = theme.fg},
                    "Press Esc to close.")
            ),
        })

        -- ── Root layout ──
        return lumina.createElement("vbox", {
            id = "showcase-root",
            style = {background = theme.bg},
            onKeyDown = handleKey,
            focusable = true,
        },
            header,
            mainContent,
            modalContent,
            footer
        )
    end,
})
