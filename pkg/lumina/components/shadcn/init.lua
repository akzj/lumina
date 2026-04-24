-- shadcn component library — init.lua
-- Exports all shadcn-style terminal UI components.
local shadcn = {}

-- Phase 21: Basic components
shadcn.Button      = require("shadcn.button")
shadcn.Badge       = require("shadcn.badge")
shadcn.Card        = require("shadcn.card")
shadcn.Alert       = require("shadcn.alert")
shadcn.Label       = require("shadcn.label")
shadcn.Separator   = require("shadcn.separator")
shadcn.Skeleton    = require("shadcn.skeleton")
shadcn.Spinner     = require("shadcn.spinner")
shadcn.Avatar      = require("shadcn.avatar")
shadcn.Breadcrumb  = require("shadcn.breadcrumb")
shadcn.Kbd         = require("shadcn.kbd")
shadcn.Input       = require("shadcn.input")
shadcn.Switch      = require("shadcn.switch")
shadcn.Progress    = require("shadcn.progress")
shadcn.Accordion   = require("shadcn.accordion")
shadcn.Tabs        = require("shadcn.tabs")
shadcn.Table       = require("shadcn.table")
shadcn.Pagination  = require("shadcn.pagination")
shadcn.Toggle      = require("shadcn.toggle")
shadcn.ToggleGroup = require("shadcn.toggle_group")

-- Phase 22: Form components
shadcn.Select       = require("shadcn.select")
shadcn.Checkbox     = require("shadcn.checkbox")
shadcn.RadioGroup   = require("shadcn.radio_group")
shadcn.Slider       = require("shadcn.slider")
shadcn.Textarea     = require("shadcn.textarea")
shadcn.Field        = require("shadcn.field")
shadcn.InputGroup   = require("shadcn.input_group")
shadcn.InputOTP     = require("shadcn.input_otp")
shadcn.Combobox     = require("shadcn.combobox")
shadcn.NativeSelect = require("shadcn.native_select")
shadcn.Form         = require("shadcn.form")

-- Phase 23: Complex components
shadcn.Command      = require("shadcn.command")
shadcn.Menubar      = require("shadcn.menubar")
shadcn.ScrollArea   = require("shadcn.scroll_area")
shadcn.Collapsible  = require("shadcn.collapsible")
shadcn.Carousel     = require("shadcn.carousel")
shadcn.Sonner       = require("shadcn.sonner")

-- Phase 23: Overlay components
shadcn.Dialog        = require("shadcn.dialog")
shadcn.AlertDialog   = require("shadcn.alert_dialog")
shadcn.Sheet         = require("shadcn.sheet")
shadcn.Drawer        = require("shadcn.drawer")
shadcn.DropdownMenu  = require("shadcn.dropdown_menu")
shadcn.ContextMenu   = require("shadcn.context_menu")
shadcn.Popover       = require("shadcn.popover")
shadcn.Tooltip       = require("shadcn.tooltip")
shadcn.HoverCard     = require("shadcn.hover_card")
return shadcn
