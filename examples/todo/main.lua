-- ============================================================================
-- Lumina Example: Todo App
-- ============================================================================
-- Showcases: useState, createStore, keyboard shortcuts, filtering,
--            conditional rendering, list navigation
--
-- Run: go run ./cmd/lumina examples/todo/main.lua
-- ============================================================================

local lumina = require("lumina")

-- ── Store ───────────────────────────────────────────────────────────────────
local store = lumina.createStore({
    state = {
        todos = {
            { id = 1, text = "Learn Lumina framework",   done = true,  priority = "high" },
            { id = 2, text = "Build a TUI app",          done = false, priority = "high" },
            { id = 3, text = "Add keyboard shortcuts",   done = false, priority = "medium" },
            { id = 4, text = "Deploy to production",     done = false, priority = "low" },
            { id = 5, text = "Write documentation",      done = false, priority = "medium" },
        },
        selected = 1,
        filter = "all",  -- "all" | "active" | "completed"
        nextId = 6,
    },
})

-- ── Theme ───────────────────────────────────────────────────────────────────
local colors = {
    bg       = "#1E1E2E",
    fg       = "#CDD6F4",
    accent   = "#89B4FA",
    success  = "#A6E3A1",
    error    = "#F38BA8",
    warning  = "#F9E2AF",
    muted    = "#6C7086",
    surface  = "#313244",
    done     = "#585B70",
    header   = "#181825",
}

