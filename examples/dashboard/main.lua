-- ============================================================================
-- Lumina Example: Admin Dashboard
-- ============================================================================
-- Showcases: Router, shadcn Card/Table/Button, Theme switcher, i18n,
--            Flexbox layout, createStore, useStore, useTheme
--
-- Run:  go run ./cmd/lumina examples/dashboard/main.lua
-- Web:  lumina.serve(8080) at bottom instead of lumina.run()
-- ============================================================================

local lumina = require("lumina")

-- ── i18n setup ─────────────────────────────────────────────────────────────
lumina.i18n.addTranslation("en", {
    ["app.title"]       = "Lumina Dashboard",
    ["nav.home"]        = "🏠 Home",
    ["nav.users"]       = "👥 Users",
    ["nav.settings"]    = "⚙ Settings",
    ["stats.users"]     = "Total Users",
    ["stats.revenue"]   = "Revenue",
    ["stats.orders"]    = "Orders",
    ["stats.growth"]    = "Growth",
    ["settings.theme"]  = "Theme",
    ["settings.lang"]   = "Language",
    ["table.name"]      = "Name",
    ["table.email"]     = "Email",
    ["table.role"]      = "Role",
    ["table.status"]    = "Status",
})

lumina.i18n.addTranslation("zh", {
    ["app.title"]       = "Lumina 控制台",
    ["nav.home"]        = "🏠 首页",
    ["nav.users"]       = "👥 用户",
    ["nav.settings"]    = "⚙ 设置",
    ["stats.users"]     = "用户总数",
    ["stats.revenue"]   = "收入",
    ["stats.orders"]    = "订单",
    ["stats.growth"]    = "增长率",
    ["settings.theme"]  = "主题",
    ["settings.lang"]   = "语言",
    ["table.name"]      = "姓名",
    ["table.email"]     = "邮箱",
    ["table.role"]      = "角色",
    ["table.status"]    = "状态",
})

-- ── Store ───────────────────────────────────────────────────────────────────
local store = lumina.createStore({
    state = {
        page = "home",
        theme = "catppuccin-mocha",
        locale = "en",
        stats = { users = 1234, revenue = 56789, orders = 890, growth = 12.5 },
        users = {
            { name = "Alice Chen",   email = "alice@example.com",   role = "Admin",  status = "Active" },
            { name = "Bob Wang",     email = "bob@example.com",     role = "Editor", status = "Active" },
            { name = "Carol Li",     email = "carol@example.com",   role = "Viewer", status = "Inactive" },
            { name = "David Zhang",  email = "david@example.com",   role = "Editor", status = "Active" },
            { name = "Eve Liu",      email = "eve@example.com",     role = "Admin",  status = "Active" },
        },
    },
})

-- ── Stat Card Component ────────────────────────────────────────────────────
local function StatCard(props)
    local theme = lumina.useTheme()
    return {
        type = "vbox",
        style = {
            width = 18, height = 5, border = "rounded",
            background = theme.colors.card or "#1E1E2E",
            padding = 1,
        },
        children = {
            { type = "text", content = props.label, style = { foreground = theme.colors.secondary or "#A6ADC8", dim = true } },
            { type = "text", content = tostring(props.value), style = { foreground = theme.colors.primary or "#89B4FA", bold = true } },
        },
    }
end

-- ── Home Page ──────────────────────────────────────────────────────────────
local function HomePage()
    local t = lumina.useTranslation()
    local state = lumina.useStore(store)
    local stats = state.stats or {}

    return {
        type = "vbox",
        style = { padding = 1 },
        children = {
            { type = "text", content = t("app.title"), style = { bold = true, foreground = "#89B4FA" } },
            { type = "text", content = "" },
            {
                type = "hbox",
                children = {
                    StatCard({ label = t("stats.users"),   value = stats.users or 0 }),
                    { type = "text", content = " " },
                    StatCard({ label = t("stats.revenue"), value = "$" .. tostring(stats.revenue or 0) }),
                    { type = "text", content = " " },
                    StatCard({ label = t("stats.orders"),  value = stats.orders or 0 }),
                    { type = "text", content = " " },
                    StatCard({ label = t("stats.growth"),  value = tostring(stats.growth or 0) .. "%" }),
                },
            },
        },
    }
end

