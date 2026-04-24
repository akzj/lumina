// Package main is the command-line entry point for the Lumina framework.
package main

import (
	"fmt"
	"os"

	"github.com/akzj/lumina/pkg/lumina"
)

const usage = `Lumina — Terminal React for AI Agents

Usage:
  lumina <script.lua>              Run a Lua script in local terminal
  lumina run <script.lua>          Same as above (explicit run mode)
  lumina serve <port> <script.lua> Start web server on port

Examples:
  lumina examples/dashboard/main.lua
  lumina run examples/todo/main.lua
  lumina serve 8080 examples/dashboard/main.lua
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: lumina run <script.lua>")
			os.Exit(1)
		}
		runScript(os.Args[2])

	case "serve":
		if len(os.Args) < 4 {
			fmt.Println("Usage: lumina serve <port> <script.lua>")
			os.Exit(1)
		}
		serveScript(os.Args[2], os.Args[3])

	case "help", "--help", "-h":
		fmt.Print(usage)

	default:
		// Default: treat first arg as script path
		runScript(os.Args[1])
	}
}

func runScript(scriptPath string) {
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", scriptPath)
		os.Exit(1)
	}

	app := lumina.NewApp()
	defer app.Close()

	if err := app.RunInteractive(scriptPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func serveScript(port, scriptPath string) {
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", scriptPath)
		os.Exit(1)
	}

	addr := ":" + port
	server := lumina.NewWebServer(addr, scriptPath)
	fmt.Printf("Lumina web server starting on http://localhost%s\n", addr)
	fmt.Printf("Serving: %s\n", scriptPath)

	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
