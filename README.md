# Lumina — React-Style Terminal UI Framework

> 一套代码 = 桌面客户端 + Web 在线应用（xterm.js）全部支持

Lumina is a **React-inspired terminal UI framework** built with Go + Lua. Write your UI once in Lua, run it as a native terminal app OR serve it to web browsers via xterm.js + WebSocket.

## ✨ Features

### React-Complete API
- **15+ Hooks**: useState, useEffect, useMemo, useCallback, useRef, useContext, useReducer, useTransition, useDeferredValue, useId, useSyncExternalStore, useLayoutEffect, useImperativeHandle, useDebugValue
- **Component Model**: defineComponent, createElement, composition, children props
- **Advanced Patterns**: Suspense + lazy loading, Error Boundaries, React.memo, Portals, forwardRef, Context tree
- **Concurrent Rendering**: useTransition, useDeferredValue for responsive UIs
- **Virtual DOM**: Efficient diff/patch rendering cycle

### Layout & Rendering
- **Flexbox**: Full flex container with grow/shrink/basis, justify/align, gap, wrap
- **CSS Grid**: Grid tracks (fixed, fr, auto), column/row span, gap, auto-flow
- **Overlay System**: z-index stacking, modal backdrop, position absolute/fixed
- **Animation**: useAnimation with 6 easing functions + presets (fadeIn, slideIn, pulse, etc.)
- **SubPixel Rendering**: Half-block characters for smooth graphics
- **Virtual Scrolling**: Efficient rendering of 10,000+ item lists

### 47 shadcn/ui Components
Button, Badge, Card, Alert, AlertDialog, Label, Separator, Skeleton, Spinner, Avatar, Breadcrumb, Input, InputGroup, InputOTP, Switch, Progress, Accordion, Tabs, Table, Pagination, Toggle, ToggleGroup, Select, NativeSelect, Checkbox, RadioGroup, Slider, Textarea, Dialog, Sheet, Drawer, DropdownMenu, ContextMenu, Popover, Tooltip, Command, Combobox, Menubar, ScrollArea, Carousel, Sonner (Toast), HoverCard, Collapsible, Form, Field, Kbd

### State Management & Data
- **createStore**: Zustand-like global state management with dispatch/subscribe
- **useForm**: React Hook Form-like validation (required, email, minLength, maxLength, pattern, min, max, custom)
- **useFetch / useQuery**: Data fetching with cache + stale-while-revalidate
- **Drag & Drop**: useDrag / useDrop with type-based acceptance filtering

### Developer Experience
- **Hot Reload**: File watcher with state preservation
- **DevTools**: Component tree inspector, props/state viewer (F12 toggle)
- **Testing Utilities**: TestRenderer, getByText, getByRole, fireEvent
- **Accessibility**: ARIA attributes, screen reader announcements
- **i18n**: Multi-language support with addTranslation, setLocale, useTranslation
- **4 Built-in Themes**: Catppuccin Mocha, Catppuccin Latte, Tokyo Night, Nord
- **Plugin System**: Extend Lumina with registerPlugin

### Dual Runtime
- `lumina run app.lua` — Native terminal application
- `lumina serve 8080 app.lua` — Web application via xterm.js + WebSocket

---

## 🚀 Quick Start

```bash
# Install
go install github.com/akzj/lumina/cmd/lumina@latest

# Run example locally
lumina run examples/dashboard/main.lua

# Or serve to web browser
lumina serve 8080 examples/dashboard/main.lua
# Open http://localhost:8080
```

---

## 📖 Hello World

```lua
local lumina = require("lumina")

local App = lumina.defineComponent({
    name = "App",
    render = function(self)
        local count, setCount = lumina.useState(0)
        return {
            type = "vbox",
            style = { padding = 1, border = "rounded" },
            children = {
                { type = "text", content = "Count: " .. count },
                lumina.createElement("button", {
                    label = "Increment",
                    onClick = function() setCount(count + 1) end,
                }),
            }
        }
    end,
})

lumina.mount(App)
lumina.run()
```

---

## 🎨 Dashboard Example

