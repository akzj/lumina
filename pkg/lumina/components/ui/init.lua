-- ui component library — init.lua
-- Exports all ui-style terminal UI components.
local ui = {}

-- Phase 21: Basic components
ui.Button      = require("lumina.ui.button")
ui.Badge       = require("lumina.ui.badge")
ui.Card        = require("lumina.ui.card")
ui.Alert       = require("lumina.ui.alert")
ui.Label       = require("lumina.ui.label")
ui.Separator   = require("lumina.ui.separator")
ui.Skeleton    = require("lumina.ui.skeleton")
ui.Spinner     = require("lumina.ui.spinner")
ui.Avatar      = require("lumina.ui.avatar")
ui.Breadcrumb  = require("lumina.ui.breadcrumb")
ui.Kbd         = require("lumina.ui.kbd")
ui.Input       = require("lumina.ui.input")
ui.Switch      = require("lumina.ui.switch")
ui.Progress    = require("lumina.ui.progress")
ui.Accordion   = require("lumina.ui.accordion")
ui.Tabs        = require("lumina.ui.tabs")
ui.Table       = require("lumina.ui.table")
ui.Pagination  = require("lumina.ui.pagination")
ui.Toggle      = require("lumina.ui.toggle")
ui.ToggleGroup = require("lumina.ui.toggle_group")

-- Phase 22: Form components
ui.Select       = require("lumina.ui.select")
ui.Checkbox     = require("lumina.ui.checkbox")
ui.RadioGroup   = require("lumina.ui.radio_group")
ui.Slider       = require("lumina.ui.slider")
ui.Textarea     = require("lumina.ui.textarea")
ui.Field        = require("lumina.ui.field")
ui.InputGroup   = require("lumina.ui.input_group")
ui.InputOTP     = require("lumina.ui.input_otp")
ui.Combobox     = require("lumina.ui.combobox")
ui.NativeSelect = require("lumina.ui.native_select")
ui.Form         = require("lumina.ui.form")

-- Phase 23: Complex components
ui.Command      = require("lumina.ui.command")
ui.Menubar      = require("lumina.ui.menubar")
ui.ScrollArea   = require("lumina.ui.scroll_area")
ui.Collapsible  = require("lumina.ui.collapsible")
ui.Carousel     = require("lumina.ui.carousel")
ui.Sonner       = require("lumina.ui.sonner")

-- Phase 23: Overlay components
ui.Dialog        = require("lumina.ui.dialog")
ui.AlertDialog   = require("lumina.ui.alert_dialog")
ui.Sheet         = require("lumina.ui.sheet")
ui.Drawer        = require("lumina.ui.drawer")
ui.DropdownMenu  = require("lumina.ui.dropdown_menu")
ui.ContextMenu   = require("lumina.ui.context_menu")
ui.Popover       = require("lumina.ui.popover")
ui.Tooltip       = require("lumina.ui.tooltip")
ui.HoverCard     = require("lumina.ui.hover_card")

-- Phase 38: Additional components
ui.AspectRatio     = require("lumina.ui.aspect_ratio")
ui.ButtonGroup     = require("lumina.ui.button_group")
ui.Calendar        = require("lumina.ui.calendar")
ui.DatePicker      = require("lumina.ui.date_picker")
ui.NavigationMenu  = require("lumina.ui.navigation_menu")
ui.Resizable       = require("lumina.ui.resizable")
ui.Sidebar         = require("lumina.ui.sidebar")
ui.Chart           = require("lumina.ui.chart")
ui.DataTable       = require("lumina.ui.data_table")
ui.ColorPicker     = require("lumina.ui.color_picker")
return ui
