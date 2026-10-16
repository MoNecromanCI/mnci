package dependencies

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/libs/plugin-ruby/scaffolding"
	"github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"
)

const gemspec = `Gem::Specification.new do |s|
  s.name = 'acme_cli'
  s.add_dependency 'acme_core'
end
`

func TestAddAndRemoveDependencies(t *testing.T) {
	out := AddDependencies(gemspec, []string{"acme_core", "thor"}, false)
	out = AddDependencies(out, []string{"webmock"}, true)
	if out != "Gem::Specification.new do |s|\n  s.name = 'acme_cli'\n  s.add_dependency 'acme_core'\n  s.add_dependency 'thor'\n  s.add_development_dependency 'webmock'\nend\n" {
		t.Fatalf("%q", out)
	}
	if out := RemoveDependencies(out, []string{"thor", "webmock"}); out != gemspec {
		t.Fatalf("%q", out)
	}
}

func TestInstall(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "cli"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "cli", "acme_cli.gemspec"), []byte(gemspec), 0o644)
	var ran []string
	real := scaffolding.Runner
	defer func() { scaffolding.Runner = real }()
	scaffolding.Runner = func(_ string, argv []string) error { ran = append(ran, strings.Join(argv, " ")); return nil }
	if _, err := Install(ws, protocol.InstallParams{}); err != nil {
		t.Fatal(err)
	}
	cli := protocol.Project{Root: "apps/cli"}
	res, err := Install(ws, protocol.InstallParams{Project: &cli, Packages: []string{"thor"}})
	if err != nil || res.Files[0] != "apps/cli/acme_cli.gemspec" {
		t.Fatal(err, res)
	}
	if src, _ := os.ReadFile(filepath.Join(ws, "apps", "cli", "acme_cli.gemspec")); !strings.Contains(string(src), "s.add_dependency 'thor'") {
		t.Fatalf("%s", src)
	}
	if _, err := Uninstall(ws, protocol.UninstallParams{Project: cli, Packages: []string{"thor"}}); err != nil || strings.Join(ran, "|") != "bundle install|bundle install|bundle install" {
		t.Fatal(err, ran)
	}
}
