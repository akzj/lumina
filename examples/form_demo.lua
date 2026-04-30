lumina.app {
    id = "form-demo",
    store = {
        values = { name = "", email = "", terms = false },
        errors = {},
        submitted = false,
    },
    keys = {
        ["q"] = function() lumina.quit() end,
        ["ctrl+c"] = function() lumina.quit() end,
    },
    render = function()
        local lux = require("lux")
        local Form = lux.Form
        local Alert = lux.Alert
        local values = lumina.useStore("values")
        local errors = lumina.useStore("errors")
        local submitted = lumina.useStore("submitted")
        local t = lumina.getTheme()

        local children = {
            lumina.createElement("text", {
                foreground = t.primary or "#89B4FA",
                bold = true,
                style = { height = 1 },
            }, "  Registration Form"),
            lumina.createElement("text", { style = { height = 1 } }, ""),
        }

        if submitted then
            children[#children + 1] = Alert {
                key = "success",
                variant = "success",
                title = "Submitted!",
                message = "Welcome, " .. (values.name or ""),
                width = 40,
            }
        else
            children[#children + 1] = Form {
                key = "reg-form",
                width = 40,
                fields = {
                    { id = "name", type = "text", label = "Name", placeholder = "Your name", required = true },
                    { id = "email", type = "text", label = "Email", placeholder = "you@example.com", required = true },
                    { id = "terms", type = "checkbox", label = "I accept the terms" },
                },
                values = values,
                errors = errors,
                onFieldChange = function(fieldId, value)
                    local v = lumina.store.get("values") or {}
                    v[fieldId] = value
                    lumina.store.set("values", v)
                    -- Clear error on change
                    local e = lumina.store.get("errors") or {}
                    e[fieldId] = nil
                    lumina.store.set("errors", e)
                end,
                onSubmit = function(vals)
                    local errs = {}
                    if not vals.name or vals.name == "" then
                        errs.name = "Name is required"
                    end
                    if not vals.email or vals.email == "" then
                        errs.email = "Email is required"
                    elseif not vals.email:find("@") then
                        errs.email = "Invalid email"
                    end
                    if not vals.terms then
                        errs.terms = "Must accept terms"
                    end
                    if next(errs) then
                        lumina.store.set("errors", errs)
                    else
                        lumina.store.set("submitted", true)
                    end
                end,
                onReset = function()
                    lumina.store.set("values", { name = "", email = "", terms = false })
                    lumina.store.set("errors", {})
                end,
            }
        end

        return lumina.createElement("vbox", {
            id = "root",
            style = { width = 50, height = 20, background = t.base or "#1E1E2E" },
        }, table.unpack(children))
    end,
}
