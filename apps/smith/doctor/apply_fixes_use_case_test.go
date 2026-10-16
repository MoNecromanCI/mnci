package doctor

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestFixes(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "go.work"), []byte("go 1.26\n"), 0o644)
	work := &protocol.Fix{Description: "fix go.work", Safe: true, Files: []protocol.FileEdit{{Path: "go.work", Content: "go 1.26\n\nuse ./api\n"}}}
	risky := &protocol.Fix{Description: "rewrite ci", Files: []protocol.FileEdit{{Path: "ci.yml", Content: "x\n"}}}
	r := Report{Findings: []Finding{
		{Finding: protocol.Finding{Check: "a", Fix: work}},
		{Finding: protocol.Finding{Check: "b", Fix: work}},
		{Finding: protocol.Finding{Check: "c", Fix: risky}},
		{Finding: protocol.Finding{Check: "d", OK: true, Fix: risky}},
	}}
	safe, manual := Fixes(r)
	if len(safe) != 1 || len(manual) != 1 {
		t.Fatalf("one fix per distinct change: %d safe, %d manual", len(safe), len(manual))
	}
	var shown bytes.Buffer
	if err := ShowFix(&shown, root, safe[0]); err != nil || !strings.Contains(shown.String(), "+use ./api") {
		t.Fatalf("%v\n%s", err, shown.String())
	}
	if err := ApplyFix(root, safe[0]); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(root, "go.work")); string(got) != "go 1.26\n\nuse ./api\n" {
		t.Fatalf("go.work %q", got)
	}
	escape := protocol.Fix{Files: []protocol.FileEdit{{Path: "../outside.txt", Content: "x"}}}
	if err := ApplyFix(root, escape); err == nil {
		t.Fatal("a fix must stay in the workspace")
	}
}
