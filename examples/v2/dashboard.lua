-- Lumina v2 Example: Dual-Scroll Dashboard
-- Demonstrates: two independent scrollable areas, clickable list,
--               mouse wheel scroll dispatch by hit-test, Catppuccin Mocha theme
--
-- Usage: lumina-v2 examples/v2/dashboard.lua
-- Quit:  q or Ctrl+Q
--
-- Left panel:  Clickable list of 30 items (mouse wheel to scroll)
-- Right panel: Detail view for selected item (mouse wheel to scroll)

-- Theme (Catppuccin Mocha)
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

-- ═══════════════════════════════════════════
-- Data: 30 items, each with 50 lines of content
-- ═══════════════════════════════════════════
local names = {
    "Server Config", "Database Pool", "Cache Layer", "Auth Service",
    "API Gateway", "Load Balancer", "Message Queue", "Task Scheduler",
    "Log Aggregator", "Metrics Engine", "Alert Manager", "DNS Resolver",
    "SSL Manager", "Rate Limiter", "Session Store", "File Storage",
    "CDN Config", "Webhook Router", "Email Service", "Payment Gateway",
    "Search Index", "Geo Lookup", "Feature Flags", "A/B Testing",
    "User Profiles", "Audit Trail", "Backup Agent", "Health Monitor",
    "Deploy Pipeline", "Secret Vault",
}

local items = {}
for i = 1, 30 do
    local lines = {}
    for j = 1, 50 do
        lines[j] = string.format(
            "Line %d: Configuration for %s — parameter %d set to value %d. "
            .. "Status: active. Last modified: 2024-01-%02d %02d:%02d.",
            j, names[i] or "Entry", j, j * 17 + i, (j % 28) + 1, (j + i) % 24, (j * 7) % 60
        )
    end
    items[i] = {
        title = string.format("%-2d  %s", i, names[i] or "Entry"),
        content = lines,
    }
end

lumina.createComponent({
    id = "dashboard",
    name = "Dashboard",

    render = function(props)
        local createElement = lumina.createElement

        -- ── State ──
        local selectedIdx, setSelectedIdx = lumina.useState("sel", 1)
        local leftScrollY, setLeftScrollY = lumina.useState("leftScroll", 0)
        local rightScrollY, setRightScrollY = lumina.useState("rightScroll", 0)

        local scrollStep = 3

        -- ── Keyboard handler ──
        local function handleKey(e)
            if e.key == "q" then
                lumina.quit()
            elseif e.key == "j" or e.key == "ArrowDown" then
                setRightScrollY(math.max(0, rightScrollY + scrollStep))
            elseif e.key == "k" or e.key == "ArrowUp" then
                setRightScrollY(math.max(0, rightScrollY - scrollStep))
            elseif e.key == "J" then
                setLeftScrollY(math.max(0, leftScrollY + scrollStep))
            elseif e.key == "K" then
                setLeftScrollY(math.max(0, leftScrollY - scrollStep))
            end
        end

        -- ═══════════════════════════════════════════
        -- Header
        -- ═══════════════════════════════════════════
        local header = createElement("hbox", {
            style = {background = theme.headerBg, height = 1},
        },
            createElement("text", {
                foreground = theme.accent, bold = true,
                style = {flex = 1},
            }, " ◆ Dashboard — Dual Scroll Demo"),
            createElement("text", {
                foreground = theme.muted,
            }, "q: quit ")
        )

        -- ═══════════════════════════════════════════
        -- Left Panel: Clickable list with scroll
        -- ═══════════════════════════════════════════
        local leftItems = {}
        for i, item in ipairs(items) do
            local isSelected = (i == selectedIdx)
            local bg = isSelected and theme.accent or theme.bg
            local fg = isSelected and theme.bg or theme.fg
            leftItems[#leftItems + 1] = createElement("text", {
                style = {height = 1, background = bg},
                foreground = fg,
                onClick = function()
                    setSelectedIdx(i)
                    setRightScrollY(0) -- reset right scroll when changing item
                end,
            }, " " .. item.title)
        end

        local leftPanel = {
            type = "vbox",
            id = "left-list",
            scrollY = leftScrollY,
            onScroll = function(e)
                if e.key == "down" then
                    setLeftScrollY(math.max(0, leftScrollY + scrollStep))
                elseif e.key == "up" then
                    setLeftScrollY(math.max(0, leftScrollY - scrollStep))
                end
            end,
            style = {
                width = 30,
                overflow = "scroll",
                background = theme.surface,
                border = "single",
            },
            children = leftItems,
        }

        -- ═══════════════════════════════════════════
        -- Right Panel: Detail content with scroll
        -- ═══════════════════════════════════════════
        local selected = items[selectedIdx]
        local contentLines = {}

        -- Title
        contentLines[#contentLines + 1] = createElement("text", {
            foreground = theme.accent, bold = true,
            style = {height = 1},
        }, " " .. (names[selectedIdx] or "Entry"))

        -- Separator
        contentLines[#contentLines + 1] = createElement("text", {
            foreground = theme.border,
            style = {height = 1},
        }, " " .. string.rep("─", 60))

        -- Content lines
        for _, line in ipairs(selected.content) do
            contentLines[#contentLines + 1] = createElement("text", {
                foreground = theme.fg,
                style = {height = 1},
            }, " " .. line)
        end

        local rightPanel = {
            type = "vbox",
            id = "right-detail",
            scrollY = rightScrollY,
            onScroll = function(e)
                if e.key == "down" then
                    setRightScrollY(math.max(0, rightScrollY + scrollStep))
                elseif e.key == "up" then
                    setRightScrollY(math.max(0, rightScrollY - scrollStep))
                end
            end,
            style = {
                flex = 1,
                overflow = "scroll",
                background = theme.bg,
                border = "single",
            },
            children = contentLines,
        }

        -- ═══════════════════════════════════════════
        -- Main content area (left + right side by side)
        -- ═══════════════════════════════════════════
        local mainContent = createElement("hbox", {
            style = {flex = 1, background = theme.bg},
        }, leftPanel, rightPanel)

        -- ═══════════════════════════════════════════
        -- Footer
        -- ═══════════════════════════════════════════
        local footer = createElement("text", {
            foreground = theme.muted,
            style = {background = theme.headerBg, height = 1},
        }, " [Click] Select  [Scroll] Wheel  [j/k] Right  [J/K] Left  [q] Quit")

        -- ═══════════════════════════════════════════
        -- Root layout
        -- ═══════════════════════════════════════════
        return createElement("vbox", {
            style = {background = theme.bg},
            onKeyDown = handleKey,
        },
            header,
            mainContent,
            footer
        )
    end,
})
