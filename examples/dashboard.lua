-- ============================================================================
-- Lumina Example: System Dashboard
-- ============================================================================
-- Showcases: Tabs, Table, Progress bars, Spinner, StatusBar, conditional
--            tab rendering, data formatting, nested layouts.
--
-- Run: lumina examples/dashboard.lua
-- ============================================================================

local lumina = require("lumina")

-- ── Theme ──────────────────────────────────────────────────────────────────
local theme = {
    bg       = "#1E1E2E",
    fg       = "#CDD6F4",
    accent   = "#89B4FA",
    success  = "#A6E3A1",
    warning  = "#F9E2AF",
    error    = "#F38BA8",
    muted    = "#6C7086",
    surface  = "#313244",
    headerBg = "#181825",
}

-- ── Helper: progress bar string ────────────────────────────────────────────
local function progressBar(value, width, label)
    width = width or 30
    local clamped = math.max(0, math.min(1, value))
    local filled = math.floor(clamped * width + 0.5)
    local empty = width - filled
    local bar = string.rep("█", filled) .. string.rep("░", empty)
    local pct = math.floor(clamped * 100 + 0.5)

    -- Color based on usage level
    local color = theme.success
    if pct > 80 then color = theme.error
    elseif pct > 60 then color = theme.warning end

    return {
        type = "hbox",
        style = { height = 1 },
        children = {
            { type = "text",
              content = string.format(" %-12s", label or ""),
              style = { foreground = theme.fg, width = 14 } },
            { type = "text",
              content = "[" .. bar .. "]",
              style = { foreground = color } },
            { type = "text",
              content = string.format(" %3d%%", pct),
              style = { foreground = color, bold = pct > 80 } },
        },
    }
end