-- ── Helpers ─────────────────────────────────────────────────────────────────
local function getFilteredTodos(state)
    local todos = state.todos or {}
    local filter = state.filter or "all"
    if filter == "all" then return todos end
    local result = {}
    for _, t in ipairs(todos) do
        if filter == "active" and not t.done then
            result[#result + 1] = t
        elseif filter == "completed" and t.done then
            result[#result + 1] = t
        end
    end
    return result
end

local function countTodos(todos)
    local active, completed = 0, 0
    for _, t in ipairs(todos) do
        if t.done then completed = completed + 1
        else active = active + 1 end
    end
    return { all = #todos, active = active, completed = completed }
end

-- ── Header Component ────────────────────────────────────────────────────────
local function Header()
    local state = lumina.useStore(store)
    local counts = countTodos(state.todos or {})
    return {
        type = "vbox",
        children = {
            { type = "text", content = "╔══════════════════════════════════════╗", style = { foreground = colors.accent } },
            { type = "text", content = "║       ✅ Lumina Todo App             ║", style = { foreground = colors.accent, bold = true } },
            { type = "text", content = "╚══════════════════════════════════════╝", style = { foreground = colors.accent } },
            { type = "text", content = string.format("  %d total · %d active · %d done", counts.all, counts.active, counts.completed),
              style = { foreground = colors.muted } },
            { type = "text", content = "" },
        },
    }
end

-- ── Filter Bar ──────────────────────────────────────────────────────────────
local function FilterBar()
    local state = lumina.useStore(store)
    local filter = state.filter or "all"
    local counts = countTodos(state.todos or {})

    local filters = {
        { key = "all",       label = "All" },
        { key = "active",    label = "Active" },
        { key = "completed", label = "Completed" },
    }

    local children = {}
    for _, f in ipairs(filters) do
        local isActive = (f.key == filter)
        local count = counts[f.key] or 0
        children[#children + 1] = {
            type = "text",
            content = isActive
                and string.format(" [ %s (%d) ] ", f.label, count)
                or  string.format("   %s (%d)   ", f.label, count),
            style = {
                foreground = isActive and colors.accent or colors.muted,
                bold = isActive,
            },
        }
    end

    return { type = "hbox", children = children }
end

-- ── Todo Item ───────────────────────────────────────────────────────────────
local function TodoItem(props)
    local todo = props.todo
    local isSelected = props.isSelected

    local checkbox = todo.done and "[✓]" or "[ ]"
    local textColor = todo.done and colors.done or colors.fg
    local checkColor = todo.done and colors.success or colors.muted

    local priorityIndicator = ""
    local priorityColor = colors.muted
    if todo.priority == "high" then
        priorityIndicator = " ●"
        priorityColor = colors.error
    elseif todo.priority == "medium" then
        priorityIndicator = " ●"
        priorityColor = colors.warning
    end

    return {
        type = "hbox",
        style = {
            height = 1,
            background = isSelected and colors.surface or nil,
        },
        children = {
            { type = "text", content = isSelected and " ▸ " or "   ", style = { foreground = colors.accent } },
            { type = "text", content = checkbox .. " ", style = { foreground = checkColor, bold = todo.done } },
            { type = "text", content = todo.text, style = { foreground = textColor, dim = todo.done } },
            priorityIndicator ~= "" and
                { type = "text", content = priorityIndicator, style = { foreground = priorityColor } }
                or nil,
        },
    }
end

-- ── Todo List ───────────────────────────────────────────────────────────────
local function TodoList()
    local state = lumina.useStore(store)
    local filtered = getFilteredTodos(state)
    local selected = state.selected or 1

    if #filtered == 0 then
        return {
            type = "vbox",
            children = {
                { type = "text", content = "" },
                { type = "text", content = "  No todos to show.", style = { foreground = colors.muted, italic = true } },
            },
        }
    end

    local children = {}
    for i, todo in ipairs(filtered) do
        children[#children + 1] = TodoItem({ todo = todo, isSelected = (i == selected) })
    end

    return { type = "vbox", children = children }
end

-- ── Footer ──────────────────────────────────────────────────────────────────
local function Footer()
    return {
        type = "vbox",
        children = {
            { type = "text", content = "" },
            { type = "text", content = string.rep("─", 40), style = { foreground = "#45475A" } },
            { type = "text", content = " [a]dd  [d]elete  [space]toggle  [tab]filter  [q]uit",
              style = { foreground = colors.muted, dim = true } },
            { type = "text", content = " [j]↓   [k]↑      [1-3]priority",
              style = { foreground = colors.muted, dim = true } },
        },
    }
end

-- ── Main App ───────────────────────────────────────────────────────────────
local App = lumina.defineComponent({
    name = "TodoApp",
    render = function(self)
        return {
            type = "vbox",
            style = { background = colors.bg },
            children = {
                Header(),
                FilterBar(),
                { type = "text", content = "" },
                TodoList(),
                Footer(),
            },
        }
    end,
})

-- ── Key bindings ───────────────────────────────────────────────────────────
lumina.onKey("j", function()
    local state = store.getState()
    local filtered = getFilteredTodos(state)
    local sel = math.min((state.selected or 1) + 1, #filtered)
    store.dispatch("setState", { selected = sel })
end)

lumina.onKey("k", function()
    local state = store.getState()
    local sel = math.max((state.selected or 1) - 1, 1)
    store.dispatch("setState", { selected = sel })
end)

lumina.onKey(" ", function()
    local state = store.getState()
    local filtered = getFilteredTodos(state)
    local sel = state.selected or 1
    if filtered[sel] then
        local todos = state.todos
        for i, t in ipairs(todos) do
            if t.id == filtered[sel].id then
                todos[i].done = not todos[i].done
                break
            end
        end
        store.dispatch("setState", { todos = todos })
    end
end)

lumina.onKey("a", function()
    local state = store.getState()
    local todos = state.todos or {}
    local nextId = state.nextId or (#todos + 1)
    todos[#todos + 1] = { id = nextId, text = "New todo #" .. nextId, done = false, priority = "medium" }
    store.dispatch("setState", { todos = todos, nextId = nextId + 1 })
end)

lumina.onKey("d", function()
    local state = store.getState()
    local filtered = getFilteredTodos(state)
    local sel = state.selected or 1
    if filtered[sel] then
        local deleteId = filtered[sel].id
        local todos = {}
        for _, t in ipairs(state.todos) do
            if t.id ~= deleteId then todos[#todos + 1] = t end
        end
        store.dispatch("setState", { todos = todos, selected = math.max(1, sel - 1) })
    end
end)

lumina.onKey("tab", function()
    local state = store.getState()
    local order = { all = "active", active = "completed", completed = "all" }
    local newFilter = order[state.filter or "all"] or "all"
    store.dispatch("setState", { filter = newFilter, selected = 1 })
end)

lumina.onKey("q", function() lumina.quit() end)

lumina.mount(App)
lumina.run()
