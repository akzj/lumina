package lumina

import (
	"math"
	"testing"

	"github.com/akzj/go-lua/pkg/lua"
)

// -----------------------------------------------------------------------
// Mouse motion event tests
// -----------------------------------------------------------------------

func TestMouseMotion_SGRMotionBit(t *testing.T) {
	// SGR format: button;x;yM where button bit 5 (32) = motion
	// button=35 = 32 (motion) + 3 (release/no button) → mousemove
	ir := &InputReader{events: make(chan AppEvent, 10)}

	// Simulate motion event: button=35 (32+3), x=10, y=5, press=M
	ir.parseSGRMouse([]byte("35;10;5M"))

	select {
	case ev := <-ir.events:
		e := ev.Payload.(*Event)
		if e.Type != "mousemove" {
			t.Fatalf("expected mousemove, got %q", e.Type)
		}
		if e.X != 9 || e.Y != 4 {
			t.Fatalf("expected (9,4), got (%d,%d)", e.X, e.Y)
		}
	default:
		t.Fatal("expected event")
	}
}

func TestMouseMotion_NonMotionStillWorks(t *testing.T) {
	// button=0 (left click, no motion bit) → mousedown
	ir := &InputReader{events: make(chan AppEvent, 10)}
	ir.parseSGRMouse([]byte("0;5;3M"))

	select {
	case ev := <-ir.events:
		e := ev.Payload.(*Event)
		if e.Type != "mousedown" {
			t.Fatalf("expected mousedown, got %q", e.Type)
		}
	default:
		t.Fatal("expected event")
	}
}

func TestMouseMotion_ReleaseStillWorks(t *testing.T) {
	// button=0, lowercase m = release → mouseup
	ir := &InputReader{events: make(chan AppEvent, 10)}
	ir.parseSGRMouse([]byte("0;5;3m"))

	select {
	case ev := <-ir.events:
		e := ev.Payload.(*Event)
		if e.Type != "mouseup" {
			t.Fatalf("expected mouseup, got %q", e.Type)
		}
	default:
		t.Fatal("expected event")
	}
}

// -----------------------------------------------------------------------
// Smooth scrolling tests
// -----------------------------------------------------------------------

func TestSmoothScroll_Velocity(t *testing.T) {
	vp := &Viewport{
		ContentH: 100,
		ViewH:    20,
	}
	vp.ScrollSmooth(5.0)

	if !vp.Animating {
		t.Fatal("expected animating=true after ScrollSmooth")
	}
	if vp.VelocityY != 5.0 {
		t.Fatalf("expected velocity=5.0, got %v", vp.VelocityY)
	}
	if vp.Damping != DefaultDamping {
		t.Fatalf("expected damping=%v, got %v", DefaultDamping, vp.Damping)
	}
}

func TestSmoothScroll_Deceleration(t *testing.T) {
	vp := &Viewport{
		ContentH: 100,
		ViewH:    20,
		Damping:  0.5, // aggressive damping for easy testing
	}
	vp.ScrollSmooth(10.0)

	// Tick 1: position = 0 + 10 = 10, velocity = 10 * 0.5 = 5
	updated := vp.Tick()
	if !updated {
		t.Fatal("expected Tick to return true")
	}
	if math.Abs(vp.ScrollYF-10) > 0.01 {
		t.Fatalf("expected ScrollYF≈10 after tick 1, got %v", vp.ScrollYF)
	}
	if math.Abs(vp.VelocityY-5) > 0.01 {
		t.Fatalf("expected VelocityY≈5 after tick 1, got %v", vp.VelocityY)
	}

	// Tick 2: position = 10 + 5 = 15, velocity = 5 * 0.5 = 2.5
	vp.Tick()
	if math.Abs(vp.ScrollYF-15) > 0.01 {
		t.Fatalf("expected ScrollYF≈15 after tick 2, got %v", vp.ScrollYF)
	}
}

func TestSmoothScroll_ClampBounds(t *testing.T) {
	vp := &Viewport{
		ContentH: 30,
		ViewH:    20,
		Damping:  0.5,
	}
	// Max scroll = 30 - 20 = 10
	vp.ScrollSmooth(50.0) // way too much

	vp.Tick()
	if vp.ScrollYF > 10 {
		t.Fatalf("expected ScrollYF clamped to 10, got %v", vp.ScrollYF)
	}
	if vp.VelocityY != 0 {
		t.Fatalf("expected velocity=0 after clamp, got %v", vp.VelocityY)
	}

	// Test negative clamp
	vp2 := &Viewport{
		ContentH: 100,
		ViewH:    20,
		Damping:  0.5,
	}
	vp2.ScrollSmooth(-50.0)
	vp2.Tick()
	if vp2.ScrollYF < 0 {
		t.Fatalf("expected ScrollYF clamped to 0, got %v", vp2.ScrollYF)
	}
}

