package protocol

// ContributeParams hands a convention plugin the projects it applies to:
// those of the languages it supports that mnci.toml enables it for.
type ContributeParams struct {
	Projects []Project `json:"projects"`
}

// ContributeResult is what a convention adds, by project name.
type ContributeResult struct {
	Projects map[string]Contribution `json:"projects,omitempty"`
}

// Contribution is a convention's addition to one project.
type Contribution struct {
	// Tasks are new tasks. A task the project already has is kept as it is.
	Tasks map[string]Task `json:"tasks,omitempty"`
	// DependsOn adds dependencies to tasks the project already has, by task
	// name ({"lint": ["slice-check"]}).
	DependsOn map[string][]string `json:"dependsOn,omitempty"`
}
