test.describe("MoveFinal", function()
    local app
    test.beforeEach(function()
        app = test.createApp(80, 24)
        app:loadFile("../examples/windows_widget.lua")
    end)
    test.afterEach(function() app:destroy() end)

    test.it("close Palette then move Editor", function()
        app:click(37, 4)
        test.assert.eq(app:screenContains("Palette"), false)
        
        app:mouseDown(10, 2)
        app:mouseMove(30, 10)
        app:mouseUp(30, 10)
        
        local s = app:screenText()
        local editorPos = s:find("Editor")
        test.log("Editor found at pos: " .. tostring(editorPos))
        if editorPos then
            local before = s:sub(1, editorPos - 1)
            local lines = select(2, before:gsub("\n", ""))
            test.log("Editor at line: " .. tostring(lines))
            test.assert.eq(lines > 5, true)
        else
            test.assert.eq(false, true)
        end
    end)
end)
