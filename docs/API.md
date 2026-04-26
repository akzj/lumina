# Lumina API Reference

This file tracks the public Lua `lumina` module as registered in `pkg/lumina/lumina.go` (`luaLoader`), plus submodules attached to that table. When in doubt, the Go loader is the source of truth.

## Table of Contents

- [Module index](#module-index)
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
- [lumina/ui Components](#lumina-ui-components)

---

## Module index

The return value of `require("lumina")` is a single table. Most functions also exist on that table; hooks are registered twice — as `lumina.useState` etc. and (subset) as `lumina.hooks.useState` for compatibility.

**Helpers:** `version`, `echo`, `info`

**App lifecycle:** `defineComponent`, `createComponent`, `createElement`, `createErrorBoundary`, `memo`, `createPortal`, `forwardRef`, `lazy`, `render`, `mount`, `run`, `quit`, `onKey`, `getSize`, `createState`

**Styling & built-in UI:** `defineStyle`, `defineGlobalStyles`, `getStyle`, `defineTheme`, `setTheme`, `Select`, `Checkbox`, `Menu`, `TextField`

**Events, focus, keyboard:** `on`, `onCapture`, `off`, `emit`, `registerShortcut`, `setFocus`, `getFocused`, `isFocused`, `isHovered`, `focusNext`, `focusPrev`, `registerFocusable`, `unregisterFocusable`, `isFocusable`, `getFocusableIDs`, `emitKeyEvent`, `pushFocusScope`, `popFocusScope`

**Output / MCP / debug tooling:** `setOutputMode`, `getOutputMode`, `getMCPFrame`, `createComponentRequest`, `createEventNotification`, `inspect`, `inspectTree`, `inspectComponent`, `inspectStyles`, `inspectFrames`, `getState`, `getAllComponents`, `simulate`, `simulateClick`, `simulateKey`, `simulateChange`, `consoleLog`, `consoleGet`, `consoleGetErrors`, `consoleClear`, `consoleSize`, `diff`, `diffFrames`, `patch`, `eval`, `profile`, `profileReset`, `profileSize` — and the **`lumina.console`** sub-table (`log`, `warn`, `error`, `get`, `clear`, `size`) plus **`lumina.debug`** (see `RegisterDebugAPI` in code).

**Async / time:** `useAsync`, `delay`

**Viewport & scroll:** `scrollTo`, `scrollToBottom`, `scrollToTop`, `scrollBy`, `getScrollInfo`, `setScrollBehavior`

**Overlays & hot reload:** `showOverlay`, `hideOverlay`, `isOverlayVisible`, `toggleOverlay`, `enableHotReload`, `disableHotReload`

**Router:** `createRouter`, `navigate`, `back`, `useRoute`, `getCurrentPath`

**Canvas (sub-pixel / Braille):** `createCanvas`

**Text input helpers:** `setInputValue`, `getInputValue`, `registerInput`, `focusInput`

**Tiling window manager (in-app):** `createWindow`, `closeWindow`, `focusWindow`, `moveWindow`, `resizeWindow`, `minimizeWindow`, `maximizeWindow`, `restoreWindow`, `tileWindows`, `getFocusedWindow`, `getWindow`, `listWindows`

**Virtual list:** `createVirtualList`

**Submodules registered on the same loader:** `lumina.hooks` (subset of hook functions), `lumina.animation` (preset-related), `lumina.i18n`, `lumina.devtools`, `lumina.Suspense`, `lumina.Profiler`, `lumina.StrictMode`, `lumina.console`, `lumina.debug`.

**Everything below** documents the most common entries in prose; the index above lists every **top-level** function on the `lumina` table from `luaLoader`.

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
    render = function(props)        -- required: return VNode tree
        return { type = "text", content = "Hello" }
    end,
})
```

### `lumina.getSize() → width, height`

Returns the current terminal size in **cells** (from the active `App`), or defaults `80, 24` if no app is running.

### `lumina.quit()`

Request a clean exit from the interactive `lumina.run()` loop.

### `lumina.createElement(component, props) → VNode`

Create a virtual element from a component and props.

```lua
lumina.createElement(MyComponent, { initial = 5, children = {...} })
```

### `lumina.mount(component)`

Register the root **component factory** for `lumina.run()` (see examples — typically a table from `defineComponent`).

### `lumina.run()`

Start the local terminal app (the Go runtime loads the script, sets up the terminal, and runs the event loop).

### `lumina.serve(port)` / `lumina.serveBackground(port)`

- **`serve`** — start the HTTP + WebSocket + embedded web UI; **blocks** the Lua thread (intended for “terminal or browser” entrypoints).
- **`serveBackground`** — start the same server in the background; returns an address string (used heavily in tests).

Documented in more detail under [Web Runtime](#web-runtime).

### `lumina.render(component, props) → VNode`

Render a component to a VNode tree (handy in tests or headless use).

---

## Hooks

All hooks are called inside a component's `render` function. The same functions are available as **`lumina.useState`**, **`lumina.useEffect`**, etc., and a subset is duplicated under **`lumina.hooks`** (`lumina.hooks.useState`, …).

### `lumina.useState(stateKey, initialValue?) → value, setter`

Local state stored **per component instance** under a string key (required). This mirrors React’s rules of hooks: the key identifies which slot in the component’s state map this call uses.

```lua
local count, setCount = lumina.useState("count", 0)
setCount(count + 1)
```

The setter accepts any new value; there is no separate functional-updater form in the Go binding — compute the next value in Lua and pass it in.

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

### Special building blocks

| API | Description |
|-----|-------------|
| `lumina.Suspense` | Factory table: shows `fallback` while a `lumina.lazy` child is `pending` |
| `lumina.createErrorBoundary({…})` | Returns an error-boundary **factory** (see E2E tests) — not a global `lumina.ErrorBoundary` constant |
| `lumina.createPortal(vnode, targetId)` | Portal VNode (see `luaCreatePortal`) |
| `{ type = "fragment", children = {…} }` | Native fragment — **no** `lumina.Fragment` symbol is required; use the `type` string |
| `lumina.Suspense` / `lumina.Profiler` / `lumina.StrictMode` | Small factory tables registered on the module (see `registerProfilerFactory` / `registerStrictModeFactory`) |

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

Lumina uses a **process-wide** `globalRouter` (see `pkg/lumina/router.go`). There is no per-instance `router.navigate` in Lua; instead you call **`lumina.navigate`**, **`lumina.back`**, and read state with **`lumina.useRoute`** / **`lumina.getCurrentPath`**.

### `lumina.createRouter(opts?) → handle`

Build the route table and optional initial path. **Routes are path patterns only** (e.g. `/users/:id`); the Lua API does not bind a component per route in Go — your app’s `render` usually switches on `useRoute()`.

```lua
lumina.createRouter({
    routes = {
        { path = "/" },
        { path = "/users/:id" },
    },
    initialPath = "/",
})
-- returns a small table, e.g. { routeCount = n }
```

### `lumina.navigate(path)`

Set the current path, parse params (e.g. `:id`), push history, and mark **all** components dirty for a re-render.

### `lumina.back() → bool`

Pop history if possible. Returns `true` on success, `false` if there is nothing to go back to; on success, components are marked dirty.

### `lumina.getCurrentPath() → string`

Return the current path string.

### `lumina.useRoute() → { path, params }`

```lua
local route = lumina.useRoute()
-- route.path, route.params (e.g. params.id for "/users/42")
```

There is **no** `query` sub-table in the returned route object today; only `path` and `params` are populated in `useRoute` (`hooks.go`).

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

Lumina’s flex engine (`computeFlexLayout` in `layout.go`) recognizes **`fragment`**, **`text`**, **`vbox`**, **`hbox`**, and treats any other `type` (e.g. **`box`**) as a **generic block** that lays out like a **vertical** box. There is **no** dedicated `type = "grid"` or CSS-style **`type = "flex"`** implementation — if you use those as `type`, they still fall into the default branch and behave like a **vbox**-style column stack, **not** browser CSS flex/grid.

### VNode types you should use

| Type | Description |
|------|-------------|
| `text` | Text; height from wrapping / `content` |
| `input` / `textarea` | Form controls (see text-input APIs) |
| `hbox` | Horizontal flow of children |
| `vbox` | Vertical flow of children |
| `box` | Generic block (vertical stack, borders/background) |
| `fragment` | Transparent pass-through; no box border of its own |

**Scrolling:** on `vbox` / `hbox` / default container, set `style.overflow = "scroll"` and give the node a stable `props.id` to pair with the viewport APIs (`scrollTo`, `getScrollInfo`, …).

### Row / column example (real flexbox for TUI)

```lua
{
    type = "hbox",
    style = { gap = 1, flex = 1 },
    children = {
        { type = "text", content = "A", style = { flex = 1 } },
        { type = "text", content = "B", style = { flex = 2 } },
    }
}
```

### Style Properties

| Property | Type | Description |
|----------|------|-------------|
| `width` | int | Fixed width |
| `height` | int | Fixed height |
| `padding` | int | Padding (all sides) |
| `border` | string | `"single"` \| `"double"` \| `"rounded"` (and `"none"`-like omissions) — no separate `"bold"` border style in the renderer |
| `background` | string | Background color (hex); empty = transparent in many paths |
| `foreground` | string | Text / border (container `foreground` also drives border line color) |
| `bold` | bool | Bold text |
| `underline` | bool | Underlined text |
| `dim` | bool | Dimmed text |
| `flex` | int | Flex grow; containers without fixed height/width get implicit `flex=1` in some cases (see `layout.go`) |
| `gap` / `padding*` / `margin*` | int | Spacing (see `Style` in `layout.go`) |
| `justify` | string | **Main-axis** distribution for boxes: `start` (default), `center`, `end`, `space-between`, `space-around` — not the CSS `flex-start` / `flex-end` spellings |
| `align` | string | **Cross-axis** (e.g. hbox row): `stretch` (default), `start`, `center`, `end` |
| `overflow` | string | `hidden` (default) or `scroll` (viewport + scrollbar) |
| `zIndex` | int | Stacking for positioned children |
| `position` | string | `""` / `relative` (default), `absolute` (offset from parent), `fixed` (offset from **terminal** 0,0; rendering still obeys parent clip — see engine notes) |

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

### `lumina.fetch(url) → body, err`

Synchronous HTTP GET. Returns the response **body** as a string, or `nil` and an error string. Intended for use **inside** `useQuery` fetcher functions (Go implementation in `lua_fetch.go`).

### `lumina.useFetch(url) → { data, loading, error }`

Returns a **single table** with keys `data`, `loading`, and `error` (not three separate return values). Results are also cached in the query cache (see `Fetch` in `pkg/lumina`).

```lua
local s = lumina.useFetch("https://example.com/api")
-- s.data, s.loading, s.error
```

### `lumina.useQuery(key, fetcherFn, opts?) → { data, loading, error }`

Cached data fetching. The returned table has **`data`**, **`loading`**, and **`error`** (there is no `refetch` function in the current `luaUseQuery` implementation). `opts` is optional: `{ staleTime = <seconds> }` (default **60** seconds in Go).

```lua
local r = lumina.useQuery("users", function()
    local body, err = lumina.fetch("https://example.com/api/users")
    if err then
        return nil
    end
    return body
end, { staleTime = 60 })
```

The query fetcher is invoked with `PCall(0, 1)` — only the **first** return value is read as `data` (returning `body, err` as two results would be ignored for the second value).

**Invalidation:** `lumina.invalidateQuery(key)`, `lumina.invalidateAllQueries()`.

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

### `lumina.devtools` submodule

Exposes the Lua-side DevTools helpers (`registerDevToolsModule` in `lua_devtools.go`), including `enable`, `disable`, `toggle`, `isVisible`, `getTree`, `selectElement`, and helpers that mirror the in-process inspector. In interactive apps, **F12** is wired in `app.go` to toggle the inspector and DevTools render path.

### `lumina.inspect` / `lumina.inspectTree` / … (MCP)

The **`inspect`**, **`inspectComponent`**, **`getState`**, **`diff`**, **`patch`**, **`eval`**, **`simulateClick`**, etc. entries on the top-level `lumina` table are the **MCP / automation** surface used by the debug server; they are listed in the [module index](#module-index) above.

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

Headless VNode test utilities (`pkg/lumina/lua_accessibility.go`). The renderer does **not** run full `defineComponent` trees — you pass a **plain VNode table** to `render`.

```lua
local renderer = lumina.createTestRenderer()
renderer.render({ type = "text", content = "Hello" })
local text = renderer.tostring()  -- snapshot-style string
```

### Test renderer methods

| Method | Description |
|--------|-------------|
| `renderer.render(vnodeTable)` | Build the internal test tree from a VNode-like table |
| `renderer.getByText(text) → node` | Find a node |
| `renderer.getByRole(role) → node` | Find by ARIA `role` |
| `renderer.getByType(vnodeType) → node` | Find by `type` string |
| `renderer.fireEvent(target, eventType)` | Fire a test event (string target + type) |
| `renderer.tostring() → string` | Serialize the current tree to a string |
| `renderer.reset()` | Clear the root |

**Note:** `lumina.renderToString` is **not** exported on the module; use `renderer.tostring()` from the test renderer.

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

### Server mode

```lua
lumina.serve(8080)          -- starts HTTP + WebSocket + embedded UI; blocks
lumina.serveBackground(0)   -- non-blocking; port 0 = choose a free port (see tests)
```

`serveBackground` returns an **address** string (e.g. `http://127.0.0.1:port`) when successful. The browser page uses xterm.js and connects to the same Go backend as the local terminal flavor.

