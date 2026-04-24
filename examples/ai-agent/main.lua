-- AI Agent Dashboard — MCP Agent Interface
-- Usage: lumina examples/ai-agent/main.lua
--
-- Features:
--   • Agent status display (running/idle/error)
--   • Scrollable message log with auto-scroll
--   • Tool call visualization (MCP tool calls with params/results)
--   • Command input at bottom
--   • Real-time updates via store
--   • Simulated agent activity for demo

local lumina = require("lumina")
local shadcn = require("shadcn")

lumina.setTheme("catppuccin-mocha")

-- ─── Store ────────────────────────────────────────────────────────────

local store = lumina.createStore({
    state = {
        status = "idle",
        messages = {
            { role = "system", content = "Agent initialized. Ready for instructions.", timestamp = "00:00:01" },
            { role = "user", content = "Analyze the project structure and suggest improvements.", timestamp = "00:00:05" },
            { role = "assistant", content = "I'll analyze the project structure now. Let me start by examining the directory layout.", timestamp = "00:00:06" },
        },
        tools = {
            { name = "list_directory", params = "path: /src", result = "15 files, 3 dirs", status = "success", timestamp = "00:00:07" },
            { name = "read_file", params = "path: /src/main.go", result = "245 lines read", status = "success", timestamp = "00:00:08" },
        },
        commandInput = "",
        scrollOffset = 0,
        activePanel = "messages",
        stats = {
            tokensUsed = 1247,
            toolCalls = 2,
            elapsed = "8s",
        },
    },
    actions = {
        addMessage = function(state, msg)
            table.insert(state.messages, {
                role = msg.role,
                content = msg.content,
                timestamp = msg.timestamp or os.date("%H:%M:%S"),
            })
        end,
        addToolCall = function(state, tool)
            table.insert(state.tools, {
                name = tool.name,
                params = tool.params or "",
                result = tool.result or "pending...",
                status = tool.status or "running",
                timestamp = tool.timestamp or os.date("%H:%M:%S"),
            })
            state.stats.toolCalls = state.stats.toolCalls + 1
        end,
        setStatus = function(state, status)
            state.status = status
        end,
        setCommand = function(state, cmd)
            state.commandInput = cmd
        end,
        sendCommand = function(state)
            if state.commandInput == "" then return end
            table.insert(state.messages, {
                role = "user",
                content = state.commandInput,
                timestamp = os.date("%H:%M:%S"),
            })
            state.commandInput = ""
            state.status = "running"
        end,
        setActivePanel = function(state, panel)
            state.activePanel = panel
        end,
        updateStats = function(state, stats)
            state.stats.tokensUsed = state.stats.tokensUsed + (stats.tokens or 0)
            state.stats.elapsed = stats.elapsed or state.stats.elapsed
        end,
    },
})

-- ─── Status Badge ─────────────────────────────────────────────────────

local statusConfig = {
    idle = { icon = "⏸", color = "#585B70", label = "Idle" },
    running = { icon = "▶", color = "#A6E3A1", label = "Running" },
    thinking = { icon = "🧠", color = "#89B4FA", label = "Thinking" },
    error = { icon = "⚠", color = "#F38BA8", label = "Error" },
    waiting = { icon = "⏳", color = "#F9E2AF", label = "Waiting" },
}

local StatusBadge = lumina.defineComponent({
    name = "StatusBadge",
    render = function(self)
        local status = self.props.status or "idle"
        local cfg = statusConfig[status] or statusConfig.idle
        return {
            type = "hbox",
            children = {
                { type = "text", content = cfg.icon .. " ", style = { foreground = cfg.color } },
                { type = "text", content = cfg.label, style = { foreground = cfg.color, bold = true } },
            },
        }
    end,
})

-- ─── Message Component ────────────────────────────────────────────────

local roleConfig = {
    user = { icon = "👤", color = "#89B4FA" },
    assistant = { icon = "🤖", color = "#A6E3A1" },
    system = { icon = "⚙", color = "#585B70" },
    error = { icon = "❌", color = "#F38BA8" },
}

