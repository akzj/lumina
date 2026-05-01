-- engine/events_keyboard_test.lua — Tests for keyboard event handling

test.describe("Engine: Keyboard Events", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("onKeyDown fires on focused node", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local key, setKey = lumina.useState("key", "none")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="target",
                        style={width=20, height=3},
                        focusable=true,
                        onKeyDown=function(e)
                            local k = type(e) == "table" and (e.key or e.Key) or tostring(e or "")
                            setKey(k)
                        end},
                        lumina.createElement("text", {}, "Focus me")
                    ),
                    lumina.createElement("text", {id="result"}, "Key:" .. key)
                )
            end})
        ]])
        -- Click to focus the target
        app:click(5, 1)
        -- Press a key
        app:keyPress("x")
        test.assert.eq(app:screenContains("Key:x"), true)
    end)

    test.it("keyboard event bubbles to parent when child has no handler", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local key, setKey = lumina.useState("key", "none")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24},
                    onKeyDown=function(e)
                        local k = type(e) == "table" and (e.key or e.Key) or tostring(e or "")
                        setKey(k)
                    end},
                    lumina.createElement("vbox", {id="child",
                        style={width=20, height=3},
                        focusable=true},
                        lumina.createElement("text", {}, "No handler")
                    ),
                    lumina.createElement("text", {id="result"}, "Key:" .. key)
                )
            end})
        ]])
        -- Click to focus child (which has no onKeyDown)
        app:click(5, 1)
        -- Key should bubble up to parent
        app:keyPress("z")
        test.assert.eq(app:screenContains("Key:z"), true)
    end)

    test.it("global keys handler receives keydown", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local pressed, setPressed = lumina.useState("pressed", "no")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24},
                    onKeyDown=function(e)
                        local k = type(e) == "table" and (e.key or e.Key) or tostring(e or "")
                        if k == "Enter" then setPressed("yes") end
                    end},
                    lumina.createElement("vbox", {id="focusable",
                        style={width=20, height=3}, focusable=true},
                        lumina.createElement("text", {}, "Press Enter")
                    ),
                    lumina.createElement("text", {id="result"}, "Pressed:" .. pressed)
                )
            end})
        ]])
        app:click(5, 1)
        app:keyPress("Enter")
        test.assert.eq(app:screenContains("Pressed:yes"), true)
    end)
end)
