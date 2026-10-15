package protocol

// VersionParams asks a plugin to record a new version in a project manifest.
type VersionParams struct {
	Project Project `json:"project"`
	Version string  `json:"version"`
	// Dependencies maps exported identifiers of other workspace projects to
	// their new versions, so the plugin can update dependency ranges.
	Dependencies map[string]string `json:"dependencies,omitempty"`
	DryRun       bool              `json:"dryRun,omitempty"`
}

// VersionResult lists the files that were (or, in a dry run, would be) changed.
type VersionResult struct {
	Files []string `json:"files"`
}

// PublishParams asks a plugin to publish one project at one version.
type PublishParams struct {
	Project Project `json:"project"`
	Version string  `json:"version"`
	DryRun  bool    `json:"dryRun,omitempty"`
}

// PublishResult reports what happened. Skipped with a reason is a success:
// an already-published version makes a re-run safe.
type PublishResult struct {
	Published bool   `json:"published"`
	Skipped   string `json:"skipped,omitempty"`
	Registry  string `json:"registry,omitempty"`
}
