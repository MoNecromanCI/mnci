package scaffolding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/libs/plugin-ruby/discovery"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

func TestGenerate(t *testing.T) {
	ws := t.TempDir()
	real := Runner
	defer func() { Runner = real }()
	installs := 0
	Runner = func(string, []string) error { installs++; return nil }
	info := protocol.WorkspaceInfo{Name: "acme-corp"}
	if _, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectPackage, Name: "core", Directory: "packages/core", Workspace: info, Stack: "ruby"}); err != nil {
		t.Fatal(err)
	}
	res, err := Generate(ws, protocol.GenerateParams{Kind: protocol.ProjectApp, Name: "web-cli", Directory: "apps/web-cli", Workspace: info})
	if err != nil || installs != 2 || !strings.Contains(strings.Join(res.Files, ","), "Gemfile,") {
		t.Fatal(err, installs, res.Files)
	}
	gemfileSrc, _ := os.ReadFile(filepath.Join(ws, "Gemfile"))
	if !strings.Contains(string(gemfileSrc), "resolution, one Gemfile.lock, for all of them.\n\ngemspec path: 'packages/core'\ngemspec path: 'apps/web-cli'\n\ngroup :development") {
		t.Fatalf("Gemfile:\n%s", gemfileSrc)
	}
	exe, _ := os.ReadFile(filepath.Join(ws, "apps/web-cli/exe/acme_corp_web_cli"))
	if !strings.Contains(string(exe), "puts AcmeCorpWebCli::Greeting.greet('world')") {
		t.Fatalf("exe:\n%s", exe)
	}
	found, err := discovery.Discover(ws, []string{"packages/core/acme_corp_core.gemspec", "apps/web-cli/acme_corp_web_cli.gemspec"})
	if err != nil || len(found.Projects) != 2 {
		t.Fatalf("%v %+v", err, found.Projects)
	}
	core, cli := found.Projects[0], found.Projects[1]
	if core.Kind != protocol.ProjectPackage || core.Stack != "ruby" || core.Exports[0] != "acme_corp_core" || core.Tasks["test"].Command == nil {
		t.Errorf("core %+v", core)
	}
	if cli.Kind != protocol.ProjectApp || cli.Tasks["start"].Command[3] != "exe/acme_corp_web_cli" {
		t.Errorf("cli %+v", cli)
	}
	if out, ok := DropGemspec(string(gemfileSrc), "packages/core"); !ok || strings.Contains(out, "packages/core") {
		t.Fatalf("dropped:\n%s", out)
	}
}

func TestAddGemspecToAnExistingGemfile(t *testing.T) {
	out := AddGemspec("source 'https://rubygems.org'\n\ngem 'rails'\n", "libs/a")
	if out != "source 'https://rubygems.org'\n\ngemspec path: 'libs/a'\n\ngem 'rails'\n" {
		t.Fatalf("%q", out)
	}
	if AddGemspec(out, "libs/a") != out {
		t.Fatal("a listed gem is not added twice")
	}
}
