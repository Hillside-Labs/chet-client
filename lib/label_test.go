package lib

import "testing"

var testCmds = map[string]string{
	"/usr/local/go/bin/go test .":     "go_test",
	"/usr/local/go/bin/go test":       "go_test",
	"go test .":                       "go_test",
	"/usr/local/go/bin/go build":      "go_build",
	"go build":                        "go_build",
	"pytest tests/":                   "pytest",
	"python -m pytest path/to/tests/": "python",
}

func TestMatchCmds(t *testing.T) {
	cm := CmdMatcher{matchers: DefaultMatches}

	for cmd, expected := range testCmds {
		name := cm.FindMatch(cmd)
		if name != expected {
			t.Fatalf("[%s] Expected: %s, got '%s'", cmd, expected, name)
		}
	}
}
