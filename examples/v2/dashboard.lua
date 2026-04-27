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
    {icon = "●", text = "Server started",           time = "08:00"},
    {icon = "●", text = "DB connection pool ready",  time = "08:01"},
    {icon = "●", text = "Cache warmed up",           time = "08:02"},
    {icon = "●", text = "Health check: all green",   time = "08:05"},
    {icon = "●", text = "User login: admin",         time = "08:12"},
    {icon = "●", text = "Config reload triggered",   time = "08:15"},
    {icon = "●", text = "Cron: log rotation",        time = "08:30"},
    {icon = "●", text = "SSL cert check passed",     time = "08:45"},
    {icon = "●", text = "Deploy v2.1.0 started",     time = "09:00"},
    {icon = "●", text = "Build artifact uploaded",   time = "09:05"},
    {icon = "●", text = "Migration 042 applied",     time = "09:08"},
    {icon = "●", text = "Deploy v2.1.0 complete",    time = "09:12"},
    {icon = "●", text = "CDN cache purged",          time = "09:13"},
    {icon = "●", text = "Webhook: deploy notify",    time = "09:14"},
    {icon = "●", text = "User login: alice",         time = "09:20"},
    {icon = "●", text = "API rate limit adjusted",   time = "09:30"},
    {icon = "●", text = "New user registered: bob",  time = "09:35"},
    {icon = "●", text = "Cron: metrics aggregation", time = "09:45"},
    {icon = "●", text = "Alert: CPU spike 92%",      time = "09:50"},
    {icon = "●", text = "Auto-scale: +2 instances",  time = "09:51"},
    {icon = "●", text = "CPU normalized to 45%",     time = "09:55"},
    {icon = "●", text = "DB backup started",         time = "10:00"},
    {icon = "●", text = "DB backup complete (2.3G)", time = "10:12"},
    {icon = "●", text = "Backup uploaded to S3",     time = "10:15"},
    {icon = "●", text = "User login: charlie",       time = "10:20"},
    {icon = "●", text = "Password reset: dave",      time = "10:25"},
    {icon = "●", text = "Cron: cleanup temp files",  time = "10:30"},
    {icon = "●", text = "Index rebuild started",     time = "10:35"},
    {icon = "●", text = "Index rebuild finished",    time = "10:42"},
    {icon = "●", text = "Search latency improved",   time = "10:43"},
    {icon = "●", text = "Webhook: Stripe payment",   time = "10:50"},
    {icon = "●", text = "Invoice #1042 generated",   time = "10:51"},
    {icon = "●", text = "Email sent: receipt",        time = "10:52"},
    {icon = "●", text = "User login: eve",            time = "11:00"},
    {icon = "●", text = "API key rotated: svc-bot",  time = "11:05"},
    {icon = "●", text = "Rate limit hit: 203.0.x",   time = "11:10"},
    {icon = "●", text = "Geo-block: 5 IPs added",    time = "11:12"},
    {icon = "●", text = "Health check: all green",   time = "11:15"},
    {icon = "●", text = "Cron: session cleanup",     time = "11:30"},
    {icon = "●", text = "Session expired: guest",    time = "11:30"},
    {icon = "●", text = "Auto-scale: -1 instance",   time = "11:35"},
    {icon = "●", text = "Metrics export complete",   time = "11:45"},
    {icon = "●", text = "Daily report generated",    time = "11:50"},
    {icon = "●", text = "Slack notify: daily stats", time = "11:51"},
    {icon = "●", text = "User logout: admin",        time = "11:55"},
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

        -- Scroll step (matches framework's scrollStep = 3)
        local scrollStep = 3

        -- Max scroll: total log entries minus visible rows in the scroll area.
        -- Layout: 24 total - 1 header - 1 footer - 2 log title/separator = 20 visible rows.
        -- The scroll container also loses 1 column for the scrollbar.
        local visibleLogRows = 20
        local maxScroll = #activityLog - visibleLogRows
        if maxScroll < 0 then maxScroll = 0 end

        -- Clamp helper: ensures scrollY stays in [0, maxScroll]
        local function clampedScroll(newY)
            if newY < 0 then newY = 0 end
            if newY > maxScroll then newY = maxScroll end
            return newY
        end

        -- Keyboard handler
        local function handleKey(e)
            if e.key == "q" then
                lumina.quit()
            elseif e.key == "j" or e.key == "ArrowDown" then
                setScrollY(clampedScroll(scrollY + scrollStep))
            elseif e.key == "k" or e.key == "ArrowUp" then
                setScrollY(clampedScroll(scrollY - scrollStep))
            end
        end

        -- Mouse wheel scroll handler (receives scroll events from framework)
        local function handleScroll(e)
            if e.key == "down" then
                setScrollY(clampedScroll(scrollY + scrollStep))
            elseif e.key == "up" then
                setScrollY(clampedScroll(scrollY - scrollStep))
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

        -- Build the scroll container via raw table (children are dynamic).
        -- The onScroll handler is attached so mouse wheel events update scrollY.
        local logScrollContainer = {
            type = "vbox",
            id = "log-scroll",
            scrollY = scrollY,
            onScroll = handleScroll,
            style = {overflow = "scroll", flex = 1, background = theme.bg},
            children = logChildren,
        }

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
            logScrollContainer
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
