# Lumina Development Guide

## MCP DevTools

Lumina MCP is **AI's TUI Chrome DevTools** — giving AI agents visual debugging capabilities.

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    AI Agent                             │
│  ┌─────────────────────────────────────────────────┐   │
│  │ MCP DevTools APIs                                │   │
│  │  inspect() | simulate() | console | diff       │   │
│  └─────────────────────────────────────────────────┘   │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────┐
│                    Lumina VM                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐          │
│  │  Lua     │  │  Frame   │  │  Event Bus   │          │
│  │  Engine  │  │  History  │  │              │          │
│  └──────────┘  └──────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────┘
```

## API Reference

### Inspect System

```lua
-- Component tree
lumina.inspectTree()  -- → JSON: [{id, type, name, x, y, w, h}]

-- Single component details
lumina.inspectComponent(id)  -- → JSON: {id, type, props, state}

-- Computed styles (final calculated values)
lumina.inspectStyles(id)  -- → JSON: {layout, visual, raw}

-- Frame history
lumina.inspectFrames(10)  -- → JSON: [{timestamp, width, height}]
```

### Simulate System

```lua
-- Simulate a click
lumina.simulateClick("button:submit")

-- Simulate keyboard event
lumina.simulateKey("Enter")
lumina.simulateKey("a", {modifiers = {ctrl = true}})

-- Simulate arbitrary event
lumina.simulate("change", "input:username", {
    value = "new_value"
})
```

### Console

```lua
-- Log messages
lumina.console.log("debug info")
lumina.console.warn("warning message")
lumina.console.error("error occurred")

-- Get logs as JSON
local logs = lumina.console.get()  -- → JSON: [{level, message, time}]
local errors = lumina.console.getErrors()  -- errors only

-- Clear logs
lumina.console.clear()

-- Check size
local n = lumina.console.size()
```

### Diff

```lua
-- Diff last two frames
lumina.diff()  -- → JSON: {patches, stats}

-- Diff frames n and n-1
lumina.diffFrames(5)  -- compare frames 5 and 4
```

### Profile

```lua
-- Get performance data
lumina.profile()  -- → JSON: {total_frames, avg_frame_time, slow_frames}
lumina.profileReset()  -- clear profiling data
lumina.profileSize()  -- → number of recorded frames
```

### Patch (Hot Fix)

```lua
-- Replace component's render function
lumina.patch("MyComponent", [[
    return function()
        return {type = "text", content = "patched!"}
    end
]])

-- Execute arbitrary Lua
lumina.eval("print('debug')")
```

## Output Modes

```lua
-- ANSI mode (default) - human readable
lumina.setOutputMode("ansi")

-- JSON mode - for AI parsing
lumina.setOutputMode("json")

-- Check current mode
local mode = lumina.getOutputMode()
```

## Implementation

### Core Files

| File | Purpose |
|------|---------|
| `pkg/lumina/inspect.go` | Component inspection, frame history |
| `pkg/lumina/inspect_api.go` | Lua bindings for inspect |
| `pkg/lumina/simulate.go` | Event simulation engine |
| `pkg/lumina/console.go` | Log management |
| `pkg/lumina/diff.go` | Frame comparison |
| `pkg/lumina/profile.go` | Performance profiling |
| `pkg/lumina/patch.go` | Hot patching |
| `pkg/lumina/json_adapter.go` | JSON output adapter |

### Data Structures

```go
// ComponentSnapshot - captured component state
type ComponentSnapshot struct {
    ID       string         // unique ID
    Type     string         // component type name
    Name     string         // display name
    X, Y     int            // position
    W, H     int            // dimensions
    Props    map[string]any // current props
    State    map[string]any // component state
}

// ComputedStyles - final calculated styles
type ComputedStyles struct {
    ComponentID string
    Layout      map[string]any // x, y, w, h, flex
    Visual      map[string]any // color, bg, border
    Raw         map[string]any // original props
}

// ConsoleEntry - log entry
type ConsoleEntry struct {
    Level   string // log, warn, error
    Message string
    Data    any
    Time    int64  // milliseconds
}

// DiffResult - frame comparison
type DiffResult struct {
    Patches []DiffPatch
    Stats   DiffStats
}
```

## Testing

```bash
# Run all tests
go test ./pkg/lumina/... -v

# Run MCP DevTools tests
go test ./pkg/lumina/... -v -run "TestInspect|TestConsole|TestOutput"
```

## Usage Example

```lua
local lumina = require("lumina")

-- Set JSON mode for AI
lumina.setOutputMode("json")

-- Create a counter
local Counter = lumina.defineComponent({
    name = "Counter",
    init = function()
        return {count = 0}
    end,
    render = function(inst)
        return {
            type = "vbox",
            children = {
                {type = "text", content = "Count: " .. inst.count},
                {
                    type = "button",
                    label = "+1",
                    onClick = function()
                        inst.count = inst.count + 1
                    end
                }
            }
        }
    end
})

lumina.createComponent(Counter)

-- AI can now inspect and debug
local tree = lumina.inspectTree()
local styles = lumina.inspectStyles("Counter")
lumina.simulateClick("Counter")
lumina.console.log("State after click")
```

## Version History

- **0.3.0**: MCP DevTools complete (inspect, simulate, console, diff, profile, patch)
- **0.2.0**: Components, hooks, theme, events, hotreload
- **0.1.0**: Initial release (basic rendering)
