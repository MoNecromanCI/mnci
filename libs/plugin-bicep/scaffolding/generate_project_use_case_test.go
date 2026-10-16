package scaffolding

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MoNecromanCI/mnci/libs/plugin-bicep/discovery"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	info := protocol.WorkspaceInfo{Name: "Acme-Corp"}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "infra", Directory: "apps/infra", Workspace: info}); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "network", Directory: "libs/network", Workspace: info}); err != nil {
		t.Fatal(err)
	}
	params, err := os.ReadFile(filepath.Join(ws, "apps", "infra", "main.bicepparam"))
	if err != nil || string(params) != "using './main.bicep'\n\nparam storageName = 'stacmecorpinfra'\n" {
		t.Fatalf("%q %v", params, err)
	}
	if _, err := os.Stat(filepath.Join(ws, "libs", "network", "main.bicepparam")); err == nil {
		t.Error("a module has no parameters")
	}
	found, err := discovery.Discover(ws, []string{"apps/infra/main.bicep", "libs/network/main.bicep"})
	if err != nil || found.Projects[0].Kind != protocol.ProjectApp || found.Projects[0].Version != "0.0.0" || found.Projects[1].Kind != protocol.ProjectLib {
		t.Fatalf("%v %+v", err, found.Projects)
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "x", Directory: "packages/x", Workspace: info}); err == nil {
		t.Error("a package is refused")
	}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "infra", Directory: "apps/infra", Workspace: info}); err == nil {
		t.Error("an existing directory is refused")
	}
}
