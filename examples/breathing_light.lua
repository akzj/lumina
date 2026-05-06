-- examples/breathing_light.lua — Breathing Light Animation Demo
--
-- Demonstrates lumina animation via setInterval + useState.
-- A dot (●) smoothly pulses between dim and bright using color interpolation.
--
-- Press q or Ctrl+C to quit.

lumina.app {
    id = "breathing-light",
    keys = {
        ["ctrl+c"] = function() lumina.quit() end,
        ["q"] = function() lumina.quit() end,
    },
    render = function()
        local t = lumina.getTheme and lumina.getTheme() or {}

        -- Animation state: phase goes from 0 to 628 (0 to 2π * 100)
        local phase, setPhase = lumina.useState("phase", 0)

        -- Timer: update phase every 50ms (20 fps)
        lumina.useEffect(function()
            local p = phase
            local timer = lumina.setInterval(function()
                p = (p + 5) % 628  -- increment by ~0.05 radians
                setPhase(p)
            end, 50)
            return function() lumina.clearInterval(timer) end
        end, {})

        -- Calculate brightness from sine wave (0.0 to 1.0)
        local radians = phase / 100.0
        local brightness = (math.sin(radians) + 1) / 2  -- 0..1

        -- Interpolate color: dim green (#1a3a1a) to bright green (#00ff88)
        local function lerp(a, b, t)
            return math.floor(a + (b - a) * t)
        end

        local r = lerp(0x1a, 0x00, brightness)
        local g = lerp(0x3a, 0xff, brightness)
        local b = lerp(0x1a, 0x88, brightness)
        local color = string.format("#%02x%02x%02x", r, g, b)

        -- Also interpolate a secondary dot (offset phase for variety)
        local brightness2 = (math.sin(radians + 2.094) + 1) / 2  -- 120° offset
        local r2 = lerp(0x1a, 0x42, brightness2)
        local g2 = lerp(0x1a, 0xb0, brightness2)
        local b2 = lerp(0x3a, 0xff, brightness2)
        local color2 = string.format("#%02x%02x%02x", r2, g2, b2)

        local brightness3 = (math.sin(radians + 4.189) + 1) / 2  -- 240° offset
        local r3 = lerp(0x3a, 0xff, brightness3)
        local g3 = lerp(0x1a, 0x55, brightness3)
        local b3 = lerp(0x1a, 0x00, brightness3)
        local color3 = string.format("#%02x%02x%02x", r3, g3, b3)

        return lumina.createElement("vbox", {
            style = {
                width = "100%", height = "100%",
                justify = "center", align = "center",
            },
        },
            lumina.createElement("text", {
                foreground = t.subtext0 or "#6C7086",
                style = { height = 1, textAlign = "center" },
            }, "Breathing Light Animation"),
            lumina.createElement("text", {
                style = { height = 1 },
            }, ""),
            -- Main breathing dots
            lumina.createElement("hbox", {
                style = { height = 1, justify = "center", gap = 3 },
            },
                lumina.createElement("text", {
                    foreground = color,
                    bold = brightness > 0.7,
                }, "●"),
                lumina.createElement("text", {
                    foreground = color2,
                    bold = brightness2 > 0.7,
                }, "●"),
                lumina.createElement("text", {
                    foreground = color3,
                    bold = brightness3 > 0.7,
                }, "●")
            ),
            lumina.createElement("text", {
                style = { height = 1 },
            }, ""),
            lumina.createElement("text", {
                foreground = color,
                style = { height = 1, textAlign = "center" },
            }, "████████████████"),
            lumina.createElement("text", {
                style = { height = 1 },
            }, ""),
            -- Hover demo: buttons that change style on hover
            lumina.createElement("hbox", {
                style = { height = 1, justify = "center", gap = 2 },
            },
                lumina.createElement("text", {
                    foreground = "#666666",
                    hoverStyle = { foreground = "#ffffff", bold = true, underline = true },
                    onClick = function() end,
                    style = { height = 1 },
                }, "[Hover me]"),
                lumina.createElement("text", {
                    foreground = "#666666",
                    hoverStyle = { foreground = "#00ff88", bold = true },
                    onClick = function() end,
                    style = { height = 1 },
                }, "[Green hover]"),
                lumina.createElement("text", {
                    foreground = "#666666",
                    hoverStyle = { foreground = "#ff5555", bold = true },
                    onClick = function() end,
                    style = { height = 1 },
                }, "[Red hover]")
            ),
            lumina.createElement("text", {
                style = { height = 1 },
            }, ""),
            lumina.createElement("text", {
                foreground = t.subtext0 or "#6C7086",
                dim = true,
                style = { height = 1, textAlign = "center" },
            }, "setInterval(50ms) + useState + math.sin + hoverStyle")
        )
    end,
}
