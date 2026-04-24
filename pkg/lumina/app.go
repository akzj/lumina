// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"os"

	"github.com/akzj/go-lua/pkg/lua"
)

// App represents an interactive Lumina application.
type App struct {
	L          *lua.State
	RenderLoop *RenderLoop
}

// NewApp creates a new interactive Lumina application.
func NewApp() *App {
	// Create Lua state with standard libraries
	L := lua.NewState()

	// Open Lumina module
	Open(L)

	// Create render loop with default terminal size
	app := &App{
		L:          L,
		RenderLoop: NewRenderLoop(L, 80, 24),
	}

	return app
}

// NewAppWithSize creates a new app with custom terminal size.
func NewAppWithSize(width, height int) *App {
	L := lua.NewState()
	Open(L)

	return &App{
		L:          L,
		RenderLoop: NewRenderLoop(L, width, height),
	}
}

// Run executes a Lua script and starts the render loop.
// This blocks until Stop() is called or an error occurs.
func (app *App) Run(scriptPath string) error {
	// Check for script argument
	if scriptPath == "" {
		fmt.Println("Lumina v" + ModuleName + " - Terminal React for AI Agents")
		fmt.Println("Usage: lumina <script.lua>")
		os.Exit(0)
	}

	// Verify the file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", scriptPath)
	}

	// Execute the user script
	if err := app.L.DoFile(scriptPath); err != nil {
		return fmt.Errorf("script error: %w", err)
	}

	// Initial render of all components
	app.RenderLoop.InitialRender()

	// Start the render loop (this blocks)
	app.RenderLoop.Start()

	return nil
}

// RunInteractive runs the app with stdin event handling.
// This is the main entry point for interactive apps.
func (app *App) RunInteractive(scriptPath string) error {
	// Load script
	if scriptPath != "" {
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", scriptPath)
		}
		if err := app.L.DoFile(scriptPath); err != nil {
			return fmt.Errorf("script error: %w", err)
		}
	}

	// Initial render
	app.RenderLoop.InitialRender()

	// Start render loop
	app.RenderLoop.Start()

	// TODO: Start event handling loop for keyboard/mouse
	// For now, just keep the render loop running
	select {}
}

// Stop stops the application.
func (app *App) Stop() {
	if app.RenderLoop != nil {
		app.RenderLoop.Stop()
	}
}

// Close closes the application and cleans up resources.
func (app *App) Close() {
	app.Stop()
	if app.L != nil {
		app.L.Close()
	}
}
