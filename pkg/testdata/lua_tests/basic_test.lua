test.describe("Basic Component", function()
    local app

    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadString([[
            lumina.createComponent({
                id = "test",
                name = "Test",
                render = function(props)
                    local selected, setSelected = lumina.useState("sel", 1)
                    return lumina.createElement("vbox", {id = "root"},
                        lumina.createElement("text", {
                            id = "item-1",
                            style = {height = 1, background = selected == 1 and "#ff0" or "#000"},
                            onClick = function() setSelected(1) end,
                        }, "Item 1"),
                        lumina.createElement("text", {
                            id = "item-2",
                            style = {height = 1, background = selected == 2 and "#ff0" or "#000"},
                            onClick = function() setSelected(2) end,
                        }, "Item 2")
                    )
                end,
            })
        ]])
    end)

    test.afterEach(function()
        app:destroy()
    end)

    test.it("initial state: item 1 selected", function()
        local item1 = app:find("item-1")
        test.assert.notNil(item1)
        test.assert.eq(item1.content, "Item 1")
        test.assert.eq(item1.style.background, "#ff0")

        local item2 = app:find("item-2")
        test.assert.eq(item2.style.background, "#000")
    end)

    test.it("click item 2 selects it", function()
        local item2 = app:find("item-2")
        app:click(item2.x + 1, item2.y)

        item2 = app:find("item-2")
        test.assert.eq(item2.style.background, "#ff0")

        local item1 = app:find("item-1")
        test.assert.eq(item1.style.background, "#000")
    end)

    test.it("vnodeTree returns full tree", function()
        local tree = app:vnodeTree()
        test.assert.eq(tree.type, "vbox")
        test.assert.eq(tree.id, "root")
        test.assert.eq(#tree.children, 2)
        test.assert.eq(tree.children[1].content, "Item 1")
        test.assert.eq(tree.children[2].content, "Item 2")
    end)
end)
