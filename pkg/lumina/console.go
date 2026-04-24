package lumina

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// ConsoleEntry represents a log entry.
type ConsoleEntry struct {
	Level   string `json:"level"` // "log" | "warn" | "error" | "debug"
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Time    int64  `json:"time"`
	Stack   string `json:"stack,omitempty"`
}

// Console holds log entries for AI inspection.
type Console struct {
	entries []ConsoleEntry
	mu      sync.RWMutex
	maxSize int
}

var globalConsole = &Console{maxSize: 1000}

// Log adds a log entry.
func (c *Console) Log(level, msg string, data any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := ConsoleEntry{
		Level:   level,
		Message: msg,
		Data:    data,
		Time:    time.Now().UnixMilli(),
	}

	c.entries = append(c.entries, entry)

	// Trim if over max size
	if len(c.entries) > c.maxSize {
		c.entries = c.entries[len(c.entries)-c.maxSize:]
	}
}

// GetEntries returns all log entries.
func (c *Console) GetEntries() []ConsoleEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]ConsoleEntry, len(c.entries))
	copy(result, c.entries)
	return result
}

// GetEntriesSince returns entries after a timestamp.
func (c *Console) GetEntriesSince(timestamp int64) []ConsoleEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]ConsoleEntry, 0)
	for _, entry := range c.entries {
		if entry.Time > timestamp {
			result = append(result, entry)
		}
	}
	return result
}

// GetErrors returns only error entries.
func (c *Console) GetErrors() []ConsoleEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]ConsoleEntry, 0)
	for _, entry := range c.entries {
		if entry.Level == "error" {
			result = append(result, entry)
		}
	}
	return result
}

// Clear removes all entries.
func (c *Console) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = nil
}

// Size returns the number of entries.
func (c *Console) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Global console functions
func log(msg string, data ...any)      { globalConsole.Log("log", msg, data) }
func logWarn(msg string, data ...any)  { globalConsole.Log("warn", msg, data) }
func logError(msg string, data ...any) { globalConsole.Log("error", msg, data) }
func logDebug(msg string, data ...any) { globalConsole.Log("debug", msg, data) }

// Console API for Lua

// consoleLog(level, message, data) — internal log function
func consoleLog(L *lua.State) int {
	level := L.CheckString(1)
	message := L.CheckString(2)

	var data any
	if L.GetTop() >= 3 {
		data = L.ToAny(3)
	}

	globalConsole.Log(level, message, data)
	return 0
}

// consoleGet() → JSON string of log entries
func consoleGet(L *lua.State) int {
	entries := globalConsole.GetEntries()
	jsonBytes, _ := json.Marshal(entries)
	L.PushString(string(jsonBytes))
	return 1
}

// consoleGetErrors() → JSON string of error entries only
func consoleGetErrors(L *lua.State) int {
	entries := globalConsole.GetErrors()
	jsonBytes, _ := json.Marshal(entries)
	L.PushString(string(jsonBytes))
	return 1
}

// consoleClear() — clear log entries
func consoleClear(L *lua.State) int {
	globalConsole.Clear()
	return 0
}

// consoleSize() → number of entries
func consoleSize(L *lua.State) int {
	L.PushInteger(int64(globalConsole.Size()))
	return 1
}
