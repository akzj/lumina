-- engine/focus_test.lua — Tests for focus management

test.describe("Engine: Focus", function()
    local app
    test.beforeEach(function() app = test.createApp(80, 24) end)
    test.afterEach(function() app:destroy() end)

    test.it("focusable node receives focus on click", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local focused, setFocused = lumina.useState("focused", false)
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="target",
                        style={width=20, height=3},
                        focusable=true,
                        onFocus=function() setFocused(true) end,
                        onBlur=function() setFocused(false) end},
                        lumina.createElement("text", {}, "Target")
                    ),
                    lumina.createElement("text", {id="result"},
                        focused and "FOCUSED" or "UNFOCUSED")
                )
            end})
        ]])
        test.assert.eq(app:screenContains("UNFOCUSED"), true)
        app:click(5, 1)
        test.assert.eq(app:screenContains("FOCUSED"), true)
    end)

    test.it("Tab moves focus to next focusable", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local which, setWhich = lumina.useState("which", "none")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="a",
                        style={width=20, height=2},
                        focusable=true,
                        onFocus=function() setWhich("A") end},
                        lumina.createElement("text", {}, "Item A")
                    ),
                    lumina.createElement("vbox", {id="b",
                        style={width=20, height=2},
                        focusable=true,
                        onFocus=function() setWhich("B") end},
                        lumina.createElement("text", {}, "Item B")
                    ),
                    lumina.createElement("text", {id="result"}, "Focus:" .. which)
                )
            end})
        ]])
        -- Click to focus A
        app:click(5, 0)
        test.assert.eq(app:screenContains("Focus:A"), true)
        -- Tab to move to B
        app:keyPress("Tab")
        test.assert.eq(app:screenContains("Focus:B"), true)
    end)

    test.it("disabled node is not focusable", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local which, setWhich = lumina.useState("which", "none")
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="disabled",
                        style={width=20, height=2},
                        focusable=true,
                        disabled=true,
                        onFocus=function() setWhich("disabled") end},
                        lumina.createElement("text", {}, "Disabled")
                    ),
                    lumina.createElement("vbox", {id="enabled",
                        style={width=20, height=2},
                        focusable=true,
                        onFocus=function() setWhich("enabled") end},
                        lumina.createElement("text", {}, "Enabled")
                    ),
                    lumina.createElement("text", {id="result"}, "Focus:" .. which)
                )
            end})
        ]])
        -- Tab should skip disabled and focus enabled
        app:keyPress("Tab")
        test.assert.eq(app:screenContains("Focus:enabled"), true)
    end)

    test.it("blur fires when focus moves away", function()
        app:loadString([[
            lumina.createComponent({id="test", name="Test", render=function()
                local blurred, setBlurred = lumina.useState("blurred", false)
                return lumina.createElement("vbox", {id="root", style={width=80, height=24}},
                    lumina.createElement("vbox", {id="a",
                        style={width=20, height=2},
                        focusable=true,
                        onBlur=function() setBlurred(true) end},
                        lumina.createElement("text", {}, "Item A")
                    ),
                    lumina.createElement("vbox", {id="b",
                        style={width=20, height=2},
                        focusable=true},
                        lumina.createElement("text", {}, "Item B")
                    ),
                    lumina.createElement("text", {id="result"},
                        blurred and "A_BLURRED" or "A_FOCUSED")
                )
            end})
        ]])
        -- Focus A
        app:click(5, 0)
        test.assert.eq(app:screenContains("A_FOCUSED"), true)
        -- Click B to move focus away from A
        app:click(5, 2)
        test.assert.eq(app:screenContains("A_BLURRED"), true)
    end)
end)
