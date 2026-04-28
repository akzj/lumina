package router

import (
	"testing"
)

func TestStaticRoutes(t *testing.T) {
	r := New()
	r.AddRoute("/about")
	r.AddRoute("/users/list")

	tests := []struct {
		path    string
		wantNil bool
		wantPat string
	}{
		{"/about", false, "/about"},
		{"/users/list", false, "/users/list"},
		{"/nope", true, ""},
	}
	for _, tt := range tests {
		result := r.Match(tt.path)
		if tt.wantNil {
			if result != nil {
				t.Errorf("Match(%q) = %v, want nil", tt.path, result)
			}
			continue
		}
		if result == nil {
			t.Fatalf("Match(%q) = nil, want match", tt.path)
		}
		if result.Route.Path() != tt.wantPat {
			t.Errorf("Match(%q).Route.Path() = %q, want %q", tt.path, result.Route.Path(), tt.wantPat)
		}
		if len(result.Params) != 0 {
			t.Errorf("Match(%q).Params = %v, want empty", tt.path, result.Params)
		}
	}
}

func TestParamRoutes(t *testing.T) {
	r := New()
	r.AddRoute("/users/:id")

	result := r.Match("/users/42")
	if result == nil {
		t.Fatal("Match(/users/42) = nil, want match")
	}
	if result.Params["id"] != "42" {
		t.Errorf("params[id] = %q, want %q", result.Params["id"], "42")
	}
}

func TestMultiParam(t *testing.T) {
	r := New()
	r.AddRoute("/users/:id/posts/:postId")

	result := r.Match("/users/7/posts/99")
	if result == nil {
		t.Fatal("Match = nil, want match")
	}
	if result.Params["id"] != "7" {
		t.Errorf("params[id] = %q, want %q", result.Params["id"], "7")
	}
	if result.Params["postId"] != "99" {
		t.Errorf("params[postId] = %q, want %q", result.Params["postId"], "99")
	}
}

func TestWildcard(t *testing.T) {
	r := New()
	r.AddRoute("/files/*path")

	tests := []struct {
		path     string
		wantPath string
	}{
		{"/files/a/b/c", "a/b/c"},
		{"/files/readme.txt", "readme.txt"},
		{"/files/deep/nested/dir/file.go", "deep/nested/dir/file.go"},
	}
	for _, tt := range tests {
		result := r.Match(tt.path)
		if result == nil {
			t.Fatalf("Match(%q) = nil, want match", tt.path)
		}
		if result.Params["path"] != tt.wantPath {
			t.Errorf("Match(%q).Params[path] = %q, want %q", tt.path, result.Params["path"], tt.wantPath)
		}
	}
}

func TestWildcardNoSegments(t *testing.T) {
	r := New()
	r.AddRoute("/files/*path")

	// "/files" with no trailing segments — should NOT match since wildcard needs ≥0 trailing.
	// Actually with our implementation: pattern has 2 segs ["files", "*path"],
	// pathSegments for "/files" has 1 seg ["files"]. len(pathSegments) < len(pattern)-1 is 1 < 1 = false,
	// so we enter the loop: i=0 matches "files", i=1 is "*path" with pathSegments[1:] = "" joined.
	// This should match with empty path param.
	result := r.Match("/files")
	if result == nil {
		t.Fatal("Match(/files) = nil, want match with empty wildcard")
	}
	if result.Params["path"] != "" {
		t.Errorf("params[path] = %q, want %q", result.Params["path"], "")
	}
}

func TestNoMatch(t *testing.T) {
	r := New()
	r.AddRoute("/users/:id")

	result := r.Match("/posts/1")
	if result != nil {
		t.Errorf("Match(/posts/1) = %v, want nil", result)
	}
}

func TestNavigate(t *testing.T) {
	r := New()
	r.AddRoute("/users/:id")

	var calledPath string
	var calledParams map[string]string
	r.OnChange(func(path string, params map[string]string) {
		calledPath = path
		calledParams = params
	})

	r.Navigate("/users/5")

	if r.CurrentPath() != "/users/5" {
		t.Errorf("CurrentPath() = %q, want %q", r.CurrentPath(), "/users/5")
	}
	if calledPath != "/users/5" {
		t.Errorf("onChange path = %q, want %q", calledPath, "/users/5")
	}
	if calledParams["id"] != "5" {
		t.Errorf("onChange params[id] = %q, want %q", calledParams["id"], "5")
	}

	// History should contain initial "/"
	hist := r.History()
	if len(hist) != 1 || hist[0] != "/" {
		t.Errorf("History() = %v, want [/]", hist)
	}
}

