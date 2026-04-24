-- ============================================================================
-- Lumina Example: File Browser
-- ============================================================================
-- Showcases: Tree component, keyboard navigation, StatusBar, conditional
--            rendering, nested VNode trees, icon mapping.
--
-- Run: lumina examples/file_browser.lua
-- ============================================================================

local lumina = require("lumina")

-- ── File type icons ────────────────────────────────────────────────────────
local icons = {
    folder_open  = "📂",
    folder_close = "📁",
    lua          = "🌙",
    go           = "🐹",
    md           = "📝",
    txt          = "📄",
    json         = "📋",
    default      = "📄",
}

local function getIcon(name, isDir, isExpanded)
    if isDir then
        return isExpanded and icons.folder_open or icons.folder_close
    end
    local ext = name:match("%.(%w+)$")
    return icons[ext] or icons.default
end

-- ── Theme ──────────────────────────────────────────────────────────────────
local theme = {
    bg       = "#1E1E2E",
    fg       = "#CDD6F4",
    accent   = "#89B4FA",
    dir      = "#F9E2AF",
    file     = "#CDD6F4",
    selected = "#45475A",
    muted    = "#6C7086",
    border   = "#585B70",
    headerBg = "#181825",
}

-- ── Sample file tree ───────────────────────────────────────────────────────
local fileTree = {
    { label = "lumina/", isDir = true, size = "-", children = {
        { label = "cmd/", isDir = true, size = "-", children = {
            { label = "lumina/", isDir = true, size = "-", children = {
                { label = "main.go", size = "1.2K" },
            }},
        }},
        { label = "pkg/", isDir = true, size = "-", children = {
            { label = "lumina/", isDir = true, size = "-", children = {
                { label = "app.go", size = "3.4K" },
                { label = "component.go", size = "8.1K" },
                { label = "hooks.go", size = "5.2K" },
                { label = "layout.go", size = "12.3K" },
                { label = "lumina.go", size = "9.7K" },
                { label = "renderer.go", size = "6.8K" },
                { label = "vdom_diff.go", size = "4.5K" },
            }},
            { label = "components/", isDir = true, size = "-", children = {
                { label = "button.lua", size = "1.5K" },
                { label = "dialog.lua", size = "3.5K" },
                { label = "list.lua", size = "1.4K" },
                { label = "modal.lua", size = "1.8K" },
                { label = "progress.lua", size = "1.2K" },
                { label = "spinner.lua", size = "1.3K" },
                { label = "table.lua", size = "2.2K" },
                { label = "tabs.lua", size = "1.1K" },
                { label = "tree.lua", size = "1.7K" },
            }},
        }},
        { label = "examples/", isDir = true, size = "-", children = {
            { label = "chat_app.lua", size = "4.2K" },
            { label = "counter.lua", size = "0.8K" },
            { label = "dashboard.lua", size = "5.1K" },
            { label = "file_browser.lua", size = "3.9K" },
            { label = "form_demo.lua", size = "1.6K" },
            { label = "todo_mvc.lua", size = "4.8K" },
        }},
        { label = "go.mod", size = "0.3K" },
        { label = "go.sum", size = "2.1K" },
        { label = "README.md", size = "5.4K" },
        { label = "LICENSE", size = "1.1K" },
    }},
}

-- ── Recursive tree renderer ────────────────────────────────────────────────
local function renderTree(nodes, expanded, selectedPath, depth)
    depth = depth or 0
    local children = {}

    for _, node in ipairs(nodes) do
        local isDir = node.isDir or (node.children and #node.children > 0)
        local isExpanded = expanded[node.label]
        local isSelected = (selectedPath == node.label)
        local indent = string.rep("  ", depth)
        local icon = getIcon(node.label, isDir, isExpanded)
        local arrow = ""
        if isDir then
            arrow = isExpanded and "▾ " or "▸ "
        else
            arrow = "  "
        end

        children[#children + 1] = {
            type = "hbox",
            style = {
                height = 1,
                background = isSelected and theme.selected or nil,
            },
            children = {
                { type = "text",
                  content = indent .. arrow .. icon .. " " .. node.label,
                  style = {
                      foreground = isDir and theme.dir or theme.file,
                      bold = isDir,
                      flex = 1,
                  } },
                { type = "text",
                  content = node.size or "",
                  style = { foreground = theme.muted, width = 8 } },
            },
        }

        -- Render children if expanded
        if isDir and isExpanded and node.children then
            local subChildren = renderTree(node.children, expanded, selectedPath, depth + 1)
            for _, child in ipairs(subChildren) do
                children[#children + 1] = child
            end
        end
    end

    return children
end

-- ── Main Component ─────────────────────────────────────────────────────────
local FileBrowser = lumina.defineComponent({
    name = "FileBrowser",

    init = function(props)
        return {
            expanded = {
                ["lumina/"]  = true,
                ["pkg/"]     = true,
                ["lumina/"]  = true,
            },
            selectedPath = "layout.go",
            fileCount = 24,
            dirCount = 7,
        }
    end,

    render = function(instance)
        local treeNodes = renderTree(
            fileTree,
            instance.expanded,
            instance.selectedPath,
            0
        )

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
                          content = " 📁 Lumina File Browser ",
                          style = { foreground = theme.accent, bold = true } },
                        { type = "text",
                          content = "  ~/projects/lumina",
                          style = { foreground = theme.muted } },
                    },
                },

                -- ── Tree View (scrollable) ─────────────────────────────
                {
                    type = "vbox",
                    id = "file-tree",
                    style = {
                        flex = 1,
                        overflow = "scroll",
                        border = "rounded",
                        background = theme.bg,
                    },
                    children = treeNodes,
                },

                -- ── File Preview / Info ────────────────────────────────
                {
                    type = "vbox",
                    style = { height = 3, border = "single", background = theme.headerBg, padding = 0 },
                    children = {
                        { type = "text",
                          content = " File: " .. instance.selectedPath,
                          style = { foreground = theme.fg, bold = true } },
                        { type = "text",
                          content = " Type: Lua source  |  Encoding: UTF-8  |  LF",
                          style = { foreground = theme.muted } },
                    },
                },

                -- ── Status Bar ─────────────────────────────────────────
                {
                    type = "hbox",
                    style = { height = 1, background = theme.headerBg },
                    children = {
                        { type = "text",
                          content = string.format(" %d files, %d dirs",
                              instance.fileCount, instance.dirCount),
                          style = { foreground = theme.muted } },
                        { type = "text",
                          content = " [↑↓] Navigate  [Enter] Open  [Space] Expand  [q] Quit ",
                          style = { foreground = theme.muted, flex = 1 } },
                    },
                },
            },
        }
    end,
})

-- ── Render ─────────────────────────────────────────────────────────────────
lumina.render(FileBrowser, {})