-- ── Helper: fit text to column width ───────────────────────────────────────
local function fit(str, width)
    str = tostring(str or "")
    if #str > width then return str:sub(1, width - 1) .. "…" end
    return str .. string.rep(" ", width - #str)
end

-- ── Tab: Processes ─────────────────────────────────────────────────────────
local function renderProcesses()
    local processes = {
        { pid = "1",    name = "systemd",       cpu = "0.1",  mem = "12.3M" },
        { pid = "142",  name = "go-lua",        cpu = "23.4", mem = "89.1M" },
        { pid = "256",  name = "lumina-server",  cpu = "15.7", mem = "45.2M" },
        { pid = "389",  name = "postgres",       cpu = "8.2",  mem = "256.8M" },
        { pid = "412",  name = "nginx",          cpu = "2.1",  mem = "18.4M" },
        { pid = "567",  name = "redis-server",   cpu = "1.3",  mem = "32.1M" },
        { pid = "723",  name = "node",           cpu = "12.8", mem = "128.5M" },
        { pid = "891",  name = "docker",         cpu = "5.6",  mem = "67.3M" },
    }

    local children = {}

    -- Header
    children[#children + 1] = {
        type = "text",
        content = fit("PID", 8) .. "│" .. fit("Name", 20) .. "│" .. fit("CPU%", 8) .. "│" .. fit("Memory", 10),
        style = { foreground = theme.accent, bold = true },
    }
    -- Separator
    children[#children + 1] = {
        type = "text",
        content = string.rep("─", 8) .. "┼" .. string.rep("─", 20) .. "┼" .. string.rep("─", 8) .. "┼" .. string.rep("─", 10),
        style = { foreground = theme.muted },
    }
    -- Rows
    for _, proc in ipairs(processes) do
        local cpuVal = tonumber(proc.cpu) or 0
        local cpuColor = theme.fg
        if cpuVal > 20 then cpuColor = theme.error
        elseif cpuVal > 10 then cpuColor = theme.warning end

        children[#children + 1] = {
            type = "hbox",
            style = { height = 1 },
            children = {
                { type = "text", content = fit(proc.pid, 8) .. "│",
                  style = { foreground = theme.muted } },
                { type = "text", content = fit(proc.name, 20) .. "│",
                  style = { foreground = theme.fg } },
                { type = "text", content = fit(proc.cpu, 8) .. "│",
                  style = { foreground = cpuColor, bold = cpuVal > 20 } },
                { type = "text", content = fit(proc.mem, 10),
                  style = { foreground = theme.fg } },
            },
        }
    end

    return {
        type = "vbox",
        style = { flex = 1, overflow = "scroll" },
        children = children,
    }
end

-- ── Tab: Memory ────────────────────────────────────────────────────────────
local function renderMemory()
    return {
        type = "vbox",
        style = { flex = 1, padding = 1 },
        children = {
            { type = "text",
              content = " Memory Usage",
              style = { foreground = theme.accent, bold = true } },
            { type = "text", content = "" },
            progressBar(0.67, 40, "RAM"),
            { type = "text",
              content = "              Used: 10.7 GB / 16.0 GB  (6 GB available)",
              style = { foreground = theme.muted } },
            { type = "text", content = "" },
            progressBar(0.12, 40, "Swap"),
            { type = "text",
              content = "              Used: 0.5 GB / 4.0 GB",
              style = { foreground = theme.muted } },
            { type = "text", content = "" },
            progressBar(0.43, 40, "Disk /"),
            { type = "text",
              content = "              Used: 215 GB / 500 GB",
              style = { foreground = theme.muted } },
            { type = "text", content = "" },
            progressBar(0.89, 40, "Disk /data"),
            { type = "text",
              content = "              Used: 890 GB / 1.0 TB  ⚠️ Low space!",
              style = { foreground = theme.warning, bold = true } },
        },
    }
end

-- ── Tab: Network ───────────────────────────────────────────────────────────
local function renderNetwork(spinnerFrame)
    return {
        type = "vbox",
        style = { flex = 1, padding = 1 },
        children = {
            { type = "text",
              content = " Network Activity",
              style = { foreground = theme.accent, bold = true } },
            { type = "text", content = "" },
            -- Active connections
            { type = "hbox",
              children = {
                  { type = "text",
                    content = "  ● eth0  ",
                    style = { foreground = theme.success } },
                  { type = "text",
                    content = "192.168.1.42  ↑ 12.3 MB/s  ↓ 45.6 MB/s",
                    style = { foreground = theme.fg } },
              } },
            { type = "hbox",
              children = {
                  { type = "text",
                    content = "  ● lo    ",
                    style = { foreground = theme.success } },
                  { type = "text",
                    content = "127.0.0.1     ↑ 1.2 MB/s   ↓ 1.2 MB/s",
                    style = { foreground = theme.fg } },
              } },
            { type = "hbox",
              children = {
                  { type = "text",
                    content = "  ○ wlan0 ",
                    style = { foreground = theme.muted } },
                  { type = "text",
                    content = "disconnected",
                    style = { foreground = theme.muted } },
              } },
            { type = "text", content = "" },
            -- Stats
            { type = "text",
              content = "  Open connections: 47",
              style = { foreground = theme.fg } },
            { type = "text",
              content = "  Packets/sec:     1,234 in  |  892 out",
              style = { foreground = theme.fg } },
            { type = "text",
              content = "  DNS queries:     56/min",
              style = { foreground = theme.fg } },
            { type = "text", content = "" },
            -- Spinner for live data
            {
                type = "hbox",
                children = {
                    { type = "text",
                      content = "  ⠋ ",
                      style = { foreground = theme.accent } },
                    { type = "text",
                      content = "Monitoring network traffic...",
                      style = { foreground = theme.muted } },
                },
            },
        },
    }
end

-- ── Main Component ─────────────────────────────────────────────────────────
local Dashboard = lumina.defineComponent({
    name = "Dashboard",

    init = function(props)
        return {
            activeTab = 1,
            tabs = { "Processes", "Memory", "Network" },
            spinnerFrame = 1,
            uptime = "3d 14h 22m",
            loadAvg = "2.34 1.89 1.56",
        }
    end,

    render = function(instance)
        local activeTab = instance.activeTab or 1

        -- Build tab bar
        local tabChildren = {}
        for i, tab in ipairs(instance.tabs) do
            local isActive = (i == activeTab)
            tabChildren[#tabChildren + 1] = {
                type = "text",
                content = isActive
                    and (" [ " .. tab .. " ] ")
                    or  ("   " .. tab .. "   "),
                style = {
                    foreground = isActive and theme.accent or theme.muted,
                    bold = isActive,
                    background = isActive and theme.surface or nil,
                },
            }
        end

        -- Render active tab content
        local tabContent
        if activeTab == 1 then
            tabContent = renderProcesses()
        elseif activeTab == 2 then
            tabContent = renderMemory()
        else
            tabContent = renderNetwork(instance.spinnerFrame)
        end

        return {
            type = "vbox",
            style = { flex = 1, background = theme.bg },
            children = {
                -- ── Header ─────────────────────────────────────────────
                {
                    type = "hbox",
                    style = { height = 1, background = theme.headerBg },
                    children = {
                        { type = "text",
                          content = " 📊 System Dashboard ",
                          style = { foreground = theme.accent, bold = true } },
                        { type = "text",
                          content = string.format("  Uptime: %s  |  Load: %s",
                              instance.uptime, instance.loadAvg),
                          style = { foreground = theme.muted } },
                    },
                },

                -- ── Tab Bar ────────────────────────────────────────────
                {
                    type = "hbox",
                    style = { height = 1, background = theme.surface },
                    children = tabChildren,
                },

                -- ── Tab Content ────────────────────────────────────────
                tabContent,

                -- ── Status Bar ─────────────────────────────────────────
                {
                    type = "hbox",
                    style = { height = 1, background = theme.headerBg },
                    children = {
                        { type = "text",
                          content = " Refresh: 2s ",
                          style = { foreground = theme.muted } },
                        { type = "text",
                          content = " [1-3] Switch Tab  [r] Refresh  [q] Quit ",
                          style = { foreground = theme.muted, flex = 1 } },
                    },
                },
            },
        }
    end,
})

-- ── Render ─────────────────────────────────────────────────────────────────
lumina.render(Dashboard, {})
