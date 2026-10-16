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
	want := map[string]bool{"go.work entry apps/api exists": true, "go.work entry libs/gone exists": false, "go.work lists apps/api": true, "go.work lists libs/new": false, "Go is at least 1.24 (go.work)": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestCIReadsGoVersion(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "go.work"), []byte("go 1.25.1\n"), 0o644)
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api"}, {Name: "core", Root: "libs/core"}}})
	if res.Toolchains[0].Version != "1.25" || res.Audit[0].Blocking || len(res.Audit) != 2 || res.Audit[1].Dir != "libs/core" {
		t.Fatalf("%+v", res)
	}
}

func TestCIInstallsGolangciLintOnlyWhenUsed(t *testing.T) {
	plain := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{{Name: "api", Root: "apps/api", Tasks: map[string]protocol.Task{"lint": {Command: []string{"go", "vet", "./..."}}}}}})
	if len(plain.Setup) != 1 {
		t.Fatalf("go vet needs no install: %+v", plain.Setup)
	}
	lint := protocol.Task{Command: []string{"golangci-lint", "run"}}
	res := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{
		{Name: "a", Root: "apps/a", Tasks: map[string]protocol.Task{"lint": lint}},
		{Name: "b", Root: "libs/b", Tasks: map[string]protocol.Task{"lint": lint}},
	}})
	if len(res.Setup) != 2 || res.Setup[1].Command[2] != "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+GolangciLintVersion {
		t.Fatalf("one pinned install: %+v", res.Setup)
	}
}

func TestCheckToolVersions(t *testing.T) {
	answers := map[string]string{}
	real := probe
	defer func() { probe = real }()
	probe = func(_ string, argv []string) (string, error) {
		if v, ok := answers[argv[0]]; ok {
			return v, nil
		}
		return "", os.ErrNotExist
	}
	byCheck := func(fs []protocol.Finding) map[string]protocol.Finding {
		m := map[string]protocol.Finding{}
		for _, f := range fs {
			m[f.Check] = f
		}
		return m
	}
	answers["go"] = "go1.23.4\n"
	answers["golangci-lint"] = "golangci-lint has version 2.13.0 built with go1.26\n"
	got := byCheck(CheckToolVersions(".", "1.24", true))
	if f := got["Go is at least 1.24 (go.work)"]; f.OK || f.Warning {
		t.Fatalf("an older Go fails: %+v", f)
	}
	if f := got["golangci-lint is 2.14.0, as in CI"]; f.OK || !f.Warning || f.Detail != "found: golangci-lint has version 2.13.0 built with go1.26" {
		t.Fatalf("another golangci-lint warns: %+v", f)
	}
	answers["go"] = "go1.26.0"
	answers["golangci-lint"] = "golangci-lint has version 2.14.0 built with go1.26"
	for check, f := range byCheck(CheckToolVersions(".", "1.24", true)) {
		if !f.OK {
			t.Errorf("%s: %+v", check, f)
		}
	}
	delete(answers, "go")
	if f := byCheck(CheckToolVersions(".", "1.24", false))["Go is installed"]; f.OK {
		t.Fatalf("a missing Go fails: %+v", f)
	}
}
