package protocol

// InstallParams asks a plugin to add packages to one of its projects with
// the ecosystem's own toolchain, or, with no project, to install every
// dependency of the projects it serves (a fresh clone's first step).
type InstallParams struct {
	// Project is the project the packages go into; nil installs everything.
	Project *Project `json:"project,omitempty"`
	// Projects are the projects the plugin serves, when Project is nil.
	Projects []Project `json:"projects,omitempty"`
	// Packages are names or pinned specs as the ecosystem writes them
	// (left-pad, requests>=2.31, github.com/pkg/errors@v0.9.1).
	Packages []string `json:"packages,omitempty"`
	// Dev adds them as development dependencies where the ecosystem has
	// such a thing; a plugin without one says so in a note.
	Dev bool `json:"dev,omitempty"`
}

// UninstallParams asks a plugin to remove packages from one of its projects.
type UninstallParams struct {
	Project  Project  `json:"project"`
	Packages []string `json:"packages"`
}

// DependencyResult lists the workspace-relative files the plugin changed and
// notes for the user.
type DependencyResult struct {
	Files []string `json:"files,omitempty"`
	Notes []string `json:"notes,omitempty"`
}
