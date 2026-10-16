package releasing

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const extensionJSON = `{"name": "tools", "publisher": "acme", "version": "0.1.0", "engines": {"vscode": "^1.96.0"}, "main": "../../dist/packages/tools/main.js", "scripts": {"build": "esbuild"}}`

func extensionWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	for dir, body := range map[string]string{"packages/tools": extensionJSON, "dist/packages/tools": ""} {
		_ = os.MkdirAll(filepath.Join(ws, filepath.FromSlash(dir)), 0o755)
		if body != "" {
			_ = os.WriteFile(filepath.Join(ws, filepath.FromSlash(dir), "package.json"), []byte(body), 0o644)
		}
	}
	return ws
}

func TestPublishExtension(t *testing.T) {
	ws := extensionWorkspace(t)
	calls := fakeNpm(t, "")
	project := protocol.Project{Name: "tools", Root: "packages/tools", Exports: []string{"tools"}, Kind: protocol.ProjectPackage}
	t.Setenv("VSCE_PAT", "")
	if _, err := Publish(ws, protocol.PublishParams{Project: project, Version: "0.1.0"}); err == nil || !strings.Contains(err.Error(), "VSCE_PAT") {
		t.Fatalf("a publish without a token says which: %v", err)
	}
	if res, _ := Publish(ws, protocol.PublishParams{Project: project, Version: "0.1.0", DryRun: true}); res.Skipped != "dry run" {
		t.Fatalf("%+v", res)
	}
	t.Setenv("VSCE_PAT", "token")
	res, err := Publish(ws, protocol.PublishParams{Project: project, Version: "0.1.0"})
	if err != nil || !res.Published || res.Registry != "https://marketplace.visualstudio.com" {
		t.Fatalf("%+v %v", res, err)
	}
	last := (*calls)[len(*calls)-1]
	if !strings.HasPrefix(last, "npx --no-install vsce publish") || !strings.Contains(last, "--skip-duplicate") {
		t.Fatalf("vsce publishes: %v", *calls)
	}
	for _, c := range *calls {
		if strings.HasPrefix(c, "npm publish") {
			t.Fatal("an extension is never published to npm")
		}
	}
	staged, _ := os.ReadFile(filepath.Join(ws, "dist", "packages", "tools", "package.json"))
	if !strings.Contains(string(staged), `"main": "./main.js"`) {
		t.Fatalf("the staged manifest points at the bundle beside it:\n%s", staged)
	}
}

func TestPublishedExtension(t *testing.T) {
	ws := extensionWorkspace(t)
	project := protocol.Project{Root: "packages/tools", Exports: []string{"tools"}}
	real := capture
	defer func() { capture = real }()
	var asked []string
	capture = func(_ string, argv []string) (string, string, error) {
		asked = argv
		return `{"versions": [{"version": "0.3.0"}, {"version": "0.2.0"}]}`, "", nil
	}
	if res, _ := Published(ws, protocol.PublishedParams{Project: project}); res.Version != "0.3.0" || strings.Join(asked, " ") != "npx --no-install vsce show acme.tools --json" {
		t.Fatalf("%+v %v", res, asked)
	}
	capture = func(string, []string) (string, string, error) {
		return "", "ERROR  Extension 'acme.tools' not found.", errors.New("exit 1")
	}
	if res, _ := Published(ws, protocol.PublishedParams{Project: project}); res.Version != "" || res.Unreachable {
		t.Fatalf("an unknown extension is unpublished: %+v", res)
	}
	capture = func(string, []string) (string, string, error) { return "", "ECONNRESET", errors.New("exit 1") }
	if res, _ := Published(ws, protocol.PublishedParams{Project: project}); !res.Unreachable {
		t.Fatalf("a failure is unreachable: %+v", res)
	}
}
