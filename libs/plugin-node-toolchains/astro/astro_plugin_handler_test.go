package astro

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("1.0.0")
	res, _ := p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if len(res.Stacks) != 1 || res.Stacks[0].Name != DocsSite || strings.Join(res.Stacks[0].Toolchain, ",") != "astro,eslint" {
		t.Fatalf("%+v", res.Stacks)
	}
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "handbook", Directory: "apps/handbook", Stack: DocsSite, Workspace: protocol.WorkspaceInfo{Name: "acme"}}); err != nil {
		t.Fatal(err)
	}
	read := func(rel string) string {
		raw, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	if config := read("apps/handbook/astro.config.mjs"); !strings.Contains(config, "outDir:       '../../dist/apps/handbook',") || !strings.Contains(config, "title:   'handbook',") {
		t.Errorf("astro.config.mjs:\n%s", config)
	}
	if lint := read("apps/handbook/eslint.config.mjs"); !strings.Contains(lint, "import workspace from '../../eslint.config.mjs'") {
		t.Errorf("the site's lint extends the workspace's:\n%s", lint)
	}
	if manifest := read("apps/handbook/package.json"); !strings.Contains(manifest, `"build": "astro check && astro build"`) {
		t.Errorf("build checks first:\n%s", manifest)
	}
	read("apps/handbook/src/content/docs/guides/getting-started.md")
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "x", Directory: "libs/x", Stack: DocsSite}); err == nil {
		t.Error("a lib is refused")
	}
}
