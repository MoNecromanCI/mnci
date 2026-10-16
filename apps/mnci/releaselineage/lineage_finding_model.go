package releaselineage

// Finding is one problem with a project's release tags.
type Finding struct {
	Project string `json:"project,omitempty"`
	// Check names what was checked, for doctor.
	Check   string `json:"check"`
	Message string `json:"message"`
	// Blocking stops a release; otherwise it is a warning.
	Blocking bool `json:"blocking"`
}

// Subject is what the checks know about one releasable project.
type Subject struct {
	Project string
	// Prefix is its release tag prefix ("sdk@", "packages/gokit/v").
	Prefix string
	// Alternatives are prefixes its tags may have been made under before
	// (its package name, scoped or not).
	Alternatives []string
	// Tags are every tag in the repository.
	Tags []string
	// Published is the newest version its registry has; empty when it has
	// none or the plugin cannot ask.
	Published string
	// Unreachable means the registry could not be asked.
	Unreachable bool
	Registry    string
	Detail      string
}
