-- examples/notes.lua
-- A simple Notes App showcasing Lumina's syntax sugar.
--
-- Demonstrates:
--   • lumina.app { ... } entry point with store + keys + render
--   • __call syntax for Go widgets: lumina.List { ... }, lumina.Button { ... }
--   • require("lux") components: Card, Badge, Divider
--   • lumina.defineComponent for custom components
--   • lumina.useStore for state management
--   • lumina.getTheme() for themed colors
--   • Global keybindings (q=quit, n=new, d=delete)
--
-- Usage: lumina examples/notes.lua

local lux = require("lux")
local Card = lux.Card
local Badge = lux.Badge
local Divider = lux.Divider

-- Custom component: NotePreview (shows title + first line of content)
local NotePreview = lumina.defineComponent("NotePreview", function(props)
    local t = lumina.getTheme()
    local note = props.note or {}
    local selected = props.selected
    local fg = selected and t.primary or t.text
    local prefix = selected and "▸ " or "  "
    return lumina.createElement("vbox", {
        style = { height = 2, background = selected and t.surface0 or t.base },
    },
        lumina.createElement("text", { foreground = fg, bold = selected },
            prefix .. (note.title or "Untitled")),
        lumina.createElement("text", { foreground = t.muted, dim = true },
            "    " .. string.sub(note.content or "", 1, 50))
    )
end)

lumina.app {
    id = "notes-app",
    store = {
        notes = {
            { id = 1, title = "Welcome to Lumina", content = "This is a TUI framework written in Go + Lua." },
            { id = 2, title = "Syntax Sugar", content = "Use lumina.Button { ... } instead of createElement." },
            { id = 3, title = "Lux Components", content = "Card, Badge, Divider, Layout and more." },
        },
        selectedIdx = 1,
        nextId = 4,
        viewing = false,
    },
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
        ["n"] = function()
            local notes = lumina.useStore("notes")
            local nextId = lumina.useStore("nextId")
            local newNotes = {}
            for _, n in ipairs(notes) do newNotes[#newNotes + 1] = n end
            newNotes[#newNotes + 1] = {
                id = nextId,
                title = "Note #" .. nextId,
                content = "New note created. Edit me!",
            }
            lumina.store.set("notes", newNotes)
            lumina.store.set("nextId", nextId + 1)
            lumina.store.set("selectedIdx", #newNotes)
        end,
        ["d"] = function()
            local notes = lumina.useStore("notes")
            local idx = lumina.useStore("selectedIdx")
            if #notes == 0 then return end
            local newNotes = {}
            for i, n in ipairs(notes) do
                if i ~= idx then newNotes[#newNotes + 1] = n end
            end
            lumina.store.set("notes", newNotes)
            if idx > #newNotes and #newNotes > 0 then
                lumina.store.set("selectedIdx", #newNotes)
            end
            lumina.store.set("viewing", false)
        end,
        ["j"] = function()
            local notes = lumina.useStore("notes")
            local idx = lumina.useStore("selectedIdx")
            if idx < #notes then lumina.store.set("selectedIdx", idx + 1) end
        end,
        ["k"] = function()
            local idx = lumina.useStore("selectedIdx")
            if idx > 1 then lumina.store.set("selectedIdx", idx - 1) end
        end,
        ["Enter"] = function()
            local viewing = lumina.useStore("viewing")
            lumina.store.set("viewing", not viewing)
        end,
        ["Escape"] = function()
            lumina.store.set("viewing", false)
        end,
    },

    render = function()
        local t = lumina.getTheme()
        local notes = lumina.useStore("notes")
        local selectedIdx = lumina.useStore("selectedIdx")
        local viewing = lumina.useStore("viewing")

        -- Header with note count badge
        local header = lumina.createElement("hbox", { style = { height = 1 } },
            lumina.createElement("text", { foreground = t.primary, bold = true }, " 📝 Notes "),
            Badge { label = tostring(#notes), variant = "default", key = "count-badge" }
        )

        -- Build note list items
        local listItems = {}
        for i, note in ipairs(notes) do
            listItems[#listItems + 1] = NotePreview {
                note = note,
                selected = (i == selectedIdx),
                key = "note-" .. note.id,
            }
        end

        -- Note list or empty state
        local noteList
        if #notes == 0 then
            noteList = lumina.createElement("text", { foreground = t.muted },
                "  No notes yet. Press [n] to create one.")
        else
            noteList = lumina.createElement("vbox", {
                style = { flex = 1 },
            }, table.unpack(listItems))
        end

        -- Detail view (shown when Enter is pressed)
        local detail
        if viewing and notes[selectedIdx] then
            local note = notes[selectedIdx]
            detail = Card {
                title = note.title,
                key = "detail-card",
                lumina.createElement("text", { foreground = t.text }, note.content),
                lumina.createElement("text", { foreground = t.muted, dim = true },
                    "\n[Escape] Back"),
            }
        end

        -- Footer with keybindings
        local footer = lumina.createElement("text", {
            foreground = t.muted, dim = true,
            style = { height = 1, background = t.surface0 },
        }, " [j/k] Navigate  [Enter] View  [n] New  [d] Delete  [q] Quit")

        -- Main layout
        return lumina.createElement("vbox", {
            id = "notes-root",
            style = { background = t.base, border = "rounded" },
            focusable = true,
        },
            header,
            Divider { width = 60, key = "div-top" },
            detail or noteList,
            Divider { width = 60, key = "div-bottom" },
            footer
        )
    end,
}