func TestSmoothScroll_StopsWhenNegligible(t *testing.T) {
	vp := &Viewport{
		ContentH: 100,
		ViewH:    20,
		Damping:  0.1, // very aggressive damping
	}
	vp.ScrollSmooth(1.0)

	// After a few ticks, velocity should be < 0.1 and animating should stop
	for i := 0; i < 20; i++ {
		if !vp.Animating {
			break
		}
		vp.Tick()
	}
	if vp.Animating {
		t.Fatal("expected animating=false after velocity becomes negligible")
	}
	if vp.VelocityY != 0 {
		t.Fatalf("expected velocity=0 when stopped, got %v", vp.VelocityY)
	}
}

func TestSmoothScroll_SyncFromInt(t *testing.T) {
	vp := &Viewport{
		ScrollY:  15,
		ScrollYF: 0,
		Damping:  0.85,
	}
	vp.ScrollSmooth(5.0)
	vp.SyncFloatFromInt()

	if vp.ScrollYF != 15 {
		t.Fatalf("expected ScrollYF=15 after sync, got %v", vp.ScrollYF)
	}
	if vp.VelocityY != 0 {
		t.Fatal("expected velocity=0 after sync")
	}
	if vp.Animating {
		t.Fatal("expected animating=false after sync")
	}
}

// -----------------------------------------------------------------------
// SubPixelCanvas tests
// -----------------------------------------------------------------------

func TestSubPixelCanvas_Creation(t *testing.T) {
	c := NewSubPixelCanvas(10, 5)
	if c.CellW != 10 || c.CellH != 5 {
		t.Fatalf("expected cell 10x5, got %dx%d", c.CellW, c.CellH)
	}
	if c.PixW != 10 || c.PixH != 10 {
		t.Fatalf("expected pixel 10x10, got %dx%d", c.PixW, c.PixH)
	}
	if len(c.Pixels) != 10 {
		t.Fatalf("expected 10 pixel rows, got %d", len(c.Pixels))
	}
	if len(c.Pixels[0]) != 10 {
		t.Fatalf("expected 10 pixel cols, got %d", len(c.Pixels[0]))
	}
}

func TestSubPixelCanvas_SetPixelAndRender(t *testing.T) {
	c := NewSubPixelCanvas(2, 1) // 2 cells wide, 1 cell tall → 2x2 subpixels

	red := Color{255, 0, 0}
	blue := Color{0, 0, 255}

	// Set different colors for top and bottom of first cell
	c.SetPixel(0, 0, red)  // top-left
	c.SetPixel(0, 1, blue) // bottom-left

	frame := NewFrame(2, 1)
	c.RenderToFrame(frame, 0, 0)

	// First cell should use half-block with red on top, blue on bottom
	cell := frame.Cells[0][0]
	if cell.Char != '▀' {
		t.Fatalf("expected ▀ for different top/bottom, got %c", cell.Char)
	}
	if cell.Foreground != red.Hex() {
		t.Fatalf("expected fg=%s, got %s", red.Hex(), cell.Foreground)
	}
	if cell.Background != blue.Hex() {
		t.Fatalf("expected bg=%s, got %s", blue.Hex(), cell.Background)
	}
}

func TestSubPixelCanvas_SameColorFullBlock(t *testing.T) {
	c := NewSubPixelCanvas(1, 1)

	green := Color{0, 255, 0}
	c.SetPixel(0, 0, green) // top
	c.SetPixel(0, 1, green) // bottom — same color

	frame := NewFrame(1, 1)
	c.RenderToFrame(frame, 0, 0)

	cell := frame.Cells[0][0]
	if cell.Char != '█' {
		t.Fatalf("expected █ for same color, got %c (%d)", cell.Char, cell.Char)
	}
	if cell.Foreground != green.Hex() {
		t.Fatalf("expected fg=%s, got %s", green.Hex(), cell.Foreground)
	}
}

func TestSubPixelCanvas_DrawLine(t *testing.T) {
	c := NewSubPixelCanvas(5, 3) // 5x6 subpixels

	white := Color{255, 255, 255}
	c.DrawLine(0, 0, 4, 4, white)

	// Diagonal line should set pixels along the diagonal
	if c.GetPixel(0, 0) != white {
		t.Fatal("expected (0,0) to be set")
	}
	if c.GetPixel(4, 4) != white {
		t.Fatal("expected (4,4) to be set")
	}
	// Check that at least some middle pixels are set
	midSet := false
	for i := 1; i < 4; i++ {
		if c.GetPixel(i, i) == white {
			midSet = true
			break
		}
	}
	if !midSet {
		t.Fatal("expected some middle diagonal pixels to be set")
	}
}