### Protocol

- **Browser → Server**: Raw bytes (keyboard input) or JSON `{"type":"resize","cols":80,"rows":24}`
- **Server → Browser**: Raw bytes (terminal escape sequences)

### Embedded Assets

Web assets (HTML, JS) are embedded in the binary via `//go:embed`, so the binary is fully self-contained.

---

## lumina/ui Components

Shadcn-style terminal components live under **`pkg/lumina/components/ui/`**. After `lumina.Open(L)` (or `require("lumina")` with the global opener), Go registers **`package.preload`** entries from `RegisterUI` in `pkg/lumina/ui_components.go`.

**Preferred entrypoint**

```lua
local ui = require("lumina.ui")
-- ui.Button, ui.Card, … (see components/ui/init.lua)
```

**Piecemeal import** (smaller scripts, same files):

```lua
local Button = require("lumina.ui.button")
```

**Legacy aliases** (`RegisterShadcn` in `pkg/lumina/shadcn.go`): `require("shadcn")` loads the **same** aggregate table as `require("lumina.ui")` (both point at `components/ui/init.lua`). You can also `require("shadcn.button")` etc.; the canonical names are **`lumina.ui.*`**.

There is **no** bare `require("ui")` preload — use **`lumina.ui`**.

### Export table (`require("lumina.ui")`)