local Message = lumina.defineComponent({
    name = "Message",
    render = function(self)
        local msg = self.props.message
        local cfg = roleConfig[msg.role] or roleConfig.system

        return {
            type = "vbox",
            style = { marginBottom = 1 },
            children = {
                {
                    type = "hbox",
                    children = {
                        { type = "text", content = cfg.icon .. " " .. msg.role,
                          style = { foreground = cfg.color, bold = true } },
                        { type = "text", content = "  " .. (msg.timestamp or ""),
                          style = { foreground = "#585B70" } },
                    },
                },
                { type = "text", content = "  " .. msg.content, style = { foreground = "#CDD6F4" } },
            },
        }
    end,
})

-- ─── Tool Call Component ──────────────────────────────────────────────

local ToolCall = lumina.defineComponent({
    name = "ToolCall",
    render = function(self)
        local tool = self.props.tool
        local statusColor = tool.status == "success" and "#A6E3A1"
            or tool.status == "error" and "#F38BA8"
            or "#F9E2AF"

        return lumina.createElement(shadcn.Card, {
            children = {
                {
                    type = "hbox",
                    children = {
                        { type = "text", content = "🔧 " .. tool.name,
                          style = { foreground = "#F9E2AF", bold = true } },
                        { type = "text", content = "  " .. tool.timestamp,
                          style = { foreground = "#585B70" } },
                        { type = "text", content = "  [" .. tool.status .. "]",
                          style = { foreground = statusColor } },
                    },
                },
                { type = "text", content = "  Params: " .. tostring(tool.params),
                  style = { foreground = "#A6ADC8" } },
                { type = "text", content = "  Result: " .. tostring(tool.result),
                  style = { foreground = statusColor } },
            },
        })
    end,
})

-- ─── Messages Panel ───────────────────────────────────────────────────

local MessagesPanel = lumina.defineComponent({
    name = "MessagesPanel",
    render = function(self)
        local state = store.getState()
        local isActive = state.activePanel == "messages"

        local msgNodes = {}
        for _, msg in ipairs(state.messages) do
            table.insert(msgNodes, lumina.createElement(Message, { message = msg }))
        end

        return {
            type = "vbox",
            style = {
                border = isActive and "double" or "rounded",
                padding = 1,
                background = "#1E1E2E",
            },
            children = {
                { type = "text", content = "💬 Messages (" .. #state.messages .. ")",
                  style = { bold = true, foreground = "#89B4FA" } },
                lumina.createElement(shadcn.Separator, {}),
                { type = "vbox", children = msgNodes },
            },
        }
    end,
})

-- ─── Tools Panel ──────────────────────────────────────────────────────

local ToolsPanel = lumina.defineComponent({
    name = "ToolsPanel",
    render = function(self)
        local state = store.getState()
        local isActive = state.activePanel == "tools"

        local toolNodes = {}
        for _, tool in ipairs(state.tools) do
            table.insert(toolNodes, lumina.createElement(ToolCall, { tool = tool }))
        end

        if #state.tools == 0 then
            table.insert(toolNodes, {
                type = "text",
                content = "  No tool calls yet",
                style = { foreground = "#585B70" },
            })
        end

        return {
            type = "vbox",
            style = {
                border = isActive and "double" or "rounded",
                padding = 1,
                background = "#1E1E2E",
            },
            children = {
                { type = "text", content = "🔧 Tool Calls (" .. #state.tools .. ")",
                  style = { bold = true, foreground = "#F9E2AF" } },
                lumina.createElement(shadcn.Separator, {}),
                { type = "vbox", children = toolNodes },
            },
        }
    end,
})

-- ─── Stats Bar ────────────────────────────────────────────────────────

