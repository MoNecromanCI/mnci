package pluginkit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/libs/plugin-go/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestScaffoldWritesAPlugin(t *testing.T) {
	var ran []string
	real := scaffolding.GoRunner
	defer func() { scaffolding.GoRunner = real }()
	scaffolding.GoRunner = func(dir string, argv []string) error {
		ran = append(ran, strings.Join(argv, " "))
		if argv[1] == "mod" && argv[2] == "init" {
			return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+argv[3]+"\n"), 0o644)
		}
		return nil
	}
	ws := t.TempDir()
	_ = os.WriteFile(filepath.Join(ws, "go.work"), []byte("go 1.26\n"), 0o644)
	s := New("1.0.0")
	info, _ := s.Describe(protocol.InitializeParams{WorkspaceRoot: ws})
	if info.Stacks[0].Name != Name || info.Stacks[0].Scaffold != Name {
		t.Fatalf("stack %+v", info.Stacks)
	}
	params := protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "lint-rules", Directory: "apps/lint-rules", Stack: Name, Workspace: protocol.WorkspaceInfo{ModulePrefix: "example.com/acme"}}
	res, err := s.Generate(params)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(res.Files, ",")
	for _, want := range []string{"apps/lint-rules/main.go", "apps/lint-rules/plugin/lintrules_plugin_handler.go", "apps/lint-rules/plugin/lintrules_plugin_handler_test.go", "apps/lint-rules/go.sum"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %s in %v", want, res.Files)
		}
	}
	if !strings.Contains(strings.Join(ran, "|"), "go get "+SDK) {
		t.Fatalf("the SDK is required: %v", ran)
	}
	main, _ := os.ReadFile(filepath.Join(ws, "apps", "lint-rules", "main.go"))
	gomod, _ := os.ReadFile(filepath.Join(ws, "apps", "lint-rules", "go.mod"))
	if !strings.Contains(string(main), `"example.com/acme/apps/lint-rules/plugin"`) || !strings.Contains(string(gomod), "// mnci:stack go-mnci-plugin") {
		t.Fatalf("main.go:\n%s\ngo.mod:\n%s", main, gomod)
	}
	params.Kind, params.Directory = protocol.ProjectLib, "libs/x"
	if _, err := s.Generate(params); err == nil {
		t.Fatal("a plugin is an app")
	}
}
