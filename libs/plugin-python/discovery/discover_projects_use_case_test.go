package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "pyproject.toml", "[tool.ruff]\nline-length = 100\n")
	write(t, ws, "apps/api/pyproject.toml", "[project]\nname = \"acme-api\"\nversion = \"0.0.0\"\ndependencies = [\"Acme-Core>=0.1\", \"requests\"]\n\n[tool.mnci]\nkind = \"app\"\n")
	write(t, ws, "apps/api/src/acme_api/__main__.py", "import os, sys\nfrom acme_utils.x import y\nfrom . import local\n")
	write(t, ws, "apps/api/build/lib/stale.py", "import stale\n")
	write(t, ws, "packages/sdk/pyproject.toml", "[project]\nname = \"acme.sdk\"\nversion = \"1.2.0\"\n")

	res, err := Discover(ws, []string{"pyproject.toml", "apps/api/pyproject.toml", "packages/sdk/pyproject.toml"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 2 {
		t.Fatalf("projects %+v", res.Projects)
	}
	api, sdk := res.Projects[0], res.Projects[1]
	if api.Kind != protocol.ProjectApp || api.Exports[0] != "acme_api" {
		t.Fatalf("api %+v", api)
	}
	if !reflect.DeepEqual(api.Imports, []string{"acme_core", "acme_utils", "os", "requests", "sys"}) {
		t.Fatalf("api imports %v", api.Imports)
	}
	if start := api.Tasks["start"].Command; start[len(start)-1] != "acme_api" || !strings.Contains(start[0], ".venv") {
		t.Fatalf("start %v", start)
	}
	if sdk.Kind != protocol.ProjectPackage || sdk.Exports[0] != "acme_sdk" || sdk.Tasks["build"].Outputs[0] != "dist" || sdk.Version != "1.2.0" {
		t.Fatalf("sdk %+v", sdk)
	}
}

func TestModuleNames(t *testing.T) {
	got := ModuleNames("import a.b, c\n    from d.e import f\nfrom .rel import g\n# import nope\n")
	if !reflect.DeepEqual(got, []string{"a", "c", "d"}) {
		t.Fatalf("got %v", got)
	}
}
