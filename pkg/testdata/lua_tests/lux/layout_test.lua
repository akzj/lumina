test.describe("Layout", function()
    test.it("renders header and main", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            local Layout = require("lux.layout")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {},
                        Layout {
                            Layout.Header { height = 1,
                                lumina.createElement("text", {id = "header"}, "My Title"),
                            },
                            Layout.Main {
                                lumina.createElement("text", {id = "content"}, "Hello World"),
                            },
                        }
                    )
                end,
            })
        ]])
        local header = app:find("header")
        test.assert.notNil(header)
        test.assert.eq(header.content, "My Title")

        local content = app:find("content")
        test.assert.notNil(content)
        test.assert.eq(content.content, "Hello World")
        app:destroy()
    end)

    test.it("renders header + sidebar + main + footer", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            local Layout = require("lux.layout")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {},
                        Layout {
                            Layout.Header { height = 1,
                                lumina.createElement("text", {id = "h"}, "Header"),
                            },
                            Layout.Sidebar { width = 20,
                                lumina.createElement("text", {id = "s"}, "Sidebar"),
                            },
                            Layout.Main {
                                lumina.createElement("text", {id = "m"}, "Main"),
                            },
                            Layout.Footer { height = 1,
                                lumina.createElement("text", {id = "f"}, "Footer"),
                            },
                        }
                    )
                end,
            })
        ]])
        test.assert.notNil(app:find("h"))
        test.assert.notNil(app:find("s"))
        test.assert.notNil(app:find("m"))
        test.assert.notNil(app:find("f"))
        test.assert.eq(app:find("h").content, "Header")
        test.assert.eq(app:find("s").content, "Sidebar")
        test.assert.eq(app:find("m").content, "Main")
        test.assert.eq(app:find("f").content, "Footer")
        app:destroy()
    end)

    test.it("sidebar gets fixed width", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            local Layout = require("lux.layout")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {},
                        Layout {
                            Layout.Sidebar { width = 25,
                                lumina.createElement("text", {id = "side"}, "Nav"),
                            },
                            Layout.Main {
                                lumina.createElement("text", {id = "main"}, "Content"),
                            },
                        }
                    )
                end,
            })
        ]])
        local side = app:find("side")
        test.assert.notNil(side)
        local main = app:find("main")
        test.assert.notNil(main)
        test.assert.eq(main.content, "Content")
        app:destroy()
    end)

    test.it("main only layout works", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            local Layout = require("lux.layout")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {},
                        Layout {
                            Layout.Main {
                                lumina.createElement("text", {id = "solo"}, "Solo Content"),
                            },
                        }
                    )
                end,
            })
        ]])
        test.assert.notNil(app:find("solo"))
        test.assert.eq(app:find("solo").content, "Solo Content")
        app:destroy()
    end)

    test.it("via lux umbrella module", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            local lux = require("lux")
            local Layout = lux.Layout

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {},
                        Layout {
                            Layout.Header { height = 1,
                                lumina.createElement("text", {id = "hdr"}, "Via Lux"),
                            },
                            Layout.Main {
                                lumina.createElement("text", {id = "body"}, "Works"),
                            },
                        }
                    )
                end,
            })
        ]])
        test.assert.eq(app:find("hdr").content, "Via Lux")
        test.assert.eq(app:find("body").content, "Works")
        app:destroy()
    end)

    test.it("header and footer appear on screen", function()
        local app = test.createApp(80, 24)
        app:loadString([[
            local Layout = require("lux.layout")

            lumina.createComponent({
                id = "test", name = "Test",
                render = function()
                    return lumina.createElement("vbox", {},
                        Layout {
                            Layout.Header { height = 1,
                                lumina.createElement("text", {}, "HEADER_TEXT"),
                            },
                            Layout.Main {
                                lumina.createElement("text", {}, "MAIN_TEXT"),
                            },
                            Layout.Footer { height = 1,
                                lumina.createElement("text", {}, "FOOTER_TEXT"),
                            },
                        }
                    )
                end,
            })
        ]])
        local text = app:screenText()
        test.assert.contains(text, "HEADER_TEXT")
        test.assert.contains(text, "MAIN_TEXT")
        test.assert.contains(text, "FOOTER_TEXT")
        app:destroy()
    end)
end)
