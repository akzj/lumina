-- shadcn component library — init.lua
-- Exports all shadcn-style terminal UI components.
local shadcn = {}

shadcn.Button      = require("shadcn.button")
shadcn.Badge       = require("shadcn.badge")
shadcn.Card        = require("shadcn.card")  -- returns table: Card, CardHeader, CardTitle, etc.
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

return shadcn
