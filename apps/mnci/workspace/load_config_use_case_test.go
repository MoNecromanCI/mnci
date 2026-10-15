package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAppliesDefaults(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"acme\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Schema != 1 || cfg.CI.Providers[0] != "github" || cfg.Release.Branch != "main" || cfg.Cache.Dir != ".mnci/cache" || cfg.Cache.MaxSizeMB != 2048 {
		t.Fatalf("defaults %+v", cfg)
	}
}

func TestParseRejectsMistakes(t *testing.T) {
	for input, want := range map[string]string{
		"[workspace]\nname = \"a\"\nnmae = \"typo\"\n": "nmae",
		"[workspace]\n": "name is required",
		"[workspace]\nname = \"a\"\n[ci]\nproviders=[\"x\"]": "unknown CI provider",
		"schema = 9\n[workspace]\nname = \"a\"\n":            "upgrade mnci",
		"[workspace\nname = \"a\"\n":                         "mnci.toml:",
	} {
		_, err := Parse([]byte(input))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("input %q: want error containing %q, got %v", input, want, err)
		}
	}
}

func TestParseOverrides(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"a\"\n[tasks.test]\ncache = false\n[projects.web.tasks.build]\ndepends_on = [\"^build\", \"codegen\"]\n"))
	if err != nil {
		t.Fatal(err)
	}
	if *cfg.Tasks["test"].Cache || cfg.Projects["web"].Tasks["build"].DependsOn[1] != "codegen" {
		t.Fatalf("overrides %+v", cfg)
	}
}

func TestFindRoot(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "apps", "web", "src")
	_ = os.MkdirAll(deep, 0o755)
	_ = os.WriteFile(filepath.Join(root, FileName), []byte("[workspace]\nname=\"a\"\n"), 0o644)
	got, err := FindRoot(deep)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.EvalSymlinks(root)
	got, _ = filepath.EvalSymlinks(got)
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
	if _, err := FindRoot(t.TempDir()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
