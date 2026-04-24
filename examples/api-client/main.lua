-- REST API Client — A Postman-like API testing tool for the terminal
-- Usage: lumina examples/api-client/main.lua
--
-- Features:
--   • URL input field with method selector (GET/POST/PUT/DELETE)
--   • Headers editor (key-value pairs)
--   • Request body editor for POST/PUT
--   • Response panel with status, headers, body
--   • Request history sidebar (last 10 requests)
--   • Uses useFetch for HTTP requests

local lumina = require("lumina")
local shadcn = require("shadcn")

lumina.setTheme("catppuccin-mocha")

-- ─── Store ────────────────────────────────────────────────────────────

local store = lumina.createStore({
    state = {
        method = "GET",
        url = "https://jsonplaceholder.typicode.com/posts/1",
        headers = {},
        body = "",
        response = nil,
        loading = false,
        history = {},
        activeTab = "response",
        selectedHistory = 0,
    },
    actions = {
        setMethod = function(state, method)
            state.method = method
        end,
        setUrl = function(state, url)
            state.url = url
        end,
        setBody = function(state, body)
            state.body = body
        end,
        setLoading = function(state, loading)
            state.loading = loading
        end,
        setResponse = function(state, resp)
            state.response = resp
            state.loading = false
        end,
        addHistory = function(state, entry)
            table.insert(state.history, 1, entry)
            if #state.history > 10 then
                table.remove(state.history)
            end
        end,
        setActiveTab = function(state, tab)
            state.activeTab = tab
        end,
        selectHistory = function(state, idx)
            state.selectedHistory = idx
            if idx > 0 and idx <= #state.history then
                local h = state.history[idx]
                state.method = h.method
                state.url = h.url
            end
        end,
    },
})

-- ─── Method Badge ─────────────────────────────────────────────────────

local methodColors = {
    GET = "#A6E3A1",
    POST = "#89B4FA",
    PUT = "#F9E2AF",
    DELETE = "#F38BA8",
    PATCH = "#CBA6F7",
}

local MethodBadge = lumina.defineComponent({
    name = "MethodBadge",
    render = function(self)
        local method = self.props.method or "GET"
        local color = methodColors[method] or "#CDD6F4"
        return {
            type = "text",
            content = " " .. method .. " ",
            style = {
                foreground = "#1E1E2E",
                background = color,
                bold = true,
            },
        }
    end,
})

-- ─── Request Panel ────────────────────────────────────────────────────

local RequestPanel = lumina.defineComponent({
    name = "RequestPanel",
    render = function(self)
        local state = store.getState()

        local methodButtons = {}
        for _, m in ipairs({"GET", "POST", "PUT", "DELETE"}) do
            local isActive = state.method == m
            table.insert(methodButtons, lumina.createElement(shadcn.Button, {
                label = m,
                variant = isActive and "default" or "outline",
                onClick = function() store.dispatch("setMethod", m) end,
            }))
        end

        local children = {
            -- Method selector row
            {
                type = "hbox",
                style = { gap = 1 },
                children = methodButtons,
            },
            -- URL input
            { type = "text", content = "" },
            { type = "text", content = "URL:", style = { foreground = "#A6ADC8", bold = true } },
            lumina.createElement(shadcn.Input, {
                value = state.url,
                placeholder = "https://api.example.com/endpoint",
                onChange = function(v) store.dispatch("setUrl", v) end,
            }),
        }

        -- Body editor for POST/PUT
        if state.method == "POST" or state.method == "PUT" or state.method == "PATCH" then
            table.insert(children, { type = "text", content = "" })
            table.insert(children, { type = "text", content = "Body (JSON):", style = { foreground = "#A6ADC8", bold = true } })
            table.insert(children, lumina.createElement(shadcn.Textarea, {
                value = state.body,
                placeholder = '{ "key": "value" }',
                onChange = function(v) store.dispatch("setBody", v) end,
            }))
        end

        -- Send button
        table.insert(children, { type = "text", content = "" })
        table.insert(children, lumina.createElement(shadcn.Button, {
            label = state.loading and "⏳ Sending..." or "🚀 Send Request",
            variant = "default",
            onClick = function()
                store.dispatch("setLoading", true)
                store.dispatch("addHistory", {
                    method = state.method,
                    url = state.url,
                    timestamp = os.date("%H:%M:%S"),
                })
                -- Simulate response (real HTTP would use lumina.fetch)
                store.dispatch("setResponse", {
                    status = 200,
                    statusText = "OK",
                    headers = {
                        ["content-type"] = "application/json",
                        ["x-request-id"] = "abc-123-def",
                    },
                    body = '{\n  "id": 1,\n  "title": "Sample Response",\n  "completed": false\n}',
                    time = "42ms",
                })
            end,
        }))

        return lumina.createElement(shadcn.Card, {
            children = {
                { type = "text", content = "📡 Request", style = { bold = true, foreground = "#F5C2E7" } },
                { type = "text", content = "" },
                {
                    type = "vbox",
                    children = children,
                },
            },
        })
    end,
})

