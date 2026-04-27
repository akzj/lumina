-- Lumina v2 Example: System Dashboard
-- Demonstrates: multiple panels, progress bars, scrollable lists,
--               keyboard navigation, Catppuccin Mocha theme
--
-- Usage: lumina-v2 examples/v2/dashboard.lua
-- Quit:  q or Ctrl+Q
--
-- Keyboard:
--   j/k          - Scroll activity log
--   q / Ctrl+Q   - Quit

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

-- Helper: build a text progress bar
local function progressBar(pct, width)
    local filled = math.floor(pct / 100 * width)
    local empty = width - filled
    return string.rep("█", filled) .. string.rep("░", empty)
end

-- Helper: format a percentage with color
local function pctColor(pct)
    if pct >= 80 then return theme.red
    elseif pct >= 60 then return theme.yellow
    else return theme.green
    end
end

-- Static data (would be live with timers)
local cpuPct  = 67
local ramPct  = 52
local diskPct = 89
local netPct  = 24

local activityLog = {
    {icon = "●", text = "Server started",          time = "09:00"},
    {icon = "●", text = "User login: admin",        time = "09:05"},
    {icon = "●", text = "Deploy v2.1.0 complete",   time = "09:12"},
    {icon = "●", text = "Cache cleared",            time = "09:15"},
    {icon = "●", text = "DB backup complete",       time = "09:30"},
    {icon = "●", text = "Config updated",           time = "09:45"},
    {icon = "●", text = "SSL cert renewed",         time = "10:00"},
    {icon = "●", text = "Health check passed",      time = "10:15"},
    {icon = "●", text = "New user: alice",           time = "10:22"},
    {icon = "●", text = "Cron job: cleanup",        time = "10:30"},
    {icon = "●", text = "API rate limit adjusted",  time = "10:45"},
    {icon = "●", text = "Metrics export complete",  time = "11:00"},
    {icon = "●", text = "Webhook delivered",        time = "11:10"},
    {icon = "●", text = "Index rebuild finished",   time = "11:20"},
    {icon = "●", text = "User login: bob",           time = "11:25"},
    {icon = "●", text = "Session expired: guest",   time = "11:30"},
}

local stats = {
    {label = "Uptime",   value = "42 days"},
    {label = "Users",    value = "128 online"},
    {label = "Requests", value = "1.2M today"},
    {label = "Errors",   value = "0.02%"},
    {label = "Latency",  value = "12ms avg"},
    {label = "Threads",  value = "64 active"},
}

