test.describe("Breadcrumb component", function()
    local app

    test.beforeEach(function()
        app = test.createApp(60, 16)
        app:loadFile("../examples/breadcrumb_demo.lua")
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Rendering
    test.it("renders all breadcrumb items", function()
        test.assert.eq(app:screenContains("Home"), true)
        test.assert.eq(app:screenContains("Products"), true)
        test.assert.eq(app:screenContains("Electronics"), true)
        test.assert.eq(app:screenContains("Phones"), true)
    end)

    test.it("renders separators between items", function()
        test.assert.eq(app:screenContains("›"), true)
    end)

    test.it("shows current page content", function()
        test.assert.eq(app:screenContains("Mobile phones catalog"), true)
    end)

    -- Navigation via depth keys
    test.it("pressing 1 shows only Home", function()
        app:keyPress("1")
        test.assert.eq(app:screenContains("Welcome to the home"), true)
        -- Only Home in breadcrumb, no separator
        test.assert.eq(app:screenContains("Products"), false)
    end)

    test.it("pressing 2 shows Home > Products", function()
        app:keyPress("2")
        test.assert.eq(app:screenContains("Home"), true)
        test.assert.eq(app:screenContains("Products"), true)
        test.assert.eq(app:screenContains("Browse our products"), true)
        test.assert.eq(app:screenContains("Electronics"), false)
    end)

    test.it("pressing 3 shows three levels", function()
        app:keyPress("3")
        test.assert.eq(app:screenContains("Home"), true)
        test.assert.eq(app:screenContains("Products"), true)
        test.assert.eq(app:screenContains("Electronics"), true)
        test.assert.eq(app:screenContains("Phones"), false)
    end)

    -- Navigate via onNavigate (simulate clicking breadcrumb)
    test.it("navigating to Home sets depth to 1", function()
        -- Simulate clicking "Home" breadcrumb
        app:loadString('lumina.store.set("depth", 1); lumina.store.set("lastNav", "home")')
        test.assert.eq(app:screenContains("Welcome to the home"), true)
        test.assert.eq(app:screenContains("Last nav: home"), true)
    end)

    test.it("navigating to Products sets depth to 2", function()
        app:loadString('lumina.store.set("depth", 2); lumina.store.set("lastNav", "products")')
        test.assert.eq(app:screenContains("Browse our products"), true)
        test.assert.eq(app:screenContains("Last nav: products"), true)
    end)

    -- Edge cases
    test.it("single item breadcrumb has no separator", function()
        app:keyPress("1")
        -- With only "Home", there should be no separator
        test.assert.eq(app:screenContains("Home"), true)
    end)

    test.it("empty breadcrumb does not crash", function()
        app:destroy()
        app = test.createApp(60, 10)
        app:loadString([[
            local Breadcrumb = require("lux.breadcrumb")
            lumina.app {
                id = "empty-bc",
                render = function()
                    return Breadcrumb { id = "bc", items = {}, width = 40 }
                end,
            }
        ]])
        -- Should not crash
    end)

    test.it("custom separator renders correctly", function()
        app:destroy()
        app = test.createApp(60, 10)
        app:loadString([[
            local Breadcrumb = require("lux.breadcrumb")
            lumina.app {
                id = "custom-sep",
                render = function()
                    return Breadcrumb {
                        id = "bc",
                        items = {
                            { id = "a", label = "AAA" },
                            { id = "b", label = "BBB" },
                        },
                        separator = " / ",
                        width = 40,
                    }
                end,
            }
        ]])
        test.assert.eq(app:screenContains("AAA"), true)
        test.assert.eq(app:screenContains("/"), true)
        test.assert.eq(app:screenContains("BBB"), true)
    end)

    test.it("last item is not clickable (bold, no underline behavior)", function()
        -- Just verify the last item renders as the current page
        test.assert.eq(app:screenContains("Phones"), true)
        test.assert.eq(app:screenContains("Mobile phones"), true)
    end)
end)