-- ── Users Page ─────────────────────────────────────────────────────────────
local function UsersPage()
    local t = lumina.useTranslation()
    local state = lumina.useStore(store)
    local users = state.users or {}

    local rows = {}
    for _, u in ipairs(users) do
        local statusColor = u.status == "Active" and "#A6E3A1" or "#F38BA8"
        rows[#rows + 1] = {
            type = "hbox",
            style = { height = 1 },
            children = {
                { type = "text", content = string.format("%-14s", u.name),   style = { foreground = "#CDD6F4" } },
                { type = "text", content = string.format("%-22s", u.email),  style = { foreground = "#A6ADC8" } },
                { type = "text", content = string.format("%-8s",  u.role),   style = { foreground = "#89B4FA" } },
                { type = "text", content = string.format("%-8s",  u.status), style = { foreground = statusColor } },
            },
        }
    end

    return {
        type = "vbox",
        style = { padding = 1 },
        children = {
            { type = "text", content = t("nav.users"), style = { bold = true, foreground = "#89B4FA" } },
            { type = "text", content = "" },
            -- Header
            {
                type = "hbox",
                style = { height = 1 },
                children = {
                    { type = "text", content = string.format("%-14s", t("table.name")),   style = { bold = true, foreground = "#F5C2E7" } },
                    { type = "text", content = string.format("%-22s", t("table.email")),  style = { bold = true, foreground = "#F5C2E7" } },
                    { type = "text", content = string.format("%-8s",  t("table.role")),   style = { bold = true, foreground = "#F5C2E7" } },
                    { type = "text", content = string.format("%-8s",  t("table.status")), style = { bold = true, foreground = "#F5C2E7" } },
                },
            },
            { type = "text", content = string.rep("─", 52), style = { foreground = "#45475A" } },
            table.unpack(rows),
        },
    }
end

-- ── Settings Page ──────────────────────────────────────────────────────────
local function SettingsPage()
    local t = lumina.useTranslation()
    local state = lumina.useStore(store)

    return {
        type = "vbox",
        style = { padding = 1 },
        children = {
            { type = "text", content = t("nav.settings"), style = { bold = true, foreground = "#89B4FA" } },
            { type = "text", content = "" },
            {
                type = "hbox",
                children = {
                    { type = "text", content = t("settings.theme") .. ": ", style = { foreground = "#CDD6F4" } },
                    { type = "text", content = state.theme or "catppuccin-mocha", style = { foreground = "#A6E3A1" } },
                },
            },
            {
                type = "hbox",
                children = {
                    { type = "text", content = t("settings.lang") .. ": ", style = { foreground = "#CDD6F4" } },
                    { type = "text", content = state.locale or "en", style = { foreground = "#A6E3A1" } },
                },
            },
            { type = "text", content = "" },
            { type = "text", content = "[T] Toggle theme  [L] Toggle language  [Q] Quit", style = { foreground = "#6C7086", dim = true } },
        },
    }
end

-- ── Sidebar ────────────────────────────────────────────────────────────────
local function Sidebar()
    local t = lumina.useTranslation()
    local state = lumina.useStore(store)
    local page = state.page or "home"

    local items = {
        { key = "home",     label = t("nav.home") },
        { key = "users",    label = t("nav.users") },
        { key = "settings", label = t("nav.settings") },
    }

    local children = {}
    for _, item in ipairs(items) do
        local isActive = (page == item.key)
        children[#children + 1] = {
            type = "text",
            content = (isActive and " ▸ " or "   ") .. item.label,
            style = {
                foreground = isActive and "#89B4FA" or "#A6ADC8",
                bold = isActive,
                background = isActive and "#313244" or nil,
            },
        }
    end

    return {
        type = "vbox",
        style = { width = 20, border = "single", background = "#181825" },
        children = children,
    }
end

-- ── Main App ───────────────────────────────────────────────────────────────
local App = lumina.defineComponent({
    name = "DashboardApp",
    render = function(self)
        local state = lumina.useStore(store)
        local page = state.page or "home"

        local content
        if page == "users" then
            content = UsersPage()
        elseif page == "settings" then
            content = SettingsPage()
        else
            content = HomePage()
        end

        return {
            type = "hbox",
            children = {
                Sidebar(),
                { type = "vbox", style = { flex = 1 }, children = { content } },
            },
        }
    end,
})

-- ── Key bindings ───────────────────────────────────────────────────────────
lumina.onKey("1", function() store.dispatch("setState", { page = "home" }) end)
lumina.onKey("2", function() store.dispatch("setState", { page = "users" }) end)
lumina.onKey("3", function() store.dispatch("setState", { page = "settings" }) end)
lumina.onKey("t", function()
    local state = store.getState()
    local newTheme = (state.theme == "catppuccin-mocha") and "catppuccin-latte" or "catppuccin-mocha"
    lumina.setTheme(newTheme)
    store.dispatch("setState", { theme = newTheme })
end)
lumina.onKey("l", function()
    local state = store.getState()
    local newLocale = (state.locale == "en") and "zh" or "en"
    lumina.i18n.setLocale(newLocale)
    store.dispatch("setState", { locale = newLocale })
end)
lumina.onKey("q", function() lumina.quit() end)

lumina.mount(App)
lumina.run()
