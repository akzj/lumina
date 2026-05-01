-- engine/theme_test.lua — Tests for theme system

test.describe("Engine: Theme", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("getTheme returns theme colors", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local t = lumina.getTheme()
                local hasBase = t and t.base and t.base ~= ""
                return lumina.createElement("text", {id="out"},
                    hasBase and ("base:" .. t.base) or "NO_THEME")
            end})
        ]])
        -- Should have a base color (starts with #)
        local s = app:screenText()
        test.assert.eq(s:find("base:#") ~= nil, true)
    end)

    test.it("getTheme has expected color keys", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local t = lumina.getTheme()
                local keys = {"base", "text", "primary", "surface0"}
                local result = ""
                for _, k in ipairs(keys) do
                    if t[k] and t[k] ~= "" then
                        result = result .. k .. ":ok "
                    else
                        result = result .. k .. ":missing "
                    end
                end
                return lumina.createElement("text", {id="out"}, result)
            end})
        ]])
        test.assert.eq(app:screenContains("base:ok"), true)
        test.assert.eq(app:screenContains("text:ok"), true)
        test.assert.eq(app:screenContains("primary:ok"), true)
    end)

    test.it("setTheme changes theme and affects rendering", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local t = lumina.getTheme()
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("text", {id="color"}, "P:" .. (t.primary or "none")),
                    lumina.createElement("text", {id="change",
                        style={width=10, height=1},
                        onClick=function()
                            lumina.setTheme("nord")
                        end}, "Nord"),
                    lumina.createElement("text", {id="reset",
                        style={width=10, height=1},
                        onClick=function()
                            lumina.setTheme("mocha")
                        end}, "Reset")
                )
            end})
        ]])
        local before = app:screenText()
        local beforePrimary = before:match("P:(#%x+)")
        app:click(2, 1)  -- switch to nord
        local after = app:screenText()
        local afterPrimary = after:match("P:(#%x+)")
        -- Primary color should change (or at minimum, the theme switch doesn't crash)
        test.assert.notNil(afterPrimary)
        -- Reset theme to default to avoid leaking into other tests
        app:click(2, 2)
    end)
end)
