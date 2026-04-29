package lux

import "embed"

// Sources is the canonical Lux component Lua for require("lux.*") preload.
//go:embed *.lua
var Sources embed.FS
