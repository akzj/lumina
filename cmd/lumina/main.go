// Package main is the command-line entry point for the Lumina framework.
package main

import (
	"fmt"
	"os"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina"
)

func main() {
	// Create Lua state with standard libraries
	L := lua.NewState()
	defer L.Close()

	// Open Lumina module (makes lumina global and register in package.preload)
	lumina.Open(L)

	// Check for script argument
	if len(os.Args) < 2 {
		// No script provided, run REPL or show version
		fmt.Println("Lumina v" + lumina.ModuleName + " - Terminal React for AI Agents")
		fmt.Println("Usage: lumina <script.lua>")
		os.Exit(0)
	}

	// Execute the user script
	scriptPath := os.Args[1]

	// Verify the file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", scriptPath)
		os.Exit(1)
	}

	// Load and execute the script
	if err := L.DoFile(scriptPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
