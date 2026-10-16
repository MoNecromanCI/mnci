package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestDoctor(t *testing.T) {
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"workspaces":["apps/*"]}`), 0o644)
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Root: "apps/web"}, {Root: "tools/x"}}}, "24")
	got := map[string]bool{}
	for _, f := range res.Findings {
		got[f.Check] = f.OK
	}
	if got["Node dependencies are installed"] || !got["npm workspaces include apps/web"] || got["npm workspaces include tools/x"] {
		t.Fatalf("findings %+v", res.Findings)
	}
}

func TestCIOnlyWhenNodeProjectsExist(t *testing.T) {
	if r := CI(protocol.CIParams{}, "24"); len(r.Toolchains) != 0 {
		t.Fatal("no projects, no contribution")
	}
	r := CI(protocol.CIParams{Projects: []protocol.Project{{Kind: protocol.ProjectPackage}}}, "24")
	if r.Toolchains[0].Version != "24" || len(r.Credentials) != 1 || !r.Setup[0].Blocking {
		t.Fatalf("contribution %+v", r)
	}
}
