# Lumina Web Frontend

A browser-based terminal renderer for Lumina's WebSocket adapter.

## How it Works

The Lumina engine can output to a WebSocket adapter (`pkg/output/ws.go`) instead of a
terminal. The web frontend connects to that WebSocket and renders the terminal grid as
a DOM-based monospace grid (one `<span>` per cell).

```
┌─────────────────────────────────────────────┐
│  Lumina Engine  ──WSAdapter──►  Browser      │
│  (Go)              :8080       (index.html)  │
└─────────────────────────────────────────────┘
```

## Quick Start

```bash
# 1. Start Lumina with web output
lumina --web :8080 examples/web_demo.lua

# 2. Open in browser
open http://localhost:8080
```

The embedded HTML is served automatically by the WebSocket adapter at the root URL (`/`).

## Standalone Development

For frontend development, you can also open `web/index.html` directly (served by any
static server) and point it at a running Lumina WebSocket server. The page auto-connects
to the same host/port it's served from.

## Protocol

### Server → Client

| Type    | Description                          |
|---------|--------------------------------------|
| `init`  | Initial grid dimensions              |
| `full`  | Complete screen buffer (all cells)   |
| `dirty` | Partial update (only changed rects)  |

### Client → Server

| Type        | Description                              |
|-------------|------------------------------------------|
| `keydown`   | Key press (key name + modifiers)         |
| `keyup`     | Key release                              |
| `mousedown` | Mouse button press (cell x,y + button)   |
| `mouseup`   | Mouse button release                     |
| `mousemove` | Mouse hover (throttled to cell changes)  |
| `scroll`    | Mouse wheel (direction: up/down)         |

### Message Format

```json
// Server → Client
{"type": "full", "data": {"width": 80, "height": 24, "cells": [[{"char": "H", "fg": "#CDD6F4", "bg": "#1E1E2E", "bold": false}, ...]]}}

// Client → Server
{"type": "keydown", "data": {"key": "Enter", "ctrl": false, "alt": false, "shift": false}}
{"type": "mousedown", "data": {"x": 5, "y": 3, "button": "left"}}
```

## Features

- **Auto-reconnect**: Reconnects every 2 seconds on disconnect
- **Dirty rendering**: Only updates changed cells for performance
- **Full keyboard capture**: All keys including modifiers (Ctrl, Alt, Shift)
- **Mouse support**: Click, hover, and scroll events with cell coordinates
- **Status indicator**: Connection status shown in top-right corner
- **No dependencies**: Single HTML file, no npm/build step required

## Browser Support

Modern browsers with WebSocket support (Chrome, Firefox, Safari, Edge).
Requires a monospace font for correct cell alignment.
