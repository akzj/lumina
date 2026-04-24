-- Event System Demo: Shows Lumina's event handling and keyboard shortcuts
local lumina = require("lumina")

print("=== Lumina Event System Demo ===")
print()

-- Track event counts
local eventLog = {}

-- Register a global keydown handler
lumina.on("keydown", "demo", function(event)
    table.insert(eventLog, { type = event.type, key = event.key })
    print("Key pressed: " .. event.key)
end)

-- Register a click handler
lumina.on("click", "demo", function(event)
    table.insert(eventLog, { type = event.type, x = event.x, y = event.y })
    print("Clicked at: " .. event.x .. "," .. event.y)
end)

-- Register a focus handler
lumina.on("focus", "demo", function(event)
    print("Focus: " .. event.target)
end)

-- Register a blur handler
lumina.on("blur", "demo", function(event)
    print("Blur: " .. event.target)
end)

-- Register a shortcut: Ctrl+Q to quit
lumina.registerShortcut({ key = "ctrl+q" }, function(event)
    print("Shortcut triggered: Ctrl+Q (quit)")
    print("Total events logged: " .. #eventLog)
    os.exit(0)
end)

-- Register another shortcut: Escape
lumina.registerShortcut({ key = "Escape" }, function(event)
    print("Shortcut triggered: Escape")
end)

print("Registered handlers:")
print("  - keydown: logs all key presses")
print("  - click: logs mouse clicks")
print("  - focus/blur: logs focus changes")
print("  - Ctrl+Q: quits (for testing)")
print("  - Escape: logs shortcut")
print()
print("Focus management test:")
lumina.setFocus("input1")
print("  Focused: " .. lumina.getFocused())

lumina.setFocus("button1")
print("  Focused: " .. lumina.getFocused())

lumina.blur()
print("  After blur: '" .. lumina.getFocused() .. "'")
print()
print("Emitting test events:")

-- Emit a custom event
lumina.emit("demo", "custom", { data = "test" })

-- Simulate key events for testing
print()
print("Simulating Ctrl+C shortcut:")
lumina.emitKeyEvent("c", { ctrl = true })
print()
print("Event system initialized!")
print("Waiting for input... (press Ctrl+Q to quit)")
