//go:build e2e

package goldenpath

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWebWorkspace scaffolds a browser app with each of the vite plugin's
// stacks and runs every task: lint, typecheck, test and the Vite build into
// dist/, then a cache replay.
func TestWebWorkspace(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildMnci(t, base), dir: base, env: identity()}
	w.mnci(true, "new", "mnci-e2e-web", "--ci", "github")
	w.dir = filepath.Join(base, "mnci-e2e-web")
	stacks := []string{"react", "react-p5", "vue", "svelte", "html"}
	for _, stack := range stacks {
		w.mnci(true, "add", "app", stack, "--stack", stack)
	}
	if out := w.mnci(false, "add", "lib", "ui", "--stack", "react"); !strings.Contains(out, "is an app") {
		t.Fatalf("a web stack scaffolds apps only:\n%s", out)
	}

	out := w.mnci(true, "run", "lint", "typecheck", "test", "build")
	for _, stack := range stacks {
		for _, task := range []string{"lint", "typecheck", "test", "build"} {
			if want := stack + ":" + task + " ✔"; !strings.Contains(out, want) {
				t.Fatalf("%s did not pass:\n%s", want, out)
			}
		}
		// Builds go to dist/, never dist/ in the app.
		if _, err := os.Stat(filepath.Join(w.dir, "dist", "apps", stack, "index.html")); err != nil {
			t.Errorf("%s did not build into dist/apps/%s", stack, stack)
		}
		if _, err := os.Stat(filepath.Join(w.dir, "apps", stack, "dist")); err == nil {
			t.Errorf("apps/%s/dist exists: builds must go to the workspace's dist/", stack)
		}
	}
	if out := w.mnci(true, "run", "lint", "typecheck", "test", "build"); !strings.Contains(out, "(20 from cache") {
		t.Fatalf("a re-run replays from the cache:\n%s", out)
	}
	if out := w.mnci(true, "stacks"); !strings.Contains(out, "react-p5") || !strings.Contains(out, "svelte") {
		t.Fatalf("the vite plugin's stacks are listed:\n%s", out)
	}
}