func TestBack(t *testing.T) {
	r := New()
	r.AddRoute("/a")
	r.AddRoute("/b")

	r.Navigate("/a")
	r.Navigate("/b")

	if r.CurrentPath() != "/b" {
		t.Fatalf("CurrentPath() = %q, want /b", r.CurrentPath())
	}

	ok := r.Back()
	if !ok {
		t.Fatal("Back() = false, want true")
	}
	if r.CurrentPath() != "/a" {
		t.Errorf("after Back(), CurrentPath() = %q, want /a", r.CurrentPath())
	}

	ok = r.Back()
	if !ok {
		t.Fatal("Back() = false, want true")
	}
	if r.CurrentPath() != "/" {
		t.Errorf("after 2nd Back(), CurrentPath() = %q, want /", r.CurrentPath())
	}

	ok = r.Back()
	if ok {
		t.Error("Back() = true on empty history, want false")
	}
}

func TestBackTriggersOnChange(t *testing.T) {
	r := New()
	r.AddRoute("/page")

	var calls int
	r.OnChange(func(path string, params map[string]string) {
		calls++
	})

	r.Navigate("/page")
	r.Back()

	if calls != 2 {
		t.Errorf("onChange called %d times, want 2", calls)
	}
}

func TestParamsAccessor(t *testing.T) {
	r := New()
	r.AddRoute("/items/:category/:id")

	r.Navigate("/items/books/42")

	params := r.Params()
	if params["category"] != "books" {
		t.Errorf("Params()[category] = %q, want %q", params["category"], "books")
	}
	if params["id"] != "42" {
		t.Errorf("Params()[id] = %q, want %q", params["id"], "42")
	}

	// Verify it's a copy — mutating returned map doesn't affect router.
	params["category"] = "MUTATED"
	if r.Params()["category"] != "books" {
		t.Error("Params() returned reference, not copy")
	}
}

func TestMultipleOnChange(t *testing.T) {
	r := New()

	var calls1, calls2 int
	r.OnChange(func(path string, params map[string]string) { calls1++ })
	r.OnChange(func(path string, params map[string]string) { calls2++ })

	r.Navigate("/x")

	if calls1 != 1 || calls2 != 1 {
		t.Errorf("calls = (%d, %d), want (1, 1)", calls1, calls2)
	}
}

func TestNavigateSamePath(t *testing.T) {
	r := New()

	var calls int
	r.OnChange(func(path string, params map[string]string) { calls++ })

	r.Navigate("/x")
	r.Navigate("/x")

	// Should still trigger onChange both times.
	if calls != 2 {
		t.Errorf("onChange called %d times, want 2", calls)
	}
	// History should have ["/", "/x"].
	hist := r.History()
	if len(hist) != 2 {
		t.Errorf("History length = %d, want 2", len(hist))
	}
}

func TestRootPath(t *testing.T) {
	r := New()
	r.AddRoute("/")

	result := r.Match("/")
	if result == nil {
		t.Fatal("Match(/) = nil, want match")
	}
	if result.Route.Path() != "/" {
		t.Errorf("Route.Path() = %q, want %q", result.Route.Path(), "/")
	}
}

func TestTrailingSlash(t *testing.T) {
	r := New()
	r.AddRoute("/users")

	// "/users/" should match "/users" since splitPath strips empty segments.
	result := r.Match("/users/")
	if result == nil {
		t.Fatal("Match(/users/) = nil, want match (trailing slash ignored)")
	}
}

func TestEmptyPath(t *testing.T) {
	r := New()
	r.AddRoute("/")

	result := r.Match("")
	// Empty path splits to zero segments, same as "/".
	if result == nil {
		t.Fatal("Match('') = nil, want match (same as /)")
	}
}

func TestRouteCount(t *testing.T) {
	r := New()
	if r.RouteCount() != 0 {
		t.Errorf("RouteCount() = %d, want 0", r.RouteCount())
	}
	r.AddRoute("/a")
	r.AddRoute("/b")
	if r.RouteCount() != 2 {
		t.Errorf("RouteCount() = %d, want 2", r.RouteCount())
	}
}

func TestHistoryCopy(t *testing.T) {
	r := New()
	r.Navigate("/a")

	hist := r.History()
	hist[0] = "MUTATED"

	if r.History()[0] != "/" {
		t.Error("History() returned reference, not copy")
	}
}

func TestInitialState(t *testing.T) {
	r := New()
	if r.CurrentPath() != "/" {
		t.Errorf("initial CurrentPath() = %q, want /", r.CurrentPath())
	}
	if len(r.Params()) != 0 {
		t.Errorf("initial Params() = %v, want empty", r.Params())
	}
	if len(r.History()) != 0 {
		t.Errorf("initial History() = %v, want empty", r.History())
	}
}
