-- lua/lux/textarea.lua — Lux Textarea: pure Lua multi-line text input with cursor blinking and CJK support
-- UNCONTROLLED by default: manages its own text state internally for performance.
-- Controlled mode: pass props.value to sync external state.
--
-- Props:
--   value: string (controlled mode) or nil (uncontrolled)
--   initialValue: string — initial text content (default "")
--   placeholder: string — shown when empty
--   onChange: function(text) — called on every text change
--   onSubmit: function(text) — called on Ctrl+J or Alt+Enter with current text
--   maxHeight: number (default 8) — max visible lines
--   foreground: string — text color
--   background: string — background color
--   style: table — additional style for root container
--   id: string
--   autoFocus: boolean
--   focusable: boolean (default true)
--   disabled: boolean

local Textarea = lumina.defineComponent("LuxTextarea", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}

    -- INTERNAL state: useState triggers re-render, useRef holds mutable latest value
    local value, setValue = lumina.useState("val", props.initialValue or "")
    local valueRef = lumina.useRef(props.initialValue or "")
    -- Sync ref → state on each render (ref is always authoritative)
    value = valueRef.current

    -- Controlled mode: if props.value is provided and different, sync it
    if props.value ~= nil and props.value ~= valueRef.current then
        valueRef.current = props.value
        value = props.value
        setValue(props.value)
    end

    local placeholder = props.placeholder or ""
    local onSubmit = props.onSubmit
    local onChange = props.onChange
    local maxHeight = props.maxHeight or 8
    local fg = props.foreground or t.text or "#CDD6F4"
    local bg = props.background or t.surface0 or "#1E1E2E"
    local inputId = props.id or "lux-textarea"

    -- Cursor position: also use ref for same stale-closure reason
    -- In controlled mode (value prop), start cursor at beginning (matches native input behavior).
    -- In uncontrolled mode, start at end of initial value.
    local charCount = utf8.len(value) or 0
    local initCursorPos = (props.value ~= nil) and 1 or (charCount + 1)
    local cursorPos, setCursorPos = lumina.useState("cp", initCursorPos)
    local cursorRef = lumina.useRef(initCursorPos)
    cursorPos = cursorRef.current
    -- Clamp
    if cursorPos > charCount + 1 then cursorPos = charCount + 1 end
    if cursorPos < 1 then cursorPos = 1 end

    -- Blink counter: increment via setInterval. Even = visible, odd = hidden.
    local blink, setBlink = lumina.useState("blink", 0)
    local blinkRef = lumina.useRef(0)

    -- Mount-only: start blink interval + focus
    lumina.useEffect(function()
        if props.autoFocus then
            lumina.focusById(inputId)
        end
        local id = lumina.setInterval(function()
            blinkRef.current = blinkRef.current + 1
            setBlink(blinkRef.current)
        end, 500)
        return function()
            lumina.clearInterval(id)
        end
    end, {})

    local cursorVisible = (blink % 2 == 0)

    -- Helper: get 1-based byte position from 1-based codepoint position
    local function cpToBytePos(cp)
        if cp <= 1 then return 1 end
        if cp > charCount then return #value + 1 end
        return utf8.offset(value, cp) or (#value + 1)
    end

    -- Split value into lines
    local lines = {}
    for line in (value .. "\n"):gmatch("([^\n]*)\n") do
        lines[#lines + 1] = line
    end
    if #lines == 0 then lines = {""} end

    -- Find cursor line and column (codepoint offset within line)
    local cursorLine = 1
    local cursorCol = 0  -- 0-based codepoint offset within current line
    local cpAccum = 0
    for i, line in ipairs(lines) do
        local lineLen = utf8.len(line) or 0
        if cursorPos <= cpAccum + lineLen + 1 then
            cursorLine = i
            cursorCol = cursorPos - cpAccum - 1
            break
        end
        cpAccum = cpAccum + lineLen + 1  -- +1 for the \n
    end

    -- Key handler (reads from refs to avoid stale closure between renders)
    local function handleKey(e)
        if props.disabled then return end
        local k = type(e) == "table" and (e.key or e.Key) or tostring(e or "")
        -- Read latest from refs (not closure-captured state)
        local val = valueRef.current
        local cp = cursorRef.current
        local cc = utf8.len(val) or 0
        -- Clamp
        if cp > cc + 1 then cp = cc + 1 end
        if cp < 1 then cp = 1 end

        -- Local byte-position helper using current val
        local function toByte(cpos)
            if cpos <= 1 then return 1 end
            if cpos > cc then return #val + 1 end
            return utf8.offset(val, cpos) or (#val + 1)
        end

        -- Compute lines and cursor line/col from current val/cp
        local hLines = {}
        for line in (val .. "\n"):gmatch("([^\n]*)\n") do
            hLines[#hLines + 1] = line
        end
        if #hLines == 0 then hLines = {""} end

        local hCursorLine = 1
        local hCursorCol = 0
        local accum = 0
        for i, line in ipairs(hLines) do
            local lineLen = utf8.len(line) or 0
            if cp <= accum + lineLen + 1 then
                hCursorLine = i
                hCursorCol = cp - accum - 1
                break
            end
            accum = accum + lineLen + 1
        end

        local newValue = val
        local newCursor = cp

        if k == "Ctrl+J" or k == "Alt+Enter" then
            if onSubmit then onSubmit(val) end
            valueRef.current = ""
            cursorRef.current = 1
            setValue("")
            setCursorPos(1)
            if onChange then onChange("") end
            return
        elseif k == "Enter" then
            -- In single-line mode (maxHeight=1), treat Enter as submit
            if maxHeight == 1 then
                if onSubmit then onSubmit(val) end
                return
            end
            local bytePos = toByte(cp)
            newValue = val:sub(1, bytePos - 1) .. "\n" .. val:sub(bytePos)
            newCursor = cp + 1
        elseif k == "Backspace" then
            if cp > 1 then
                local startByte = toByte(cp - 1)
                local endByte = toByte(cp)
                newValue = val:sub(1, startByte - 1) .. val:sub(endByte)
                newCursor = cp - 1
            end
        elseif k == "Delete" then
            if cp <= cc then
                local startByte = toByte(cp)
                local endByte = toByte(cp + 1)
                newValue = val:sub(1, startByte - 1) .. val:sub(endByte)
            end
        elseif k == "ArrowLeft" or k == "Left" then
            if cp > 1 then
                newCursor = cp - 1
            end
        elseif k == "ArrowRight" or k == "Right" then
            if cp <= cc then
                newCursor = cp + 1
            end
        elseif k == "ArrowUp" or k == "Up" then
            if hCursorLine > 1 then
                local prevLine = hLines[hCursorLine - 1]
                local prevLineLen = utf8.len(prevLine) or 0
                local col = math.min(hCursorCol, prevLineLen)
                local p = 0
                for i = 1, hCursorLine - 2 do
                    p = p + (utf8.len(hLines[i]) or 0) + 1
                end
                newCursor = p + col + 1
            end
        elseif k == "ArrowDown" or k == "Down" then
            if hCursorLine < #hLines then
                local nextLine = hLines[hCursorLine + 1]
                local nextLineLen = utf8.len(nextLine) or 0
                local col = math.min(hCursorCol, nextLineLen)
                local p = 0
                for i = 1, hCursorLine do
                    p = p + (utf8.len(hLines[i]) or 0) + 1
                end
                newCursor = p + col + 1
            end
        elseif k == "Home" then
            newCursor = cp - hCursorCol
        elseif k == "End" then
            local lineLen = utf8.len(hLines[hCursorLine]) or 0
            newCursor = cp + (lineLen - hCursorCol)
        elseif utf8.len(k) == 1 then
            -- Single character (ASCII printable or multi-byte CJK/emoji)
            local byte1 = k:byte(1)
            if byte1 >= 32 or byte1 >= 128 then
                local bytePos = toByte(cp)
                newValue = val:sub(1, bytePos - 1) .. k .. val:sub(bytePos)
                newCursor = cp + 1
            else
                return  -- control character, ignore
            end
        else
            return  -- unknown key, don't consume
        end

        -- Update refs synchronously (so next keystroke sees latest)
        if newCursor ~= cp then
            cursorRef.current = newCursor
            setCursorPos(newCursor)
        end
        if newValue ~= val then
            valueRef.current = newValue
            setValue(newValue)
            if onChange then onChange(newValue) end
        end

        -- Reset blink on any action (keep cursor visible while interacting)
        blinkRef.current = 0
        setBlink(0)
    end

    -- Build display lines with cursor
    local lineElements = {}
    local totalLines = #lines
    local visibleCount = math.min(totalLines, maxHeight)

    -- Scroll to keep cursor visible
    local scrollY = 0
    if cursorLine > maxHeight then
        scrollY = cursorLine - maxHeight
    end

    -- Empty state
    if #value == 0 then
        -- Show cursor block + placeholder text (dimmed) after cursor
        local phText = placeholder or ""
        if cursorVisible then
            local phAfter = ""
            if utf8.len(phText) and utf8.len(phText) > 1 then
                phAfter = utf8.sub(phText, 2, -1)
            end
            lineElements[1] = lumina.createElement("hbox", {
                key = "ln-1",
                style = { height = 1, width = "100%" },
            },
                lumina.createElement("text", {
                    key = "cursor",
                    foreground = bg or "#1E1E2E",
                    background = fg or "#CDD6F4",
                }, utf8.len(phText) and utf8.len(phText) > 0 and utf8.sub(phText, 1, 1) or " "),
                lumina.createElement("text", {
                    key = "ph-rest",
                    foreground = t.muted or "#6C7086",
                    dim = true,
                }, phAfter)
            )
        else
            lineElements[1] = lumina.createElement("text", {
                key = "ph",
                foreground = t.muted or "#6C7086",
                dim = true,
                style = { height = 1, width = "100%" },
            }, phText)
        end
    else
        for i = 1, visibleCount do
            local lineIdx = i + scrollY
            local line = lines[lineIdx] or ""

            if lineIdx == cursorLine then
                if cursorVisible then
                    local lineLen = utf8.len(line) or 0
                    local before = ""
                    local cursorChar = " "  -- space if at end of line
                    local after = ""

                    if cursorCol > 0 then
                        before = utf8.sub(line, 1, cursorCol)
                    end
                    if cursorCol < lineLen then
                        cursorChar = utf8.sub(line, cursorCol + 1, cursorCol + 1)
                        after = utf8.sub(line, cursorCol + 2, -1)
                    end

                    -- Build hbox with text segments + inverted cursor
                    local segments = {}
                    if #before > 0 then
                        segments[#segments + 1] = lumina.createElement("text", {
                            key = "before",
                            foreground = fg,
                        }, before)
                    end
                    -- Cursor char with inverted colors
                    segments[#segments + 1] = lumina.createElement("text", {
                        key = "cursor",
                        foreground = bg or "#1E1E2E",
                        background = fg or "#CDD6F4",
                    }, cursorChar)
                    if #after > 0 then
                        segments[#segments + 1] = lumina.createElement("text", {
                            key = "after",
                            foreground = fg,
                        }, after)
                    end

                    lineElements[#lineElements + 1] = lumina.createElement("hbox", {
                        key = "ln-" .. i,
                        style = { height = 1, width = "100%" },
                    }, table.unpack(segments))
                else
                    -- Blink off: show line normally
                    lineElements[#lineElements + 1] = lumina.createElement("text", {
                        key = "ln-" .. i,
                        foreground = fg,
                        style = { height = 1, width = "100%" },
                    }, line)
                end
            else
                -- Non-cursor line: plain text
                lineElements[#lineElements + 1] = lumina.createElement("text", {
                    key = "ln-" .. i,
                    foreground = fg,
                    style = { height = 1, width = "100%" },
                }, line)
            end
        end
    end

    -- Merge background into style so the vbox fills its area with bg color
    local baseStyle = props.style or { flex = 1, minHeight = 1 }
    if bg and bg ~= "" then
        baseStyle = {}
        local src = props.style or { flex = 1, minHeight = 1 }
        for k, v in pairs(src) do baseStyle[k] = v end
        baseStyle.background = bg
    end

    -- Compute cursor hint for hardware cursor positioning (0-based offsets from node origin)
    local hintCol = cursorCol  -- codepoint offset = display width for ASCII/Latin
    local hintRow = cursorLine - 1 - scrollY  -- 0-based visible row
    if hintRow < 0 then hintRow = 0 end

    return lumina.createElement("vbox", {
        id = inputId,
        focusable = props.focusable ~= false,
        autoFocus = props.autoFocus,
        onKeyDown = handleKey,
        cursorHintCol = hintCol,
        cursorHintRow = hintRow,
        style = baseStyle,
    }, table.unpack(lineElements))
end)

return Textarea
