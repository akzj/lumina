// Package main is the MCP server entry point for Lumina.
// This allows AI agents to control Lumina apps via stdio.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/akzj/lumina/pkg/lumina"
)

// MCPNotification represents a notification sent to the agent.
type MCPNotification struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

func main() {
	// Create Lumina app
	app := lumina.NewApp()

	// Load script if provided (non-blocking load, no render loop)
	scriptPath := ""
	if len(os.Args) > 1 {
		scriptPath = os.Args[1]
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", scriptPath)
			os.Exit(1)
		}
		if err := app.L.DoFile(scriptPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading script: %v\n", err)
			os.Exit(1)
		}
	}

	// Send initial state to agent
	sendNotification(MCPNotification{
		Method: "ready",
		Params: map[string]interface{}{
			"version": lumina.ModuleName,
			"script":  scriptPath,
			"frame":   nil,
		},
	})

	// Process MCP requests (user can start render loop via MCP command)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req lumina.MCPRequest
		if err := json.Unmarshal(line, &req); err != nil {
			sendResponse(lumina.MCPResponse{
				ID:    nil,
				Error: &lumina.MCPError{Code: -32700, Message: "Parse error"},
			})
			continue
		}

		// Handle the request
		resp := app.HandleMCPRequest(req)
		sendResponse(resp)
	}
}

// sendResponse sends a JSON-RPC response to stdout.
func sendResponse(resp lumina.MCPResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	fmt.Println(string(data))
}

// sendNotification sends a JSON-RPC notification to stdout.
func sendNotification(notif MCPNotification) {
	data, err := json.Marshal(notif)
	if err != nil {
		return
	}
	fmt.Println(string(data))
}