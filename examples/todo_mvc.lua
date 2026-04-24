-- Lumina Example: Todo MVC
-- Showcases: useState, defineComponent, TextInput, filtering, conditional rendering
local lumina = require("lumina")
local theme = {
    bg="#1E1E2E", fg="#CDD6F4", accent="#89B4FA", success="#A6E3A1",
    error="#F38BA8", muted="#6C7086", surface="#313244", headerBg="#181825", done="#585B70",
}
local function renderTodo(todo, isSelected)
    local checkbox = todo.done and "[✓]" or "[ ]"
    local textColor = todo.done and theme.done or theme.fg
    local checkColor = todo.done and theme.success or theme.muted
    return {type="hbox",style={height=1,background=isSelected and theme.surface or nil},children={
        {type="text",content=isSelected and " ▸ " or "   ",style={foreground=theme.accent}},
        {type="text",content=checkbox.." ",style={foreground=checkColor,bold=todo.done}},
        {type="text",content=todo.text,style={foreground=textColor,dim=todo.done}},
        todo.priority == "high" and {type="text",content=" ●",style={foreground=theme.error}}
        or todo.priority == "medium" and {type="text",content=" ●",style={foreground="#F9E2AF"}}
        or nil,
    }}
end
local function renderFilterBar(activeFilter, counts)
    local filters = {{key="all",label="All"},{key="active",label="Active"},{key="completed",label="Completed"}}
    local ch = {}
    for _, f in ipairs(filters) do
        local isActive = (f.key == activeFilter)
        local count = counts[f.key] or 0
        ch[#ch+1] = {type="text",
            content=isActive and string.format(" [ %s (%d) ] ",f.label,count) or string.format("   %s (%d)   ",f.label,count),
            style={foreground=isActive and theme.accent or theme.muted,bold=isActive}}
    end
    return {type="hbox",style={height=1,background=theme.surface},children=ch}
end
local TodoMVC = lumina.defineComponent({
    name = "TodoMVC",
    init = function(props) return {
        todos = {
            {id=1,text="Learn Lumina framework",done=true,priority="high"},
            {id=2,text="Build a TUI application",done=false,priority="high"},
            {id=3,text="Add keyboard shortcuts",done=false,priority="medium"},
            {id=4,text="Write documentation",done=false,priority="medium"},
            {id=5,text="Set up CI/CD pipeline",done=false,priority="low"},
            {id=6,text="Review pull requests",done=true,priority="low"},
            {id=7,text="Deploy to production",done=false,priority="high"},
        },
        inputText = "", filter = "all", selectedIndex = 1, nextId = 8,
    } end,
    render = function(inst)
        local todos = inst.todos or {}
        local filter = inst.filter or "all"
        local selectedIdx = inst.selectedIndex or 1
        local activeCount, completedCount = 0, 0
        for _, todo in ipairs(todos) do
            if todo.done then completedCount = completedCount + 1
            else activeCount = activeCount + 1 end
        end
        local counts = {all=#todos, active=activeCount, completed=completedCount}
        local filtered = {}
        for _, todo in ipairs(todos) do
            if filter == "all" or (filter == "active" and not todo.done) or (filter == "completed" and todo.done) then
                filtered[#filtered+1] = todo
            end
        end
        local todoCh = {}
        if #filtered == 0 then
            todoCh[#todoCh+1] = {type="text",content="  No todos to show.",style={foreground=theme.muted}}
        else
            for i, todo in ipairs(filtered) do
                todoCh[#todoCh+1] = renderTodo(todo, i == selectedIdx)
            end
        end
        return {type="vbox",style={flex=1,background=theme.bg},children={
            {type="hbox",style={height=1,background=theme.headerBg},children={
                {type="text",content=" Todo MVC ",style={foreground=theme.accent,bold=true}},
                {type="text",content=string.format("  %d items, %d active",#todos,activeCount),style={foreground=theme.muted}},
            }},
            {type="hbox",style={height=1,border="single",background=theme.surface},children={
                {type="text",content=" + ",style={foreground=theme.success,bold=true}},
                {type="input",id="todo-input",value=inst.inputText,placeholder="What needs to be done?",style={flex=1,foreground=theme.fg}},
            }},
            renderFilterBar(filter, counts),
            {type="vbox",id="todo-list",style={flex=1,overflow="scroll",border="rounded",background=theme.bg},children=todoCh},
            {type="hbox",style={height=1,background=theme.headerBg},children={
                {type="text",content=string.format(" %d item%s left ",activeCount,activeCount==1 and "" or "s"),style={foreground=theme.fg}},
                {type="text",content=" [Enter] Add  [Space] Toggle  [d] Delete  [f] Filter  [q] Quit ",style={foreground=theme.muted,flex=1}},
            }},
        }}
    end,
})
lumina.render(TodoMVC, {})
