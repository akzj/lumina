-- Lumina v2 Example: Todo MVC (using Lux component library + theme)
-- Demonstrates: useState, keyboard events, dynamic lists, filtering,
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

lumina.createComponent({
    id = "todo-app",
    name = "TodoMVC",
    x = 0, y = 0,
    w = 80, h = 24,
    zIndex = 0,

    render = function(state, props)
        local t = lumina.getTheme()

        -- State
        local todos, setTodos = lumina.useState("todos", {
            {id=1, text="Learn Lumina v2", done=true, priority="high"},
            {id=2, text="Build a TUI app", done=false, priority="high"},
            {id=3, text="Add keyboard shortcuts", done=false, priority="medium"},
            {id=4, text="Write documentation", done=false, priority="medium"},
            {id=5, text="Deploy to production", done=false, priority="low"},
        })
        local filter, setFilter = lumina.useState("filter", "all")
        local selectedIdx, setSelectedIdx = lumina.useState("selectedIdx", 1)
        local inputText, setInputText = lumina.useState("inputText", "")
        local nextId, setNextId = lumina.useState("nextId", 6)
        local mode, setMode = lumina.useState("mode", "list")

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
                    setMode("list")
                    setInputText("")
                elseif e.key == "Enter" then
                    if #inputText > 0 then
                        local newTodos = {}
                        for _, tt in ipairs(todos) do
                            newTodos[#newTodos + 1] = tt
                        end
                        newTodos[#newTodos + 1] = {
                            id = nextId,
                            text = inputText,
                            done = false,
                            priority = "low",
                        }
                        setTodos(newTodos)
                        setNextId(nextId + 1)
                        setInputText("")
                        setMode("list")
                    end
                elseif e.key == "Backspace" then
                    if #inputText > 0 then
                        local bytePos = utf8.offset(inputText, -1)
                        if bytePos then
                            setInputText(string.sub(inputText, 1, bytePos - 1))
                        else
                            setInputText("")
                        end
                    end
                elseif utf8.len(e.key) == 1 then
                    setInputText(inputText .. e.key)
                end
            else
                -- List mode
                if e.key == "j" or e.key == "ArrowDown" then
                    if selectedIdx < #filtered then
                        setSelectedIdx(selectedIdx + 1)
                    end
                elseif e.key == "k" or e.key == "ArrowUp" then
                    if selectedIdx > 1 then
                        setSelectedIdx(selectedIdx - 1)
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
                        setTodos(newTodos)
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
                        setTodos(newTodos)
                        if selectedIdx > #filtered - 1 and selectedIdx > 1 then
                            setSelectedIdx(selectedIdx - 1)
                        end
                    end
                elseif e.key == "a" or e.key == "i" then
                    setMode("input")
                elseif e.key == "f" then
                    if filter == "all" then
                        setFilter("active")
                    elseif filter == "active" then
                        setFilter("completed")
                    else
                        setFilter("all")
                    end
                    setSelectedIdx(1)
                elseif e.key == "1" then
                    setFilter("all"); setSelectedIdx(1)
                elseif e.key == "2" then
                    setFilter("active"); setSelectedIdx(1)
                elseif e.key == "3" then
                    setFilter("completed"); setSelectedIdx(1)
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
                local checkbox = todo.done and "[x]" or "[ ]"
                local prefix = isSelected and " > " or "   "
                local line = prefix .. checkbox .. " " .. todo.text

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
                        foreground = isSelected and t.primary or (todo.done and t.surface2 or t.text),
                        bold = isSelected,
                    }, line),
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

        -- Input bar
        local inputContent
        if mode == "input" then
            inputContent = " + " .. inputText .. "_"
        else
            inputContent = " + [a] Add new todo"
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

            -- Input bar
            lumina.createElement("text", {
                foreground = mode == "input" and t.text or t.muted,
                style = {background = t.surface0, height = 1},
            }, inputContent),

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
})
