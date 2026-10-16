package scaffolding

import (
	"strings"
	"testing"
)

func TestForgetWorkspace(t *testing.T) {
	lock := `{
  "name": "ws",
  "lockfileVersion": 3,
  "packages": {
    "": {"name": "ws", "workspaces": ["apps/*"]},
    "apps/a": {"name": "@x/a", "version": "0.0.0"},
    "apps/b": {"name": "@x/b", "version": "0.0.0"},
    "node_modules/@x/a": {"resolved": "apps/a", "link": true},
    "node_modules/@x/b": {"resolved": "apps/b", "link": true},
    "node_modules/left-pad": {"version": "1.3.0", "resolved": "https://registry.npmjs.org/left-pad/-/left-pad-1.3.0.tgz"}
  }
}`
	out, changed, err := ForgetWorkspace([]byte(lock), "apps/b")
	if err != nil || !changed {
		t.Fatalf("%v %v", changed, err)
	}
	text := string(out)
	if strings.Contains(text, "apps/b") || strings.Contains(text, "@x/b") || !strings.Contains(text, `"apps/a"`) || !strings.Contains(text, "left-pad") {
		t.Fatalf("%s", text)
	}
	if _, changed, _ := ForgetWorkspace(out, "apps/b"); changed {
		t.Fatal("nothing left to forget")
	}
}