func TestSubPixelCanvas_DrawCircle(t *testing.T) {
	c := NewSubPixelCanvas(20, 10) // 20x20 subpixels

	white := Color{255, 255, 255}
	c.DrawCircle(10, 10, 5, white)

	// Check that the top of the circle is set (cx, cy-r)
	if c.GetPixel(10, 5) != white {
		t.Fatal("expected top of circle to be set")
	}
	// Check that the bottom is set (cx, cy+r)
	if c.GetPixel(10, 15) != white {
		t.Fatal("expected bottom of circle to be set")
	}
	// Check that the center is NOT set (it's a circle outline)
	if c.GetPixel(10, 10) == white {
		t.Fatal("expected center to NOT be set (outline only)")
	}
}

func TestSubPixelCanvas_DrawRoundedRect(t *testing.T) {
	c := NewSubPixelCanvas(20, 10) // 20x20 subpixels

	white := Color{255, 255, 255}
	c.DrawRoundedRect(2, 2, 16, 16, 3, white)

	// Top edge (straight part) should be set
	if c.GetPixel(10, 2) != white {
		t.Fatal("expected top edge to be set")
	}
	// Left edge (straight part) should be set
	if c.GetPixel(2, 10) != white {
		t.Fatal("expected left edge to be set")
	}
}

func TestSubPixelCanvas_FillRect(t *testing.T) {
	c := NewSubPixelCanvas(5, 3)

	red := Color{255, 0, 0}
	c.FillRect(1, 1, 3, 3, red)

	// Inside should be set
	if c.GetPixel(2, 2) != red {
		t.Fatal("expected inside of filled rect to be set")
	}
	// Outside should not be set
	if c.GetPixel(0, 0) == red {
		t.Fatal("expected outside of filled rect to NOT be set")
	}
}

func TestSubPixelCanvas_BoundsChecking(t *testing.T) {
	c := NewSubPixelCanvas(5, 3)

	// Out of bounds should not panic
	c.SetPixel(-1, -1, Color{255, 0, 0})
	c.SetPixel(100, 100, Color{255, 0, 0})
	c.SetPixel(5, 6, Color{255, 0, 0})

	// GetPixel out of bounds returns BgColor
	got := c.GetPixel(-1, 0)
	if got != c.BgColor {
		t.Fatal("expected BgColor for out of bounds")
	}
}

func TestColorFromHex(t *testing.T) {
	c := ColorFromHex("#ff8000")
	if c.R != 255 || c.G != 128 || c.B != 0 {
		t.Fatalf("expected (255,128,0), got (%d,%d,%d)", c.R, c.G, c.B)
	}

	// Round-trip
	hex := c.Hex()
	if hex != "#ff8000" {
		t.Fatalf("expected #ff8000, got %s", hex)
	}

	// Invalid input
	c2 := ColorFromHex("invalid")
	if c2.R != 0 || c2.G != 0 || c2.B != 0 {
		t.Fatal("expected zero color for invalid hex")
	}
}

// -----------------------------------------------------------------------
// Lua Canvas API test
// -----------------------------------------------------------------------

func TestLua_CreateCanvasAPI(t *testing.T) {
	ClearComponents()
	ClearContextValues()

	L := lua.NewState()
	Open(L)
	defer L.Close()

	err := L.DoString(`
		local c = lumina.createCanvas(10, 5)
		_w = c.width
		_h = c.height
		_pw = c.pixelWidth
		_ph = c.pixelHeight
		c.setPixel(0, 0, "#ff0000")
		c.drawLine(0, 0, 9, 9, "#00ff00")
		c.drawCircle(5, 5, 3, "#0000ff")
		c.clear()
	`)
	if err != nil {
		t.Fatalf("createCanvas error: %v", err)
	}

	L.GetGlobal("_w")
	w, _ := L.ToNumber(-1)
	L.Pop(1)
	if w != 10 {
		t.Fatalf("expected width=10, got %v", w)
	}

	L.GetGlobal("_h")
	h, _ := L.ToNumber(-1)
	L.Pop(1)
	if h != 5 {
		t.Fatalf("expected height=5, got %v", h)
	}

	L.GetGlobal("_pw")
	pw, _ := L.ToNumber(-1)
	L.Pop(1)
	if pw != 10 {
		t.Fatalf("expected pixelWidth=10, got %v", pw)
	}

	L.GetGlobal("_ph")
	ph, _ := L.ToNumber(-1)
	L.Pop(1)
	if ph != 10 {
		t.Fatalf("expected pixelHeight=10, got %v", ph)
	}
}
