package tooling

// Setup is what a convention adds when a Node project is generated. Nothing
// already there is changed: every file, script, field and dependency is
// added only when absent.
type Setup struct {
	// RootFiles are written at the workspace root (path → content).
	RootFiles map[string]string
	// DevDependencies join the root package.json.
	DevDependencies map[string]string
	// Files are written in the project (project-relative path → content).
	Files map[string]string
	// Scripts join the project's package.json scripts.
	Scripts map[string]string
	// Fields join the top level of the project's package.json.
	Fields map[string]any
}
