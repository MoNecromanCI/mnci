package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCI(t *testing.T) {
	ws := t.TempDir()
	core := protocol.Project{Root: "libs/core"}
	if res := CI(ws, protocol.CIParams{Projects: []protocol.Project{core}}); len(res.Toolchains) != 1 || res.Toolchains[0].Name != "cpp" || len(res.Credentials) != 0 {
		t.Fatalf("%+v", res)
	}
	_ = os.MkdirAll(filepath.Join(ws, "libs", "core"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "libs", "core", "vcpkg.json"), []byte(`{"dependencies": ["fmt"]}`), 0o644)
	if res := CI(ws, protocol.CIParams{Projects: []protocol.Project{core}}); len(res.Toolchains) != 2 || res.Toolchains[1].Name != "vcpkg" {
		t.Fatalf("a vcpkg.json brings vcpkg: %+v", res)
	}
}

func TestDoctor(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(_ string, argv []string) (string, error) {
		if argv[0] == "cppcheck" {
			return "", errors.New("missing")
		}
		return "", nil
	}
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "libs", "core"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "libs", "core", "vcpkg.json"), []byte(`{}`), 0o644)
	t.Setenv("VCPKG_ROOT", "")
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Root: "libs/core"}}})
	if len(res.Findings) != 5 || !res.Findings[0].OK || res.Findings[3].OK || res.Findings[4].OK {
		t.Fatalf("%+v", res)
	}
}
