package taskcache

import (
	"strings"
	"testing"
)

func TestDiffNamesWhatChanged(t *testing.T) {
	before := KeyInputs{
		Task: "web:build", Command: []string{"npm", "run", "build"}, Env: map[string]string{"MODE": "a", "OLD": "1"},
		Files:    []FileDigest{{"apps/web/a.ts", "1"}, {"apps/web/b.ts", "2"}, {"apps/web/gone.ts", "3"}},
		Upstream: map[string]string{"ui": "u1"}, Deps: map[string]string{"ui:build": "k1"}, Platform: "linux/amd64", Tool: "0.1.0/node",
	}
	now := KeyInputs{
		Task: "web:build", Command: []string{"npm", "run", "build", "--silent"}, Env: map[string]string{"MODE": "b", "NEW": "1"},
		Files:    []FileDigest{{"apps/web/a.ts", "1"}, {"apps/web/b.ts", "9"}, {"apps/web/new.ts", "4"}},
		Upstream: map[string]string{"ui": "u2"}, Deps: map[string]string{"ui:build": "k2"}, Platform: "linux/amd64", Tool: "0.2.0/node",
	}
	var got []string
	for _, d := range Diff(before, now) {
		got = append(got, d.String())
	}
	want := []string{
		"tool       0.1.0/node → 0.2.0/node",
		"command    npm run build → npm run build --silent",
		"env        MODE changed",
		"env        NEW added",
		"env        OLD removed",
		"file       apps/web/b.ts changed",
		"file       apps/web/gone.ts removed",
		"file       apps/web/new.ts added",
		"upstream   ui changed",
		"dependency ui:build changed",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if d := Diff(now, now); len(d) != 0 {
		t.Fatalf("no change, no differences: %v", d)
	}
}
