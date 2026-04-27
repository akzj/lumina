package buffer

import "testing"

func TestBuffer_NewAndSize(t *testing.T) {
	b := New(10, 5)
	if b.Width() != 10 {
		t.Errorf("Width() = %d, want 10", b.Width())
	}
	if b.Height() != 5 {
		t.Errorf("Height() = %d, want 5", b.Height())
	}
}

func TestBuffer_SetGet(t *testing.T) {
	b := New(10, 10)
	c := Cell{Char: 'A', Foreground: "#ff0000", Bold: true}
	b.Set(3, 2, c)
	got := b.Get(3, 2)
	if got != c {
		t.Errorf("Get(3,2) = %+v, want %+v", got, c)
	}
}

func TestBuffer_OutOfBounds(t *testing.T) {
	b := New(10, 10)
	c := Cell{Char: 'X', Foreground: "#ffffff"}

	// Get out of bounds returns zero.
	zero := Cell{}
	if got := b.Get(-1, 0); got != zero {
		t.Errorf("Get(-1,0) = %+v, want zero", got)
	}
	if got := b.Get(0, -1); got != zero {
		t.Errorf("Get(0,-1) = %+v, want zero", got)
	}
	if got := b.Get(10, 0); got != zero {
		t.Errorf("Get(10,0) = %+v, want zero", got)
	}
	if got := b.Get(0, 10); got != zero {
		t.Errorf("Get(0,10) = %+v, want zero", got)
	}

	// Set out of bounds is a no-op.
	b.Set(100, 0, c)
	b.Set(-1, 0, c)
	b.Set(0, -1, c)
	// Verify buffer is still all zeros.
	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			if got := b.Get(x, y); got != zero {
				t.Errorf("After OOB sets, Get(%d,%d) = %+v, want zero", x, y, got)
			}
		}
	}
}

func TestBuffer_Fill(t *testing.T) {
	b := New(10, 10)
	c := Cell{Char: '#', Background: "#00ff00"}
	r := Rect{X: 2, Y: 3, W: 4, H: 3}
	b.Fill(r, c)

	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			got := b.Get(x, y)
			if r.Contains(x, y) {
				if got != c {
					t.Errorf("Get(%d,%d) inside fill = %+v, want %+v", x, y, got, c)
				}
			} else {
				if !got.Zero() {
					t.Errorf("Get(%d,%d) outside fill = %+v, want zero", x, y, got)
				}
			}
		}
	}
}

func TestBuffer_Clear(t *testing.T) {
	b := New(5, 5)
	b.Set(0, 0, Cell{Char: 'A', Foreground: "#ff0000"})
	b.Set(4, 4, Cell{Char: 'Z', Background: "#0000ff"})
	b.Clear()

	zero := Cell{}
	for y := 0; y < b.Height(); y++ {
		for x := 0; x < b.Width(); x++ {
			if got := b.Get(x, y); got != zero {
				t.Errorf("After Clear, Get(%d,%d) = %+v, want zero", x, y, got)
			}
		}
	}
}

func TestBuffer_Resize_Grow(t *testing.T) {
	b := New(3, 3)
	c := Cell{Char: 'X', Foreground: "#aabbcc"}
	b.Set(1, 1, c)

	b.Resize(5, 5)
	if b.Width() != 5 || b.Height() != 5 {
		t.Fatalf("After Resize(5,5): Width=%d, Height=%d", b.Width(), b.Height())
	}
	// Old content preserved.
	if got := b.Get(1, 1); got != c {
		t.Errorf("After grow, Get(1,1) = %+v, want %+v", got, c)
	}
	// New cells are zero.
	if got := b.Get(4, 4); !got.Zero() {
		t.Errorf("After grow, Get(4,4) = %+v, want zero", got)
	}
}

func TestBuffer_Resize_Shrink(t *testing.T) {
	b := New(5, 5)
	c := Cell{Char: 'Y', Foreground: "#112233"}
	b.Set(1, 1, c)
	b.Set(4, 4, Cell{Char: 'Z'}) // will be clipped

	b.Resize(3, 3)
	if b.Width() != 3 || b.Height() != 3 {
		t.Fatalf("After Resize(3,3): Width=%d, Height=%d", b.Width(), b.Height())
	}
	// Content within new bounds preserved.
	if got := b.Get(1, 1); got != c {
		t.Errorf("After shrink, Get(1,1) = %+v, want %+v", got, c)
	}
	// Out of new bounds returns zero.
	if got := b.Get(4, 4); !got.Zero() {
		t.Errorf("After shrink, Get(4,4) should be zero (OOB)")
	}
}

func TestBuffer_Blit(t *testing.T) {
	dst := New(20, 20)
	src := New(3, 2)
	c1 := Cell{Char: 'A', Foreground: "#ff0000"}
	c2 := Cell{Char: 'B', Foreground: "#00ff00"}
	src.Set(0, 0, c1)
	src.Set(2, 1, c2)

	clip := Rect{X: 0, Y: 0, W: 20, H: 20}
	dirty := Blit(dst, src, 5, 3, clip)

	if got := dst.Get(5, 3); got != c1 {
		t.Errorf("dst.Get(5,3) = %+v, want %+v", got, c1)
	}
	if got := dst.Get(7, 4); got != c2 {
		t.Errorf("dst.Get(7,4) = %+v, want %+v", got, c2)
	}

	// Dirty rect should cover what was written.
	if dirty.W <= 0 || dirty.H <= 0 {
		t.Errorf("dirty rect is empty: %+v", dirty)
	}
}

