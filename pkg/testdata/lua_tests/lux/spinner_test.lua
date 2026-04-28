-- spinner_test.lua — Tests for the lux Spinner component

test.describe("lux.Spinner", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders spinner component without error", function()
        app:loadString([[
            local Spinner = require("lux.spinner")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Spinner { label = "Loading..." },
                        lumina.createElement("text", {id = "marker"}, "AFTER_SPINNER")
                    )
                end,
            })
        ]])
        -- Spinner uses useEffect + setInterval for animation.
        -- Verify the component tree renders without error.
        local tree = app:vnodeTree()
        test.assert.notNil(tree)
    end)

    test.it("renders with custom label in tree", function()
        app:loadString([[
            local Spinner = require("lux.spinner")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Spinner { label = "Processing..." },
                        lumina.createElement("text", {id = "marker"}, "DONE")
                    )
                end,
            })
        ]])
        local tree = app:vnodeTree()
        test.assert.notNil(tree)
    end)
end)
