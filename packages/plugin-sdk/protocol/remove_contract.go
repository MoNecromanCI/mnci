package protocol

// RemoveParams tells a plugin a project is being removed. The host deletes
// the project's directory first; the plugin cleans what it owns outside it
// (a go.work entry, a lockfile, an installed package).
type RemoveParams struct {
	Project Project `json:"project"`
	// DryRun asks what would change, without changing it.
	DryRun bool `json:"dryRun,omitempty"`
}

// RemoveResult lists the workspace-relative files the plugin changed (or, in
// a dry run, would change), and notes for the user.
type RemoveResult struct {
	Files []string `json:"files,omitempty"`
	Notes []string `json:"notes,omitempty"`
}
