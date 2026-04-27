-- ============================================================================
-- Lumina v2 Example: Admin Dashboard
-- ============================================================================
-- Sidebar + home / users / settings, stat cards, user table, theme + locale.
-- v2: lumina.createComponent / useState / createElement / quit (no require,
--     createStore, mount, onKey, i18n API, setTheme).
--
-- Run:  lumina-v2 examples/dashboard/main.lua
-- Web:  lumina-v2 --web :8080 examples/dashboard/main.lua
--
-- Table columns use fixed style.width so headers align with cells under hbox.
--
-- Keys (focus root: click the top hint line if shortcuts are ignored):
--   1 / 2 / 3   Home / Users / Settings
--   t / T       Theme mocha ↔ latte
--   l / L       en ↔ zh
--   q / Q       Quit
-- ============================================================================

-- "lumina" is injected by lumina-v2. No require().

local i18n = {
    en = {
        ["app.title"] = "Lumina Dashboard",
        ["nav.home"] = "Home",
        ["nav.users"] = "Users",
        ["nav.settings"] = "Settings",
        ["stats.users"] = "Total Users",
        ["stats.revenue"] = "Revenue",
        ["stats.orders"] = "Orders",
        ["stats.growth"] = "Growth",
        ["settings.theme"] = "Theme",
        ["settings.lang"] = "Language",
        ["table.name"] = "Name",
        ["table.email"] = "Email",
        ["table.role"] = "Role",
        ["table.status"] = "Status",
    },
    zh = {
        ["app.title"] = "Lumina 控制台",
        ["nav.home"] = "首页",
        ["nav.users"] = "用户",
        ["nav.settings"] = "设置",
        ["stats.users"] = "用户总数",
        ["stats.revenue"] = "收入",
        ["stats.orders"] = "订单",
        ["stats.growth"] = "增长率",
        ["settings.theme"] = "主题",
        ["settings.lang"] = "语言",
        ["table.name"] = "姓名",
        ["table.email"] = "邮箱",
        ["table.role"] = "角色",
        ["table.status"] = "状态",
    },
}

local function T(locale, key)
    local pack = i18n[locale] or i18n.en
    return pack[key] or key
end

local function themePalette(name)
    if name == "catppuccin-latte" then
        return {
            card = "#EFF1F5",
            secondary = "#6C6F85",
            primary = "#1E66F5",
            muted = "#8C8FA1",
            sidebarBg = "#E6E9EF",
            sidebarFg = "#4C4F69",
            accentBg = "#DCE0E8",
        }
    end
    return {
        card = "#1E1E2E",
        secondary = "#A6ADC8",
        primary = "#89B4FA",
        muted = "#6C7086",
        sidebarBg = "#181825",
        sidebarFg = "#A6ADC8",
        accentBg = "#313244",
    }
end

local function StatCard(theme, label, value)
    return lumina.createElement("vbox", {
        style = {
            flex = 1,
            height = 5,
            border = "rounded",
            background = theme.card,
            padding = 1,
        },
    },
        lumina.createElement("text", {
            foreground = theme.secondary,
            dim = true,
        }, label),
        lumina.createElement("text", {
            foreground = theme.primary,
            bold = true,
        }, value)
    )
end

local function HomePage(theme, tFn, stats)
    return lumina.createElement("vbox", {
        style = { padding = 1 },
    },
        lumina.createElement("text", {
            foreground = theme.primary,
            bold = true,
        }, tFn("app.title")),
        lumina.createElement("text", {}, ""),
        lumina.createElement("hbox", {
            style = { gap = 1 },
        },
            StatCard(theme, tFn("stats.users"), tostring(stats.users or 0)),
            StatCard(theme, tFn("stats.revenue"), "$" .. tostring(stats.revenue or 0)),
            StatCard(theme, tFn("stats.orders"), tostring(stats.orders or 0)),
            StatCard(theme, tFn("stats.growth"), tostring(stats.growth or 0) .. "%")
        )
    )
end

-- Fixed widths: hbox flex + string.format byte padding misaligns CJK headers vs ASCII cells.
local COL_NAME, COL_EMAIL, COL_ROLE, COL_STATUS = 16, 28, 12, 12

local function usersTableCell(width, opts, text)
    return lumina.createElement("text", {
        style = { width = width, height = 1 },
        foreground = opts.fg,
        bold = opts.bold,
        dim = opts.dim,
    }, text)
end