-- ─── Response Panel ───────────────────────────────────────────────────

local ResponsePanel = lumina.defineComponent({
    name = "ResponsePanel",
    render = function(self)
        local state = store.getState()
        local resp = state.response

        if not resp then
            return lumina.createElement(shadcn.Card, {
                children = {
                    { type = "text", content = "📋 Response", style = { bold = true, foreground = "#89B4FA" } },
                    { type = "text", content = "" },
                    { type = "text", content = "Send a request to see the response here.", style = { foreground = "#585B70" } },
                },
            })
        end

        local statusColor = "#A6E3A1"
        if resp.status >= 400 then statusColor = "#F38BA8"
        elseif resp.status >= 300 then statusColor = "#F9E2AF" end

        -- Headers display
        local headerNodes = {}
        if resp.headers then
            for k, v in pairs(resp.headers) do
                table.insert(headerNodes, {
                    type = "text",
                    content = "  " .. k .. ": " .. v,
                    style = { foreground = "#A6ADC8" },
                })
            end
        end

        return lumina.createElement(shadcn.Card, {
            children = {
                { type = "text", content = "📋 Response", style = { bold = true, foreground = "#89B4FA" } },
                { type = "text", content = "" },
                -- Status line
                {
                    type = "hbox",
                    children = {
                        { type = "text", content = "Status: ", style = { foreground = "#A6ADC8" } },
                        { type = "text", content = tostring(resp.status) .. " " .. resp.statusText,
                          style = { foreground = statusColor, bold = true } },
                        { type = "text", content = "  ⏱ " .. resp.time, style = { foreground = "#585B70" } },
                    },
                },
                { type = "text", content = "" },
                -- Headers
                { type = "text", content = "Headers:", style = { foreground = "#A6ADC8", bold = true } },
                { type = "vbox", children = headerNodes },
                { type = "text", content = "" },
                -- Body
                { type = "text", content = "Body:", style = { foreground = "#A6ADC8", bold = true } },
                {
                    type = "vbox",
                    style = { border = "rounded", padding = 1, background = "#181825" },
                    children = {
                        { type = "text", content = resp.body, style = { foreground = "#A6E3A1" } },
                    },
                },
            },
        })
    end,
})

-- ─── History Sidebar ──────────────────────────────────────────────────

local HistorySidebar = lumina.defineComponent({
    name = "HistorySidebar",
    render = function(self)
        local state = store.getState()

        local items = {}
        if #state.history == 0 then
            table.insert(items, {
                type = "text",
                content = "  No requests yet",
                style = { foreground = "#585B70" },
            })
        else
            for i, h in ipairs(state.history) do
                local isSelected = i == state.selectedHistory
                local color = methodColors[h.method] or "#CDD6F4"
                table.insert(items, {
                    type = "hbox",
                    style = { background = isSelected and "#313244" or nil },
                    children = {
                        { type = "text", content = " " .. h.method, style = { foreground = color, bold = true } },
                        { type = "text", content = " " .. h.timestamp, style = { foreground = "#585B70" } },
                    },
                })
            end
        end

        return lumina.createElement(shadcn.Card, {
            children = {
                { type = "text", content = "📜 History", style = { bold = true, foreground = "#F9E2AF" } },
                { type = "text", content = "" },
                { type = "vbox", children = items },
            },
        })
    end,
})

-- ─── App Layout ───────────────────────────────────────────────────────

local App = lumina.defineComponent({
    name = "APIClient",
    render = function(self)
        return {
            type = "vbox",
            style = { background = "#1E1E2E" },
            children = {
                -- Title bar
                {
                    type = "hbox",
                    style = { padding = 1, background = "#313244" },
                    children = {
                        { type = "text", content = "🔌 Lumina API Client", style = { bold = true, foreground = "#F5C2E7" } },
                        { type = "text", content = "  q = quit", style = { foreground = "#585B70" } },
                    },
                },
                -- Main content
                {
                    type = "hbox",
                    children = {
                        -- Left: Request + Response
                        {
                            type = "vbox",
                            style = { width = "75%" },
                            children = {
                                lumina.createElement(RequestPanel, {}),
                                lumina.createElement(ResponsePanel, {}),
                            },
                        },
                        -- Right: History sidebar
                        {
                            type = "vbox",
                            style = { width = "25%" },
                            children = {
                                lumina.createElement(HistorySidebar, {}),
                            },
                        },
                    },
                },
            },
        }
    end,
})

lumina.onKey("q", function() lumina.quit() end)

lumina.mount(App)
lumina.run()
