package vite

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDescribe(t *testing.T) {
	res, err := New("1.0.0").Describe(protocol.InitializeParams{})
	if err != nil || len(res.Stacks) != 5 {
		t.Fatal(err, res.Stacks)
	}
	for _, s := range res.Stacks {
		if s.Scaffold != Name || strings.Join(s.Toolchain, ",") != "vite,eslint,vitest" || s.Description == "" {
			t.Errorf("stack %+v", s)
		}
	}
}

func TestGenerate(t *testing.T) {
	real := scaffolding.Installer
	defer func() { scaffolding.Installer = real }()
	scaffolding.Installer = func(string) error { return nil }
	for stack, want := range map[string]struct{ entry, typecheck, dep, devDep string }{
		"react":    {"src/main.tsx", "tsc -p tsconfig.json --noEmit", "react-dom", "@vitejs/plugin-react"},
		"react-p5": {"src/bouncing-ball/bouncing-ball.client.ts", "tsc -p tsconfig.json --noEmit", "@p5-wrapper/react", "@types/react"},
		"vue":      {"src/app/app.component.vue", "vue-tsc -p tsconfig.json --noEmit", "vue", "vue-tsc"},
		"svelte":   {"src/app/app.component.svelte", "svelte-check --tsconfig ./tsconfig.json --fail-on-warnings", "svelte", "svelte-check"},
		"html":     {"src/app/show-greeting.handler.ts", "tsc -p tsconfig.json --noEmit", "", "vite"},
	} {
		t.Run(stack, func(t *testing.T) {
			ws := t.TempDir()
			p := New("1.0.0")
			_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
			if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: stack, Workspace: protocol.WorkspaceInfo{Name: "acme", Scope: "@acme"}}); err != nil {
				t.Fatal(err)
			}
			read := func(rel string) []byte {
				raw, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(rel)))
				if err != nil {
					t.Fatal(err)
				}
				return raw
			}
			read("apps/web/" + want.entry)
			read("apps/web/src/greeting/greet.use-case.ts")
			var manifest struct {
				Name         string
				Smith        struct{ Kind, Stack string }
				Scripts      map[string]string
				Dependencies map[string]string
			}
			if err := json.Unmarshal(read("apps/web/package.json"), &manifest); err != nil {
				t.Fatal(err)
			}
			if manifest.Name != "@acme/web" || manifest.Smith.Kind != "app" || manifest.Smith.Stack != stack || manifest.Scripts["typecheck"] != want.typecheck || manifest.Scripts["build"] != "vite build" {
				t.Errorf("manifest %+v", manifest)
			}
			if want.dep != "" && manifest.Dependencies[want.dep] == "" {
				t.Errorf("the app depends on %s: %+v", want.dep, manifest.Dependencies)
			}
			var root struct{ DevDependencies map[string]string }
			if err := json.Unmarshal(read("package.json"), &root); err != nil {
				t.Fatal(err)
			}
			if root.DevDependencies[want.devDep] == "" || root.DevDependencies["vite"] == "" {
				t.Errorf("the root has the tooling: %+v", root.DevDependencies)
			}
			if config := string(read("apps/web/vite.config.ts")); !strings.Contains(config, "outDir:      '../../dist/apps/web',") {
				t.Errorf("the build goes to dist/:\n%s", config)
			}
			var tsconfig map[string]any
			if err := json.Unmarshal(read("apps/web/tsconfig.json"), &tsconfig); err != nil {
				t.Fatalf("tsconfig is JSON: %v", err)
			}
			if index := string(read("apps/web/index.html")); !strings.Contains(index, `src="/src/`+apps[stack].Entry+`"`) {
				t.Errorf("index.html loads the entry:\n%s", index)
			}
		})
	}
}

func TestGenerateRefuses(t *testing.T) {
	p := New("1.0.0")
	_, _ = p.Describe(protocol.InitializeParams{WorkspaceRoot: t.TempDir()})
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectLib, Name: "ui", Directory: "libs/ui", Stack: "react"}); err == nil || !strings.Contains(err.Error(), "smith add app ui --stack react") {
		t.Errorf("a lib is refused: %v", err)
	}
	if _, err := p.Generate(protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web", Directory: "apps/web", Stack: "angular"}); err == nil {
		t.Error("an unknown stack is refused")
	}
}
