-- test_log_test.lua — Verify test.log() API works

test.describe("test.log API", function()
    test.it("test.log outputs to test results", function()
        test.log("hello from test.log")
        test.log("multiple", "args", 42)
        test.assert.eq(1, 1)
    end)

    test.it("test.log works with screenText", function()
        local app = test.createApp(40, 10)
        app:loadFile("../examples/notes.lua")
        test.log("Screen dump:")
        test.log(app:screenText())
        test.assert.eq(app:screenContains("Notes"), true)
        app:destroy()
    end)
end)