```lua
local lumina = require("lumina")
local shadcn = require("shadcn")

-- Global state
local store = lumina.createStore({
    state = { users = 1234, revenue = 56789, orders = 890 },
    actions = {
        refresh = function(state)
            state.users = state.users + math.random(10)
        end,
    }
})

-- Theme
lumina.setTheme("catppuccin-mocha")

-- i18n
lumina.i18n.addTranslation("en", {
    ["dashboard.users"] = "Users",
    ["dashboard.revenue"] = "Revenue",
})
lumina.i18n.addTranslation("zh", {
    ["dashboard.users"] = "用户数",
    ["dashboard.revenue"] = "收入",
})

local Dashboard = lumina.defineComponent({
    name = "Dashboard",
    render = function(self)
        local state = lumina.useStore(store)
        local t = lumina.useTranslation()
        local theme = lumina.useTheme()

        return {
            type = "grid",
            style = { columns = "1fr 1fr", gap = 1 },
            children = {
                lumina.createElement(shadcn.Card, {
                    children = {
                        lumina.createElement(shadcn.CardTitle, {
                            children = {{ type = "text", content = t("dashboard.users") }}
                        }),
                        lumina.createElement(shadcn.CardContent, {
                            children = {{ type = "text", content = tostring(state.users) }}
                        }),
                    }
                }),
                lumina.createElement(shadcn.Card, {
                    children = {
                        lumina.createElement(shadcn.CardTitle, {
                            children = {{ type = "text", content = t("dashboard.revenue") }}
                        }),
                        lumina.createElement(shadcn.CardContent, {
                            children = {{ type = "text", content = "$" .. tostring(state.revenue) }}
                        }),
                    }
                }),
            }
        }
    end,
})

lumina.mount(Dashboard)
lumina.serve(8080)  -- or lumina.run() for terminal
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Lua Application                      │
│   Components · Hooks · Store · Router · shadcn UI        │
├─────────────────────────────────────────────────────────┤
│                    Lumina Go Engine                       │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────┐ │
│  │ Renderer │ │  Hooks   │ │  Layout  │ │   Events   │ │
│  │  (VDOM)  │ │  (15+)   │ │Flex/Grid │ │  Bubbling  │ │
│  └──────────┘ └──────────┘ └──────────┘ └────────────┘ │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────┐ │
│  │ Overlay  │ │Animation │ │  Theme   │ │    i18n    │ │
│  └──────────┘ └──────────┘ └──────────┘ └────────────┘ │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────┐ │
│  │  Store   │ │  Router  │ │ DevTools │ │  Plugins   │ │
│  └──────────┘ └──────────┘ └──────────┘ └────────────┘ │
├─────────────────────────────────────────────────────────┤
│                     Terminal I/O                          │
│  ┌─────────────────────┐  ┌────────────────────────────┐│
│  │   LocalTerminal     │  │  WebTerminal (WebSocket)   ││
│  │   (stdin/stdout)    │  │  (xterm.js + browser)      ││
│  └─────────────────────┘  └────────────────────────────┘│
└─────────────────────────────────────────────────────────┘
```

---

## 📚 API Overview

See [docs/API.md](docs/API.md) for the full API reference.

### Hooks

| Hook | Description |
|------|-------------|
| `useState(initial)` | Local state, returns `value, setter` |
| `useEffect(fn, deps)` | Side effects with dependency tracking |
| `useMemo(fn, deps)` | Memoized computation |
| `useCallback(fn, deps)` | Memoized callback |
| `useRef(initial)` | Mutable ref object |
| `useContext(ctx)` | Read context value |
| `useReducer(reducer, init)` | State with reducer pattern |
| `useTransition()` | Non-urgent state updates |
| `useDeferredValue(value)` | Deferred value for expensive renders |
| `useId()` | Stable unique ID |
| `useSyncExternalStore(sub, snap)` | External store subscription |
| `useLayoutEffect(fn, deps)` | Synchronous effect after render |
| `useImperativeHandle(ref, fn)` | Customize ref value |
| `useDebugValue(value)` | DevTools debug label |
| `useAnimation(opts)` | Animation with easing |

### Components

```lua
-- Define
local MyComponent = lumina.defineComponent({
    name = "MyComponent",
    render = function(self) return { type = "text", content = "Hello" } end,
})

-- Create element
lumina.createElement(MyComponent, { prop1 = "value" })

-- Special components
lumina.Suspense       -- Async loading boundary
lumina.ErrorBoundary  -- Error catching
lumina.Portal         -- Render outside parent tree
lumina.Fragment       -- Group without wrapper
```

### State Management

```lua
local store = lumina.createStore({
    state = { count = 0 },
    actions = {
        increment = function(state) state.count = state.count + 1 end,
    }
})

-- In component:
local state = lumina.useStore(store)
store.dispatch("increment")
```

### Router

