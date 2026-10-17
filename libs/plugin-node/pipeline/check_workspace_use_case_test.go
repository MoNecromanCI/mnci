package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
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
	if r := CI(t.TempDir(), protocol.CIParams{}, "24"); len(r.Toolchains) != 0 {
		t.Fatal("no projects, no contribution")
	}
	ws := t.TempDir()
	r := CI(ws, protocol.CIParams{Projects: []protocol.Project{{Kind: protocol.ProjectPackage, Root: "packages/sdk"}}}, "24")
	if r.Toolchains[0].Version != "24" || len(r.Credentials) != 1 || r.Credentials[0].Env != "NODE_AUTH_TOKEN" || !r.Setup[0].Blocking {
		t.Fatalf("contribution %+v", r)
	}
	dir := filepath.Join(ws, "packages", "tools")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name": "tools", "engines": {"vscode": "^1.96.0"}}`), 0o644)
	r = CI(ws, protocol.CIParams{Projects: []protocol.Project{{Kind: protocol.ProjectPackage, Root: "packages/tools"}}}, "24")
	if len(r.Credentials) != 1 || r.Credentials[0].Env != "VSCE_PAT" {
		t.Errorf("an extension publishes to the Marketplace, not npm: %+v", r.Credentials)
	}
}
