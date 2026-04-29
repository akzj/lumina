-- lua/lux/init.lua
-- Lux: component library for Lumina
-- Usage: local lux = require("lux")

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

return M
