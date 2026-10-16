//go:build e2e

package goldenpath

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRustWorkspace builds Rust crates with the built-in rust plugin: an app
// and a library in one Cargo workspace, the app depending on the library by
// path, every cargo task, the cache, and a release plan.
func TestRustWorkspace(t *testing.T) {
	if _, err := exec.LookPath("cargo"); err != nil {
		t.Skip("cargo is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildMnci(t, base), dir: base, env: identity()}
	// A name no crate on crates.io has: the release checks ask the registry.
	w.mnci(true, "new", "mnci-e2e-rusty", "--ci", "github")
	w.dir = filepath.Join(base, "mnci-e2e-rusty")
	w.mnci(true, "add", "app", "cli", "--lang", "rust")
	w.mnci(true, "add", "package", "core", "--lang", "rust")
	if root := w.read("Cargo.toml"); !strings.Contains(root, `members = ["apps/cli", "packages/core"]`) {
		t.Fatalf("root Cargo.toml:\n%s", root)
	}

	// cli uses core by path: a graph edge, and core's code in cli's output.
	w.write("apps/cli/Cargo.toml", strings.Replace(w.read("apps/cli/Cargo.toml"), "[dependencies]\n", "[dependencies]\nmnci-e2e-rusty-core = { path = \"../../packages/core\", version = \"0.0.0\" }\n", 1))
	w.write("apps/cli/src/main.rs", "mod greeting;\n\nfn main() {\n    let core = mnci_e2e_rusty_core::greeting::greet(\"core\");\n    println!(\"{} {}\", greeting::greet(\"world\"), core);\n}\n")
	if !hasEdge(w.mnci(true, "graph", "--format", "json"), "cli", "core") {
		t.Fatal("a path dependency is a graph edge")
	}
	// The first cargo run writes Cargo.lock, an input of every task.
	w.mnci(true, "run", "typecheck")
	out := w.mnci(true, "run", "format", "lint", "typecheck", "test", "build")
	for _, want := range []string{"cli:lint ✔", "core:test ✔", "cli:build ✔", "core:format ✔"} {
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
	if out := w.mnci(true, "doctor"); !strings.Contains(out, "the Cargo workspace lists apps/cli") {
		t.Fatalf("doctor:\n%s", out)
	}
	if ci := w.read(".github/workflows/ci.yml"); !strings.Contains(ci, "dtolnay/rust-toolchain") || !strings.Contains(ci, "CARGO_REGISTRY_TOKEN") {
		t.Fatalf("ci.yml installs Rust and passes the crates.io token:\n%s", ci)
	}

	// The package releases a first version, through cargo publish.
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: a cli and its core")
	if plan := w.mnci(true, "release", "--dry-run"); !strings.Contains(plan, "core") || !strings.Contains(plan, "0.1.0") {
		t.Fatalf("release plan:\n%s", plan)
	}
}