lumina.createComponent({
    id = "dashboard",
    name = "Dashboard",
    x = 0, y = 0,
    w = 80, h = 24,
    zIndex = 0,

    render = function(state, props)
        local scrollY, setScrollY = lumina.useState("scrollY", 0)

        -- TODO: When lumina.setInterval is available, add live updates:
        -- lumina.setInterval(function()
        --     setCpuPct(math.random(30, 95))
        --     setRamPct(math.random(40, 80))
        --     setNetPct(math.random(10, 60))
        -- end, 2000)

        -- Keyboard handler
        local function handleKey(e)
            if e.key == "q" then
                lumina.quit()
            elseif e.key == "j" or e.key == "ArrowDown" then
                setScrollY(scrollY + 1)
            elseif e.key == "k" or e.key == "ArrowUp" then
                if scrollY > 0 then
                    setScrollY(scrollY - 1)
                end
            end
        end

        -- ═══════════════════════════════════════════
        -- Header
        -- ═══════════════════════════════════════════
        local header = lumina.createElement("hbox", {
            id = "header",
            style = {background = theme.headerBg, height = 1},
        },
            lumina.createElement("text", {
                foreground = theme.accent,
                bold = true,
                style = {flex = 1},
            }, " Dashboard"),
            lumina.createElement("text", {
                foreground = theme.muted,
            }, "q: quit ")
        )

        -- ═══════════════════════════════════════════
        -- System Resources Panel
        -- ═══════════════════════════════════════════
        local barWidth = 16
        local resources = {
            {name = "CPU ", pct = cpuPct},
            {name = "RAM ", pct = ramPct},
            {name = "Disk", pct = diskPct},
            {name = "Net ", pct = netPct},
        }

        local resourceChildren = {
            lumina.createElement("text", {
                foreground = theme.accent,
                bold = true,
            }, " System Resources"),
            lumina.createElement("text", {
                foreground = theme.border,
            }, " " .. string.rep("─", 28)),
        }
        for _, r in ipairs(resources) do
            local bar = progressBar(r.pct, barWidth)
            local line = string.format(" %s [%s] %3d%%", r.name, bar, r.pct)
            resourceChildren[#resourceChildren + 1] = lumina.createElement("text", {
                foreground = pctColor(r.pct),
            }, line)
        end
        -- Add empty line after resources
        resourceChildren[#resourceChildren + 1] = lumina.createElement("text", {
            foreground = theme.bg,
        }, "")

        local resourcePanel = {
            type = "vbox",
            id = "resource-panel",
            style = {background = theme.bg, flex = 1},
            children = resourceChildren,
        }

        -- ═══════════════════════════════════════════
        -- Quick Stats Panel
        -- ═══════════════════════════════════════════
        local statsChildren = {
            lumina.createElement("text", {
                foreground = theme.accent,
                bold = true,
            }, " Quick Stats"),
            lumina.createElement("text", {
                foreground = theme.border,
            }, " " .. string.rep("─", 28)),
        }
        for _, s in ipairs(stats) do
            local line = string.format(" %-10s %s", s.label, s.value)
            statsChildren[#statsChildren + 1] = lumina.createElement("text", {
                foreground = theme.fg,
            }, line)
        end

        local statsPanel = {
            type = "vbox",
            id = "stats-panel",
            style = {background = theme.bg, flex = 1},
            children = statsChildren,
        }

        -- ═══════════════════════════════════════════
        -- Left Column (Resources + Stats stacked)
        -- ═══════════════════════════════════════════
        local leftColumn = lumina.createElement("vbox", {
            id = "left-col",
            style = {flex = 1, background = theme.bg},
        },
            resourcePanel,
            statsPanel
        )

        -- ═══════════════════════════════════════════
        -- Activity Log Panel (scrollable)
        -- ═══════════════════════════════════════════
        local logChildren = {}
        for _, entry in ipairs(activityLog) do
            local line = string.format(" %s %s  %s", entry.icon, entry.text, entry.time)
            logChildren[#logChildren + 1] = lumina.createElement("text", {
                foreground = theme.fg,
                style = {height = 1},
            }, line)
        end

        local activityPanel = lumina.createElement("vbox", {
            id = "activity-panel",
            style = {flex = 1, background = theme.bg},
        },
            lumina.createElement("text", {
                foreground = theme.accent,
                bold = true,
            }, " Activity Log"),
            lumina.createElement("text", {
                foreground = theme.border,
            }, " " .. string.rep("─", 34)),
            {
                type = "vbox",
                id = "log-scroll",
                scrollY = scrollY,
                style = {overflow = "scroll", flex = 1, background = theme.bg},
                children = logChildren,
            }
        )

        -- ═══════════════════════════════════════════
        -- Main content area (left + right side by side)
        -- ═══════════════════════════════════════════
        local mainContent = lumina.createElement("hbox", {
            id = "main-content",
            style = {flex = 1, background = theme.bg},
        },
            leftColumn,
            activityPanel
        )

        -- ═══════════════════════════════════════════
        -- Footer
        -- ═══════════════════════════════════════════
        local footer = lumina.createElement("text", {
            id = "footer",
            foreground = theme.muted,
            style = {background = theme.headerBg, height = 1},
        }, " [j/k] Scroll Log  [q] Quit")

        -- ═══════════════════════════════════════════
        -- Root layout
        -- ═══════════════════════════════════════════
        return lumina.createElement("vbox", {
            id = "dashboard-root",
            style = {background = theme.bg},
            onKeyDown = handleKey,
            focusable = true,
        },
            header,
            mainContent,
            footer
        )
    end,
})
