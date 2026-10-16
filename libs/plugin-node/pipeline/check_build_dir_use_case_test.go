package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestCheckBuildDir(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	builds := map[string]protocol.Task{"build": {}}
	write("apps/web/tsconfig.json", "{\n  // ours\n  \"compilerOptions\": { \"outDir\": \"../../dist/apps/web\" }\n}\n")
	if f, ok := CheckBuildDir(ws, protocol.Project{Name: "web", Root: "apps/web", Tasks: builds}); !ok || !f.OK {
		t.Fatalf("web: %+v %v", f, ok)
	}
	write("apps/old/tsconfig.build.json", `{"compilerOptions": {"outDir": "dist"}}`)
	f, ok := CheckBuildDir(ws, protocol.Project{Name: "old", Root: "apps/old", Tasks: builds})
	if !ok || f.OK || f.Remedy == "" {
		t.Fatalf("old: %+v", f)
	}
	if _, ok := CheckBuildDir(ws, protocol.Project{Name: "ui", Root: "libs/ui"}); ok {
		t.Fatal("a project without a build has nothing to check")
	}
}
