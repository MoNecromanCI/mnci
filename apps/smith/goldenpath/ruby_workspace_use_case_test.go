//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRubyWorkspace builds gems with the built-in ruby plugin: an app
// depending on a package in one Bundler workspace, RuboCop and RSpec, the
// cache, and a release plan.
func TestRubyWorkspace(t *testing.T) {
	for _, tool := range []string{"ruby", "bundle"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is not installed")
		}
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildMnci(t, base), dir: base, env: identity()}
	// A name no gem on RubyGems has: the release checks ask RubyGems.
	w.mnci(true, "new", "mnci-e2e-rubyish", "--ci", "github")
	w.dir = filepath.Join(base, "mnci-e2e-rubyish")
	w.mnci(true, "add", "package", "core", "--lang", "ruby")
	w.mnci(true, "add", "app", "cli", "--lang", "ruby")

	// cli uses core: a gem dependency is a graph edge.
	spec := "apps/cli/mnci_e2e_rubyish_cli.gemspec"
	w.write(spec, strings.Replace(w.read(spec), "  spec.metadata['mnci_kind'] = 'app'\n", "  spec.metadata['mnci_kind'] = 'app'\n\n  spec.add_dependency 'mnci_e2e_rubyish_core'\n", 1))
	exe := "apps/cli/exe/mnci_e2e_rubyish_cli"
	w.write(exe, strings.Replace(w.read(exe), "puts MnciE2eRubyishCli::Greeting.greet('world')\n",
		"require 'mnci_e2e_rubyish_core'\n\nputs \"#{MnciE2eRubyishCli::Greeting.greet('world')} #{MnciE2eRubyishCore::Greeting.greet('core')}\"\n", 1))
	w.mnci(true, "install")
	if !hasEdge(w.mnci(true, "graph", "--format", "json"), "cli", "core") {
		t.Fatal("a gem dependency is a graph edge")
	}
	out := w.mnci(true, "run", "lint", "test")
	for _, want := range []string{"cli:lint ✔", "core:lint ✔", "cli:test ✔", "core:test ✔"} {
		if !strings.Contains(out, want) {
			t.Fatalf("%s did not pass:\n%s", want, out)
		}
	}
	if out := w.mnci(true, "run", "lint", "test"); !strings.Contains(out, "(4 from cache") {
		t.Fatalf("a re-run replays from the cache:\n%s", out)
	}
	if out := w.mnci(true, "run", "start", "-p", "cli"); !strings.Contains(out, "Hello, world! Hello, core!") {
		t.Fatalf("start:\n%s", out)
	}
	if out := w.mnci(true, "doctor"); !strings.Contains(out, "✔ Bundler is installed") {
		t.Fatalf("doctor:\n%s", out)
	}
	if ci := w.read(".github/workflows/ci.yml"); !strings.Contains(ci, "ruby/setup-ruby") || !strings.Contains(ci, "GEM_HOST_API_KEY") {
		t.Fatalf("ci.yml installs Ruby and passes the RubyGems key:\n%s", ci)
	}

	w.git("add", "-A")
	w.git("commit", "-qm", "feat: a cli and its core")
	if plan := w.mnci(true, "release", "--dry-run"); !strings.Contains(plan, "core") || !strings.Contains(plan, "0.1.0") {
		t.Fatalf("release plan:\n%s", plan)
	}
}
