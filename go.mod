module github.com/akzj/lumina

go 1.26.1

require github.com/akzj/go-lua v0.9.6

require (
	github.com/mattn/go-runewidth v0.0.23
	golang.org/x/sys v0.43.0
	nhooyr.io/websocket v1.8.17
)

require (
	github.com/alecthomas/chroma/v2 v2.24.1
	github.com/clipperhouse/uax29/v2 v2.2.0 // indirect
	github.com/dlclark/regexp2 v1.12.0 // indirect
)

replace github.com/akzj/go-lua => ../go-lua
