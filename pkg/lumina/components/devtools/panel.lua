-- DevTools Inspector Panel (Lua implementation)
-- Replaces the Go-side BuildInspectorVNode with framework-native rendering.
-- Called by Go via _devtools_render() when inspector state changes.
local lumina = require("lumina")

-- Colors
local c = {
    bg      = "#111827",
    fg      = "#D1D5DB",
    header  = "#1E40AF",
    section = "#93C5FD",
    selected = "#FCD34D",
    highlight = "#34D399",
    muted   = "#6B7280",
}

-- Build element tree lines from structured VNode data
local function buildTreeLines(node, depth, lines, panelW)
    if not node then return end
    depth = depth or 0
    lines = lines or {}
    
    local indent = string.rep("  ", depth)
    local prefix = depth == 0 and "▸" or "├─"
    
    local label = indent .. prefix .. " " .. (node.type or "?")
    if node.id and node.id ~= "" then
        label = label .. "#" .. node.id
    end
    if (node.w or 0) > 0 and (node.h or 0) > 0 then
        label = label .. string.format(" (%dx%d)", node.w, node.h)
    end
    
    -- Truncate to fit panel
    local maxLen = (panelW or 40) - 4
    if maxLen > 0 and #label > maxLen then
        label = string.sub(label, 1, maxLen - 1) .. "…"
    end
    
    local nodeID = node.id
    if not nodeID or nodeID == "" then
        nodeID = string.format("%s-%d-%d", node.type or "?", node.x or 0, node.y or 0)
    end
    
    table.insert(lines, { text = " " .. label, id = nodeID })
    
    if node.children then
        for _, child in ipairs(node.children) do
            buildTreeLines(child, depth + 1, lines, panelW)
        end
    end
    
    return lines
end

-- Build style info lines
local function buildStyleLines(styles, panelW)
    local lines = {}
    if not styles or not styles.selected then
        table.insert(lines, " (hover or click to inspect)")
        return lines
    end
    
    table.insert(lines, string.format(" Element: %s", styles.type or "?"))
    if styles.id and styles.id ~= "" then
        table.insert(lines, string.format(" ID: %s", styles.id))
    end
    table.insert(lines, string.format(" Position: (%d, %d)", styles.x or 0, styles.y or 0))
    table.insert(lines, string.format(" Size: %d × %d", styles.w or 0, styles.h or 0))
    
    if styles.content then
        local content = styles.content
        if #content > 30 then
            content = string.sub(content, 1, 27) .. "..."
        end
        table.insert(lines, string.format(' Content: "%s"', content))
    end
    
    table.insert(lines, " ─── Style ───")
    if styles.style then
        for k, v in pairs(styles.style) do
            table.insert(lines, string.format("  %s: %s", k, tostring(v)))
        end
    end
    
    -- Truncate lines
    local maxLen = (panelW or 40) - 3
    if maxLen > 0 then
        for i, line in ipairs(lines) do
            if #line > maxLen then
                lines[i] = string.sub(line, 1, maxLen - 1) .. "…"
            end
        end
    end
    
    return lines
end

-- Main render function called by Go
function _devtools_render()
    local devtools = lumina.devtools
    if not devtools.isInspectorEnabled() then
        lumina.hideOverlay("devtools-panel")
        return
    end
    
    local w, h = lumina.getSize()
    local panelW = devtools.getPanelWidth()
    if panelW > math.floor(w / 2) then
        panelW = math.floor(w / 2)
    end
    
    local selectedID = devtools.getSelectedID()
    local highlightID = devtools.getHighlightID()
    local scrollY = devtools.getScrollY()
    
    -- Get data from Go
    local tree = devtools.getElementTree()
    local styles = devtools.getSelectedStyles()
    
    -- Build tree lines
    local treeLines = buildTreeLines(tree, 0, nil, panelW)
    
    -- Build children VNodes
    local children = {}
    
    -- Header
    table.insert(children, {
        type = "text",
        content = " [*] DevTools Inspector ",
        style = {
            bold = true,
            foreground = "#FFFFFF",
            background = c.header,
            height = 1,
        },
    })
    
    -- Element tree header
    table.insert(children, {
        type = "text",
        content = " ─── Element Tree ───",
        style = { foreground = c.section, height = 1 },
    })
    
    -- Element tree lines (scrollable)
    local maxTreeLines = math.floor(h / 2) - 4
    local startLine = scrollY
    if startLine > #treeLines - maxTreeLines then
        startLine = #treeLines - maxTreeLines
    end
    if startLine < 0 then startLine = 0 end
    local endLine = startLine + maxTreeLines
    if endLine > #treeLines then endLine = #treeLines end
    
    for i = startLine + 1, endLine do
        local line = treeLines[i]
        if line then
            local fg = c.fg
            if line.id == selectedID then
                fg = c.selected
            elseif line.id == highlightID then
                fg = c.highlight
            end
            table.insert(children, {
                type = "text",
                id = "dt-tree-" .. line.id,
                content = line.text,
                style = { foreground = fg, height = 1 },
            })
        end
    end
    
    -- Style inspector header
    table.insert(children, {
        type = "text",
        content = " ─── Computed Styles ───",
        style = { foreground = c.section, height = 1 },
    })
    
    -- Style lines
    local styleLines = buildStyleLines(styles, panelW)
    local maxStyleLines = math.floor(h / 2) - 2
    for i = 1, math.min(#styleLines, maxStyleLines) do
        table.insert(children, {
            type = "text",
            content = styleLines[i],
            style = { foreground = c.fg, height = 1 },
        })
    end
    
    -- Close hint
    table.insert(children, {
        type = "text",
        content = " [F12] Close",
        style = { foreground = c.muted, height = 1 },
    })
    
    -- Show as overlay
    lumina.showOverlay({
        id = "devtools-panel",
        x = w - panelW,
        y = 0,
        width = panelW,
        height = h,
        zIndex = 9999,
        content = {
            type = "vbox",
            style = {
                width = panelW,
                height = h,
                background = c.bg,
                foreground = c.fg,
                border = "single",
            },
            children = children,
        },
    })
end
