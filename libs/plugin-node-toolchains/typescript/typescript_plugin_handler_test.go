package typescript

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/libs/plugin-node/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/serve"
)

func TestTypeScriptScaffoldsAndRecordsTheStack(t *testing.T) {
	scaffolding.Installer = func(string) error { return nil }
	ws := t.TempDir()
	p := New("1.0.0")
	if caps := serve.CapabilitiesOf(p); !caps.Generate || caps.Discover {
		t.Fatalf("capabilities %+v", caps)
	}
	info, _ := p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if info.Stacks[0].Scaffold != Name || strings.Join(info.Provides, ",") != "build,typecheck" {
		t.Fatalf("describe %+v", info)
	}
	res, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "sdk", Directory: "packages/sdk", Stack: "node-ts", Workspace: protocol.WorkspaceInfo{Name: "Acme", Scope: "acme"}})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(res.Files, ",")
	for _, want := range []string{"tsconfig.base.json", ".npmrc", "packages/sdk/src/greeting/greet.use-case.ts", "packages/sdk/tsconfig.build.json"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %s in %v", want, res.Files)
		}
	}
	manifest, _ := os.ReadFile(filepath.Join(ws, "packages", "sdk", "package.json"))
	for _, want := range []string{`"stack": "node-ts"`, `"build": "tsc -p tsconfig.build.json"`, `"name": "@acme/sdk"`} {
		if !strings.Contains(string(manifest), want) {
			t.Errorf("package.json lacks %s:\n%s", want, manifest)
		}
	}
	root, _ := os.ReadFile(filepath.Join(ws, "package.json"))
	if !strings.Contains(string(root), `"typescript"`) {
		t.Fatalf("root devDependencies:\n%s", root)
	}
}