```lua
local router = lumina.createRouter()
router.addRoute("/", HomePage)
router.addRoute("/users", UsersPage)
router.addRoute("/users/:id", UserDetailPage)
router.navigate("/users/42")
```

### Form Validation

```lua
local form = lumina.useForm({
    defaultValues = { email = "", name = "" },
    rules = {
        email = {
            { type = "required", message = "Email is required" },
            { type = "email", message = "Invalid email" },
        },
        name = {
            { type = "minLength", value = 2, message = "Too short" },
        },
    },
    onSubmit = function(values) print("Submitted!") end,
})

form.setValue("email", "user@example.com")
form.handleSubmit()  -- validates then calls onSubmit
```

### Theme

```lua
lumina.setTheme("catppuccin-mocha")  -- or "catppuccin-latte", "tokyo-night", "nord"
local theme = lumina.useTheme()      -- { colors = { primary = "...", ... } }
```

### i18n

```lua
lumina.i18n.addTranslation("en", { ["hello"] = "Hello" })
lumina.i18n.addTranslation("zh", { ["hello"] = "你好" })
lumina.i18n.setLocale("zh")
local t = lumina.useTranslation()
t("hello")  -- "你好"
```

### Layout

```lua
-- Flexbox
{ type = "flex", style = { direction = "row", justify = "space-between", align = "center", gap = 1 } }

-- CSS Grid
{ type = "grid", style = { columns = "1fr 2fr 1fr", rows = "auto auto", gap = 1 } }
```

### Drag & Drop

```lua
local drag = lumina.useDrag({ type = "card", data = { id = 1 } })
local drop = lumina.useDrop({ accept = { "card" }, onDrop = function(data) print(data.id) end })

drag.start("card-1")
drop.drop()
```

---

## 🧩 Plugin System

```lua
lumina.registerPlugin({
    name = "my-charts",
    version = "1.0.0",
    init = function(app)
        lumina.defineComponent({ name = "BarChart", ... })
    end,
})

lumina.usePlugin("my-charts")
```

---

## 📊 Project Stats

- **668+ tests** passing
- **47 shadcn/ui components**
- **15+ React hooks** implemented
- **107 Go source files**
- **Dual runtime**: terminal + web (xterm.js)
- **Zero JavaScript build step** — embedded assets via `//go:embed`

---

## 📁 Project Structure

```
lumina/
├── cmd/lumina/              # CLI entry point (run/serve)
├── pkg/lumina/
│   ├── app.go               # Application lifecycle
│   ├── renderer.go          # Virtual DOM renderer
│   ├── vdom.go              # VNode + diff/patch
│   ├── hooks.go             # All React hooks
│   ├── component.go         # Component system
│   ├── store.go             # State management (createStore)
│   ├── router.go            # Client-side router
│   ├── flexbox.go           # Flexbox layout engine
│   ├── grid.go              # CSS Grid layout engine
│   ├── overlay.go           # Overlay/modal system
│   ├── animation.go         # Animation engine
│   ├── theme.go             # Theme system (4 built-in)
│   ├── i18n.go              # Internationalization
│   ├── form_validation.go   # Form validation (useForm)
│   ├── dnd.go               # Drag & drop
│   ├── suspense.go          # Suspense + lazy loading
│   ├── concurrent.go        # useTransition, useDeferredValue
│   ├── accessibility.go     # ARIA attributes
│   ├── devtools.go          # Component inspector
│   ├── fetch.go             # useFetch, useQuery
│   ├── testing_utils.go     # TestRenderer, getByText
│   ├── virtual_scroll.go    # Virtual scrolling
│   ├── web_server.go        # HTTP + WebSocket server
│   ├── web_terminal.go      # Terminal over WebSocket
│   ├── hot_reload.go        # File watcher + HMR
│   ├── plugin.go            # Plugin system
│   ├── lumina.go            # Lua API registration
│   ├── components/
│   │   └── shadcn/          # 47 shadcn/ui components (Lua)
│   │       ├── init.lua
│   │       ├── button.lua
│   │       ├── card.lua
│   │       └── ...
│   └── web/
│       ├── index.html        # Embedded web page
│       └── lumina-client.js   # xterm.js client
├── examples/
│   ├── dashboard/            # Admin dashboard
│   ├── todo/                 # Todo app
│   ├── file-explorer/        # File browser
│   ├── chat/                 # Chat application
│   └── components-showcase/  # Component gallery
└── docs/
    └── API.md                # Full API reference
```

---

## License

MIT
