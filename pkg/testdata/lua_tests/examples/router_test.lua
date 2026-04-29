-- router_test.lua — Tests for Router demo (multi-page navigation)

test.describe("Router demo", function()
    local app

    test.beforeEach(function()
        app = test.createApp(60, 20)
        app:loadFile("../examples/router_demo.lua")
    end)

    test.afterEach(function()
        app:destroy()
    end)

    -- Initial rendering
    test.it("shows home page initially", function()
        test.assert.eq(app:screenContains("Welcome to Router Demo"), true)
        test.assert.eq(app:screenContains("Press 1 for Home"), true)
    end)

    test.it("shows breadcrumb with Home", function()
        test.assert.eq(app:screenContains("Home"), true)
    end)

    test.it("shows header bar", function()
        test.assert.eq(app:screenContains("Router Demo"), true)
    end)

    -- Navigation via keyboard
    test.it("key 2 navigates to users page", function()
        app:keyPress("2")
        test.assert.eq(app:screenContains("Name"), true)
        test.assert.eq(app:screenContains("Role"), true)
        test.assert.eq(app:screenContains("Alice"), true)
    end)

    test.it("key 3 navigates to settings page", function()
        app:keyPress("3")
        test.assert.eq(app:screenContains("Settings"), true)
        test.assert.eq(app:screenContains("Catppuccin"), true)
    end)

    test.it("key 1 returns to home", function()
        app:keyPress("3")
        app:keyPress("1")
        test.assert.eq(app:screenContains("Welcome to Router Demo"), true)
    end)

    -- Breadcrumb updates
    test.it("breadcrumb shows Users on /users", function()
        app:keyPress("2")
        test.assert.eq(app:screenContains("Users"), true)
    end)

    test.it("breadcrumb shows Settings on /settings", function()
        app:keyPress("3")
        test.assert.eq(app:screenContains("Settings"), true)
    end)

    -- Back navigation
    test.it("b key goes back to previous page", function()
        app:keyPress("2")
        test.assert.eq(app:screenContains("Alice"), true)
        app:keyPress("b")
        test.assert.eq(app:screenContains("Welcome to Router Demo"), true)
    end)

    test.it("multiple back navigations work", function()
        app:keyPress("2")
        app:keyPress("3")
        app:keyPress("b")
        test.assert.eq(app:screenContains("Alice"), true)
        app:keyPress("b")
        test.assert.eq(app:screenContains("Welcome to Router Demo"), true)
    end)

    -- User detail page via store-driven navigation
    test.it("navigating to user detail shows user info", function()
        app:loadString('lumina.router.navigate("/users/1")')
        test.assert.eq(app:screenContains("User: Alice"), true)
        test.assert.eq(app:screenContains("Role: Admin"), true)
        test.assert.eq(app:screenContains("Status: active"), true)
    end)

    test.it("user detail breadcrumb shows User #id", function()
        app:loadString('lumina.router.navigate("/users/3")')
        test.assert.eq(app:screenContains("User #3"), true)
    end)

    test.it("non-existent user shows error alert", function()
        app:loadString('lumina.router.navigate("/users/99")')
        test.assert.eq(app:screenContains("Not Found"), true)
        test.assert.eq(app:screenContains("User #99"), true)
    end)

    -- Users page shows all users
    test.it("users page shows all user names", function()
        app:keyPress("2")
        test.assert.eq(app:screenContains("Alice"), true)
        test.assert.eq(app:screenContains("Bob"), true)
        test.assert.eq(app:screenContains("Charlie"), true)
    end)

    -- Edge cases
    test.it("back on home does nothing (stays on home)", function()
        app:keyPress("b")
        test.assert.eq(app:screenContains("Welcome to Router Demo"), true)
    end)

    test.it("rapid navigation works correctly", function()
        app:keyPress("2")
        app:keyPress("3")
        app:keyPress("1")
        test.assert.eq(app:screenContains("Welcome to Router Demo"), true)
    end)
end)
