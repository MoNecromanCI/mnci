package configedit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoNecromanCI/mnci/apps/mnci/workspace"
)

func TestSetChecksTheSchemaAndKeepsComments(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, workspace.FileName)
	_ = os.WriteFile(file, []byte("schema = 1\n\n[workspace]\nname = \"acme\" # ours\n\n[ci]\n# which pipelines\nproviders = [\"github\"]\n"), 0o644)
	for _, tc := range []struct {
		path string
		args []string
		err  string
	}{
		{"ci.providers", []string{"github", "gitlab"}, `takes github or azure, not "gitlab"`},
		{"ci.providers", []string{"github", "github"}, "twice"},
		{"ci.devcontainer", []string{"yes please"}, "true or false"},
		{"cache.max_size_mb", []string{"0"}, "at least 1"},
		{"workspace.name", []string{""}, "cannot be empty"},
		{"workspace.name", []string{"a", "b"}, "one value"},
		{"stacks.web", []string{"x"}, "not a setting"},
	} {
		if err := Set(root, tc.path, tc.args); err == nil || !strings.Contains(err.Error(), tc.err) {
			t.Errorf("%s %v: %v, want %q", tc.path, tc.args, err, tc.err)
		}
	}
	for _, step := range [][]string{{"ci.providers", "github", "azure"}, {"ci.devcontainer", "true"}, {"cache.max_size_mb", "512"}, {"workspace.name", `say "hi"`}} {
		if err := Set(root, step[0], step[1:]); err != nil {
			t.Fatal(err)
		}
	}
	raw, _ := os.ReadFile(file)
	cfg, err := workspace.Parse(raw)
	if err != nil {
		t.Fatalf("%v\n%s", err, raw)
	}
	if strings.Join(cfg.CI.Providers, ",") != "github,azure" || !cfg.CI.Devcontainer || cfg.Cache.MaxSizeMB != 512 || cfg.Workspace.Name != `say "hi"` {
		t.Fatalf("config %+v", cfg)
	}
	if !strings.Contains(string(raw), "# ours") || !strings.Contains(string(raw), "# which pipelines") {
		t.Fatalf("comments lost:\n%s", raw)
	}
	if err := Unset(root, "ci.devcontainer"); err != nil {
		t.Fatal(err)
	}
	if _, set, _ := Get(root, "ci.devcontainer"); set {
		t.Fatal("unset left the value")
	}
	if err := Unset(root, "ci.devcontainer"); err == nil {
		t.Fatal("unsetting a default is an error")
	}
}

func TestFieldsCoverTheFixedTables(t *testing.T) {
	fields, err := Fields()
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, f := range fields {
		paths = append(paths, f.Path)
		if f.Description == "" {
			t.Errorf("%s has no description", f.Path)
		}
	}
	got := strings.Join(paths, " ")
	for _, want := range []string{"workspace.name", "ci.providers", "ci.native_runners", "release.github_releases", "cache.max_size_mb", "plugins.enabled"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in %s", want, got)
		}
	}
	if strings.Contains(got, "plugins.external") {
		t.Error("a table of named entries is edited in the file")
	}
}
