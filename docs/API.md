# Lumina API Reference

## Table of Contents

- [Core](#core)
- [Hooks](#hooks)
- [Components](#components)
- [State Management](#state-management)
- [Router](#router)
- [Form Validation](#form-validation)
- [Theme](#theme)
- [i18n](#i18n)
- [Layout](#layout)
- [Animation](#animation)
- [Data Fetching](#data-fetching)
- [Drag & Drop](#drag--drop)
- [Suspense & Lazy Loading](#suspense--lazy-loading)
- [Concurrent Rendering](#concurrent-rendering)
- [Accessibility](#accessibility)
- [DevTools](#devtools)
- [Testing Utilities](#testing-utilities)
- [Plugin System](#plugin-system)
- [Web Runtime](#web-runtime)
- [shadcn/ui Components](#shadcnui-components)

---

## Core

### `lumina.defineComponent(opts) → Component`

Define a new component.

```lua
local MyComponent = lumina.defineComponent({
    name = "MyComponent",         -- required: unique name
    init = function(props)         -- optional: initialize state from props
        return { count = props.initial or 0 }
    end,
    render = function(self)        -- required: return VNode tree
        return { type = "text", content = "Hello" }
    end,
})
```

### `lumina.createElement(component, props) → VNode`

Create a virtual element from a component and props.

```lua
lumina.createElement(MyComponent, { initial = 5, children = {...} })
```

### `lumina.mount(component)`

Mount a root component for rendering.

### `lumina.run()`

Start the application in local terminal mode.

### `lumina.serve(port)`

Start the application as a web server on the given port.

### `lumina.render(component, props) → VNode`

Render a component to a VNode tree (useful for testing).

---

## Hooks

All hooks are called inside a component's `render` function.

### `lumina.useState(initialValue) → value, setter`

Local state. `setter(newValue)` triggers re-render.

```lua
local count, setCount = lumina.useState(0)
setCount(count + 1)
-- or functional update:
setCount(function(prev) return prev + 1 end)
```

### `lumina.useEffect(fn, deps)`

Side effect that runs after render. `deps` is a table of dependency values; effect re-runs when deps change. Pass `{}` for mount-only. Omit deps for every render.

```lua
lumina.useEffect(function()
    print("count changed to", count)
    return function()  -- cleanup
        print("cleaning up")
    end
end, { count })
```

### `lumina.useMemo(fn, deps) → value`

Memoized computation. Only recomputes when deps change.

```lua
local sorted = lumina.useMemo(function()
    return expensiveSort(items)
end, { items })
```

### `lumina.useCallback(fn, deps) → fn`

Memoized callback. Returns same function reference when deps haven't changed.

```lua
local handleClick = lumina.useCallback(function()
    setCount(count + 1)
end, { count })
```

### `lumina.useRef(initialValue) → ref`

Mutable reference that persists across renders. `ref.current` holds the value.

```lua
local inputRef = lumina.useRef(nil)
inputRef.current = "some value"
```

### `lumina.useContext(context) → value`

Read a context value from the nearest provider ancestor.

```lua
local ThemeContext = lumina.createContext("dark")
local theme = lumina.useContext(ThemeContext)
```

### `lumina.useReducer(reducer, initialState) → state, dispatch`

State management with a reducer function.

```lua
local state, dispatch = lumina.useReducer(function(state, action)
    if action.type == "increment" then
        return { count = state.count + 1 }
    end
    return state
end, { count = 0 })

dispatch({ type = "increment" })
```

### `lumina.useTransition() → isPending, startTransition`

Mark state updates as non-urgent. UI stays responsive during transition.

```lua
local isPending, startTransition = lumina.useTransition()
startTransition(function()
    setLargeList(newData)  -- non-urgent update
end)
```

### `lumina.useDeferredValue(value, opts) → deferredValue`

Returns a deferred version of a value. Useful for expensive re-renders.

```lua
local deferredSearch = lumina.useDeferredValue(searchInput, { timeoutMs = 300 })
```

### `lumina.useId() → string`

Generate a unique, stable ID. Useful for accessibility attributes.

```lua
local id = lumina.useId()  -- "lumina-1", "lumina-2", etc.
```

### `lumina.useSyncExternalStore(subscribe, getSnapshot) → state`

Subscribe to an external data source.

```lua
local state = lumina.useSyncExternalStore(
    function(callback) return store.subscribe(callback) end,
    function() return store.getState() end
)
```

### `lumina.useLayoutEffect(fn, deps)`

Like `useEffect`, but fires synchronously after render (before paint).

### `lumina.useImperativeHandle(ref, createHandle)`

Customize the value exposed via a ref when using `forwardRef`.

### `lumina.useDebugValue(value)`

Display a label in DevTools for custom hooks.

### `lumina.useAnimation(opts) → value`

Animate a value over time.

```lua
local opacity = lumina.useAnimation({
    from = 0, to = 1,
    duration = 500,
    easing = "easeInOut",  -- "linear"|"easeIn"|"easeOut"|"easeInOut"|"bounce"|"elastic"
})
```

**Presets**: `lumina.useAnimation({ preset = "fadeIn" })` — available presets: `fadeIn`, `fadeOut`, `slideInLeft`, `slideInRight`, `slideInUp`, `slideInDown`, `pulse`, `bounce`

---

## Components

### Special Components

| Component | Description |
|-----------|-------------|
| `lumina.Suspense` | Shows fallback while children load |
| `lumina.ErrorBoundary` | Catches render errors in children |
| `lumina.Portal` | Renders children outside parent tree |
| `lumina.Fragment` | Groups children without a wrapper node |

### `lumina.lazy(loader) → LazyComponent`

Lazy-load a component. Works with Suspense.

```lua
local LazyPage = lumina.lazy(function()
    return require("pages.heavy_page")
end)

lumina.createElement(lumina.Suspense, {
    fallback = { type = "text", content = "Loading..." },
    children = { lumina.createElement(LazyPage, {}) }
})
```

### `lumina.memo(component) → MemoComponent`

Memoize a component — only re-renders when props change.

### `lumina.forwardRef(renderFn) → Component`

Forward a ref to a child component.

### `lumina.createContext(defaultValue) → Context`

Create a context for passing data through the component tree.

---

## State Management

### `lumina.createStore(opts) → store`

Create a global state store (Zustand-like).

```lua
local store = lumina.createStore({
    state = {
        count = 0,
        todos = {},
    },
    actions = {
        increment = function(state)
            state.count = state.count + 1
        end,
        addTodo = function(state, text)
            table.insert(state.todos, { text = text, done = false })
        end,
    }
})
```

### `store.dispatch(action, payload)`

Dispatch an action to update state.

```lua
store.dispatch("increment")
store.dispatch("addTodo", "Buy milk")
```

### `store.getState() → table`

Get current state snapshot.

### `store.subscribe(listener) → unsubscribe`

Subscribe to state changes.

### `lumina.useStore(store) → state`

Hook to subscribe a component to store updates. Component re-renders when state changes.

```lua
local state = lumina.useStore(store)
-- state.count, state.todos
```

---

## Router

### `lumina.createRouter() → router`

Create a client-side router.

```lua
local router = lumina.createRouter()
```

### `router.addRoute(path, component)`

Register a route. Supports path parameters with `:param`.

```lua
router.addRoute("/", HomePage)
router.addRoute("/users", UsersPage)
router.addRoute("/users/:id", UserDetailPage)
```

### `router.navigate(path)`

Navigate to a path.

```lua
router.navigate("/users/42")
```

### `router.back()` / `router.forward()`

Navigate through history.

### `lumina.useRoute() → { path, params, query }`

Hook to read current route info.

```lua
local route = lumina.useRoute()
-- route.path = "/users/42"
-- route.params.id = "42"
```

### `router.Outlet`

Component that renders the matched route's component.

---

## Form Validation

### `lumina.useForm(opts) → form`

Create a form with validation (React Hook Form-like).

```lua
local form = lumina.useForm({
    defaultValues = {
        name = "",
        email = "",
        age = 0,
    },
    rules = {
        name = {
            { type = "required", message = "Name is required" },
            { type = "minLength", value = 2, message = "Too short" },
        },
        email = {
            { type = "required", message = "Email required" },
            { type = "email", message = "Invalid email" },
        },
        age = {
            { type = "min", value = 0, message = "Must be positive" },
            { type = "max", value = 150, message = "Invalid age" },
        },
    },
    onSubmit = function(values)
        print("Submitted:", values.name, values.email)
    end,
})
```

### Form Methods

| Method | Description |
|--------|-------------|
| `form.setValue(field, value)` | Set a field value |
| `form.getValue(field) → value` | Get a field value |
| `form.validate() → bool` | Validate all fields |
| `form.validateField(name) → bool, error` | Validate one field |
| `form.getErrors() → table` | Get all validation errors |
| `form.getValues() → table` | Get all field values |
| `form.handleSubmit() → bool` | Validate + call onSubmit if valid |
| `form.reset()` | Reset to default values |

### Validation Rule Types

| Type | Value | Description |
|------|-------|-------------|
| `required` | — | Field must not be empty/nil |
| `minLength` | `int` | Minimum string length |
| `maxLength` | `int` | Maximum string length |
| `pattern` | `string` (regex) | Must match regex pattern |
| `email` | — | Must be valid email format |
| `min` | `number` | Minimum numeric value |
| `max` | `number` | Maximum numeric value |
| `custom` | — | Custom validator function |

---

## Theme

### `lumina.setTheme(name|table)`

Set the active theme by name or custom definition.

```lua
-- Built-in themes
lumina.setTheme("catppuccin-mocha")  -- dark (default)
lumina.setTheme("catppuccin-latte")  -- light
lumina.setTheme("tokyo-night")
lumina.setTheme("nord")

-- Custom theme
lumina.setTheme({
    colors = {
        background = "#000000",
        foreground = "#FFFFFF",
        primary = "#FF0000",
    }
})
```

### `lumina.useTheme() → theme`

Hook to access the current theme in a component.

```lua
local theme = lumina.useTheme()
-- theme.colors.primary, theme.colors.background, etc.
```

### Theme Color Tokens

| Token | Description |
|-------|-------------|
| `background` | App background |
| `foreground` | Default text color |
| `primary` | Primary accent color |
| `secondary` | Secondary color |
| `accent` | Accent/highlight |
| `destructive` | Error/danger color |
| `muted` | Muted/disabled background |
| `border` | Border color |
| `ring` | Focus ring color |
| `card` | Card background |
| `popover` | Popover background |
| `success` | Success state |
| `warning` | Warning state |
| `info` | Info state |

---

## i18n

### `lumina.i18n.addTranslation(locale, translations)`

Add translations for a locale.

```lua
lumina.i18n.addTranslation("en", {
    ["app.title"] = "My App",
    ["button.submit"] = "Submit",
})
lumina.i18n.addTranslation("zh", {
    ["app.title"] = "我的应用",
    ["button.submit"] = "提交",
})
```

### `lumina.i18n.setLocale(locale)`

Switch the active locale.

```lua
lumina.i18n.setLocale("zh")
```

### `lumina.useTranslation() → t`

Hook that returns a translation function.

```lua
local t = lumina.useTranslation()
t("app.title")  -- "我的应用" (when locale is "zh")
```

---

## Layout

### VNode Types

| Type | Description |
|------|-------------|
| `text` | Text content |
| `hbox` | Horizontal box |
| `vbox` | Vertical box |
| `flex` | Flexbox container |
| `grid` | CSS Grid container |

### Flexbox

```lua
{
    type = "flex",
    style = {
        direction = "row",          -- "row" | "column"
        justify = "space-between",  -- "flex-start"|"center"|"flex-end"|"space-between"|"space-around"|"space-evenly"
        align = "center",           -- "flex-start"|"center"|"flex-end"|"stretch"
        gap = 1,                    -- gap between items
        wrap = "wrap",              -- "nowrap"|"wrap"
    },
    children = {
        { type = "text", content = "A", style = { flex = 1 } },
        { type = "text", content = "B", style = { flex = 2 } },
    }
}
```

### CSS Grid

```lua
{
    type = "grid",
    style = {
        columns = "1fr 2fr 1fr",   -- column track definitions
        rows = "auto auto",         -- row track definitions
        gap = 1,                    -- gap between cells
    },
    children = {
        { type = "text", content = "A", style = { gridColumn = "1/3" } },  -- spans 2 cols
        { type = "text", content = "B", style = { gridRow = "1/3" } },     -- spans 2 rows
    }
}
```

### Style Properties

| Property | Type | Description |
|----------|------|-------------|
| `width` | int | Fixed width |
| `height` | int | Fixed height |
| `padding` | int | Padding (all sides) |
| `border` | string | `"single"` \| `"double"` \| `"rounded"` \| `"bold"` |
| `background` | string | Background color (hex) |
| `foreground` | string | Text color (hex) |
| `bold` | bool | Bold text |
| `italic` | bool | Italic text |
| `underline` | bool | Underlined text |
| `dim` | bool | Dimmed text |
| `zIndex` | int | Overlay stacking order |
| `position` | string | `"relative"` \| `"absolute"` \| `"fixed"` |

---

## Animation

### `lumina.useAnimation(opts) → value`

```lua
local value = lumina.useAnimation({
    from = 0,           -- start value
    to = 100,           -- end value
    duration = 1000,    -- milliseconds
    easing = "easeInOut",
    loop = false,       -- repeat animation
    delay = 0,          -- delay before start (ms)
})
```

### Easing Functions

`linear`, `easeIn`, `easeOut`, `easeInOut`, `bounce`, `elastic`

### Presets

```lua
lumina.useAnimation({ preset = "fadeIn" })
lumina.useAnimation({ preset = "slideInLeft" })
lumina.useAnimation({ preset = "pulse" })
```

Available: `fadeIn`, `fadeOut`, `slideInLeft`, `slideInRight`, `slideInUp`, `slideInDown`, `pulse`, `bounce`

---

## Data Fetching

### `lumina.useFetch(url) → data, loading, error`

Simple data fetching hook.

```lua
local data, loading, error = lumina.useFetch("/api/users")
if loading then return { type = "text", content = "Loading..." } end
```

### `lumina.useQuery(key, fetcher, opts) → result`

Cached data fetching with stale-while-revalidate.

```lua
local result = lumina.useQuery("users", function()
    return lumina.fetch("/api/users")
end, { staleTime = 60 })

-- result.data, result.loading, result.error, result.refetch()
```

---

## Drag & Drop

### `lumina.useDrag(opts) → drag`

Make an element draggable.

```lua
local drag = lumina.useDrag({
    type = "card",                    -- drag type for filtering
    data = { id = 1, title = "Task" }, -- data to transfer
})

drag.start("card-1")        -- begin drag from source ID
drag.stop()                  -- end drag, returns sourceID, targetID, data
drag.isDragging()            -- bool
drag.updatePosition(x, y)   -- update position
```

### `lumina.useDrop(opts) → drop`

Make an element a drop target.

```lua
local drop = lumina.useDrop({
    accept = { "card", "task" },      -- accepted drag types (empty = all)
    onDrop = function(data)            -- called on successful drop
        print("Dropped:", data.title)
    end,
})

drop.canDrop()       -- bool: can current drag be dropped here?
drop.setTarget(id)   -- set this as current drop target
drop.drop()          -- execute drop, returns bool
drop.isOver()        -- bool: is something being dragged?
```

---

## Suspense & Lazy Loading

### `lumina.Suspense`

Shows a fallback while children are loading.

```lua
lumina.createElement(lumina.Suspense, {
    fallback = { type = "text", content = "Loading..." },
    children = { lumina.createElement(LazyComponent, {}) }
})
```

### `lumina.lazy(loader) → LazyComponent`

```lua
local HeavyPage = lumina.lazy(function()
    return require("pages.heavy")
end)
```

---

## Concurrent Rendering

### `lumina.useTransition() → isPending, startTransition`

```lua
local isPending, startTransition = lumina.useTransition()
startTransition(function()
    setExpensiveState(newValue)
end)
-- isPending is true while transition is in progress
```

### `lumina.useDeferredValue(value, opts) → deferred`

```lua
local deferred = lumina.useDeferredValue(searchInput, { timeoutMs = 300 })
```

### `lumina.useId() → string`

```lua
local id = lumina.useId()  -- "lumina-1"
```

---

## Accessibility

### ARIA Attributes

Add `aria` table to any VNode:

```lua
{
    type = "hbox",
    aria = {
        role = "button",
        label = "Submit form",
        description = "Submits the registration form",
        expanded = false,
        selected = true,
        disabled = false,
        hidden = false,
        live = "polite",       -- "polite"|"assertive"|"off"
        controls = "panel-1",  -- ID of controlled element
        labelledBy = "label-1",
    },
    children = {...}
}
```

### `lumina.announce(message, priority)`

Queue a screen reader announcement.

```lua
lumina.announce("Form submitted successfully", "polite")
lumina.announce("Error: invalid input", "assertive")
```

---

## DevTools

### `lumina.devtools.enable()`

Enable the developer tools panel.

### `lumina.devtools.toggle()`

Toggle the DevTools panel (also available via F12 or Ctrl+Shift+D).

### Features

- **Component Tree**: Hierarchical view of all mounted components
- **Props Inspector**: View selected component's props
- **State Inspector**: View useState values
- **Hook Inspector**: View all hooks (useEffect, useMemo, etc.)
- **Re-render Counter**: Track render frequency per component
- **Performance**: Render time per component

---

## Testing Utilities

### `lumina.createTestRenderer() → renderer`

Create a headless renderer for testing.

```lua
local renderer = lumina.createTestRenderer()
renderer.render(MyComponent, { name = "World" })
```

### Test Renderer Methods

| Method | Description |
|--------|-------------|
| `renderer.render(component, props)` | Render a component |
| `renderer.getByText(text) → node` | Find node by text content |
| `renderer.getByRole(role) → node` | Find node by ARIA role |
| `renderer.fireEvent(node, event)` | Simulate an event on a node |

### `lumina.renderToString(component, props) → string`

Render a component to a string (useful for snapshot testing).

---

## Plugin System

### `lumina.registerPlugin(opts)`

Register a plugin.

```lua
lumina.registerPlugin({
    name = "my-charts",
    version = "1.0.0",
    init = function(app)
        -- Register custom components, hooks, etc.
        lumina.defineComponent({ name = "BarChart", ... })
    end,
    hooks = {
        useChart = function(data, options)
            -- custom hook implementation
        end,
    },
})
```

### `lumina.usePlugin(name)`

Activate a registered plugin.

```lua
lumina.usePlugin("my-charts")
```

---

## Web Runtime

### Server Mode

```lua
lumina.serve(8080)  -- Start HTTP + WebSocket server on port 8080
```

Opens a web page with xterm.js that connects via WebSocket to the Go backend. The same Lua app runs identically in both terminal and web modes.

### Protocol

- **Browser → Server**: Raw bytes (keyboard input) or JSON `{"type":"resize","cols":80,"rows":24}`
- **Server → Browser**: Raw bytes (terminal escape sequences)

### Embedded Assets

Web assets (HTML, JS) are embedded in the binary via `//go:embed`, so the binary is fully self-contained.

---

## shadcn/ui Components

All components are in `pkg/lumina/components/shadcn/` and loaded via:

```lua
local shadcn = require("shadcn")
```

### Component List

| Component | File | Description |
|-----------|------|-------------|
| `shadcn.Button` | button.lua | Button with variants: default/outline/secondary/ghost/destructive/link |
| `shadcn.Badge` | badge.lua | Badge with variants: default/secondary/destructive/outline |
| `shadcn.Card` | card.lua | Card container with Header/Title/Description/Content/Footer |
| `shadcn.Alert` | alert.lua | Alert with variants: default/destructive |
| `shadcn.AlertDialog` | alert_dialog.lua | Confirmation dialog |
| `shadcn.Label` | label.lua | Form label |
| `shadcn.Separator` | separator.lua | Horizontal/vertical separator |
| `shadcn.Skeleton` | skeleton.lua | Loading placeholder |
| `shadcn.Spinner` | spinner.lua | Animated loading spinner |
| `shadcn.Avatar` | avatar.lua | Avatar with fallback |
| `shadcn.Breadcrumb` | breadcrumb.lua | Breadcrumb navigation |
| `shadcn.Kbd` | kbd.lua | Keyboard shortcut display |
| `shadcn.Input` | input.lua | Text input field |
| `shadcn.InputGroup` | input_group.lua | Input with prefix/suffix |
| `shadcn.InputOTP` | input_otp.lua | OTP code input |
| `shadcn.Switch` | switch.lua | Toggle switch |
| `shadcn.Progress` | progress.lua | Progress bar |
| `shadcn.Accordion` | accordion.lua | Collapsible sections |
| `shadcn.Tabs` | tabs.lua | Tab navigation |
| `shadcn.Table` | table.lua | Data table |
| `shadcn.Pagination` | pagination.lua | Page navigation |
| `shadcn.Toggle` | toggle.lua | Toggle button |
| `shadcn.ToggleGroup` | toggle_group.lua | Group of toggles |
| `shadcn.Select` | select.lua | Dropdown select |
| `shadcn.NativeSelect` | native_select.lua | Native select |
| `shadcn.Checkbox` | checkbox.lua | Checkbox |
| `shadcn.RadioGroup` | radio_group.lua | Radio button group |
| `shadcn.Slider` | slider.lua | Range slider |
| `shadcn.Textarea` | textarea.lua | Multi-line text input |
| `shadcn.Dialog` | dialog.lua | Modal dialog |
| `shadcn.Sheet` | sheet.lua | Side panel |
| `shadcn.Drawer` | drawer.lua | Bottom drawer |
| `shadcn.DropdownMenu` | dropdown_menu.lua | Dropdown menu |
| `shadcn.ContextMenu` | context_menu.lua | Right-click menu |
| `shadcn.Popover` | popover.lua | Popover |
| `shadcn.Tooltip` | tooltip.lua | Tooltip |
| `shadcn.Command` | command.lua | Command palette |
| `shadcn.Combobox` | combobox.lua | Searchable select |
| `shadcn.Menubar` | menubar.lua | Menu bar |
| `shadcn.ScrollArea` | scroll_area.lua | Scrollable area |
| `shadcn.Carousel` | carousel.lua | Carousel/slider |
| `shadcn.Sonner` | sonner.lua | Toast notifications |
| `shadcn.HoverCard` | hover_card.lua | Hover card |
| `shadcn.Collapsible` | collapsible.lua | Collapsible section |
| `shadcn.Form` | form.lua | Form wrapper |
| `shadcn.Field` | field.lua | Form field |

### Component Usage

```lua
local shadcn = require("shadcn")

-- Button
lumina.createElement(shadcn.Button, {
    label = "Click me",
    variant = "default",  -- "default"|"outline"|"secondary"|"ghost"|"destructive"|"link"
    size = "default",     -- "default"|"sm"|"lg"|"icon"
    disabled = false,
    onClick = function() print("clicked") end,
})

-- Card
lumina.createElement(shadcn.Card, {
    children = {
        lumina.createElement(shadcn.CardTitle, { children = {{ type = "text", content = "Title" }} }),
        lumina.createElement(shadcn.CardContent, { children = {{ type = "text", content = "Content" }} }),
    }
})

-- Dialog
lumina.createElement(shadcn.Dialog, {
    open = showDialog,
    onClose = function() setShowDialog(false) end,
    title = "Confirm",
    children = {{ type = "text", content = "Are you sure?" }},
})

-- Tabs
lumina.createElement(shadcn.Tabs, {
    tabs = { "Tab 1", "Tab 2", "Tab 3" },
    children = {
        { type = "text", content = "Content 1" },
        { type = "text", content = "Content 2" },
        { type = "text", content = "Content 3" },
    }
})
```
