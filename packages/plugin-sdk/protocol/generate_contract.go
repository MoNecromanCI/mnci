package protocol

// GenerateParams asks a plugin to scaffold a project. The host chooses the
// directory from the kind (apps/, libs/, packages/).
type GenerateParams struct {
	Kind ProjectKind `json:"kind"`
	Name string      `json:"name"`
	// Directory is workspace-relative, forward slashes; it does not exist yet.
	Directory string `json:"directory"`
	// Workspace carries what a template needs to know about the monorepo.
	Workspace WorkspaceInfo `json:"workspace"`
}

// WorkspaceInfo is the subset of workspace settings plugins may use.
type WorkspaceInfo struct {
	Name string `json:"name"`
	// Scope is the npm scope for packages, with its @ (e.g. "@acme").
	Scope string `json:"scope,omitempty"`
	// ModulePrefix is the Go module path prefix, e.g. github.com/acme/repo.
	ModulePrefix string `json:"modulePrefix,omitempty"`
}

// GenerateResult lists what the plugin wrote, workspace-relative.
type GenerateResult struct {
	Files []string `json:"files"`
	// Notes are printed to the user after generation.
	Notes []string `json:"notes,omitempty"`
}
