package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestNewWorkspaceInAnExistingClone(t *testing.T) {
	dir := t.TempDir()
	g := gitclient.New(dir)
	if err := g.Init(); err != nil {
		t.Skip("git unavailable")
	}
	_, _ = g.Run("config", "user.email", "t@e.com")
	_, _ = g.Run("config", "user.name", "T")
	_, _ = g.Run("config", "commit.gpgsign", "false")
	_, _ = g.Run("remote", "add", "origin", "git@github.com:Acme/acme-repo.git")
	_ = os.WriteFile(filepath.Join(dir, "LICENSE"), []byte("MIT"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("# mine"), 0o644)

	var out bytes.Buffer
	if _, err := NewWorkspace(NewOptions{Name: "acme", Dir: dir, Providers: []string{"github", "azure"}, Version: "1.0.0", Out: &out}); err != nil {
		t.Fatal(err)
	}
	cfg, err := workspace.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Workspace.Scope != "@acme" || cfg.Workspace.ModulePrefix != "github.com/Acme/acme-repo" || len(cfg.CI.Providers) != 2 {
		t.Fatalf("config %+v", cfg.Workspace)
	}
	if body, _ := os.ReadFile(filepath.Join(dir, "README.md")); string(body) != "# mine" {
		t.Fatal("an existing README must be kept")
	}
	for _, f := range []string{".github/workflows/ci.yml", "azure-pipelines.yml", ".gitignore", "LICENSE"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(f))); err != nil {
			t.Errorf("missing %s", f)
		}
	}
	if !strings.Contains(out.String(), "committed") || !g.HasCommits() {
		t.Fatalf("expected a first commit:\n%s", out.String())
	}
	if _, err := NewWorkspace(NewOptions{Name: "acme", Dir: dir, Out: &out}); err == nil {
		t.Fatal("creating over an existing workspace must fail")
	}
}

func TestValidateName(t *testing.T) {
	for _, ok := range []string{"web", "web-app", "a1"} {
		if ValidateName("app", ok) != nil {
			t.Errorf("%s must be valid", ok)
		}
	}
	for _, bad := range []string{"Web", "web_app", "1web", "web--app", "-web", "", "web/app"} {
		if ValidateName("app", bad) == nil {
			t.Errorf("%q must be invalid", bad)
		}
	}
}

func TestSplitKind(t *testing.T) {
	for in, want := range map[string][2]string{"node-app": {"app", "node"}, "python-package": {"package", "python"}, "my-lang-lib": {"lib", "my-lang"}} {
		if k, l, ok := SplitKind(in); !ok || k != want[0] || l != want[1] {
			t.Errorf("SplitKind(%q) = %s %s %v", in, k, l, ok)
		}
	}
	for _, in := range []string{"app", "node-web", "-app"} {
		if _, _, ok := SplitKind(in); ok {
			t.Errorf("SplitKind(%q) must not split", in)
		}
	}
}
