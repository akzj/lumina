-- ============================================================================
-- Lumina Example: Navigation & Routing
-- ============================================================================
-- Demonstrates: Router, NavigationMenu, Breadcrumb, Sidebar, nested routes
-- Run: lumina examples/navigation/main.lua
-- ============================================================================
local lumina = require("lumina")
local shadcn = require("shadcn")

lumina.setTheme("catppuccin-mocha")

local c = {
    bg = "#1E1E2E", fg = "#CDD6F4", accent = "#89B4FA",
    success = "#A6E3A1", error = "#F38BA8", muted = "#6C7086",
    surface = "#313244", border = "#45475A", warning = "#F9E2AF",
    pink = "#F5C2E7",
}

-- Simple hash router
local store = lumina.createStore({
    state = {
        route = "#/dashboard",
        sidebarOpen = true,
        sidebarWidth = 20,
    },
})

local function setRoute(path)
    store.dispatch("setState", { route = path })
end

local function getCurrentRoute()
    return store.getState().route or "#/dashboard"
end

-- Page definitions
local pages = {
    { path = "#/dashboard", label = "Dashboard", icon = "📊", content = "Dashboard" },
    { path = "#/users", label = "Users", icon = "👥", content = "Users" },
    { path = "#/users/1", label = "User Detail", icon = "👤", content = "User Detail" },
    { path = "#/settings", label = "Settings", icon = "⚙", content = "Settings" },
    { path = "#/settings/profile", label = "Profile", icon = "👤", content = "Profile Settings" },
    { path = "#/settings/security", label = "Security", icon = "🔒", content = "Security Settings" },
    { path = "#/about", label = "About", icon = "ℹ", content = "About" },
}

