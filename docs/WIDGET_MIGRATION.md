# Widget Architecture: Lua-First (Zero Go Widgets)

## Principle
> Go provides mechanisms (layout, paint, events, layers, focus, cursor, animation).
> Lua provides policy (appearance, interaction, theming).

## Architecture

### Engine Primitives (Go)
The Go engine provides only layout primitives — no UI widgets:
- `vbox`, `hbox`, `box` — flex layout containers
- `text` — text rendering
- `input`, `textarea` — text input with cursor/selection
- Style system (CSS-like properties: position, flex, overflow, border, etc.)
- Event system (click, hover, keydown, scroll, focus)
- Layer management (z-order, absolute/fixed positioning)
- Animation hooks (`lumina.useAnimation()`)
- Theme system (`lumina.getTheme()`, `lumina.setTheme()`)

### Lua Lux Components (All UI)
All UI components are implemented in `lua/lux/` as pure Lua components.
They use `lumina.createElement("vbox"/"hbox"/"text"/...)` with CSS-style properties.

Available components:
- `lux.Button` — buttons with hover/press states
- `lux.Checkbox` — checkbox with label
- `lux.Radio` — radio button with label
- `lux.Switch` — toggle switch with label
- `lux.Dialog` — modal dialog
- `lux.Toast` — toast notifications
- `lux.List` — scrollable list with row rendering
- `lux.Pagination` — page navigation
- `lux.Card` — card container with title
- `lux.Badge` — status badge
- `lux.Divider` — horizontal divider
- `lux.Progress` — progress bar
- `lux.Form` — form layout
- `lux.Tree` — tree view

Benefits:
- Hot-reloadable
- Theme-customizable
- Composable via slots
- No Go rebuild required

### Removed Go Widgets
ALL Go widgets have been removed. The following migrations apply:
- `lumina.Button` → `require("lux.button")`
- `lumina.Dialog` → `require("lux.dialog")`
- `lumina.Toast` → `require("lux.toast")`
- `lumina.List` → `require("lux.list")`
- `lumina.Pagination` → `require("lux.pagination")`
- `lumina.Checkbox` → `require("lux.checkbox")`
- `lumina.Switch` → `require("lux.switch")`
- `lumina.Radio` → `require("lux.radio")`
- `lumina.Label` → `lumina.createElement("text", {...}, "content")`
- `lumina.Spacer` → `lumina.createElement("box", {style = {flex = 1}})`
- `lumina.Select` → build with vbox/text + state management
- `lumina.Dropdown` → build with vbox/text + layer management
- `lumina.Table` → build with hbox/vbox/text rows
- `lumina.Menu` → build with vbox/text items
- `lumina.Tooltip` → build with absolute-positioned text
- `lumina.Window` → build with vbox + absolute positioning (see `lua/lux/wm.lua`)
- `lumina.ScrollView` → use `overflow = "scroll"` style on any container

## Creating New Lua Components

See `lua/lux/button.lua` as the reference implementation for:
- Theme integration via `lumina.getTheme()`
- Controlled component pattern (props drive state)
- Local UI state via `lumina.useState()` (hover, pressed)
- Event handling via `onClick`, `onKeyDown`, `onMouseEnter/Leave`
- Style composition via `props.style` merge

### Component Template

```lua
local MyWidget = lumina.defineComponent("LuxMyWidget", function(props)
    local t = lumina.getTheme and lumina.getTheme() or {}
    local hovered, setHovered = lumina.useState("hover", false)

    -- Derive visual state from props (controlled pattern)
    local disabled = props.disabled == true

    return lumina.createElement("hbox", {
        id = props.id,
        key = props.key,
        focusable = not disabled,
        disabled = disabled,
        style = props.style,
        onClick = not disabled and function()
            if props.onChange then props.onChange(newValue) end
        end or nil,
        onKeyDown = not disabled and function(e)
            local k = type(e) == "table" and (e.key or e.Key) or tostring(e or "")
            if k == " " or k == "Enter" then
                if props.onChange then props.onChange(newValue) end
            end
        end or nil,
        onMouseEnter = not disabled and function() setHovered(true) end or nil,
        onMouseLeave = not disabled and function() setHovered(false) end or nil,
    },
        lumina.createElement("text", {foreground = t.primary or "#89B4FA"}, "content")
    )
end)

return MyWidget
```

### Key Patterns

1. **Controlled state**: Widget has no internal state for the value. Parent passes
   `checked`/`value`/`selected` as props, and `onChange` callback to update.

2. **Local UI state**: Only cosmetic state (hover, pressed, animation frame) uses
   `lumina.useState()`. This state is ephemeral and doesn't affect parent logic.

3. **Theme integration**: Always read theme at render time via `lumina.getTheme()`.
   Never hardcode colors — use theme values with fallback defaults.

4. **Keyboard accessibility**: Always handle Space and Enter for toggle/action widgets.
   Use `onKeyDown` with key string parsing.

5. **Disabled state**: When `disabled = true`, don't attach event handlers (`nil`),
   set `focusable = false`, and use muted theme colors.
