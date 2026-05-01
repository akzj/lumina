-- pagination_test.lua — Tests for the Lux Pagination component (rendering)

test.describe("Pagination widget", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("renders without error", function()
        app:loadString([[
            local Pagination = require("lux.pagination")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Pagination {
                            currentPage = 3,
                            pageCount = 10,
                            key = "pg1",
                        },
                        lumina.createElement("text", {id = "marker"}, "END")
                    )
                end,
            })
        ]])
        local tree = app:vnodeTree()
        test.assert.notNil(tree)
        test.assert.eq(app:screenContains("END"), true)
    end)

    test.it("renders single page", function()
        app:loadString([[
            local Pagination = require("lux.pagination")
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Pagination {
                            currentPage = 1,
                            pageCount = 1,
                            key = "pg2",
                        },
                        lumina.createElement("text", {id = "marker"}, "DONE")
                    )
                end,
            })
        ]])
        local tree = app:vnodeTree()
        test.assert.notNil(tree)
        test.assert.eq(app:screenContains("DONE"), true)
    end)

    test.it("updates when page changes via store", function()
        app:loadString([[
            local Pagination = require("lux.pagination")
            lumina.store.set("page", 1)
            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    local page = lumina.useStore("page")
                    return lumina.createElement("vbox", {id = "root",
                        style = {width = 80, height = 24}},
                        Pagination {
                            currentPage = page,
                            pageCount = 5,
                            key = "pg3",
                        },
                        lumina.createElement("text", {id = "out"},
                            "page:" .. tostring(page))
                    )
                end,
            })
        ]])
        test.assert.eq(app:screenContains("page:1"), true)

        -- Change page via store
        app:loadString([[lumina.store.set("page", 4)]])
        app:render()
        test.assert.eq(app:screenContains("page:4"), true)
    end)
end)