| Field on `ui` | Piecemeal `require` | Lua file |
|-----------------|----------------------|----------|
| `Button` | `lumina.ui.button` | `button.lua` |
| `Badge` | `lumina.ui.badge` | `badge.lua` |
| `Card` | `lumina.ui.card` | `card.lua` |
| `Alert` | `lumina.ui.alert` | `alert.lua` |
| `Label` | `lumina.ui.label` | `label.lua` |
| `Separator` | `lumina.ui.separator` | `separator.lua` |
| `Skeleton` | `lumina.ui.skeleton` | `skeleton.lua` |
| `Spinner` | `lumina.ui.spinner` | `spinner.lua` |
| `Avatar` | `lumina.ui.avatar` | `avatar.lua` |
| `Breadcrumb` | `lumina.ui.breadcrumb` | `breadcrumb.lua` |
| `Kbd` | `lumina.ui.kbd` | `kbd.lua` |
| `Input` | `lumina.ui.input` | `input.lua` |
| `Switch` | `lumina.ui.switch` | `switch.lua` |
| `Progress` | `lumina.ui.progress` | `progress.lua` |
| `Accordion` | `lumina.ui.accordion` | `accordion.lua` |
| `Tabs` | `lumina.ui.tabs` | `tabs.lua` |
| `Table` | `lumina.ui.table` | `table.lua` |
| `Pagination` | `lumina.ui.pagination` | `pagination.lua` |
| `Toggle` | `lumina.ui.toggle` | `toggle.lua` |
| `ToggleGroup` | `lumina.ui.toggle_group` | `toggle_group.lua` |
| `Select` | `lumina.ui.select` | `select.lua` |
| `Checkbox` | `lumina.ui.checkbox` | `checkbox.lua` |
| `RadioGroup` | `lumina.ui.radio_group` | `radio_group.lua` |
| `Slider` | `lumina.ui.slider` | `slider.lua` |
| `Textarea` | `lumina.ui.textarea` | `textarea.lua` |
| `Field` | `lumina.ui.field` | `field.lua` |
| `InputGroup` | `lumina.ui.input_group` | `input_group.lua` |
| `InputOTP` | `lumina.ui.input_otp` | `input_otp.lua` |
| `Combobox` | `lumina.ui.combobox` | `combobox.lua` |
| `NativeSelect` | `lumina.ui.native_select` | `native_select.lua` |
| `Form` | `lumina.ui.form` | `form.lua` |
| `Command` | `lumina.ui.command` | `command.lua` |
| `Menubar` | `lumina.ui.menubar` | `menubar.lua` |
| `ScrollArea` | `lumina.ui.scroll_area` | `scroll_area.lua` |
| `Collapsible` | `lumina.ui.collapsible` | `collapsible.lua` |
| `Carousel` | `lumina.ui.carousel` | `carousel.lua` |
| `Sonner` | `lumina.ui.sonner` | `sonner.lua` |
| `Dialog` | `lumina.ui.dialog` | `dialog.lua` |
| `AlertDialog` | `lumina.ui.alert_dialog` | `alert_dialog.lua` |
| `Sheet` | `lumina.ui.sheet` | `sheet.lua` |
| `Drawer` | `lumina.ui.drawer` | `drawer.lua` |
| `DropdownMenu` | `lumina.ui.dropdown_menu` | `dropdown_menu.lua` |
| `ContextMenu` | `lumina.ui.context_menu` | `context_menu.lua` |
| `Popover` | `lumina.ui.popover` | `popover.lua` |
| `Tooltip` | `lumina.ui.tooltip` | `tooltip.lua` |
| `HoverCard` | `lumina.ui.hover_card` | `hover_card.lua` |
| `AspectRatio` | `lumina.ui.aspect_ratio` | `aspect_ratio.lua` |
| `ButtonGroup` | `lumina.ui.button_group` | `button_group.lua` |
| `Calendar` | `lumina.ui.calendar` | `calendar.lua` |
| `DatePicker` | `lumina.ui.date_picker` | `date_picker.lua` |
| `NavigationMenu` | `lumina.ui.navigation_menu` | `navigation_menu.lua` |
| `Resizable` | `lumina.ui.resizable` | `resizable.lua` |
| `Sidebar` | `lumina.ui.sidebar` | `sidebar.lua` |
| `Chart` | `lumina.ui.chart` | `chart.lua` |
| `DataTable` | `lumina.ui.data_table` | `data_table.lua` |
| `ColorPicker` | `lumina.ui.color_picker` | `color_picker.lua` |

### Usage

`lumina.ui.button` (and most single-file widgets) return **one** component factory. `lumina.ui.card` returns a **table** of factories (`Card`, `CardHeader`, `CardTitle`, …) — use `ui.Card.Card` when going through the aggregate `require("lumina.ui")`, or destructure:

```lua
local ui = require("lumina.ui")

lumina.createElement(ui.Button, {
    label = "Click me",
    variant = "default",
    size = "default",
    disabled = false,
    onClick = function() print("clicked") end,
})

lumina.createElement(ui.Card.Card, {
    children = {
        lumina.createElement(ui.Card.CardTitle, { children = {{ type = "text", content = "Title" }} }),
        lumina.createElement(ui.Card.CardContent, { children = {{ type = "text", content = "Content" }} }),
    }
})
```
