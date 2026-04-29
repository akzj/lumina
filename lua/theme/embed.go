package theme

import "embed"

// Sources is the canonical theme Lua for require("theme") preload.
//go:embed *.lua
var Sources embed.FS
