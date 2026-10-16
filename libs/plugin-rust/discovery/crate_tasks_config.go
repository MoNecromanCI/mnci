package discovery

import "github.com/MoNecromanCI/mnci/packages/plugin-sdk/protocol"

// inputs are what a crate's results depend on: its sources and manifest, and
// the workspace's manifest and lockfile.
var inputs = []string{"**/*", "!target/**", "{workspaceRoot}/Cargo.toml", "{workspaceRoot}/Cargo.lock"}

// Tasks are cargo's, scoped to one package with -p: build, test, lint
// (clippy, warnings are errors), typecheck (check), format (fmt --check) and
// package (cargo package, what crates.io would get). Cargo builds the
// crates a crate depends on itself, so no task depends on another project's.
func Tasks(name string, kind protocol.ProjectKind) map[string]protocol.Task {
	cargo := func(args ...string) []string { return append([]string{"cargo"}, args...) }
	tasks := map[string]protocol.Task{
		"build":     {Command: cargo("build", "-p", name), Inputs: inputs, Description: "cargo build"},
		"test":      {Command: cargo("test", "-p", name), Inputs: inputs, Cache: true, Description: "cargo test"},
		"lint":      {Command: cargo("clippy", "-p", name, "--all-targets", "--", "-D", "warnings"), Inputs: inputs, Cache: true, Description: "cargo clippy, warnings are errors"},
		"typecheck": {Command: cargo("check", "-p", name, "--all-targets"), Inputs: inputs, Cache: true, Description: "cargo check"},
		"format":    {Command: cargo("fmt", "-p", name, "--check"), Inputs: inputs, Cache: true, Description: "cargo fmt --check"},
	}
	if kind == protocol.ProjectApp {
		tasks["start"] = protocol.Task{Command: cargo("run", "-p", name), Description: "Run the app"}
	}
	if kind == protocol.ProjectPackage {
		tasks["package"] = protocol.Task{Command: cargo("package", "-p", name, "--allow-dirty"), Description: "cargo package: the .crate crates.io would get"}
	}
	return tasks
}
