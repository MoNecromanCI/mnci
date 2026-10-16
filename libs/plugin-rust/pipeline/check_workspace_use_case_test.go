package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDoctor(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "Cargo.toml"), []byte("[workspace]\nmembers = [\"apps/cli\", \"libs/gone\", \"crates/*\"]\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(ws, "apps", "cli"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "cli", "Cargo.toml"), []byte("[package]\nname = \"cli\"\n"), 0o644)
	real := probe
	defer func() { probe = real }()
	probe = func(string, []string) (string, error) { return "", errors.New("not found") }
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Root: "apps/cli"}, {Root: "crates/x"}, {Root: "libs/new"}}})
	got := map[string]bool{}
	for _, f := range res.Findings {
		got[f.Check] = f.OK
	}
	want := map[string]bool{
		"cargo is installed": false, "workspace member apps/cli exists": true, "workspace member libs/gone exists": false,
		"the Cargo workspace lists apps/cli": true, "the Cargo workspace lists crates/x": true, "the Cargo workspace lists libs/new": false,
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: %v, want %v", k, got[k], v)
		}
	}
}

func TestCI(t *testing.T) {
	ws := t.TempDir()
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{{Kind: protocol.ProjectPackage}}})
	if res.Toolchains[0].Version != RustVersion || len(res.Credentials) != 1 || res.Setup[0].Command[1] != "fetch" {
		t.Fatalf("%+v", res)
	}
	_ = os.WriteFile(filepath.Join(ws, "rust-toolchain.toml"), []byte("[toolchain]\nchannel = \"1.85.0\"\n"), 0o644)
	if v := CI(ws, protocol.CIParams{Projects: []protocol.Project{{}}}).Toolchains[0].Version; v != "1.85.0" {
		t.Fatalf("rust-toolchain.toml pins %s", v)
	}
}
