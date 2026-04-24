package lumina

import (
	"encoding/json"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/lumina/proto"
	"github.com/stretchr/testify/assert"
)

func TestSetOutputMode(t *testing.T) {
	SetOutputMode(ModeANSI)
	assert.Equal(t, ModeANSI, GetOutputMode())

	SetOutputMode(ModeJSON)
	assert.Equal(t, ModeJSON, GetOutputMode())

	SetOutputMode(ModeANSI) // reset
}

func TestOutputModeString(t *testing.T) {
	assert.Equal(t, "ansi", ModeANSI.String())
	assert.Equal(t, "json", ModeJSON.String())
}

func TestOutputModeFromString(t *testing.T) {
	assert.Equal(t, ModeANSI, OutputModeFromString("ansi"))
	assert.Equal(t, ModeJSON, OutputModeFromString("json"))
	assert.Equal(t, ModeANSI, OutputModeFromString("unknown")) // default
}

func TestSetOutputModeLua(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		lumina.setOutputMode("json")
		local mode = lumina.getOutputMode()
		assert(mode == "json", "mode should be json")
	`)
	assert.NoError(t, err)
}

func TestGetMCPFrame(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		local frame = lumina.getMCPFrame()
		assert(frame ~= nil, "frame should not be nil")
		assert(#frame > 0, "frame should have content")
	`)
	assert.NoError(t, err)
}

func TestCreateComponentRequest(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		local req = lumina.createComponentRequest("Button", { label = "Click me" })
		assert(req ~= nil, "request should not be nil")
		-- Verify JSON structure
		assert(string.find(req, '"type":"Button"') ~= nil)
		assert(string.find(req, '"label":"Click me"') ~= nil)
	`)
	assert.NoError(t, err)
}

func TestCreateEventNotification(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	err := L.DoString(`
		local lumina = require("lumina")
		local notif = lumina.createEventNotification("btn1", "click", { x = "10", y = "20" })
		assert(notif ~= nil, "notification should not be nil")
		assert(string.find(notif, '"component_id":"btn1"') ~= nil)
		assert(string.find(notif, '"event_type":"click"') ~= nil)
	`)
	assert.NoError(t, err)
}

func TestProtoFrame(t *testing.T) {
	frame := proto.NewFrame()
	assert.NotNil(t, frame)
	assert.True(t, frame.Timestamp > 0)
}

func TestProtoFrameAddPatch(t *testing.T) {
	frame := proto.NewFrame()
	frame.AddPatch(0, 0, proto.Cell{
		Char:       "H",
		Foreground: "cyan",
	})

	assert.Len(t, frame.Patches, 1)
	assert.Equal(t, 0, frame.Patches[0].X)
	assert.Equal(t, 0, frame.Patches[0].Y)
	assert.Equal(t, "H", frame.Patches[0].Cell.Char)
}

func TestProtoEventNotification(t *testing.T) {
	notif := proto.NewEventNotification("comp1", "click", map[string]string{
		"x": "10",
		"y": "20",
	})

	assert.Equal(t, "comp1", notif.ComponentID)
	assert.Equal(t, "click", notif.EventType)
	assert.Equal(t, "10", notif.EventData["x"])
	assert.True(t, notif.Timestamp > 0)
}

func TestProtoMCPResponse(t *testing.T) {
	resp := proto.NewMCPResponse("req123", true, "")
	assert.Equal(t, "req123", resp.RequestID)
	assert.True(t, resp.Success)

	resp2 := proto.NewMCPResponse("req456", false, "error")
	assert.False(t, resp2.Success)
	assert.Equal(t, "error", resp2.Error)
}

func TestToMCPFrame(t *testing.T) {
	frame := NewFrame(10, 5)
	frame.Cells[0][0] = Cell{Char: 'H', Foreground: "cyan", Bold: true}
	frame.Cells[0][1] = Cell{Char: 'i', Foreground: "white"}

	pf := ToMCPFrame(frame)
	assert.NotNil(t, pf)
	assert.Len(t, pf.Cells, 5)
	assert.Len(t, pf.Cells[0], 10)
	assert.Equal(t, "H", pf.Cells[0][0].Char)
	assert.Equal(t, "cyan", pf.Cells[0][0].Foreground)
	assert.True(t, pf.Cells[0][0].Bold)
}

func TestJSONOutput(t *testing.T) {
	frame := proto.NewFrame()
	frame.Cells = append(frame.Cells, []proto.Cell{
		{Char: "H", Foreground: "cyan"},
		{Char: "i", Foreground: "white"},
	})

	data, err := json.Marshal(frame)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"char":"H"`)
	assert.Contains(t, string(data), `"fg":"cyan"`)
}
