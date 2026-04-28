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
			if !r.Passed {
				t.Errorf("%s", r.Error)
			} else {
				t.Logf("PASS (%v)", r.Duration)
			}
		})
	}
}
