package pipeline

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestParseGoWork(t *testing.T) {
	src := "go 1.24.3\n\nuse (\n\t./apps/api // the api\n\t\"./libs/core\"\n)\nuse ./packages/sdk\n"
	w := ParseGoWork(src)
	if w.GoVersion != "1.24" || !reflect.DeepEqual(w.Use, []string{"apps/api", "libs/core", "packages/sdk"}) {
		t.Fatalf("%+v", w)
	}
}

func TestDoctorFindsStaleAndMissingEntries(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "api"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "api", "go.mod"), []byte("module x\n"), 0o644)
	_ = os.WriteFile(filepath.Join(ws, "go.work"), []byte("go 1.24\nuse (\n\t./apps/api\n\t./libs/gone\n)\n"), 0o644)
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Root: "apps/api"}, {Root: "libs/new"}}})
	got := map[string]bool{}
	for _, f := range res.Findings {
		got[f.Check] = f.OK
	}
	want := map[string]bool{"go.work entry apps/api exists": true, "go.work entry libs/gone exists": false, "go.work lists apps/api": true, "go.work lists libs/new": false}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestCIReadsGoVersion(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "go.work"), []byte("go 1.25.1\n"), 0o644)
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{{}}})
	if res.Toolchains[0].Version != "1.25" || res.Audit[0].Blocking {
		t.Fatalf("%+v", res)
	}
}
