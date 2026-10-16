//go:build e2e

package goldenpath

import (
	"archive/zip"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

// TestVscodeExtension scaffolds a VS Code extension with the
// vscode-extension stack, runs every task, packs a .vsix from the staged
// build in dist/, and plans its release.
func TestVscodeExtension(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildMnci(t, base), dir: base, env: identity()}
	w.mnci(true, "new", "mnci-e2e-vscode", "--ci", "github")
	w.dir = filepath.Join(base, "mnci-e2e-vscode")
	w.mnci(true, "add", "package", "tools", "--stack", "vscode-extension")

	out := w.mnci(true, "run", "lint", "typecheck", "test", "build", "package")
	for _, want := range []string{"tools:lint ✔", "tools:typecheck ✔", "tools:test ✔", "tools:build ✔", "tools:package ✔"} {
		if !strings.Contains(out, want) {
			t.Fatalf("%s did not pass:\n%s", want, out)
		}
	}
	// The .vsix holds the bundle and a manifest pointing at it.
	vsix, err := zip.OpenReader(filepath.Join(w.dir, "dist", "drop", "tools-0.0.0.vsix"))
	if err != nil {
		t.Fatalf("package wrote dist/drop/tools-0.0.0.vsix: %v", err)
	}
	defer func() { _ = vsix.Close() }()
	var manifest string
	found := map[string]bool{}
	for _, f := range vsix.File {
		found[f.Name] = true
		if f.Name == "extension/package.json" {
			r, _ := f.Open()
			raw, _ := io.ReadAll(r)
			_ = r.Close()
			manifest = string(raw)
		}
	}
	if !found["extension/main.js"] || !strings.Contains(manifest, `"main": "./main.js"`) {
		t.Fatalf("the .vsix has the bundle and its manifest points at it: %v\n%s", found, manifest)
	}
	if out := w.mnci(true, "run", "lint", "typecheck", "test", "build"); !strings.Contains(out, "(4 from cache") {
		t.Fatalf("a re-run replays from the cache:\n%s", out)
	}

	w.git("add", "-A")
	w.git("commit", "-qm", "feat: a hello command")
	if plan := w.mnci(true, "release", "--dry-run"); !strings.Contains(plan, "tools") || !strings.Contains(plan, "0.1.0") {
		t.Fatalf("release plan:\n%s", plan)
	}
}
