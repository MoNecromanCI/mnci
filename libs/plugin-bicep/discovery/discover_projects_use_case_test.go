package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("apps/infra/main.bicep", `metadata version = '1.2.0'

module net '../../libs/network/main.bicep' = {
  name: 'net'
}
module shared 'modules/shared.bicep' = {
  name: 'shared'
}
module registry 'br:contoso.azurecr.io/bicep/storage:1.0.0' = {
  name: 'registry'
}
`)
	write("apps/infra/main.bicepparam", "using './main.bicep'\n")
	write("libs/network/main.bicep", "metadata smithStack = 'azure'\n")
	res, err := Discover(ws, []string{"main.bicep", "apps/infra/main.bicep", "libs/network/main.bicep"})
	if err != nil || len(res.Projects) != 2 {
		t.Fatal(err, res.Projects)
	}
	infra, network := res.Projects[0], res.Projects[1]
	if infra.Kind != protocol.ProjectApp || infra.Version != "1.2.0" || strings.Join(infra.Imports, ",") != "libs/network" || infra.Exports[0] != "apps/infra" {
		t.Errorf("infra %+v", infra)
	}
	if build := infra.Tasks["build"]; build.Outputs[0] != "../../dist/apps/infra/template" || build.DependsOn[0] != "params" || build.Command[6] != "../../dist/apps/infra/template" {
		t.Errorf("infra build %+v", build)
	}
	if params := infra.Tasks["params"]; params.Command[len(params.Command)-1] != "../../dist/apps/infra/parameters/main.parameters.json" {
		t.Errorf("infra params %+v", params)
	}
	if network.Kind != protocol.ProjectLib || network.Stack != "azure" || network.Tasks["build"].DependsOn != nil {
		t.Errorf("network %+v", network)
	}
	if _, ok := network.Tasks["params"]; ok {
		t.Error("a project without main.bicepparam has no params task")
	}
}

func TestModuleRoot(t *testing.T) {
	for file, want := range map[string]string{
		"../../libs/network/main.bicep": "libs/network",
		"modules/x.bicep":               "apps/infra/modules",
		"../../../outside/main.bicep":   "",
	} {
		if got := ModuleRoot("apps/infra", file); got != want {
			t.Errorf("ModuleRoot(%q) = %q, want %q", file, got, want)
		}
	}
}
