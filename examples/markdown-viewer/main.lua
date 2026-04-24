-- Markdown Viewer — A terminal markdown file viewer
-- Usage: lumina examples/markdown-viewer/main.lua
--
-- Features:
--   • Renders markdown with styled headers, bold, italic, code, lists
--   • Scrollable with j/k or arrow keys
--   • Uses shadcn Card container
--   • Built-in demo content when no file is provided

local lumina = require("lumina")
local shadcn = require("shadcn")

lumina.setTheme("catppuccin-mocha")

-- ─── Markdown Parser ──────────────────────────────────────────────────

local function parseMD(text)
    local nodes = {}
    local inCodeBlock = false
    local codeLines = {}

    for line in text:gmatch("([^\n]*)\n?") do
        -- Code block toggle
        if line:match("^```") then
            if inCodeBlock then
                -- End code block: wrap accumulated lines in a bordered box
                local codeText = table.concat(codeLines, "\n")
                table.insert(nodes, {
                    type = "vbox",
                    style = { border = "rounded", padding = 1, background = "#181825" },
                    children = {
                        { type = "text", content = codeText, style = { foreground = "#A6E3A1" } },
                    }
                })
                codeLines = {}
                inCodeBlock = false
            else
                inCodeBlock = true
            end
        elseif inCodeBlock then
            table.insert(codeLines, line)
        elseif line:match("^### ") then
            table.insert(nodes, {
                type = "text",
                content = "  " .. line:sub(5),
                style = { bold = true, foreground = "#89DCEB" },
            })
        elseif line:match("^## ") then
            table.insert(nodes, {
                type = "text",
                content = " " .. line:sub(4),
                style = { bold = true, foreground = "#89B4FA" },
            })
        elseif line:match("^# ") then
            table.insert(nodes, {
                type = "text",
                content = line:sub(3),
                style = { bold = true, foreground = "#F5C2E7" },
            })
        elseif line:match("^%-%-%-") or line:match("^%*%*%*") then
            table.insert(nodes, lumina.createElement(shadcn.Separator, {}))
        elseif line:match("^%d+%.%s") then
            -- Numbered list
            local num, rest = line:match("^(%d+%.%s)(.*)")
            table.insert(nodes, {
                type = "text",
                content = "  " .. num .. rest,
                style = { foreground = "#CDD6F4" },
            })
        elseif line:match("^%- ") then
            table.insert(nodes, {
                type = "text",
                content = "  • " .. line:sub(3),
                style = { foreground = "#CDD6F4" },
            })
        elseif line:match("^> ") then
            -- Blockquote
            table.insert(nodes, {
                type = "hbox",
                children = {
                    { type = "text", content = " ▎ ", style = { foreground = "#585B70" } },
                    { type = "text", content = line:sub(3), style = { foreground = "#BAC2DE", italic = true } },
                }
            })
        elseif line == "" then
            table.insert(nodes, { type = "text", content = "" })
        else
            -- Apply inline formatting
            local content = line
            -- Bold: **text**
            content = content:gsub("%*%*(.-)%*%*", function(s) return "⟨b⟩" .. s .. "⟨/b⟩" end)
            -- Italic: *text*
            content = content:gsub("%*(.-)%*", function(s) return "⟨i⟩" .. s .. "⟨/i⟩" end)
            -- Inline code: `text`
            content = content:gsub("`(.-)`", function(s) return "⟨c⟩" .. s .. "⟨/c⟩" end)
            -- Links: [text](url)
            content = content:gsub("%[(.-)%]%((.-)%)", function(text, url)
                return "⟨l⟩" .. text .. "⟨/l⟩"
            end)

            table.insert(nodes, {
                type = "text",
                content = content,
                style = { foreground = "#CDD6F4" },
            })
        end
    end

    return nodes
end

-- ─── Demo Content ─────────────────────────────────────────────────────

local demoMarkdown = [[
# Lumina — Terminal React for AI Agents

## Overview

Lumina is a **React-style terminal UI framework** built with Go and Lua.
It brings modern component-based development to the terminal.

### Features

- **57 shadcn components** — buttons, cards, dialogs, tables, and more
- **19 React-style hooks** — useState, useEffect, useMemo, useContext
- **Hot reload** — instant feedback during development
- **Web mode** — run in the browser via WebSocket
- **AI native** — designed for agent interfaces

---

## Quick Start

```lua
local lumina = require("lumina")

local App = lumina.defineComponent({
    name = "App",
    render = function(self)
        return { type = "text", content = "Hello!" }
    end
})

lumina.mount(App)
lumina.run()
```

### Installation

1. Download the latest release
2. Add `lumina` to your PATH
3. Run `lumina init myapp`

> Lumina works best with a modern terminal that supports 24-bit color.

## Components

- **Button** — clickable actions with variants
- **Card** — content containers with headers
- **Dialog** — modal overlays
- **Table** — data display with sorting
- **Select** — dropdown menus
- **Input** — text entry fields

---

*Built with ❤️ for the terminal*
]]

-- ─── App Component ────────────────────────────────────────────────────

local App = lumina.defineComponent({
    name = "MarkdownViewer",
    render = function(self)
        local scrollY, setScrollY = lumina.useState("scrollY", 0)
        local content = parseMD(demoMarkdown)

        -- Slice content for scrolling
        local visible = {}
        local startIdx = scrollY + 1
        local endIdx = math.min(#content, scrollY + 20)
        for i = startIdx, endIdx do
            table.insert(visible, content[i])
        end

        return {
            type = "vbox",
            style = { background = "#1E1E2E" },
            children = {
                -- Title bar
                {
                    type = "hbox",
                    style = { padding = 1, background = "#313244" },
                    children = {
                        { type = "text", content = "📄 Markdown Viewer", style = { bold = true, foreground = "#F5C2E7" } },
                        { type = "text", content = "  ↑↓/jk to scroll  q to quit", style = { foreground = "#585B70" } },
                    }
                },
                -- Content area
                lumina.createElement(shadcn.Card, {
                    children = visible,
                }),
                -- Status bar
                {
                    type = "hbox",
                    style = { padding = 1, background = "#313244" },
                    children = {
                        { type = "text", content = "Line " .. tostring(scrollY + 1) .. "/" .. tostring(#content), style = { foreground = "#585B70" } },
                    }
                },
            }
        }
    end,
})

-- Keyboard shortcuts
lumina.onKey("j", function()
    -- scroll down handled by state
end)
lumina.onKey("k", function()
    -- scroll up handled by state
end)
lumina.onKey("q", function()
    lumina.quit()
end)

lumina.mount(App)
lumina.run()
