-- Lumina v2 Example: Todo MVC
-- Demonstrates: useState, keyboard events, dynamic lists, filtering
--
-- Usage: lumina-v2 examples/v2/todo_mvc.lua
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

lumina.createComponent({
    id = "todo-app",
    name = "TodoMVC",
    x = 0, y = 0,
    w = 80, h = 24,
    zIndex = 0,

    render = function(state, props)
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
                        for _, t in ipairs(todos) do
                            newTodos[#newTodos + 1] = t
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
                    -- Single printable character (supports Unicode/CJK)
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
                    -- Toggle done
                    if filtered[selectedIdx] then
                        local id = filtered[selectedIdx].id
                        local newTodos = {}
                        for _, t in ipairs(todos) do
                            if t.id == id then
                                newTodos[#newTodos + 1] = {
                                    id = t.id,
                                    text = t.text,
                                    done = not t.done,
                                    priority = t.priority,
                                }
                            else
                                newTodos[#newTodos + 1] = t
                            end
                        end
                        setTodos(newTodos)
                    end
                elseif e.key == "d" then
                    -- Delete selected
                    if filtered[selectedIdx] then
                        local id = filtered[selectedIdx].id
                        local newTodos = {}
                        for _, t in ipairs(todos) do
                            if t.id ~= id then
                                newTodos[#newTodos + 1] = t
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
                    -- Cycle filter
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

        -- Theme (Catppuccin Mocha)
        local theme = {
            bg = "#1E1E2E",
            fg = "#CDD6F4",
            accent = "#89B4FA",
            success = "#A6E3A1",
            error = "#F38BA8",
            warning = "#F9E2AF",
            muted = "#6C7086",
            surface = "#313244",
            headerBg = "#181825",
            done = "#585B70",
        }

        -- Build todo item children
        local todoChildren = {}
        if #filtered == 0 then
            todoChildren[#todoChildren + 1] = lumina.createElement("text", {
                foreground = theme.muted,
            }, "  No todos to show.")
        else
            for i, todo in ipairs(filtered) do
                local isSelected = (i == selectedIdx)
                local checkbox = todo.done and "[x]" or "[ ]"
                local textColor = todo.done and theme.done or theme.fg
                local priorityMark = ""
                if todo.priority == "high" then
                    priorityMark = " !"
                elseif todo.priority == "medium" then
                    priorityMark = " *"
                end

                local prefix = isSelected and " > " or "   "
                local line = prefix .. checkbox .. " " .. todo.text .. priorityMark

                todoChildren[#todoChildren + 1] = lumina.createElement("text", {
                    foreground = isSelected and theme.accent or textColor,
                    bold = isSelected,
                    style = {background = isSelected and theme.surface or theme.bg, height = 1},
                }, line)
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
        local inputFg = mode == "input" and theme.fg or theme.muted

        -- Build the VNode tree using a raw table for the todo list section
        -- (since table.unpack may not work well with many children)
        local todoList = {
            type = "vbox",
            id = "todo-list",
            style = {background = theme.bg},
            children = todoChildren,
        }

        return lumina.createElement("vbox", {
            id = "todo-root",
            style = {background = theme.bg},
            onKeyDown = handleKey,
            focusable = true,
        },
            -- Header
            lumina.createElement("text", {
                foreground = theme.accent,
                bold = true,
                style = {background = theme.headerBg, height = 1},
            }, " Todo MVC  " .. #todos .. " items, " .. activeCount .. " active"),

            -- Input bar
            lumina.createElement("text", {
                foreground = inputFg,
                style = {background = theme.surface, height = 1},
            }, inputContent),

            -- Filter bar
            lumina.createElement("text", {
                foreground = theme.muted,
                style = {background = theme.surface, height = 1},
            }, filterText),

            -- Todo list (raw table, not createElement, for dynamic children)
            todoList,

            -- Footer
            lumina.createElement("text", {
                foreground = theme.muted,
                style = {background = theme.headerBg, height = 1},
            }, " [j/k] Navigate  [Space] Toggle  [d] Delete  [a] Add  [f/1-3] Filter")
        )
    end,
})
