-- ============================================================================
-- Lumina Example: Chat Application
-- ============================================================================
-- Showcases: defineComponent, VNode trees, scrollable lists, TextInput,
--            StatusBar, theming, message formatting, conditional rendering.
--
-- Run: lumina examples/chat_app.lua
-- ============================================================================

local lumina = require("lumina")

-- ── Theme ──────────────────────────────────────────────────────────────────
local darkTheme = {
    bg        = "#1E1E2E",
    fg        = "#CDD6F4",
    accent    = "#89B4FA",
    success   = "#A6E3A1",
    error     = "#F38BA8",
    muted     = "#6C7086",
    surface   = "#313244",
    headerBg  = "#181825",
}

-- ── Helper: format a chat message ──────────────────────────────────────────
local function formatMessage(msg, theme)
    local senderColor
    if msg.sender == "System" then
        senderColor = theme.muted
    elseif msg.sender == "You" then
        senderColor = theme.accent
    else
        senderColor = theme.success
    end

    return {
        type = "hbox",
        children = {
            { type = "text",
              content = string.format("[%s] ", msg.time),
              style = { foreground = theme.muted, dim = true } },
            { type = "text",
              content = msg.sender .. ": ",
              style = { foreground = senderColor, bold = true } },
            { type = "text",
              content = msg.text,
              style = { foreground = theme.fg } },
        },
    }
end

-- ── Helper: build the message list (scrollable) ───────────────────────────
local function renderMessageList(messages, theme)
    local children = {}
    for _, msg in ipairs(messages) do
        children[#children + 1] = formatMessage(msg, theme)
    end

    if #children == 0 then
        children[#children + 1] = {
            type = "text",
            content = "  No messages yet...",
            style = { foreground = theme.muted },
        }
    end

    return {
        type = "vbox",
        id = "message-list",
        style = {
            flex = 1,
            overflow = "scroll",
            border = "rounded",
            background = theme.bg,
            padding = 1,
        },
        children = children,
    }
end

-- ── Main Component ─────────────────────────────────────────────────────────
local ChatApp = lumina.defineComponent({
    name = "ChatApp",

    init = function(props)
        return {
            messages = {
                { sender = "System", text = "Welcome to Lumina Chat!", time = "12:00" },
                { sender = "Alice",  text = "Hello everyone!",        time = "12:01" },
                { sender = "Bob",    text = "Hey Alice! 👋",          time = "12:01" },
                { sender = "Alice",  text = "How's the new TUI framework?", time = "12:02" },
                { sender = "Bob",    text = "Lumina is amazing — Lua + Go = ❤️", time = "12:02" },
                { sender = "System", text = "Carol has joined the chat.", time = "12:03" },
                { sender = "Carol",  text = "Hi all! What did I miss?", time = "12:03" },
            },
            inputText = "",
            connected = true,
            theme = darkTheme,
            unread = 0,
        }
    end,

    render = function(instance)
        local theme = instance.theme
        local msgCount = #instance.messages

        return {
            type = "vbox",
            style = { flex = 1, background = theme.bg },
            children = {
                -- ── Header ─────────────────────────────────────────────
                {
                    type = "hbox",
                    style = { height = 1, background = theme.headerBg, padding = 0 },
                    children = {
                        { type = "text",
                          content = " 💬 Lumina Chat ",
                          style = { foreground = theme.accent, bold = true } },
                        { type = "text",
                          content = "  #general",
                          style = { foreground = theme.muted } },
                    },
                },

                -- ── Message List (scrollable) ──────────────────────────
                renderMessageList(instance.messages, theme),

                -- ── Typing indicator ───────────────────────────────────
                {
                    type = "text",
                    content = "  Alice is typing...",
                    style = { foreground = theme.muted, dim = true, height = 1 },
                },

                -- ── Input Area ─────────────────────────────────────────
                {
                    type = "hbox",
                    style = { height = 1, border = "single", background = theme.surface },
                    children = {
                        { type = "text",
                          content = " > ",
                          style = { foreground = theme.accent } },
                        { type = "input",
                          id = "chat-input",
                          value = instance.inputText,
                          placeholder = "Type a message...",
                          style = { flex = 1, foreground = theme.fg } },
                    },
                },

                -- ── Status Bar ─────────────────────────────────────────
                {
                    type = "hbox",
                    style = { height = 1, background = theme.headerBg },
                    children = {
                        { type = "text",
                          content = instance.connected
                              and " ● Connected"
                              or  " ○ Disconnected",
                          style = {
                              foreground = instance.connected
                                  and theme.success
                                  or  theme.error,
                          } },
                        { type = "text",
                          content = string.format("  %d messages ", msgCount),
                          style = { foreground = theme.muted } },
                        { type = "text",
                          content = " [Ctrl+T] Theme  [Ctrl+Q] Quit ",
                          style = { foreground = theme.muted, flex = 1 } },
                    },
                },
            },
        }
    end,
})

-- ── Render ─────────────────────────────────────────────────────────────────
lumina.render(ChatApp, {})
