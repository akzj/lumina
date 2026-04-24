-- Lumina Component Library — init.lua
-- Exports all built-in components.
local lumina = require("lumina")

local function loadComponent(name)
    local path = "pkg/lumina/components/" .. name .. ".lua"
    if lumina._getComponentPath then
        path = lumina._getComponentPath(name)
    end
    local ok, result = pcall(dofile, path)
    if ok then return result end
    return nil
end

-- Load all components
local Button = loadComponent("button")
local Input = loadComponent("input")
local Dialog = loadComponent("dialog")
local List = loadComponent("list")
local Tabs = loadComponent("tabs")
local Modal = loadComponent("modal")
local Progress = loadComponent("progress")
local Spinner = loadComponent("spinner")
local Table = loadComponent("table")
local Tree = loadComponent("tree")
local StatusBar = loadComponent("statusbar")

return {
    -- Original components
    Button = Button,
    Input = Input,
    Dialog = Dialog,
    -- New components (Phase 8)
    List = List,
    Tabs = Tabs,
    Modal = Modal,
    Progress = Progress,
    Spinner = Spinner,
    Table = Table,
    Tree = Tree,
    StatusBar = StatusBar,
}
