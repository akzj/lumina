-- lua/lux/init.lua
-- Lux: component library for Lumina
-- Usage: local lux = require("lux")
--
-- Architecture: Lua-first. All components are pure Lua implementations using
-- lumina.createElement() with CSS-style properties. Go widgets (lumina.Checkbox,
-- lumina.Switch, etc.) are deprecated in favor of these Lua versions.
-- See docs/WIDGET_MIGRATION.md for details.

local M = {}

M.Card = require("lux.card")
M.Badge = require("lux.badge")
M.Divider = require("lux.divider")
M.Progress = require("lux.progress")
M.Spinner = require("lux.spinner")
M.Dialog = require("lux.dialog")
M.Layout = require("lux.layout")
M.CommandPalette = require("lux.command_palette")
M.ListView = require("lux.list")
M.DataGrid = require("lux.data_grid")
M.WM = require("lux.wm")
M.Pagination = require("lux.pagination")
M.Tabs = require("lux.tabs")
M.Alert = require("lux.alert")
M.Accordion = require("lux.accordion")
M.Breadcrumb = require("lux.breadcrumb")
M.TextInput = require("lux.text_input")
M.Button = require("lux.button")
M.Checkbox = require("lux.checkbox")
M.Radio = require("lux.radio")
M.Switch = require("lux.switch")
-- Dropdown: use lumina.Dropdown (Go widget) directly — no lux wrapper needed
M.Toast = require("lux.toast")
M.Tree = require("lux.tree")
M.Form = require("lux.form")
M.Atlantis = require("lux.atlantis")
M.AtlantisFormDemo = require("lux.atlantis_form_demo")

return M
