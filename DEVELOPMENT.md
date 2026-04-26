# Lumina Development Guide

This guide focuses on **MCP / DevTools-style APIs** and where they live in the Go tree. **Authoritative Lua signatures** for the whole `lumina` module are in [`docs/API.md`](docs/API.md). High-level architecture is in [`DESIGN.md`](DESIGN.md).

---

## MCP DevTools (Lua surface)

These functions exist on the table returned by `require("lumina")` after `lumina.Open` (or the global opener). They are primarily for **automation, MCP, and headless debugging**.

For a **long-running HTTP MCP** server that speaks JSON-RPC to tools, see `cmd/lumina-mcp-http/main.go`.

### Architecture (conceptual)

```
┌─────────────────────────────────────────────────────────┐
│                    AI Agent / MCP client                 │
│  inspect* | simulate* | console | diff | profile | …    │
└───────────────────────────┬─────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────┐
│                    Lumina App                          │
│  Lua VM  +  Component registry  +  last Frame history  │
└─────────────────────────────────────────────────────────┘
```

---

## Inspect

All inspect helpers return a **JSON string** (Lua single return), unless noted.

### Direct functions (recommended)

```lua
-- Full component tree (array of snapshots)
local json = lumina.inspectTree()

-- One component by registry ID (see VNode props["id"] / component ID)
local json = lumina.inspectComponent("my-widget")

-- Computed layout/visual snapshot for that ID
local json = lumina.inspectStyles("my-widget")

-- Recent frame metadata (width/height/timestamp/dirty rect count)
local json = lumina.inspectFrames(10)   -- optional count, default 10, max 100
```

### `lumina.inspect(action, …)` (dispatcher)

`lumina.inspect` routes string actions in `inspect_api.go`:

- `lumina.inspect("tree")` → same as `inspectTree()`
- `lumina.inspect("component", id)` → component detail
- `lumina.inspect("styles", id)` → styles
- `lumina.inspect("frames", count?)` → frames list

Prefer the **direct** `lumina.inspectTree` / `lumina.inspectComponent` / … calls in scripts to keep argument order obvious.

### Other registry helpers

```lua
local json = lumina.getState(componentId)
local json = lumina.getAllComponents()   -- JSON array of IDs
```

---

## Simulate

Targets are **string IDs** — usually the `id` you set on a focusable / clickable VNode (`props.id`), not the Lua component `name` unless they match.

```lua
-- Click (synthesized mousedown/up on target id)
lumina.simulateClick("counter-inc")

-- Key (keydown); optional modifiers table
lumina.simulateKey("enter")
lumina.simulateKey("a", { modifiers = { ctrl = true } })

-- Generic simulated event (see simulate.go / SimulatedEvent)
lumina.simulate("change", "input:username", { value = "new" })
```

Return values: **`simulate`** pushes `success` and optional `error` string; **`simulateClick`** / **`simulateKey`** push a boolean success (see `simulate.go`).

---

## Console

Two surfaces:

1. **`lumina.console`** sub-table — ergonomic logging from Lua:

```lua
lumina.console.log("debug", { detail = 1 })
lumina.console.warn("heads up")
lumina.console.error("boom")

local json = lumina.console.get()   -- all entries as JSON string
lumina.console.clear()
local n = lumina.console.size()
```

2. **Top-level** helpers (also used internally / MCP):

```lua
lumina.consoleLog("info", "message", optionalDataTable)
local errs = lumina.consoleGetErrors()   -- JSON string, errors only
```

Log entry shape is `ConsoleEntry` in `console.go` (`level`, `message`, `data`, `time`, optional `stack`).

---

## Diff & frames

Frame history is populated when the renderer records frames (`RecordFrame`); diff compares entries from that ring buffer.

```lua
local json = lumina.diff()           -- last two recorded frames
local json = lumina.diffFrames(5)  -- frames[len-5] vs frames[len-1]; n >= 2, default 2
```

JSON matches `DiffResult` in `diff.go` (`patches`, `stats`, optional `before`/`after` frame snapshots depending on marshaling).

---

## Profile

```lua
local json = lumina.profile()       -- ProfileResult JSON (avg/min/max ms, slow frame count, …)
lumina.profileReset()
local n = lumina.profileSize()
```

---

## Patch & eval

