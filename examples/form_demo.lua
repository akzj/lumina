-- Form demo with Select, Checkbox, and Menu components
local Form = lumina.defineComponent({
    name = "FormDemo",
    render = function()
        return {
            type = "vbox",
            children = {
                { type = "text", content = "=== Form Demo ===" },
                { type = "text", content = "" },
                lumina.Select({
                    id = "country-select",
                    placeholder = "Select country",
                    options = {"USA", "China", "Japan"},
                    value = "USA"
                }),
                lumina.Checkbox({
                    id = "agree-checkbox",
                    label = "I agree to terms"
                }),
                lumina.Menu({
                    id = "file-menu",
                    items = {
                        { label = "New" },
                        { label = "Open" },
                        { label = "Save" },
                        { label = "Exit" }
                    }
                }),
                { type = "text", content = "" },
                lumina.TextField({
                    id = "username",
                    value = "user@example.com",
                    placeholder = "Enter username"
                })
            }
        }
    end
})

lumina.render(Form)
