package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("package.json", `{"name":"root","private":true}`)
	write("apps/web/package.json", `{"name":"@acme/web","version":"0.0.0","private":true,"scripts":{"build":"tsc","test":"jest","custom":"x"}}`)
	write("apps/web/src/main.ts", `import { greet } from '@acme/ui'`)
	write("libs/ui/package.json", `{"name":"@acme/ui","version":"0.0.0","private":true,"mnci":{"kind":"lib"},"dependencies":{"zod":"^3"}}`)

	res, err := Discover(ws, []string{"package.json", "apps/web/package.json", "libs/ui/package.json"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 2 {
		t.Fatalf("root manifest must be skipped: %+v", res.Projects)
	}
	web, ui := res.Projects[0], res.Projects[1]
	if web.Name != "web" || web.Kind != protocol.ProjectApp || web.Exports[0] != "@acme/web" {
		t.Fatalf("web: %+v", web)
	}
	if len(web.Imports) != 1 || web.Imports[0] != "@acme/ui" {
		t.Fatalf("web imports: %v", web.Imports)
	}
	build := web.Tasks["build"]
	if !build.Cache || build.Outputs[0] != "../../dist/apps/web" || build.DependsOn[0] != "^build" || build.Command[2] != "build" {
		t.Fatalf("build task: %+v", build)
	}
	if web.Tasks["custom"].Cache {
		t.Fatal("unknown scripts are not cached")
	}
	if ui.Kind != protocol.ProjectLib || ui.Imports[0] != "zod" {
		t.Fatalf("ui: %+v", ui)
	}
}

func TestDiscoverReportsBadJSON(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "x"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "x", "package.json"), []byte("{"), 0o644)
	if _, err := Discover(ws, []string{"apps/x/package.json"}); err == nil {
		t.Fatal("expected an error naming the bad manifest")
	}
}

func TestStartDependsOnBuildOnlyWhenThereIsOne(t *testing.T) {
	scripts := map[string]string{"start": "node src/main.js"}
	if deps := TaskFor("start", "apps/web", scripts).DependsOn; len(deps) != 0 {
		t.Fatalf("plain JavaScript has no build to wait for: %v", deps)
	}
	scripts["build"] = "tsc"
	if deps := TaskFor("start", "apps/web", scripts).DependsOn; len(deps) != 1 || deps[0] != "build" {
		t.Fatalf("start waits for the project's build: %v", deps)
	}
	if deps := TaskFor("test", "apps/web", scripts).DependsOn; len(deps) != 1 || deps[0] != "^build" {
		t.Fatalf("upstream builds are always kept: %v", deps)
	}
}
