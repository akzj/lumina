// Package main is the command-line entry point for the Lumina framework.
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	_ "net/http/pprof" // registers handlers on DefaultServeMux

	"github.com/akzj/lumina/pkg/lumina"
)

const usage = `Lumina — Terminal React for AI Agents

Usage:
  lumina <script.lua>              Run a Lua script in local terminal
  lumina run <script.lua>          Same as above (explicit run mode)
  lumina dev <script.lua>          Dev mode: hot reload + DevTools
  lumina dev <port> <script.lua>   Dev mode with web server
  lumina serve <port> <script.lua> Start web server on port
  lumina init <name>               Create a new Lumina project
  lumina version                   Show version info

Examples:
  lumina examples/dashboard/main.lua
  lumina init myapp
  lumina dev examples/todo/main.lua
  lumina serve 8080 examples/dashboard/main.lua

Profiling (net/http/pprof on DefaultServeMux):
  LUMINA_PPROF=:6060 lumina examples/mouse_test.lua
  lumina -pprof=:6060 examples/mouse_test.lua
  go run ./cmd/lumina -pprof=:6060 examples/mouse_test.lua
  curl -sS 'http://127.0.0.1:6060/debug/pprof/heap' -o heap.prof
  curl -sS 'http://127.0.0.1:6060/debug/pprof/profile?seconds=30' -o cpu.prof
  # then: go tool pprof http://127.0.0.1:6060/debug/pprof/profile?seconds=30
`

// extractPprofAndArgs removes -pprof from argv and returns optional listen address
// from the flag or from LUMINA_PPROF when the flag is absent.
func extractPprofAndArgs(orig []string) (pprofAddr string, args []string) {
	args = []string{orig[0]}
	for i := 1; i < len(orig); i++ {
		a := orig[i]
		switch {
		case strings.HasPrefix(a, "-pprof="):
			pprofAddr = strings.TrimPrefix(a, "-pprof=")
		case a == "-pprof" && i+1 < len(orig):
			i++
			pprofAddr = orig[i]
		default:
			args = append(args, a)
		}
	}
	if env := strings.TrimSpace(os.Getenv("LUMINA_PPROF")); env != "" && pprofAddr == "" {
		pprofAddr = env
	}
	return pprofAddr, args
}

// startPprof listens for Go's pprof HTTP endpoints (cpu, heap, goroutine, etc.).
// addr may be ":6060", "6060", or "127.0.0.1:6060".
func startPprof(addr string) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return
	}
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	} else if !strings.Contains(addr, ":") {
		addr = "127.0.0.1:" + addr
	}
	go func() {
		fmt.Fprintf(os.Stderr, "lumina: pprof listening on http://%s/debug/pprof/\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			fmt.Fprintf(os.Stderr, "lumina: pprof server: %v\n", err)
		}
	}()
}

func main() {
	pprofAddr, cmdArgs := extractPprofAndArgs(os.Args)
	if pprofAddr != "" {
		startPprof(pprofAddr)
	}

	if len(cmdArgs) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch cmdArgs[1] {
	case "run":
		if len(cmdArgs) < 3 {
			fmt.Println("Usage: lumina run <script.lua>")
			os.Exit(1)
		}
		runScript(cmdArgs[2])

	case "dev":
		runDev(cmdArgs[2:])

	case "serve":
		if len(cmdArgs) < 4 {
			fmt.Println("Usage: lumina serve <port> <script.lua>")
			os.Exit(1)
		}
		serveScript(cmdArgs[2], cmdArgs[3])

	case "init":
		if len(cmdArgs) < 3 {
			fmt.Println("Usage: lumina init <project-name>")
			os.Exit(1)
		}
		initProject(cmdArgs[2])

	case "version", "--version", "-v":
		printVersion()

	case "help", "--help", "-h":
		fmt.Print(usage)

	default:
		// Default: treat first arg as script path
		runScript(cmdArgs[1])
	}
}

func runScript(scriptPath string) {
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%s", lumina.FormatLuaError(
			fmt.Errorf("%s: file not found", scriptPath), ""))
		os.Exit(1)
	}

	app := lumina.NewApp()
	defer app.Close()

	if err := app.RunInteractive(scriptPath); err != nil {
		fmt.Fprintf(os.Stderr, "%s", lumina.FormatLuaError(err, ""))
		os.Exit(1)
	}
}

func runDev(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: lumina dev <script.lua>")
		fmt.Println("       lumina dev <port> <script.lua>")
		os.Exit(1)
	}

	var scriptPath string
	var port string

	if len(args) >= 2 {
		// lumina dev <port> <script.lua>
		port = args[0]
		scriptPath = args[1]
	} else {
		// lumina dev <script.lua>
		scriptPath = args[0]
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", scriptPath)
		os.Exit(1)
	}

	// Dev mode banner
	fmt.Println("\033[36m🔧 Lumina Dev Mode\033[0m")
	fmt.Println("   Hot reload: \033[32menabled\033[0m")
	fmt.Println("   DevTools:   \033[32mF12\033[0m")
	fmt.Printf("   Script:     %s\n", scriptPath)

	if port != "" {
		// Web dev mode
		fmt.Printf("   Web:        http://localhost:%s\n", port)
		fmt.Println()

		addr := ":" + port
		server := lumina.NewWebServer(addr, scriptPath)
		fmt.Printf("Lumina dev server starting on http://localhost%s\n", addr)

		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Terminal dev mode with hot reload
		fmt.Println()

		app := lumina.NewApp()
		defer app.Close()

		// Enable hot reload before running
		app.L.DoString(`
			local lumina = require("lumina")
			lumina.enableHotReload({ paths = {"` + scriptPath + `"}, interval = 500 })
		`)

		if err := app.RunInteractive(scriptPath); err != nil {
			fmt.Fprintf(os.Stderr, "%s", lumina.FormatLuaError(err, ""))
			os.Exit(1)
		}
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

func initProject(name string) {
	fmt.Printf("Creating new Lumina project: %s\n", name)

	if err := lumina.ScaffoldProject(name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  \033[32m✓\033[0m Created %s/main.lua\n", name)
	fmt.Printf("  \033[32m✓\033[0m Created %s/components/hello.lua\n", name)
	fmt.Printf("  \033[32m✓\033[0m Created %s/lumina.json\n", name)
	fmt.Printf("  \033[32m✓\033[0m Created %s/README.md\n", name)
	fmt.Println()
	fmt.Println("Get started:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  lumina main.lua")
}

func printVersion() {
	fmt.Println("Lumina v0.3.0")
	fmt.Println("Terminal React for AI Agents")
	fmt.Println()
	fmt.Println("  Engine:     Go + Lua")
	fmt.Println("  Components: 57 shadcn")
	fmt.Println("  Hooks:      19 React-style")
	fmt.Println("  Tests:      751+")
}
