-- ============================================================================
-- Lumina Example: File Explorer
-- ============================================================================
-- Showcases: Tree view, breadcrumbs, keyboard navigation, preview panel,
--            collapsible nodes, conditional rendering, flexbox layout
--
-- Run: go run ./cmd/lumina examples/file-explorer/main.lua
-- ============================================================================

local lumina = require("lumina")

-- ── File System Data (simulated) ────────────────────────────────────────────
local filesystem = {
    { name = "src",       type = "dir", children = {
        { name = "main.go",     type = "file", size = "2.4 KB", content = "package main\n\nfunc main() {\n    // ...\n}" },
        { name = "app.go",      type = "file", size = "8.1 KB", content = "package main\n\ntype App struct {\n    // ...\n}" },
        { name = "handlers.go", type = "file", size = "5.3 KB", content = "package main\n\nfunc handleRequest() {\n    // ...\n}" },
        { name = "utils",       type = "dir", children = {
            { name = "helpers.go", type = "file", size = "1.2 KB", content = "package utils\n\nfunc FormatSize() string {\n    // ...\n}" },
            { name = "config.go",  type = "file", size = "0.8 KB", content = "package utils\n\nvar Config = map[string]string{}" },
        }},
    }},
    { name = "docs",      type = "dir", children = {
        { name = "README.md",   type = "file", size = "3.2 KB", content = "# Project Documentation\n\nWelcome to the project." },
        { name = "API.md",      type = "file", size = "6.7 KB", content = "# API Reference\n\n## Endpoints\n\n### GET /api/v1/..." },
        { name = "CHANGELOG.md", type = "file", size = "1.5 KB", content = "# Changelog\n\n## v1.0.0\n- Initial release" },
    }},
    { name = "tests",     type = "dir", children = {
        { name = "main_test.go", type = "file", size = "4.1 KB", content = "package main\n\nfunc TestMain(t *testing.T) {\n    // ...\n}" },
    }},
    { name = "go.mod",    type = "file", size = "0.3 KB", content = "module example.com/project\n\ngo 1.21" },
    { name = "go.sum",    type = "file", size = "12.4 KB", content = "(checksum file)" },
    { name = "Makefile",  type = "file", size = "0.5 KB", content = "build:\n\tgo build ./..." },
    { name = ".gitignore", type = "file", size = "0.1 KB", content = "*.exe\n*.o\n/bin/" },
}

-- ── Store ───────────────────────────────────────────────────────────────────
local store = lumina.createStore({
    state = {
        path = {},           -- current directory path (breadcrumb)
        entries = filesystem, -- current directory entries
        selected = 1,        -- selected index
        expanded = {},       -- expanded directory names
        preview = nil,       -- preview content
    },
})

-- ── Colors ──────────────────────────────────────────────────────────────────
local colors = {
    bg       = "#1E1E2E",
    fg       = "#CDD6F4",
    accent   = "#89B4FA",
    muted    = "#6C7086",
    surface  = "#313244",
    border   = "#45475A",
    dir      = "#89B4FA",
    file     = "#CDD6F4",
    lua      = "#A6E3A1",
    go       = "#89DCEB",
    md       = "#F9E2AF",
    hidden   = "#585B70",
}

-- ── Helpers ─────────────────────────────────────────────────────────────────
local function getIcon(entry)
    if entry.type == "dir" then return "📁 " end
    local ext = entry.name:match("%.(%w+)$") or ""
    local icons = {
        go = "🔷 ", lua = "🌙 ", md = "📝 ", mod = "📦 ",
        sum = "🔒 ", gitignore = "🙈 ", Makefile = "⚙ ",
    }
    return icons[ext] or "📄 "
end

local function getColor(entry)
    if entry.type == "dir" then return colors.dir end
    if entry.name:sub(1, 1) == "." then return colors.hidden end
    local ext = entry.name:match("%.(%w+)$") or ""
    local map = { go = colors.go, lua = colors.lua, md = colors.md }
    return map[ext] or colors.file
end

local function getCurrentEntries(state)
    local entries = filesystem
    for _, dir in ipairs(state.path or {}) do
        for _, e in ipairs(entries) do
            if e.name == dir and e.type == "dir" then
                entries = e.children or {}
                break
            end
        end
    end
    return entries
end

