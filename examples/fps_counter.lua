-- Lumina Example: 60 FPS Counter
-- Verifies engine outputs 60 frames per second without dropping states.
-- Usage: lumina examples/fps_counter.lua
-- Test: Record screen with slow-motion camera. Every number should appear.
-- Quit: q or Ctrl+Q

lumina.app {
    id = "fps-counter",
    store = {
        count = 0,
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["Ctrl+q"] = function() lumina.quit() end,
    },

    render = function()
        local count = lumina.useStore("count")

        -- Timer: increment every ~16ms (60 fps)
        lumina.useEffect(function()
            local timer = lumina.setInterval(function()
                local c = lumina.store.get("count")
                lumina.store.set("count", c + 1)
            end, 16)
            return function() lumina.clearInterval(timer) end
        end, {})

        -- Calculate elapsed time
        local seconds = string.format("%.1f", count / 60)

        return lumina.createElement("box", {
            id = "fps-box",
            style = { background = "#1E1E2E", width = "100%", height = "100%" },
        },
            lumina.createElement("text", {
                foreground = "#89B4FA",
                bold = true,
            }, "60 FPS Counter Test"),

            lumina.createElement("text", {
                foreground = "#CDD6F4",
            }, ""),

            lumina.createElement("text", {
                foreground = "#A6E3A1",
                bold = true,
                style = { height = 1 },
            }, "Frame: " .. tostring(count)),

            lumina.createElement("text", {
                foreground = "#CDD6F4",
            }, "Time:  " .. seconds .. "s"),

            lumina.createElement("text", {
                foreground = "#CDD6F4",
            }, ""),

            lumina.createElement("text", {
                foreground = "#6C7086",
            }, "Each number should appear exactly once."),

            lumina.createElement("text", {
                foreground = "#6C7086",
            }, "Record with slow-motion camera to verify."),

            lumina.createElement("text", {
                foreground = "#6C7086",
            }, ""),

            lumina.createElement("text", {
                foreground = "#6C7086",
            }, "q to quit")
        )
    end,
}
