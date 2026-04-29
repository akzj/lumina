-- lua/lux/wm.lua
-- WindowManager module for Lux
--
-- Manages multiple overlapping windows with stable id-based access.
-- State is stored in lumina.store for reactivity.
--
-- Usage:
--   local WM = require("lux.wm")
--   local mgr = WM.create("wm_state", {
--       { id = "editor", title = "Editor", x = 2, y = 1, w = 35, h = 12 },
--       { id = "palette", title = "Palette", x = 10, y = 3, w = 30, h = 10 },
--   })
--
-- In render:
--   local windows = mgr.getWindows()   -- ordered bottom-to-top, only open windows
--   for _, win in ipairs(windows) do ... end
--
-- In onChange:
--   mgr.activate(winId)                -- bring to front
--   mgr.setFrame(winId, { x=e.x, y=e.y })  -- update position
--   mgr.close(winId)                   -- close (preserves frame for reopen)

local M = {}

--- Create a new WindowManager instance backed by lumina.store.
--- Idempotent: if the store key already has data, it is not overwritten.
---
--- @param storeKey string  Key in lumina.store to store WM state
--- @param initialWindows table  Array of { id, title, x, y, w, h } initial window definitions
--- @return table  Manager object with methods: register, close, reopen, activate, setFrame, getWindows, getActiveId
function M.create(storeKey, initialWindows)
    -- Only initialize if store doesn't have this key yet (idempotent)
    local existing = lumina.store.get(storeKey)
    if existing == nil then
        local state = { order = {}, frames = {}, activeId = nil }
        for _, win in ipairs(initialWindows) do
            state.frames[win.id] = {
                x = win.x, y = win.y,
                w = win.w, h = win.h,
                title = win.title,
                open = true,
            }
            state.order[#state.order + 1] = win.id
        end
        -- activeId = last window (topmost)
        if #state.order > 0 then
            state.activeId = state.order[#state.order]
        end
        lumina.store.set(storeKey, state)
    end

    local mgr = {}

    --- Register a new window at the top of the z-order.
    function mgr.register(id, frame)
        local s = lumina.store.get(storeKey)
        s.frames[id] = {
            x = frame.x or 0, y = frame.y or 0,
            w = frame.w or 30, h = frame.h or 10,
            title = frame.title or id,
            open = true,
        }
        s.order[#s.order + 1] = id
        s.activeId = id
        lumina.store.set(storeKey, s)
    end

    --- Close a window: remove from order, mark as closed, preserve frame.
    function mgr.close(id)
        local s = lumina.store.get(storeKey)
        -- Remove from order
        for i, oid in ipairs(s.order) do
            if oid == id then
                table.remove(s.order, i)
                break
            end
        end
        -- Mark closed but keep frame (remember position)
        if s.frames[id] then
            s.frames[id].open = false
        end
        -- Update activeId if the closed window was active
        if s.activeId == id then
            s.activeId = s.order[#s.order] or nil
        end
        lumina.store.set(storeKey, s)
    end

    --- Reopen a previously closed window at the top of the z-order.
    function mgr.reopen(id)
        local s = lumina.store.get(storeKey)
        if not s.frames[id] then return end
        s.frames[id].open = true
        s.order[#s.order + 1] = id
        s.activeId = id
        lumina.store.set(storeKey, s)
    end

    --- Activate (bring to front): move to top of z-order, set as active.
    --- Should only be called on mousedown/"activate" events, NOT on every move/resize.
    function mgr.activate(id)
        local s = lumina.store.get(storeKey)
        -- Remove from current position in order
        for i, oid in ipairs(s.order) do
            if oid == id then
                table.remove(s.order, i)
                break
            end
        end
        -- Append to end (top of z-order)
        s.order[#s.order + 1] = id
        s.activeId = id
        lumina.store.set(storeKey, s)
    end

    --- Update a window's frame (position/size). Does NOT change z-order.
    --- @param id string  Window id
    --- @param patch table  Fields to merge: { x, y, w, h, title, ... }
    function mgr.setFrame(id, patch)
        local s = lumina.store.get(storeKey)
        if not s.frames[id] then return end
        for k, v in pairs(patch) do
            s.frames[id][k] = v
        end
        lumina.store.set(storeKey, s)
    end

    --- Get ordered list of open windows (bottom to top) for rendering.
    --- Uses lumina.useStore for reactivity (subscribes to changes).
    --- @return table  Array of { id, title, x, y, w, h }
    function mgr.getWindows()
        local s = lumina.useStore(storeKey)
        local result = {}
        for _, id in ipairs(s.order) do
            local f = s.frames[id]
            if f and f.open then
                result[#result + 1] = {
                    id = id,
                    title = f.title,
                    x = f.x, y = f.y,
                    w = f.w, h = f.h,
                }
            end
        end
        return result
    end

    --- Get the active window id.
    --- Uses lumina.useStore for reactivity.
    --- @return string|nil  Active window id, or nil if none
    function mgr.getActiveId()
        local s = lumina.useStore(storeKey)
        return s.activeId
    end

    return mgr
end

return M
