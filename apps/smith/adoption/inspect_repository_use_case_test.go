package adoption

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceName(t *testing.T) {
	for dir, want := range map[string]string{
		"Legacy_Repo": "legacy-repo",
		"my.app":      "my-app",
		"42-things":   "workspace-42-things",
		"acme":        "acme",
	} {
		if got := workspaceName(filepath.Join("/repos", dir)); got != want {
			t.Errorf("%s: %s, want %s", dir, got, want)
		}
	}
}

func TestScopeAndProviders(t *testing.T) {
	root := t.TempDir()
	if got := scopeOf(root, "acme"); got != "@acme" {
		t.Errorf("no package.json: %s", got)
	}
	_ = os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name": "@globex/monorepo", "dependencies": {"zod": "^3", "chalk": "^5"}}`), 0o644)
	if got := scopeOf(root, "acme"); got != "@globex" {
		t.Errorf("scoped root: %s", got)
	}
	if deps := rootRuntimeDependencies(root); len(deps) != 2 || deps[0] != "chalk" {
		t.Errorf("deps %v", deps)
	}
	exists := func(set ...string) func(string) bool {
		return func(p string) bool {
			for _, s := range set {
				if s == p {
					return true
				}
			}
			return false
		}
	}
	if got := providersOf(exists()); len(got) != 1 || got[0] != "github" {
		t.Errorf("default %v", got)
	}
	if got := providersOf(exists("azure-pipelines.yml")); len(got) != 1 || got[0] != "azure" {
		t.Errorf("azure %v", got)
	}
	if got := providersOf(exists(".github/workflows", "azure-pipelines.yml")); len(got) != 2 {
		t.Errorf("both %v", got)
	}
}

func TestAdoptable(t *testing.T) {
	r := Report{Blockers: []Item{{Blocks: BlocksRelease}}}
	if !r.Adoptable() {
		t.Error("a release blocker does not stop --init")
	}
	r.Blockers = append(r.Blockers, Item{Blocks: BlocksAdoption})
	if r.Adoptable() {
		t.Error("an adoption blocker does")
	}
}
