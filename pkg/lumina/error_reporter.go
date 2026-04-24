package lumina

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// LuaError represents a parsed Lua error with context.
type LuaError struct {
	File      string // source file
	Line      int    // line number (0 if unknown)
	Component string // component name (empty if unknown)
	Message   string // error message
	Stack     string // full Lua stack trace (if available)
}

// ParseLuaError parses a Lua error string into structured form.
// Lua errors typically look like: "file.lua:42: error message"
// or: "runtime error: file.lua:42: error message"
func ParseLuaError(err error, componentName string) LuaError {
	if err == nil {
		return LuaError{}
	}

	msg := err.Error()
	le := LuaError{
		Component: componentName,
		Message:   msg,
	}

	// Strip common prefixes
	for _, prefix := range []string{"script error: ", "runtime error: ", "render: "} {
		msg = strings.TrimPrefix(msg, prefix)
	}

	// Try to parse "file:line: message" pattern
	re := regexp.MustCompile(`^(.+?):(\d+):\s*(.*)$`)
	if m := re.FindStringSubmatch(msg); m != nil {
		le.File = m[1]
		le.Line, _ = strconv.Atoi(m[2])
		le.Message = m[3]
	}

	return le
}

// FormatLuaError formats a Lua error into a readable terminal-friendly string.
func FormatLuaError(err error, componentName string) string {
	le := ParseLuaError(err, componentName)
	return le.Format()
}

// Format renders the LuaError as a terminal-friendly string with box drawing.
func (le LuaError) Format() string {
	var sb strings.Builder

	sb.WriteString("\n\033[31m") // red
	sb.WriteString("╭─ Lumina Error ─────────────────────────────╮\n")

	if le.Component != "" {
		sb.WriteString(fmt.Sprintf("│ Component: %-33s │\n", le.Component))
	}
	if le.File != "" {
		sb.WriteString(fmt.Sprintf("│ File:      %-33s │\n", le.File))
	}
	if le.Line > 0 {
		sb.WriteString(fmt.Sprintf("│ Line:      %-33d │\n", le.Line))
	}

	sb.WriteString("├───────────────────────────────────────────────┤\n")

	// Wrap message to fit in box
	msg := le.Message
	if msg == "" {
		msg = "(unknown error)"
	}
	for len(msg) > 0 {
		line := msg
		if len(line) > 45 {
			// Find last space before 45 chars
			cutAt := 45
			for cutAt > 20 && line[cutAt] != ' ' {
				cutAt--
			}
			if cutAt <= 20 {
				cutAt = 45
			}
			line = msg[:cutAt]
			msg = msg[cutAt:]
			msg = strings.TrimLeft(msg, " ")
		} else {
			msg = ""
		}
		sb.WriteString(fmt.Sprintf("│ %-45s │\n", line))
	}

	if le.Stack != "" {
		sb.WriteString("├───────────────────────────────────────────────┤\n")
		sb.WriteString(fmt.Sprintf("│ %-45s │\n", "Stack:"))
		for _, line := range strings.Split(le.Stack, "\n") {
			if line == "" {
				continue
			}
			if len(line) > 45 {
				line = line[:42] + "..."
			}
			sb.WriteString(fmt.Sprintf("│ %-45s │\n", line))
		}
	}

	sb.WriteString("╰───────────────────────────────────────────────╯\n")
	sb.WriteString("\033[0m") // reset

	return sb.String()
}

// FormatLuaErrorPlain formats without ANSI colors (for testing/logging).
func FormatLuaErrorPlain(err error, componentName string) string {
	le := ParseLuaError(err, componentName)
	return le.FormatPlain()
}

// FormatPlain renders the LuaError without ANSI color codes.
func (le LuaError) FormatPlain() string {
	var sb strings.Builder

	sb.WriteString("╭─ Lumina Error ─────────────────────────────╮\n")

	if le.Component != "" {
		sb.WriteString(fmt.Sprintf("│ Component: %-33s │\n", le.Component))
	}
	if le.File != "" {
		sb.WriteString(fmt.Sprintf("│ File:      %-33s │\n", le.File))
	}
	if le.Line > 0 {
		sb.WriteString(fmt.Sprintf("│ Line:      %-33d │\n", le.Line))
	}

	sb.WriteString("├───────────────────────────────────────────────┤\n")

	msg := le.Message
	if msg == "" {
		msg = "(unknown error)"
	}
	sb.WriteString(fmt.Sprintf("│ %-45s │\n", msg))

	if le.Stack != "" {
		sb.WriteString("├───────────────────────────────────────────────┤\n")
		sb.WriteString(fmt.Sprintf("│ %-45s │\n", "Stack:"))
		for _, line := range strings.Split(le.Stack, "\n") {
			if line == "" {
				continue
			}
			if len(line) > 45 {
				line = line[:42] + "..."
			}
			sb.WriteString(fmt.Sprintf("│ %-45s │\n", line))
		}
	}

	sb.WriteString("╰───────────────────────────────────────────────╯\n")

	return sb.String()
}
