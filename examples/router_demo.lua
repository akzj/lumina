-- examples/router_demo.lua — Multi-page TUI with lumina.router
-- Usage: lumina examples/router_demo.lua
-- Keys: 1=Home, 2=Users, 3=Settings, b=Back, q=Quit

local lux = require("lux")
local Breadcrumb = lux.Breadcrumb
local DataGrid = lux.DataGrid
local Alert = lux.Alert

local users = {
    { id = 1, name = "Alice", role = "Admin", status = "active" },
    { id = 2, name = "Bob", role = "Editor", status = "active" },
    { id = 3, name = "Charlie", role = "Viewer", status = "inactive" },
    { id = 4, name = "Diana", role = "Admin", status = "active" },
    { id = 5, name = "Eve", role = "Editor", status = "pending" },
}

local function renderHome(t)
    return lumina.createElement("vbox", { style = { height = 10 } },
        lumina.createElement("text", {
            foreground = t.text or "#CDD6F4",
            bold = true,
            style = { height = 1 },
        }, "  Welcome to Router Demo"),
        lumina.createElement("text", { style = { height = 1 } }, ""),
        lumina.createElement("text", {
            foreground = t.primary or "#89B4FA",
            style = { height = 1 },
        }, "  Press 1 for Home"),
        lumina.createElement("text", {
            foreground = t.primary or "#89B4FA",
            style = { height = 1 },
        }, "  Press 2 for Users"),
        lumina.createElement("text", {
            foreground = t.primary or "#89B4FA",
            style = { height = 1 },
        }, "  Press 3 for Settings"),
        lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            style = { height = 1 },
        }, "  Press b to go back")
    )
end

local function renderUsers(t, selectedIdx)
    return DataGrid {
        id = "users-grid",
        key = "users-grid",
        width = 50,
        height = 8,
        columns = {
            { id = "id", header = "#", width = 5, key = "id" },
            { id = "name", header = "Name", width = 15, key = "name" },
            { id = "role", header = "Role", width = 12, key = "role" },
            { id = "status", header = "Status", width = 10, key = "status" },
        },
        rows = users,
        selectedIndex = selectedIdx,
        onChangeIndex = function(i)
            lumina.store.set("selectedIdx", i)
        end,
        onActivate = function(i, row)
            lumina.router.navigate("/users/" .. tostring(row.id))
        end,
        autoFocus = true,
    }
end

local function renderUserDetail(t, params)
    local userId = tonumber(params.id) or 0
    local user = nil
    for _, u in ipairs(users) do
        if u.id == userId then user = u; break end
    end
    if not user then
        return Alert {
            key = "not-found",
            variant = "error",
            title = "Not Found",
            message = "User #" .. tostring(userId) .. " not found",
            width = 50,
        }
    end
    return lumina.createElement("vbox", { style = { height = 6 } },
        lumina.createElement("text", {
            foreground = t.text or "#CDD6F4",
            bold = true,
            style = { height = 1 },
        }, "  User: " .. user.name),
        lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            style = { height = 1 },
        }, "  Role: " .. user.role),
        lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            style = { height = 1 },
        }, "  Status: " .. user.status),
        lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            style = { height = 1 },
        }, "  Press b to go back to users list")
    )
end

local function renderSettings(t)
    return lumina.createElement("vbox", { style = { height = 4 } },
        lumina.createElement("text", {
            foreground = t.text or "#CDD6F4",
            bold = true,
            style = { height = 1 },
        }, "  Settings"),
        lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            style = { height = 1 },
        }, "  Theme: Catppuccin Mocha"),
        lumina.createElement("text", {
            foreground = t.muted or "#6C7086",
            style = { height = 1 },
        }, "  (No other settings available yet)")
    )
end

lumina.app {
    id = "router-demo",
    store = {
        selectedIdx = 1,
    },
    routes = {
        ["/"] = true,
        ["/users"] = true,
        ["/users/:id"] = true,
        ["/settings"] = true,
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
        ["1"] = function() lumina.router.navigate("/") end,
        ["2"] = function() lumina.router.navigate("/users") end,
        ["3"] = function() lumina.router.navigate("/settings") end,
        ["b"] = function() lumina.router.back() end,
    },
    render = function()
        local t = lumina.getTheme()
        local route = lumina.useRoute()
        local path = route.path
        local params = route.params
        local selectedIdx = lumina.useStore("selectedIdx")

        -- Breadcrumb items
        local crumbs = {{ id = "home", label = "Home" }}
        if path == "/users" then
            crumbs[#crumbs + 1] = { id = "users", label = "Users" }
        elseif path:match("^/users/") then
            crumbs[#crumbs + 1] = { id = "users", label = "Users" }
            crumbs[#crumbs + 1] = { id = "detail", label = "User #" .. (params.id or "?") }
        elseif path == "/settings" then
            crumbs[#crumbs + 1] = { id = "settings", label = "Settings" }
        end

        -- Route content
        local content
        if path == "/users" then
            content = renderUsers(t, selectedIdx)
        elseif path:match("^/users/") then
            content = renderUserDetail(t, params)
        elseif path == "/settings" then
            content = renderSettings(t)
        else
            content = renderHome(t)
        end

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 60, height = 20, background = t.base or "#1E1E2E" },
        },
            lumina.createElement("text", {
                foreground = t.primary or "#89B4FA",
                bold = true,
                style = { height = 1 },
            }, "  Router Demo  [1:Home 2:Users 3:Settings b:Back q:Quit]"),
            Breadcrumb {
                key = "bc",
                items = crumbs,
                onNavigate = function(id)
                    if id == "home" then lumina.router.navigate("/")
                    elseif id == "users" then lumina.router.navigate("/users")
                    elseif id == "settings" then lumina.router.navigate("/settings")
                    end
                end,
            },
            lumina.createElement("text", { style = { height = 1 } }, ""),
            content
        )
    end,
}