local function UsersPage(theme, tFn, users)
    local sepW = COL_NAME + COL_EMAIL + COL_ROLE + COL_STATUS
    local rows = {}
    for _, u in ipairs(users) do
        local statusColor = (u.status == "Active") and "#A6E3A1" or "#F38BA8"
        rows[#rows + 1] = lumina.createElement("hbox", {
            style = { height = 1 },
        },
            usersTableCell(COL_NAME, { fg = "#CDD6F4" }, u.name),
            usersTableCell(COL_EMAIL, { fg = "#A6ADC8" }, u.email),
            usersTableCell(COL_ROLE, { fg = "#89B4FA" }, u.role),
            usersTableCell(COL_STATUS, { fg = statusColor }, u.status)
        )
    end

    local children = {
        lumina.createElement("text", { foreground = theme.primary, bold = true }, tFn("nav.users")),
        lumina.createElement("text", {}, ""),
        lumina.createElement("hbox", { style = { height = 1 } },
            usersTableCell(COL_NAME, { fg = "#F5C2E7", bold = true }, tFn("table.name")),
            usersTableCell(COL_EMAIL, { fg = "#F5C2E7", bold = true }, tFn("table.email")),
            usersTableCell(COL_ROLE, { fg = "#F5C2E7", bold = true }, tFn("table.role")),
            usersTableCell(COL_STATUS, { fg = "#F5C2E7", bold = true }, tFn("table.status"))
        ),
        lumina.createElement("text", { foreground = "#45475A" }, string.rep("─", sepW)),
    }
    for i = 1, #rows do
        children[#children + 1] = rows[i]
    end

    return lumina.createElement("vbox", { style = { padding = 1 } }, table.unpack(children))
end

local function SettingsPage(theme, tFn, themeName, locale)
    return lumina.createElement("vbox", {
        style = { padding = 1 },
    },
        lumina.createElement("text", { foreground = theme.primary, bold = true }, tFn("nav.settings")),
        lumina.createElement("text", {}, ""),
        lumina.createElement("hbox", {},
            lumina.createElement("text", { foreground = "#CDD6F4" }, tFn("settings.theme") .. ": "),
            lumina.createElement("text", { foreground = "#A6E3A1" }, themeName)
        ),
        lumina.createElement("hbox", {},
            lumina.createElement("text", { foreground = "#CDD6F4" }, tFn("settings.lang") .. ": "),
            lumina.createElement("text", { foreground = "#A6E3A1" }, locale)
        ),
        lumina.createElement("text", {}, ""),
        lumina.createElement("text", { foreground = "#6C7086", dim = true },
            "[T] Theme  [L] Language  [1-3] Pages  [Q] Quit")
    )
end

local function Sidebar(theme, tFn, page, setPage)
    local items = {
        { key = "home", label = tFn("nav.home") },
        { key = "users", label = tFn("nav.users") },
        { key = "settings", label = tFn("nav.settings") },
    }
    local children = {}
    for _, item in ipairs(items) do
        local isActive = (page == item.key)
        local prefix = isActive and " > " or "   "
        children[#children + 1] = lumina.createElement("text", {
            id = "nav-" .. item.key,
            foreground = isActive and theme.primary or theme.sidebarFg,
            bold = isActive,
            background = isActive and theme.accentBg or nil,
            onClick = function()
                setPage(item.key)
            end,
        }, prefix .. item.label)
    end
    return lumina.createElement("vbox", {
        style = {
            width = 20,
            border = "single",
            background = theme.sidebarBg,
        },
    }, table.unpack(children))
end

lumina.createComponent({
    id = "dashboard",
    name = "DashboardApp",
    x = 0,
    y = 0,
    w = 100,
    h = 35,
    zIndex = 0,

    render = function(state, props)
        local page, setPage = lumina.useState("page", "home")
        local themeName, setThemeName = lumina.useState("theme", "catppuccin-mocha")
        local locale, setLocale = lumina.useState("locale", "en")
        local stats, _setStats = lumina.useState("stats", {
            users = 1234,
            revenue = 56789,
            orders = 890,
            growth = 12.5,
        })
        local users, _setUsers = lumina.useState("users", {
            { name = "Alice Chen", email = "alice@example.com", role = "Admin", status = "Active" },
            { name = "Bob Wang", email = "bob@example.com", role = "Editor", status = "Active" },
            { name = "Carol Li", email = "carol@example.com", role = "Viewer", status = "Inactive" },
            { name = "David Zhang", email = "david@example.com", role = "Editor", status = "Active" },
            { name = "Eve Liu", email = "eve@example.com", role = "Admin", status = "Active" },
        })

        local function tFn(key)
            return T(locale, key)
        end

        local pal = themePalette(themeName)

        local content
        if page == "users" then
            content = UsersPage(pal, tFn, users)
        elseif page == "settings" then
            content = SettingsPage(pal, tFn, themeName, locale)
        else
            content = HomePage(pal, tFn, stats)
        end

        local function onKeyDown(e)
            local k = e.key
            if k == "1" then
                setPage("home")
            elseif k == "2" then
                setPage("users")
            elseif k == "3" then
                setPage("settings")
            elseif k == "t" or k == "T" then
                if themeName == "catppuccin-mocha" then
                    setThemeName("catppuccin-latte")
                else
                    setThemeName("catppuccin-mocha")
                end
            elseif k == "l" or k == "L" then
                if locale == "en" then
                    setLocale("zh")
                else
                    setLocale("en")
                end
            elseif k == "q" or k == "Q" then
                lumina.quit()
            end
        end

        return lumina.createElement("vbox", {
            id = "dashboard-root",
            style = {
                background = pal.card,
                height = 35,
            },
            onKeyDown = onKeyDown,
            focusable = true,
        },
            lumina.createElement("text", {
                foreground = pal.muted,
                dim = true,
                style = { height = 1 },
            }, " [1-3] Pages  [T] Theme  [L] Language  [Q] Quit — click here if keys ignored "),
            lumina.createElement("hbox", {
                style = { flex = 1 },
            },
                Sidebar(pal, tFn, page, setPage),
                lumina.createElement("vbox", { style = { flex = 1 } }, content)
            )
        )
    end,
})
