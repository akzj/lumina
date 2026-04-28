package v2

import "testing"

func TestLuaTestFramework(t *testing.T) {
	runner := NewTestRunner()
	results, err := runner.RunDir("testdata/lua_tests/")
	if err != nil {
		t.Fatal(err)
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