-- ── Breadcrumb Component ────────────────────────────────────────────────────
local function Breadcrumb()
    local state = lumina.useStore(store)
    local path = state.path or {}

    local parts = { { type = "text", content = " 🏠 root", style = { foreground = colors.accent, bold = true } } }
    for _, p in ipairs(path) do
        parts[#parts + 1] = { type = "text", content = " / ", style = { foreground = colors.muted } }
        parts[#parts + 1] = { type = "text", content = p, style = { foreground = colors.accent } }
    end

    return { type = "hbox", children = parts }
end

-- ── File List Component ─────────────────────────────────────────────────────
local function FileList()
    local state = lumina.useStore(store)
    local entries = getCurrentEntries(state)
    local selected = state.selected or 1

    if #entries == 0 then
        return { type = "text", content = "  (empty directory)", style = { foreground = colors.muted, italic = true } }
    end

    local children = {}
    for i, entry in ipairs(entries) do
        local isSelected = (i == selected)
        local icon = getIcon(entry)
        local color = getColor(entry)
        local sizeStr = entry.size and ("  " .. entry.size) or ""

        children[#children + 1] = {
            type = "hbox",
            style = {
                height = 1,
                background = isSelected and colors.surface or nil,
            },
            children = {
                { type = "text", content = isSelected and " ▸ " or "   ", style = { foreground = colors.accent } },
                { type = "text", content = icon },
                { type = "text", content = entry.name, style = { foreground = color, bold = entry.type == "dir" } },
                { type = "text", content = sizeStr, style = { foreground = colors.muted, dim = true } },
            },
        }
    end

    return { type = "vbox", children = children }
end

-- ── Preview Panel ───────────────────────────────────────────────────────────
local function PreviewPanel()
    local state = lumina.useStore(store)
    local entries = getCurrentEntries(state)
    local selected = state.selected or 1
    local entry = entries[selected]

    if not entry then
        return { type = "text", content = "" }
    end

    if entry.type == "dir" then
        local count = entry.children and #entry.children or 0
        return {
            type = "vbox",
            style = { border = "single", width = 35 },
            children = {
                { type = "text", content = " 📁 " .. entry.name, style = { foreground = colors.dir, bold = true } },
                { type = "text", content = " " .. count .. " items", style = { foreground = colors.muted } },
            },
        }
    end

    local content = entry.content or "(no preview)"
    local lines = {}
    for line in (content .. "\n"):gmatch("(.-)\n") do
        lines[#lines + 1] = { type = "text", content = " " .. line, style = { foreground = colors.fg } }
    end

    return {
        type = "vbox",
        style = { border = "single", width = 35 },
        children = {
            { type = "text", content = " " .. getIcon(entry) .. entry.name, style = { foreground = getColor(entry), bold = true } },
            { type = "text", content = " Size: " .. (entry.size or "?"), style = { foreground = colors.muted } },
            { type = "text", content = string.rep("─", 33), style = { foreground = colors.border } },
            table.unpack(lines),
        },
    }
end

-- ── Main App ───────────────────────────────────────────────────────────────
local App = lumina.defineComponent({
    name = "FileExplorer",
    render = function(self)
        return {
            type = "vbox",
            style = { background = colors.bg },
            children = {
                { type = "text", content = " 📂 Lumina File Explorer ", style = { foreground = colors.accent, bold = true, background = "#181825" } },
                Breadcrumb(),
                { type = "text", content = string.rep("─", 70), style = { foreground = colors.border } },
                {
                    type = "hbox",
                    children = {
                        { type = "vbox", style = { width = 35 }, children = { FileList() } },
                        { type = "text", content = "│", style = { foreground = colors.border } },
                        PreviewPanel(),
                    },
                },
                { type = "text", content = "" },
                { type = "text", content = " [j]↓ [k]↑ [enter]open [backspace]back [q]uit",
                  style = { foreground = colors.muted, dim = true } },
            },
        }
    end,
})

-- ── Key bindings ───────────────────────────────────────────────────────────
lumina.onKey("j", function()
    local state = store.getState()
    local entries = getCurrentEntries(state)
    local sel = math.min((state.selected or 1) + 1, #entries)
    store.dispatch("setState", { selected = sel })
end)

lumina.onKey("k", function()
    local state = store.getState()
    local sel = math.max((state.selected or 1) - 1, 1)
    store.dispatch("setState", { selected = sel })
end)

lumina.onKey("enter", function()
    local state = store.getState()
    local entries = getCurrentEntries(state)
    local sel = state.selected or 1
    local entry = entries[sel]
    if entry and entry.type == "dir" then
        local path = state.path or {}
        path[#path + 1] = entry.name
        store.dispatch("setState", { path = path, selected = 1 })
    end
end)

lumina.onKey("backspace", function()
    local state = store.getState()
    local path = state.path or {}
    if #path > 0 then
        table.remove(path)
        store.dispatch("setState", { path = path, selected = 1 })
    end
end)

lumina.onKey("q", function() lumina.quit() end)

lumina.mount(App)
lumina.run()