local StatsBar = lumina.defineComponent({
    name = "StatsBar",
    render = function(self)
        local state = store.getState()
        local stats = state.stats

        return {
            type = "hbox",
            style = { padding = 1, background = "#181825" },
            children = {
                { type = "text", content = "📊 ", style = { foreground = "#CBA6F7" } },
                { type = "text", content = "Tokens: " .. tostring(stats.tokensUsed),
                  style = { foreground = "#89B4FA" } },
                { type = "text", content = "  │  ", style = { foreground = "#585B70" } },
                { type = "text", content = "Tools: " .. tostring(stats.toolCalls),
                  style = { foreground = "#F9E2AF" } },
                { type = "text", content = "  │  ", style = { foreground = "#585B70" } },
                { type = "text", content = "Time: " .. stats.elapsed,
                  style = { foreground = "#A6E3A1" } },
            },
        }
    end,
})

-- ─── Command Input ────────────────────────────────────────────────────

local CommandInput = lumina.defineComponent({
    name = "CommandInput",
    render = function(self)
        local state = store.getState()

        return {
            type = "hbox",
            style = { padding = 1, background = "#313244" },
            children = {
                { type = "text", content = "❯ ", style = { foreground = "#A6E3A1", bold = true } },
                lumina.createElement(shadcn.Input, {
                    value = state.commandInput,
                    placeholder = "Type a command for the agent...",
                    onChange = function(v) store.dispatch("setCommand", v) end,
                }),
                lumina.createElement(shadcn.Button, {
                    label = "Send",
                    variant = "default",
                    onClick = function() store.dispatch("sendCommand") end,
                }),
            },
        }
    end,
})

-- ─── App Layout ───────────────────────────────────────────────────────

local App = lumina.defineComponent({
    name = "AIAgentDashboard",
    render = function(self)
        local state = store.getState()

        return {
            type = "vbox",
            style = { background = "#1E1E2E" },
            children = {
                -- Title bar with status
                {
                    type = "hbox",
                    style = { padding = 1, background = "#313244" },
                    children = {
                        { type = "text", content = "🤖 Lumina AI Agent", style = { bold = true, foreground = "#F5C2E7" } },
                        { type = "text", content = "  │  ", style = { foreground = "#585B70" } },
                        lumina.createElement(StatusBadge, { status = state.status }),
                        { type = "text", content = "                    ", style = {} },
                        { type = "text", content = "Tab=panel  Enter=send  q=quit", style = { foreground = "#585B70" } },
                    },
                },
                -- Main content: Messages + Tools side by side
                {
                    type = "hbox",
                    children = {
                        -- Left: Messages (65%)
                        {
                            type = "vbox",
                            style = { width = "65%" },
                            children = {
                                lumina.createElement(MessagesPanel, {}),
                            },
                        },
                        -- Right: Tools (35%)
                        {
                            type = "vbox",
                            style = { width = "35%" },
                            children = {
                                lumina.createElement(ToolsPanel, {}),
                            },
                        },
                    },
                },
                -- Stats bar
                lumina.createElement(StatsBar, {}),
                -- Command input
                lumina.createElement(CommandInput, {}),
            },
        }
    end,
})

-- ─── Keyboard Shortcuts ───────────────────────────────────────────────

lumina.onKey("Tab", function()
    local state = store.getState()
    if state.activePanel == "messages" then
        store.dispatch("setActivePanel", "tools")
    else
        store.dispatch("setActivePanel", "messages")
    end
end)

lumina.onKey("q", function()
    lumina.quit()
end)

-- ─── Demo: Simulate Agent Activity ───────────────────────────────────

-- Add a simulated response after a brief "thinking" period
store.dispatch("setStatus", "thinking")
store.dispatch("addToolCall", {
    name = "analyze_code",
    params = "target: /src/**/*.go",
    result = "Found 12 files, 3 potential improvements",
    status = "success",
})
store.dispatch("addMessage", {
    role = "assistant",
    content = "Analysis complete. I found 3 areas for improvement:\n  1. Error handling in api/handlers.go could use structured errors\n  2. The database layer could benefit from connection pooling\n  3. Test coverage for auth middleware is at 45% — should be >80%",
})
store.dispatch("setStatus", "idle")
store.dispatch("updateStats", { tokens = 856 })

lumina.mount(App)
lumina.run()
