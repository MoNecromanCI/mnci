package angular

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/libs/plugin-node/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("1.0.0")
	res, _ := p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if len(res.Stacks) != 1 || strings.Join(res.Stacks[0].Toolchain, ",") != "angular,eslint,vitest" {
		t.Fatalf("%+v", res.Stacks)
	}
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "admin", Directory: "apps/admin", Stack: Name, Workspace: protocol.WorkspaceInfo{Name: "acme"}}); err != nil {
		t.Fatal(err)
	}
	read := func(rel string) string {
		raw, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	if config := read("apps/admin/angular.json"); !strings.Contains(config, `"base": "../../dist/apps/admin"`) || !strings.Contains(config, `"admin:build:development"`) {
		t.Errorf("angular.json builds into dist/:\n%s", config)
	}
	if root := read("package.json"); !strings.Contains(root, `"@angular/cli"`) || strings.Contains(root, `"@angular/core"`) {
		t.Errorf("the CLI is root tooling, the framework the app's:\n%s", root)
	}
	read("apps/admin/src/app/app.component.ts")
	read("apps/admin/src/greeting/greet.use-case.ts")
	if read("apps/admin/.gitignore") != ".angular/\n" {
		t.Error("the CLI's cache is ignored")
	}
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "ui", Directory: "packages/ui", Stack: Name}); err == nil {
		t.Error("a package is refused")
	}
}
