-- Todo App: Demonstrates Lumina with stateful components
local lumina = require("lumina")

print("=== Lumina Todo App v0.2 ===")
print()

-- Todo item helper
local function makeTodo(id, text, done)
    return { id = id, text = text, done = done }
end

-- Todo App component
local TodoApp = lumina.defineComponent({
    name = "TodoApp",
    
    init = function(props)
        local todos, setTodos = lumina.useState({
            makeTodo(1, "Learn Lumina", true),
            makeTodo(2, "Build TUI app", false),
            makeTodo(3, "Deploy to production", false)
        })
        local inputValue, setInputValue = lumina.useState("")
        local nextId, setNextId = lumina.useState(4)
        
        return {
            todos = todos,
            setTodos = setTodos,
            inputValue = inputValue,
            setInputValue = setInputValue,
            nextId = nextId,
            setNextId = setNextId
        }
    end,
    
    render = function(instance)
        -- Get theme
        local theme = lumina.hooks.useTheme and lumina.hooks.useTheme() or {}
        local colors = theme.colors or {primary="cyan", text="white", success="green", secondary="gray"}
        
        -- Count stats
        local total = #instance.todos
        local done = 0
        for _, t in ipairs(instance.todos) do
            if t.done then done = done + 1 end
        end
        
        -- Render todo list
        local function renderTodo(todo)
            local checkbox = todo.done and "[✓]" or "[ ]"
            local textColor = todo.done and colors.secondary or colors.text
            local checkboxColor = todo.done and colors.success or colors.secondary
            
            return {
                type = "hbox",
                children = {
                    {
                        type = "text",
                        content = "  " .. checkbox .. " ",
                        color = checkboxColor
                    },
                    {
                        type = "text",
                        content = todo.text,
                        color = textColor,
                        dim = todo.done
                    }
                }
            }
        end
        
        return {
            type = "vbox",
            children = {
                -- Header
                {
                    type = "text",
                    content = "╔═══════════════════════════════╗",
                    color = colors.secondary
                },
                {
                    type = "text",
                    content = "║       Lumina Todo App         ║",
                    color = colors.primary,
                    bold = true
                },
                {
                    type = "text",
                    content = "╠═══════════════════════════════╣",
                    color = colors.secondary
                },
                
                -- Stats
                {
                    type = "text",
                    content = "║  Total: " .. total .. "  Done: " .. done .. "/" .. total .. string.rep(" ", 10) .. "║",
                    color = colors.text
                },
                {
                    type = "text",
                    content = "╠═══════════════════════════════╣",
                    color = colors.secondary
                },
                
                -- Todo items
                {
                    type = "vbox",
                    children = (function()
                        local items = {}
                        for _, todo in ipairs(instance.todos) do
                            table.insert(items, renderTodo(todo))
                        end
                        if #items == 0 then
                            table.insert(items, {
                                type = "text",
                                content = "║  No todos yet!               ║",
                                color = colors.secondary
                            })
                        end
                        return items
                    end)()
                },
                
                -- Input area
                {
                    type = "text",
                    content = "╠═══════════════════════════════╣",
                    color = colors.secondary
                },
                {
                    type = "text",
                    content = "║  Add new todo:                ║",
                    color = colors.secondary
                },
                {
                    type = "text",
                    content = "║  > " .. instance.inputValue .. "_",
                    color = colors.text
                },
                {
                    type = "text",
                    content = "╚═══════════════════════════════╝",
                    color = colors.secondary
                }
            }
        }
    end
})

-- Render the app
lumina.render(TodoApp)

print()
print("Tip: Use lumina.setState() to modify todos in your app!")
print("Todo app ready!")
