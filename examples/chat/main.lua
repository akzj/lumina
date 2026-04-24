-- ============================================================================
-- Lumina Example: Chat Application
-- ============================================================================
-- Showcases: useState, message list, input handling, timestamps,
--            auto-scroll, user avatars, conditional rendering
--
-- Run: go run ./cmd/lumina examples/chat/main.lua
-- ============================================================================

local lumina = require("lumina")

-- ── Store ───────────────────────────────────────────────────────────────────
local store = lumina.createStore({
    state = {
        messages = {
            { id = 1, user = "Alice",  text = "Hey everyone! 👋",           time = "09:00", avatar = "A" },
            { id = 2, user = "Bob",    text = "Hi Alice! How's the project?", time = "09:01", avatar = "B" },
            { id = 3, user = "Alice",  text = "Going great! Just finished the UI.", time = "09:02", avatar = "A" },
            { id = 4, user = "Carol",  text = "Nice! Can I see a demo?",    time = "09:03", avatar = "C" },
            { id = 5, user = "Alice",  text = "Sure, let me share my screen.", time = "09:04", avatar = "A" },
            { id = 6, user = "Bob",    text = "The new theme looks amazing 🎨", time = "09:05", avatar = "B" },
        },
        currentUser = "You",
        nextId = 7,
        inputMode = false,
    },
})

-- ── Colors ──────────────────────────────────────────────────────────────────
local colors = {
    bg       = "#1E1E2E",
    fg       = "#CDD6F4",
    accent   = "#89B4FA",
    muted    = "#6C7086",
    surface  = "#313244",
    border   = "#45475A",
    avatars  = {
        A = "#F5C2E7",  -- Alice (pink)
        B = "#89B4FA",  -- Bob (blue)
        C = "#A6E3A1",  -- Carol (green)
        Y = "#F9E2AF",  -- You (yellow)
    },
}

-- ── Message Bubble ──────────────────────────────────────────────────────────
local function MessageBubble(props)
    local msg = props.message
    local isOwn = (msg.user == "You")
    local avatarColor = colors.avatars[msg.avatar] or colors.accent

    return {
        type = "vbox",
        style = { height = 3 },
        children = {
            {
                type = "hbox",
                children = {
                    -- Avatar
                    { type = "text",
                      content = isOwn and "" or string.format(" [%s] ", msg.avatar),
                      style = { foreground = avatarColor, bold = true } },
                    -- Username + time
                    { type = "text",
                      content = msg.user,
                      style = { foreground = avatarColor, bold = true } },
                    { type = "text",
                      content = "  " .. msg.time,
                      style = { foreground = colors.muted, dim = true } },
                },
            },
            {
                type = "hbox",
                children = {
                    { type = "text", content = isOwn and "     " or "       " },
                    { type = "text",
                      content = msg.text,
                      style = {
                          foreground = isOwn and "#F9E2AF" or colors.fg,
                          background = isOwn and "#313244" or nil,
                      } },
                },
            },
        },
    }
end

-- ── Message List ────────────────────────────────────────────────────────────
local function MessageList()
    local state = lumina.useStore(store)
    local messages = state.messages or {}

    local children = {}
    for _, msg in ipairs(messages) do
        children[#children + 1] = MessageBubble({ message = msg })
    end

    return {
        type = "vbox",
        style = { scroll = true },
        children = children,
    }
end

-- ── Input Bar ───────────────────────────────────────────────────────────────
local function InputBar()
    local state = lumina.useStore(store)
    local inputMode = state.inputMode

    return {
        type = "vbox",
        children = {
            { type = "text", content = string.rep("─", 60), style = { foreground = colors.border } },
            {
                type = "hbox",
                children = {
                    { type = "text",
                      content = inputMode and " ✏ Type message... " or " Press [i] to type, [q] to quit ",
                      style = { foreground = inputMode and colors.accent or colors.muted } },
                },
            },
        },
    }
end

-- ── Main App ───────────────────────────────────────────────────────────────
local App = lumina.defineComponent({
    name = "ChatApp",
    render = function(self)
        return {
            type = "vbox",
            style = { background = colors.bg },
            children = {
                -- Header
                { type = "text", content = " 💬 Lumina Chat ", style = { foreground = colors.accent, bold = true, background = "#181825" } },
                { type = "text", content = "" },
                -- Messages
                MessageList(),
                -- Input
                InputBar(),
            },
        }
    end,
})

-- ── Key bindings ───────────────────────────────────────────────────────────
local messageTexts = {
    "That's awesome! 🚀",
    "I agree, let's do it!",
    "Can we schedule a meeting?",
    "Great progress everyone! 👏",
    "Let me check and get back to you.",
    "Sounds good to me! 👍",
}
local msgIdx = 1

lumina.onKey("i", function()
    -- Simulate sending a message
    local state = store.getState()
    local messages = state.messages or {}
    local nextId = state.nextId or (#messages + 1)
    local text = messageTexts[msgIdx]
    msgIdx = (msgIdx % #messageTexts) + 1

    local hour = 9 + math.floor(nextId / 6)
    local min = (nextId * 7) % 60
    messages[#messages + 1] = {
        id = nextId,
        user = "You",
        text = text,
        time = string.format("%02d:%02d", hour, min),
        avatar = "Y",
    }
    store.dispatch("setState", { messages = messages, nextId = nextId + 1 })
end)

lumina.onKey("q", function() lumina.quit() end)

lumina.mount(App)
lumina.run()