func TestBuffer_Blit_Clip(t *testing.T) {
	dst := New(20, 20)
	src := New(5, 5)
	c := Cell{Char: 'C', Foreground: "#aaaaaa"}
	src.Fill(Rect{X: 0, Y: 0, W: 5, H: 5}, c)

	// Clip to only allow writing in a 2x2 area.
	clip := Rect{X: 10, Y: 10, W: 2, H: 2}
	Blit(dst, src, 10, 10, clip)

	// Inside clip: should be written.
	if got := dst.Get(10, 10); got != c {
		t.Errorf("dst.Get(10,10) = %+v, want %+v", got, c)
	}
	if got := dst.Get(11, 11); got != c {
		t.Errorf("dst.Get(11,11) = %+v, want %+v", got, c)
	}
	// Outside clip: should be zero.
	if got := dst.Get(12, 10); !got.Zero() {
		t.Errorf("dst.Get(12,10) outside clip = %+v, want zero", got)
	}
}

func TestBuffer_Blit_Transparent(t *testing.T) {
	dst := New(10, 10)
	bg := Cell{Char: '.', Background: "#333333"}
	dst.Fill(Rect{X: 0, Y: 0, W: 10, H: 10}, bg)

	src := New(3, 3)
	fg := Cell{Char: 'X', Foreground: "#ffffff"}
	src.Set(1, 1, fg) // only center cell is non-zero

	clip := Rect{X: 0, Y: 0, W: 10, H: 10}
	Blit(dst, src, 2, 2, clip)

	// Center cell should be overwritten.
	if got := dst.Get(3, 3); got != fg {
		t.Errorf("dst.Get(3,3) = %+v, want %+v", got, fg)
	}
	// Adjacent cells should still be bg (zero src cells are transparent).
	if got := dst.Get(2, 2); got != bg {
		t.Errorf("dst.Get(2,2) = %+v, want bg %+v (transparent skip)", got, bg)
	}
	if got := dst.Get(4, 4); got != bg {
		t.Errorf("dst.Get(4,4) = %+v, want bg %+v (transparent skip)", got, bg)
	}
}

func TestBuffer_Equal(t *testing.T) {
	a := New(5, 5)
	b := New(5, 5)
	if !Equal(a, b) {
		t.Error("Two empty buffers should be equal")
	}

	c := Cell{Char: 'Q', Foreground: "#abcdef"}
	a.Set(2, 3, c)
	if Equal(a, b) {
		t.Error("Buffers with different content should not be equal")
	}

	b.Set(2, 3, c)
	if !Equal(a, b) {
		t.Error("Buffers with same content should be equal")
	}

	// Different sizes.
	d := New(3, 3)
	if Equal(a, d) {
		t.Error("Buffers with different sizes should not be equal")
	}
}

func TestBuffer_FlatBacking(t *testing.T) {
	b := New(10, 5)
	// Verify single allocation: cap should be exactly w*h.
	if cap(b.cells) != 50 {
		t.Errorf("cap(cells) = %d, want 50 (single allocation)", cap(b.cells))
	}
	if len(b.cells) != 50 {
		t.Errorf("len(cells) = %d, want 50", len(b.cells))
	}
}

func TestRect_Contains(t *testing.T) {
	r := Rect{X: 5, Y: 5, W: 10, H: 10}

	tests := []struct {
		x, y int
		want bool
	}{
		{5, 5, true},   // top-left corner (inclusive)
		{14, 14, true},  // bottom-right corner (inclusive, last valid)
		{10, 10, true},  // center
		{4, 5, false},   // left of rect
		{15, 5, false},  // right of rect (exclusive boundary)
		{5, 4, false},   // above rect
		{5, 15, false},  // below rect (exclusive boundary)
	}

	for _, tt := range tests {
		if got := r.Contains(tt.x, tt.y); got != tt.want {
			t.Errorf("Rect%+v.Contains(%d,%d) = %v, want %v", r, tt.x, tt.y, got, tt.want)
		}
	}
}

func TestRect_Intersect(t *testing.T) {
	// Overlapping rects.
	a := Rect{X: 0, Y: 0, W: 10, H: 10}
	b := Rect{X: 5, Y: 5, W: 10, H: 10}
	got := a.Intersect(b)
	want := Rect{X: 5, Y: 5, W: 5, H: 5}
	if got != want {
		t.Errorf("Intersect(%+v, %+v) = %+v, want %+v", a, b, got, want)
	}

	// Non-overlapping rects.
	c := Rect{X: 20, Y: 20, W: 5, H: 5}
	got = a.Intersect(c)
	if got.W != 0 || got.H != 0 {
		t.Errorf("Non-overlapping Intersect = %+v, want zero rect", got)
	}
}

func TestRect_Union(t *testing.T) {
	a := Rect{X: 0, Y: 0, W: 5, H: 5}
	b := Rect{X: 10, Y: 10, W: 5, H: 5}
	got := a.Union(b)
	want := Rect{X: 0, Y: 0, W: 15, H: 15}
	if got != want {
		t.Errorf("Union(%+v, %+v) = %+v, want %+v", a, b, got, want)
	}

	// Union with empty rect.
	empty := Rect{}
	if got := a.Union(empty); got != a {
		t.Errorf("Union with empty = %+v, want %+v", got, a)
	}
	if got := empty.Union(b); got != b {
		t.Errorf("Empty union with b = %+v, want %+v", got, b)
	}
}

func TestRect_Translated(t *testing.T) {
	r := Rect{X: 0, Y: 0, W: 4, H: 10}
	got := r.Translated(6, 0)
	want := Rect{X: 6, Y: 0, W: 4, H: 10}
	if got != want {
		t.Errorf("Translated(6,0) = %+v, want %+v", got, want)
	}
	if r.X != 0 || r.Y != 0 {
		t.Errorf("Translated must not mutate receiver, got r=%+v", r)
	}
}
