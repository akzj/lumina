-- Lumina v2 Example: Todo MVC (using lumina.app + Lux component library + theme)
-- Demonstrates: lumina.app, useStore, keyboard events, dynamic lists, filtering,
--               Lux Badge/Divider components, lumina.getTheme()
--
-- Usage: lumina examples/todo_mvc.lua
-- Quit:  Ctrl+C or Ctrl+Q
--
-- Keyboard:
--   j/k or Arrow keys  - Navigate todos
--   Space              - Toggle done
--   d                  - Delete todo
--   a/i                - Add new todo (enter input mode)
--   Enter              - Submit new todo (in input mode)
--   Escape             - Cancel input mode
--   f or 1/2/3         - Cycle/select filter (All/Active/Completed)

local lux = require("lux")
local Badge = lux.Badge
local Divider = lux.Divider

-- Module-level variable to track pending input text.
-- Avoids stale closure issues: onChange writes here, onSubmit reads from here.
local _pendingInput = ""

lumina.app {
    id = "todo-app",
    store = {
        todos = {
            {id=1, text="Learn Lumina v2", done=true, priority="high"},
            {id=2, text="Build a TUI app", done=false, priority="high"},
            {id=3, text="Add keyboard shortcuts", done=false, priority="medium"},
            {id=4, text="Write documentation", done=false, priority="medium"},
            {id=5, text="Deploy to production", done=false, priority="low"},
        },
        filter = "all",
        selectedIdx = 1,
        inputText = "",
        nextId = 6,
        mode = "list",
    },
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
    },

    render = function()
        local t = lumina.getTheme()

        -- State from store
        local todos = lumina.useStore("todos")
        local filter = lumina.useStore("filter")
        local selectedIdx = lumina.useStore("selectedIdx")
        local mode = lumina.useStore("mode")
        local nextId = lumina.useStore("nextId")

        -- Computed: filtered todos and counts
        local filtered = {}
        local activeCount, completedCount = 0, 0
        for _, todo in ipairs(todos) do
            if todo.done then
                completedCount = completedCount + 1
            else
                activeCount = activeCount + 1
            end
            if filter == "all"
               or (filter == "active" and not todo.done)
               or (filter == "completed" and todo.done) then
                filtered[#filtered + 1] = todo
            end
        end

        -- Clamp selectedIdx to valid range
        if selectedIdx > #filtered and #filtered > 0 then
            selectedIdx = #filtered
        end
        if selectedIdx < 1 then
            selectedIdx = 1
        end

        -- Keyboard handler
        local function handleKey(e)
            if mode == "input" then
                if e.key == "Escape" then
                    lumina.store.set("mode", "list")
                    lumina.store.set("inputText", "")
                    _pendingInput = ""
                end
                -- All other keys (chars, backspace, arrows) handled by native input.
                -- Enter is handled by onSubmit on the input element.
                return
            else
                -- List mode
                if e.key == "j" or e.key == "ArrowDown" then
                    if selectedIdx < #filtered then
                        lumina.store.set("selectedIdx", selectedIdx + 1)
                    end
                elseif e.key == "k" or e.key == "ArrowUp" then
                    if selectedIdx > 1 then
                        lumina.store.set("selectedIdx", selectedIdx - 1)
                    end
                elseif e.key == " " then
                    if filtered[selectedIdx] then
                        local id = filtered[selectedIdx].id
                        local newTodos = {}
                        for _, tt in ipairs(todos) do
                            if tt.id == id then
                                newTodos[#newTodos + 1] = {
                                    id = tt.id,
                                    text = tt.text,
                                    done = not tt.done,
                                    priority = tt.priority,
                                }
                            else
                                newTodos[#newTodos + 1] = tt
                            end
                        end
                        lumina.store.set("todos", newTodos)
                    end
                elseif e.key == "d" then
                    if filtered[selectedIdx] then
                        local id = filtered[selectedIdx].id
                        local newTodos = {}
                        for _, tt in ipairs(todos) do
                            if tt.id ~= id then
                                newTodos[#newTodos + 1] = tt
                            end
                        end
                        lumina.store.set("todos", newTodos)
                        if selectedIdx > #filtered - 1 and selectedIdx > 1 then
                            lumina.store.set("selectedIdx", selectedIdx - 1)
                        end
                    end
                elseif e.key == "a" or e.key == "i" then
                    lumina.store.set("mode", "input")
                elseif e.key == "f" then
                    if filter == "all" then
                        lumina.store.set("filter", "active")
                    elseif filter == "active" then
                        lumina.store.set("filter", "completed")
                    else
                        lumina.store.set("filter", "all")
                    end
                    lumina.store.set("selectedIdx", 1)
                elseif e.key == "1" then
                    lumina.store.batch({filter = "all", selectedIdx = 1})
                elseif e.key == "2" then
                    lumina.store.batch({filter = "active", selectedIdx = 1})
                elseif e.key == "3" then
                    lumina.store.batch({filter = "completed", selectedIdx = 1})
                end
            end
        end

        -- Build todo item children
        local todoChildren = {}
        if #filtered == 0 then
            todoChildren[#todoChildren + 1] = lumina.createElement("text", {
                foreground = t.muted,
            }, "  No todos to show.")
        else
            for i, todo in ipairs(filtered) do
                local isSelected = (i == selectedIdx)
                local prefix = isSelected and " > " or "   "

                -- Priority badge using Lux Badge component
                local priorityBadge = nil
                if todo.priority == "high" then
                    priorityBadge = lumina.createElement(Badge, {
                        label = "!",
                        variant = "error",
                    })
                elseif todo.priority == "medium" then
                    priorityBadge = lumina.createElement(Badge, {
                        label = "*",
                        variant = "warning",
                    })
                end

                local rowChildren = {
                    lumina.createElement("text", {
                        foreground = isSelected and t.primary or t.text,
                        bold = isSelected,
                    }, prefix),
                    lumina.createElement(lumina.Checkbox, {
                        key = "cb-" .. tostring(todo.id),
                        label = todo.text,
                        checked = todo.done,
                        onChange = function(newChecked)
                            local id = todo.id
                            local newTodos = {}
                            for _, tt in ipairs(todos) do
                                if tt.id == id then
                                    newTodos[#newTodos + 1] = {
                                        id = tt.id,
                                        text = tt.text,
                                        done = newChecked,
                                        priority = tt.priority,
                                    }
                                else
                                    newTodos[#newTodos + 1] = tt
                                end
                            end
                            lumina.store.set("todos", newTodos)
                        end,
                    }),
                }
                if priorityBadge then
                    rowChildren[#rowChildren + 1] = priorityBadge
                end

                todoChildren[#todoChildren + 1] = lumina.createElement("hbox", {
                    style = {
                        background = isSelected and t.surface0 or t.base,
                        height = 1,
                    },
                }, table.unpack(rowChildren))
            end
        end

        -- Filter bar text
        local filterLabels = {
            {key = "all", label = "All", count = #todos},
            {key = "active", label = "Active", count = activeCount},
            {key = "completed", label = "Completed", count = completedCount},
        }
        local filterText = ""
        for _, f in ipairs(filterLabels) do
            if f.key == filter then
                filterText = filterText .. " [" .. f.label .. "(" .. f.count .. ")] "
            else
                filterText = filterText .. "  " .. f.label .. "(" .. f.count .. ")  "
            end
        end

        -- Input bar: native input element (handles CJK, cursor movement)
        -- or placeholder text when not in input mode
        local inputBar
        if mode == "input" then
            inputBar = lumina.createElement("input", {
                id = "todo-input",
                style = {width = 76, height = 1, background = t.surface0},
                foreground = t.text,
                placeholder = "Type a new todo and press Enter...",
                autoFocus = true,
                focusable = true,
                onChange = function(text)
                    _pendingInput = text
                end,
                onSubmit = function()
                    local text = _pendingInput
                    if #text > 0 then
                        local newTodos = {}
                        for _, tt in ipairs(todos) do
                            newTodos[#newTodos + 1] = tt
                        end
                        newTodos[#newTodos + 1] = {
                            id = nextId,
                            text = text,
                            done = false,
                            priority = "low",
                        }
                        lumina.store.set("todos", newTodos)
                        lumina.store.set("nextId", nextId + 1)
                        lumina.store.set("inputText", "")
                        _pendingInput = ""
                        lumina.store.set("mode", "list")
                    end
                end,
            })
        else
            inputBar = lumina.createElement("text", {
                foreground = t.muted,
                style = {background = t.surface0, height = 1},
            }, " + [a] Add new todo")
        end

        -- Completion progress bar
        local totalCount = #todos
        local progressPct = totalCount > 0 and math.floor(completedCount * 100 / totalCount) or 0
        local barWidth = 20
        local filled = math.floor(barWidth * progressPct / 100)
        local progressBar = string.rep("█", filled) .. string.rep("░", barWidth - filled)

        -- Build the todo list section (raw table for dynamic children)
        local todoList = {
            type = "vbox",
            id = "todo-list",
            style = {background = t.base},
            children = todoChildren,
        }

        return lumina.createElement("vbox", {
            id = "todo-root",
            style = {background = t.base, border = "rounded"},
            onKeyDown = handleKey,
            focusable = true,
        },
            -- Header
            lumina.createElement("text", {
                foreground = t.primary,
                bold = true,
                style = {background = t.surface0, height = 1},
            }, " Todo MVC  " .. totalCount .. " items, " .. activeCount .. " active"),

            -- Input bar (native input in input mode, or placeholder text)
            inputBar,

            -- Divider (Lux component)
            lumina.createElement(Divider, {width = 78}),

            -- Filter bar
            lumina.createElement("text", {
                foreground = t.muted,
                style = {background = t.surface0, height = 1},
            }, filterText),

            -- Todo list
            todoList,

            -- Divider before progress
            lumina.createElement(Divider, {width = 78}),

            -- Progress bar
            lumina.createElement("hbox", {
                style = {height = 1},
            },
                lumina.createElement("text", {
                    foreground = t.muted,
                    dim = true,
                }, " Progress: "),
                lumina.createElement("text", {
                    foreground = t.primary,
                }, progressBar),
                lumina.createElement("text", {
                    foreground = t.text,
                }, " " .. progressPct .. "%")
            ),

            -- Footer
            lumina.createElement("text", {
                foreground = t.muted,
                dim = true,
                style = {background = t.surface0, height = 1},
            }, " [j/k] Navigate  [Space] Toggle  [d] Delete  [a] Add  [f/1-3] Filter")
        )
    end,
}
