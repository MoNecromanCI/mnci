package pipeline

import (
	"errors"
	"strings"
	"testing"
)

func TestCheckPythonVersion(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	var asked string
	for _, c := range []struct {
		out string
		err error
		ok  bool
	}{{"3.12\n", nil, true}, {"3.11\n", nil, false}, {"", errors.New("missing"), false}} {
		probe = func(_ string, argv []string) (string, error) { asked = argv[0]; return c.out, c.err }
		if f := CheckPythonVersion("/ws"); f.OK != c.ok || !f.Warning {
			t.Errorf("%q: %+v", c.out, f)
		}
	}
	if !strings.HasPrefix(asked, "/ws/.venv/") {
		t.Fatalf("asked %s, not the venv's python", asked)
	}
}