```lua
-- patch(componentIdOrNameOrType, luaSource) → success, error?
-- Current implementation: validates Lua compiles (DoString); it does NOT
-- replace the component's render closure yet — see patch.go TODO.
lumina.patch("MyComponent", [[ print("syntax check") ]])

lumina.eval([[ return 1+1 ]])   -- run arbitrary Lua in the VM (use with care)
```

---

## Output modes

```lua
lumina.setOutputMode("ansi")   -- default terminal output
lumina.setOutputMode("json")   -- machine-oriented frame path (see json_adapter.go)

local mode = lumina.getOutputMode()
```

---

## Implementation map

| File | Role |
|------|------|
| `pkg/lumina/inspect_api.go` | Lua bindings: `inspect*` dispatch, `getState`, `getAllComponents` |
| `pkg/lumina/inspect.go` | `ComponentSnapshot`, `ComputedStyles`, tree/style helpers, frame history |
| `pkg/lumina/simulate.go` | `Simulate`, `simulateClick`, `simulateKey`, generic `simulate` |
| `pkg/lumina/console.go` | `Console`, `ConsoleEntry`, `consoleGet` / `makeConsoleLog` |
| `pkg/lumina/diff.go` | `DiffFrames`, `diff` / `diffFrames` Lua |
| `pkg/lumina/profile.go` | Frame timing ring, `profile` Lua |
| `pkg/lumina/patch.go` | `patch` (validation stub), related eval |
| `pkg/lumina/json_adapter.go` | JSON `OutputAdapter` |
| `pkg/lumina/mcp_debug.go` | MCP tool wiring / debug bridge (large surface) |

---

## Data structures (Go reference)

```go
// inspect.go
type ComponentSnapshot struct {
    ID       string         `json:"id"`
    Type     string         `json:"type"`
    Name     string         `json:"name"`
    X, Y, W, H int          `json:"x","y","w","h"`
    Props    map[string]any `json:"props"`
    State    map[string]any `json:"state"`
    Children []string       `json:"children,omitempty"`
}

type ComputedStyles struct {
    ComponentID string         `json:"component_id"`
    Layout      map[string]any `json:"layout"`
    Visual      map[string]any `json:"visual"`
    Raw         map[string]any `json:"raw"`
}

// console.go
type ConsoleEntry struct {
    Level   string `json:"level"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
    Time    int64  `json:"time"`
    Stack   string `json:"stack,omitempty"`
}

// diff.go
type DiffResult struct {
    Before  *Frame      `json:"before,omitempty"`
    After   *Frame      `json:"after,omitempty"`
    Patches []DiffPatch `json:"patches"`
    Stats   DiffStats   `json:"stats"`
}
```

---

## Testing

```bash
go test ./pkg/lumina/... -count=1

# Focus DevTools-related tests (names may drift — use -run regex)
go test ./pkg/lumina/... -run 'Inspect|Console|Simulate|Diff|Profile' -count=1
```

---

## Minimal Lua example (aligned with current hooks)

```lua
local lumina = require("lumina")

local Counter = lumina.defineComponent({
    name = "Counter",
    render = function()
        local n, setN = lumina.useState("n", 0)
        return {
            type = "vbox",
            children = {
                { type = "text", content = "Count: " .. tostring(n) },
                {
                    type = "box",
                    id = "counter-inc",
                    style = { border = "rounded" },
                    onClick = function() setN(n + 1) end,
                    children = { { type = "text", content = " +1 " } },
                },
            },
        }
    end,
})

lumina.mount(Counter)
-- lumina.run()  -- interactive apps: then inspectTree / simulateClick("counter-inc") etc.
```

`simulateClick` / `inspectStyles` need a real **`id`** on the VNode; the component **`name`** is not automatically the DOM-style id.

---

## Version notes (documentation)

| Doc / area | Note |
|------------|------|
| **0.3.x** | This file + `DESIGN.md` + `docs/API.md` refreshed to match `luaLoader` and actual DevTools behavior. |
| **Patch** | Documented as **syntax-validation only** until `patch.go` TODO is implemented. |
| **Older samples** | Any example using `useState` without a string key, hooks inside `init`, or `type = "button"` may predate the current renderer / hook rules—prefer `docs/API.md` + `examples/`. |
