# Widget Architecture: Lua-First

## Principle
> Go provides mechanisms (layout, paint, events, layers, focus, cursor).
> Lua provides policy (appearance, interaction, theming).

## Widget Layers

### Lua Lux Components (Preferred)
All new UI components should be implemented in `lua/lux/` as pure Lua components.
They use `lumina.createElement("vbox"/"hbox"/"text"/...)` with CSS-style properties.

Benefits:
- Hot-reloadable
- Theme-customizable
- Composable via slots
- No Go rebuild required

### Go Widgets (Capability Layer)
Go widgets (`pkg/widget/`) should only be used when Lua cannot provide the needed capability:
- **Layer management**: Select, Dropdown, Tooltip, Window (create overlays)
- **Complex input**: TextInput (cursor, selection, IME)
- **Virtual scrolling**: Table, ScrollView (performance-critical)
- **System integration**: Menu (OS-level menus)

### Removed Go Widgets
The following Go widgets have been fully removed. Use the Lua replacements:
- `lumina.Button` → `require("lux.button")`
- `lumina.Dialog` → `require("lux.dialog")`
- `lumina.Toast` → `require("lux.toast")`
- `lumina.List` → `require("lux.list")`
- `lumina.Pagination` → `require("lux.pagination")`
- `lumina.Checkbox` → `require("lux.checkbox")`
- `lumina.Switch` → `require("lux.switch")`
- `lumina.Radio` → `require("lux.radio")`
- `lux.Dropdown` → use `lumina.Dropdown` (Go widget) directly (no lux wrapper)

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