-- Breadcrumb
local function getBreadcrumb(route)
    local items = {}
    local parts = {}
    for part in route:gmatch("[^/]+") do
        table.insert(parts, part)
    end

    for i, part in ipairs(parts) do
        local path = table.concat(parts, "/", 1, i)
        local label = part:gsub("#", "")
        label = label:sub(1, 1):upper() .. label:sub(2)

        for _, p in ipairs(pages) do
            if p.path == path then
                label = p.label
                break
            end
        end

        table.insert(items, { label = label, path = path, active = (i == #parts) })
    end

    return items
end

local function Breadcrumb(props)
    local items = props.items or {}
    local children = {}

    for i, item in ipairs(items) do
        children[#children + 1] = {
            type = "text",
            content = item.label,
            style = { foreground = item.active and c.accent or c.muted, bold = item.active },
        }
        if i < #items then
            children[#children + 1] = {
                type = "text",
                content = " / ",
                style = { foreground = c.border },
            }
        end
    end

    return { type = "hbox", children = children }
end

-- Sidebar
local function SidebarItem(props)
    local page = props.page
    local active = props.active
    local collapsed = props.collapsed

    if collapsed then
        return {
            type = "text",
            content = page.icon,
            style = {
                foreground = active and c.accent or c.fg,
                bold = active,
            },
        }
    else
        return {
            type = "hbox",
            style = {
                background = active and c.surface or "",
                foreground = active and c.accent or c.fg,
            },
            children = {
                { type = "text", content = "  " .. page.icon .. "  " },
                { type = "text", content = page.label, bold = active },
            },
        }
    end
end

local function Sidebar(props)
    local collapsed = props.collapsed
    local currentRoute = props.route

    local navItems = {
        { path = "#/dashboard", label = "Dashboard", icon = "📊" },
        { path = "#/users", label = "Users", icon = "👥" },
        { path = "#/settings", label = "Settings", icon = "⚙" },
        { path = "#/about", label = "About", icon = "ℹ" },
    }

    -- Get current section for users/settings
    local currentSection = currentRoute:match("^#/(%w+)")

    local children = {
        { type = "text", content = collapsed and "" or " Navigation ", style = { foreground = c.muted, bold = true } },
        { type = "text", content = "" },
    }

    for _, item in ipairs(navItems) do
        local active = currentRoute == item.path or currentRoute:find("^" .. item.path) == 1
        children[#children + 1] = SidebarItem({ page = item, active = active, collapsed = collapsed })
    end

    -- Sub-items for Users
    if currentSection == "users" then
        children[#children + 1] = { type = "text", content = "  ↳ User #1", style = { foreground = c.muted } }
    end

    -- Sub-items for Settings
    if currentSection == "settings" then
        children[#children + 1] = { type = "text", content = "  ↳ Profile", style = { foreground = c.muted } }
        children[#children + 1] = { type = "text", content = "  ↳ Security", style = { foreground = c.muted } }
    end

    local width = collapsed and 3 or 20

    return {
        type = "vbox",
        style = {
            border = "right",
            borderColor = c.border,
            background = c.bg,
            foreground = c.fg,
            padding = 0,
            width = width,
        },
        children = children,
    }
end

-- Page content
local function getPageContent(route)
    for _, page in ipairs(pages) do
        if page.path == route then
            return page.content
        end
    end
    return "Page not found"
end

local function getPageIcon(route)
    for _, page in ipairs(pages) do
        if page.path == route then
            return page.icon
        end
    end
    return "?"
end

local App = lumina.defineComponent({
    name = "NavigationApp",
    render = function(self)
        local state = lumina.useStore(store)
        local route = state.route or "#/dashboard"
        local sidebarOpen = state.sidebarOpen

        local breadcrumb = getBreadcrumb(route)
        local content = getPageContent(route)
        local icon = getPageIcon(route)

        -- Mock content for each page
        local pageBody
        if route == "#/dashboard" then
            pageBody = {
                { type = "text", content = " Welcome to the Dashboard! ", style = { foreground = c.accent, bold = true } },
                { type = "text", content = "" },
                { type = "text", content = " This is your main workspace. ", style = { foreground = c.fg } },
                { type = "text", content = "" },
                { type = "text", content = " Quick Stats:", style = { foreground = c.fg, bold = true } },
                { type = "text", content = "  📊 1,234 active users", style = { foreground = c.muted } },
                { type = "text", content = "  📈 89% satisfaction", style = { foreground = c.muted } },
                { type = "text", content = "  ⚡ 99.9% uptime", style = { foreground = c.muted } },
            }
        elseif route == "#/users" then
            pageBody = {
                { type = "text", content = " User Management ", style = { foreground = c.accent, bold = true } },
                { type = "text", content = "" },
                { type = "text", content = " All registered users (1,234 total):", style = { foreground = c.fg } },
                { type = "text", content = "" },
                { type = "text", content = "  👤 Alice Chen      alice@example.com      Active", style = { foreground = c.fg } },
                { type = "text", content = "  👤 Bob Wang       bob@example.com       Active", style = { foreground = c.fg } },
                { type = "text", content = "  👤 Carol Li       carol@example.com     Inactive", style = { foreground = c.muted } },
                { type = "text", content = "" },
                { type = "text", content = "  Press [1] to view User #1 detail", style = { foreground = c.muted, dim = true } },
            }
        elseif route:find("^#/settings") then
            pageBody = {
                { type = "text", content = " Settings ", style = { foreground = c.accent, bold = true } },
                { type = "text", content = "" },
                { type = "text", content = " Configure your preferences:", style = { foreground = c.fg } },
                { type = "text", content = "" },
                { type = "text", content = "  [p] Profile Settings", style = { foreground = c.fg } },
                { type = "text", content = "  [s] Security Settings", style = { foreground = c.fg } },
                { type = "text", content = "  [n] Notifications", style = { foreground = c.fg } },
                { type = "text", content = "  [b] Billing", style = { foreground = c.fg } },
            }
        elseif route == "#/about" then
            pageBody = {
                { type = "text", content = " About Lumina ", style = { foreground = c.accent, bold = true } },
                { type = "text", content = "" },
                { type = "text", content = " Version 1.0.0", style = { foreground = c.fg } },
                { type = "text", content = "" },
                { type = "text", content = " A React-style TUI framework for Go + Lua.", style = { foreground = c.muted } },
                { type = "text", content = " Build beautiful terminal interfaces.", style = { foreground = c.muted } },
            }
        else
            pageBody = {
                { type = "text", content = " " .. content .. " ", style = { foreground = c.accent, bold = true } },
                { type = "text", content = "" },
                { type = "text", content = " Page content goes here.", style = { foreground = c.fg } },
            }
        end

        local body = {
            { type = "text", content = " " .. icon .. " " .. content, style = { foreground = c.accent, bold = true, background = "#181825" } },
            { type = "text", content = "" },
            Breadcrumb({ items = breadcrumb }),
            { type = "text", content = "" },
            { type = "text", content = string.rep("─", 60), style = { foreground = c.border } },
            { type = "text", content = "" },
            unpack(pageBody),
        }

        local mainContent = {
            type = "vbox",
            style = { padding = 1 },
            children = body,
        }

        if sidebarOpen then
            return {
                type = "hbox",
                style = { background = c.bg },
                children = {
                    Sidebar({ collapsed = false, route = route }),
                    mainContent,
                },
            }
        else
            return mainContent
        end
    end,
})

-- Key bindings for navigation
lumina.onKey("1", function() setRoute("#/dashboard") end)
lumina.onKey("2", function() setRoute("#/users") end)
lumina.onKey("3", function() setRoute("#/settings") end)
lumina.onKey("4", function() setRoute("#/about") end)
lumina.onKey("u", function() setRoute("#/users/1") end)
lumina.onKey("p", function() setRoute("#/settings/profile") end)
lumina.onKey("o", function() setRoute("#/settings/security") end)

lumina.onKey("t", function()
    local state = store.getState()
    store.dispatch("setState", { sidebarOpen = not state.sidebarOpen })
end)

lumina.onKey("q", function() lumina.quit() end)

lumina.mount(App)
lumina.run()
