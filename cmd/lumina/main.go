// Package main is the command-line entry point for the Lumina framework.
package main

import (
	"fmt"
	"os"

	"github.com/akzj/lumina/pkg/lumina"
)

func main() {
	// Create interactive app
	app := lumina.NewApp()
	defer app.Close()

	// Get script path from args
	scriptPath := ""
	if len(os.Args) >= 2 {
		scriptPath = os.Args[1]
	}

	// Run the app
	if err := app.Run(scriptPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
