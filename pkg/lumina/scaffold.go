package lumina

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ProjectConfig is the lumina.json project configuration.
type ProjectConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Entry   string `json:"entry"`
	Theme   string `json:"theme"`
}

// ScaffoldProject creates a new Lumina project directory with starter files.
func ScaffoldProject(name string) error {
	// Create project directory
	if err := os.MkdirAll(name, 0o755); err != nil {
		return fmt.Errorf("create project dir: %w", err)
	}

	// Create components subdirectory
	compDir := filepath.Join(name, "components")
	if err := os.MkdirAll(compDir, 0o755); err != nil {
		return fmt.Errorf("create components dir: %w", err)
	}

	// Write main.lua
	mainLua := filepath.Join(name, "main.lua")
	if err := os.WriteFile(mainLua, []byte(scaffoldMainLua(name)), 0o644); err != nil {
		return fmt.Errorf("write main.lua: %w", err)
	}

	// Write components/hello.lua
	helloLua := filepath.Join(compDir, "hello.lua")
	if err := os.WriteFile(helloLua, []byte(scaffoldHelloLua), 0o644); err != nil {
		return fmt.Errorf("write hello.lua: %w", err)
	}

	// Write lumina.json
	cfg := ProjectConfig{
		Name:    name,
		Version: "0.1.0",
		Entry:   "main.lua",
		Theme:   "catppuccin-mocha",
	}
	cfgJSON, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	cfgPath := filepath.Join(name, "lumina.json")
	if err := os.WriteFile(cfgPath, cfgJSON, 0o644); err != nil {
		return fmt.Errorf("write lumina.json: %w", err)
	}

	// Write README.md
	readmePath := filepath.Join(name, "README.md")
	if err := os.WriteFile(readmePath, []byte(scaffoldReadme(name)), 0o644); err != nil {
		return fmt.Errorf("write README.md: %w", err)
	}

	return nil
}

func scaffoldMainLua(name string) string {
	return `local lumina = require("lumina")
local shadcn = require("shadcn")

lumina.setTheme("catppuccin-mocha")

local App = lumina.defineComponent({
    name = "App",
    render = function(self)
        local count, setCount = lumina.useState("count", 0)
        return {
            type = "vbox",
            style = { padding = 1, border = "rounded", background = "#1E1E2E" },
            children = {
                { type = "text", content = "🌟 Welcome to ` + name + `!" },
                { type = "text", content = "" },
                { type = "text", content = "Count: " .. tostring(count) },
                lumina.createElement(shadcn.Button, {
                    label = "Click me",
                    onClick = function() setCount(count + 1) end,
                }),
            }
        }
    end,
})

lumina.mount(App)
lumina.run()
`
}

const scaffoldHelloLua = `-- Example custom component
local lumina = require("lumina")

local Hello = lumina.defineComponent({
    name = "Hello",
    init = function(props)
        return { name = props.name or "World" }
    end,
    render = function(self)
        return {
            type = "hbox",
            style = { padding = 1 },
            children = {
                { type = "text", content = "👋 Hello, " .. self.name .. "!" },
            }
        }
    end,
})

return Hello
`

func scaffoldReadme(name string) string {
	return `# ` + name + `

A terminal UI application built with [Lumina](https://github.com/akzj/lumina).

## Getting Started

` + "```bash" + `
# Run in terminal
lumina main.lua

# Run in dev mode (hot reload + DevTools)
lumina dev main.lua

# Run as web app
lumina serve 8080 main.lua
` + "```" + `

## Project Structure

` + "```" + `
` + name + `/
├── main.lua          # Application entry point
├── components/       # Custom components
│   └── hello.lua     # Example component
├── lumina.json       # Project configuration
└── README.md         # This file
` + "```" + `

## Learn More

- [Lumina Documentation](https://github.com/akzj/lumina)
- [shadcn Components](https://github.com/akzj/lumina/tree/main/pkg/lumina/components/shadcn)
`
}
