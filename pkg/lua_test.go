package v2

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLuaTestFramework runs all Lua tests under testdata/lua_tests/, or a single
// file when LUMINA_LUA_TEST is set (path relative to this package directory, e.g.
// testdata/lua_tests/examples/scrollview_test.lua). See docs/TESTING.md and scripts/lua-test.sh.
func TestLuaTestFramework(t *testing.T) {
	runner := NewTestRunner()
	var results []TestResult
	var err error
	if p := os.Getenv("LUMINA_LUA_TEST"); p != "" {
		p = filepath.Clean(p)
		results, err = runner.RunFile(p)
	} else {
		results, err = runner.RunDir("testdata/lua_tests/")
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 && os.Getenv("LUMINA_LUA_TEST") != "" {
		t.Fatalf("LUMINA_LUA_TEST=%q: no test cases ran (missing file, or no test.describe in file)", os.Getenv("LUMINA_LUA_TEST"))
	}
	for _, r := range results {
		t.Run(r.Suite+"/"+r.Name, func(t *testing.T) {
			// Output any test logs (visible with -v or on failure)
			for _, log := range r.Logs {
				t.Log(log)
			}
			if !r.Passed {
				t.Errorf("%s", r.Error)
			} else {
				t.Logf("PASS (%v)", r.Duration)
			}
		})
	}
}
