// Package main is the command-line entry point for the Lumina framework.
package main

import (
	"fmt"
	"os"

	"github.com/akzj/lumina/pkg/lumina"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Lumina — Terminal React for AI Agents")
		fmt.Println("Usage: lumina <script.lua>")
		os.Exit(1)
	}

	app := lumina.NewApp()
	defer app.Close()

	if err := app.RunInteractive(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
